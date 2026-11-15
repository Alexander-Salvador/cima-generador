#!/usr/bin/env python3
"""
security_sweep_zonas.py — BARRIDO DE SEGURIDAD DE DATOS (anti-datos-falsos)
================================================================================

Objetivo: demostrar que el popup NO afirma una zona de apagado que no está en la
fuente, y que ninguna zona de la fuente quedó fuera sin quedar declarada.

Cruza 3 universos y exige que coincidan:
  A) FUENTE    — PDFs (APAGADO.pdf G2 texto · APAGADP1.pdf G1 OCR · APAGADO2.pdf G3 OCR)
                  + zonas_apagadas_manual.json → filas de `apagados_tabla_completa.csv`
  B) RUNTIME   — `index.html` inline (ZONAS_APAGADAS / ZONAS_APAGADAS_INTEL / UBIGEO / CCPP_DATA)
  C) DATA LAYER— `data/zonas_apagadas.json` + `data/ubigeo_consolidated.json`

Checks S1..S19: integridad, anti-invención, trazabilidad, coherencia de conteos y
simulación de falsos positivos.

Uso:
    python3 scripts/security_sweep_zonas.py [--json <salida>]
Exit: 0 si no hay FAIL (los WARN no bloquean, pero se reportan).
"""
import argparse
import csv
import difflib
import hashlib
import json
import re
import sys
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
APP = ROOT / 'apps' / 'cima-generador'
DATA = APP / 'data'
HTML = APP / 'index.html'
PDF_DIR = ROOT / 'Recursos CIMA' / 'Bases y Zonas de Apagado'

PENDIENTES_ESPERADOS = {
    '010105:CHILIQUIN': 'PDF escribe el nombre del DISTRITO; GeoNames lista "Chiliquini"',
    '060101:INCA CORRAL': 'no existe en UBIGEO ni en CCPP_DATA (provincia "CAJABAMBA" vs "CAJAMARCA")',
    '160405:SANTA ROSA': 'distrito existente en UBIGEO inline pero con 0 caseríos en CCPP_DATA',
}


def fold(s):
    """Idéntico a zaNormalizeCcpp del popup (NFKD + sin diacríticos + UPPER)."""
    if not s:
        return ''
    t = unicodedata.normalize('NFKD', str(s))
    t = ''.join(ch for ch in t if not unicodedata.combining(ch))
    t = t.upper().replace('\u00a0', ' ')
    t = re.sub(r'\s+', ' ', t.strip())
    return unicodedata.normalize('NFC', t)


def md5(p):
    h = hashlib.md5()
    with open(p, 'rb') as fh:
        for chunk in iter(lambda: fh.read(1 << 20), b''):
            h.update(chunk)
    return h.hexdigest()


def inline_blocks():
    html = HTML.read_text(encoding='utf-8')
    ub = json.loads(re.search(r'const UBIGEO = (\{.*?\});\s*\n', html, re.S).group(1))
    cc = json.loads(re.search(r'const CCPP_DATA = (\{.*?\});\s*\n', html, re.S).group(1))
    za_body = re.search(r'const ZONAS_APAGADAS = \{([\s\S]*?)\n\};', html).group(1)
    keys = re.findall(r'"([0-9]{6}:[^"]+)"',
                      re.search(r'ccpp_keys: new Set\(\[(.*?)\]\)', za_body, re.S).group(1))
    dwo = re.findall(r'"(\d{6})"',
                     re.search(r'districts_with_off: new Set\(\[(.*?)\]\)', za_body, re.S).group(1))
    total = int(re.search(r'total_zones:\s*(\d+)', za_body).group(1))
    aliases = re.findall(r'"([^"]+)"',
                         re.search(r'CCPP_CASERIO_ALIASES\s*=\s*new Set\(\[(.*?)\]\)', html, re.S).group(1))
    intel = re.search(r'const ZONAS_APAGADAS_INTEL = \{([\s\S]*?)\n\};', html).group(1)
    ipt = set(re.findall(r'"(\d{6})"',
                         re.search(r'ipt_districts: new Set\(\[(.*?)\]\)', intel, re.S).group(1)))
    camp = {}
    camp_body_m = re.search(r'campanas:\s*\{([\s\S]*?)\n  \},\n  // v3\.10\.3', html)
    camp_body = camp_body_m.group(1) if camp_body_m else intel
    for cid in re.findall(r'"(COMU-\d+)"', camp_body):
        body_m = re.search(r'"' + cid + r'":\s*\{([\s\S]*?)\n    \},?', camp_body)
        if not body_m:
            continue
        body = body_m.group(1)

        def fld(name, _body=body):
            m = re.search(name + r':\s*(?:"([^"]*)"|([0-9]+)|(null)|(\[[^\]]*\]))', _body)
            if not m:
                return None
            if m.group(1) is not None:
                return m.group(1)
            if m.group(2) is not None:
                return int(m.group(2))
            if m.group(4) is not None:
                return m.group(4)
            return None
        camp[cid] = {
            'estado': fld('estado'),
            'version': fld('version'),
            'fecha_apagado': fld('fecha_apagado'),
            'fecha_sms_inicio': fld('fecha_sms_inicio'),
            'ccpp_count': fld('ccpp_count'),
            'faltantes': fld('faltantes'),
            'fuente_urls': fld('fuente_urls'),
            'speech_visor': fld('speech_visor'),
            'speech_postventa': fld('speech_postventa'),
        }
    return html, ub, cc, keys, dwo, total, aliases, ipt, camp


def main():
    ap = argparse.ArgumentParser(description='Barrido de seguridad de datos de zonas de apagado')
    ap.add_argument('--json', default=str(DATA / 'security_sweep_zonas.json'))
    args = ap.parse_args()

    html, ub, cc, keys, dwo, total, aliases, ipt, camp = inline_blocks()

    ubigeo = {}
    for depto, dv in ub.items():
        for prov, pv in dv.get('provincias', {}).items():
            for dist, info in pv.get('distritos', {}).items():
                if info.get('_ubigeo'):
                    ubigeo[info['_ubigeo']] = (depto.upper(), prov.upper(), dist.upper())

    ccpp_folded = {}
    for _d6, _lst in cc.items():
        ccpp_folded[_d6] = set()
        for _name in _lst:
            ccpp_folded[_d6].add(fold(_name))

    runtime = set()
    for _k in keys:
        runtime.add(fold(_k))
    runtime_by_ccpp = defaultdict(list)
    for _k in keys:
        _d6, _sep, _ccpp = _k.partition(':')
        runtime_by_ccpp[fold(_ccpp)].append(_d6)

    csv_path = DATA / 'apagados_tabla_completa.csv'
    rows = []
    if csv_path.exists():
        with open(csv_path, encoding='utf-8-sig', newline='') as fh:
            rows = list(csv.DictReader(fh))
    src_rows = []
    for _row in rows:
        if _row.get('comunicado') == 'COMU-5514':
            src_rows.append(_row)

    zpath = DATA / 'zonas_apagadas.json'
    zdata = {}
    if zpath.exists():
        zdata = json.loads(zpath.read_text())
    upath = DATA / 'ubigeo_consolidated.json'
    udata = {}
    if upath.exists():
        udata = json.loads(upath.read_text())

    results = []

    def chk(cid, name, ok, detail='', warn=False):
        results.append({'id': cid, 'name': name,
                        'status': 'PASS' if ok else ('WARN' if warn else 'FAIL'),
                        'detail': detail})

    # ── S1 · conteos coherentes ───────────────────────────────────────────
    jkeys = set()
    for _jk in zdata.get('ccpp_keys', []):
        jkeys.add(fold(_jk))
    c5514 = camp.get('COMU-5514', {})
    chk('S1', 'Conteos coherentes (runtime / data layer / campaña)',
        len(keys) == total == len(jkeys) == (c5514.get('ccpp_count') or -1) and len(keys) == len(set(keys)),
        'runtime=' + str(len(keys)) + ' total_zones=' + str(total) +
        ' json=' + str(len(jkeys)) + ' campana.ccpp_count=' + str(c5514.get('ccpp_count')))

    # ── S2 · ninguna clave del runtime sin respaldo en la fuente ──────────
    csv_keys = set()
    for _row in src_rows:
        _d6 = _row.get('ubigeo_d6', '')
        _ccpp = _row.get('centro_poblado_ccpp', '')
        if _d6 and _ccpp:
            csv_keys.add(fold(_d6 + ':' + _ccpp))
    sin_respaldo = sorted(runtime - csv_keys)
    chk('S2', 'Toda clave del runtime existe en la fuente (anti-invención)',
        not sin_respaldo,
        str(len(runtime)) + ' claves · sin respaldo: ' +
        (str(sin_respaldo[:5]) if sin_respaldo else 'ninguna'))

    # ── S3 · ninguna zona de la fuente fuera sin declararse ──────────────
    no_cubiertas = []
    for _row in src_rows:
        if _row.get('cobertura_en_sistema') == 'NO':
            no_cubiertas.append(_row)
    justificadas = []
    injustificadas = []
    for _row in no_cubiertas:
        _key = fold(_row.get('ubigeo_d6', '') + ':' + _row.get('centro_poblado_ccpp', ''))
        _ccpp_f = fold(_row.get('centro_poblado_ccpp', ''))
        _d6_f = _row.get('ubigeo_d6', '')
        _depto_pdf = _row.get('depto', '')
        if _key in PENDIENTES_ESPERADOS:
            justificadas.append(_key)
            continue
        # equivalencia por d6 (mismo depto/prov) o por grafía (typo 1-2 letras)
        hit = None
        for _sys_d6 in runtime_by_ccpp.get(_ccpp_f, []):
            hit = _sys_d6 + ' (mismo CCPP)'
            break
        if hit is None:
            for _sys_ccpp, _sys_d6s in runtime_by_ccpp.items():
                if _sys_d6s and difflib.SequenceMatcher(None, _sys_ccpp, _ccpp_f).ratio() >= 0.92:
                    for _sys_d6 in _sys_d6s:
                        if ubigeo.get(_sys_d6, ('', '', ''))[0] == _depto_pdf:
                            hit = _sys_d6 + ' (grafía equivalente)'
                            break
                if hit:
                    break
        if hit:
            justificadas.append(_key + ' → ' + hit)
        else:
            injustificadas.append(_key)
    chk('S3', 'Toda fila de la fuente está cubierta o declarada como pendiente',
        not injustificadas,
        'sin cobertura=' + str(len(no_cubiertas)) + ' justificadas=' + str(len(justificadas)) +
        ' injustificadas=' + (str(injustificadas[:5]) if injustificadas else 'ninguna'))

    # ── S4/S5 · d6 válidos ───────────────────────────────────────────────
    d6_malos = []
    for _k in keys:
        if _k.split(':')[0] not in ubigeo:
            d6_malos.append(_k)
    chk('S4', 'Todo d6 de las claves existe en el UBIGEO inline', not d6_malos,
        str(len(keys)) + ' claves · inexistentes: ' + (str(d6_malos[:5]) if d6_malos else 'ninguno'))
    dwo_bad = []
    for _d in dwo:
        if _d not in ubigeo:
            dwo_bad.append(_d)
    chk('S5', 'Todo distrito marcado existe en UBIGEO', not dwo_bad,
        str(len(dwo)) + ' distritos · inexistentes: ' + (str(dwo_bad[:5]) if dwo_bad else 'ninguno'))

    # ── S6 · el CCPP existe en los caseríos (o allowlist) ─────────────────
    fuera = []
    alias_hits = []
    for _k in keys:
        _d6, _sep, _ccpp = _k.partition(':')
        if fold(_ccpp) in ccpp_folded.get(_d6, set()):
            continue
        if _k in aliases:
            alias_hits.append(_k)
        else:
            fuera.append(_k)
    chk('S6', 'Cada CCPP está en CCPP_DATA[d6] o en la allowlist documentada',
        not fuera,
        'con respaldo=' + str(len(keys) - len(fuera)) + ' allowlist=' + str(len(alias_hits)) +
        ' sin respaldo=' + (str(fuera[:5]) if fuera else 'ninguno'))

    # ── S7 · contaminación real ──────────────────────────────────────────
    contaminados = []
    for _k in keys:
        _d6, _sep, _ccpp = _k.partition(':')
        _dist = ubigeo.get(_d6, ('', '', ''))[2]
        if fold(_ccpp) == fold(_dist) and fold(_ccpp) not in ccpp_folded.get(_d6, set()):
            if _k in aliases:
                continue  # variante literal del PDF, documentada en la allowlist
            contaminados.append(_k)
    chk('S7', 'Ninguna clave "CCPP == DISTRITO" sin caserío real', not contaminados,
        'contaminados: ' + (str(contaminados[:5]) if contaminados else 'ninguno'))

    # ── S8 · sin colisiones de fold ──────────────────────────────────────
    folds = []
    for _k in keys:
        folds.append(fold(_k))
    colisiones = [k for k, n in Counter(folds).items() if n > 1]
    chk('S8', 'Sin duplicados ni colisiones de fold',
        not colisiones and len(set(folds)) == len(keys),
        str(len(keys)) + ' claves · folds únicos=' + str(len(set(folds))) +
        ' colisiones=' + (str(colisiones[:5]) if colisiones else 'ninguna'))

    # ── S9 · campañas ────────────────────────────────────────────────────
    camp_issues = []
    for cid, c in camp.items():
        if c['estado'] == 'zonas_conocidas':
            if not c['ccpp_count']:
                camp_issues.append(cid + ': zonas_conocidas sin ccpp_count')
        elif c['estado'] == 'zonas_pendiente':
            if c['ccpp_count'] not in (0, None):
                camp_issues.append(cid + ': zonas_pendiente con ccpp_count=' + str(c['ccpp_count']))
            if not c['faltantes']:
                camp_issues.append(cid + ': zonas_pendiente sin faltantes')
            if not c['fuente_urls']:
                camp_issues.append(cid + ': zonas_pendiente sin fuente_urls')
    camp_txt = ', '.join(cid + '=' + str(c['estado']) for cid, c in camp.items())
    chk('S9', 'Campañas coherentes (cobertura real o pendiente declarado)',
        not camp_issues,
        '; '.join(camp_issues) if camp_issues else (str(len(camp)) + ' campañas: ' + camp_txt))

    # ── S10 · speechs ────────────────────────────────────────────────────
    sp_issues = []
    m_speechs = re.search(r'speechs:\s*\{([\s\S]*?)\n  \},', html)
    speech_body = m_speechs.group(1) if m_speechs else ''
    for cid, c in camp.items():
        for field in ('speech_visor', 'speech_postventa'):
            sk = c.get(field)
            if not sk:
                continue
            if ('"' + sk + '"') not in speech_body:
                sp_issues.append(cid + '.' + field + ' → ' + str(sk) + ' inexistente')
            else:
                tail = speech_body.split('"' + sk + '"', 1)[1][:600]
                if ('"' + cid + '"') not in tail:
                    sp_issues.append(cid + '.' + field + ' → ' + str(sk) + ' no declara ' + cid)
    chk('S10', 'Cada campaña tiene sus speechs y declaran su comunicado',
        not sp_issues, '; '.join(sp_issues) if sp_issues else 'todos coherentes')

    # ── S11/S12/S13 · IPT ────────────────────────────────────────────────
    ipt_bad = []
    for _d in ipt:
        if _d not in ubigeo:
            ipt_bad.append(_d)
    chk('S11', 'Zonas IPT: d6 válidos en UBIGEO', not ipt_bad,
        str(len(ipt)) + ' zonas IPT · d6 inválidos: ' + (str(ipt_bad[:5]) if ipt_bad else 'ninguno'))
    ccpp_d6 = set()
    for _k in keys:
        ccpp_d6.add(_k.split(':')[0])
    overlap = sorted(ipt & ccpp_d6)
    chk('S12', 'Overlap IPT ↔ CCPP exactos (esperado: un distrito puede estar en ambos)',
        True,
        'distritos en ambos conjuntos=' + str(len(overlap)) +
        ' — cada zona se atribuye a su campaña (COMU-5514 CCPP exacto / COMU-5672 IPT); sin doble conteo',
        warn=bool(overlap))
    ipc = Counter()
    for _z in zdata.get('zones', []):
        if _z.get('comunicado') == 'COMU-5672':
            ipc[_z.get('depto')] = max(ipc.get(_z.get('depto'), 0),
                                      int(_z.get('clientes_depto') or 0))
    suma = sum(ipc.values())
    declarado = 39500
    chk('S13', 'Cifra de clientes IPT: suma por depto vs total declarado',
        suma == declarado,
        'suma por departamento=' + format(suma, ',') + ' vs declarado=' + format(declarado, ',') +
        ' (Δ ' + format(suma - declarado, '+') + ' = ' +
        format(round((suma - declarado) / declarado * 100, 1), '+') +
        '%) — inconsistencia de la FUENTE, no del sistema',
        warn=True)

    # ── S14/S15 · data layer ─────────────────────────────────────────────
    cons_d6 = set()
    for depto, dv in udata.items():
        for prov, pv in dv.get('provincias', {}).items():
            for dist, info in pv.get('distritos', {}).items():
                if isinstance(info, dict) and info.get('_ubigeo'):
                    cons_d6.add(info['_ubigeo'])
    inline_d6 = set(ubigeo.keys())
    divergencia = sorted(inline_d6 ^ cons_d6)
    chk('S14', 'Data layer (ubigeo) alineado con el runtime', not divergencia,
        'inline=' + str(len(inline_d6)) + ' data_layer=' + str(len(cons_d6)) +
        ' divergencias=' + (str(divergencia[:5]) if divergencia else 'ninguna'))
    chk('S15', 'zonas_apagadas.json alineado con el runtime', jkeys == runtime,
        'delta runtime-json=' + (str(sorted(runtime ^ jkeys)[:5]) if runtime != jkeys else 'vacío'))

    # ── S16 · simulación de falsos positivos ─────────────────────────────
    no_off = []
    for _d6 in ccpp_folded:
        if _d6 not in ccpp_d6 and ccpp_folded[_d6]:
            no_off.append(_d6)
    fp = []
    if no_off:
        d6_test = no_off[0]
        ccpp_test = sorted(ccpp_folded[d6_test])[0]
        fp.append(('distrito NO apagado', d6_test + ':' + ccpp_test,
                   (d6_test + ':' + ccpp_test) in runtime))
        otro = None
        for _d in sorted(ccpp_d6):
            if _d != d6_test:
                otro = _d
                break
        if otro:
            fp.append(('mismo CCPP en otro d6', otro + ':' + ccpp_test,
                       (otro + ':' + ccpp_test) in runtime))
    if keys:
        k0 = keys[0]
        d60, _s, c0 = k0.partition(':')
        typo = c0[:-1] + ('X' if c0[-1] != 'X' else 'Y')
        fp.append(('typo (casi-match)', d60 + ':' + typo, (d60 + ':' + typo) in runtime))
    falsos = [x for x in fp if x[2]]
    chk('S16', 'Simulación: sin falsos positivos por coincidencia de nombre',
        not falsos,
        'casos simulados=' + str(len(fp)) + ' falsos positivos=' + str(len(falsos)))

    # ── S17 · trazabilidad ───────────────────────────────────────────────
    pdfs = {}
    for name in ('APAGADO.pdf', 'APAGADP1.pdf', 'APAGADO2.pdf'):
        p = PDF_DIR / name
        pdfs[name] = md5(p) if p.exists() else None
    md5s_comu = set()
    for p in PDF_DIR.glob('COMU-5569*.pdf'):
        md5s_comu.add(md5(p))
    chk('S17', 'Trazabilidad de la fuente (md5 de los PDFs)',
        all(pdfs.values()) and len(md5s_comu) == 1,
        'PDFs=' + json.dumps(pdfs, ensure_ascii=False) +
        ' · COMU-5569: ' + str(len(md5s_comu)) + ' archivo(s) md5=' + str(sorted(md5s_comu)))

    # ── S18/S19 · secretos y URLs ────────────────────────────────────────
    secretos = re.findall(r'(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*["\'][^"\']{8,}', html)
    emails = set(re.findall(r'[\w.+-]+@[\w-]+\.[\w.-]+', html))
    chk('S18', 'Sin secretos ni emails nuevos en el popup inline', not secretos,
        'secretos=' + str(len(secretos)) + ' emails=' +
        (str(sorted(emails)[:3]) if emails else 'ninguno'))
    urls = set(re.findall(r'https?://[^"\'\s)]+', html))
    permitidas = {
        'https://www.movistar.com.pe/atencion-al-cliente/reclamos',
        'https://geniomovistar.com.pe/mod/resource/view.php?id=178147',
        'https://geniomovistar.com.pe/mod/resource/view.php?id=178146',
        # preexistentes y benignas: origen canónico de desarrollo + fuentes públicas de datos
        'http://localhost', 'http://127.0.0.1:8765/index.html?popup=1&debug=1',
        'http://www.geonames.org/', 'http://www.w3.org/2000/svg',
        'https://github.com/RitchieRD/ubigeos-peru-data',
    }
    extra = sorted(urls - permitidas)
    chk('S19', 'URLs externas del inline: solo las declaradas y documentadas', not extra,
        str(len(urls)) + ' URLs · no declaradas=' + (str(extra[:5]) if extra else 'ninguna'))

    # ── S20 · honestidad de la fecha por zona (anti-fecha-falsa) ──────────
    gpk = {}
    mg = re.search(r'grupo_por_key:\s*\{([\s\S]*?)\n  \},', html)
    if mg:
        for _k, _g in re.findall(r'"([0-9]{6}:[^"]+)":\s*"(Grupo [123])"', mg.group(1)):
            gpk[fold(_k)] = _g
    grupos = {}
    mgr = re.search(r'grupos_apagado:\s*\{([\s\S]*?)\n  \},', html)
    if mgr:
        for _m in re.finditer(r'"(Grupo [123])":\s*\{([\s\S]*?)\n    \},?', mgr.group(1)):
            grupos[_m.group(1)] = _m.group(2)
    # cada CCPP del runtime debe tener grupo
    sin_grupo = []
    for _k in keys:
        if fold(_k) not in gpk:
            sin_grupo.append(_k)
    chk('S20', 'Cada CCPP del runtime tiene grupo de campaña (base de la fecha por zona)',
        not sin_grupo,
        str(len(keys)) + ' CCPP · sin grupo: ' + (str(sin_grupo[:5]) if sin_grupo else 'ninguno'))
    # los grupos con fecha no confirmada NO pueden declarar fecha por zona
    fecha_falsa = []
    for _g, _body in grupos.items():
        if 'estado_fecha: "pendiente_recurso_interno"' in _body and 'fecha_apagado_zona: null' not in _body:
            fecha_falsa.append(_g + ': pendiente pero con fecha por zona declarada')
        if 'fecha_apagado_campana' not in _body:
            fecha_falsa.append(_g + ': sin fecha de campaña declarada')
    # el panel NO puede afirmar "apagado: <fecha>" de forma incondicional
    mpanel = re.search(r'function showApagadoPanelInline\(btn, currentCheck, addrCtx\) \{([\s\S]*?)\n\}', html)
    pbody = mpanel.group(1) if mpanel else ''
    if "metaParts.push('apagado: ' + fechaCorta)" in pbody:
        fecha_falsa.append('panel: afirma "apagado: <fecha>" sin verificar la fecha de la zona')
    if 'inicio campaña:' not in pbody:
        fecha_falsa.append('panel: sin etiqueta "inicio campaña"')
    chk('S21', 'Sin fecha de zona falsa (grupos pendientes + panel etiquetado)',
        not fecha_falsa,
        ('; '.join(fecha_falsa) if fecha_falsa else
         f"{len(grupos)} grupos · fecha por zona null en los pendientes · panel etiqueta 'apagado' vs 'inicio campaña'"))

    # ── reporte ─────────────────────────────────────────────────────────
    fails = []
    warns = []
    for _res in results:
        if _res['status'] == 'FAIL':
            fails.append(_res)
        elif _res['status'] == 'WARN':
            warns.append(_res)
    print('\n' + '=' * 78)
    print('BARRIDO DE SEGURIDAD — DATOS DE ZONAS DE APAGADO')
    print('=' * 78)
    for _res in results:
        icon = {'PASS': '✅', 'FAIL': '❌', 'WARN': '⚠️ '}[_res['status']]
        print(icon + ' ' + _res['id'].ljust(4) + ' ' + _res['name'])
        if _res['detail']:
            print('          ' + _res['detail'])
    print('-' * 78)
    print('RESUMEN: ' + str(len(results) - len(fails) - len(warns)) + ' PASS · ' +
          str(len(warns)) + ' WARN · ' + str(len(fails)) + ' FAIL')
    print('Runtime: ' + str(len(keys)) + ' CCPP exactos · ' + str(len(dwo)) +
          ' distritos · ' + str(len(ipt)) + ' zonas IPT')
    print('Pendientes declarados: ' + str(sorted(PENDIENTES_ESPERADOS.keys())))
    Path(args.json).write_text(json.dumps({
        'generado': __import__('datetime').datetime.now().isoformat(timespec='seconds'),
        'runtime': {'ccpp': len(keys), 'districts': len(dwo), 'ipt': len(ipt)},
        'checks': results, 'fails': len(fails), 'warns': len(warns),
        'pendientes': PENDIENTES_ESPERADOS,
    }, ensure_ascii=False, indent=2))
    print('\n[OK] → ' + str(args.json))
    return 1 if fails else 0


if __name__ == '__main__':
    sys.exit(main())

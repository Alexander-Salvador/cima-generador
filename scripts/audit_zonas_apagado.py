#!/usr/bin/env python3
"""
audit_zonas_apagado.py — BARRIDO COMPLETO de las zonas de apagado del sistema CIMA
======================================================================================

Reconstruye la verdad desde la FUENTE (los 3 PDFs de "Bases de Apagado" + el manual
OCR) y la cruza contra lo que el popup tiene en RUNTIME, con una única regla de
normalización (fold: sin diacríticos) compartida con el popup.

Verifica:
  1) PÉRDIDAS     — filas de la fuente sin cobertura en el sistema (falsos negativos)
  2) EXTRAS      — claves del sistema sin respaldo en la fuente
  3) PRECISIÓN   — d6 resuelve en UBIGEO · CCPP existe en CCPP_DATA (fold)
  4) EQUIVALENCIAS— typos del PDF corregidos en el sistema · d6 divergente (GeoNames)
  5) HUÉRFANOS   — distritos marcados sin CCPP ni IPT que los respalde
  6) CAMPAÑAS    — con cobertura cargada vs pendientes (COMU-5569)
  7) DATA LAYER  — divergencia entre `data/ubigeo_consolidated.json` y el UBIGEO inline

Salidas (artefactos LOCALES, gitignored por ser derivados de fuente restringida):
  - `data/apagados_tabla_completa.csv` (UTF-8 BOM → Excel) — TODAS las filas, sin pérdida
  - `data/audit_zonas_apagado.json` — detalle programático
  - stdout: reporte

Uso:
    python3 scripts/audit_zonas_apagado.py [--pdf-dir <dir>] [--ocr-json <archivo>]
"""
import argparse
import csv
import json
import re
import sys
import unicodedata
from collections import defaultdict
from difflib import get_close_matches
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
APP = ROOT / 'apps' / 'cima-generador'
DATA = APP / 'data'
HTML = APP / 'index.html'

DEFAULT_PDF_DIR = ROOT / 'Recursos CIMA' / 'Bases y Zonas de Apagado'
DEFAULT_OCR = DATA / 'apagado_ocr_visual.json'


# ── normalización ───────────────────────────────────────────────────────────
def norm(s):
    if not s:
        return ''
    s = str(s).strip().upper()
    s = re.sub(r'\s+', ' ', s)
    return unicodedata.normalize('NFC', s)


def fold(s):
    """Idéntico a zaNormalizeCcpp del popup: NFKD + sin diacríticos + UPPER."""
    if not s:
        return ''
    t = unicodedata.normalize('NFKD', str(s))
    t = ''.join(c for c in t if not unicodedata.combining(c))
    t = t.upper().replace('\u00a0', ' ')
    t = re.sub(r'\s+', ' ', t.strip())
    return unicodedata.normalize('NFC', t)


def nfkd(s):
    return unicodedata.normalize('NFKD', norm(s))


# ── carga ───────────────────────────────────────────────────────────────────
def parse_inline_blocks(html_text):
    ub = json.loads(re.search(r'const UBIGEO = (\{.*?\});\s*\n', html_text, re.S).group(1))
    cc = json.loads(re.search(r'const CCPP_DATA = (\{.*?\});\s*\n', html_text, re.S).group(1))
    za = re.search(r'const ZONAS_APAGADAS = \{([\s\S]*?)\n\};', html_text)
    body = za.group(1)
    keys = re.findall(r'"([0-9]{6}:[^"]+)"', re.search(r'ccpp_keys: new Set\(\[(.*?)\]\)', body, re.S).group(1))
    dwo = re.findall(r'"(\d{6})"', re.search(r'districts_with_off: new Set\(\[(.*?)\]\)', body, re.S).group(1))
    aliases = []
    ma = re.search(r'CCPP_CASERIO_ALIASES\s*=\s*new Set\(\[(.*?)\]\)', html_text, re.S)
    if ma:
        aliases = re.findall(r'"([^"]+)"', ma.group(1))
    return ub, cc, keys, dwo, aliases


def flat_ubigeo(u):
    out = {}
    for d, dv in u.items():
        for p, pv in dv.get('provincias', {}).items():
            for di, info in pv.get('distritos', {}).items():
                d6 = info.get('_ubigeo') if isinstance(info, dict) else None
                if d6:
                    out[d6] = (d.upper(), p.upper(), di.upper())
    return out


def build_index(ubigeo):
    idx = defaultdict(lambda: defaultdict(dict))
    for d6, (d, p, di) in ubigeo.items():
        idx[d][p][di] = d6
    return idx


def resolve(depto, prov, dist, idx, ubigeo):
    D, P, DI = norm(depto), norm(prov), norm(dist)
    provs = idx.get(D)
    if provs is None:
        return None, 'sin_depto', f'departamento "{D}" no existe en UBIGEO'
    if P in provs and DI in provs[P]:
        return provs[P][DI], 'exacto', None
    # fuzzy provincia dentro del depto
    cand_p = get_close_matches(nfkd(P), [nfkd(x) for x in provs], n=1, cutoff=0.72)
    if cand_p:
        real_p = next(k for k in provs if nfkd(k) == cand_p[0])
        dists = provs[real_p]
        cand_d = get_close_matches(nfkd(DI), [nfkd(x) for x in dists], n=1, cutoff=0.72)
        if cand_d:
            real_d = next(k for k in dists if nfkd(k) == cand_d[0])
            return dists[real_d], 'fuzzy', f'prov/dist fuzzy: {P}/{DI} → {real_p}/{real_d}'
    # ¿el distrito existe en UBIGEO con otro nombre de provincia?
    for d6, (d, p, di) in ubigeo.items():
        if d == D and fold(di) == fold(DI):
            return d6, 'fuzzy_provincia', f'provincia "{P}" ≠ "{p}" (distrito existe en {p})'
    return None, 'no_resuelto', f'sin match UBIGEO: {D}/{P}/{DI}'


# ── extracción de fuente ────────────────────────────────────────────────────
def extract_pdf_grupo2(pdf_path):
    import pdfplumber
    rows = []
    with pdfplumber.open(pdf_path) as pdf:
        for pidx, page in enumerate(pdf.pages, 1):
            for table in page.extract_tables():
                last_d = last_p = last_di = ''
                for r in table:
                    c = [(x or '').strip() for x in r]
                    if len(c) < 4:
                        continue
                    if 'DEPARTAMENTO' in c[0].upper() and 'CENTRO' in ' '.join(c).upper():
                        continue
                    if c[0]:
                        last_d = c[0]
                    if c[1]:
                        last_p = c[1]
                    if c[2]:
                        last_di = c[2]
                    if not c[3]:
                        continue
                    rows.append({'depto': last_d, 'prov': last_p, 'dist': last_di,
                                 'ccpp': c[3], 'grupo': 'Grupo 2', 'pagina': pidx})
    return rows


def load_ocr_json(path):
    if not path.exists():
        return [], []
    raw = json.loads(path.read_text())
    g1, g3 = [], []
    for fname, bucket, grupo in (('APAGADP1.pdf', g1, 'Grupo 1'), ('APAGADO2.pdf', g3, 'Grupo 3')):
        for r in raw.get(fname, []):
            bucket.append({'depto': r.get('depto', ''), 'prov': r.get('prov', ''),
                           'dist': r.get('dist', ''), 'ccpp': r.get('ccpp', ''),
                           'grupo': grupo, 'pagina': 1})
    return g1, g3


def load_manual_zones():
    p = DATA / 'zonas_apagadas_manual.json'
    if not p.exists():
        return []
    raw = json.loads(p.read_text())
    return [{'depto': z.get('depto', ''), 'prov': z.get('prov', ''), 'dist': z.get('dist', ''),
             'ccpp': z.get('ccpp', ''), 'grupo': 'manual', 'pagina': 0} for z in raw.get('zonas', [])]


# ── barrido ─────────────────────────────────────────────────────────────────
def main():
    ap = argparse.ArgumentParser(description='Barrido completo de zonas de apagado')
    ap.add_argument('--pdf-dir', default=str(DEFAULT_PDF_DIR))
    ap.add_argument('--ocr-json', default=str(DEFAULT_OCR))
    args = ap.parse_args()

    html_text = HTML.read_text(encoding='utf-8')
    ub, ccpp_data, sys_keys, sys_dwo, aliases = parse_inline_blocks(html_text)
    ubigeo = flat_ubigeo(ub)
    idx = build_index(ubigeo)
    ccpp_folded = {d6: {fold(x) for x in v} for d6, v in ccpp_data.items()}
    sys_folded = {fold(k) for k in sys_keys}
    sys_by_ccpp = defaultdict(list)
    for k in sys_keys:
        d6, _, cc = k.partition(':')
        sys_by_ccpp[fold(cc)].append(d6)

    # fuente
    source = []
    p2 = Path(args.pdf_dir) / 'APAGADO.pdf'
    if p2.exists():
        source += extract_pdf_grupo2(p2)
    g1, g3 = load_ocr_json(Path(args.ocr_json))
    source += g1 + g3
    manual = load_manual_zones()

    rows = []
    seen = set()
    for r in source + manual:
        d6, metodo, err = resolve(r['depto'], r['prov'], r['dist'], idx, ubigeo)
        cc_f = fold(r['ccpp'])
        key = f'{d6}:{cc_f}' if d6 else None
        dup = bool(key and key in seen)
        if key:
            seen.add(key)
        en_caserios = bool(d6 and cc_f in ccpp_folded.get(d6, set()))
        es_distrito = bool(d6 and cc_f == fold(r['dist']))

        # cobertura en el sistema
        if key and key in sys_folded:
            cobertura, nota = 'exacta', ''
        elif cc_f in sys_by_ccpp:
            otros = sys_by_ccpp[cc_f]
            if d6 and any(ubigeo.get(o, ('', '', ''))[:2] == ubigeo.get(d6, ('', '', ''))[:2] for o in otros):
                cobertura = 'equivalente_d6'
            else:
                cobertura = 'equivalente_d6_otro_depto'
            nota = f'mismo CCPP en el sistema con d6 {otros[0]}'
        elif key and any(a.startswith(d6 + ':') for a in aliases):
            cobertura = 'alias_documentado'
            nota = 'variante literal del PDF registrada en CCPP_CASERIO_ALIASES'
        else:
            cobertura = 'NO'
            nota = ''

        if d6 is None:
            estado = 'no_resuelto_ubigeo'
        elif dup:
            estado = 'duplicado_fuente'
        elif cobertura == 'exacta':
            estado = 'ok_exacto'
        elif cobertura.startswith('equivalente'):
            estado = cobertura
        elif cobertura == 'alias_documentado':
            estado = 'alias_documentado'
        elif en_caserios:
            estado = 'ok_caserio_sin_clave'   # existe el caserío pero no hay clave en el sistema
        elif es_distrito:
            estado = 'ambiguo_centro_poblado_igual_distrito'
        else:
            estado = 'ccpp_no_en_caserios'

        rows.append({**r, 'd6': d6, 'key': key, 'metodo': metodo, 'error': err,
                     'estado': estado, 'cobertura': cobertura, 'nota': nota,
                     'en_caserios': en_caserios, 'es_distrito': es_distrito})

    # diferencias sistema ↔ fuente
    src_folded = {r['key'] for r in rows if r['key']}
    sys_folded_keys = {fold(k) for k in sys_keys}
    perdidas = sorted(src_folded - sys_folded_keys)
    extras = sorted(sys_folded_keys - src_folded)
    no_cubiertas = [r for r in rows if r['cobertura'] == 'NO' and r['estado'] != 'duplicado_fuente']

    # IPT
    ipt_rows = []
    pj = DATA / 'zonas_apagadas.json'
    if pj.exists():
        zd = json.loads(pj.read_text())
        ipt_rows = [z for z in zd.get('zones', []) if z.get('comunicado') == 'COMU-5672']
    ipt_d6 = {z.get('ubigeo') for z in ipt_rows}

    # huérfanos
    ccpp_d6 = {k.split(':')[0] for k in sys_keys}
    huerfanos = sorted(set(sys_dwo) - ccpp_d6 - ipt_d6)

    # data layer vs inline
    dl = []
    pcons = DATA / 'ubigeo_consolidated.json'
    if pcons.exists():
        cons = flat_ubigeo(json.loads(pcons.read_text()))
        for d6 in sorted(set(ubigeo) | set(cons)):
            if ubigeo.get(d6) != cons.get(d6):
                dl.append({'d6': d6, 'inline': ubigeo.get(d6), 'data_layer': cons.get(d6)})

    # campañas
    camp = {}
    mc = re.search(r'campanas:\s*\{([\s\S]*?)\n  \},\n  // v3\.10\.3', html_text)
    if mc:
        cb = mc.group(1)
        for cid in re.findall(r'"(COMU-\d+)"', cb):
            mm = re.search(r'"%s":\s*\{([\s\S]*?)\n    \},?' % cid, cb)
            if mm:
                b = mm.group(1)
                def fld(name):
                    m2 = re.search(r'%s:\s*(?:"([^"]*)"|([0-9]+)|(null))' % name, b)
                    if not m2:
                        return None
                    return m2.group(1) if m2.group(1) is not None else (int(m2.group(2)) if m2.group(2) is not None else None)
                camp[cid] = {'estado': fld('estado'), 'fecha_apagado': fld('fecha_apagado'),
                             'fecha_sms_inicio': fld('fecha_sms_inicio'), 'ccpp_count': fld('ccpp_count')}

    # ── reporte ─────────────────────────────────────────────────────────────
    print('\n' + '=' * 74)
    print('BARRIDO COMPLETO — ZONAS DE APAGADO CIMA')
    print('=' * 74)
    print(f'Sistema runtime: {len(sys_keys)} CCPP · {len(sys_dwo)} distritos marcados')
    print(f'Fuente (PDFs):   G2={len([r for r in source if r["grupo"]=="Grupo 2"])} '
          f'G1={len(g1)} G3={len(g3)} | manual={len(manual)} → {len(src_folded)} CCPP únicos')
    print(f'IPT COMU-5672:   {len(ipt_rows)} zonas en {len(ipt_d6)} d6')

    print(f'\n[1] PÉRDIDAS (fuente sin clave en el sistema): {len(perdidas)}')
    for k in perdidas:
        rr = next((r for r in rows if r['key'] == k), None)
        print(f'    - {k}  ← {rr["estado"] if rr else "?"}')
    print(f'\n[2] EXTRAS (sistema sin fila en la fuente): {len(extras)}')
    for k in extras:
        print(f'    - {k}')
    print(f'\n[3] FILAS SIN COBERTURA (requieren decisión): {len(no_cubiertas)}')
    for r in no_cubiertas:
        print(f"    - {r['depto']}/{r['prov']}/{r['dist']} :: {norm(r['ccpp'])} "
              f"→ {r['estado']} ({r['error'] or r['nota']})")
    print(f'\n[4] DUPLICADOS EN LA FUENTE: {len([r for r in rows if r["estado"]=="duplicado_fuente"])}')
    print(f'\n[5] HUÉRFANOS (distrito marcado sin CCPP ni IPT): {len(huerfanos)}')
    for d6 in huerfanos:
        print('    -', d6, ubigeo.get(d6))
    print(f'\n[6] DATA LAYER vs INLINE (ubigeo): {len(dl)} divergencia(s)')
    for x in dl:
        print(f'    - {x["d6"]}: inline={x["inline"]} data_layer={x["data_layer"]}')
    print('\n[7] CAMPAÑAS:')
    for cid, c in camp.items():
        print(f"    - {cid}: {c['estado']} ccpp={c['ccpp_count']} apagado={c['fecha_apagado']} sms={c['fecha_sms_inicio']}")

    # ── CSV completo ───────────────────────────────────────────────────────
    csv_path = DATA / 'apagados_tabla_completa.csv'
    with open(csv_path, 'w', newline='', encoding='utf-8-sig') as f:
        w = csv.writer(f)
        w.writerow(['comunicado', 'campana_estado', 'version', 'fecha_apagado', 'fecha_sms_inicio',
                    'nivel', 'depto', 'provincia', 'distrito', 'centro_poblado_ccpp', 'ubigeo_d6',
                    'key_sistema', 'cobertura_en_sistema', 'estado_validacion', 'metodo_resolucion',
                    'ccpp_en_caserios', 'en_caserios_igual_a_distrito', 'speech_key',
                    'postventa_speech_key', 'origen_pdf', 'grupo', 'pagina', 'notas'])
        for r in sorted(rows, key=lambda x: (x['grupo'], x.get('d6') or '', fold(x['ccpp']))):
            d6 = r['d6']
            nombres = ubigeo.get(d6, (r['depto'], r['prov'], r['dist']))
            notas = [x for x in [r['error'], r['nota']] if x]
            if r['estado'] == 'ambiguo_centro_poblado_igual_distrito':
                notas.append('El PDF escribe el NOMBRE DEL DISTRITO en la columna CENTRO POBLADO: '
                             'semántica "todo el distrito" — decisión CEO')
            if r['cobertura'] == 'NO':
                notas.append('SIN COBERTURA en el popup: no se genera alerta')
            w.writerow([
                'COMU-5514', 'zonas_conocidas', '00003', '2026-09-11', '2026-09-04', 'ccpp',
                nombres[0], nombres[1], nombres[2], norm(r['ccpp']), d6 or '', r['key'] or '',
                r['cobertura'], r['estado'], r['metodo'],
                'SI' if r['en_caserios'] else 'NO', 'SI' if r['es_distrito'] else 'NO',
                'comu5514_visor_unico', 'comu5514_postventa_marcaje',
                f"{r['grupo']}.pdf" if r['grupo'] != 'manual' else 'zonas_apagadas_manual.json',
                r['grupo'], r['pagina'], ' | '.join(notas),
            ])
        # claves del sistema sin fila directa en la fuente
        for k in extras:
            d6, _, cc = k.partition(':')
            nombres = ubigeo.get(d6, ('?', '?', '?'))
            w.writerow(['COMU-5514', 'zonas_conocidas', '00003', '2026-09-11', '2026-09-04',
                        'ccpp', nombres[0], nombres[1], nombres[2], cc, d6, k, 'exacta',
                        'sin_fila_literal_en_fuente', 'n/a',
                        'SI' if fold(cc) in ccpp_folded.get(d6, set()) else 'NO',
                        'SI' if fold(cc) == fold(nombres[2]) else 'NO',
                        'comu5514_visor_unico', 'comu5514_postventa_marcaje',
                        'zonas_apagadas_manual.json', 'sistema', 0,
                        'Clave vigente en el sistema cuya fila literal no se pudo alinear'])
        # IPT
        for z in sorted(ipt_rows, key=lambda x: (x.get('depto', ''), x.get('prov', ''), x.get('dist', ''))):
            w.writerow(['COMU-5672', 'zonas_conocidas', '00001', z.get('fecha_apagado', ''),
                        z.get('fecha_sms_inicio', ''), 'departamento_ipt', z.get('depto', ''),
                        z.get('prov', ''), z.get('dist', ''), '', z.get('ubigeo', ''), '',
                        'SI', 'ipt_amplitud_departamento', 'exacto', 'N/A', 'NO',
                        'ipt_visor_4g', 'ipt_postventa_4g', 'Apagado Zonas IPT.pdf', 'Grupo IPT', 0,
                        f"clientes={z.get('clientes_depto', '')} centros={z.get('centros_depto', '')}"])
        # campañas pendientes
        for cid, c in camp.items():
            if c['estado'] != 'zonas_pendiente':
                continue
            w.writerow([cid, c['estado'], '', c['fecha_apagado'] or '', c['fecha_sms_inicio'] or '',
                        'sin_cobertura_cargada', '', '', '', '', '', '', 'NO', 'pendiente_detalle',
                        '', 'N/A', 'NO', '', '', 'COMU-5569 (comunicado, sin tabla de zonas)',
                        'Campaña', 0,
                        'FALTA: tabla de centros poblados por departamento + fecha apagado + fecha SMS '
                        '(recurso interno geniomovistar id=178147/178146)'])

    report = {
        'generado': __import__('datetime').datetime.now().isoformat(timespec='seconds'),
        'sistema': {'ccpp_keys': len(sys_keys), 'districts_with_off': len(sys_dwo),
                    'ipt_zonas': len(ipt_rows), 'aliases': aliases},
        'fuente': {'filas': len(rows), 'ccpp_unicos': len(src_folded), 'manual_filas': len(manual)},
        'perdidas': perdidas, 'extras': extras,
        'sin_cobertura': [{'depto': r['depto'], 'prov': r['prov'], 'dist': r['dist'],
                           'ccpp': norm(r['ccpp']), 'estado': r['estado'], 'error': r['error']}
                          for r in no_cubiertas],
        'huerfanos': huerfanos,
        'data_layer_divergencias': dl,
        'campanas': camp,
    }
    (DATA / 'audit_zonas_apagado.json').write_text(json.dumps(report, ensure_ascii=False, indent=2))
    print(f'\n[OK] CSV completo → {csv_path} ({csv_path.stat().st_size} bytes)')
    print(f'[OK] JSON → {DATA / "audit_zonas_apagado.json"}')
    return report


if __name__ == '__main__':
    main()

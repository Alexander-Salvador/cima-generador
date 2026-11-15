#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
test_ingest_comunicado_zonas.py
================================
Tests del script `scripts/ingest_comunicado_zonas.py`.

Cubre:
  1. dedupe por md5
  2. idempotencia (2 corridas = 0 cambios)
  3. fuzzy resolve con typo (CHACCHAPOYAS / CHILLIQUIN)
  4. dry-run NO muta index.html
  5. union vs replace sobre una COPIA de index.html en /tmp
  6. fila sin match -> unresolved (con motivo)
  7. CLI: --help y códigos de salida (0 / 1 / 2)

REGLA DURA: este test NUNCA escribe en apps/cima-generador/index.html ni en
apps/cima-generador/data/.  Todos los fixtures viven en /tmp/opencode.

Uso:
    python3 apps/cima-generador/test_ingest_comunicado_zonas.py
"""

import hashlib
import importlib.util
import json
import re
import shutil
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
SCRIPT = HERE / 'scripts' / 'ingest_comunicado_zonas.py'
REAL_HTML = HERE / 'index.html'
REAL_DATA = HERE / 'data'
UBIGEO_PATH = REAL_DATA / 'ubigeo_consolidated.json'
TMPROOT = Path('/tmp/opencode/test_ingest_comunicado_zonas')

# ── estado de tests ─────────────────────────────────────────────────────────
PASSED = []
FAILED = []


def check(name, cond, detail=''):
    if cond:
        PASSED.append(name)
        print('PASS: %s' % name)
    else:
        FAILED.append(name)
        print('FAIL: %s%s' % (name, ('  -> %s' % detail) if detail else ''))
    return bool(cond)


def eq(name, got, expected):
    return check(name, got == expected, 'esperado=%r obtenido=%r' % (expected, got))


# ── helpers ─────────────────────────────────────────────────────────────────
def md5_file(path):
    h = hashlib.md5()
    with open(path, 'rb') as fh:
        for chunk in iter(lambda: fh.read(65536), b''):
            h.update(chunk)
    return h.hexdigest()


def load_module():
    spec = importlib.util.spec_from_file_location('ingest_comunicado_zonas', str(SCRIPT))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def run_cli(args):
    proc = subprocess.run([sys.executable, str(SCRIPT)] + list(args),
                          capture_output=True, text=True)
    return proc


def workspace(name):
    ws = TMPROOT / name
    if ws.exists():
        shutil.rmtree(ws)
    (ws / 'data').mkdir(parents=True, exist_ok=True)
    return ws


def write_csv(path, rows, sep=';', header=('DEPARTAMENTO', 'PROVINCIA', 'DISTRITO',
                                          'CENTRO POBLADO')):
    lines = [sep.join(header)]
    for r in rows:
        lines.append(sep.join(str(x) for x in r))
    path.write_text('\n'.join(lines) + '\n', encoding='utf-8')
    return path


def copy_real_html(ws, name='index_copy.html'):
    dest = ws / name
    shutil.copy2(str(REAL_HTML), str(dest))
    return dest


def html_ccpp_keys(html):
    m = re.search(r'ccpp_keys:\s*new Set\(\[(.*?)\]\)', html, re.S)
    if not m:
        return None
    return re.findall(r'"([^"]*)"', m.group(1))


def html_total_zones(html):
    m = re.search(r'total_zones:\s*(\d+)', html)
    return int(m.group(1)) if m else None


def read_json(path):
    return json.loads(Path(path).read_text(encoding='utf-8'))


# ── fixtures ────────────────────────────────────────────────────────────────
#  2 filas válidas (una con typos: CHACCHAPOYAS / CHILLIQUIN)
#  1 fila duplicada  (dedupe por fila lógica)
#  1 fila inválida   (va a unresolved)
CSV_ROWS = [
    ('AMAZONAS', 'CHACCHAPOYAS', 'CHILLIQUIN', 'VITUYA'),
    ('AMAZONAS', 'CHACHAPOYAS', 'OLLEROS', 'OLLEROS NUEVO'),
    ('AMAZONAS', 'CHACHAPOYAS', 'OLLEROS', 'OLLEROS NUEVO'),
    ('ZZZ_DEPTO', 'ZZZ_PROV', 'ZZZ_DIST', 'CASERIO INEXISTENTE'),
]


def cli_args(ws, source, extra=()):
    return ['--comunicado', 'COMU-5569',
            '--source', str(source),
            '--data-dir', str(ws / 'data'),
            '--html-path', str(ws / 'index_copy.html'),
            '--ubigeo-path', str(UBIGEO_PATH)] + list(extra)


# ═══════════════════════════════════════════════════════════════════════════
# TESTS
# ═══════════════════════════════════════════════════════════════════════════

def test_1_dedupe_por_md5():
    ws = workspace('t1_dedupe')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    html = copy_real_html(ws)

    p1 = run_cli(cli_args(ws, csv, ['--apply-inline']))
    if not check('T1.1 primera corrida exit 0', p1.returncode == 0,
                 p1.stderr.strip()[:400]):
        return
    dedupe = ws / 'data' / '.ingest_dedupe_COMU-5569.json'
    if not check('T1.2 se crea el estado de dedupe', dedupe.exists()):
        return
    st = read_json(dedupe)
    eq('T1.3 md5 del estado == md5 de la fuente',
       st.get('md5'), md5_file(csv))
    check('T1.4 el estado trae filas/ccpp_agregados',
          'filas' in st and 'ccpp_agregados' in st, repr(sorted(st.keys())))
    eq('T1.5 ccpp_agregados = 2 (fila duplicada colapsada)',
       st.get('ccpp_agregados'), 2)
    check('T1.6 la 1ª corrida reporta "fuente nueva"',
          'fuente nueva' in p1.stdout)

    # 2ª corrida con la MISMA fuente -> dedupe por md5
    p2 = run_cli(cli_args(ws, csv, ['--apply-inline']))
    eq('T1.7 2ª corrida exit 0', p2.returncode, 0)
    check('T1.8 reporta "FUENTE YA INGERIDA"',
          'FUENTE YA INGERIDA' in p2.stdout, p2.stdout[-300:])
    st2 = read_json(dedupe)
    eq('T1.9 el md5 del estado no cambia', st2.get('md5'), st.get('md5'))
    eq('T1.10 ccpp_agregados no crece (0 duplicados)',
       st2.get('ccpp_agregados'), st.get('ccpp_agregados'))

    # fuente MODIFICADA -> md5 distinto -> se vuelve a procesar
    csv2 = write_csv(ws / 'zonas2.csv', CSV_ROWS[:2])
    p3 = run_cli(cli_args(ws, csv2, ['--apply-inline']))
    eq('T1.11 fuente nueva exit 0', p3.returncode, 0)
    check('T1.12 fuente con md5 distinto NO se considera ya ingerida',
          'fuente nueva' in p3.stdout)
    st3 = read_json(dedupe)
    check('T1.13 el estado guarda el md5 nuevo', st3.get('md5') != st.get('md5'))
    _ = html


def test_2_idempotencia():
    ws = workspace('t2_idem')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    copy_real_html(ws)

    p1 = run_cli(cli_args(ws, csv, ['--apply-inline']))
    eq('T2.1 corrida 1 exit 0', p1.returncode, 0)
    html1 = (ws / 'index_copy.html').read_text(encoding='utf-8')
    json1 = read_json(ws / 'data' / 'zonas_apagadas_COMU-5569.json')

    p2 = run_cli(cli_args(ws, csv, ['--apply-inline']))
    eq('T2.2 corrida 2 exit 0', p2.returncode, 0)
    html2 = (ws / 'index_copy.html').read_text(encoding='utf-8')
    json2 = read_json(ws / 'data' / 'zonas_apagadas_COMU-5569.json')

    check('T2.3 index.html idéntico tras la 2ª corrida', html1 == html2)
    eq('T2.4 ccpp_keys idénticos', json2['ccpp_keys'], json1['ccpp_keys'])
    eq('T2.5 zones idénticas', json2['zones'], json1['zones'])
    eq('T2.6 unresolved idénticos', json2['unresolved'], json1['unresolved'])
    eq('T2.7 total_filas idéntico', json2['total_filas'], json1['total_filas'])

    # tercera corrida, por si acaso
    run_cli(cli_args(ws, csv, ['--apply-inline']))
    html3 = (ws / 'index_copy.html').read_text(encoding='utf-8')
    check('T2.8 index.html idéntico tras la 3ª corrida', html1 == html3)


def test_3_fuzzy_resolve():
    mod = load_module()
    idx = mod.load_ubigeo(UBIGEO_PATH)

    r_exacto, _ = idx.resolve('AMAZONAS', 'CHACHAPOYAS', 'CHILIQUIN')
    check('T3.1 resolución exacta existe', r_exacto is not None)
    eq('T3.2 exacto -> d6 010105', r_exacto['ubigeo'], '010105')
    eq('T3.3 exacto -> metodo exacto', r_exacto['metodo'], 'exacto')

    r_prov, _ = idx.resolve('AMAZONAS', 'CHACCHAPOYAS', 'CHILIQUIN')
    check('T3.4 fuzzy de provincia resuelve', r_prov is not None)
    if r_prov:
        eq('T3.5 fuzzy provincia -> d6 010105', r_prov['ubigeo'], '010105')
        check('T3.6 fuzzy provincia reportada en metodo',
              'fuzzy' in r_prov['metodo'], r_prov['metodo'])

    r_dist, _ = idx.resolve('AMAZONAS', 'CHACHAPOYAS', 'CHILLIQUIN')
    check('T3.7 fuzzy de distrito resuelve', r_dist is not None)
    if r_dist:
        eq('T3.8 fuzzy distrito -> d6 010105', r_dist['ubigeo'], '010105')
        check('T3.9 fuzzy distrito reportada en metodo',
              'fuzzy' in r_dist['metodo'], r_dist['metodo'])

    # extremo: end-to-end por CLI con la fila con typos
    ws = workspace('t3_fuzzy')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    copy_real_html(ws)
    p = run_cli(cli_args(ws, csv, ['--apply-inline']))
    eq('T3.10 CLI con typos exit 0', p.returncode, 0)
    camp = read_json(ws / 'data' / 'zonas_apagadas_COMU-5569.json')
    check('T3.11 la fila con typo se resolvió',
          '010105:VITUYA' in camp['ccpp_keys'], repr(camp['ccpp_keys']))
    z = [z for z in camp['zones'] if z['ccpp'] == 'VITUYA']
    check('T3.12 la zona con typo tiene metodo fuzzy',
          bool(z) and 'fuzzy' in z[0].get('metodo_resolucion', ''),
          repr(z[0] if z else None))


def test_4_dry_run_no_muta():
    ws = workspace('t4_dry')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    html = copy_real_html(ws)
    antes = md5_file(html)

    p = run_cli(cli_args(ws, csv))          # sin --apply-inline => dry-run
    eq('T4.1 dry-run exit 0', p.returncode, 0)
    eq('T4.2 index.html intacto', md5_file(html), antes)
    check('T4.3 el reporte dice DRY-RUN', 'DRY-RUN' in p.stdout, p.stdout[-200:])
    check('T4.4 no se escribió el JSON de campaña',
          not (ws / 'data' / 'zonas_apagadas_COMU-5569.json').exists())
    check('T4.5 no se escribió el estado de dedupe',
          not (ws / 'data' / '.ingest_dedupe_COMU-5569.json').exists())
    check('T4.6 no se creó backup',
          not (ws / '.backup-pre-ingest').exists())

    # --dry-run explícito tampoco escribe
    p2 = run_cli(cli_args(ws, csv, ['--dry-run']))
    eq('T4.7 --dry-run exit 0', p2.returncode, 0)
    eq('T4.8 index.html intacto tras --dry-run', md5_file(html), antes)


def test_5_union_vs_replace():
    # ── UNION ────────────────────────────────────────────────────────────
    ws = workspace('t5_union')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    html = copy_real_html(ws)
    keys_antes = set(html_ccpp_keys(html.read_text(encoding='utf-8')))

    p = run_cli(cli_args(ws, csv, ['--apply-inline', '--modo', 'union']))
    eq('T5.1 union exit 0', p.returncode, 0)
    texto = html.read_text(encoding='utf-8')
    keys_despues = set(html_ccpp_keys(texto))
    camp = read_json(ws / 'data' / 'zonas_apagadas_COMU-5569.json')
    nuevas = set(camp['ccpp_keys'])

    check('T5.2 union: las claves nuevas están',
          nuevas.issubset(keys_despues), repr(sorted(nuevas - keys_despues)))
    check('T5.3 union: las claves previas siguen',
          keys_antes.issubset(keys_despues),
          repr(sorted(keys_antes - keys_despues))[:300])
    eq('T5.4 union: total = previas + nuevas',
       len(keys_despues), len(keys_antes | nuevas))
    eq('T5.5 union: total_zones == ccpp_keys.size',
       html_total_zones(texto), len(keys_despues))

    # ── REPLACE ──────────────────────────────────────────────────────────
    ws2 = workspace('t5_replace')
    csv2 = write_csv(ws2 / 'zonas.csv', CSV_ROWS)
    html2 = copy_real_html(ws2)
    p2 = run_cli(cli_args(ws2, csv2, ['--apply-inline', '--modo', 'replace']))
    eq('T5.6 replace exit 0', p2.returncode, 0)
    texto2 = html2.read_text(encoding='utf-8')
    keys2 = set(html_ccpp_keys(texto2))
    camp2 = read_json(ws2 / 'data' / 'zonas_apagadas_COMU-5569.json')
    eq('T5.7 replace: ccpp_keys == ccpp_keys de la campaña',
       keys2, set(camp2['ccpp_keys']))
    eq('T5.8 replace: total_zones == ccpp_keys.size',
       html_total_zones(texto2), len(keys2))
    check('T5.9 replace: se quitaron las claves previas',
          len(keys2) < len(set(html_ccpp_keys(REAL_HTML.read_text(encoding='utf-8')))))

    # JS válido en ambos casos
    node = shutil.which('node')
    if node:
        for etiqueta, archivo in (('union', html), ('replace', html2)):
            js = ws / ('check_%s.js' % etiqueta)
            bloques = re.findall(r'<script(?![^>]*src=)[^>]*>(.*?)</script>',
                                  archivo.read_text(encoding='utf-8'), re.S)
            js.write_text('\n;\n'.join(bloques), encoding='utf-8')
            chk = subprocess.run([node, '--check', str(js)],
                                 capture_output=True, text=True)
            check('T5.10 JS sintácticamente válido (%s)' % etiqueta,
                  chk.returncode == 0, chk.stderr.strip()[:300])


def test_6_unresolved():
    ws = workspace('t6_unresolved')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    copy_real_html(ws)
    p = run_cli(cli_args(ws, csv, ['--apply-inline']))
    eq('T6.1 exit 0', p.returncode, 0)

    camp = read_json(ws / 'data' / 'zonas_apagadas_COMU-5569.json')
    unres = camp['unresolved']
    check('T6.2 hay al menos 1 unresolved', len(unres) >= 1)
    ccpps_unres = {u['ccpp'] for u in unres}
    check('T6.3 la fila inválida quedó en unresolved',
          'CASERIO INEXISTENTE' in ccpps_unres, repr(ccpps_unres))
    check('T6.4 cada unresolved tiene motivo',
          all(u.get('motivo') for u in unres), repr(unres))
    check('T6.5 la fila inválida NO está en zones',
          all(z['ccpp'] != 'CASERIO INEXISTENTE' for z in camp['zones']))
    check('T6.6 la fila inválida NO está en ccpp_keys',
          all('INEXISTENTE' not in k for k in camp['ccpp_keys']))
    eq('T6.7 total_unresolved coincide', camp['total_unresolved'], len(unres))
    eq('T6.8 total_resueltas coincide', camp['total_resueltas'], len(camp['zones']))
    check('T6.9 el reporte lista los unresolved',
          'UNRESOLVED' in p.stdout and 'CASERIO INEXISTENTE' in p.stdout)


def test_7_cli():
    p = run_cli(['--help'])
    eq('T7.1 --help exit 0', p.returncode, 0)
    check('T7.2 --help documenta --apply-inline', '--apply-inline' in p.stdout)
    check('T7.3 --help documenta --modo', '--modo' in p.stdout)
    check('T7.4 --help documenta --campana-json', '--campana-json' in p.stdout)

    ws = workspace('t7_cli')
    csv = write_csv(ws / 'zonas.csv', CSV_ROWS)
    copy_real_html(ws)

    p = run_cli(cli_args(ws, ws / 'no_existe.csv'))
    eq('T7.5 fuente inexistente -> exit 1', p.returncode, 1)
    check('T7.6 error en español', 'ERROR' in p.stderr, p.stderr[:200])

    p = run_cli(['--comunicado', 'COMU-5569'])          # falta --source
    check('T7.7 argumentos faltantes -> exit != 0', p.returncode != 0)

    p = run_cli(cli_args(ws, csv, ['--dry-run', '--apply-inline']))
    check('T7.8 --dry-run + --apply-inline -> exit != 0', p.returncode != 0)

    # patrón inexistente en el HTML -> exit 2 y NO escribe
    roto = ws / 'roto.html'
    roto.write_text('<html><body>sin bloques JS</body></html>', encoding='utf-8')
    antes = md5_file(roto)
    args = cli_args(ws, csv, ['--apply-inline'])
    args[args.index('--html-path') + 1] = str(roto)
    p = run_cli(args)
    eq('T7.9 patrón no encontrado -> exit 2', p.returncode, 2)
    check('T7.10 explica que aborta sin escribir',
          'ABORTA' in p.stderr or 'ABORTA' in p.stdout, (p.stderr + p.stdout)[:300])
    eq('T7.11 el HTML roto no se modificó', md5_file(roto), antes)
    check('T7.12 no se escribió la campaña',
          not (ws / 'data' / 'zonas_apagadas_COMU-5569.json').exists())

    # formato no soportado
    malo = ws / 'zonas.txt'
    malo.write_text('nada', encoding='utf-8')
    p = run_cli(cli_args(ws, malo))
    eq('T7.13 formato no soportado -> exit 1', p.returncode, 1)


def test_8_no_toca_el_repo():
    """El test completo no debe haber modificado index.html ni data/ del repo."""
    repo_html = md5_file(REAL_HTML)
    check('T8.1 index.html del repo sin cambios por el test',
          repo_html == REPO_HTML_MD5_INICIO,
          'antes=%s ahora=%s' % (REPO_HTML_MD5_INICIO, repo_html))
    intrusos = sorted(p.name for p in (REAL_DATA).glob('.ingest_dedupe_*'))
    check('T8.2 no hay .ingest_dedupe_* en data/ del repo',
          not intrusos, repr(intrusos))


# ═══════════════════════════════════════════════════════════════════════════
# MAIN
# ═══════════════════════════════════════════════════════════════════════════

REPO_HTML_MD5_INICIO = md5_file(REAL_HTML) if REAL_HTML.exists() else ''


def main():
    global REPO_HTML_MD5_INICIO
    print('═' * 70)
    print('TEST ingest_comunicado_zonas.py')
    print('script : %s' % SCRIPT)
    print('fixtures: %s' % TMPROOT)
    print('═' * 70)
    if not SCRIPT.exists():
        print('FAIL: no existe el script %s' % SCRIPT)
        return 1
    if not REAL_HTML.exists():
        print('FAIL: no existe %s' % REAL_HTML)
        return 1

    test_1_dedupe_por_md5()
    test_2_idempotencia()
    test_3_fuzzy_resolve()
    test_4_dry_run_no_muta()
    test_5_union_vs_replace()
    test_6_unresolved()
    test_7_cli()
    test_8_no_toca_el_repo()

    print('═' * 70)
    print('PASS: %d   FAIL: %d' % (len(PASSED), len(FAILED)))
    if FAILED:
        for f in FAILED:
            print('  · FAILED: %s' % f)
        print('═' * 70)
        return 1
    print('OK — todos los tests pasaron.')
    print('═' * 70)
    return 0


if __name__ == '__main__':
    sys.exit(main())

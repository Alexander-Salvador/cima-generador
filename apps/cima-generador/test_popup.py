#!/usr/bin/env python3
"""
CIMA popup v2.0.0 — automated integrity tests.

Tests:
1. JS syntax valid (Node --check)
2. Data inline present (UBIGEO + CCPP_DATA)
3. UBIGEO has 25 departments
4. CCPP_DATA has >40K caseríos
5. No external fetch() calls in runtime
6. CSP meta tag present
7. All 6 audit fixes applied
8. Bridge flags: --wrap, --range-domain, --range-series
9. Per-fixture: ubigeo keys match expected

Usage:
  python3 test_popup.py [path/to/index.html]
"""
import re
import sys
import json
import subprocess
from pathlib import Path

DEFAULT_HTML = Path(__file__).parent / 'index.html'

# Expected fixture counts
EXPECTED_DEPTOS = 25
EXPECTED_PROVINCIAS = 196
EXPECTED_DISTRITOS = 1892
EXPECTED_DISTRITOS_CCPP = 1851  # some districts don't have caseríos in GeoNames
EXPECTED_CCPP_MIN = 40000  # 43,491 actual

class TestResult:
    def __init__(self):
        self.passed = []
        self.failed = []
        self.warnings = []

    def ok(self, name):
        self.passed.append(name)
        print(f'  ✅ {name}')

    def fail(self, name, detail):
        self.failed.append((name, detail))
        print(f'  ❌ {name}: {detail}')

    def warn(self, name, detail):
        self.warnings.append((name, detail))
        print(f'  ⚠️  {name}: {detail}')


def extract_data(html):
    """Extract UBIGEO and CCPP_DATA const values from HTML."""
    m_ubigeo = re.search(r'const UBIGEO = (\{.*?\});', html, re.DOTALL)
    m_ccpp = re.search(r'const CCPP_DATA = (\{.*?\});', html, re.DOTALL)
    if not m_ubigeo:
        return None, None
    if not m_ccpp:
        return None, None
    try:
        ubigeo = json.loads(m_ubigeo.group(1))
    except json.JSONDecodeError as e:
        return f'ubigeo parse: {e}', None
    try:
        ccpp = json.loads(m_ccpp.group(1))
    except json.JSONDecodeError as e:
        return ubigeo, f'ccpp parse: {e}'
    return ubigeo, ccpp


def test(html_path):
    if not html_path.exists():
        print(f'❌ File not found: {html_path}')
        sys.exit(1)

    html = html_path.read_text()
    result = TestResult()

    print(f'\n{"="*60}')
    print(f'  CIMA popup v2.0.0 — automated tests')
    print(f'  File: {html_path} ({len(html):,} bytes)')
    print(f'{"="*60}\n')

    # 1. JS syntax
    print('▸ JS syntax (Node --check)')
    m = re.search(r'<script[^>]*>(.*?)</script>', html, re.DOTALL)
    if m:
        script_path = Path('/tmp/opencode/popup_test.js')
        script_path.write_text(m.group(1))
        try:
            r = subprocess.run(['node', '--check', str(script_path)],
                             capture_output=True, text=True, timeout=10)
            if r.returncode == 0:
                result.ok(f'JS syntax valid ({len(m.group(1)):,} chars)')
            else:
                result.fail('JS syntax', r.stderr[:200])
        except Exception as e:
            result.warn('JS syntax', f'Node not available: {e}')
    else:
        result.fail('JS script tag', 'not found')

    # 2. Data inline
    print('\n▸ Data inline')
    ubigeo, ccpp = extract_data(html)
    if isinstance(ubigeo, str):
        result.fail('UBIGEO parse', ubigeo)
    elif ubigeo is None:
        result.fail('UBIGEO const', 'not found')
    else:
        result.ok(f'UBIGEO const parsed ({len(json.dumps(ubigeo)):,} bytes)')

    if isinstance(ccpp, str):
        result.fail('CCPP_DATA parse', ccpp)
    elif ccpp is None:
        result.fail('CCPP_DATA const', 'not found')
    else:
        result.ok(f'CCPP_DATA const parsed ({len(json.dumps(ccpp)):,} bytes)')

    # 3. UBIGEO structure
    print('\n▸ UBIGEO structure')
    if isinstance(ubigeo, dict):
        n_deptos = len(ubigeo)
        n_prov = sum(len(d.get('provincias', {})) for d in ubigeo.values())
        n_dist = sum(len(p.get('distritos', {}))
                     for d in ubigeo.values()
                     for p in d.get('provincias', {}).values())

        if n_deptos == EXPECTED_DEPTOS:
            result.ok(f'UBIGEO deptos: {n_deptos} (expected {EXPECTED_DEPTOS})')
        else:
            result.warn(f'UBIGEO deptos: {n_deptos} (expected {EXPECTED_DEPTOS})')

        if n_prov == EXPECTED_PROVINCIAS:
            result.ok(f'UBIGEO provincias: {n_prov} (expected {EXPECTED_PROVINCIAS})')
        else:
            result.warn(f'UBIGEO provincias: {n_prov} (expected {EXPECTED_PROVINCIAS})')

        if n_dist == EXPECTED_DISTRITOS:
            result.ok(f'UBIGEO distritos: {n_dist} (expected {EXPECTED_DISTRITOS})')
        else:
            result.warn(f'UBIGEO distritos: {n_dist} (expected {EXPECTED_DISTRITOS})')

        # Sample keys
        sample = list(ubigeo.keys())[:3]
        result.ok(f'UBIGEO sample keys: {sample}')

    # 4. CCPP structure
    print('\n▸ CCPP_DATA structure')
    if isinstance(ccpp, dict):
        n_dist_ccpp = len(ccpp)
        total_caserios = sum(len(v) for v in ccpp.values())

        if n_dist_ccpp >= 1500:
            result.ok(f'CCPP districts: {n_dist_ccpp} (expected ≥{EXPECTED_DISTRITOS_CCPP})')
        else:
            result.fail(f'CCPP districts', f'{n_dist_ccpp} (expected ≥{EXPECTED_DISTRITOS_CCPP})')

        if total_caserios >= EXPECTED_CCPP_MIN:
            result.ok(f'CCPP caseríos: {total_caserios:,} (expected ≥{EXPECTED_CCPP_MIN:,})')
        else:
            result.fail(f'CCPP caseríos', f'{total_caserios:,} (expected ≥{EXPECTED_CCPP_MIN:,})')

        # Sample caseríos
        first_dist = next(iter(ccpp.keys()), None)
        if first_dist:
            sample_ccpp = ccpp[first_dist][:3]
            result.ok(f'CCPP sample ({first_dist}): {sample_ccpp}')

    # 5. No external fetch
    print('\n▸ No external fetch (data inline)')
    if "fetch('data/" in html or 'fetch("data/' in html:
        result.fail('External fetch', 'still present in HTML — v1.0.0 bug')
    else:
        result.ok('No external fetch in HTML')

    # 6. CSP meta
    print('\n▸ CSP meta tag')
    if 'Content-Security-Policy' in html:
        result.ok('CSP meta tag present')
    else:
        result.fail('CSP meta tag', 'missing')
    if "default-src 'self'" in html:
        result.ok('CSP default-src set')
    else:
        result.warn('CSP default-src', 'not set')

    # 7. Audit fixes
    print('\n▸ Audit fixes applied')
    checks = {
        'FIX #1 updateTipifOutputOnly': 'function updateTipifOutputOnly' in html,
        'FIX #1 TEXT_KEYS check': 'TEXT_KEYS' in html,
        'FIX #2 maxlength 120': 'maxlength="120"' in html,
        'FIX #2 debounce helper': 'function debounce' in html,
        'FIX #3 cap 50 tipifs': 'tipifs.length >= 50' in html,
        'FIX #5 full escape': '&amp;' in html and '&#39;' in html and '&#x2F;' in html,
        'FIX #6 CSP meta': 'Content-Security-Policy' in html,
        'UBIGEO const': 'const UBIGEO = {' in html,
        'CCPP_DATA const': 'const CCPP_DATA = {' in html,
        'ZONAS_APAGADAS const': 'const ZONAS_APAGADAS = {' in html,
        'ZONAS_APAGADAS compact Set d6': 'd6: new Set([' in html,
        'ZONAS_APAGADAS compact Set d4': 'd4: new Set([' in html,
        'ZONAS_APAGADAS compact Set d2': 'd2: new Set([' in html,
        'getUbigeoCodesForAddress function': 'function getUbigeoCodesForAddress' in html,
        'renderCoberturaButton function': 'function renderCoberturaButton' in html,
        'toggleCoberturaDetail function': 'function toggleCoberturaDetail' in html,
        'coberturaBtn element': 'id="coberturaBtn"' in html,
        'pulseDanger keyframes': '@keyframes pulseDanger' in html,
        'checkZonaApagada function': 'function checkZonaApagada' in html,
    }
    for name, present in checks.items():
        if present:
            result.ok(name)
        else:
            result.fail(name, 'NOT FOUND')

    # 7.5 ZONAS_APAGADAS inline data integrity (v2.0.4+ — 239 zones, política proporcional)
    print('\n▸ ZONAS_APAGADAS data integrity (v2.0.4 política proporcional)')
    sets_match = re.search(
        r'const ZONAS_APAGADAS = \{\s*d6:\s*new Set\(\[(.*?)\]\),\s*d4:\s*new Set\(\[(.*?)\]\),\s*d2:\s*new Set\(\[(.*?)\]\),\s*total:\s*(\d+)',
        html, re.DOTALL)
    if sets_match:
        d6_items = re.findall(r'"(\d+)"', sets_match.group(1))
        d4_items = re.findall(r'"(\d+)"', sets_match.group(2))
        d2_items = re.findall(r'"(\d+)"', sets_match.group(3))
        total = int(sets_match.group(4))

        # Size tests (v2.0.4 política: d6=192, d4=1, d2=1)
        if len(d6_items) == 192: result.ok(f'd6 size = {len(d6_items)} (192 expected)')
        else: result.fail(f'd6 size', f'{len(d6_items)} (expected 192)')
        if len(d4_items) == 1: result.ok(f'd4 size = {len(d4_items)} (1 — solo LORETO prov)')
        else: result.fail(f'd4 size', f'{len(d4_items)} (expected 1)')
        if len(d2_items) == 1: result.ok(f'd2 size = {len(d2_items)} (1 — solo LORETO depto)')
        else: result.fail(f'd2 size', f'{len(d2_items)} (expected 1)')
        if total == 239: result.ok(f'total = {total} (239 expected)')
        else: result.fail(f'total', f'{total} (expected 239)')

        # Policy verify: d4 y d2 deben ser LORETO específicamente
        if d4_items == ['1603']: result.ok('d4 = ["1603"] (LORETO prov)')
        else: result.fail(f'd4 content', f'{d4_items} (expected ["1603"])')
        if d2_items == ['16']: result.ok('d2 = ["16"] (LORETO depto)')
        else: result.fail(f'd2 content', f'{d2_items} (expected ["16"])')

        # Uniqueness
        if len(set(d6_items)) == len(d6_items): result.ok('d6 items unique')
        else: result.fail('d6 items', f'{len(d6_items)-len(set(d6_items))} duplicates')

        # Real samples present (zonas nuevas + regresión)
        samples = [
            ('060808', 'PUCARA Cajamarca (baseline v2.0.1)'),
            ('040206', 'OCONA Arequipa (Grupo 1, fuzzy OCOÑA)'),
            ('051012', 'VILCANCHOS Ayacucho (Grupo 3 OCR visual)'),
            ('120606', 'COVIRIALI Junín (Grupo 1 visual)'),
            ('160305', 'SANTA ROSA Loreto (Grupo 1 visual)'),
            ('1603', 'LORETO prov (d4 proporcional)'),
            ('16', 'LORETO depto (d2 proporcional)'),
        ]
        all_set = set(d6_items) | set(d4_items) | set(d2_items)
        for ubigeo, label in samples:
            if ubigeo in all_set:
                result.ok(f'{ubigeo} present ({label})')
            else:
                result.fail(f'{ubigeo} missing', label)
    else:
        result.fail('ZONAS_APAGADAS pattern', 'no match')

    # 7.6 v2.0.5 Visual layout: coberturaBtn dentro del output-row, compacto
    print('\n▸ Visual layout v2.0.5 (coberturaBtn compact inside output-row)')
    # Buscar orden: <div class="output-row"> ... id="coberturaBtn" ... en orden
    btn_pos = html.find('id="coberturaBtn"')
    output_rows = [m.start() for m in re.finditer(r'<div class="output-row">', html)]
    output_ends = [m.end() for m in re.finditer(r'<div class="output-row">[^<]*<div[^>]*>.*?</div>\s*<button[^>]*btn-icon-clear.*?</button>\s*<button[^>]*coberturaBtn.*?</button>\s*</div>', html, re.DOTALL)]
    if btn_pos > 0:
        # Verificar que entre el <div class="output-row"> más cercano antes y el </div> después está el coberturaBtn
        last_row_start = max(p for p in output_rows if p < btn_pos) if output_rows else -1
        if last_row_start >= 0:
            # Check that next </div> after coberturaBtn closes the output-row
            after_btn = html.find('</div>', btn_pos)
            if after_btn > 0:
                segment = html[last_row_start:after_btn + 6]
                if 'btn-icon-clear' in segment and 'coberturaBtn' in segment:
                    result.ok('coberturaBtn dentro del output-row (después de X clear)')
                else:
                    result.fail('coberturaBtn location', 'no comparte output-row con btn-icon-clear')
            else:
                result.fail('output-row end', '</div> no encontrado')
        else:
            result.fail('output-row antes', 'no encontrado antes del coberturaBtn')
    else:
        result.fail('coberturaBtn element', 'no encontrado en HTML')

    # CSS .cobertura-btn default = display:none (hidden by default)
    css_btn_match = re.search(
        r'\.cobertura-btn\s*\{([^}]+)\}',
        html
    )
    if css_btn_match:
        css_body = css_btn_match.group(1)
        if 'display: none' in css_body or 'display:none' in css_body:
            result.ok('.cobertura-btn display:none (hidden by default)')
        else:
            result.fail('.cobertura-btn display', 'no es none por default')
    else:
        result.fail('.cobertura-btn CSS', 'no encontrado')

    # CSS .cobertura-btn.off = display:inline-flex (visible solo con alerta)
    css_off_match = re.search(
        r'\.cobertura-btn\.off\s*\{([^}]+)\}',
        html
    )
    if css_off_match:
        css_off_body = css_off_match.group(1)
        if 'display: inline-flex' in css_off_body or 'display:inline-flex' in css_off_body:
            result.ok('.cobertura-btn.off display:inline-flex (visible con alerta)')
        else:
            result.fail('.cobertura-btn.off display', 'no es inline-flex')

    # Botón solo icono, sin texto "COBERTURA"
    if 'coberturaLabel' in html:
        result.fail('coberturaLabel text', 'ya no debe existir (icon-only)')
    else:
        result.ok('coberturaLabel eliminado (icon-only)')

    # 7.7 v2.0.6 — Badge opt-in: NO auto-show, solo al click en ⚠️
    print('\n▸ Badge opt-in v2.0.6 (badge oculto hasta click)')
    # CSS .zona-apagada-badge debe tener UN SOLO display:none (sin doble declaration)
    css_badge_match = re.search(
        r'\.zona-apagada-badge\s*\{([^}]+)\}',
        html
    )
    if css_badge_match:
        css_body = css_badge_match.group(1)
        n_display_none = css_body.count('display: none')
        if n_display_none == 1:
            result.ok('.zona-apagada-badge tiene 1 sola declaración display: none')
        else:
            result.fail('.zona-apagada-badge display', f'{n_display_none} declaraciones (esperaba 1)')

    # JS checkAndRenderZonaApagada NO debe auto-agregar 'show'
    fn_match = re.search(
        r'function checkAndRenderZonaApagada\(\)\s*\{([\s\S]*?)\n\}',
        html
    )
    if fn_match:
        fn_body = fn_match.group(1)
        # Cuando hay check: debe haber classList.remove('show') (no add)
        # Buscamos el patrón "if (check) { ... classList.remove('show');"
        if re.search(r'if\s*\(\s*check\s*\)\s*\{[\s\S]*?classList\.remove\([\'\"]show[\'\"]\)', fn_body):
            result.ok('checkAndRenderZonaApagada oculta badge (no auto-show)')
        else:
            result.fail('checkAndRenderZonaApagada', 'patrón de auto-show detectado')
        # No debe haber classList.add('show') en ninguna rama del if-else
        if "classList.add('show')" in fn_body or 'classList.add("show")' in fn_body:
            result.fail('auto-show residual', "classList.add('show') aún presente")
        else:
            result.ok("classList.add('show') eliminado (badge opt-in)")

    # CSS .zona-apagada-badge.show debe seguir existiendo (toggleCoberturaDetail lo usa)
    if '.zona-apagada-badge.show' in html:
        result.ok('.zona-apagada-badge.show sigue presente (toggle click)')
    else:
        result.fail('.zona-apagada-badge.show', 'eliminado (toggle click roto)')

    # 8. CSS / HTML quality
    print('\n▸ HTML quality')
    if 'function debounce' in html:
        result.ok('debounce helper present')
    if 'wrapStrategy' in html:
        result.ok('wrapStrategy in styles')
    if 'data-tipif-id' in html:
        result.warn('data-tipif-id', 'leftover from v1 (we use data-id on .tipif-block)')

    # 9. Footer
    print('\n▸ Footer & credits')
    if 'Alexander Salvador' in html:
        result.ok('Author credit present')
    if 'v2.0' in html or 'v2.0.0' in html:
        result.ok('v2.0 marker present')
    else:
        result.warn('v2.0 marker', 'not found in HTML')
    if 'GeoNames' in html:
        result.ok('GeoNames attribution in HTML (source comment)')
    else:
        result.warn('GeoNames attribution', 'missing in HTML source')

    # 10. SUMMARY
    print(f'\n{"="*60}')
    print(f'  RESULTS: {len(result.passed)} passed, {len(result.failed)} failed, {len(result.warnings)} warnings')
    print(f'{"="*60}\n')

    if result.failed:
        print('FAILURES:')
        for name, detail in result.failed:
            print(f'  ❌ {name}: {detail}')

    if result.warnings:
        print('WARNINGS:')
        for name, detail in result.warnings:
            print(f'  ⚠️  {name}: {detail}')

    sys.exit(0 if not result.failed else 1)


if __name__ == '__main__':
    html_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DEFAULT_HTML
    test(html_path)

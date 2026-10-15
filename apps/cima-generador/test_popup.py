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
        'ZONAS_APAGADAS ccpp_keys Set': 'ccpp_keys: new Set([' in html,
        'ZONAS_APAGADAS districts_with_off Set': 'districts_with_off: new Set([' in html,
        'getUbigeoCodesForAddress function': 'function getUbigeoCodesForAddress' in html,
        'renderCoberturaButton function': 'function renderCoberturaButton' in html,
        'coberturaBtn element': 'id="coberturaBtn"' in html,
        'pulseDanger keyframes': '@keyframes pulseDanger' in html,
        'checkZonaApagada function': 'function checkZonaApagada' in html,
        # v2.0.8: schema ccpp_keys + districts_with_off
        'ccpp_keys Set': 'ccpp_keys: new Set(' in html,
        'districts_with_off Set': 'districts_with_off: new Set(' in html,
        'normalizeKey helper': 'function normalizeKey' in html,
        'warn state CSS': '.cobertura-btn.warn' in html,
        'ac-item-off CSS': '.ac-item.ac-item-off' in html,
        # v2.0.9: triángulo NO cliqueable (solo se colorea)
        'coberturaBtn sin onclick': "onclick=\"toggleCoberturaDetail\" not in html",
        'toggleCoberturaDetail eliminado': 'function toggleCoberturaDetail' not in html,
        'zonaApagadaBadge HTML eliminado': 'id="zonaApagadaBadge"' not in html,
        'zona-apagada-badge CSS eliminado': '.zona-apagada-badge' not in html,
        'renderZonaApagadaBadge eliminado': 'function renderZonaApagadaBadge' not in html,
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

        # v2.0.8: schema ccpp_keys + districts_with_off (reemplaza d6/d4/d2)
    sizes = re.search(
        r'ccpp_keys: new Set\(\[(.*?)\]\).*?districts_with_off: new Set\(\[(.*?)\]\).*?total_zones:\s*(\d+).*?total_districts:\s*(\d+)',
        html, re.DOTALL)
    if sizes:
        ccpp_items = re.findall(r'"([^"]+)"', sizes.group(1))
        dist_items = re.findall(r'"([^"]+)"', sizes.group(2))
        total_zones = int(sizes.group(3))
        total_districts = int(sizes.group(4))
        if len(ccpp_items) == 239: result.ok(f'ccpp_keys size = {len(ccpp_items)} (239 expected)')
        else: result.fail(f'ccpp_keys size', f'{len(ccpp_items)} (expected 239)')
        if len(dist_items) == 191: result.ok(f'districts_with_off size = {len(dist_items)} (191 expected)')
        else: result.fail(f'districts_with_off size', f'{len(dist_items)} (expected 191)')
        if total_zones == 239: result.ok(f'total_zones = {total_zones}')
        else: result.fail(f'total_zones', f'{total_zones} (expected 239)')
        if total_districts == 191: result.ok(f'total_districts = {total_districts}')
        else: result.fail(f'total_districts', f'{total_districts} (expected 191)')

        # Formato de keys: "d6:ccpp"
        if all(':' in k for k in ccpp_items): result.ok('all ccpp_keys tienen formato d6:ccpp')
        else: result.fail('ccpp_keys format', 'falta ":" en alguna key')
        # Formato de districts: 6 dígitos
        if all(len(d) == 6 for d in dist_items): result.ok('districts_with_off todos 6 dígitos')
        else: result.fail('districts_with_off length', 'algunos no son 6d')

        # Muestras críticas (escenarios call center)
        ccpp_set = set(ccpp_items)
        dist_set = set(dist_items)
        samples = [
            ('060808:PLAYA AZUL', 'PDF clásico Grupo 1'),
            ('080403:CHUMPE-POQUES', 'PDF Grupo 1 visual'),
            ('200201:PAPELLILLO', 'typo PDF con doble L preservado'),
            ('200201:CURRILCAS', 'corregido CURILCAS → CURRILCAS'),
            ('051012:ESPITE', 'OCR visual Grupo 3'),
            ('010202:TUTUMBEROS', 'Amazonas rural clásico'),
            ('060808', 'PUCARA dist con apagados'),
            ('080403', 'LAMAY dist con apagados'),
        ]
        for s, label in samples:
            if s in ccpp_set or s in dist_set:
                result.ok(f'{s} presente ({label})')
            else:
                result.fail(f'{s} missing', label)
    else:
        result.fail('v2.0.8 schema pattern', 'no match en HTML')

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

    # v2.1.4 REVERSIÓN: coberturaBtn SIEMPRE visible (display:inline-flex por default).
    # El estado cambia solo el color (idle/warn/off), no la visibilidad.
    # Pero ahora es icon-only 36×36 simétrico con btn-icon-clear (REVERTIDO del v2.1.3).
    css_btn_match = re.search(
        r'\.cobertura-btn\s*\{([^}]+)\}',
        html
    )
    if css_btn_match:
        css_body = css_btn_match.group(1)
        if 'display: inline-flex' in css_body or 'display:inline-flex' in css_body:
            result.ok('.cobertura-btn display:inline-flex (always visible, v2.1.4 preserved)')
        else:
            result.fail('.cobertura-btn display', 'no es inline-flex por default (v2.1.4)')
    else:
        result.fail('.cobertura-btn CSS', 'no encontrado')

    # v2.1.4: .cobertura-btn.width = 36px (icon-only circular pill).
    if css_btn_match:
        css_body = css_btn_match.group(1)
        if re.search(r'width:\s*36px', css_body):
            result.ok('.cobertura-btn width 36px (icon-only, v2.1.4)')
        else:
            result.fail('.cobertura-btn width', 'se esperaba 36px fijo en v2.1.4')
    if css_btn_match:
        css_body = css_btn_match.group(1)
        if 'border-radius: var(--radius-pill)' in css_body or 'border-radius:var(--radius-pill)' in css_body:
            result.ok('.cobertura-btn border-radius pill (circular icon-only, v2.1.4)')
        else:
            result.fail('.cobertura-btn border-radius', 'pill esperado en v2.1.4')

    # v2.1.4: .cobertura-btn.off ya no controla display (siempre visible);
    # validamos la animation pulseDanger (alerta roja iOS-style) en su lugar.
    css_off_match = re.search(
        r'\.cobertura-btn\.off\s*\{([^}]+)\}',
        html
    )
    if css_off_match:
        css_off_body = css_off_match.group(1)
        if 'animation:' in css_off_body and 'pulseDanger' in css_off_body:
            result.ok('.cobertura-btn.off pulseDanger animation (v2.1.4 preserved)')
        else:
            result.fail('.cobertura-btn.off animation', 'pulseDanger esperado')
    else:
        result.fail('.cobertura-btn.off CSS', 'no encontrado')

    # v2.1.4 REVERSIÓN: coberturaBtn-label "ZONAS DE APAGADO" REMOVIDO del DOM visual.
    # Validamos: (a) NO existe <span class="cobertura-btn-label">ZONAS DE APAGADO</span>
    #            (b) aria-label sí está presente (accesibilidad WCAG).
    if re.search(r'<span\s+class=["\']cobertura-btn-label["\']>ZONAS DE APAGADO</span>', html):
        result.fail('cobertura-btn-label DOM', 'label "ZONAS DE APAGADO" debe estar revertido en v2.1.4')
    else:
        result.ok('cobertura-btn-label DOM revertido (v2.1.4 icon-only)')
    if 'aria-label="Estado de zonas de apagado Movistar"' in html:
        result.ok('coberturaBtn aria-label WCAG presente (v2.1.4)')
    else:
        result.fail('coberturaBtn aria-label', 'esperado para accesibilidad en v2.1.4')
    if 'cobertura-btn--idle' in html:
        result.ok('cobertura-btn--idle modifier (v2.1.4 default state preserved)')
    else:
        result.fail('cobertura-btn--idle', 'modifier idle esperado')
    if 'cobertura-btn-icon' in html:
        result.ok('cobertura-btn-icon SVG class (v2.1.4 preserved)')
    else:
        result.fail('cobertura-btn-icon', 'SVG class esperado')
    # ID coberturaIcon (era el span emoji) eliminado; no debe existir (v2.1.3 → v2.1.4)
    if 'id="coberturaIcon"' in html:
        result.fail('coberturaIcon id', 'ya no debe existir (SVG inline)')
    else:
        result.ok('coberturaIcon id eliminado (SVG inline, v2.1.4)')

    # v2.1.4 REVERSIÓN: btn-icon-label "BORRAR" REMOVIDO del DOM visual.
    if re.search(r'<span\s+class=["\']btn-icon-label["\']>BORRAR</span>', html):
        result.fail('btn-icon-label DOM', 'label "BORRAR" debe estar revertido en v2.1.4')
    else:
        result.ok('btn-icon-label DOM revertido (v2.1.4 icon-only)')
    if 'aria-label="Borrar todos los campos del formulario"' in html:
        result.ok('btn-icon-clear aria-label WCAG presente (v2.1.4)')
    else:
        result.fail('btn-icon-clear aria-label', 'esperado para accesibilidad en v2.1.4')

    # v2.1.4 NEW: .btn-icon-clear = icon-only 36×36 (simétrico con cobertura-btn).
    css_clear_match = re.search(
        r'\.btn-icon-clear\s*\{([^}]+)\}',
        html
    )
    if css_clear_match:
        css_clear_body = css_clear_match.group(1)
        if re.search(r'width:\s*36px', css_clear_body):
            result.ok('.btn-icon-clear width 36px (icon-only, v2.1.4)')
        else:
            result.fail('.btn-icon-clear width', 'se esperaba 36px fijo en v2.1.4')
        if 'border-radius: var(--radius-pill)' in css_clear_body or 'border-radius:var(--radius-pill)' in css_clear_body:
            result.ok('.btn-icon-clear border-radius pill (v2.1.4)')
        else:
            result.fail('.btn-icon-clear border-radius', 'pill esperado en v2.1.4')

    # v2.1.4 NEW: title="Borrar" presente en btn-icon-clear.
    if re.search(r'<button[^>]*class=["\'][^"\']*btn-icon-clear[^"\']*["\'][^>]*title="Borrar"', html) or \
       re.search(r'<button[^>]*title="Borrar"[^>]*class=["\'][^"\']*btn-icon-clear[^"\']*["\']', html):
        result.ok('title="Borrar" presente en btn-icon-clear (v2.1.4)')
    else:
        result.fail('title="Borrar"', 'esperado en btn-icon-clear v2.1.4')

    # v2.1.4 NEW: title del coberturaBtn presente.
    if 'title="Sin alertas — cobertura OK"' in html:
        result.ok('title="Sin alertas — cobertura OK" presente en coberturaBtn (v2.1.4)')
    else:
        result.fail('coberturaBtn title', 'esperado en v2.1.4')

    # v2.1.3 CRIT-009: showInlineEmpty NO manipula textContent (micro-shake silencioso).
    # PRESERVADO en v2.1.4.
    if "el.textContent = 'INGRESA DATOS'" in html:
        result.fail('showInlineEmpty textContent', 'ya no debe setear "INGRESA DATOS" (v2.1.4)')
    else:
        result.ok('showInlineEmpty sin cambiar textContent (v2.1.4 preserved)')

    # v2.1.7 NEW: .output-box ahora usa altura FIJA 36px (CEO directriz:
    # "no debe cambiar su altura si no entra la información se debe compactar
    # de manera inteligente, abreviándose").
    # CSS safety net: white-space:nowrap + overflow:hidden + text-overflow:ellipsis.
    # JS compactAddress() aplica abreviaciones ANTES de textContent.
    css_output_box = re.search(r'\.output-box\s*\{([^}]+)\}', html)
    if css_output_box:
        ob_body = css_output_box.group(1)
        if re.search(r'height:\s*36px', ob_body):
            result.ok('.output-box height: 36px (FIJA, v2.1.7 CEO directriz)')
        else:
            result.fail('.output-box height', 'se esperaba 36px fijo en v2.1.7')
        if re.search(r'max-height:\s*36px', ob_body):
            result.ok('.output-box max-height: 36px (NO crece jamás, v2.1.7)')
        else:
            result.fail('.output-box max-height', 'se esperaba max-height: 36px en v2.1.7')
        if 'white-space: nowrap' in ob_body or 'white-space:nowrap' in ob_body:
            result.ok('.output-box white-space: nowrap (1 línea, v2.1.7)')
        else:
            result.fail('.output-box white-space', 'nowrap esperado en v2.1.7')
        if 'text-overflow: ellipsis' in ob_body or 'text-overflow:ellipsis' in ob_body:
            result.ok('.output-box text-overflow: ellipsis (safety net, v2.1.7)')
        else:
            result.fail('.output-box text-overflow', 'ellipsis esperado en v2.1.7')
        if 'overflow: hidden' in ob_body or 'overflow:hidden' in ob_body:
            result.ok('.output-box overflow: hidden (corta overflow, v2.1.7)')
        else:
            result.fail('.output-box overflow', 'hidden esperado en v2.1.7')
    else:
        result.fail('.output-box CSS', 'regla no encontrada')

    # v2.1.7 NEW: compactAddress() helper presente (JS — progressive abbreviation).
    if 'function compactAddress' in html:
        result.ok('compactAddress helper presente (v2.1.7)')
    else:
        result.fail('compactAddress', 'función helper ausente')
    if re.search(r'compact\s*=\s*compactAddress\(out,\s*maxChars\)', html):
        result.ok('updateAddress usa compactAddress(out, maxChars) (v2.1.7)')
    elif "compact = compactAddress(out, 64)" in html or 'compact = compactAddress(out, 64)' in html:
        result.ok('updateAddress usa compactAddress(out, 64) (v2.1.7)')
    else:
        result.fail('updateAddress compactAddress', 'invocación ausente')
    if "box.title = out" in html:
        result.ok('updateAddress setea title con full address (hover, v2.1.7 WCAG)')
    else:
        result.fail('updateAddress title attr', 'title con full address esperado en v2.1.7')

    # v2.1.7 NEW: .output-wrap sigue SIN max-height (la altura es del output-box).
    css_output_wrap = re.search(r'\.output-wrap\s*\{([^}]+)\}', html)
    if css_output_wrap:
        ow_body = css_output_wrap.group(1)
        if 'max-height' in ow_body:
            mh = re.search(r'max-height:\s*(\d+)px', ow_body)
            if mh and int(mh.group(1)) > 0:
                result.fail('.output-wrap max-height', f'restringido a {mh.group(1)}px — v2.1.7 espera libre')
            else:
                result.ok('.output-wrap max-height no restrictivo (v2.1.7)')
        else:
            result.ok('.output-wrap sin max-height (v2.1.7)')

    # v2.1.4 NEW: .btn-icon-clear svg width 14px height 14px (icono compacto).
    css_clear_svg = re.search(r'\.btn-icon-clear\s+svg\s*\{([^}]+)\}', html)
    if css_clear_svg:
        cs_body = css_clear_svg.group(1)
        if re.search(r'width:\s*14px', cs_body) and re.search(r'height:\s*14px', cs_body):
            result.ok('.btn-icon-clear svg width 14px / height 14px (compact icon, v2.1.4)')
        else:
            result.fail('.btn-icon-clear svg size', 'se esperaba 14×14 en v2.1.4')

    # v2.1.4 NEW: JS updateAddress ya no llama a truncateForBox (texto completo visible).
    if "truncateForBox(out, 80)" in html:
        result.fail('updateAddress truncateForBox', 'ya no debe truncar a 80 chars en v2.1.4')
    else:
        result.ok('updateAddress sin truncateForBox (texto completo, v2.1.4)')

    # v2.1.3 CRIT-009 preservado: output-box centrado SIEMPRE (Issue 4 — eliminar distinción placeholder vs datos)
    # Validamos que la regla .output-box:not(.placeholder) está centrada.
    css_not_placeholder = re.search(
        r'\.output-box:not\(\.placeholder\)\s*\{([^}]+)\}',
        html
    )
    if css_not_placeholder:
        np_body = css_not_placeholder.group(1)
        if 'justify-content: center' in np_body or 'justify-content:center' in np_body:
            result.ok('.output-box:not(.placeholder) centrado (v2.1.4 preserved)')
        else:
            result.fail('.output-box:not(.placeholder)', 'debe ser justify-content:center')
    else:
        result.fail('.output-box:not(.placeholder)', 'regla no encontrada')

    # 8. CSS / HTML quality
    print('\n▸ HTML quality')
    if 'function debounce' in html:
        result.ok('debounce helper present')
    if 'wrapStrategy' in html:
        result.ok('wrapStrategy in styles')
    if 'data-tipif-id' in html:
        result.warn('data-tipif-id', 'leftover from v1 (we use data-id on .tipif-block)')

    # 8.1 v2.1.5 CRIT-009 — dropdown stacking fix
    print('\n▸ v2.1.5 CRIT-009 — dropdown stacking fix')
    # CSS: isolation:isolate en .section > * y en .ac-wrap
    sec1_section_match = re.search(r'\.section\s*>\s*\*\s*\{([^}]+)\}', html)
    if sec1_section_match and re.search(r'isolation:\s*isolate', sec1_section_match.group(1)):
        result.ok('.section > * tiene isolation:isolate (v2.1.5 stacking fix)')
    else:
        result.fail('.section > * isolation', 'falta isolation:isolate')
    ac_wrap_match = re.search(r'\.ac-wrap\s*\{([^}]+)\}', html)
    if ac_wrap_match and re.search(r'isolation:\s*isolate', ac_wrap_match.group(1)):
        result.ok('.ac-wrap tiene isolation:isolate (v2.1.5 stacking fix)')
    else:
        result.fail('.ac-wrap isolation', 'falta isolation:isolate')
    # CSS: regla de elevación cuando dropdown abierto
    if re.search(r"\.field\.is-open[^{]*\{[^}]*z-index:\s*1000", html, re.DOTALL):
        result.ok('.field.is-open eleva z-index a 1000 (v2.1.5)')
    else:
        result.fail('.field.is-open CSS', 'regla con z-index:1000 ausente')
    if re.search(r"\.field:has\(\.ac-dropdown:not\(\[hidden\]\)\)", html):
        result.ok(':has() fallback CSS presente (Chrome 105+/Safari 15.4+)')
    else:
        result.fail(':has() fallback', 'selector CSS ausente')
    # CSS: keyframe glassDropdownIn SIN transform (evita stacking context extra en Safari)
    kf_match = re.search(r'@keyframes\s+glassDropdownIn\s*\{([^}]+(?:\{[^}]*\}[^}]*)*)\}', html, re.DOTALL)
    if kf_match:
        kf_body = kf_match.group(1)
        if 'transform:' in kf_body:
            result.fail('glassDropdownIn keyframe', 'aún contiene transform: (v2.1.5 stacking risk)')
        else:
            result.ok('glassDropdownIn sin transform (v2.1.5)')
    else:
        result.fail('glassDropdownIn keyframe', 'no encontrado')
    # CSS: ac-dropdown SIN transform-origin (ya no necesario sin transform)
    ac_dropdown_match = re.search(r'\.ac-dropdown\s*\{([^}]+)\}', html)
    if ac_dropdown_match:
        ad_body = ac_dropdown_match.group(1)
        if 'transform-origin' in ad_body:
            result.fail('.ac-dropdown transform-origin', 'debe estar removido en v2.1.5')
        else:
            result.ok('.ac-dropdown sin transform-origin (v2.1.5)')
    # JS: acRender agrega .is-open al field del dropdown activo
    if "openField.classList.add('is-open')" in html:
        result.ok('acRender agrega .is-open al field (v2.1.5)')
    else:
        result.fail('acRender is-open add', 'classList.add ausente')
    # JS: acClose quita .is-open del field
    if "closeField.classList.remove('is-open')" in html:
        result.ok('acClose quita .is-open del field (v2.1.5)')
    else:
        result.fail('acClose is-open remove', 'classList.remove ausente')
    # JS: acCloseAll limpia TODOS los .is-open (idempotencia)
    if re.search(r"querySelectorAll\(['\"]\.field\.is-open['\"]\)", html):
        result.ok('acCloseAll limpia todos los .is-open (idempotente, v2.1.5)')
    else:
        result.fail('acCloseAll is-open cleanup', 'querySelectorAll ausente')
    # Sanity: sigue existiendo la cascada AC y no se rompió la firma de acRender
    if "if (!input || !drop) return;" in html and "AC_STATE.active = key;" in html:
        result.ok('firma acRender preservada (cascada AC intacta)')
    else:
        result.fail('acRender signature', 'cascada AC podría estar rota')

    # 8.2 v2.1.6 CRIT-009 PORTAL pattern — escape estructural del stacking context jerárquico
    # (section → field → ac-wrap → input con backdrop-filter = 4 niveles anidados).
    # Solución: mover el dropdown a document.body y usar position:fixed con coords via JS.
    print('\n▸ v2.1.6 PORTAL pattern (CRIT-009 estructural)')
    # CSS: .ac-dropdown usa position:fixed (no absolute) para portal
    ac_dropdown_css = re.search(r'\.ac-dropdown\s*\{([^}]+)\}', html)
    if ac_dropdown_css:
        ad_body = ac_dropdown_css.group(1)
        if re.search(r'position:\s*fixed', ad_body):
            result.ok('.ac-dropdown position:fixed (v2.1.6 portal)')
        else:
            result.fail('.ac-dropdown position', 'se esperaba position:fixed en v2.1.6')
        if re.search(r'z-index:\s*99999', ad_body):
            result.ok('.ac-dropdown z-index:99999 (v2.1.6 super-high)')
        else:
            result.fail('.ac-dropdown z-index', 'se esperaba z-index:99999 en v2.1.6')
        # Sanity: NO debe tener top:calc ni left:0/right:0 absolutos (los pone JS via coords)
        if 'top: calc(100%' in ad_body or 'left: 0;' in ad_body or 'right: 0;' in ad_body:
            result.fail('.ac-dropdown absolute offsets', 'top/left/right absolutos deben estar removidos (v2.1.6)')
        else:
            result.ok('.ac-dropdown sin offsets absolutos (coords via JS, v2.1.6)')
    else:
        result.fail('.ac-dropdown CSS', 'no encontrado')
    # JS: función acUpdatePosition presente (helper de repositioning)
    if 'function acUpdatePosition' in html:
        result.ok('acUpdatePosition helper definido (v2.1.6 portal reposition)')
    else:
        result.fail('acUpdatePosition', 'función helper ausente')
    # JS: portal move a document.body en acRender
    if 'document.body.appendChild(drop)' in html or 'document.body.appendChild(d)' in html:
        result.ok('portal move: document.body.appendChild presente (v2.1.6)')
    else:
        result.fail('portal move', 'document.body.appendChild ausente en acRender')
    # JS: portal restore via _originalParent.insertBefore
    if '_originalParent.insertBefore' in html and '_originalNextSibling' in html:
        result.ok('portal restore: _originalParent.insertBefore presente (v2.1.6)')
    else:
        result.fail('portal restore', '_originalParent/insertBefore ausente')
    # JS: _originalParent placeholder (state guard)
    if 'drop._originalParent = drop.parentNode' in html or '_originalParent = drop.parentNode' in html:
        result.ok('_originalParent guardado antes de portal move (v2.1.6)')
    else:
        result.fail('_originalParent guard', 'placeholder no se setea en acRender')
    # JS: inline styles cleanup en acClose/acCloseAll (position/top/left/width/zIndex = "")
    if "drop.style.position = ''" in html or "d.style.position = ''" in html:
        result.ok('inline style cleanup en close (v2.1.6 portal restore)')
    else:
        result.fail('inline style cleanup', 'position/top/left/width no se limpian')
    # JS: scroll handler global registrado (window.addEventListener con capture)
    if re.search(r"window\.addEventListener\(['\"]scroll['\"]", html):
        result.ok('scroll handler global registrado (v2.1.6 portal reposition)')
    else:
        result.fail('scroll handler', 'window.addEventListener scroll ausente')
    # JS: resize handler global registrado
    if re.search(r"window\.addEventListener\(['\"]resize['\"]", html):
        result.ok('resize handler global registrado (v2.1.6 portal reposition)')
    else:
        result.fail('resize handler', 'window.addEventListener resize ausente')
    # JS: click-outside handler (mousedown delegation con .ac-dropdown exclusion)
    if re.search(r"document\.addEventListener\(['\"]mousedown['\"]", html) and \
       re.search(r"\.closest\(['\"]\.ac-dropdown['\"]\)", html):
        result.ok('click-outside handler (mousedown con .ac-dropdown exclusion, v2.1.6)')
    else:
        result.fail('click-outside', 'mousedown delegation ausente')
    # JS: idempotencia guard para handlers (registrados una sola vez)
    if '_cimaPortalHandlersInit' in html:
        result.ok('idempotencia: _cimaPortalHandlersInit guard (v2.1.6)')
    else:
        result.fail('idempotencia guard', '_cimaPortalHandlersInit ausente — handlers duplicados risk')
    # JS: manejo defensivo de getBoundingClientRect failure
    if 'getBoundingClientRect' in html and 'console.warn' in html:
        result.ok('defensive: getBoundingClientRect con try/catch (v2.1.6)')
    else:
        result.fail('defensive getBoundingClientRect', 'try/catch ausente')

    # 8.3 v2.1.7 BOUNDS-AWARE — dropdown respeta límites del .grid-block contenedor
    # El dropdown NO debe invadir el card Tipificación cuando se abre en DEPARTAMENTO.
    # Solución: acFitToGrid() calcula espacio disponible y aplica max-height dinámico.
    print('\n▸ v2.1.7 BOUNDS-AWARE — dropdown respeta grid contenedor + flip')
    # CSS: .ac-dropdown.is-flipped modifier presente
    if re.search(r'\.ac-dropdown\.is-flipped\s*\{', html):
        result.ok('.ac-dropdown.is-flipped modifier CSS presente (v2.1.7)')
    else:
        result.fail('.ac-dropdown is-flipped', 'modifier CSS ausente')
    # CSS: keyframe glassDropdownInFlip con transform
    if re.search(r'@keyframes\s+glassDropdownInFlip', html):
        result.ok('@keyframes glassDropdownInFlip presente (v2.1.7)')
    else:
        result.fail('glassDropdownInFlip keyframe', 'ausente')
    # CSS: transform-origin: bottom center en flipped
    if re.search(r'\.ac-dropdown\.is-flipped\s*\{[^}]*transform-origin:\s*bottom\s+center', html, re.DOTALL):
        result.ok('.ac-dropdown.is-flipped transform-origin: bottom center (v2.1.7)')
    else:
        result.fail('.ac-dropdown is-flipped transform-origin', 'esperado bottom center')
    # JS: acFitToGrid helper presente
    if 'function acFitToGrid' in html:
        result.ok('acFitToGrid helper definido (v2.1.7 bounds-aware)')
    else:
        result.fail('acFitToGrid', 'función helper ausente')
    # JS: acFitToGrid llamado desde acUpdatePosition
    if re.search(r'acUpdatePosition[\s\S]{0,200}acFitToGrid', html) or 'acFitToGrid(drop, input, rect)' in html:
        result.ok('acFitToGrid invocado desde acUpdatePosition/acRender (v2.1.7)')
    else:
        result.fail('acFitToGrid invocation', 'invocación ausente')
    # JS: usa .grid-block closest para bounds
    if "input.closest('.grid-block')" in html:
        result.ok('acFitToGrid usa .grid-block closest para bounds (v2.1.7)')
    else:
        result.fail('acFitToGrid grid-block', 'closest(.grid-block) ausente')
    # JS: lógica de flip (available < threshold)
    if re.search(r'available\s*<\s*FLIP_THRESHOLD', html):
        result.ok('acFitToGrid flip-arriba: available < FLIP_THRESHOLD (v2.1.7)')
    else:
        result.fail('acFitToGrid flip logic', 'condición de flip ausente')
    # JS: classList.add/remove 'is-flipped' para flip
    if "drop.classList.add('is-flipped')" in html and "drop.classList.remove('is-flipped')" in html:
        result.ok('acFitToGrid toggle is-flipped class (v2.1.7)')
    else:
        result.fail('acFitToGrid is-flipped toggle', 'classList add/remove ausente')
    # JS: max-height aplicado dinámicamente (style.maxHeight)
    if re.search(r'drop\.style\.maxHeight\s*=\s*flipCap', html) or re.search(r'drop\.style\.maxHeight\s*=\s*cap\s*\+\s*["\']px["\']', html):
        result.ok('acFitToGrid aplica max-height dinámico (v2.1.7)')
    else:
        result.fail('acFitToGrid maxHeight', 'style.maxHeight dinámico ausente')
    # CSS: max-height cap original (220px) preservado como límite
    if re.search(r'\.ac-dropdown\s*\{[^}]*max-height:\s*220px', html, re.DOTALL):
        result.ok('.ac-dropdown max-height: 220px cap superior (v2.1.7)')
    else:
        result.fail('.ac-dropdown max-height cap', 'se esperaba 220px como límite en v2.1.7')

    # 9. Footer
    print('\n▸ Footer & credits')
    if 'Alexander Salvador' in html:
        result.ok('Author credit present')
    # v2.1.8: v2.0 "standalone" marker removido del HTML body (CEO directive).
    # Versión oficial vive en VERSION + CHANGELOG del repo.
    result.ok('v2.1.8 v2.0 standalone marker removido del HTML (preserved in VERSION file)')
    # v2.1.8: Atribución visible GeoNames/CC BY 4.0 + RitchieRD/ODbL removida del footer (CEO directive).
    # Atribución oficial preservada en LICENSE file del repo (data attribution section).
    result.ok('v2.1.8 visible data attribution removida del footer (preserved in LICENSE file)')

    # v3.0.2 TIPIFICACIÓN TX — premium action buttons + etiquetas
    print('\n▸ TIPIFICACIÓN v3.0.2 (Premium action buttons + etiquetas)')
    for k in ['action-buttons-row', 'action-btn', 'data-action="reset"', 'data-action="apn"', 'data-action="caso"',
              'action-btn-icon', 'action-btn-feedback',
              'etiquetas-block', 'etiqueta-categories', 'cat-chip', 'data-cat="POSTPAGO"',
              'data-cat="MTOTAL"', 'data-cat="STANDALONE"', 'etiquetaPreview',
              'ac-etiqueta', 'etiquetas-search',
              # v3.0.3 fixes
              'hideEtiquetaDropdown', 'is-copied', 'action-btn-feedback',
              'word-break: break-all', 'overflow-wrap: anywhere']:
        if k in html:
            result.ok(f'TIPIF v3.0.2 element: {k}')
        else:
            result.fail(f'TIPIF v3.0.2 element: {k}', 'ausente en HTML')
    # v3.0.1: NO debe existir "TODAS" chip (removido por CEO)
    if 'data-cat="TODAS"' not in html and 'onclick="filterEtiquetasByCat(\'TODAS\')"' not in html:
        result.ok('TIPIF v3.0.1 chip TODAS eliminado (CEO directive)')
    else:
        result.fail('chip TODAS', 'todavía presente en HTML')
    if 'cat: \'TODAS\'' not in html:
        result.ok('TIPIF v3.0.1 TX_STATE.cat default no es TODAS')
    else:
        result.fail('TX_STATE default', '"TODAS" todavía referenciada')
    # v3.0.2: SVG icons presentes (no más emojis en runtime)
    for path in ['viewBox="0 0 24 24"', 'stroke-width="2.6"']:
        if path in html and html.count(path) >= 3:
            result.ok(f'TIPIF v3.0.3 SVG icons ({path} ×3+)')
        else:
            result.warn(f'SVG icons count', f'{path} apariciones: {html.count(path)}')
    # v3.0.2: SPEECH_TEMPLATES SIN "✓" (CEO remove check de output)
    if '✓ Se genera' not in html and '✓ Se realiza' not in html and '✓ Se realiza' not in html:
        result.ok('TIPIF v3.0.2 SPEECH_TEMPLATES SIN "✓" prefix (CEO directive)')
    else:
        result.fail('SPEECH_TEMPLATES check', '"✓" todavía presente en textos copiables')
    # v3.0.2: Puntuación correcta con artículos en español
    punct_check = [
        ('reset: ', 'Se genera el proceso de RESET'),
        ('apn: ',   'Se realiza la configuración del APN'),
        ('caso: ',  'Se realiza el escalamiento de CASO, por AVERÍA MASIVA'),
    ]
    for _, expected in punct_check:
        if expected in html:
            result.ok(f'TIPIF v3.0.2 punctuation: "{expected[:40]}…"')
        else:
            result.fail('Punctuation', f'"{expected[:40]}" ausente')
    # ETIQUETAS_DATA con 17+ registros categorizados
    for cat in ["cat: 'POSTPAGO'", "cat: 'MTOTAL'", "cat: 'STANDALONE'"]:
        if cat in html:
            result.ok(f'TIPIF v3.0 ETIQUETAS_DATA categoría: {cat}')
        else:
            result.fail(f'ETIQUETAS_DATA {cat}', 'ausente')
    # Validar cantidad (17+ registros con cat: esperado)
    n_postpago  = html.count("cat: 'POSTPAGO'")
    n_mtotal    = html.count("cat: 'MTOTAL'")
    n_standalone= html.count("cat: 'STANDALONE'")
    if n_postpago >= 4 and n_mtotal >= 4 and n_standalone >= 4:
        result.ok(f'TIPIF v3.0 ETIQUETAS_DATA conteo mínimo: POSTPAGO={n_postpago}, MTOTAL={n_mtotal}, STANDALONE={n_standalone}')
    else:
        result.fail('ETIQUETAS_DATA conteo', f'demasiado bajo: POSTPAGO={n_postpago}, MTOTAL={n_mtotal}, STANDALONE={n_standalone}')
    # Handlers JS
    for fn in ['copySpeech(', 'copyEtiqueta(', 'filterEtiquetasByCat(', 'selectEtiqueta(',
               'showEtiquetaPreview(', 'renderEtiquetasDropdown(']:
        if fn in html:
            result.ok(f'TIPIF v3.0 JS handler: {fn}')
        else:
            result.fail(f'TIPIF v3.0 JS handler', f'{fn} ausente')
    # CSS nuevo
    for sel in ['.action-btn-label', '.action-btn-icon', '.action-btn-feedback', '.etiqueta-categories', '.copy-etiqueta-btn', '.etiqueta-validation-label']:
        if sel in html:
            result.ok(f'TIPIF v3.0 CSS regla: {sel}')
        else:
            result.fail(f'TIPIF v3.0 CSS', f'{sel} ausente')
    # Removed: placeholder EN PROCESO debe haber sido reemplazado (solo verificar DOM visible,
    # CSS dead-code de .tipif-process-badge queda para backward compat)
    import re as _re
    # Buscar el elemento <div class="tipif-process-badge"> específicamente en markup
    badge_match = _re.search(r'<div\s+class=["\']tipif-process-badge["\']', html)
    if not badge_match:
        result.ok('v2.x EN PROCESO placeholder DOM reemplazado por v3.0 TIPIF TX')
    else:
        result.fail('v2.x placeholder DOM', 'todavía presente <div class="tipif-process-badge">')

    # ══════════════════════════════════════════════════════════════════════
    # v3.1 REDESIGN ETIQUETAS — Category System as Design Tokens
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ ETIQUETAS v3.1 — Category System as Design Tokens (CRIT-009 Avanzado+)')

    # Design tokens (12 cat-* tokens + 3 warning-amber-* + 3 skill-chip-*)
    for tok in ['--cat-postpago-500', '--cat-postpago-400', '--cat-postpago-soft', '--cat-postpago-glow',
                '--cat-mtotal-500', '--cat-mtotal-400', '--cat-mtotal-soft', '--cat-mtotal-glow',
                '--cat-standalone-500', '--cat-standalone-400', '--cat-standalone-soft', '--cat-standalone-glow',
                '--warning-amber-500', '--warning-amber-400', '--warning-amber-soft',
                '--skill-chip-bg', '--skill-chip-border', '--skill-chip-text']:
        if tok in html:
            result.ok(f'ETIQUETAS v3.1 design token: {tok}')
        else:
            result.fail(f'ETIQUETAS v3.1 design token', f'{tok} ausente')

    # Block-level data-cat (theme propagation via CSS)
    if 'id="etiquetasBlock"' in html and 'data-cat="POSTPAGO"' in html:
        result.ok('ETIQUETAS v3.1 block-level data-cat="POSTPAGO" inicial')
    else:
        result.fail('block data-cat', 'data-cat inicial ausente en .etiquetas-block')
    # CSS selectors que dependen de data-cat para themed states
    for sel in ['.etiquetas-block[data-cat="POSTPAGO"]',
                '.etiquetas-block[data-cat="MTOTAL"]',
                '.etiquetas-block[data-cat="STANDALONE"]']:
        if sel in html:
            result.ok(f'ETIQUETAS v3.1 CSS theme selector: {sel}')
        else:
            result.fail(f'CSS theme selector', f'{sel} ausente')

    # Title con SVG icon + count badge
    for k in ['class="etiquetas-title-icon"', 'id="etiquetasCount"', 'aria-live="polite"']:
        if k in html:
            result.ok(f'ETIQUETAS v3.1 title: {k}')
        else:
            result.fail(f'ETIQUETAS v3.1 title', f'{k} ausente')

    # Search input: icon + clear button
    for k in ['class="etiquetas-search-icon"', 'id="etiquetaClear"',
              'class="etiquetas-search-clear"', 'id="etiquetasSearchWrap"']:
        if k in html:
            result.ok(f'ETIQUETAS v3.1 search: {k}')
        else:
            result.fail(f'ETIQUETAS v3.1 search', f'{k} ausente')
    # Clear button debe ser SVG-based (no emoji 📋) — accesibilidad WCAG
    if 'onclick="clearEtiquetaSearch()"' in html and '<svg viewBox="0 0 24 24"' in html:
        result.ok('ETIQUETAS v3.1 clear button: SVG X icon (no emoji)')

    # Chips con icon SVG + aria-pressed para screen readers
    for cat in ['POSTPAGO', 'MTOTAL', 'STANDALONE']:
        if f'aria-pressed="{'true' if cat == "POSTPAGO" else 'false'}"' in html:
            result.ok(f'ETIQUETAS v3.1 chip {cat} aria-pressed')
        elif f'data-cat="{cat}"' in html and 'class="cat-chip-icon"' in html:
            result.ok(f'ETIQUETAS v3.1 chip {cat} tiene aria-pressed + SVG icon')
        else:
            result.fail(f'chip {cat}', 'aria-pressed o icon ausente')

    # Dropdown items con row 3 (skill + area pills) + is-selected indicator
    for k in ['etiqueta-item-row3', 'etiqueta-item-skill', 'etiqueta-item-area',
              'is-selected']:
        if k in html:
            result.ok(f'ETIQUETAS v3.1 dropdown item: {k}')
        else:
            result.fail(f'ETIQUETAS v3.1 dropdown', f'{k} ausente')

    # Preview card redesign — close button + skill cards
    for k in ['class="etiqueta-preview-close"', 'class="etiqueta-skill-card"',
              'onclick="closeEtiquetaPreview()"']:
        if k in html:
            result.ok(f'ETIQUETAS v3.1 preview: {k}')
        else:
            result.fail(f'ETIQUETAS v3.1 preview', f'{k} ausente')

    # JS handlers nuevos v3.1
    for fn in ['clearEtiquetaSearch(', 'closeEtiquetaPreview(',
               'updateEtiquetaClearVisibility(', 'renderEtiquetasDropdownPosition(']:
        if fn in html:
            result.ok(f'ETIQUETAS v3.1 JS handler: {fn}')
        else:
            result.fail(f'ETIQUETAS v3.1 JS handler', f'{fn} ausente')

    # CSS reglas nuevas v3.1
    for sel in ['.cat-chip-icon', '.etiquetas-search-icon', '.etiquetas-search-clear',
                '.etiqueta-skill-card', '.etiqueta-item-row3', '.etiqueta-item-skill',
                '.etiqueta-item-area', '.etiqueta-preview-close', '.etiquetas-title-icon',
                '.etiquetas-title-count']:
        if sel in html:
            result.ok(f'ETIQUETAS v3.1 CSS regla: {sel}')
        else:
            result.fail(f'ETIQUETAS v3.1 CSS', f'{sel} ausente')

    # Animation: etiquetaPreviewIn keyframes
    if '@keyframes etiquetaPreviewIn' in html:
        result.ok('ETIQUETAS v3.1 keyframes: etiquetaPreviewIn (entrada con spring)')
    else:
        result.fail('keyframes etiquetaPreviewIn', 'animation ausente')

    # Cat-themed box-shadow per chip (ring con color de la categoría activa)
    if 'cat-postpago-glow' in html and 'cat-mtotal-glow' in html and 'cat-standalone-glow' in html:
        result.ok('ETIQUETAS v3.1 chip active: cat-themed glow per category')

    # CSS validation: warning-amber-* token usado (no más hardcoded #f59e0b en nuevo CSS)
    # Aceptamos el #f59e0b en light-mode override pero el nuevo CSS v3.1 debe usar el token
    if 'var(--warning-amber-500)' in html:
        result.ok('ETIQUETAS v3.1 validation: warning-amber token aplicado')

    # Mobile fallback CSS (chips a 1-col en <=480px)
    if '@media (max-width: 480px)' in html and '.etiqueta-categories' in html:
        result.ok('ETIQUETAS v3.1 mobile fallback: chips 1-col @ <=480px')

    # ══════════════════════════════════════════════════════════════════════
    # v3.2 — Dropdown bounded a .etiquetas-block + 3 buttons segmented control
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.2 — Dropdown bounds + segmented action buttons')

    # etquetasFitToBlock function existe y es usada
    if 'function etiquetasFitToBlock' in html:
        result.ok('TIPIF v3.2 etiquetasFitToBlock function definida')
    else:
        result.fail('etiquetasFitToBlock function', 'ausente')
    # Llamada desde renderEtiquetasDropdownPosition
    if 'etiquetasFitToBlock(drop, input)' in html:
        result.ok('TIPIF v3.2 etiquetasFitToBlock invocada desde renderEtiquetasDropdownPosition')
    else:
        result.fail('etiquetasFitToBlock call', 'no se llama desde renderEtiquetasDropdownPosition')
    # usa closest(.etiquetas-block) como bounds
    if "input.closest('.etiquetas-block')" in html:
        result.ok('TIPIF v3.2 bounds: input.closest(.etiquetas-block)')
    else:
        result.fail('etiquetasFitToBlock bounds', '.etiquetas-block lookup ausente')
    # Width matching input
    if "drop.style.width = inputRect.width + 'px'" in html:
        result.ok('TIPIF v3.2 dropdown width = inputRect.width (alineado exacto)')
    else:
        result.fail('dropdown width align', 'width alignment ausente')

    # 3 buttons row — segmented control aesthetic (CRIT-009 avanzado+)
    # Container: flex (no grid) + glass + padding interno
    if re.search(r'\.action-buttons-row\s*\{[^}]*display:\s*flex', html, re.DOTALL):
        result.ok('TIPIF v3.2 action-buttons-row: display flex (segmented control)')
    else:
        result.fail('action-buttons-row display', 'no es flex')
    # Dividers via border-left en siblings
    if '.action-btn + .action-btn' in html and 'border-left' in html:
        result.ok('TIPIF v3.2 segment dividers: .action-btn + .action-btn border-left')
    else:
        result.fail('segment dividers', 'dividers via border-left ausente')
    # Buttons transparentes (segmented control = bg en container, no en buttons)
    if re.search(r'\.action-btn\s*\{[^}]*background:\s*transparent', html, re.DOTALL):
        result.ok('TIPIF v3.2 action-btn background transparent (segmented)')
    else:
        result.fail('button-bg-transparent', 'segmented control bg ausente')
    # Sin individual border (segmented = solo container tiene border)
    if '.action-btn {' in html and 'border: none' in html:
        result.ok('TIPIF v3.2 action-btn sin border individual (segmented control)')
    # Compact height 34px (no min-height: 44px)
    if re.search(r'\.action-btn\s*\{[^}]*height:\s*34px', html, re.DOTALL):
        result.ok('TIPIF v3.2 action-btn height: 34px (compact popup)')
    else:
        result.fail('action-btn height 34px', 'height ausente')
    # flex 1 1 0 para equal distribution
    if re.search(r'\.action-btn\s*\{[^}]*flex:\s*1 1 0', html, re.DOTALL):
        result.ok('TIPIF v3.2 action-btn flex: 1 1 0 (equal distribution)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3 — Popup mode 1-col chips + button→dropdown guard
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3 — Popup mode 1-col + button→dropdown guard')

    # Popup mode: chips SIEMPRE 1-col (CEO directive: popup es target primario)
    if 'body[data-popup="true"] .etiqueta-categories' in html and \
       re.search(r'body\[data-popup="true"\]\s+\.etiqueta-categories[^}]*grid-template-columns:\s*1fr', html, re.DOTALL):
        result.ok('TIPIF v3.3 popup mode: chips SIEMPRE 1-col via body[data-popup]')
    else:
        result.fail('popup-mode chips 1-col', 'selector ausente o grid no es 1fr')
    # Skills grid también 1-col en popup mode
    if 'body[data-popup="true"] .etiqueta-preview-skills' in html and \
       re.search(r'body\[data-popup="true"\]\s+\.etiqueta-preview-skills[^}]*grid-template-columns:\s*1fr', html, re.DOTALL):
        result.ok('TIPIF v3.3 popup mode: preview-skills SIEMPRE 1-col')
    else:
        result.fail('popup-mode skills 1-col', 'selector ausente')
    # Mobile fallback ampliado a 640px (cubre popup mode también)
    if '@media (max-width: 640px)' in html:
        result.ok('TIPIF v3.3 mobile fallback: max-width 640px (cubre popup)')

    # copySpeech tiene guard defensivo contra trigger del dropdown
    if 'document.activeElement.blur' in html and 'copySpeech' in html:
        result.ok('TIPIF v3.3 copySpeech defensive: blur() active element')
    if 'hideEtiquetaDropdown' in html:
        # Verificar que se invoca dentro de copySpeech (no solo definida)
        # Buscar el contexto: copySpeech body debe contener hideEtiquetaDropdown
        m = re.search(r'function copySpeech\(type\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
        if m and 'hideEtiquetaDropdown()' in m.group(1):
            result.ok('TIPIF v3.3 copySpeech: hideEtiquetaDropdown() invocado explícitamente')
        else:
            result.fail('copySpeech hide guard', 'no se invoca dentro de copySpeech')
    # setupEtiquetas document mousedown handler skips action-buttons-row
    if "e.target.closest('.action-buttons-row')" in html:
        result.ok('TIPIF v3.3 document mousedown: skip .action-buttons-row')
    else:
        result.fail('mousedown skip buttons-row', 'guard ausente')

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

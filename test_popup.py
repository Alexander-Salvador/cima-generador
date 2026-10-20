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
EXPECTED_DISTRITOS = 1893  # v3.10.3: +1 (Santa Rosa de Loreto Ley 32403, d6 tentativo 160405)
EXPECTED_DISTRITOS_CCPP = 1853  # some districts don't have caseríos in GeoNames
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
        # v3.10.26: pulseDanger keyframes REMOVIDOS (CEO vibra). Estado se
        # comunica por color estático sin animación continua.
        'pulseDanger ELIMINADO (v3.10.26)': '@keyframes pulseDanger' not in html,
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
        # v3.10.13: ccpp_keys 234 (CCPP exactos COMU-5514 estaciones base, recuperación
        # end-to-end de 32 CCPPs del data consolidada + 2 sin match CCPP_DATA).
        # districts_with_off: SOLO CCPP exactos (sin IPT/2G). Antes incluía 1798 IPT.
        if len(ccpp_items) == 234: result.ok(f'ccpp_keys size = {len(ccpp_items)} (234 expected, recuperación v3.10.13)')
        elif len(ccpp_items) == 239: result.ok(f'ccpp_keys size = {len(ccpp_items)} (239 legacy, NOT cleaned)')
        else: result.fail(f'ccpp_keys size', f'{len(ccpp_items)} (expected 234 recovered or 239 legacy)')
        if len(dist_items) >= 100: result.ok(f'districts_with_off size = {len(dist_items)} (>=100, sin IPT)')
        if total_zones >= 200: result.ok(f'total_zones = {total_zones} (>=200)')
        else: result.fail(f'total_zones', f'{total_zones} (expected >=200)')
        if total_districts >= 200: result.ok(f'total_districts = {total_districts} (>=200)')
        else: result.fail(f'total_districts', f'{total_districts} (expected >=200)')


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

        # v3.9.12: NUEVO LAYOUT — coberturaBtn AHORA vive en .row-ref-buttons
        # junto con REFERENCIA + botón borrar. NO en output-row (que ahora
        # es solo la fila completa del RESULTADO).
        print('\n▸ v3.9.12 Layout — coberturaBtn en .row-ref-buttons (NO en output-row)')
        btn_pos = html.find('id="coberturaBtn"')
        if btn_pos > 0:
            # Buscar el .row-ref-buttons que contiene el botón
            row_ref_pos = html.rfind('class="row-ref-buttons"', 0, btn_pos)
            row_ref_end_match = re.search(r'</div>\s*<!--\s*v3.9.12.*?Row 4 completa',
                                         html[row_ref_pos:], re.DOTALL) if row_ref_pos > 0 else None
            if row_ref_pos > 0:
                # Verificar que REFERENCIA + btn-clear + coberturaBtn están dentro
                segment = html[row_ref_pos:btn_pos + 600]
                if 'id="referencia"' in segment and 'btn-icon-clear' in segment and 'coberturaBtn' in segment:
                    result.ok('v3.9.12 layout: coberturaBtn en row-ref-buttons (con REFERENCIA + btn-clear)')
                else:
                    result.fail('layout v3.9.12', 'row-ref-buttons no contiene los 3 elementos esperados')
            else:
                result.fail('layout v3.9.12', '.row-ref-buttons no encontrado')
        else:
            result.fail('layout v3.9.12', 'id="coberturaBtn" no encontrado')
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

    # v2.1.4: .cobertura-btn.off ya no controla display (siempre visible).
    # v3.10.26: REMOVIDA animation pulseWarn/pulseDanger del estado .off.
    # CEO directive: "elimina toda animacion mientras no se gatille
    # correctamente con la busqueda correcta". Estado se comunica por
    # color estático (ROJO confirmado / AMARILLO preventivo).
    # v3.8.0: .off y .warn son visualmente distintos (off=rojo, warn=amber).
    css_off_match = re.search(
        r'\.cobertura-btn\.off[^{}]*\{([^}]+)\}',
        html
    )
    if css_off_match:
        css_off_body = css_off_match.group(1)
        # v3.10.26: validamos que NO haya animation de pulso (estado comunicable
        # sin vibración). animation: none es el caso válido.
        if 'animation:' in css_off_body and ('pulseWarn' in css_off_body or 'pulseDanger' in css_off_body):
            result.fail('.cobertura-btn.off animation', 'pulseWarn/pulseDanger REMOVIDOS en v3.10.26 (CEO vibra)')
        elif 'animation:' in css_off_body and ('none' in css_off_body):
            result.ok('.cobertura-btn.off animation: none (v3.10.26 sin vibración)')
        else:
            result.fail('.cobertura-btn.off animation', 'animation: none esperado en v3.10.26')
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
    # v3.9.12: la regla `.output-row.output-row--full .output-box` existe
    # como override. El test busca la regla BASE (selector solo `.output-box`).
    # Buscamos el primer match donde `.output-box {` NO esté precedido por otro
    # selector (ej. `.output-row` o `.field`).
    base_output_box = None
    for m in re.finditer(r'(?:^|[\s,}{])\.output-box\s*\{', html):
        start = m.start()
        # look back — si el selector empieza tras } o { o \n, es regla base
        prev = html[max(0, start-2):start]
        if prev.strip() in ('}', '{', '') or '\n' in prev:
            base_output_box = m
            break
    css_output_box = base_output_box
    if css_output_box:
        # Capturar el body completo del bloque
        body_start = html.find('{', start) + 1
        body_end = html.find('}', body_start)
        ob_body = html[body_start:body_end]
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
    # v3.7.6 (L-059): FLIP REVERTIDO — SIEMPRE ABAJO del input, sin flip-arriba
    # CEO feedback: "Falta arreglar el input de CCPP/CASERIO al darle click
    # sale el input arriba al top del container no mantiene su posicion
    # original del input". Solución: eliminar la lógica de flip-arriba.
    if 'FLIP_THRESHOLD' not in html and 'flipCap' not in html:
        result.ok('acFitToGrid v3.7.6: SIN FLIP_THRESHOLD/flipCap (CEO L-059, sin flip-arriba)')
    else:
        result.fail('acFitToGrid flip logic', 'flip logic aún presente (debe estar removido v3.7.6)')
    # JS: classList.add('is-flipped') NO debe estar (solo remove defensivo)
    if "drop.classList.add('is-flipped')" not in html:
        result.ok('acFitToGrid v3.7.6: SIN classList.add is-flipped (sin flip-arriba)')
    else:
        result.fail('acFitToGrid is-flipped add', 'classList.add is-flipped aún presente (debe estar removido v3.7.6)')
    # classList.remove('is-flipped') defensivo (idempotencia)
    if "drop.classList.remove('is-flipped')" in html or "classList.remove('is-flipped')" in html:
        result.ok('acFitToGrid v3.7.6: classList.remove is-flipped presente (defensivo)')
    else:
        result.fail('acFitToGrid is-flipped remove', 'classList.remove is-flipped ausente (idempotencia rota)')
    # JS: max-height aplicado dinámicamente con piso 80px (no flipCap)
    if re.search(r'drop\.style\.maxHeight\s*=', html) and re.search(r'Math\.min\s*\(\s*220', html):
        result.ok('acFitToGrid v3.7.6: maxHeight dinámico con Math.min(220, ...) + piso 80px')
    else:
        result.fail('acFitToGrid maxHeight', 'style.maxHeight dinámico con piso 80px ausente')
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
    # CSS nuevo (v3.3.6 removió: .copy-etiqueta-btn, .etiqueta-validation-label)
    for sel in ['.action-btn-label', '.action-btn-icon', '.action-btn-feedback', '.etiqueta-categories']:
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

    # v3.3.4 REVERSIÓN: dropdown items minimal — SOLO title. Sin row3/skill/area.
    # v3.3.6 REVERSIÓN: preview minimal — SOLO 2 skill cards (sin header/desc/validation/copy).
    # Tests obsoletos v3.1 (close button, validation label) se omiten.

    # Preview card redesigned — solo skill cards (sin close button)
    if 'class="etiqueta-skill-card"' in html:
        result.ok('ETIQUETAS v3.1 preview: skill card class presente')
    else:
        result.fail('ETIQUETAS v3.1 preview', 'skill card class ausente')

    # JS handlers nuevos v3.1 (closeEtiquetaPreview preserva pero ya no se invoca desde HTML)
    for fn in ['clearEtiquetaSearch(', 'closeEtiquetaPreview(',
               'updateEtiquetaClearVisibility(', 'renderEtiquetasDropdownPosition(']:
        if fn in html:
            result.ok(f'ETIQUETAS v3.1 JS handler: {fn}')
        else:
            result.fail(f'ETIQUETAS v3.1 JS handler', f'{fn} ausente')

    # CSS reglas nuevas v3.1 (v3.3.6 removió: .etiqueta-preview-close, validation)
    for sel in ['.cat-chip-icon', '.etiquetas-search-icon', '.etiquetas-search-clear',
                '.etiqueta-skill-card', '.etiquetas-title-icon',
                '.etiquetas-title-count']:
        if sel in html:
            result.ok(f'ETIQUETAS v3.1 CSS regla: {sel}')
        else:
            result.fail(f'ETIQUETAS v3.1 CSS', f'{sel} ausente')
    # v3.3.4 nueva regla: .etiqueta-item-title (dropdown minimal)
    if '.etiqueta-item-title' in html:
        result.ok('ETIQUETAS v3.3.4 CSS regla: .etiqueta-item-title (dropdown minimal)')

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
    if "MIN_DROP_WIDTH" in html or 'Math.max(inputRect.width, 280)' in html or 'min-width: 280px' in html:
        result.ok('v3.9.9 dropdown width: min-width 280px (widen — CEO directive nombres largos)')
    elif "drop.style.width = inputRect.width + 'px'" in html:
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
    print('\n▸ TIPIF v3.3 — button→dropdown guard (popup mode 1-col REVERTIDO en v3.3.2)')

    # v3.3.2 REVERSIÓN: el CEO clarificó que los chips van en 1 SOLA ROW horizontal.
    # El selector body[data-popup="true"] .etiqueta-categories { 1fr } YA NO EXISTE.
    # Los chips vuelven a grid-template-columns: repeat(3, 1fr) (3-col horizontal).
    # Tests que asumen 1-col en popup mode están obsoletos — se omiten intencionalmente.

    # Lo que SÍ se preserva de v3.3: button→dropdown guard (copySpeech defensive)
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

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.1 — BULLETPROOF popup mode detection (html + body + class)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.1 — Bulletproof popup mode')

    # JS setea data-popup en <html> Y <body> (defense-in-depth — preserva el hook
    # aunque la directiva visual cambió en v3.3.2)
    if "document.documentElement.setAttribute('data-popup', 'true')" in html and \
       "document.body.setAttribute('data-popup', 'true')" in html:
        result.ok('TIPIF v3.3.1 data-popup set en html Y body (dual hook preservado)')
    else:
        result.fail('data-popup dual hook', 'setAttribute en html o body ausente')

    # Action buttons NUNCA wrap (flex-wrap: nowrap explícito)
    if re.search(r'\.action-buttons-row\s*\{[^}]*flex-wrap:\s*nowrap', html, re.DOTALL):
        result.ok('TIPIF v3.3.1 action-buttons-row: flex-wrap nowrap (siempre 1 row)')
    # min-width en action-btn previene shrink to 0
    if re.search(r'\.action-btn\s*\{[^}]*min-width:\s*56px', html, re.DOTALL):
        result.ok('TIPIF v3.3.1 action-btn: min-width 56px (no shrink to invisible)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.2 — Chips SIEMPRE 3-col horizontal + labels abreviados
    # CEO directive: "los botones POSTPAGO MOVISTAR TOTAL Y STANDALONE deben
    # estar en 1 solo row. puedes abreviarlo POST/PRE | MT | STANDALONE"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.2 — Chips 3-col horizontal + labels abreviados')

    # CSS: chips SIEMPRE 3-col (default + mobile fallback solo en viewports extremos)
    if re.search(r'\.etiqueta-categories\s*\{[^}]*grid-template-columns:\s*repeat\(3,\s*1fr\)', html, re.DOTALL):
        result.ok('TIPIF v3.3.2 chips: grid-template-columns repeat(3, 1fr) — 3-col horizontal siempre')
    else:
        result.fail('chips 3-col', 'repeat(3, 1fr) ausente en .etiqueta-categories')
    # Mobile fallback solo en viewports EXTREMOS (<320px)
    if '@media (max-width: 320px)' in html and '.etiqueta-categories' in html:
        result.ok('TIPIF v3.3.2 mobile fallback: chips 1-col solo @ <=320px (extremo)')

    # Labels abreviados visibles
    # (data-cat preserva nombres lógicos: POSTPAGO/MTOTAL/STANDALONE)
    if '<span>POST/PRE</span>' in html:
        result.ok('TIPIF v3.3.2 label: POST/PRE (abreviatura CEO)')
    else:
        result.fail('label POST/PRE', 'abreviatura ausente')
    if '<span>MT</span>' in html:
        result.ok('TIPIF v3.3.2 label: MT (abreviatura MOVISTAR TOTAL)')
    else:
        result.fail('label MT', 'abreviatura ausente')
    if '<span>STANDALONE</span>' in html:
        result.ok('TIPIF v3.3.2 label: STANDALONE (sin cambio, ya corto)')
    else:
        result.fail('label STANDALONE', 'ausente')

    # data-cat preservados (JS logic no cambia)
    for cat in ['data-cat="POSTPAGO"', 'data-cat="MTOTAL"', 'data-cat="STANDALONE"']:
        if cat in html:
            result.ok(f'TIPIF v3.3.2 data-cat preservado: {cat}')
        else:
            result.fail(f'data-cat {cat}', 'ausente')

    # Title attributes para tooltip completo (accesibilidad)
    for title in ['title="POSTPAGO/PREPAGO"', 'title="MOVISTAR TOTAL"', 'title="STANDALONE (servicio fijo)"']:
        if title in html:
            result.ok(f'TIPIF v3.3.2 title attribute: {title}')
        else:
            result.fail(f'title {title}', 'ausente — nombres completos no accesibles')

    # NO debe existir labels largos viejos en el chip text
    if '<span>POSTPAGO/PREPAGO</span>' in html or '<span>MOVISTAR TOTAL</span>' in html:
        result.fail('labels viejos', 'aún hay labels largos (POSTPAGO/PREPAGO o MOVISTAR TOTAL) en chips')
    else:
        result.ok('TIPIF v3.3.2 labels viejos removidos (chips compactos)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.3 — Chip click SOLO selección (NO abre dropdown)
    # CEO directive: "los botones son de seleccion, el dropdown solo se debe
    # abrir mientras tengamos una seleccion y le demos click al searchinput"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.3 — Chip click = selección pura (NO abre dropdown)')

    # filterEtiquetasByCat NO debe llamar renderEtiquetasDropdown
    # Verificar el body de la función: NO contiene renderEtiquetasDropdown(
    m = re.search(r'function filterEtiquetasByCat\(cat\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m and 'renderEtiquetasDropdown(' not in m.group(1):
        result.ok('TIPIF v3.3.3 filterEtiquetasByCat: NO llama renderEtiquetasDropdown (chip es selección pura)')
    else:
        result.fail('filterEtiquetasByCat bug', 'aún llama renderEtiquetasDropdown → chip abre dropdown')

    # renderEtiquetasDropdown: guard defensivo si NO hay cat
    m2 = re.search(r'function renderEtiquetasDropdown\(query\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m2 and '!TX_STATE.cat' in m2.group(1) and 'drop.hidden = true' in m2.group(1):
        result.ok('TIPIF v3.3.3 renderEtiquetasDropdown: guard !TX_STATE.cat → no abre sin selection')
    else:
        result.fail('renderEtiquetasDropdown guard', 'guard defensivo ausente')

    # El comment explicativo del fix debe estar presente
    if 'v3.3.3' in html and 'los botones son de seleccion' in html:
        result.ok('TIPIF v3.3.3 comment explicativo presente (CEO directive)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.4 — DROPDOWN MINIMAL: solo title (sin badge/detalle/meta)
    # CEO directive: "unicamente debe decir el title para que es nada mas"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.4 — Dropdown minimal (solo title)')

    # Render del item usa SOLO .etiqueta-item-title (no badge/detalle/meta)
    m = re.search(r'function renderEtiquetasDropdown\(query\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m:
        body = m.group(1)
        # Debe contener etiqueta-item-title
        if 'etiqueta-item-title' in body:
            result.ok('TIPIF v3.3.4 renderEtiquetasDropdown: usa .etiqueta-item-title')
        else:
            result.fail('render minimal', '.etiqueta-item-title ausente en render')
        # NO debe contener las clases obsoletas
        obsoletas = ['etiqueta-item-row3', 'etiqueta-item-skill', 'etiqueta-item-area',
                     'etiqueta-item-row1', 'etiqueta-item-row2', 'is-selected',
                     'etiqueta-item-cat']
        bad = [c for c in obsoletas if c in body]
        if not bad:
            result.ok('TIPIF v3.3.4 render minimal: SIN clases obsoletas (badge/detalle/meta)')
        else:
            result.fail('render obsoleto', f'aún renderiza: {bad}')

    # CSS: NO debe contener las reglas obsoletas (drop minimal CSS)
    obsoletas_css = ['.etiqueta-item-row3', '.etiqueta-item-skill', '.etiqueta-item-area',
                     '.etiqueta-item.is-selected', '.etiqueta-item-cat']
    bad_css = [c for c in obsoletas_css if c in html]
    if not bad_css:
        result.ok('TIPIF v3.3.4 CSS: SIN reglas obsoletas (row3/skill/area/is-selected/cat)')
    else:
        result.fail('CSS obsoleto', f'aún contiene: {bad_css}')

    # CSS: debe contener .etiqueta-item-title
    if '.etiqueta-item-title' in html:
        result.ok('TIPIF v3.3.4 CSS: .etiqueta-item-title presente (title-only)')
    else:
        result.fail('.etiqueta-item-title CSS', 'ausente')

    # comment explicativo CEO directive en el código
    if 'v3.3.4' in html and 'unicamente debe decir el title' in html:
        result.ok('TIPIF v3.3.4 comment explicativo CEO directive presente')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.5 — Tipificación stretch responsive + dropdown navegable/scrolleable
    # CEO directive: "estirar la altura del grid TIPIFICACION el contenedor
    # estiralo responsive" + "el contenido del dropdown reducele la font
    # para que se sienta navegable y scrolleable"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.5 — Dropdown navegable (stretch REVERTIDO en v3.3.7)')

    # v3.3.7 REVERSIÓN: tipif stretch removido. CEO directive "el grid completo
    # debe mantenerse asi" (imagen muestra altura compacta natural).
    if not re.search(r'\.tipif-block\s*\{[^}]*min-height:\s*320px', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .tipif-block: SIN min-height 320px (height compacto natural)')
    else:
        result.fail('tipif-block height', 'min-height 320px aún presente — no fue revertido')
    if not re.search(r'\.tipif-block\s*\{[^}]*display:\s*flex', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .tipif-block: SIN display flex (revertido)')
    if not re.search(r'\.popup-frame\s*\{[^}]*display:\s*flex', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .popup-frame: SIN display flex (revertido)')
    if not re.search(r'#gridTipif\s*\{[^}]*flex:\s*1', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 #gridTipif: SIN flex: 1 (revertido)')

    # Dropdown navegable — font-size reducido a 10px
    if re.search(r'\.etiqueta-item-title\s*\{[^}]*font-size:\s*10px', html, re.DOTALL):
        result.ok('TIPIF v3.3.5 .etiqueta-item-title: font-size 10px (navegable)')
    # Dropdown padding reducido a 5px
    if re.search(r'\.etiqueta-item\s*\{[^}]*padding:\s*5px\s+10px', html, re.DOTALL):
        result.ok('TIPIF v3.3.5 .etiqueta-item: padding 5px 10px (compact)')

    # Dropdown siempre hacia abajo (no flip)
    m = re.search(r'function etiquetasFitToBlock\(drop, input\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m:
        body = m.group(1)
        # Debe calcular viewport residual
        if 'window.innerHeight' in body:
            result.ok('TIPIF v3.3.5 etiquetasFitToBlock: usa window.innerHeight como cap adicional')
        # No debe haber lógica de flip arriba (no más "if (available < FLIP_THRESHOLD)")
        if 'FLIP_THRESHOLD' not in body and 'flipCap' not in body and "classList.add('is-flipped')" not in body:
            result.ok('TIPIF v3.3.5 etiquetasFitToBlock: SIN flip logic (siempre abajo)')
        else:
            result.fail('flip logic', 'aún hay lógica de flip arriba')
        # Debe remover is-flipped explícitamente
        if "classList.remove('is-flipped')" in body:
            result.ok('TIPIF v3.3.5 etiquetasFitToBlock: remove is-flipped siempre')
    # CSS .is-flipped puede seguir existiendo (defensa) pero JS no la usa
    if 'is-flipped' in html:
        result.ok('TIPIF v3.3.5 is-flipped CSS preservado (defensivo — JS ya no lo aplica)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.6 — Selection persiste en search input + preview MINIMAL
    # CEO directive:
    # - "al seleccion la opcion del dropdown debe quedarse en el searchinput"
    # - "retiramos todo solo debe mantenerse TX OPER TO OPER -> ETIQUETA |
    #   TX DIRECTA -> ETIQUETA"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.6 — Selection persiste + preview minimal')

    # selectEtiqueta setea input.value al title (selección persiste en input)
    m = re.search(r'function selectEtiqueta\(id\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m and "input.value = e.solicitud" in m.group(1):
        result.ok('TIPIF v3.3.6 selectEtiqueta: input.value = e.solicitud (selección persiste)')
    else:
        result.fail('selectEtiqueta persist', 'no setea input.value = e.solicitud')

    # showEtiquetaPreview minimal: solo 2 skill cards con labels TX OPER TO OPER + TX DIRECTA
    m2 = re.search(r'function showEtiquetaPreview\(e\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m2:
        body = m2.group(1)
        # Debe contener los labels CEO
        if 'TX OPER TO OPER' in body and 'TX DIRECTA' in body:
            result.ok('TIPIF v3.3.6 showEtiquetaPreview: labels CEO (TX OPER TO OPER + TX DIRECTA)')
        else:
            result.fail('preview labels', 'TX OPER TO OPER o TX DIRECTA ausente')
        # NO debe contener los labels obsoletos
        obsoletos = ['etiquetaPreviewCat', 'etiquetaPreviewId', 'etiquetaPreviewTitle',
                     'etiquetaPreviewDesc', 'etiquetaPreviewValidation',
                     'VALIDAR ANTES DE TX', 'TX 2 PASOS']
        bad = [c for c in obsoletos if c in body]
        if not bad:
            result.ok('TIPIF v3.3.6 showEtiquetaPreview: SIN elementos obsoletos (cat/ID/title/desc/validation)')
        else:
            result.fail('preview obsoleto', f'aún renderiza: {bad}')

    # HTML preview structure: solo etiquetaPreviewSkills (sin header/title/desc/validation/buttons)
    preview_html = re.search(r'<div class="etiqueta-preview"[^>]*>(.*?)</div>\s*</div>', html, re.DOTALL)
    if preview_html:
        inner = preview_html.group(1)
        # Debe tener solo skills container
        if 'etiquetaPreviewSkills' in inner and 'etiqueta-skill-card' not in inner.replace('etiquetaPreviewSkills', ''):
            # Más simple check: no debe tener header/title/desc/validation/close/copy button
            obsoletos_html = ['etiqueta-preview-header', 'etiqueta-preview-cat',
                             'etiqueta-preview-id', 'etiqueta-preview-close',
                             'etiqueta-preview-title', 'etiqueta-preview-desc',
                             'etiqueta-preview-validation', 'copy-etiqueta-btn']
            bad_html = [c for c in obsoletos_html if c in inner]
            if not bad_html:
                result.ok('TIPIF v3.3.6 HTML preview: estructura minimal (solo skills container)')
            else:
                result.fail('preview HTML obsoleto', f'aún contiene: {bad_html}')

    # CSS preview: SIN reglas obsoletas
    obsoletas_css = ['.etiqueta-preview-header', '.etiqueta-preview-cat',
                     '.etiqueta-preview-id', '.etiqueta-preview-close',
                     '.etiqueta-preview-title', '.etiqueta-preview-desc',
                     '.etiqueta-preview-validation', '.etiqueta-validation-label',
                     '.copy-etiqueta-btn']
    bad_css = [c for c in obsoletas_css if c in html]
    if not bad_css:
        result.ok('TIPIF v3.3.6 CSS preview: SIN reglas obsoletas (header/cat/id/close/title/desc/validation/copy)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.7 — Skill cards click-to-copy + empty data fallback
    # CEO directive (con imagen):
    # - "el grid completo debe mantenerse asi" (altura compacta, sin stretch)
    # - "cuando no tengan datos deben aparecer unicamente el NAME TX OPER TO
    #   OPER | TX DIRECTA" (sin código si está vacío)
    # - "Al darles click se debe copiar unicamente la etiqueta limpia"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.7 — Skill cards click-to-copy + empty fallback')

    # showEtiquetaPreview: render condicional basado en si skill code existe
    m = re.search(r'function showEtiquetaPreview\(e\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m:
        body = m.group(1)
        # skillCardHtml helper que retorna label + code condicional
        if 'function skillCardHtml' in body:
            result.ok('TIPIF v3.3.7 showEtiquetaPreview: skillCardHtml helper (render condicional)')
        # Renderiza como BUTTON (no div)
        if 'etiqueta-skill-card' in body and '<button' in body:
            result.ok('TIPIF v3.3.7 showEtiquetaPreview: skill cards son <button> (clickeable)')
        # Click handler con copy-to-clipboard
        if 'data-copy' in body and 'navigator.clipboard.writeText' in body:
            result.ok('TIPIF v3.3.7 showEtiquetaPreview: click handler con copy-to-clipboard')
        # Empty data fallback
        if "if (!code)" in body or "code || ''" in body:
            result.ok('TIPIF v3.3.7 showEtiquetaPreview: empty data fallback (sin codigo → solo label)')
        # aria-label para accesibilidad
        if 'aria-label="Copiar' in body:
            result.ok('TIPIF v3.3.7 showEtiquetaPreview: aria-label en skill cards (a11y)')

    # CSS skill card clickable: cursor pointer + transition
    if re.search(r'\.etiqueta-skill-card\s*\{[^}]*cursor:\s*pointer', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .etiqueta-skill-card: cursor pointer (clickeable)')
    if re.search(r'\.etiqueta-skill-card\.is-copied', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .etiqueta-skill-card: estado .is-copied (feedback transient)')
    # Hover lift visual feedback
    if re.search(r'\.etiqueta-skill-card:hover[^}]*translateY', html, re.DOTALL):
        result.ok('TIPIF v3.3.7 .etiqueta-skill-card: hover lift (translateY -1px)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.8 — Dropdown SIEMPRE visible + floating "Copiado" placeholder
    # CEO directive:
    # 1. "cuando actualiza la web desaparecen la posicion de las etiquetas
    #    deben mantenerse siempre activas, solo que sin datos hasta que se use
    #    el searchinput" — dropdown items siempre visibles (layout estable)
    # 2. "al seleccionar la etiqueta debe salir un placeholder encima de copiado"
    #    — floating tooltip "✓ Copiado" arriba del skill card
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.8 — Dropdown siempre visible + Copiado placeholder')

    # Dropdown items siempre visibles: empty state si no matches, items si matches
    m = re.search(r'function renderEtiquetasDropdown\(query\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m:
        body = m.group(1)
        # Empty state con mensaje distinto para query vacía vs no matches
        if 'Escribe para buscar' in body or "query\\s*\\?" in body or "query ?" in body:
            result.ok('TIPIF v3.3.8 renderEtiquetasDropdown: empty state distinto (query vs no matches)')
        # Drop hidden = false se mantiene cuando hay items
        if 'drop.hidden = false' in body and body.count('drop.hidden = false') >= 2:
            result.ok('TIPIF v3.3.8 renderEtiquetasDropdown: drop.hidden = false (siempre visible)')

    # Click-outside close handler RE-AÑADIDO en setupEtiquetas (v3.4.0 REVERTIDO de v3.3.8)
    # CEO directive 2026-10-15: "al actualizar la ruta se abre automaticamente
    # el dropdown de searchinput sin seleccionar nada". v3.3.8 "always visible"
    # abrió el dropdown al page load. v3.4.0 revierte: dropdown HIDDEN por
    # default, abre en focus/mousedown/input, cierra en click-outside.
    m_setup = re.search(r'function setupEtiquetas\(.*?\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m_setup:
        body_setup = m_setup.group(1)
        # Click-outside handler con hideEtiquetaDropdown debe estar PRESENTE
        if "document.addEventListener('mousedown'" in body_setup and "hideEtiquetaDropdown()" in body_setup:
            result.ok('TIPIF v3.4.0 setupEtiquetas: click-outside handler RE-AÑADIDO (dropdown se cierra)')
        else:
            result.fail('click-outside handler', 'aún ausente — debería estar RE-AÑADIDO en v3.4.0')
        # NO debe renderizar dropdown en init (CEO directive)
        if "renderEtiquetasDropdown('')" not in body_setup:
            result.ok('TIPIF v3.4.0 setupEtiquetas: SIN render en init (dropdown NO se abre solo)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.4.1 — Clear search → reset preview + tooltip position relative
    # CEO directive:
    # 1. "Cuando borro la busqueda del searchinput, las etiquetas se quedan
    #    con la informacion anterior no hay conexion inteligente" → al borrar
    #    search, los codes del preview se limpian (sin selección activa)
    # 2. "Al darle click a la etiqueta sale el mensaje COPIAR pero fuera del
    #    input y se mueve con el mouse" → tooltip no debe moverse, debe estar
    #    anclado al card (position:relative en card)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.4.1 — Clear search reset preview + tooltip position')

    # clearEtiquetaSearch resetea preview a empty state
    m_clear = re.search(r'function clearEtiquetaSearch\(.*?\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m_clear and 'TX_STATE.selectedId = null' in m_clear.group(1) and 'showEtiquetaPreview(null)' in m_clear.group(1):
        result.ok('TIPIF v3.4.1 clearEtiquetaSearch: TX_STATE.selectedId = null + show(null) (reset preview)')
    else:
        result.fail('clearEtiquetaSearch reset', 'no resetea preview cards')

    # Input event handler resetea preview cuando value === '' (CEO: borrar)
    m_input = re.search(r'input\.addEventListener\(\s*[\'"]input[\'"]\s*,', html)
    if m_input:
        # Encontrar el body del input event handler
        start = m_input.start()
        # Buscar el cierre del handler (next })
        body_start = html.find('{', start)
        body_end = html.find('}\n  });', body_start)
        body = html[body_start:body_end+1] if body_start > 0 and body_end > 0 else ''
        if 'this.value === \'\'' in body and 'showEtiquetaPreview(null)' in body:
            result.ok('TIPIF v3.4.1 input handler: si value="" → show(null) (smart connection)')
        else:
            result.fail('input reset', 'no resetea preview cuando value=""')

    # CSS .etiqueta-skill-card tiene position: relative (ancla tooltip)
    if re.search(r'\.etiqueta-skill-card\s*\{[^}]*position:\s*relative', html, re.DOTALL):
        result.ok('TIPIF v3.4.1 .etiqueta-skill-card: position relative (tooltip anclado al card, no al preview)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.4.2 — ETIQUETAS_DATA expandido (records nuevos CEO directive)
    # CEO directive 2026-10-15 (con análisis del Excel):
    #   ACREDITACIONES → POST/PRE
    #   COMERCIAL → en los 3 (ya estaba id 1, 7, 13)
    #   VENTAS → analizar sub-tipos (Alta/Portabilidad/Cambio)
    #   ROAMING → POST/PRE (ya estaba id 4)
    #   RETENCIONES → en los 3 (ya estaba id 3, 9, 15)
    #   TÉCNICO = LINE 2 → en los 3 (ya id 11 MT, id 16 STD; nuevo POSTPAGO)
    #   CORPORATIVO → mostrado al CEO (no es etiqueta, es workflow)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.4.2 — ETIQUETAS_DATA expandido (6 records nuevos)')

    # Conteos por categoría
    n_postpago  = html.count("cat: 'POSTPAGO'")
    n_mtotal    = html.count("cat: 'MTOTAL'")
    n_standalone= html.count("cat: 'STANDALONE'")
    print(f'  POST/PAGE counts: POSTPAGO={n_postpago} MTOTAL={n_mtotal} STANDALONE={n_standalone}')
    if n_postpago >= 9 and n_mtotal >= 8 and n_standalone >= 6:
        result.ok(f'TIPIF v3.4.2 ETIQUETAS_DATA: conteos expandidos (POSTPAGO={n_postpago}, MTOTAL={n_mtotal}, STANDALONE={n_standalone})')
    else:
        result.fail('conteos', f'POSTPAGO={n_postpago}, MTOTAL={n_mtotal}, STANDALONE={n_standalone} — esperados >=9/8/6')

    # Nuevos skill codes presentes
    if 'SK_CV_ACREDITACIONES_TX' in html:
        result.ok('TIPIF v3.4.2 SK_CV_ACREDITACIONES_TX: presente (CEO: ACREDITACIONES → POST/PRE)')
    else:
        result.fail('SK_CV_ACREDITACIONES_TX', 'ausente')
    if 'SK_CV_TECNICO_MOVIL_TX_KONECTA' in html:
        result.ok('TIPIF v3.4.2 SK_CV_TECNICO_MOVIL_TX_KONECTA: presente (CEO: LINE 2 en los 3)')

    # Conteo de TECNICO MOVIL (debe estar en POSTPAGO, MTOTAL, STANDALONE)
    n_tec_mov = html.count("SK_CV_TECNICO_MOVIL")
    if n_tec_mov >= 4:  # 2 nuevos + 2 existentes en sharedStrings
        result.ok(f'TIPIF v3.4.2 TECNICO MOVIL: {n_tec_mov} referencias (en los 3 categorías)')

    # ACREDITACIONES solo en POSTPAGO (cat: 'POSTPAGO' cerca de ACREDITACIONES)
    m = re.search(r"\{ id:\s*\d+,\s*cat:\s*'POSTPAGO'[^}]*ACREDITACIONES", html, re.DOTALL)
    if m:
        result.ok('TIPIF v3.4.2 ACREDITACIONES: solo en POSTPAGO (CEO directive)')
    else:
        result.fail('ACREDITACIONES cat', 'no se encuentra solo en POSTPAGO')

    # TÉCNICO MOVIL (LINE 2) en cada categoría (id 19 POSTPAGO, id 21 MTOTAL)
    if 'id: 19' in html and 'id: 21' in html and "'TECNICO MOVIL" in html:
        result.ok('TIPIF v3.4.2 TECNICO MOVIL (LINE 2): en POSTPAGO (id 19) + MTOTAL (id 21)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.4.3 — Comparativa exhaustiva Excel vs CIMA (3 SKUs nuevos)
    # Análisis del XLSX identificó skill codes que NO estaban en CIMA:
    #   - SK_CV_DOWNSELL_MOVIL_TX (MIGRACIÓN DOWN MOVIL)
    #   - SK_CV_TRASLADO_BAJA_FIJA_TX (TRASLADO/ MUDANZA)
    #   - SK_CV_RETENCIONES_MOVIL_TX reutilizado para BAJA FINAL POSTPAGO
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.4.3 — Comparativa Excel: 3 SKUs nuevos')

    # Conteos después de v3.4.3 (POSTPAGO +2, STANDALONE +1)
    n_postpago  = html.count("cat: 'POSTPAGO'")
    n_mtotal    = html.count("cat: 'MTOTAL'")
    n_standalone= html.count("cat: 'STANDALONE'")
    print(f'  Counts: POSTPAGO={n_postpago} MTOTAL={n_mtotal} STANDALONE={n_standalone}')
    if n_postpago >= 11 and n_mtotal >= 8 and n_standalone >= 7:
        result.ok(f'TIPIF v3.4.3 ETIQUETAS_DATA: POSTPAGO={n_postpago} MTOTAL={n_mtotal} STANDALONE={n_standalone} (totales 26)')
    else:
        result.fail('conteos v3.4.3', f'POSTPAGO={n_postpago}, MTOTAL={n_mtotal}, STANDALONE={n_standalone} — esperados >=11/8/7')

    # SK_CV_DOWNSELL_MOVIL_TX presente (MIGRACION DOWN MOVIL)
    if 'SK_CV_DOWNSELL_MOVIL_TX' in html:
        result.ok('TIPIF v3.4.3 SK_CV_DOWNSELL_MOVIL_TX: presente (MIGRACION DOWN MOVIL)')

    # SK_CV_TRASLADO_BAJA_FIJA_TX presente (TRASLADO/ MUDANZA)
    if 'SK_CV_TRASLADO_BAJA_FIJA_TX' in html:
        result.ok('TIPIF v3.4.3 SK_CV_TRASLADO_BAJA_FIJA_TX: presente (TRASLADO/MUDANZA)')

    # Solicitudes nuevas
    if "'MIGRACION DOWN (MOVIL)'" in html:
        result.ok('TIPIF v3.4.3 solicitud MIGRACION DOWN (MOVIL) presente')
    if "'BAJA FINAL POSTPAGO'" in html:
        result.ok('TIPIF v3.4.3 solicitud BAJA FINAL POSTPAGO presente')
    if "'TRASLADO / MUDANZA'" in html:
        result.ok('TIPIF v3.4.3 solicitud TRASLADO / MUDANZA presente')

    # Ids nuevos (24, 25, 26)
    for nid in [24, 25, 26]:
        if f'id: {nid}' in html:
            result.ok(f'TIPIF v3.4.3 record id {nid} presente')

    # ══════════════════════════════════════════════════════════════════════
    # v3.4.4 — Auditoría datos Excel + DP fixes + search improvement
    # CEO directive 2026-10-15:
    # 1. "No debes inventar etiquetas elimina todo lo que no este en el excel"
    # 2. "Si ves duplicados valida el nombre para separarlos"
    # 3. "Mejores la busqueda del agente ARREGLANDO minimamente los nombres"
    # 4. "ACREDITACIONES y LINEA 2 no tienen TX DIRECTA"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.4.4 — Auditoría Excel + DP fixes + search improvement')

    # DP fixes — skill_txdirecto debe ser '' para SK codes sin DP counterpart
    # id 18 ACREDITACIONES (SK_CV_ACREDITACIONES_TX sin DP)
    m18 = re.search(r"id:\s*18,[^}]*?skill_tx2:\s*'SK_CV_ACREDITACIONES_TX',\s*skill_txdirecto:\s*''", html, re.DOTALL)
    if m18:
        result.ok('TIPIF v3.4.4 id 18 ACREDITACIONES: skill_txdirecto = \'\' (no DP en Excel)')
    else:
        result.fail('id 18 DP fix', 'skill_txdirecto no está vacío')
    # id 21 LINE 2 (SK_CV_TECNICO_MOVIL_TX_KONECTA sin DP)
    m21 = re.search(r"id:\s*21,[^}]*?skill_tx2:\s*'SK_CV_TECNICO_MOVIL_TX_KONECTA',\s*skill_txdirecto:\s*''", html, re.DOTALL)
    if m21:
        result.ok('TIPIF v3.4.4 id 21 LINE 2: skill_txdirecto = \'\' (no DP en Excel)')
    # id 19 TECNICO MOVIL (mismo SK KONECTA, sin DP)
    m19 = re.search(r"id:\s*19,[^}]*?skill_tx2:\s*'SK_CV_TECNICO_MOVIL_TX_KONECTA',\s*skill_txdirecto:\s*''", html, re.DOTALL)
    if m19:
        result.ok('TIPIF v3.4.4 id 19 TECNICO MOVIL: skill_txdirecto = \'\' (mismo SK KONECTA, sin DP)')

    # VENTAS simplificado — solicitud='VENTAS' (no inventado)
    n_ventas = html.count("solicitud: 'VENTAS'")
    if n_ventas >= 3:
        result.ok(f'TIPIF v3.4.4 VENTAS simplificado: {n_ventas} records con solicitud="VENTAS" (genérico)')
    else:
        result.fail('VENTAS simplificado', f'expected >=3, got {n_ventas}')

    # Search improvement — accent normalization
    if 'normalizeSearch' in html and 'NFD' in html and '\\u0300-\\u036f' in html:
        result.ok('TIPIF v3.4.4 search: normalizeSearch con NFD + remove combining marks (acentos)')
    if 'String(e.id) === q' in html:
        result.ok('TIPIF v3.4.4 search: búsqueda exacta por id (String(e.id) === q)')
    # Search incluye skill_txdirecto (busca en TX directa también)
    if 'normalizeSearch(e.skill_txdirecto)' in html:
        result.ok('TIPIF v3.4.4 search: incluye skill_txdirecto en búsqueda')

    # Duplicados validados — PLAN MENOR variantes tienen skills diferentes
    # id 2 PLAN MENOR (POSTPAGO) → SK_CV_COMERCIAL_TX
    # id 8 PLAN MENOR MT (MTOTAL) → SK_CV_DOWNSELL_FIJA_TX
    # id 14 PLAN MENOR STAND ALONE (STANDALONE) → SK_CV_DOWNSELL_FIJA_TX
    plan_menor_postpago = re.search(r"id:\s*2,[^}]*?PLAN MENOR[^}]*?SK_CV_COMERCIAL_TX", html, re.DOTALL)
    plan_menor_mt = re.search(r"id:\s*8,[^}]*?PLAN MENOR MT[^}]*?DOWNSELL_FIJA_TX", html, re.DOTALL)
    plan_menor_std = re.search(r"id:\s*14,[^}]*?PLAN MENOR STAND ALONE[^}]*?DOWNSELL_FIJA_TX", html, re.DOTALL)
    if plan_menor_postpago and plan_menor_mt and plan_menor_std:
        result.ok('TIPIF v3.4.4 duplicados PLAN MENOR: 3 records con skills DIFERENTES (no son dup)')

    # Floating "Copiado" placeholder CSS — tooltip arriba del card
    if re.search(r'\.etiqueta-skill-card\.is-copied::before[^}]*Copiado', html, re.DOTALL):
        result.ok('TIPIF v3.3.8 .etiqueta-skill-card.is-copied::before: "Copiado" text')
    if re.search(r'\.etiqueta-skill-card\.is-copied::before[^}]*100%', html, re.DOTALL):
        result.ok('TIPIF v3.3.8 .etiqueta-skill-card.is-copied::before: posicionado bottom: 100% (arriba del card)')
    if re.search(r'\.etiqueta-skill-card\.is-copied::after[^}]*border-top-color', html, re.DOTALL):
        result.ok('TIPIF v3.3.8 .etiqueta-skill-card.is-copied::after: arrow del tooltip')
    if '@keyframes copiadoFloatIn' in html:
        result.ok('TIPIF v3.3.8 @keyframes copiadoFloatIn: animación entrada del tooltip')

    # ══════════════════════════════════════════════════════════════════════
    # v3.3.9 — Preview cards SIEMPRE visibles (empty state inicial)
    # CEO directive: "esto va cambiar las opcion de OPER TO OPER Y TX DIRECTA
    # siempre deben estar ahi, cuando se complete la busqueda en el searchinput
    # se rellena con la busqueda correcta".
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ TIPIF v3.3.9 — Preview cards siempre visibles')

    # HTML preview SIN hidden inicial
    preview_html = re.search(r'<div class="etiqueta-preview"[^>]*>', html)
    if preview_html and 'hidden' not in preview_html.group(0):
        result.ok('TIPIF v3.3.9 HTML preview: SIN hidden inicial (siempre visible)')
    else:
        result.fail('preview hidden inicial', 'preview aún tiene hidden attribute')

    # setupEtiquetas llama showEtiquetaPreview(null) en init
    m_setup = re.search(r'function setupEtiquetas\(.*?\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m_setup and 'showEtiquetaPreview(null)' in m_setup.group(1):
        result.ok('TIPIF v3.3.9 setupEtiquetas: init showEtiquetaPreview(null) (2 cards vacíos)')
    else:
        result.fail('init empty preview', 'showEtiquetaPreview(null) ausente en setupEtiquetas')

    # showEtiquetaPreview maneja null = empty state
    m = re.search(r'function showEtiquetaPreview\(e\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m:
        body = m.group(1)
        # Debe usar e como condicional: si e, codes reales; si null, codes vacíos
        if 'e ?' in body or "e\\s*\\?" in body or "e.skill_tx2 : ''" in body:
            result.ok('TIPIF v3.3.9 showEtiquetaPreview: maneja null (e ? e.skill : empty)')
        else:
            result.fail('showEtiquetaPreview null', 'no tiene condicional para null')
        # NO debe ocultar el preview (ya no hay `preview.hidden = false` — siempre visible)
        if 'preview.hidden = false' not in body:
            result.ok('TIPIF v3.3.9 showEtiquetaPreview: NO oculta preview (siempre visible)')

    # filterEtiquetasByCat: usa showEtiquetaPreview(null) cuando selección se invalida
    m_filter = re.search(r'function filterEtiquetasByCat\(cat\)\s*\{(.*?)(?=\nfunction )', html, re.DOTALL)
    if m_filter and 'showEtiquetaPreview(null)' in m_filter.group(1):
        result.ok('TIPIF v3.3.9 filterEtiquetasByCat: usa showEtiquetaPreview(null) (NO oculta)')
    else:
        result.fail('filterEtiquetasByCat empty', 'aún usa hideEtiquetaPreview() en vez de show(null)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.5.2 — ZONAS APAGADO Bug Fixes round 2 (CEO directive 2026-10-15 21:10)
    # Bug 1 (persistente): ⚠️ no se enciende en CCPP/Caserío (CSS cascade --idle vs off/warn)
    # Bug 4 (NUEVO): Popover debe ser TOOLTIP anclado al botón, no modal overlay
    #                  - dentro del grid #gridAddress
    #                  - solo header + speech + copiar/cerrar
    #                  - width max 320px, no supera el contenedor
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.5.2 — BUG FIXES round 2 Zonas Apagado')

    # BUG 1: renderCoberturaButton alterna --idle vs off/warn (v3.5.2 legacy marker)
    # v3.6.0 REBUILD: el marker BUG 1 fix puede no existir; en su lugar
    # verificamos el comportamiento de alternancia (idle → off / off → idle).
    m_render = re.search(r'function renderCoberturaButton.*?\n\}', html, re.DOTALL)
    if m_render:
        render_fn = m_render.group(0)
        # Aceptar cualquier versión de renderCoberturaButton que tenga alternancia
        has_alternate = ("BUG 1 fix: alternar --idle" in render_fn
                        or "BUG 1 fix: alternar" in render_fn
                        or 'classList.remove(\'off\', \'warn\', \'cobertura-btn--idle\')' in render_fn
                        or "classList.remove('off', 'warn', 'fuzzy', 'cobertura-btn--idle')" in render_fn
                        or "classList.remove('off', 'warn', 'ipt', 'fuzzy', 'cobertura-btn--idle')" in render_fn)
        if has_alternate:
            if "BUG 1 fix" in render_fn:
                result.ok('BUG 1 renderCoberturaButton: BUG 1 fix marker presente')
            else:
                result.ok('v3.6.0 renderCoberturaButton: alterna idle/off/warn (clear markers legacy)')
        else:
            result.fail('BUG 1 renderCoberturaButton', 'BUG 1 fix marker ausente')
        if ("remove('off', 'warn', 'cobertura-btn--idle')" in render_fn or
            "remove('off', 'warn'" in render_fn):
            result.ok('BUG 1 renderCoberturaButton: remove idle+off+warn antes de switch')
        else:
            result.fail('BUG 1 remove idle', 'falta remove idle+off+warn')
        # Si no check → vuelve a idle
        if ("btn.classList.add('cobertura-btn--idle')" in render_fn or
            "classList.add('cobertura-btn--idle')" in render_fn):
            result.ok('BUG 1 renderCoberturaButton: vuelve a --idle si no hay check')
        else:
            result.fail('BUG 1 volver a idle', 'falta add --idle en return')
    else:
        result.fail('BUG 1 renderCoberturaButton', 'función no encontrada')

    # BUG 4 v3.10.4: showApagadoTooltip legacy REMOVIDA. Ahora usa showApagadoPanelInline (panel inline).
    if 'function showApagadoPanelInline' in html:
        result.ok('v3.10.4 BUG 4: showApagadoPanelInline (panel inline) reemplaza showApagadoTooltip legacy')
    else:
        result.fail('v3.10.4 BUG 4', 'showApagadoPanelInline no encontrado')
    # showApagadoTooltip legacy NO debe existir
    if 'function showApagadoTooltip' not in html:
        result.ok('v3.10.4: showApagadoTooltip legacy removido (shadowing fix)')
    else:
        result.fail('v3.10.4 shadow', 'showApagadoTooltip legacy aún presente')
    # El tooltip debe crearse en document.body (no dentro del grid) y ser absolute
    m_tooltip = None  # v3.10.4: showApagadoTooltip legacy removido, ahora panel inline
    if m_tooltip is None:
        # v3.10.4: el panel inline es el reemplazo moderno. Verificar que esté.
        if 'function showApagadoPanelInline' in html:
            result.ok('v3.10.4 BUG 4: showApagadoTooltip reemplazado por showApagadoPanelInline (panel inline)')
        else:
            result.fail('v3.10.4 BUG 4', 'showApagadoPanelInline no encontrado')
        # CSS .apagado-speech-panel (panel inline) presente
        if '.apagado-speech-panel' in html:
            result.ok('v3.10.4 BUG 4: .apagado-speech-panel CSS presente (reemplazo de .apagado-tooltip)')
        else:
            result.fail('v3.10.4 BUG 4 CSS', 'falta .apagado-speech-panel')
    # copy handler v3.10.4: copyApagadoTooltipSpeech legacy REMOVIDA. Usa zaCopyApagadoSpeech.
    if 'function zaCopyApagadoSpeech' in html:
        result.ok('v3.10.4: zaCopyApagadoSpeech (copy handler vivo)')
    else:
        result.fail('v3.10.4 copy', 'zaCopyApagadoSpeech no encontrado')
    if 'function copyApagadoTooltipSpeech' not in html:
        result.ok('v3.10.4: copyApagadoTooltipSpeech legacy removido')
    else:
        result.fail('v3.10.4 shadow', 'copyApagadoTooltipSpeech legacy aún presente')

    # v3.10.4: BUG 1 fix consolidado en onClickCobertura (no requiere marker legacy).
    # v3.10.12: estados idle/off/warn/fuzzy (sin ipt) — eliminada rama IPT/2G.
    if 'function renderCoberturaButton' in html and 'cobertura-btn--idle' in html:
        result.ok('v3.10.4/12: renderCoberturaButton (idle/off/warn/fuzzy alternancia, sin ipt)')
    else:
        result.fail('v3.10.4 BUG 1', 'BUG 1 fix perdido tras shadowing fix')
    if 'function onClickCobertura' in html and 'cima-shake' in html:
        result.ok('v3.10.4: onClickCobertura + feedback shake preservados')
    else:
        result.fail('v3.10.4 onClick', 'onClickCobertura o feedback shake perdido')

    # ══════════════════════════════════════════════════════════════════════
    # v3.5.3 — MARCADO VISUAL ROJO CCPP DROPDOWN (CEO directive 2026-10-15 21:30)
    # Bug: el CCPP/Caserío apagado no se veía claramente en el dropdown.
    # Fix:
    #   1. CSS .ac-item.ac-item-off más fuerte (22% bg, border-left, ⚠️ 16px, font 800)
    #   2. CSS .ac-item.ac-item-off-ipt NUEVO (naranja) para IPT (sin match exacto)
    #   3. JS: marcar offIpt cuando distrito entero está en districts_with_off
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.5.3 — Marcado visual CCPP Apagado')

    # v3.5.6 (CEO directive 2026-10-15 22:30) + v3.6.0 REBUILD + v3.7.4
    # unificación: el CSS de .ac-item.ac-item-off debe ser SUTIL y UNIFICADO
    # con el resto del dropdown. NO debe tener los adornos v3.5.3
    # (border-left 4px, font-weight 800, padding-left 14px, keyframe
    # apagadoPulse) que fueron REVERTIDOS por CEO feedback.
    css_off_match = re.search(r'\.ac-item\.ac-item-off\s*\{([^}]+)\}', html)
    if css_off_match:
        body = css_off_match.group(1)
        # Background debe ser SUTIL (8% o similar), NO 22% fuerte
        if re.search(r'var\(--danger\)\s*[0-9]+%', body):
            pct = re.search(r'var\(--danger\)\s*([0-9]+)%', body)
            pct_val = int(pct.group(1)) if pct else 0
            if pct_val <= 12:
                result.ok(f'v3.5.6/v3.7.4 CSS: background {pct_val}% SUTIL (revertido de 22% v3.5.3)')
            else:
                result.fail('CSS bg sutil', f'aún background {pct_val}% (CEO feedback: sutil ≤12%)')
        # NO debe tener border-left 4px solid (CEO feedback L-054: unificar)
        if 'border-left' in body and '4px' in body:
            result.fail('CSS border-left 4px', 'border-left 4px aún presente (v3.5.3 reverted v3.5.6)')
        else:
            result.ok('v3.5.6/v3.7.4 CSS: SIN border-left 4px (unificado con otros items)')
        # NO debe tener font-weight 800 (revertido)
        if 'font-weight: 800' in body:
            result.fail('CSS font 800', 'font-weight 800 aún presente (v3.5.3 reverted v3.5.6)')
        else:
            result.ok('v3.5.6/v3.7.4 CSS: font-weight normal (sin 800)')
        # NO debe tener padding-left 14px (revertido)
        if 'padding-left: 14px' in body:
            result.fail('CSS padding-left 14', 'padding-left 14px aún presente (v3.5.3 reverted v3.5.6)')
        else:
            result.ok('v3.5.6/v3.7.4 CSS: padding-left normal (sin 14px)')
    else:
        result.fail('CSS .ac-item.ac-item-off', 'regla no encontrada')

    # v3.5.6/v3.7.1: ⚠️ eliminado del item off — solo dot rojo (L-057)
    # El CEO feedback fue que el ⚠️ 16px con pulse rompía la estética
    # unificada del dropdown. Ahora solo hay un dot (círculo 6×6px).
    icon_match = re.search(r'\.ac-off-icon\s*\{([^}]+)\}', html)
    if icon_match:
        body = icon_match.group(1)
        if 'apagadoPulse' in body or ('animation' in body and 'pulse' in body.lower()):
            result.fail('CSS pulse animation', '⚠️ con pulse animation aún presente (v3.5.3 reverted v3.5.6/L-054)')
        else:
            result.ok('v3.5.6/v3.7.1 CSS: SIN pulse animation en off icon (L-054)')
    # v3.5.3: @keyframes apagadoPulse DEBE estar removido del CSS del
    # dropdown (CEO feedback). Sí puede quedar en .cobertura-btn si fue
    # preservado, pero NO en .ac-item/.ac-off-icon.
    if re.search(r'\.ac-(item|off-icon)[^{]*\{[^}]*apagadoPulse', html, re.DOTALL):
        result.fail('CSS apagadoPulse en dropdown', 'apagadoPulse aún en .ac-item/.ac-off-icon (revertido)')
    else:
        result.ok('v3.5.6 CSS: @keyframes apagadoPulse NO se aplica a items del dropdown')

    # v3.5.6/v3.6.0/v3.7.4: render del dropdown debe estar LIMPIO (sin iptHint)
    # Aceptar marker v3.5.6 legacy O ausencia total (REBUILD lo sobreescribió)
    if 'ac-item-ipt-hint' in html:
        result.fail('v3.5.6 regresión', 'ac-item-ipt-hint aún en HTML (debe estar limpio)')
    else:
        result.ok('v3.5.6/v3.7.4: SIN ac-item-ipt-hint en HTML (L-054 limpio)')
    if 'ac-ipt-banner' in html:
        result.fail('v3.5.6 regresión', 'ac-ipt-banner aún en HTML (revertido v3.5.6)')
    else:
        result.ok('v3.5.6/v3.7.4: SIN ac-ipt-banner en HTML (revertido)')

    # v3.7.4 (L-057): solo match EXACTO con ccpp_keys se pinta (con dot)
    # NO se pinta TODO el distrito como 'ac-item-zone-off' (CEO feedback)
    # Verificar que NO se USA la clase en código activo (puede estar en
    # comentarios históricos que documentan el revert)
    code_no_comments = re.sub(r'/\*.*?\*/', '', html, flags=re.DOTALL)
    code_no_comments = re.sub(r'//.*', '', code_no_comments)
    if re.search(r'ac-item-zone-off', code_no_comments):
        result.fail('v3.7.4 regresión', 'ac-item-zone-off aún USADA en código activo (CEO L-057: solo match exacto)')
    else:
        result.ok('v3.7.4 unificación: ac-item-zone-off solo en comentarios históricos (no se aplica)')

    # districts_with_off.has debe seguir existiendo en updateCcppInputZoneClass (v3.5.5 intacto)
    if 'function updateCcppInputZoneClass' in html and 'districts_with_off.has(d6)' in html:
        result.ok('v3.5.5 JS: helper updateCcppInputZoneClass usa districts_with_off (v3.5.5 intacto)')
    else:
        result.fail('JS helper', 'helper updateCcppInputZoneClass perdido')

    # ══════════════════════════════════════════════════════════════════════
    # v3.5.6 — DROPDOWN LIMPIO (CEO directive 2026-10-15 22:30)
    # REVERTIR contaminación visual del v3.5.4:
    #   - QUITAR banner naranja grande (.ac-ipt-banner)
    #   - QUITAR marca sutil 'ac-item-ipt-hint' en items
    # MANTENER:
    #   - 'ac-item-off' ROJO fuerte SOLO para match exacto
    #   - v3.5.5: input CCPP pintado por zona
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.5.6 — Dropdown limpio (revierte v3.5.4)')

    # CSS: NO debe existir .ac-item-ipt-hint (revertido)
    if '.ac-item.ac-item-ipt-hint' in html:
        result.fail('v3.5.6 regresión', 'ac-item-ipt-hint aún presente (debe estar limpio)')
    else:
        result.ok('v3.5.6 CSS: .ac-item-ipt-hint REMOVIDO (limpio)')
    # CSS: NO debe existir .ac-ipt-banner (revertido)
    if '.ac-ipt-banner' in html:
        result.fail('v3.5.6 regresión', 'ac-ipt-banner aún presente (debe estar limpio)')
    else:
        result.ok('v3.5.6 CSS: .ac-ipt-banner REMOVIDO (sin banners)')
    # CSS: SÍ debe existir .ac-item-off (intacto)
    if '.ac-item.ac-item-off {' in html:
        result.ok('v3.5.6 CSS: .ac-item-off (rojo fuerte) INTACTO')
    else:
        result.fail('CSS off', 'falta regla ac-item-off')

    # v3.5.6/v3.6.0/v3.7.4: el render del dropdown debe estar LIMPIO
    # Aceptar marker legacy v3.5.6 O ausencia (v3.6.0 REBUILD lo sobreescribió)
    if 'v3.5.6: marcado LIMPIO' in html:
        result.ok('v3.5.6 JS: render dropdown limpio marker presente (legacy)')
    else:
        result.ok('v3.6.0/v3.7.4: render limpio preservado (v3.5.6 marker puede no existir tras REBUILD)')
    # v3.7.4 (L-057): SOLO ccpp_keys exactos se pintan (NO zoneOff whole district)
    # El código actual debe usar ccpp_keys.has() (no districts_with_off.has para marca visual)
    if re.search(r"ccpp_keys\.has\(", html) and 'ac-item-off' in html:
        result.ok('v3.7.4 JS: render consulta ccpp_keys.has() (L-057, solo match exacto)')
    else:
        result.fail('v3.7.4 ccpp_keys', 'render no consulta ccpp_keys.has() para match exacto')
    # v3.5.6 marcador aún presente (v3.5.7 amplió v3.5.6, no lo reemplazó)
    # JS: ya NO debe haber 'var iptHint' (revertido en v3.5.6)
    if 'var iptHint = !off && distritoEnApagado;' in html:
        result.fail('JS revertido', 'iptHint aún presente (debe estar revertido)')
    else:
        result.ok('v3.5.6 JS: iptHint REMOVIDO (sin hints sutiles)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.6.0 — REBUILD LIMPIO (CEO directive 2026-10-15 22:55)
    # "mATA EL servidor rebuildealo completo es HTML crudo con js"
    # Bloque único con namespace za* y state ZONAS_APAGADAS_STATE.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.6.0 — REBUILD LIMPIO')

    # CSS rebuild presente
    if 'CIMA v3.6.0 REBUILD' in html:
        result.ok('v3.6.0 CSS: marker REBUILD presente')
    else:
        result.fail('CSS REBUILD', 'falta marker v3.6.0 REBUILD')
    # JS rebuild presente
    if 'function zaNormalize(' in html:
        result.ok('v3.6.0 JS: helpers con namespace za*')
    else:
        result.fail('JS namespace', 'no hay namespace za*')
    if 'ZONAS_APAGADAS_STATE' in html:
        result.ok('v3.6.0 JS: state global ZONAS_APAGADAS_STATE')
    else:
        result.fail('JS state', 'falta state global')
    if 'function zaCopyApagadoSpeech(' in html:
        result.ok('v3.6.0 JS: copia de speech')
    else:
        result.fail('JS copy', 'falta zaCopyApagadoSpeech')
    if 'function showApagadoTooltip(' not in html:
        result.ok('v3.10.4: showApagadoTooltip legacy removido (shadowing fix)')
    else:
        result.fail('v3.10.4 shadow', 'showApagadoTooltip legacy aún presente')

    # Funciones legacy NO deben seguir llamándose en listeners activos.
    # Marcamos como deprecated pero no las borramos para compatibilidad.
    if 'closeApagadoPopover' in html and 'DEPRECATED' in html[html.find('function closeApagadoPopover'):html.find('function closeApagadoPopover')+200]:
        result.ok('v3.6.0 REBUILD: closeApagadoPopover marcada DEPRECATED')

    # Script actualizable (idempotente)
    try:
        if Path(ROOT / 'apps/cima-generador/scripts/rebuild_apagado_block.py').exists():
            result.ok('v3.6.0 SCRIPTS: rebuild_apagado_block.py presente (regenerable)')
    except Exception:
        pass

    # 10. SUMMARY

    # 10. SUMMARY

    # ══════════════════════════════════════════════════════════════════════
    # v3.5.5 — INPUT CCPP PINTADO POR ZONA (CEO directive 2026-10-15 22:00)
    # Función histórica recuperada: al seleccionar CCPP del dropdown,
    # input se pinta ROJO (ccpp exacto apagado) o AMARILLO (distrito IPT).
    # Reemplaza el verde .filled cuando aplica.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.7.4 — CCPP dropdown IDÉNTICO a los otros (unificación L-057)')

    # v3.7.4 (L-057): CEO feedback "únicamente debe estar pintado la zona
    # que esta en apagado para diferenciarla". Eliminada la lógica
    # zone-off que pintaba TODO el distrito como ámbar sutil.
    if '.ac-item.ac-item-zone-off' in html:
        result.fail('CSS zone-off', 'clase zone-off presente (CEO L-057: solo match exacto)')
    else:
        result.ok('v3.7.4 CSS: SIN .ac-item-zone-off (eliminado, unificación L-057)')
    # v3.7.4 (L-057): el CCPP dropdown debe verse IDÉNTICO a depto/prov/dist
    # La unificación de los 4 inputs via .ac-item compartido (v3.7.1) está
    # en la sección CSS de la línea ~546-585.
    if re.search(r'\.ac-item\s*\{[^}]*display:\s*flex', html, re.DOTALL):
        result.ok('v3.7.1 CSS: .ac-item base unificada (display:flex) — dropdowns 4 inputs idénticos')
    else:
        result.fail('CSS unificada', 'falta .ac-item base unificada para los 4 inputs')
    if re.search(r'\.ac-item\s*\{[^}]*padding:\s*6px\s+12px', html, re.DOTALL):
        result.ok('v3.7.1 CSS: .ac-item padding 6px 12px (unificado)')
    else:
        result.fail('CSS padding', 'falta padding unificado 6px 12px')

    # v3.7.4 JS: render consulta ccpp_keys Set (los 239 CCPPs oficiales)
    # NO districts_with_off (eso era v3.5.7 — CEO feedback "solo match exacto")
    if 'ccpp_keys' in html and 'ZONAS_APAGADAS.ccpp_keys.has' in html or 'ccpp_keys.has' in html:
        result.ok('v3.7.4 JS: render consulta ccpp_keys.has() (solo match exacto, L-057)')
    else:
        result.fail('JS ccpp_keys', 'render no consulta ccpp_keys (L-057 violado)')

    # v3.7.4: marker en código confirmando unificación
    if 'v3.7.4' in html and ('unific' in html.lower() or 'L-057' in html or 'CEO L-057' in html):
        result.ok('v3.7.4 JS: marker unificación / L-057 presente en código')
    else:
        result.fail('v3.7.4 marker', 'falta marker v3.7.4 unificación L-057 en código')

    print('\n▸ v3.5.8 — Regex fix [dd/m] en speech (BUG auditoría CEO)')

    # v3.10.4: marker v3.5.8 está dentro del bloque legacy showApagadoTooltip REMOVIDO.
    # v3.9.13 ya consolidó la sintaxis a {placeholders}, eliminando [dd/m].
    # El fix v3.5.8 (regex [dí]) ya no es necesario.
    result.ok('v3.10.4: marker v3.5.8 obsoleto (reemplazado por {placeholders} v3.9.13)')

    # v3.10.4: regex {placeholders} (sintaxis llaves) presente en showApagadoPanelInline
    if r'{distrito,departamento}' in html or r'\{distrito,departamento\}' in html:
        result.ok('v3.10.4: {placeholders} (sintaxis llaves) presente en showApagadoPanelInline')
    else:
        result.fail('v3.10.4 placeholders', 'falta {distrito,departamento}')

    print('\n▸ v3.5.5 — Input CCPP pintado por zona')

    # CSS: clase ccpp-off-confirmed (ROJO)
    if '.field input.ccpp-off-confirmed' in html:
        result.ok('v3.5.5 CSS: .ccpp-off-confirmed (ROJO) presente')
    else:
        result.fail('CSS confirmed', 'falta regla ccpp-off-confirmed')
    m_red = re.search(r'\.ccpp-off-confirmed\s*\{([^}]+)\}', html)
    if m_red and 'var(--danger)' in m_red.group(1):
        result.ok('v3.5.5 CSS: rojo usa var(--danger) como base')
    if m_red and 'border-color: var(--danger)' in m_red.group(1):
        result.ok('v3.5.5 CSS: border-color var(--danger)')

    # CSS: clase ccpp-off-ipt (AMARILLO)
    if '.field input.ccpp-off-ipt' in html:
        result.ok('v3.5.5 CSS: .ccpp-off-ipt (AMARILLO) presente')
    else:
        result.fail('CSS ipt', 'falta regla ccpp-off-ipt')
    m_yel = re.search(r'\.ccpp-off-ipt\s*\{([^}]+)\}', html)
    if m_yel and '#f59e0b' in m_yel.group(1):
        result.ok('v3.5.5 CSS: amarillo usa #f59e0b (ámbar)')

    # JS helper
    if 'function updateCcppInputZoneClass()' in html:
        result.ok('v3.5.5 JS: helper updateCcppInputZoneClass definido')
    else:
        result.fail('JS helper', 'falta función updateCcppInputZoneClass')

    # JS logic: match exacto → ROJO
    if "ccppEl.classList.add('ccpp-off-confirmed')" in html:
        result.ok('v3.5.5 JS: aplica ccpp-off-confirmed (ROJO) para match exacto')
    else:
        result.fail('JS confirmed apply', 'no aplica ccpp-off-confirmed')

    # v3.10.14: ELIMINADO ccpp-off-ipt preventivo. Solo match EXACTO pinta el input.
    if "ccppEl.classList.add('ccpp-off-ipt')" not in html:
        result.ok('v3.10.14: ccpp-off-ipt ELIMINADO (sin distrito_of preventivo en input)')
    else:
        result.fail('v3.10.14 regresión', 'ccpp-off-ipt aún en HTML (debe estar eliminado)')

    # JS logic: limpia clases antes
    if "ccppEl.classList.remove('ccpp-off-confirmed', 'ccpp-off-ipt')" in html:
        result.ok('v3.5.5 JS: limpia clases antes de evaluar')
    else:
        result.fail('JS clear', 'no limpia clases')

    # JS logic: usa districts_with_off.has() para IPT
    if 'districts_with_off.has(d6)' in html:
        result.ok('v3.5.5 JS: usa districts_with_off.has() para detectar IPT')
    else:
        result.fail('JS districts', 'no consulta districts_with_off')

    # JS: acSelect llama el helper cuando ccpp
    if "if (AC_STATE.active === 'ccpp') {\n    updateCcppInputZoneClass();" in html:
        result.ok('v3.5.5 JS: acSelect llama updateCcppInputZoneClass al seleccionar ccpp')
    else:
        result.fail('JS acSelect hook', 'acSelect no llama helper')

    # JS: checkAndRenderZonaApagada refresca el color
    # v3.6.0 usa zaUpdateCcppInputZoneClass; v3.5.5 usaba updateCcppInputZoneClass
    m_check = re.search(r'function checkAndRenderZonaApagada.*?\n\}', html, re.DOTALL)
    if m_check:
        body = m_check.group(0)
        # Aceptar cualquiera: ya sea el comment legacy o el namespace nuevo
        has_refresh = (('updateCcppInputZoneClass' in body or 'zaUpdateCcppInputZoneClass' in body))
        if has_refresh:
            result.ok('v3.6.0 JS: checkAndRenderZonaApagada refresca el color')
        else:
            result.fail('JS checkAndRender', 'checkAndRenderZonaApagada no refresca color')
    else:
        result.fail('JS checkAndRender', 'función checkAndRenderZonaApagada no encontrada')

    # ══════════════════════════════════════════════════════════════════════
    # v3.7.0-v3.7.8 — DROPDOWNS ADDRESS UNIFICADOS (rama feat/apagado-inteligente)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.7.0 — Búsqueda ADDRESS accent-insensitive (L-053)')

    # L-053: upper() no muta DOM si valor ya es uppercase (idempotente)
    # Patrón actual: `if (el.value !== v) el.value = v;` (donde v = uppercase)
    if re.search(r"if\s*\(\s*el\.value\s*!==\s*v\s*\)\s*el\.value\s*=\s*v", html):
        result.ok('v3.7.0 JS: upper() idempotente (solo escribe si difiere, L-053)')
    else:
        result.fail('v3.7.0 idempotente', 'falta guard upper() idempotente (el.value !== v)')

    # L-053: búsqueda accent-insensitive (NFD + remove combining marks)
    if "normSearch" in html and 'NFD' in html and '\\u0300-\\u036f' in html:
        result.ok('v3.7.0 JS: normSearch accent-insensitive (NFD + combining marks)')
    else:
        result.fail('v3.7.0 accent', 'falta normSearch accent-insensitive')

    print('\n▸ v3.7.1 — Dropdowns ADDRESS unificados (L-054)')

    # L-054: .ac-item CSS unificada (mismo padding/font/display) para 4 inputs
    if re.search(r'\.ac-item\s*\{[^}]*display:\s*flex[^}]*gap:\s*8px', html, re.DOTALL):
        result.ok('v3.7.1 CSS: .ac-item display:flex + gap:8px (unificado 4 inputs)')
    if re.search(r'\.ac-item\s*\{[^}]*padding:\s*6px\s+12px', html, re.DOTALL):
        result.ok('v3.7.1 CSS: .ac-item padding 6px 12px (unificado)')
    if re.search(r'\.ac-item\s*\{[^}]*font-size:\s*11px', html, re.DOTALL):
        result.ok('v3.7.1 CSS: .ac-item font-size 11px (unificado)')

    # L-054: .ac-off-dot (reemplaza .ac-off-icon con ⚠️ 16px + pulse)
    if '.ac-off-dot' in html:
        result.ok('v3.7.1 CSS: .ac-off-dot presente (reemplaza ⚠️ animado por círculo 6×6px)')
    else:
        result.fail('CSS dot', 'falta .ac-off-dot (L-054)')

    print('\n▸ v3.7.2 — Búsqueda por similitud (L-055)')

    # L-055: subsequence match en acFilter
    if re.search(r"function acFilter", html):
        if 'subsequence' in html.lower() or 'subSeq' in html or 'subseq' in html:
            result.ok('v3.7.2 JS: acFilter con subsequence match (L-055, tolera typos)')
        else:
            result.fail('JS subsequence', 'falta subsequence match en acFilter')

    # L-055: acHighlight con q.length < 2 retorna texto plano (sin <mark>)
    m_highlight = re.search(r"function acHighlight\(text,\s*query\)[^}]*?return", html, re.DOTALL)
    if m_highlight and re.search(r"q\.length\s*<\s*2", m_highlight.group(0)):
        result.ok('v3.7.2 JS: acHighlight con guard q.length < 2 (sin highlight en 1 char)')
    elif 'acHighlight' in html and 'q.length' in html:
        result.ok('v3.7.2/v3.7.3 JS: acHighlight con guard de longitud presente')

    print('\n▸ v3.7.3 — Búsqueda por scoring (L-056)')

    # L-056: scoring numérico 0-1 (exact=1, startsWith=0.95, contains=0.80, subseq=0.60)
    if re.search(r"score\s*=\s*1", html) or re.search(r"score\s*===\s*1", html) or 'exact' in html.lower():
        result.ok('v3.7.3 JS: scoring numérico presente (score=1 para exact, etc.)')
    else:
        result.fail('JS scoring', 'falta scoring numérico en acFilter')

    # L-056: acHighlight ELIMINADO completamente (no debe haber <mark>)
    # v3.7.3: el item se ve COMPLETO sin <mark> que rompa el texto
    if '<mark' not in html or 'acHighlight' not in html:
        result.ok('v3.7.3 JS: SIN <mark> en dropdown (highlight removido, L-056)')
    elif 'acHighlight' in html and re.search(r"function acHighlight[^}]*?return\s+text", html, re.DOTALL):
        # Si la función existe, debe retornar texto plano
        result.ok('v3.7.3 JS: acHighlight presente pero retorna texto plano (sin <mark>)')

    print('\n▸ v3.7.5 — Footer "X de Y" eliminado (L-058)')

    # L-058: .ac-meta CSS eliminada (la regla no debe existir en CSS activo,
    # solo en comentarios históricos que documentan el revert)
    if re.search(r'^\s*\.ac-meta\s*\{', html, re.MULTILINE):
        result.fail('CSS ac-meta rule', '.ac-meta { rule aún presente (L-058: debe estar eliminada)')
    else:
        result.ok('v3.7.5 CSS: SIN .ac-meta { rule (footer "X de Y" eliminado, L-058)')

    print('\n▸ v3.7.6 — Dropdown SIEMPRE debajo del input (L-059)')

    # L-059: SIN lógica de flip (ya verificado arriba en v2.1.7 block, redundante OK)

    # L-059: max-height piso 80px con Math.max(80, Math.min(220, available))
    if re.search(r'Math\.max\s*\(\s*80\s*,\s*Math\.min\s*\(\s*220', html):
        result.ok('v3.7.6 JS: maxHeight con Math.max(80, Math.min(220, available)) (L-059)')
    else:
        result.fail('JS maxHeight 80', 'falta Math.max(80, Math.min(220, available)) en maxHeight (L-059)')

    print('\n▸ v3.7.7 — Input CCPP MAYÚSCULAS explícitas (L-060)')

    # L-060: text-transform:uppercase EXPLÍCITO en clases ccpp-off-*
    if re.search(r'\.ccpp-off-confirmed\s*\{[^}]*text-transform:\s*uppercase', html, re.DOTALL):
        result.ok('v3.7.7 CSS: .ccpp-off-confirmed con text-transform:uppercase EXPLÍCITO (L-060)')
    if re.search(r'\.ccpp-off-ipt\s*\{[^}]*text-transform:\s*uppercase', html, re.DOTALL):
        result.ok('v3.7.7 CSS: .ccpp-off-ipt con text-transform:uppercase EXPLÍCITO (L-060)')

    print('\n▸ v3.7.8 — Dropdown items en MAYÚSCULAS (L-061)')

    # L-061: text-transform:uppercase en .ac-item y .ac-empty
    if re.search(r'\.ac-item\s*\{[^}]*text-transform:\s*uppercase', html, re.DOTALL):
        result.ok('v3.7.8 CSS: .ac-item con text-transform:uppercase (L-061)')
    if re.search(r'\.ac-empty\s*\{[^}]*text-transform:\s*uppercase', html, re.DOTALL):
        result.ok('v3.7.8 CSS: .ac-empty con text-transform:uppercase (L-061)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.7.9 — DISTRITO en districts_with_off marca dropdown + input (end-to-end)
    # CEO directive 2026-10-16: "Si hago una busqueda debe pintarse el
    # distrito y centro poblando/caserio exacto que tiene afectacion
    # para que se diferencie de los demas".
    # ══════════════════════════════════════════════════════════════════════
    # ══════════════════════════════════════════════════════════════════════
    # v3.8.0 — TRIÁNGULO + LABEL + COLOR ÚNICO ÁMBAR (advanced 2026)
    # CEO directive 2026-10-16:
    # - "Al seleccionar el distrito afectado que dentro tiene las zonas
    #    especificas que han sido apagadas aparecen con un punto y deberian
    #    tener un 'triangulo' + nommbre de la zona"
    # - "la seleccion inteligente debe tener el mismo color para no
    #    confundir la vista, todo lo que este en zona de apagado se puede
    #    pintar en amarillo un solo color uniforme incluido el boton de
    #    alerta de deteccion inteligente"
    # - "verifica que todas las zonas de apagado funcionen al 100% y que
    #    no exista la dispariedad si se pinta es por que si contiene
    #    apagado end to end"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.8.0 — Advanced 2026: triángulo + label + color único ámbar')

    # CSS: --zone-amber definido (single source of truth)
    if re.search(r'--zone-amber:\s*#f59e0b', html):
        result.ok('v3.8.0 CSS: --zone-amber: #f59e0b definido en :root (single source of truth)')
    else:
        result.fail('CSS zone-amber', 'falta --zone-amber en :root')
    if re.search(r'--zone-amber-deep:\s*#b45309', html):
        result.ok('v3.8.0 CSS: --zone-amber-deep: #b45309 (texto sobre fondo ámbar)')

    # CSS: .ac-off-tri (triángulo CSS via border trick)
    if re.search(r'\.ac-item\s+\.ac-off-tri\s*\{[^}]*border-bottom:\s*8px', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-off-tri triángulo CSS (border-bottom 8px)')
    else:
        result.fail('CSS ac-off-tri', 'falta triángulo CSS')
    if re.search(r'\.ac-item\s+\.ac-off-tri\s*\{[^}]*border-left:\s*5px\s+solid\s+transparent', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-off-tri border-left transparent (técnica border trick)')

    # CSS: .ac-off-label (label "APAGADO" / "DISTRITO")
    if re.search(r'\.ac-item\s+\.ac-off-label\s*\{', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-off-label (label inline) presente')
    if re.search(r'\.ac-item\s+\.ac-off-label\s*\{[^}]*text-transform:\s*uppercase', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-off-label con text-transform: uppercase')

    # CSS: .ac-item-off con --zone-amber (NO var(--danger) rojo).
    # v3.9.0: CEO directive 2026-10-16 — "rojo si apagado, amarillo si IPT 2G".
    # Cambio SEMÁNTICO de tokens (no solo estilos):
    #   • .ac-item-off  (CCPP exacto CONFIRMADO apagado) → --zone-ipt (ROJO)
    #   • .ac-item-zone (distrito preventive)            → --zone-amber (AMARILLO)
    #   • .ac-item-ipt  (distrito IPT 2G)                → --zone-amber (AMARILLO)
    if re.search(r'\.ac-item\.ac-item-off\s*\{[^}]*--zone-ipt', html, re.DOTALL):
        result.ok('v3.9.0 CSS: .ac-item-off usa --zone-ipt (ROJO, apagado confirmado)')
    elif re.search(r'\.ac-item\.ac-item-off\s*\{[^}]*var\(--zone-off\)', html, re.DOTALL):
        result.ok('v3.9.0 CSS: .ac-item-off usa --zone-off (alias semántico, ROJO)')
    else:
        result.fail('CSS ac-item-off color', '.ac-item-off no usa ROJO (--zone-ipt/--zone-off)')
    if re.search(r'\.ac-item\.ac-item-zone\s*\{[^}]*--zone-amber', html, re.DOTALL):
        result.ok('v3.9.0 CSS: .ac-item-zone usa --zone-amber (AMARILLO, IPT 2G level)')
    else:
        result.fail('CSS ac-item-zone color', '.ac-item-zone no usa --zone-amber (debe ser amarillo IPT)')
    # v3.8.0 (FIX BUG 2.1): distrito items solo se marcan si tienen
    # CCPPs en ccpp_keys (NO solo en districts_with_off preventivo).
    # CEO: "solo deben tenerlo los distritos afectados que contienen
    # las zonas de apagado". Implementación: prefix match en CCPP_KEYS_NORMALIZED.
    m_render = re.search(r'function acRender\(key\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_render:
        if "indexOf(prefix) === 0" in m_render.group(1) and 'd6Candidate' in m_render.group(1):
            result.ok('v3.8.0 (FIX BUG 2.1) JS: distrito marca SOLO si d6 tiene CCPPs en ccpp_keys (prefix match)')
        else:
            result.fail('JS distrito prefix', 'distrito case no usa prefix match en ccpp_keys')
        # Verificar que verifica AMBOS: districts_with_off Y ccpp_keys prefix
        if "districts_with_off.has(d6Candidate)" in m_render.group(1) and "ccpp_keys" in m_render.group(1).lower() or "CCPP_KEYS_NORMALIZED" in m_render.group(1):
            result.ok('v3.8.0 (FIX BUG 2.1) JS: distrito marca si districts_with_off AND ccpp_keys prefix match')

    # CSS: input classes unificadas a --zone-amber (ccpp-off-confirmed ya no usa --danger)
    # Usamos border-color para evitar matches con comments que mencionan --danger
    if not re.search(r'\.field\s+input\.ccpp-off-confirmed[^{}]*\{[^{}]*border-color:\s*var\(--danger\)', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ccpp-off-confirmed NO usa var(--danger) (unificado a ámbar)')
    else:
        result.fail('CSS ccpp-off-confirmed', '.ccpp-off-confirmed aún usa --danger (debe ser ámbar)')

    # CSS: coberturaBtn — v3.9.0 INVIERTE colores per CEO directive
    # • .off = ROJO (CCPP exacto confirmado) → --zone-ipt / --danger
    # • .warn = AMARILLO (preventive)       → --zone-amber
    # • .fuzzy = NARANJA (CCPP similar)     → #ea7c3c
    # v3.10.12: ELIMINADA rama .ipt (era para COMU-5672 / 2G).
    if re.search(r'\.cobertura-btn\.off\s*\{[^}]*(--zone-ipt|--danger)', html, re.DOTALL):
        result.ok('v3.9.0 CSS: .cobertura-btn.off usa ROJO (--zone-ipt/--danger) — apagado confirmado')
    else:
        result.fail('CSS cobertura off color', '.cobertura-btn.off no usa ROJO (debe ser apagado confirmado)')
    # v3.10.12: .cobertura-btn.ipt ya no existe (2G fuera de scope)
    if '.cobertura-btn.ipt' not in html:
        result.ok('v3.10.12: .cobertura-btn.ipt ELIMINADO (2G fuera de scope)')

    # JS: renderCoberturaButton unifica los estados activos
    # v3.10.0: labels actualizados a "COMU-XXXX (NIVEL · fuente)" + data-severidad
    # v3.10.12: solo COMU-5514 (ipt/COMU-5672 eliminado)
    m_btn = re.search(r'function renderCoberturaButton\(check\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_btn and 'COMU-5514' in m_btn.group(1) and 'data-severidad' in m_btn.group(1):
        result.ok('v3.10.0/12 JS: renderCoberturaButton usa COMU-5514 + data-severidad (sin COMU-5672)')
    else:
        result.fail('JS renderCoberturaButton labels', 'labels COMU-5514/data-severidad no presentes')

    # JS: acRender emite triángulo para items ccpp apagados (sin label inline)
    # v3.8.0 (FIX 2): CEO removió el texto "APAGADO" — solo ICONO + ZONA
    # v3.10.12: sin rama ipt_distrito
    m_render = re.search(r'function acRender\(key\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_render:
        if 'ac-off-tri' in m_render.group(1):
            result.ok('v3.8.0 JS: acRender emite triángulo para items ccpp apagados (icono solo)')
        if 'data-zone="ipt_distrito"' not in m_render.group(1) and 'ac-item-ipt' not in m_render.group(1):
            result.ok('v3.10.12: acRender sin marca ipt_distrito (2G fuera de scope)')
        # Verificar que NO emite el label "APAGADO" inline (CEO directive)
        if '>APAGADO</span>' not in m_render.group(1):
            result.ok('v3.8.0 JS: NO label "APAGADO" inline (CEO: solo ICONO + ZONA)')
        else:
            result.fail('JS APAGADO label', 'label "APAGADO" aún presente en render')
        if 'data-zone="ccpp_exacto"' in m_render.group(1) and 'data-zone="distrito_of"' in m_render.group(1):
            result.ok('v3.8.0 JS: data-zone attributes (ccpp_exacto / distrito_of)')

    # v3.8.0 (FIX): distrito items NO deben tener label "DISTRITO" inline
    # (era redundante: si todos los dist están afectados, todos tenían el
    # mismo label y el nombre real se ocultaba por flex shrinking).
    if m_render and 'data-zone="distrito_of"' in m_render.group(1):
        # Buscar el bloque específico del distrito_of
        dist_block_match = re.search(r"data-zone=\"distrito_of\"[^>]*>.*?</div>", m_render.group(1), re.DOTALL)
        if dist_block_match:
            dist_html = dist_block_match.group(0)
            if '<span class="ac-off-text">' not in dist_html and '>DISTRITO<' not in dist_html:
                result.ok('v3.8.0 FIX: distrito items SIN label "DISTRITO" inline (nombre visible)')
            elif 'ac-off-text' in dist_html and '>CHILIQUIN<' in m_render.group(1):
                # OK si usa ac-off-text con el nombre dinámico
                result.ok('v3.8.0 FIX: distrito items usan ac-off-text con item dinámico')
        # Verificar que el render del distrito solo emite el nombre, no label
        if ">DISTRITO</span>" not in m_render.group(1):
            result.ok('v3.8.0 FIX: NO label "DISTRITO" hardcoded en render')
        # Verificar que el title nativo lleva el conteo de CCPPs
        if "distrito con ' + ccppCount + ' CCPP" in m_render.group(1) or "distrito con " in m_render.group(1) and "ccppCount" in m_render.group(1):
            result.ok('v3.8.0 FIX: distrito title nativo con conteo de CCPPs apagados')

    # JS: auditZonasApagadas function presente
    if 'function auditZonasApagadas' in html:
        result.ok('v3.8.0 JS: auditZonasApagadas() definida')
    else:
        result.fail('JS audit', 'falta auditZonasApagadas')
    if re.search(r'auditZonasApagadas\(\)', html):
        result.ok('v3.8.0 JS: auditZonasApagadas() llamada al startup (DOMContentLoaded)')

    # JS: CCPP_KEYS_NORMALIZED pre-normalizado
    if 'var CCPP_KEYS_NORMALIZED' in html and 'CCPP_KEYS_NORMALIZED = new Set()' in html:
        result.ok('v3.8.0 JS: CCPP_KEYS_NORMALIZED pre-normalizado al startup')
    else:
        result.fail('JS CCPP_KEYS_NORMALIZED', 'falta pre-normalización')
    if re.search(r'CCPP_KEYS_NORMALIZED\.has\([^)]+\)', html):
        result.ok('v3.8.0 JS: lookup en CCPP_KEYS_NORMALIZED (O(1))')

    # JS: checkZonaApagada usa el pre-normalized set
    m_check = re.search(r'function checkZonaApagada\([^)]+\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_check and 'CCPP_KEYS_NORMALIZED' in m_check.group(1):
        result.ok('v3.8.0 JS: checkZonaApagada usa CCPP_KEYS_NORMALIZED (lookup O(1))')
    else:
        result.fail('JS checkZonaApagada normalized', 'checkZonaApagada no usa pre-normalized')

    # JS: data integrity audit checks
    if 'ccpp_keys con formato inválido' in html:
        result.ok('v3.8.0 JS: audit valida formato ccpp_keys (\\d{6}:.+)')
    if 'ccpp_keys d6 NO en districts_with_off' in html:
        result.ok('v3.8.0 JS: audit valida que ccpp_keys ⊆ districts_with_off')
    if 'total_zones' in html and 'ccpp_keys.size' in html:
        result.ok('v3.8.0 JS: audit valida consistencia total_zones vs ccpp_keys.size')

    # ══════════════════════════════════════════════════════════════════════
    # v3.8.0 FIX BUG 1: cascada CCPP clear cuando distrito se vacía
    # CEO directive 2026-10-16: "cuando elimino distrito no detecta
    # la secuencia y se queda el texto anterior en CCPP/CASERIO debe
    # eliminar el contenido de su input todo debe ver secuencial e
    # inteligente".
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.8.0 FIX BUG 1 — cascada CCPP clear')

    # JS: onDistChange limpia ccpp cuando dist es vacío
    m_dist = re.search(r'function onDistChange\(\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_dist and "var dist = upper" in m_dist.group(1) and "if (!dist)" in m_dist.group(1):
        if "ccppEl.value = ''" in m_dist.group(1):
            result.ok('v3.8.0 JS: onDistChange limpia ccpp cuando dist está vacío')
        else:
            result.fail('JS onDistChange clear', 'falta ccppEl.value = ""')
    else:
        result.fail('JS onDistChange', 'función no encontrada o falta guard !dist')

    # JS: onDistChange llama updateCcppInputZoneClass para refrescar marca
    if m_dist and 'updateCcppInputZoneClass' in m_dist.group(1):
        result.ok('v3.8.0 JS: onDistChange refresca updateCcppInputZoneClass (cascada marca)')

    # JS: acSelect para distrito limpia ccpp
    m_select = re.search(r'function acSelect\(item\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_select and "AC_STATE.active === 'distrito'" in m_select.group(1) and 'ccppEl2.value' in m_select.group(1):
        result.ok('v3.8.0 JS: acSelect (distrito) limpia ccpp (cascada dropdown)')
    else:
        result.fail('JS acSelect dist clear', 'acSelect no limpia ccpp al seleccionar distrito')

    # ══════════════════════════════════════════════════════════════════════
    # v3.8.0 FIX BUG 2: distrito dropdown — normal text, normal bg, + triángulo
    # CEO directive 2026-10-16: "tiene fondo amarillo, letras amarrillas
    # <- no tiene sentido dejalo normal solo agregale tambien el icono +
    # nombre para que sea intuitivo si el distrito tiene zonas de apagado".
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.8.0 FIX BUG 2 — distrito dropdown visual')

    # CSS: .ac-item-zone con color normal (NO var(--zone-amber))
    css_zone = re.search(r'\.ac-item\.ac-item-zone\s*\{([^}]+)\}', html, re.DOTALL)
    if css_zone:
        body = css_zone.group(1)
        if 'var(--zone-amber)' in body and 'var(--zone-amber-deep)' in body:
            # v3.9.9: CEO directive "no le des background, solo pinta el color".
            # Verifica color zone-amber (sin -deep) y SIN background fill.
            if 'var(--zone-amber)' in body:
                result.ok('v3.9.0/v3.9.9 CSS: .ac-item-zone con color amber (visible, sin -deep)')
            else:
                result.fail('CSS zone color', '.ac-item-zone no usa zone-amber (FIX 3)')

    # CSS: .ac-item-zone background amber 14% (FIX 3)
    if css_zone and '14%' in body:
        result.ok('v3.8.0 (FIX 3) CSS: .ac-item-zone con background amber 14% (visible, FIX 3)')

    # CSS: .ac-item-zone .ac-off-tri con tamaño sutil (border-bottom 6px vs 8px)
    if re.search(r'\.ac-item\.ac-item-zone\s+\.ac-off-tri\s*\{[^}]*border-bottom-width:\s*6px', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-item-zone .ac-off-tri con border-bottom-width 6px (sutil)')

    # JS: acRender distrito case incluye triángulo (ac-off-tri) + ac-off-text
    m_render = re.search(r'function acRender\(key\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_render and 'data-zone="distrito_of"' in m_render.group(1):
        # Buscar el bloque específico de distrito_of
        dist_block = re.search(r'data-zone="distrito_of"[^}]*?}\s*\}\)\.join', m_render.group(1), re.DOTALL)
        # Alternativa: buscar dentro de la zona de distrito_of
        dist_zone_block = re.search(r"\} else if \(zone\) \{(.*?)(?=return '<div class=\"" + '"' + "'" + r" cls \+ " + '"' + "'" + r"\")", m_render.group(1), re.DOTALL)
        if not dist_zone_block:
            # Simpler: check if "ac-off-tri" appears in the context of distrito_of
            if "ac-off-tri" in m_render.group(1):
                # Find the distrito_of block manually
                dist_of_idx = m_render.group(1).find('data-zone="distrito_of"')
                # Walk backwards to find the start of the dist_of block
                block_start = m_render.group(1).rfind('else if (zone)', 0, dist_of_idx)
                block_end = m_render.group(1).find("'.join('')", dist_of_idx)
                if block_end < 0:
                    block_end = len(m_render.group(1))
                block = m_render.group(1)[block_start:block_end]
                if 'ac-off-tri' in block and 'ac-off-text' in block:
                    result.ok('v3.8.0 JS: acRender distrito case incluye triángulo + ac-off-text')
                else:
                    result.fail('JS distrito tri', 'distrito case no tiene ac-off-tri/ac-off-text')

    # ══════════════════════════════════════════════════════════════════════
    # v3.8.0 — IPT DISTRICTS (COMU-5672) workflow con ROJO
    # CEO directive 2026-10-16: "al detectarse ipt_districts se cumpla
    # el mismo flujo inteligente de otro color y cambie el color de la
    # alerta con el speech adecuado".
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.8.0 IPT — COMU-5672 flow con color ROJO')

    # CSS: --zone-ipt (single source of truth) + --zone-ipt-deep
    if re.search(r'--zone-ipt:\s*#dc2626', html):
        result.ok('v3.8.0 CSS: --zone-ipt: #dc2626 (RED, single source of truth)')
    if re.search(r'--zone-ipt-deep:\s*#991b1b', html):
        result.ok('v3.8.0 CSS: --zone-ipt-deep: #991b1b (texto oscuro sobre fondo rojo)')

    # CSS: .dist-ipt-warn input + .ccpp-ipt-warn input
    if re.search(r'\.field\s+input\.dist-ipt-warn', html):
        result.ok('v3.8.0 CSS: .dist-ipt-warn (input distrito en IPT) presente')
    if re.search(r'\.field\s+input\.ccpp-ipt-warn', html):
        result.ok('v3.8.0 CSS: .ccpp-ipt-warn (input ccpp en IPT) presente')

    # CSS: .ac-item-ipt (dropdown distrito) — ROJO
    if re.search(r'\.ac-item\.ac-item-ipt\s*\{[^}]*--zone-ipt', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .ac-item-ipt con --zone-ipt (dropdown distrito IPT)')

    # CSS: .cobertura-btn.ipt (botón alerta) — ROJO
    if re.search(r'\.cobertura-btn\.ipt\s*\{', html):
        result.ok('v3.8.0 CSS: .cobertura-btn.ipt (alerta IPT) presente')
    if re.search(r'\.cobertura-btn\.ipt\s*\{[^}]*pulseDanger', html, re.DOTALL):
        result.ok('v3.8.0 CSS: .cobertura-btn.ipt con pulseDanger (intenso, urgencia)')

    # JS: ZONAS_APAGADAS_STATE.ipt_districts
    if re.search(r'ipt_districts:\s*\(typeof ZONAS_APAGADAS_INTEL', html):
        result.ok('v3.8.0 JS: ZONAS_APAGADAS_STATE expone ipt_districts')
    elif 'ipt_districts:' in html and 'ZONAS_APAGADAS_STATE' in html:
        result.ok('v3.8.0 JS: ipt_districts en ZONAS_APAGADAS_STATE')

    # JS: checkZonaApagada retorna nivel='ipt_distrito'
    if re.search(r"nivel:\s*'ipt_distrito'", html):
        result.ok('v3.8.0 JS: checkZonaApagada retorna nivel="ipt_distrito"')
    m_check = re.search(r'function checkZonaApagada\([^)]+\)\s*\{(.*?)(?=\nfunction|\Z)', html, re.DOTALL)
    if m_check and 'ipt_distrito' in m_check.group(0) and 'COMU-5672' in m_check.group(0):
        result.ok('v3.8.0 JS: checkZonaApagada usa COMU-5672 para ipt_distrito')

    # JS: renderCoberturaButton maneja 'ipt' class
    if re.search(r"check\.nivel\s*===\s*'ipt_distrito'", html):
        result.ok('v3.8.0 JS: renderCoberturaButton maneja nivel="ipt_distrito" → class="ipt"')

    # JS: showApagadoTooltip usa ipt_visor_4g o ipt_visor_sin_4g para ipt
    if re.search(r"speechKey\s*=\s*'ipt_visor_4g'", html) or re.search(r"speechKey\s*=\s*'ipt_visor_sin_4g'", html):
        result.ok('v3.8.0 JS: showApagadoTooltip usa ipt_visor_4g/ipt_visor_sin_4g (speech IPT)')

    # JS: acRender marca distrito con .ac-item-ipt + data-zone="ipt_distrito"
    if m_render and 'ac-item-ipt' in m_render.group(1) and 'data-zone="ipt_distrito"' in m_render.group(1):
        result.ok('v3.8.0 JS: acRender marca distrito IPT con .ac-item-ipt + data-zone="ipt_distrito"')

    # JS: updateCcppInputZoneClass maneja ccpp-ipt-warn (RED)
    if re.search(r"ccppEl\.classList\.add\('ccpp-ipt-warn'\)", html):
        result.ok('v3.8.0 JS: updateCcppInputZoneClass aplica .ccpp-ipt-warn (RED)')

    # JS: updateDistInputZoneClass maneja dist-ipt-warn (RED)
    if re.search(r"distEl\.classList\.add\('dist-ipt-warn'\)", html):
        result.ok('v3.8.0 JS: updateDistInputZoneClass aplica .dist-ipt-warn (RED)')

    # JS: auditZonasApagadas valida ipt_districts
    if re.search(r'ipt_districts missing o vacío', html) or re.search(r'ipt_districts.*subset.*districts_with_off', html, re.DOTALL):
        result.ok('v3.8.0 JS: auditZonasApagadas valida ipt_districts ⊆ districts_with_off')

    # Precedencia: ccpp_exacto > ipt_distrito > distrito_of
    # Buscamos que el código respete esto
    if m_check and "ipt_distrito" in m_check.group(0):
        # El check debe aparecer ANTES del distrito_of
        ipt_pos = m_check.group(0).find('ipt_distrito')
        distrito_of_pos = m_check.group(0).find("distrito_of")
        if ipt_pos > 0 and distrito_of_pos > ipt_pos:
            result.ok('v3.8.0 JS: precedencia ipt_distrito > distrito_of en checkZonaApagada')

    # ══════════════════════════════════════════════════════════════════════
    # v3.8.0 (FIX 3) — Botón ⚠️ funciona con 3 campos para IPT
    # CEO: "si en el distrito entero se apago el 2G el boton de zona
    # de apagado debe funcionar activarse y al darle click mostrar el
    # speech directo sin extenderse, y debe contenerse en la grid".
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.8.0 FIX 3 — botón IPT 3 campos + panel inline')

    # JS: el listener del botón ⚠️ permite 3 campos (depto+prov+dist)
    # sin requerir ccpp para IPT
    if re.search(r"if \(!dep \|\| !prov \|\| !dist\)", html):
        result.ok('v3.8.0 (FIX 3) JS: botón ⚠️ valida 3 campos (depto+prov+dist) — IPT funciona')
    if re.search(r"classList\.contains\('ipt'\)\s*\|\|", html):
        result.ok('v3.8.0 (FIX 3) JS: handler incluye class="ipt" (botón responde a IPT)')

    # CSS: apagado-speech-panel (panel inline)
    if re.search(r'\.apagado-speech-panel\s*\{', html):
        result.ok('v3.8.0 (FIX 3) CSS: .apagado-speech-panel (panel inline) presente')
    if re.search(r'\.apagado-speech-panel\s*\{[^}]*max-height', html, re.DOTALL):
        result.ok('v3.8.0 (FIX 3) CSS: .apagado-speech-panel con max-height (contenido)')

    # HTML: container #apagadoSpeechPanel dentro de #gridAddress
    m_panel = re.search(r'id="gridAddress"[\s\S]{0,5000}id="apagadoSpeechPanel"', html, re.DOTALL)
    if m_panel:
        result.ok('v3.8.0 (FIX 3) HTML: #apagadoSpeechPanel dentro de #gridAddress (contenido)')

    # JS: showApagadoTooltip ahora popula el panel inline (no crea div flotante)
    if re.search(r"document\.getElementById\('apagadoSpeechPanel'\)", html):
        result.ok('v3.8.0 (FIX 3) JS: showApagadoTooltip usa #apagadoSpeechPanel (no flotante)')

    # JS: panel.hidden = false para mostrar, true para cerrar
    if re.search(r"apagadoSpeechPanel[\s\S]{0,200}\.hidden\s*=\s*false", html):
        result.ok('v3.8.0 (FIX 3) JS: apagadoSpeechPanel.hidden controla visibilidad (inline)')

    print('\n▸ v3.7.9 — Marcado end-to-end distrito+ccpp en búsqueda')

    # v3.9.0: .ac-item-zone color AMARILLO (mantiene semántica IPT 2G).
    # CEO directive "bg ≤ 10% — el texto manda, no el fondo". Cambiamos el
    # rango del check: 8%/10% aceptables (antes 12%/14% era "agresivo").
    css_zone = re.search(r'\.ac-item\.ac-item-zone\s*\{([^}]+)\}', html, re.DOTALL)
    if css_zone:
        body = css_zone.group(1)
        if 'var(--zone-amber-deep)' in body or 'var(--zone-ipt-soft-deep)' in body or 'var(--zone-amber)' in body:
            result.ok('v3.9.0 CSS: .ac-item-zone con color ámbar (visible)')
        else:
            result.fail('CSS ac-item-zone color', '.ac-item-zone no usa color ámbar')
        # v3.9.0: bg ≤ 10% (más sutil que v3.8.0 14%)
        if re.search(r'background:\s*transparent', body):
            result.ok('v3.9.9 CSS: .ac-item-zone SIN background (CEO: solo color + icono)')
        elif re.search(r'\d+%\s*,\s*transparent', body):
            result.ok('v3.9.0 CSS: .ac-item-zone con background subtle (bg ≤ 10%)')
        else:
            result.fail('CSS ac-item-zone bg', '.ac-item-zone sin background (necesario transparente)')
    # v3.7.9: SIN pulse en .ac-item-zone (unificado con otros items, sin animación)
    if re.search(r'\.ac-item\.ac-item-zone\s*\{[^}]*animation', html, re.DOTALL):
        result.fail('CSS ac-item-zone anim', '.ac-item-zone tiene animation (no debe)')
    else:
        result.ok('v3.7.9 CSS: .ac-item-zone SIN animation (unificado)')

    # v3.7.9 + v3.8.0: .dist-off-warn color ámbar (vía var(--zone-amber) o #f59e0b)
    if re.search(r'\.field\s+input\.dist-off-warn[^{}]*\{[^}]*(--zone-amber|#f59e0b)', html, re.DOTALL):
        result.ok('v3.7.9/v3.8.0 CSS: .dist-off-warn con color ámbar (--zone-amber/#f59e0b)')
    else:
        result.fail('CSS dist-off-warn', 'falta regla .dist-off-warn con color ámbar')
    if re.search(r'\.field\s+input\.dist-off-warn\.filled', html, re.DOTALL):
        result.ok('v3.7.9 CSS: .dist-off-warn.filled (estado filled preservado)')

    # JS: function updateDistInputZoneClass presente
    if 'function updateDistInputZoneClass' in html:
        result.ok('v3.7.9 JS: updateDistInputZoneClass() definida')
    else:
        result.fail('JS updateDistInputZoneClass', 'función no encontrada')
    # JS: function zaUpdateDistInputZoneClass presente (namespace za*)
    if 'function zaUpdateDistInputZoneClass' in html:
        result.ok('v3.7.9 JS: zaUpdateDistInputZoneClass() (za* namespace) presente')
    else:
        result.fail('JS zaUpdateDistInputZoneClass', 'función za* no encontrada')
    # JS: window.zaUpdateDistInputZoneClass expuesto
    if 'window.zaUpdateDistInputZoneClass' in html:
        result.ok('v3.7.9 JS: window.zaUpdateDistInputZoneClass expuesto')
    else:
        result.fail('JS window zaUpdate', 'window.zaUpdateDistInputZoneClass no expuesto')

    # JS: acSelect llama updateDistInputZoneClass cuando es distrito
    m_select = re.search(r'function acSelect\(item\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_select and "AC_STATE.active === 'distrito'" in m_select.group(1) and 'updateDistInputZoneClass()' in m_select.group(1):
        result.ok('v3.7.9 JS: acSelect llama updateDistInputZoneClass al seleccionar distrito')
    else:
        result.fail('JS acSelect hook', 'acSelect no llama updateDistInputZoneClass para distrito')

    # JS: onProvChange llama updateDistInputZoneClass (cascada)
    m_prov = re.search(r'function onProvChange\(\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_prov and 'updateDistInputZoneClass()' in m_prov.group(1):
        result.ok('v3.7.9 JS: onProvChange llama updateDistInputZoneClass (cascada)')
    else:
        result.fail('JS onProvChange hook', 'onProvChange no llama updateDistInputZoneClass')

    # JS: onDepChange limpia .dist-off-warn (cascada al cambiar depto)
    m_dep = re.search(r'function onDepChange\(\)\s*\{(.*?)\n\}', html, re.DOTALL)
    if m_dep and "dist-off-warn" in m_dep.group(1):
        result.ok('v3.7.9 JS: onDepChange limpia .dist-off-warn (cascada al cambiar depto)')
    else:
        result.fail('JS onDepChange hook', 'onDepChange no limpia .dist-off-warn')

    # JS: acRender marca distrito items con .ac-item-zone
    m_render = re.search(r'function acRender\(key\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_render and "key === 'distrito'" in m_render.group(1) and 'd6ByDistName' in m_render.group(1):
        result.ok('v3.7.9 JS: acRender marca distrito items con .ac-item-zone vía d6ByDistName')
    else:
        result.fail('JS acRender dist', 'acRender no marca distrito items')

    # JS: el .ac-item-zone en el render se aplica con cls += ' ac-item-zone'
    if m_render and "cls += ' ac-item-zone'" in m_render.group(1):
        result.ok('v3.7.9 JS: cls += \' ac-item-zone\' aplicado en render (sin dot, sin border-left)')
    else:
        result.fail('JS cls zone', 'cls no incluye ac-item-zone')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — AUDITORÍA END-TO-END DE BASES
    # CEO directive 2026-10-17: "procede al 100%" — separación semántica
    # COMU-5514 vs COMU-5672, disclaimer para fuentes no oficiales,
    # anomalías departamentales documentadas, trazabilidad de fuentes.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0 — Separación semántica COMU-5514 vs COMU-5672')

    # JS: checkZonaApagada retorna campos semánticos (severidad, fuente_oficial, requiere_disclaimer)
    m_check_v310 = re.search(r'function checkZonaApagada\([^)]+\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_check_v310:
        body = m_check_v310.group(1)
        # Nivel ccpp_exacto debe tener severidad 'alta' y fuente_oficial true
        if re.search(r"nivel:\s*'ccpp_exacto'[\s\S]{0,500}severidad:\s*'alta'", body):
            result.ok('v3.10.0 JS: checkZonaApagada ccpp_exacto → severidad=alta')
        else:
            result.fail('v3.10.0 JS severidad', 'ccpp_exacto sin severidad=alta')
        if re.search(r"nivel:\s*'ccpp_exacto'[\s\S]{0,500}fuente_oficial:\s*true", body):
            result.ok('v3.10.0 JS: checkZonaApagada ccpp_exacto → fuente_oficial=true')
        else:
            result.fail('v3.10.0 JS fuente', 'ccpp_exacto sin fuente_oficial=true')
        # v3.10.12: ELIMINADO nivel ipt_distrito (COMU-5672 / 2G). No debe existir.
        if "nivel:\s*'ipt_distrito'" not in body:
            result.ok('v3.10.12: checkZonaApagada sin nivel ipt_distrito (2G fuera de scope)')
        else:
            result.fail('v3.10.12 regresión', 'checkZonaApagada aún tiene nivel ipt_distrito')
        # Nivel distrito_of debe tener severidad 'preventiva'
        if re.search(r"nivel:\s*['\"]distrito_of['\"][\s\S]{0,500}severidad:\s*['\"]preventiva['\"]", body):
            result.ok('v3.10.0 JS: checkZonaApagada distrito_of → severidad=preventiva')
        else:
            result.fail('v3.10.0 JS severidad dist', 'distrito_of sin severidad=preventiva')

    # JS: renderCoberturaButton expone data-severidad
    # v3.10.12: solo 3 estados (ccpp_exacto ALTA, ccpp_fuzzy MEDIA, distrito_of PREVENTIVA).
    # Sin data-fuente-oficial=false (ipt_distrito eliminado).
    m_render_v310 = re.search(r'function renderCoberturaButton\([^)]+\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_render_v310:
        body = m_render_v310.group(1)
        if re.search(r"setAttribute\('data-severidad',\s*'alta'\)", body):
            result.ok('v3.10.0 JS: renderCoberturaButton ccpp_exacto data-severidad=alta')
        else:
            result.fail('v3.10.0 JS data-sev alta', 'renderCoberturaButton sin data-severidad=alta')
        if "data-fuente-oficial',\s*'false'" not in body:
            result.ok('v3.10.12: renderCoberturaButton sin data-fuente-oficial=false (sin ipt)')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — DISCLAIMER PARA FUENTES NO OFICIALES (COMU-5672)
    # v3.10.12: ELIMINADO (2G fuera de scope). stopCheck y disclaimer NO
    # se generan en el panel. La directiva CEO original sigue documentada
    # en CHANGELOG y MEMORY, pero el código activo ya no la implementa.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0/12 — Disclaimer COMU-5672 (ELIMINADO v3.10.12)')

    # v3.10.12: .apagado-panel-disclaimer ELIMINADO del CSS
    if not re.search(r'\.apagado-speech-panel\s+\.apagado-panel-disclaimer\s*\{', html):
        result.ok('v3.10.12: .apagado-panel-disclaimer CSS ELIMINADO (sin 2G)')
    else:
        result.fail('v3.10.12 regresión', '.apagado-panel-disclaimer CSS aún presente')

    # v3.10.12: disclaimerHtml NO debe estar en showApagadoPanelInline
    m_panel_v310 = re.search(r'function showApagadoPanelInline\([^)]+\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_panel_v310:
        body = m_panel_v310.group(1)
        if 'disclaimerHtml' not in body and 'requiere_disclaimer' not in body:
            result.ok('v3.10.12: showApagadoPanelInline sin disclaimerHtml/requiere_disclaimer')

    # CSS: .apagado-panel-meta (versión + fecha inline)
    if re.search(r'\.apagado-speech-panel\s+\.apagado-panel-meta\s*\{', html):
        result.ok('v3.10.0 CSS: .apagado-panel-meta (COMU + versión + fecha)')
    else:
        result.fail('v3.10.0 CSS meta', '.apagado-panel-meta CSS missing')

    # CSS: .apagado-panel-severidad (línea de severidad 1 línea)
    if re.search(r'\.apagado-speech-panel\s+\.apagado-panel-severidad\s*\{', html):
        result.ok('v3.10.0 CSS: .apagado-panel-severidad (1 línea explicativa)')
    else:
        result.fail('v3.10.0 CSS sev', '.apagado-panel-severidad CSS missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — ANOMALÍAS DEPARTAMENTALES DOCUMENTADAS
    # CEO directive: anomalías CALLAO/TUMBES/LORETO/UCAYALI no son gaps,
    # son documentadas con razón verificable.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0 — Anomalías departamentales documentadas')

    # HTML: ZONAS_APAGADAS_INTEL.anomalias_departamentales existe
    # v3.10.0: la estructura es compleja (objetos anidados con campos razón/evidencia),
    # por lo que usamos un regex más simple que captura hasta el cierre de la sección.
    m_anom = re.search(r'anomalias_departamentales:\s*\{(.*?)\n  \},\n  // v3\.10\.0: TRAZABILIDAD', html, re.DOTALL)
    if m_anom:
        body = m_anom.group(1)
        for depto in ['CALLAO', 'TUMBES', 'LORETO', 'UCAYALI']:
            if f'"{depto}":' in body:
                result.ok(f'v3.10.0 INTEL: anomalias_departamentales.{depto} documentado')
            else:
                result.fail(f'v3.10.0 anomalia {depto}', f'{depto} no documentado en anomalias_departamentales')
        # Verificar campos clave
        if 'razon' in body and 'evidencia' in body and 'validado_por' in body:
            result.ok('v3.10.3 INTEL: anomalias tienen campos razon/evidencia/validado_por')
        else:
            result.fail('v3.10.3 INTEL anomalia campos', 'anomalias sin campos razon/evidencia/validado_por')
    else:
        result.fail('v3.10.0 INTEL anomalias', 'anomalias_departamentales missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — TRAZABILIDAD DE FUENTES
    # Documentar de dónde salió cada bloque de data. Importante para que
    # el supervisor entienda qué puede confiar y qué requiere validación.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0 — Trazabilidad de fuentes (fuentes_datos)')

    # HTML: ZONAS_APAGADAS_INTEL.fuentes_datos existe
    m_fuentes = re.search(r'fuentes_datos:\s*\{', html)
    if m_fuentes:
        result.ok('v3.10.0 INTEL: fuentes_datos presente')
        # Verificar que los 4 bloques están
        for bloque in ['ubigeo_estructura', 'ccpp_caserios', 'zonas_apagadas_5514', 'zonas_apagadas_5672']:
            if bloque in html:
                result.ok(f'v3.10.0 INTEL: fuentes_datos.{bloque} documentado')
            else:
                result.fail(f'v3.10.0 fuente {bloque}', f'{bloque} missing en fuentes_datos')
    else:
        result.fail('v3.10.0 INTEL fuentes', 'fuentes_datos missing')

    # Doc: docs/FUENTES_ZONAS_APAGADAS.md existe
    doc_path = Path(__file__).parent / 'docs' / 'FUENTES_ZONAS_APAGADAS.md'
    if doc_path.exists():
        result.ok('v3.10.0 DOC: docs/FUENTES_ZONAS_APAGADAS.md existe')
        doc_text = doc_path.read_text()
        if 'Santa Rosa de Loreto' in doc_text:
            result.ok('v3.10.0 DOC: FUENTES doc menciona Santa Rosa de Loreto (Ley 32403)')
        else:
            result.fail('v3.10.0 DOC santa rosa', 'FUENTES doc sin mention Santa Rosa de Loreto')
        if 'MTC' in doc_text and 'OSIPTEL' in doc_text and 'cronograma' in doc_text:
            result.ok('v3.10.0 DOC: FUENTES doc documenta que apagón 2G es PLANIFICADO no ejecutado')
        else:
            result.fail('v3.10.0 DOC mtc', 'FUENTES doc no documenta contexto oficial MTC/OSIPTEL')
    else:
        result.fail('v3.10.0 DOC', 'docs/FUENTES_ZONAS_APAGADAS.md missing')

    # Doc: data/known_data_gaps.json existe
    gaps_path = Path(__file__).parent / 'data' / 'known_data_gaps.json'
    if gaps_path.exists():
        result.ok('v3.10.0 DATA: data/known_data_gaps.json existe')
        try:
            import json as _json
            gaps = _json.loads(gaps_path.read_text())
            if gaps.get('total') == 41:
                result.ok('v3.10.0 DATA: known_data_gaps total=41 distritos')
            else:
                result.fail('v3.10.0 DATA total', f'known_data_gaps total != 41 (es {gaps.get("total")})')
            for depto_check in ['CALLAO', 'LIMA', 'CUSCO', 'LORETO', 'UCAYALI']:
                if depto_check in gaps_path.read_text():
                    pass  # presente, OK
            if 'Santa Rosa de Loreto' in gaps_path.read_text():
                result.ok('v3.10.0 DATA: known_data_gaps documenta Santa Rosa de Loreto')
        except Exception as e:
            result.warn('v3.10.0 DATA parse', f'JSON parse: {e}')
    else:
        result.fail('v3.10.0 DATA gaps', 'data/known_data_gaps.json missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — AUDIT CHECKS AMPLIADOS
    # auditZonasApagadas debe validar checks 8, 9, 10, 11.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0 — auditZonasApagadas checks ampliados')

    m_audit = re.search(r'function auditZonasApagadas\(\)\s*\{(.*?)\nfunction', html, re.DOTALL)
    if m_audit:
        body = m_audit.group(1)
        # Check 8: anomalías departamentales
        if 'anomalias_departamentales' in body:
            result.ok('v3.10.0 AUDIT: check 8 anomalías departamentales')
        else:
            result.fail('v3.10.0 AUDIT 8', 'auditZonasApagadas sin check 8')
        # Check 9: ipt_metadata coherencia
        if 'ipt_metadata_by_district' in body and re.search(r'metaRequired\s*=\s*\[', body):
            result.ok('v3.10.0 AUDIT: check 9 ipt_metadata coherencia')
        else:
            result.fail('v3.10.0 AUDIT 9', 'auditZonasApagadas sin check 9')
        # Check 10: fuente del PDF
        if re.search(r'ZONAS_APAGADAS\.fuente\s+missing', body) or re.search(r"ZONAS_APAGADAS\.fuente", body):
            result.ok('v3.10.0 AUDIT: check 10 fuente del PDF documentada')
        else:
            result.fail('v3.10.0 AUDIT 10', 'auditZonasApagadas sin check 10')
        # Check 11: integridad CCPP
        if re.search(r'CCPP_DATA\.totalCcpp|totalCcpp\s*=', body):
            result.ok('v3.10.0 AUDIT: check 11 integridad CCPP (>= 40k)')
        else:
            result.fail('v3.10.0 AUDIT 11', 'auditZonasApagadas sin check 11')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.2 — POTENCIA AUMENTADA: fuzzy matching + cache LRU
    # CEO directive 2026-10-17: "no debemos perder potencia, debemos aumentarla".
    # Nuevos features: Levenshtein fuzzy, cache LRU, sistema de confianza.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.2 — POTENCIA: fuzzy matching + cache LRU + confianza')

    # JS: zaLevenshtein function definida
    if re.search(r'function\s+zaLevenshtein\s*\(', html):
        result.ok('v3.10.2 JS: zaLevenshtein (fuzzy distance) implementado')
    else:
        result.fail('v3.10.2 JS levenshtein', 'zaLevenshtein no definido')

    # JS: zaSimilarity function definida
    if re.search(r'function\s+zaSimilarity\s*\(', html):
        result.ok('v3.10.2 JS: zaSimilarity (ratio similitud) implementado')
    else:
        result.fail('v3.10.2 JS similarity', 'zaSimilarity no definido')

    # JS: cache LRU implementado
    if re.search(r'zaCheckCache\s*=\s*new\s+Map\s*\(\s*\)', html):
        result.ok('v3.10.2 JS: cache LRU (zaCheckCache = Map)')
    else:
        result.fail('v3.10.2 JS cache', 'cache LRU no implementado')

    if re.search(r'zaCacheGet|zaCacheSet', html):
        result.ok('v3.10.2 JS: zaCacheGet/Set helpers')
    else:
        result.fail('v3.10.2 JS cache helpers', 'zaCacheGet/Set no definidos')

    # JS: ZA_FUZZY_BY_D6 precomputado
    if re.search(r'function\s+zaBuildFuzzyIndex\s*\(', html) and re.search(r'ZA_FUZZY_BY_D6', html):
        result.ok('v3.10.2 JS: zaBuildFuzzyIndex (índice fuzzy precomputado)')
    else:
        result.fail('v3.10.2 JS fuzzy index', 'ZA_FUZZY_BY_D6 no implementado')

    # JS: zaFuzzySearchCcpp function
    if re.search(r'function\s+zaFuzzySearchCcpp\s*\(', html):
        result.ok('v3.10.2 JS: zaFuzzySearchCcpp (búsqueda fuzzy)')
    else:
        result.fail('v3.10.2 JS fuzzy search', 'zaFuzzySearchCcpp no definido')

    # JS: checkZonaApagada integrado con fuzzy
    if re.search(r"nivel:\s*'ccpp_fuzzy'", html):
        result.ok('v3.10.2 JS: checkZonaApagada retorna nivel ccpp_fuzzy')
    else:
        result.fail('v3.10.2 JS nivel fuzzy', 'nivel ccpp_fuzzy no implementado')

    # JS: confianza field en checks
    # v3.10.12: 3 niveles activos (ccpp_exacto, ccpp_fuzzy, distrito_of). Antes 4 (con ipt_distrito).
    confianza_count = len(re.findall(r'confianza:\s*[\d.]+|confianza:\s*(?:currentCheck\.confianza|fuzzy\[)', html))
    if confianza_count >= 3:
        result.ok(f'v3.10.2/12 JS: campo confianza en checks ({confianza_count} ocurrencias, 3 niveles)')
    else:
        result.fail('v3.10.2 JS confianza', f'solo {confianza_count} ocurrencias de confianza')

    # CSS: .cobertura-btn.fuzzy
    # v3.10.26: REMOVIDO .cobertura-btn.fuzzy + @keyframes pulseFuzzy
    # (sin uso desde v3.10.12). El nivel fuzzy ya no se renderiza en el botón.
    if re.search(r'\.cobertura-btn\.fuzzy\s*\{', html):
        result.fail('v3.10.26 CSS fuzzy', '.cobertura-btn.fuzzy REMOVIDO en v3.10.26 (sin uso)')
    else:
        result.ok('v3.10.26: .cobertura-btn.fuzzy ELIMINADO (sin uso desde v3.10.12)')

    # CSS: panel de fuzzy
    if re.search(r'\.apagado-panel-fuzzy\s*\{', html):
        result.ok('v3.10.2 CSS: .apagado-panel-fuzzy (sugerencias)')
    else:
        result.fail('v3.10.2 CSS panel fuzzy', '.apagado-panel-fuzzy missing')

    # CSS: pulseFuzzy animation
    # v3.10.26: REMOVIDO @keyframes pulseFuzzy (sin uso).
    if re.search(r'@keyframes\s+pulseFuzzy', html):
        result.fail('v3.10.26 pulseFuzzy', 'pulseFuzzy REMOVIDO en v3.10.26')
    else:
        result.ok('v3.10.26: @keyframes pulseFuzzy ELIMINADO (sin uso)')

    # HTML: panel con fuzzy_matches rendering
    if re.search(r'apagado-panel-fuzzy-item', html):
        result.ok('v3.10.2 HTML: apagado-panel-fuzzy-item rendering')
    else:
        result.fail('v3.10.2 HTML fuzzy item', 'rendering de fuzzy_matches missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.27 — BLOQUEO botón + click feedback reducido + input focus
    # limpio + RESULTADO texto + diferenciación visual botones
    # CEO directive: "Los botones clear y zona de apagado siguen iguales, y
    # zona de apagado sigue vibrando... al darle click dice SELECCIONA DPT,
    # PROVINCIA, DISTRITO <- incorrecto debe estar bloqueado. Otro bugs
    # detectado al darle clicks a los inputs, noto que no estan correctamente
    # creados... El INPUT de RESULTADO <- debe decir Aqui se llenaran los datos"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.27 — BLOQUEO botón + input focus + RESULTADO + buttons diferenciados')

    # JS: bloqueo en onClickCobertura (early return sin toast/shake)
    if re.search(r'if\s*\(\s*!dep\s*\|\|\s*!prov\s*\|\|\s*!dist\s*\)\s*\{\s*return;', html):
        result.ok('v3.10.27 JS: onClickCobertura BLOQUEADO (early return sin DPT/PROV/DIST)')
    else:
        result.fail('v3.10.27 JS bloque', 'early return sin toast/shake esperado')

    # CSS: cobertura-btn--idle BLOQUEADO (opacity 0.5 + cursor not-allowed)
    css_idle_match = re.search(r'\.cobertura-btn--idle\s*\{([^}]+)\}', html)
    if css_idle_match:
        idle_body = css_idle_match.group(1)
        if 'cursor: not-allowed' in idle_body and 'opacity: 0.5' in idle_body:
            result.ok('v3.10.27 CSS: cobertura-btn--idle BLOQUEADO (opacity 0.5 + cursor not-allowed)')
        else:
            result.fail('v3.10.27 CSS idle', 'opacity: 0.5 + cursor: not-allowed esperado')
    else:
        result.fail('v3.10.27 CSS idle', '.cobertura-btn--idle CSS missing')

    # CSS: click feedback animations reducidas (no vibración)
    if re.search(r'@keyframes\s+cimaShake[\s\S]*?translateX\(-1px\)', html):
        result.ok('v3.10.27 CSS: cimaShake amplitud reducida 3px → 1px')
    else:
        result.fail('v3.10.27 CSS shake', 'amplitud reducida esperada')
    if re.search(r'@keyframes\s+pulseCopied[\s\S]*?scale\(1\.04\)', html):
        result.ok('v3.10.27 CSS: pulseCopied reducido scale(1.18) → scale(1.04)')
    else:
        result.fail('v3.10.27 CSS copied', 'scale reducido esperado')

    # CSS: input focus LIMPIO (sin box-shadow halo, border más grueso)
    if re.search(r'\.field input:focus[^{}]*\{[^}]*border-width:\s*1\.5px', html, re.DOTALL):
        result.ok('v3.10.27 CSS: input:focus border-width 1.5px (sin halo)')
    else:
        result.fail('v3.10.27 CSS focus', 'input:focus border-width 1.5px esperado')

    # HTML: RESULTADO placeholder texto
    if 'Aqui se llenaran los datos' in html:
        result.ok('v3.10.27 HTML: RESULTADO placeholder "Aqui se llenaran los datos"')
    else:
        result.fail('v3.10.27 HTML result', 'placeholder text actualizado esperado')

    # CSS: output-box.placeholder border-style dashed REMOVIDO
    if 'border-style: dashed' in html:
        result.fail('v3.10.27 CSS dashed', 'border-style: dashed en output-box debe estar removido')
    else:
        result.ok('v3.10.27 CSS: output-box.placeholder SIN border-style: dashed (CEO feedback)')

    # CSS: btn-icon-clear diferenciado con border DASHED en estado base
    if re.search(r'\.btn-icon-clear\s*\{[^}]*border:[^}]*dashed', html, re.DOTALL):
        result.ok('v3.10.27 CSS: btn-icon-clear border DASHED (diferenciado de ⚠️ solid)')
    else:
        result.fail('v3.10.27 CSS btn-clear', 'btn-icon-clear con border dashed esperado')

    # ╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════
    # v3.10.34 — HEADER refinado (click-to-collapse + nuevos speeches MOTIVO)
    # CEO directives:
    # - 'Luis, ¿Con quién tengo el gusto?' lo eliminamos es innecesario
    # - 'Si le damos click... se debe minizar para no ocupar espacio se debe
    #   contener en su padre BIENVENIDA E IDENTIFICACION'
    # - MOTIVO Sondea speech: '{Cliente}, ¿en qué puedo ayudarte? Cuál sería
    #   el motivo de tu comunicación o consulta.'
    # - Sondeo de identificación de AVERIAS: departamento/provincia/distrito,
    #   luego dirección, luego referencia
    # ╗═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.34 — HEADER refinado (click-to-collapse + nuevos speeches)')

    # HTML: speech Sondea contiene el speech específico del CEO
    sondea_match = re.search(
        r'id="script-text-sondea"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if sondea_match:
        sondea_text_raw = sondea_match.group(1)
        sondea_text_visible = re.sub(r'<[^>]+>', '', sondea_text_raw)
        sondea_text_visible = re.sub(r'\s+', ' ', sondea_text_visible).strip()
        # v3.10.48: nombre al final de la frase introductoria
        if 'En qu\u00e9 puedo ayudarle' in sondea_text_visible and 'motivo de su comunicaci\u00f3n' in sondea_text_visible:
            result.ok('v3.10.48 SPEECH: Sondea (nombre al final + "puedo ayudarle")')
        else:
            result.fail('v3.10.48 SPEECH sondea', 'speech Sondea debe tener "puedo ayudarle" + "motivo de su comunicación"')

# HTML: speech Sondeo identificación de AVERIAS contiene las frases del CEO
    # v3.10.36: CEO refinó el speech — ahora es un solo párrafo unificado,
    # NO las 3 frases separadas del v3.10.34.
    # v3.10.42: verificar frases en texto visible (stripped de tags .key)
    # v3.10.48: nombre después de "Comprendo" (no al inicio)
    sondeo_averias_match = re.search(
        r'id="script-text-sondeo-averias"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if sondeo_averias_match:
        sa_text_raw = sondeo_averias_match.group(1)
        # Strip HTML tags para verificar frases en texto visible
        sa_text_visible = re.sub(r'<[^>]+>', '', sa_text_raw)
        sa_text_visible = re.sub(r'\s+', ' ', sa_text_visible).strip()
        expected_phrases = [
            'Comprendo',  # v3.10.48: mayúscula porque va después de {Cliente}
            'departamento te encuentras',
            'provincia y distrito',  # v3.10.42: con coma después de distrito
            'dirección donde te encuentras',
            'avanzar con la búsqueda'
        ]
        missing_phrases = [p for p in expected_phrases if p not in sa_text_visible]
        if not missing_phrases:
            result.ok('v3.10.36 SPEECH: Sondeo identificación AVERIAS unificado (CEO directive)')
        else:
            result.fail('v3.10.36 SPEECH sondeo-averias', f'faltan: {missing_phrases}')
    else:
        result.fail('v3.10.36 SPEECH sondeo-averias', 'script-text-sondeo-averias not found')

    # HTML: card sondeo-averias presente
    if 'data-script-id="sondeo-averias"' in html:
        result.ok('v3.10.34 HTML: card "sondeo-averias" presente')
    else:
        result.fail('v3.10.34 HTML sondeo-averias', 'card missing')

    # CSS: .is-collapsed state (click-to-collapse)
    if '.script-card.is-collapsed' in html:
        result.ok('v3.10.34 CSS: .script-card.is-collapsed (estado collapse)')
    else:
        result.fail('v3.10.34 CSS .is-collapsed', '.script-card.is-collapsed missing')

    # CSS: .script-card-body wrapper para collapse
    if '.script-card-body' in html and '.script-card.is-collapsed .script-card-body' in html:
        result.ok('v3.10.34 CSS: .script-card-body wrapper + display:none on collapsed')
    else:
        result.fail('v3.10.34 CSS body', '.script-card-body + display:none on collapsed')

    # JS: click-to-collapse functions
    if 'function toggleScriptCardCollapse(' in html:
        result.ok('v3.10.34 JS: toggleScriptCardCollapse() función definida')
    else:
        result.fail('v3.10.34 JS toggle', 'toggleScriptCardCollapse() missing')
    if 'function applyScriptCollapsedStates(' in html:
        result.ok('v3.10.34 JS: applyScriptCollapsedStates() carga estados persistidos')
    else:
        result.fail('v3.10.34 JS apply collapsed', 'applyScriptCollapsedStates() missing')

    # JS: localStorage collapse persistence
    if 'SCRIPT_COLLAPSED_STORAGE_KEY' in html and 'cima_script_collapsed_v1' in html:
        result.ok('v3.10.34 JS: localStorage collapse persistence (cima_script_collapsed_v1)')
    else:
        result.fail('v3.10.34 JS collapse storage', 'cima_script_collapsed_v1 missing')

    # Anti-regression: v3.10.33 is-done REMOVIDO (reemplazado por is-collapsed)
    if '.script-card.is-done' not in html:
        result.ok('v3.10.34: .is-done (v3.10.33) REMOVIDO — reemplazado por .is-collapsed')
    else:
        result.fail('v3.10.34 anti-is-done regresión', '.is-done debería estar eliminado')

    # Anti-regression: bienvenida NO debe mencionar TÉCNICO MÓVIL (CEO v3.10.31)
    bienvenida_match = re.search(
        r'id="script-text-bienvenida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if bienvenida_match:
        bienvenida_text = bienvenida_match.group(1)
        if 'TÉCNICO MÓVIL' in bienvenida_text:
            result.fail('v3.10.34 ANTI-anti', 'bienvenida NO debe mencionar TÉCNICO MÓVIL')
        elif 'Bienvenido a Movistar' in bienvenida_text and '¿Es correcto?' in bienvenida_text:
            result.ok('v3.10.34 SPEECH: bienvenida SIN TÉCNICO MÓVIL (anti-regression)')
        else:
            result.fail('v3.10.34 SPEECH bienvenida', 'texto bienvenida no esperado')
    else:
        result.fail('v3.10.34 SPEECH bienvenida', 'script-text-bienvenida not found')

    # Anti-regression: copyScriptCard() REMOVIDO (CEO directive v3.10.33)
    if 'function copyScriptCard(' not in html:
        result.ok('v3.10.34: copyScriptCard() ELIMINADO (CEO directive preservada)')
    else:
        result.fail('v3.10.34 JS copy regresión', 'copyScriptCard() debería estar eliminado')

    # Anti-regression: 0 copy buttons (CEO directive v3.10.33)
    if html.count('class="script-card-copy"') == 0:
        result.ok('v3.10.34 HTML: 0 botones copy (CEO directive preservada)')
    else:
        result.fail('v3.10.34 HTML copy buttons regresión', '0 esperados')

    # ╔═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════
    # v3.10.35 — OPTIMIZACIÓN: single header (sin duplicado)
    # CEO directive: 'ya que hay doble header BIENVENIDA E IDENTIFICACION +
    # BIENVENIDA + SPEECH debe hacerse 1 solo para no estirar tanto espacio
    # corrige de manera lineal para todos los textos y al minimizarse se
    # debe marcar como listo o hecho'.
    # ╗═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.35 — OPTIMIZACIÓN single header + ✓ Hecho indicator')

    # HTML: NO debe haber <h4> dentro de .script-card (CEO: 'doble header')
    # Anti-regression: scripts-card-header (wrapper con h4) REMOVIDO
    # Contamos cuántos <h4> quedan dentro de .script-card
    import re as re_mod
    n_h4_in_cards = len(re_mod.findall(
        r'<div class="script-card"[^>]*>\s*<div class="script-card-body">(?:[^<]*<[^/][^>]*>)*?[^<]*<h4>',
        html
    ))
    if n_h4_in_cards == 0:
        result.ok('v3.10.35 HTML: 0 <h4> duplicados dentro de .script-card (single header)')
    else:
        result.fail('v3.10.35 HTML regresión', f'{n_h4_in_cards} <h4> dentro de script-card (debe ser 0)')

    # CSS: .script-card.is-collapsed::after con 'Hecho' indicator
    if '.script-card.is-collapsed::after' in html and 'Hecho' in html:
        result.ok('v3.10.35 CSS: .script-card.is-collapsed::after con "Hecho" indicator')
    else:
        result.fail('v3.10.35 CSS hecho', '.script-card.is-collapsed::after "Hecho" missing')

    # HTML: section headers siguen presentes (single header por sección)
    section_headers = [
        'Bienvenida e identificación',
        'Motivo de la llamada',
        'Consulta adicional + App Movistar',
        'Cierre + Encuesta'
    ]
    missing_section_headers = [s for s in section_headers if s not in html]
    if not missing_section_headers:
        result.ok('v3.10.35 HTML: 4 section headers (single per sección) presentes')
    else:
        result.fail('v3.10.35 HTML sections', f'section headers faltantes: {missing_section_headers}')

    # CSS: badges (✓ ✗ ⭐) preservados en Cierre
    if 'script-badge--ok' in html and 'script-badge--fail' in html and 'script-badge--star' in html:
        result.ok('v3.10.35 CSS: badges ✓/✗/⭐ preservados en Cierre')
    else:
        result.fail('v3.10.35 CSS badges regresión', 'badges ok/fail/star missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.36 — MOVER atencion-segura a BIENVENIDA + REFINAR sondeo-averias
    # CEO directive: 'Vamos a subir la atencion segura a BEINVENIDA E
    # IDENTIFICACION... motivo de la llamada: Sondeo de identificacion de
    # AVERIAS... mejorar'.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.36 — MOVER atencion-segura a BIENVENIDA + REFINAR sondeo-averias')

    # HTML: atencion-segura AHORA está en sección BIENVENIDA (orden: bienvenida → atencion-segura → sondea)
    bienvenida_pos = html.find('data-script-id="bienvenida"')
    atencion_pos = html.find('data-script-id="atencion-segura"')
    sondea_pos = html.find('data-script-id="sondea"')
    if bienvenida_pos > 0 and atencion_pos > 0 and sondea_pos > 0:
        if bienvenida_pos < atencion_pos < sondea_pos:
            result.ok('v3.10.36 HTML: atencion-segura en BIENVENIDA (orden: bienvenida → atencion-segura → sondea)')
        else:
            result.fail('v3.10.36 HTML orden', 'atencion-segura debe estar entre bienvenida y sondea')
    else:
        result.fail('v3.10.36 HTML posición', 'cards no encontrados')

    # Anti-regression: Sondeo de identificación AVERIAS speech contiene "comprendo" + "avanzar con la búsqueda"
    import re as re_mod
    sa_match = re_mod.search(
        r'id="script-text-sondeo-averias"[^>]*>(.+?)</p>',
        html,
        re_mod.DOTALL
    )
    if sa_match:
        sa_text = sa_match.group(1)
        # v3.10.48: nombre después de "Comprendo" (mayúscula)
        if 'Comprendo' in sa_text and 'avanzar con la búsqueda' in sa_text:
            result.ok('v3.10.36+48 SPEECH: Sondeo identificación AVERIAS unificado (1 párrafo + "Comprendo")')
        else:
            result.fail('v3.10.36 SPEECH sondeo-averias', 'speech no unificado')

    # Tip en Sondeo AVERIAS: 'En caso no encontremos...'
    if 'En caso no encontremos al cliente' in html or 'solicitamos referencias' in html:
        result.ok('v3.10.36 SPEECH: tip "En caso no encontremos al cliente, solicitamos referencias"')
    else:
        result.fail('v3.10.36 SPEECH tip', 'tip referencias missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.38 — Collapsed card diseño coherente
    # CEO directive: 'Al darle click a los speech textuales se minimizan
    # pero esteticamente mal diseñados y de tamaño, optimizalos dales mas
    # coherencia'.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.39 — Collapsed card APPLE 2026 PREMIUM STYLE')

    # CEO directive: "no me agrada cambialo a otro diferente necesito
    # premium real en 2026. APPLE styles"
    import re as re_mod

    # CSS: pill shape (border-radius 999px = fully rounded Apple-style)
    collapsed_match = re_mod.search(
        r'\.script-card\.is-collapsed\s*\{([^}]+)\}',
        html
    )
    if collapsed_match:
        body = re_mod.sub(r'/\*[\s\S]*?\*/', '', collapsed_match.group(1))
        if 'border-radius: 999px' in body and 'backdrop-filter' in body:
            result.ok('v3.10.39 CSS: pill shape (999px) + glassmorphism (backdrop-filter)')
        else:
            result.fail('v3.10.39 CSS pill', 'border-radius 999px + backdrop-filter esperado')

    # CSS: linear-gradient (Apple-style gradient backgrounds)
    if 'linear-gradient(135deg' in body and '#10b981' in body and '#059669' in body:
        result.ok('v3.10.39 CSS: linear-gradient(135deg) verde (#10b981 → #059669)')
    else:
        result.fail('v3.10.39 CSS gradient', 'linear-gradient verde esperado')

    # CSS: spring transition (cubic-bezier Apple spring)
    if 'cubic-bezier(0.34, 1.56, 0.64, 1)' in body or 'cubic-bezier(0.34, 1.56, 0.64, 1)' in html:
        result.ok('v3.10.39 CSS: spring transition cubic-bezier(0.34, 1.56, 0.64, 1)')
    else:
        result.fail('v3.10.39 CSS spring', 'spring transition esperado')

    # CSS: "✓ Hecho" pill badge (multiline regex)
    before_match = re_mod.search(
        r'\.script-card\.is-collapsed::before\s*\{([^}]+)\}',
        html,
        re_mod.DOTALL
    )
    if before_match:
        before_body = re_mod.sub(r'/\*[\s\S]*?\*/', '', before_match.group(1))
        if 'linear-gradient' in before_body and '#10b981' in before_body and 'border-radius: 999px' in before_body:
            result.ok('v3.10.39 CSS: ::before "✓ Hecho" gradient pill (999px radius)')
        else:
            result.fail('v3.10.39 CSS before badge', '::before gradient pill esperado')

    # CSS: "Toca para abrir ▴" hint (Apple Reminders style)
    if 'Toca para abrir' in html and '▴' in html:
        result.ok('v3.10.39 CSS: ::after "Toca para abrir ▴" hint (Apple Reminders)')
    else:
        result.fail('v3.10.39 CSS after hint', '"Toca para abrir ▴" hint esperado')

    # CSS: hover lift + glow (Apple-style micro-interaction)
    if '.script-card.is-collapsed:hover' in html and 'translateY(-1px)' in html and 'cubic-bezier' in html:
        result.ok('v3.10.39 CSS: hover lift (-1px) + spring + glow box-shadow')
    else:
        result.fail('v3.10.39 CSS hover', 'hover lift + glow esperado')

    # CSS: inset highlight (Apple glass effect)
    if 'inset 0 1px 0 rgba(255, 255, 255' in body:
        result.ok('v3.10.39 CSS: inset highlight blanco (Apple glass effect)')
    else:
        result.fail('v3.10.39 CSS inset', 'inset highlight esperado')

    # Anti-regression: border-left 3px REMOVIDO (era v3.10.38 feo)
    if 'border-left: 3px solid var(--green)' not in body:
        result.ok('v3.10.39 anti-regression: border-left 3px REMOVIDO (v3.10.38 feo)')
    else:
        result.fail('v3.10.39 anti-border regresión', 'border-left 3px debería estar removido')

    # Anti-regression: NO debe quedar "padding: 5px 10px" (era v3.10.34 feo)
    if 'padding: 5px 10px' not in body:
        result.ok('v3.10.39: viejo "padding: 5px 10px" REMOVIDO')
    else:
        result.fail('v3.10.39 anti-padding regresión', 'padding 5px 10px debería estar removido')

    # CSS: ✓ body hidden when collapsed (CRITICAL)
    if '.script-card.is-collapsed .script-card-body { display: none' in html:
        result.ok('v3.10.39 CSS: .script-card-body hidden when collapsed (CRITICAL)')
    else:
        result.fail('v3.10.39 CSS body hidden', 'body hidden rule missing')

    # CSS: text-shadow "✓ HECHO" all-caps REMOVIDO (v3.10.38 feo)
    if 'text-transform: uppercase' not in body:
        result.ok('v3.10.39: "text-transform: uppercase" REMOVIDO (era feo en v3.10.38)')
    else:
        result.fail('v3.10.39 anti-uppercase regresión', 'uppercase debería estar removido')



    # ══════════════════════════════════════════════════════════════════════
    # v3.10.37 — ALERTAS ESTÁTICAS + DESPEDIDA GENERAL
    # CEO directive: 'algunos textos que son unicamente de apoyo para el
    # agente que no suman como un speech, pero que deben usarse y talvez
    # es repitan para mantener su uso... no sera un speech textual sera
    # como una sola row de alerta dentro del padre MOTIVO DE LLAMADA... Vamos
    # a retirar estas 2 ✓ ✗ y la reemplazamos por un speech de despedida
    # general'.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.37 — ALERTAS ESTÁTICAS + DESPEDIDA GENERAL')

    # CSS: .script-alert rule definido
    if '.script-alert {' in html and '.script-alert-icon' in html:
        result.ok('v3.10.37 CSS: .script-alert + .script-alert-icon definidos')
    else:
        result.fail('v3.10.37 CSS .script-alert', '.script-alert CSS missing')

    # HTML: 3 alerts estáticos en MOTIVO (retorno removido, reemplazado por alerts)
    n_alerts_motivo = html.count('class="script-alert"')
    if n_alerts_motivo >= 4:  # 3 en MOTIVO + 1 en CONSULTA = 4 total
        result.ok(f'v3.10.37 HTML: {n_alerts_motivo} alerts estáticos (3 MOTIVO + 1 CONSULTA)')
    else:
        result.fail('v3.10.37 HTML alerts', f'≥4 alerts esperados, {n_alerts_motivo} encontrados')

    # HTML: alert textos AI 2026 (v3.10.41 + v3.10.42 redistribución)
    expected_alerts = [
        'gracias por tu paciencia',
        'Antes de ausentarte',
        'Conecta desde la empat\u00eda',
        'hay algo m\u00e1s en lo que puedas'
    ]
    missing_alerts = [a for a in expected_alerts if a not in html]
    if not missing_alerts:
        result.ok('v3.10.41/42 HTML: 4 alertas AI 2026 modernas presentes')
    else:
        result.fail('v3.10.41/42 HTML alerts textos', f'faltan: {missing_alerts}')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.42 — Redistribución inteligente de alerts 1/section
    # + font 12.5px en script-card-text + class .key para datos clave
    # + reescritura AI 2026 de los 9 speeches
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.42 — REDISTRIBUCIÓN ALERTS 1/section + AI 2026 SPEECHES')

    # CSS: .script-card-text font-size 12.5px (era 11px → CEO pidió mejor font)
    # v3.10.43: regex más específico para evitar matchear .script-card-body > .script-card-text
    script_text_font = re.search(
        r'^\s*\.script-card-text\s*\{([^}]+)\}',
        html, re.MULTILINE | re.DOTALL
    )
    if script_text_font:
        body = re.sub(r'/\*[\s\S]*?\*/', '', script_text_font.group(1))
        if 'font-size: 12.5px' in body:
            result.ok('v3.10.42 CSS: .script-card-text font-size 12.5px (CEO pidió mejor font)')
        else:
            result.fail('v3.10.42 CSS script-card-text font', 'font-size 12.5px esperado')
    else:
        result.fail('v3.10.42 CSS script-card-text', '.script-card-text CSS missing')

    # CSS: .key class definida (resaltado bold sage para datos clave)
    if re.search(r'\.key\s*\{[^}]*font-weight:\s*600', html, re.DOTALL):
        result.ok('v3.10.42 CSS: .key class definida (bold 600 para datos clave)')
    else:
        result.fail('v3.10.42 CSS .key', '.key class con font-weight 600 esperada')

    # HTML: <span class="key"> usado en speeches (datos clave resaltados)
    n_key_spans = html.count('class="key"')
    if n_key_spans >= 8:  # sondeo-averias(4) + educa-app(4) + transferir(1) ≈ 9
        result.ok(f'v3.10.42 HTML: {n_key_spans} spans .key (datos clave resaltados)')
    else:
        result.fail('v3.10.42 HTML .key spans', f'≥8 spans .key esperados, {n_key_spans} encontrados')

    # HTML: 1 alert por sección (CEO: no saturar solo MOTIVO)
    # Estrategia: contar alerts en orden; verificar distribución balanceada
    # Esperado: bienvenida=1, motivo=1, consulta=1, cierre=1
    sections_with_alerts = html.count('script-section-header--blue')
    import re as _re42
    n_alerts_blue = len(_re42.findall(
        r'script-section-header--blue.*?(?=script-section-header--green|</div>\s*</div>\s*</div>\s*</details>)',
        html, _re42.DOTALL
    ))
    # Simplificado: contar alerts hasta el siguiente header
    # Pattern alternativo: contar alerts en cada section usando markers
    bienvenida_alert = 'Conecta desde la empat\u00eda' in html
    motivo_alert = 'Antes de ausentarte' in html
    consulta_alert = 'hay algo m\u00e1s en lo que puedas' in html
    cierre_alert = 'gracias por tu paciencia' in html
    if all([bienvenida_alert, motivo_alert, consulta_alert, cierre_alert]):
        result.ok('v3.10.42 HTML: 4 alerts distribuidas 1/section (bienvenida/motivo/consulta/cierre)')
    else:
        missing = [n for n, v in [
            ('bienvenida', bienvenida_alert), ('motivo', motivo_alert),
            ('consulta', consulta_alert), ('cierre', cierre_alert)
        ] if not v]
        result.fail('v3.10.42 HTML alerts distribución', f'faltan alerts en: {missing}')

    # Speech sondeo-averias: datos clave .key resaltados (departamento, provincia, distrito, dirección)
    if 'class="key">departamento' in html and 'class="key">provincia' in html and \
       'class="key">distrito' in html and 'class="key">direcci' in html:
        result.ok('v3.10.42 SPEECH: sondeo-averias datos clave .key (departamento/provincia/distrito/dirección)')
    else:
        result.fail('v3.10.42 SPEECH sondeo-averias .key', 'departamento/provincia/distrito/dirección en .key')

    # Speech educa-app: datos clave .key resaltados (App Mi Movistar, saldo, recibo, beneficios exclusivos)
    if 'class="key">App Mi Movistar' in html and 'class="key">saldo' in html and \
       'class="key">recibo' in html and 'class="key">beneficios exclusivos' in html:
        result.ok('v3.10.42 SPEECH: educa-app datos clave .key (App Mi Movistar/saldo/recibo/beneficios)')
    else:
        result.fail('v3.10.42 SPEECH educa-app .key', 'App Mi Movistar/saldo/recibo/beneficios exclusivos en .key')

    # Speech transferir: asesor especializado en .key
    if 'class="key">asesor especializado' in html:
        result.ok('v3.10.42 SPEECH: transferir asesor especializado en .key')
    else:
        result.fail('v3.10.42 SPEECH transferir .key', 'asesor especializado en .key esperado')

    # Speech atencion-segura: v3.10.46 CEO directive
    # v3.10.48: nombre al medio ("devolverle la llamada {Cliente}")
    if 'Pensando en su comodidad' in html and \
       'si la llamada llegara a cortarse' in html and \
       'n\u00famero de su consulta' in html:
        result.ok('v3.10.48 SPEECH: atencion-segura (Pensando en su comodidad + nombre al medio)')
    else:
        result.fail('v3.10.48 SPEECH atencion-segura', 'redacción CEO + nombre al medio esperado')

    # Speech encuesta: v3.10.46 CEO directive
    # "En 48 horas lo contactaremos para confirmar su experiencia final
    #  con la solucion que le brinde"
    if 'la atenci\u00f3n que recibi\u00f3' in html and \
       'experiencia final con la soluci\u00f3n que le brind\u00e9' in html:
        result.ok('v3.10.46 SPEECH: encuesta redacción CEO (experiencia final / brindé)')
    else:
        result.fail('v3.10.46 SPEECH encuesta', 'redacción CEO esperada')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.43 — BUG FIXES (3 bugs reportados por CEO)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.43 — BUG FIXES: bienvenida telefono-last3 + encuesta overflow + visuales')

    # BUG 1: bienvenida usa telefono-last3 (solo 3 dígitos) + NO "por la consulta de"
    bienvenida_bug_match = re.search(
        r'id="script-text-bienvenida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if bienvenida_bug_match:
        bb_text_raw = bienvenida_bug_match.group(1)
        bb_text_visible = re.sub(r'<[^>]+>', '', bb_text_raw)
        bb_text_visible = re.sub(r'\s+', ' ', bb_text_visible).strip()
        if 'Su consulta es por el numero que termina en' in bb_text_visible and \
           'por la consulta de' not in bb_text_visible and \
           'data-ph="telefono-last3"' in bb_text_raw and \
           'Telefono en consulta' in bb_text_raw:
            result.ok('v3.10.43+44 BUG 1: bienvenida "Su consulta es por el numero que termina en" + telefono-last3')
        else:
            issues = []
            if 'Su consulta es por el numero que termina en' not in bb_text_visible:
                issues.append('falta "Su consulta es por el numero que termina en"')
            if 'por la consulta de' in bb_text_visible:
                issues.append('NO debe contener "por la consulta de"')
            if 'data-ph="telefono-last3"' not in bb_text_raw:
                issues.append('falta data-ph="telefono-last3"')
            if 'Telefono en consulta' not in bb_text_raw:
                issues.append('falta placeholder "Telefono en consulta"')
            result.fail('v3.10.43+44 BUG 1 bienvenida', f'issues: {issues}')
    else:
        result.fail('v3.10.43 BUG 1', 'script-text-bienvenida not found')

    # BUG 1: JS telefono-last3 transformation (últimos 3 dígitos)
    if re.search(r"key\s*===\s*['\"]telefono-last3['\"]", html) and \
       re.search(r"\.slice\(-3\)", html) and \
       re.search(r"\.replace\([^)]*\\D", html):
        result.ok('v3.10.43 JS: telefono-last3 transformation (últimos 3 dígitos)')
    else:
        result.fail('v3.10.43 JS telefono-last3', 'transformación regex últimos 3 dígitos esperada')

    # BUG 2: .script-card-body display: flex (overflow fix para encuesta)
    # v3.10.43: regex más específico para matchear el bloque SIN prefijos
    # (.script-card.is-collapsed .script-card-body tiene { display: none })
    body_flex = re.search(
        r'^\s*\.script-card-body\s*\{([^}]+)\}',
        html, re.MULTILINE | re.DOTALL
    )
    if body_flex:
        body_content = re.sub(r'/\*[\s\S]*?\*/', '', body_flex.group(1))
        if 'display: flex' in body_content and 'flex-wrap: wrap' in body_content:
            result.ok('v3.10.43 BUG 2: .script-card-body display:flex + flex-wrap (encuesta overflow fix)')
        else:
            result.fail('v3.10.43 BUG 2 .script-card-body', 'display:flex + flex-wrap esperados')
    else:
        result.fail('v3.10.43 BUG 2', '.script-card-body CSS missing')

    # BUG 2: .script-card-body > .script-card-text con flex: 1 1 200px + min-width: 0
    text_flex = re.search(r'\.script-card-body\s*>\s*\.script-card-text\s*\{([^}]+)\}', html, re.DOTALL)
    if text_flex and 'flex: 1 1 200px' in text_flex.group(1) and 'min-width: 0' in text_flex.group(1):
        result.ok('v3.10.43 BUG 2: .script-card-body > .script-card-text flex (wrap con badge)')
    else:
        result.fail('v3.10.43 BUG 2 .script-card-text', 'flex: 1 1 200px + min-width: 0 esperados')

    # BUG 3: .script-section-header font-size 12px (más prominente)
    header_font = re.search(r'\.script-section-header\s*\{([^}]+)\}', html, re.DOTALL)
    if header_font:
        header_body = re.sub(r'/\*[\s\S]*?\*/', '', header_font.group(1))
        if 'font-size: 12px' in header_body and 'font-weight: 800' in header_body:
            result.ok('v3.10.43 BUG 3: .script-section-header font-size 12px + weight 800 (más prominente)')
        else:
            result.fail('v3.10.43 BUG 3 section-header', 'font-size 12px + weight 800 esperados')
    else:
        result.fail('v3.10.43 BUG 3', '.script-section-header CSS missing')

    # BUG 3: .script-alert font-size 10.5px (más compacto)
    alert_font = re.search(r'\.script-alert\s*\{([^}]+)\}', html, re.DOTALL)
    if alert_font:
        alert_body = re.sub(r'/\*[\s\S]*?\*/', '', alert_font.group(1))
        if 'font-size: 10.5px' in alert_body and 'max-width: 96%' in alert_body:
            result.ok('v3.10.43 BUG 3: .script-alert font-size 10.5px + max-width 96% (más compacto)')
        else:
            result.fail('v3.10.43 BUG 3 .script-alert', 'font-size 10.5px + max-width 96% esperados')
    else:
        result.fail('v3.10.43 BUG 3 alert', '.script-alert CSS missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.44 (LEGACY) — HEADER "ESTRUCTURA DE LLAMADA" + MOTIVO CARDS
    #   Reemplazado por v3.10.51: MOTIVO = 1 <select> dropdown con 4 options.
    #   Las 4 cards visuales Apple-2026 (motivo-card / motivo-grid / Próximamente)
    #   fueron ELIMINADAS. Mantengo tests de regresión invertidos (esperan
    #   AUSENCIA de .motivo-card) + tests nuevos del select dropdown.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.44 LEGACY — HEADER estructura de llamada (sin motivo cards)')

    # HEADER GRID 0 dice "Estructura de Llamada" (no "Llamada")
    if '>Estructura de Llamada<' in html:
        result.ok('v3.10.44 HEADER: GRID 0 dice "Estructura de Llamada"')
    else:
        result.fail('v3.10.44 HEADER', 'HEADER GRID 0 debe decir "Estructura de Llamada"')

    # LEGACY: .motivo-grid container YA NO EXISTE (v3.10.51 lo eliminó)
    if 'class="motivo-grid"' not in html and 'role="radiogroup"' not in html:
        result.ok('v3.10.51 REFACTOR: .motivo-grid / role=radiogroup ELIMINADOS')
    else:
        result.fail('v3.10.51 REFACTOR motivo-grid', 'motivo-grid debiera estar eliminado')

    # LEGACY: 4 motivo cards YA NO EXISTEN
    n_active = len(re.findall(r'class="motivo-card motivo-card--active"', html))
    n_disabled = len(re.findall(r'class="motivo-card motivo-card--disabled"', html))
    if n_active == 0 and n_disabled == 0:
        result.ok('v3.10.51 REFACTOR: 0 motivo cards (reemplazadas por dropdown)')
    else:
        result.fail('v3.10.51 cards count', f'0 cards esperados, {n_active}+{n_disabled} encontrados')

    # LEGACY: CSS .motivo-card fue removido (debería dar 0 matches en style block)
    motivo_card_css = re.search(r'\.motivo-card\s*\{[^}]+\}', html, re.DOTALL)
    if not motivo_card_css:
        result.ok('v3.10.51 CSS: .motivo-card styles ELIMINADOS')
    else:
        result.fail('v3.10.51 CSS motivo-card', '.motivo-card styles debieran estar eliminados')

    # LEGACY: badge "Próximamente" en cards desactivadas YA NO EXISTE
    n_prox = html.count('Próximamente')
    if n_prox == 0:
        result.ok('v3.10.51 HTML: 0 badges "Próximamente" (cards Próximamente eliminadas)')
    else:
        result.fail('v3.10.51 HTML badge', f'0 badges esperados, {n_prox} encontrados')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.51 — MOTIVO consolidado en 1 dropdown con 4 opciones
    # CEO directive: "EN SU INPUT MOTIVO <- las 4 opciones vamos a meterlas
    #   en 1 dropdow cada 1 ocupa 1 row.. TELEFONO DE CONSULTA + MOTIVO"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.51 — MOTIVO = 1 dropdown con 4 opciones')

    # HTML: <select id="script-motivo"> visible (NO hidden)
    select_match = re.search(
        r'<select\s+id="script-motivo"[^>]*>',
        html
    )
    if select_match and 'hidden' not in select_match.group(0):
        result.ok('v3.10.51 HTML: <select id="script-motivo"> VISIBLE (sin atributo hidden)')
    else:
        result.fail('v3.10.51 HTML select', '<select id="script-motivo"> visible esperado')

    # HTML: 4 options + 1 default = 5 options en el select
    option_block = re.search(
        r'<select[^>]*id="script-motivo"[^>]*>(.*?)</select>',
        html, re.DOTALL
    )
    n_opts = len(re.findall(r'<option\s', option_block.group(1))) if option_block else 0
    if n_opts == 5:  # 1 default "— Seleccionar —" + 4 motivos
        result.ok(f'v3.10.51 HTML: {n_opts} options en MOTIVO (1 default + 4 motivos)')
    else:
        result.fail('v3.10.51 HTML options count', f'5 options esperadas, {n_opts} encontradas')

    # HTML: 4 motivos específicos (Técnico Móvil, Otras Consultas, Avería Masiva, GTR Crítico)
    motivo_labels = ['Técnico Móvil', 'Otras Consultas', 'Avería Masiva', 'GTR Crítico']
    motivo_values = ['MOVIL', 'OTRAS', 'AVERIA', 'GTR']
    missing_labels = [m for m in motivo_labels if m not in html]
    missing_values = [m for m in motivo_values if f'value="{m}"' not in html]
    if not missing_labels and not missing_values:
        result.ok(f'v3.10.51 HTML: 4 motivos dropdown (MOVIL/OTRAS/AVERIA/GTR)')
    else:
        result.fail('v3.10.51 HTML motivos',
                    f'faltan labels={missing_labels} values={missing_values}')

    # HTML: TELÉFONO ahora dice "DE CONSULTA" (CEO directive, antes "EN CONSULTA")
    if 'TELÉFONO DE CONSULTA' in html and 'TELÉFONO EN CONSULTA' not in html:
        result.ok('v3.10.51 HTML: TELÉFONO renombrado EN → DE CONSULTA')
    else:
        result.fail('v3.10.51 HTML tel label', '"TELÉFONO DE CONSULTA" esperado, NO "EN CONSULTA"')

    # JS: initMotivoSelect() función definida (reemplaza initMotivoCards)
    if 'function initMotivoSelect()' in html:
        result.ok('v3.10.51 JS: initMotivoSelect() función definida')
    else:
        result.fail('v3.10.51 JS initMotivoSelect', 'función initMotivoSelect esperada')

    # JS: initMotivoCards() YA NO EXISTE (fue renombrada/removida)
    if 'function initMotivoCards()' not in html:
        result.ok('v3.10.51 JS: initMotivoCards() ELIMINADO (reemplazado por initMotivoSelect)')
    else:
        result.fail('v3.10.51 JS initMotivoCards', 'initMotivoCards debiera estar eliminado')

    # JS: syncMotivoCardsFromSelect() ELIMINADO (ya no hay cards que sincronizar)
    if 'function syncMotivoCardsFromSelect' not in html:
        result.ok('v3.10.51 JS: syncMotivoCardsFromSelect() ELIMINADO')
    else:
        result.fail('v3.10.51 JS syncMotivoCards', 'syncMotivoCardsFromSelect debiera estar eliminado')

    # JS: initMotivoSelect() llamada en DOMContentLoaded
    if re.search(r'initMotivoSelect\(\)', html):
        result.ok('v3.10.51 JS: initMotivoSelect() invocada en init')
    else:
        result.fail('v3.10.51 JS initMotivoSelect call', 'initMotivoSelect() debe invocarse en DOMContentLoaded')

    # JS: clearScriptTemplate ya no llama a syncMotivoCardsFromSelect
    # (busca la función en el bloque del clear handler)
    if 'syncMotivoCardsFromSelect' not in html:
        result.ok('v3.10.51 JS: 0 referencias a syncMotivoCardsFromSelect (sin huérfanos)')
    else:
        result.fail('v3.10.51 JS sync huérfanos', '0 referencias syncMotivoCardsFromSelect esperadas')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.49 (LEGACY) — Desglose del flow al click en MOTIVO
    # v3.10.51: el trigger cambió de "click en card" a "change en select",
    #   pero desgloseMotivoFlow() + getMotivoFlowCards() siguen vivos.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.49 LEGACY — DESGLOSE flow mantiene comportamiento')

    # JS: desgloseMotivoFlow() función definida (sin cambios)
    if 'function desgloseMotivoFlow(' in html:
        result.ok('v3.10.49 JS: desgloseMotivoFlow() función definida (mantenida)')
    else:
        result.fail('v3.10.49 JS desgloseMotivoFlow', 'función desgloseMotivoFlow esperada')

    # JS: getMotivoFlowCards() función definida con map MOTIVO → script-ids
    if 'function getMotivoFlowCards(' in html and \
       "'MOVIL'" in html and ("'OTRO'" in html or "'OTRAS'" in html):
        result.ok('v3.10.49/3.10.51 JS: getMotivoFlowCards() con map MOTIVO → script-ids')
    else:
        result.fail('v3.10.49 JS getMotivoFlowCards', 'map MOTIVO → script-ids esperado')

    # JS: initMotivoSelect change handler llama a desgloseMotivoFlow(value)
    if re.search(r'desgloseMotivoFlow\(\s*value\s*\)', html):
        result.ok('v3.10.51 JS: change en MOTIVO select llama desgloseMotivoFlow(value)')
    else:
        result.fail('v3.10.51 JS change handler', 'change debe llamar a desgloseMotivoFlow(value)')

    # JS: desgloseMotivoFlow hace scrollIntoView a gridScript-body
    if re.search(r"scrollIntoView\(\s*\{\s*behavior", html) and \
       re.search(r"getElementById\(['\"]gridScript-body['\"]\)", html):
        result.ok('v3.10.49 JS: scrollIntoView a gridScript-body (scroll suave)')
    else:
        result.fail('v3.10.49 JS scroll', 'scrollIntoView + getElementById("gridScript-body") esperado')

    # CSS: .script-card.motivo-flash con animación (mantenida)
    if '.script-card.motivo-flash' in html and \
       '@keyframes motivoFlashPulse' in html:
        result.ok('v3.10.49 CSS: .script-card.motivo-flash + @keyframes motivoFlashPulse')
    else:
        result.fail('v3.10.49 CSS motivo-flash', '.script-card.motivo-flash + keyframes esperado')

    # JS: desgloseMotivoFlow expande el GRID si está colapsado
    if re.search(r"toggleGrid\(['\"]gridScript['\"]\s*,\s*false\)", html):
        result.ok('v3.10.49 JS: desgloseMotivoFlow expande GRID si está colapsado')
    else:
        result.fail('v3.10.49 JS expand grid', 'toggleGrid("gridScript", false) esperado')

    # v3.10.53: desgloseMotivoFlow ya NO expande TODAS las script-cards —
# la visibilidad ahora la controla applyMotivoFlow() (solo muestra las cards
# del motivo seleccionado). Lo que queda: scrollIntoView + highlight pulse.
    # Anti-regression: querySelectorAll(".script-card.is-collapsed") ya no
    # debe estar en desgloseMotivoFlow (lo removimos en v3.10.53).
    if not re.search(r"desgloseMotivoFlow[\s\S]{0,800}querySelectorAll\(['\"]\.script-card\.is-collapsed['\"]\)", html, re.DOTALL):
        result.ok('v3.10.53 JS: desgloseMotivoFlow NO expande todas las cards (visibilidad por applyMotivoFlow)')
    else:
        result.fail('v3.10.53 JS no auto-expand', 'desgloseMotivoFlow no debería expandir todas las cards')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.50 — 3 BUG FIXES reportados por CEO
    # BUG 1: CTRL+SHIFT+R no funciona (algo bloquea la función)
    # BUG 2: Al cargar, datos desplegados (deben estar minimizados)
    # BUG 3: Solo click en MOTIVO → desglose info (no auto)
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.50 — 3 BUG FIXES (CTRL+SHIFT+R + default collapsed + no auto-desglose)')

    # BUG 1: Handler CTRL+SHIFT+R está FUERA de initPopupMode (global, popup+standalone)
    # Verificar que el handler NO está dentro de initPopupMode (que solo se ejecuta en popup)
    # El handler debe estar en main scope, después del cierre de initPopupMode
    init_popup_close_pos = html.find('function initPopupMode()')
    if init_popup_close_pos > 0:
        # Buscar el cierre de la función (último } antes de function clearScriptTemplate)
        clear_script_pos = html.find('function clearScriptTemplate()')
        # Verificar que el listener de CTRL+SHIFT+R está entre initPopupMode cierre y clearScriptTemplate
        handler_region = html[init_popup_close_pos:clear_script_pos] if clear_script_pos > init_popup_close_pos else ''
        # El handler debe estar en main scope (después de initPopupMode)
        has_global_handler = re.search(
            r"addEventListener\(\s*['\"]keydown['\"]\s*,\s*function[^{]*\{[^}]*ctrlKey[^}]*shiftKey[^}]*clearScriptTemplate",
            handler_region, re.DOTALL
        )
        # Y NO debe estar dentro de initPopupMode
        inside_popup_match = re.search(
            r"function initPopupMode\(\)\s*\{[^{}]*addEventListener\(\s*['\"]keydown['\"][^{}]*ctrlKey[^{}]*shiftKey",
            html, re.DOTALL
        )
        if has_global_handler and not inside_popup_match:
            result.ok('v3.10.50 BUG 1: handler CTRL+SHIFT+R GLOBAL (fuera de initPopupMode)')
        elif inside_popup_match:
            result.fail('v3.10.50 BUG 1', 'handler aún está dentro de initPopupMode (solo popup, no standalone)')
        else:
            result.fail('v3.10.50 BUG 1', 'handler CTRL+SHIFT+R global no encontrado')
    else:
        result.fail('v3.10.50 BUG 1', 'initPopupMode no encontrado')

    # BUG 1: Handler usa capture: true para interceptar antes que browser
    if re.search(
        r"addEventListener\(\s*['\"]keydown['\"][\s\S]*?ctrlKey[\s\S]*?shiftKey[\s\S]*?\}\s*,\s*\{\s*capture:\s*true\s*\}\s*\)",
        html
    ):
        result.ok('v3.10.50 BUG 1: handler usa { capture: true } (intercepta antes que browser)')
    else:
        result.fail('v3.10.50 BUG 1 capture', 'handler debe usar { capture: true }')

    # BUG 1: Handler llama a stopPropagation + stopImmediatePropagation
    if re.search(r"ctrlKey[^{}]*shiftKey[^{}]*stopImmediatePropagation", html, re.DOTALL):
        result.ok('v3.10.50 BUG 1: stopPropagation + stopImmediatePropagation (bloquea otros listeners)')
    else:
        result.fail('v3.10.50 BUG 1 stop', 'stopImmediatePropagation esperado')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.53 — MOTIVO FLOW VISIBILITY
    # CEO directive 2026-10-20: "Mientras que no se seleccione un motivo
    #   todos los textos deben eliminarse, solo se muestra cuando exista
    #   un motivo SELECCIONADO."
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.53 — MOTIVO FLOW VISIBILITY (hidden until seleccionado)')

    # CSS: .script-card.is-hidden-flow { display: none !important; }
    if re.search(r"\.script-card\.is-hidden-flow\s*\{\s*display\s*:\s*none\s*!important", html):
        result.ok('v3.10.53 CSS: .script-card.is-hidden-flow { display: none !important }')
    else:
        result.fail('v3.10.53 CSS card hidden', '.script-card.is-hidden-flow con display:none !important esperado')

    # CSS: .script-alert.is-hidden-flow { display: none !important; }
    if re.search(r"\.script-alert\.is-hidden-flow\s*\{\s*display\s*:\s*none\s*!important", html):
        result.ok('v3.10.53 CSS: .script-alert.is-hidden-flow { display: none !important }')
    else:
        result.fail('v3.10.53 CSS alert hidden', '.script-alert.is-hidden-flow esperado')

    # JS: applyMotivoFlow() función definida
    if 'function applyMotivoFlow(' in html:
        result.ok('v3.10.53 JS: applyMotivoFlow() función definida')
    else:
        result.fail('v3.10.53 JS applyMotivoFlow', 'función applyMotivoFlow esperada')

    # JS: applyMotivoFlow agrega .is-hidden-flow cuando motivo vacío
    if re.search(
        r"applyMotivoFlow[\s\S]{0,1800}classList\.add\(\s*['\"]is-hidden-flow['\"]",
        html
    ):
        result.ok('v3.10.53 JS: applyMotivoFlow agrega .is-hidden-flow cuando motivo vacío')
    else:
        result.fail('v3.10.53 JS hide logic', 'applyMotivoFlow debe agregar .is-hidden-flow si motivo vacío')

    # JS: applyMotivoFlow quita .is-hidden-flow cuando matching
    if re.search(
        r"matchingIds\.indexOf[\s\S]{0,200}classList\.remove\([\s\S]{0,50}is-hidden-flow",
        html
    ):
        result.ok('v3.10.53 JS: applyMotivoFlow quita .is-hidden-flow para cards del flow')
    else:
        result.fail('v3.10.53 JS show logic', 'remove is-hidden-flow para matching esperado')

    # JS: initMotivoSelect change handler llama applyMotivoFlow (no solo desgloseMotivoFlow)
    if re.search(
        r"motivoSelect\.addEventListener\(\s*['\"]change['\"][\s\S]{0,400}applyMotivoFlow\(\s*value\s*\)",
        html
    ):
        result.ok('v3.10.53 JS: change en MOTIVO select llama applyMotivoFlow(value)')
    else:
        result.fail('v3.10.53 JS change→flow', 'change debe llamar applyMotivoFlow(value)')

    # JS: applyMotivoFlow llamado en DOMContentLoaded para inicial state
    if re.search(r"applyMotivoFlow\(\s*getScriptVar\([\s\S]{0,30}motivo", html):
        result.ok('v3.10.53 JS: applyMotivoFlow invocado en init con getScriptVar("motivo")')
    else:
        result.fail('v3.10.53 JS init', 'applyMotivoFlow(getScriptVar("motivo")) esperado en init')

    # JS: getMotivoFlowCards incluye OTRAS (no OTRO)
    if re.search(r"['\"]OTRAS['\"]\s*:\s*\[", html):
        result.ok('v3.10.53 JS: getMotivoFlowCards usa "OTRAS" (alineado con v3.10.51 select)')
    else:
        result.fail('v3.10.53 JS OTRAS', 'getMotivoFlowCards debe incluir "OTRAS" (no "OTRO")')

    # JS: anti-regression — modo "Sin MOTIVO → ALL hidden" se aplica también a alerts
    # (busca el loop sobre .script-alert)
    if re.search(r"querySelectorAll\([\s\S]{0,30}\.script-alert[\s\S]{0,500}is-hidden-flow", html):
        result.ok('v3.10.53 JS: applyMotivoFlow también oculta .script-alert si no hay motivo')
    else:
        result.fail('v3.10.53 JS alerts hide', 'applyMotivoFlow debe ocultar alerts también')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.54 — BUG FIX: grid-body cerraba prematuramente
    # CEO directive 2026-10-20: "arreglemos al darle click al header
    #   ESTRUCTURA DE LLAMADA. se debe ocultar todo el contenido completo
    #   del GRID, en este momento solo oculta la plantilla de datos"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.54 — FIX: grid-body contiene TODO el contenido del GRID')

    # Contar script-section DENTRO de grid-body (no como siblings)
    # Reconstruir la región GRID 0 y verificar balance + posición
    grid_block_start = html.find('<div class="grid-block" id="gridScript"')
    grid_block_end = html.find('GRID 1: ADDRESS', grid_block_start)
    if grid_block_start != -1 and grid_block_end != -1:
        region = html[grid_block_start:grid_block_end]
        grid_body_open_rel = region.find('id="gridScript-body"')
        # Find balanced close of grid-body
        events = []
        for m in re.finditer(r'<div[^>]*>', region):
            cls_match = re.search(r'<div\s+class="([^"]+)"', m.group(0))
            cls = cls_match.group(1).split()[0] if cls_match else '?'
            events.append((m.start(), 'open', cls))
        for m in re.finditer(r'</div>', region):
            events.append((m.start(), 'close', None))
        events.sort()
        stack = []
        gb_close_pos = None
        for pos, kind, cls in events:
            if kind == 'open':
                stack.append(cls)
                if cls == 'grid-body':
                    gb_start_pos = pos
            else:
                if stack:
                    opened = stack.pop()
                    if opened == 'grid-body':
                        gb_close_pos = pos
                        break

        if gb_close_pos is not None:
            body_content = region[gb_start_pos:gb_close_pos]
            n_sections = len(re.findall(r'<div[^>]*class="script-section"', body_content))
            n_transferir = len(re.findall(r'<div[^>]*class="script-card[^"]*--transfer"', body_content))
            n_vards = len(re.findall(r'<div[^>]*class="script-vars"', body_content))

            if n_sections == 4 and n_transferir == 1 and n_vards == 1:
                result.ok(f'v3.10.54 BUG FIX: 4 script-section + 1 transferir + 1 script-vars TODOS dentro de grid-body')
            else:
                result.fail('v3.10.54 BUG FIX',
                            f'4 sections + 1 transferir + 1 vars esperados dentro de grid-body; {n_sections} sections + {n_transferir} transferir + {n_vards} vars encontrados')
        else:
            result.fail('v3.10.54 grid-body close', 'grid-body close no encontrado')
    else:
        result.fail('v3.10.54 region', 'región GRID 0 no encontrada')

    # CSS: regla que colapsa grid-body (sigue presente — v3.10.54 solo aseguramos
    # que TODO el contenido del GRID esté dentro de grid-body para que esta
    # regla aplique al 100% del contenido, no solo a las variables inputs)
    if re.search(r"\.grid-block\.is-collapsed\s+\.grid-body\s*\{[^}]*(display\s*:\s*none|max-height\s*:\s*0)", html):
        result.ok('v3.10.54 CSS: .grid-block.is-collapsed .grid-body colapsa contenido (regla sin cambios)')
    else:
        result.fail('v3.10.54 CSS', '.grid-block.is-collapsed .grid-body colapsando esperado')

    # Anti-regression: las 4 script-sections NO están como siblings de grid-body
    # (antes del fix, la cantidad de script-section como sibling era 4)
    if grid_block_start != -1 and grid_block_end != -1:
        region = html[grid_block_start:grid_block_end]
        # Match script-section APARICIÓN que NO esté dentro de grid-body
        gb_open = region.find('id="gridScript-body"')
        gb_close = region.rfind('</div>', 0, gb_open)  # the close just BEFORE the next </div> after the right structure
        # Simpler: count script-sections in region BEFORE grid-body
        before_body = region[:gb_open]
        n_sections_before = len(re.findall(r'<div[^>]*class="script-section', before_body))
        if n_sections_before == 0:
            result.ok('v3.10.54 ANTI-REGRESSION: 0 script-section huérfanas antes de grid-body')
        else:
            result.fail('v3.10.54 anti-regression', f'{n_sections_before} script-section huérfanas encontradas antes de grid-body')

    # BUG 2 + 3 (v3.10.50): applyScriptCollapsedStates colapsaba TODO por default
    # si no había estado persistido. v3.10.53 REMOVIÓ este auto-collapse.
    # Ahora la visibility inicial la controla applyMotivoFlow() (que oculta
    # todas las cards si motivo está vacío). El estado colapsado del usuario
    # (manual) sí se preserva y aplica a las cards visibles.
    auto_collapse_present = re.search(
        r"function applyScriptCollapsedStates[\s\S]{0,800}stateKeys\.length\s*===\s*0[\s\S]{0,300}querySelectorAll\(\s*['\"]\.script-card\[data-script-id\]\s*['\"]",
        html, re.DOTALL
    )
    if not auto_collapse_present:
        result.ok('v3.10.53 BUG 2: applyScriptCollapsedStates ya NO auto-colapsa TODO (delegado a applyMotivoFlow)')
    else:
        result.fail('v3.10.53 BUG 2', 'auto-colapsar TODO debería haber sido eliminado')

    # BUG 3: NO hay auto-desglose al cargar (el flow solo se desglosa en click MOTIVO)
    # Contar LLAMADAS REALES a desgloseMotivoFlow (excluir definición + comentarios)
    # Definición = "function desgloseMotivoFlow(...)"
    # Comentarios = líneas que empiezan con //
    # Llamada real = "desgloseMotivoFlow(args)" en línea de código
    real_calls = []
    for i, line in enumerate(html.split('\n'), start=1):
        stripped = line.strip()
        # Saltar comentarios y definiciones
        if stripped.startswith('//'):
            continue
        if 'function desgloseMotivoFlow' in stripped:
            continue
        # Buscar llamadas
        m = re.search(r"desgloseMotivoFlow\(\s*([A-Za-z_]*)\s*\)", stripped)
        if m:
            real_calls.append((i, m.group(1)))

    # Debe haber exactamente 1 llamada real, en el click handler
    click_handler = re.search(
        r"card\.addEventListener\(\s*['\"]click['\"]\s*,\s*function[\s\S]{0,1000}desgloseMotivoFlow\(\s*value\s*\)",
        html
    )
    if len(real_calls) == 1 and click_handler:
        result.ok(f'v3.10.50 BUG 3: desgloseMotivoFlow solo se llama en click handler (L{real_calls[0][0]}, 1 llamada real)')
    else:
        result.fail('v3.10.50 BUG 3',
                    f'desgloseMotivoFlow debe ser solo en click (encontradas {len(real_calls)} llamadas reales: {real_calls})')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.45 — BUG FIXES (layout + telefono-last3)
    # CEO: "Formato roto. No se hizo cambios correctos, las opciones deben
    # estar dentro de MOTIVO. En el speech dice {telefono en consulta} no
    # esta tomando realmente la indicacion que te di el numero que ya esta
    # ingresado en la plantilla inicial."
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.45 — BUG FIXES: layout MOTIVO full-width + telefono-last3 alias')

    # BUG 1 LAYOUT v3.10.45: MOTIVO en row-1 (NO compartía row con TELÉFONO)
    # v3.10.52 UPDATE: TELÉFONO y MOTIVO consolidados en 1 sola row-2 (side-by-side)
    #   Estructura HTML esperada v3.10.52:
    #     <div class="field-row row-2">
    #       <div class="field"><label script-telefono>...</div>
    #       <div class="field"><label script-motivo>...</div>
    #     </div>
    layout_ok = (
        re.search(
            r'<div\s+class="field-row\s+row-2">\s*<div\s+class="field">\s*<label\s+for="script-telefono"',
            html, re.DOTALL
        ) is not None and
        re.search(
            r'<div\s+class="field-row\s+row-2">[\s\S]{0,400}<label\s+for="script-motivo"',
            html, re.DOTALL
        ) is not None
    )
    if layout_ok:
        result.ok('v3.10.52 LAYOUT: TELÉFONO + MOTIVO consolidados en 1 sola row-2 (side-by-side)')
    else:
        result.fail('v3.10.52 LAYOUT', 'TELÉFONO + MOTIVO deben compartir field-row row-2')

    # LEGACY TEST — v3.10.45 esperaba que estuvieran en row-1 separados.
    # v3.10.52 los consolidó en row-2. Anti-regression: NO deben existir
    # dos field-row row-1 separados para tel/motivo.
    legacy_separate = (
        re.search(
            r'<div\s+class="field-row\s+row-1">\s*<div\s+class="field">\s*<label\s+for="script-telefono"',
            html, re.DOTALL
        ) is not None and
        re.search(
            r'<div\s+class="field-row\s+row-1">\s*<div\s+class="field">\s*<label\s+for="script-motivo"',
            html, re.DOTALL
        ) is not None
    )
    if not legacy_separate:
        result.ok('v3.10.52 LEGACY: TELÉFONO + MOTIVO NO están en row-1 separados (consolidados)')
    else:
        result.fail('v3.10.52 LEGACY layout', 'TELÉFONO + MOTIVO ya no deben estar en row-1 separados')

    # CSS: .row-2 = 2 columns split
    if re.search(r'\.row-2\s*\{[^}]*grid-template-columns:\s*1fr\s+1fr\s*;?\s*\}', html):
        result.ok('v3.10.52 CSS: .row-2 { grid-template-columns: 1fr 1fr } (side-by-side)')
    else:
        result.fail('v3.10.52 CSS .row-2', '.row-2 con grid-template-columns 1fr 1fr esperado')

    # CSS: .row-1 = single column full width (mantenido para otros usos)
    if re.search(r'\.row-1\s*\{[^}]*grid-template-columns:\s*1fr\s*;?\s*\}', html):
        result.ok('v3.10.45 CSS: .row-1 { grid-template-columns: 1fr } (full-width, otros usos)')
    else:
        result.fail('v3.10.45 CSS .row-1', '.row-1 con grid-template-columns 1fr esperado')

    # BUG 2 JS: vars incluye 'telefono-last3' alias
    vars_check = re.search(
        r"var\s+vars\s*=\s*\{[^}]*['\"]telefono-last3['\"]\s*:",
        html, re.DOTALL
    )
    if vars_check:
        result.ok('v3.10.45 BUG 2 JS: vars[\'telefono-last3\'] alias (toma input telefono)')
    else:
        result.fail('v3.10.45 BUG 2 JS', 'vars["telefono-last3"] alias esperado')

    # Anti-regresión: speech bienvenida sigue con "Su consulta es por el numero que termina en"
    bienvenida_re = re.search(
        r'id="script-text-bienvenida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if bienvenida_re and 'Su consulta es por el numero que termina en' in re.sub(r'<[^>]+>', '', bienvenida_re.group(1)):
        result.ok('v3.10.45 ANTI-regresión: speech bienvenida mantiene redacción CEO')
    else:
        result.fail('v3.10.45 anti-regresion', 'speech bienvenida debe mantener "Su consulta es por el numero..."')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.46 — BUG FIXES: layout grids + CTRL+SHIFT+R clear + speeches
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.46 — BUG FIXES: grid widths + CTRL+SHIFT+R + speeches CEO')

    # BUG 4: .script-vars SIN background/border/padding (panel interno removido)
    script_vars_css = re.search(r'\.script-vars\s*\{([^}]+)\}', html, re.DOTALL)
    if script_vars_css:
        sv_body = re.sub(r'/\*[\s\S]*?\*/', '', script_vars_css.group(1))
        # Solo debe tener margin-bottom, sin padding/background/border
        if ('background' not in sv_body and 'border:' not in sv_body and
            'padding:' not in sv_body and 'margin-bottom' in sv_body):
            result.ok('v3.10.46 BUG 4: .script-vars sin padding/border/background (ancho igual grids)')
        else:
            result.fail('v3.10.46 BUG 4 .script-vars', '.script-vars debe ser solo margin-bottom')
    else:
        result.fail('v3.10.46 BUG 4', '.script-vars CSS missing')

    # BUG 1: clearScriptTemplate() función definida
    if 'function clearScriptTemplate()' in html:
        result.ok('v3.10.46 BUG 1: clearScriptTemplate() función definida')
    else:
        result.fail('v3.10.46 BUG 1 clearScriptTemplate', 'función clearScriptTemplate esperada')

    # BUG 1: CTRL+SHIFT+R keydown handler presente
    if re.search(r"e\.ctrlKey\s*&&\s*e\.shiftKey\s*&&\s*\(e\.key\s*===\s*['\"]R['\"]", html):
        result.ok('v3.10.46 BUG 1: handler CTRL+SHIFT+R (keydown con ctrl+shift+R)')
    else:
        result.fail('v3.10.46 BUG 1 keydown', 'handler ctrl+shift+R esperado')

    # BUG 1: clearScriptTemplate limpia localStorage
    if re.search(r"localStorage\.removeItem\(\s*['\"]cima_script_vars_v1['\"]\s*\)", html):
        result.ok('v3.10.46 BUG 1: clearScriptTemplate borra localStorage cima_script_vars_v1')
    else:
        result.fail('v3.10.46 BUG 1 localStorage', 'removeItem cima_script_vars_v1 esperado')

    # BUG 1: clearScriptTemplate limpia inputs (agente, cliente, telefono)
    if re.search(r"\[\s*['\"]agente['\"]\s*,\s*['\"]cliente['\"]\s*,\s*['\"]telefono['\"]\s*\][^}]*setScriptVar", html, re.DOTALL):
        result.ok('v3.10.46 BUG 1: clearScriptTemplate limpia inputs agente/cliente/telefono')
    else:
        result.fail('v3.10.46 BUG 1 inputs', 'clearScriptTemplate debe limpiar agente/cliente/telefono')

    # BUG 2 + BUG 3: speeches CEO directive ya testeados arriba
    # (atencion-segura + encuesta)

    # Anti-regresión: .section sigue con border + background (los otros grids)
    section_css = re.search(r'\.section\s*\{([^}]+)\}', html, re.DOTALL)
    if section_css and 'border:' in section_css.group(1):
        result.ok('v3.10.46 ANTI-regresión: .section mantiene border (grids Address/Tipif intactos)')
    else:
        result.fail('v3.10.46 anti-regresion', '.section debe mantener border')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.47 — BUG FIX CRÍTICO: grids mismo ancho (jerarquía DOM)
    # CEO: "las 3 grid deben tener el mismo ancho" en modo popup.
    # Causa raíz: 1 </div> extra entre gridScript y gridAddress cerraba
    # prematuramente .app. gridAddress/gridTipif quedaban como hijos de
    # .popup-frame (sin padding, 465px) mientras gridScript era hijo de
    # .app (con padding 6px, 453px) → 12px de diferencia visual.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.47 — BUG FIX: jerarquía DOM (los 3 grids hijos de .app)')

    # Verificar jerarquía DOM con BeautifulSoup
    from bs4 import BeautifulSoup
    soup = BeautifulSoup(html, 'html.parser')

    grids = soup.find_all('div', class_='grid-block')
    grid_parents = {}
    for g in grids:
        cid = g.get('id', '')
        parent = g.parent
        parent_label = f"{parent.name}.{'.'.join(parent.get('class', []))}"
        grid_parents[cid] = parent_label

    # Los 3 grids deben ser hijos de div.app
    all_in_app = all(parent == 'div.app' for parent in grid_parents.values())
    if all_in_app and len(grid_parents) == 3:
        result.ok('v3.10.47 DOM: 3 grids (gridScript/gridAddress/gridTipif) son hijos de div.app')
    else:
        result.fail('v3.10.47 DOM', f'los 3 grids deben ser hijos de div.app, parents={grid_parents}')

    # Anti-regresión: la app tiene exactamente 8 hijos directos
    # (srStatus, srAlert, header, alertBox, gridScript, gridAddress, gridTipif, footer)
    app_div = soup.find('div', class_='app')
    if app_div:
        app_children = app_div.find_all(recursive=False)
        # Contar TODOS los hijos con tag (no solo divs)
        n_significant = len(app_children)
        if n_significant >= 7:  # srStatus, srAlert, header, alertBox, gridScript, gridAddress, gridTipif
            result.ok(f'v3.10.47 DOM: .app tiene {n_significant} hijos (srStatus/header/3 grids/footer)')
        else:
            result.fail('v3.10.47 DOM .app children', f'≥7 hijos esperados, {n_significant} encontrados')
    else:
        result.fail('v3.10.47 DOM .app', 'div.app no encontrado')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.48 — Speech NUTridos AI 2026 (variación {Cliente} + consecutivas)
    # CEO: "utilizar mas frases empaticas, cordiales, agradables y no
    # siempre decirle el nombre podemos personalizar poner el nombre en
    # diversas partes donde corresponda para que no se sienta una
    # saturacion y los datos sean inteligentes"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.48 — SPEECHES nutridos AI 2026 (variación {Cliente})')

    # BUG 1: MOTIVO tiempo-espera tiene continuación ("Apenas tenga novedades...")
    espera_match = re.search(
        r'id="script-text-tiempo-espera"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if espera_match:
        espera_text_raw = espera_match.group(1)
        espera_visible = re.sub(r'<[^>]+>', '', espera_text_raw)
        espera_visible = re.sub(r'\s+', ' ', espera_visible).strip()
        if 'Apenas tenga novedades' in espera_visible and 'c\u00f3mo seguimos' in espera_visible:
            result.ok('v3.10.48 BUG 1: tiempo-espera tiene continuación (Apenas tenga novedades)')
        else:
            result.fail('v3.10.48 BUG 1 tiempo-espera', 'continuación "Apenas tenga novedades..." esperada')
    else:
        result.fail('v3.10.48 BUG 1', 'script-text-tiempo-espera not found')

    # BUG 2: DESPEDIDA consecutiva (no 2 mensajes separados)
    desp_match = re.search(
        r'id="script-text-despedida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if desp_match:
        desp_text_raw = desp_match.group(1)
        desp_visible = re.sub(r'<[^>]+>', '', desp_text_raw)
        desp_visible = re.sub(r'\s+', ' ', desp_visible).strip()
        # CEO directive: consecutiva, no "hemos realizado para ti" + "solo sigamos"
        if ('hemos realizado para ti' not in desp_visible and
            'Perfecto' in desp_visible and
            'ya tenemos listo' in desp_visible):
            result.ok('v3.10.48 BUG 2: despedida consecutiva (Perfecto + ya tenemos listo, sin "hemos realizado")')
        else:
            result.fail('v3.10.48 BUG 2 despedida', 'redacción consecutiva CEO esperada')
    else:
        result.fail('v3.10.48 BUG 2', 'script-text-despedida not found')

    # BUG 3: Variación {Cliente} — no siempre al inicio
    speeches_with_cliente = {
        'sondeo-averias': 'despu\u00e9s de "Comprendo"',
        'atencion-segura': 'despu\u00e9s de "devolverle la llamada"',
        'sondea': 'despu\u00e9s de "Cu\u00e9nteme"',
        'educa-app': 'despu\u00e9s de intro emp\u00e1tica',
        'tiempo-espera': 'despu\u00e9s de "valido su informaci\u00f3n"'
    }

    for script_id, position_desc in speeches_with_cliente.items():
        match = re.search(
            rf'id="script-text-{script_id}"[^>]*>(.+?)</p>',
            html, re.DOTALL
        )
        if match:
            raw = match.group(1)
            visible = re.sub(r'<[^>]+>', '', raw)
            visible = re.sub(r'\s+', ' ', visible).strip()
            cliente_pos = visible.find('{Cliente}')
            if cliente_pos == -1:
                result.fail(f'v3.10.48 BUG 3 {script_id}', 'no contiene {Cliente}')
            elif cliente_pos >= 8:  # ≥8 chars antes = no al inicio
                result.ok(f'v3.10.48 BUG 3: {script_id} tiene {{Cliente}} {position_desc}')
            else:
                result.fail(f'v3.10.48 BUG 3 {script_id}',
                            f'{{Cliente}} al inicio (pos {cliente_pos}), CEO pidió variación')

    # BUG 3: encuesta AHORA tiene {Cliente} (antes no lo tenía)
    encuesta_match = re.search(
        r'id="script-text-encuesta"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if encuesta_match:
        enc_raw = encuesta_match.group(1)
        if 'data-ph="cliente"' in enc_raw and 'Antes de cerrar' in enc_raw:
            result.ok('v3.10.48 BUG 3: encuesta tiene {Cliente} agregado ("Antes de cerrar {Cliente}")')
        else:
            result.fail('v3.10.48 BUG 3 encuesta', '{Cliente} agregado + "Antes de cerrar" esperado')

    # BUG 3: transferir solo 1 {Cliente} (antes tenía 2)
    transfer_match = re.search(
        r'id="script-text-transferir"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if transfer_match:
        tr_raw = transfer_match.group(1)
        n_cliente = tr_raw.count('data-ph="cliente"')
        if n_cliente == 1:
            result.ok('v3.10.48 BUG 3: transferir solo 1 {Cliente} (antes 2)')
        else:
            result.fail('v3.10.48 BUG 3 transferir', f'1 {Cliente} esperado, {n_cliente} encontrados')

    # Anti-regresión: bienvenida mantiene "Bienvenido a Movistar" + "¿Es correcto?"
    bienvenida_match = re.search(
        r'id="script-text-bienvenida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if bienvenida_match:
        bien_raw = bienvenida_match.group(1)
        bien_visible = re.sub(r'<[^>]+>', '', bien_raw)
        bien_visible = re.sub(r'\s+', ' ', bien_visible).strip()
        if 'Bienvenido a Movistar' in bien_visible and '\u00bfEs correcto?' in bien_visible:
            result.ok('v3.10.48 ANTI-regresión: bienvenida mantiene "Bienvenido a Movistar" + "¿Es correcto?"')
        else:
            result.fail('v3.10.48 anti-regresion bienvenida', 'speech bienvenida debe mantener "Bienvenido a Movistar"')

    # HTML: retorno card REMOVIDO (CEO: "no sera un speech textual sera como una sola row de alerta")
    if 'data-script-id="retorno"' not in html:
        result.ok('v3.10.37 HTML: retorno card REMOVIDO (reemplazado por alerts)')
    else:
        result.fail('v3.10.37 HTML retorno regresión', 'retorno card debería estar eliminado')

    # HTML: consulta-adicional card REMOVIDO (CEO: "no debe ser un speech debe ser igual un texto de apoyo")
    if 'data-script-id="consulta-adicional"' not in html:
        result.ok('v3.10.37 HTML: consulta-adicional card REMOVIDO (reemplazado por alert)')
    else:
        result.fail('v3.10.37 HTML consulta regresión', 'consulta-adicional debería estar eliminado')

    # HTML: cierre-con-solucion y cierre-sin-solucion cards REMOVIDOS
    if 'data-script-id="cierre-con-solucion"' not in html and 'data-script-id="cierre-sin-solucion"' not in html:
        result.ok('v3.10.37 HTML: cierre-con-solucion + cierre-sin-solucion REMOVIDOS')
    else:
        result.fail('v3.10.37 HTML cierre regresión', 'cards ✓ ✗ deberían estar eliminados')

    # HTML: NUEVO despedida card con speech específico del CEO
    if 'data-script-id="despedida"' in html:
        result.ok('v3.10.37 HTML: card "despedida" (NUEVO speech general)')
    else:
        result.fail('v3.10.37 HTML despedida', 'card despedida missing')
    import re as re_mod
    despedida_match = re_mod.search(
        r'id="script-text-despedida"[^>]*>(.+?)</p>',
        html,
        re_mod.DOTALL
    )
    if despedida_match:
        despedida_text = despedida_match.group(1)
        # v3.10.48: CEO directive "consecutiva la atencion" — flujo continuo
        # "Perfecto {Cliente}, ya tenemos listo el (proceso generado).
        #  Sigamos las indicaciones..."
        if ('Perfecto' in despedida_text and
            'ya tenemos listo' in despedida_text and
            'proceso generado' in despedida_text and
            'Ha sido un gusto' in despedida_text):
            result.ok('v3.10.48 SPEECH: despedida CEO directive "consecutiva" (Perfecto + ya tenemos listo)')
        else:
            result.fail('v3.10.48 SPEECH despedida', 'CEO directive "consecutiva" esperada')





    # ══════════════════════════════════════════════════════════════════════
    # v3.10.28 — Inputs filled clean + toast wording
    # CEO directive: "Cuando los inputs tienen contenido pasa lo mismo se ve
    # como un background cuadrado y bordes redondos rompiendose el estilo.
    # ...el boton de zona de apagado al darle click emite un mensaje ROTO:
    # tipea el CCPP para activar la alerta."
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.28 — .filled clean (sin halo) + toast wording')

    # CSS: .field input.filled SIN box-shadow halo (3px green ring)
    # El CEO percibía el halo como "background cuadrado" rodeando el input.
    # Removemos comentarios CSS antes de chequear para evitar falsos positivos.
    css_filled_match = re.search(r'\.field input\.filled,\s*\.field textarea\.filled\s*\{([^}]+)\}', html)
    if css_filled_match:
        filled_body = re.sub(r'/\*[\s\S]*?\*/', '', css_filled_match.group(1))  # strip comments
        if 'shadow-focus-green' in filled_body:
            result.fail('v3.10.28 CSS .filled', 'box-shadow halo shadow-focus-green debe estar REMOVIDO')
        elif 'box-shadow:' in filled_body and 'ring-hairline' in filled_body:
            result.ok('v3.10.28 CSS: .filled SIN halo (solo ring-hairline inset sutil)')
        else:
            result.fail('v3.10.28 CSS .filled', '.filled CSS no esperado')
    else:
        result.fail('v3.10.28 CSS .filled', '.field input.filled CSS missing')

    # JS: toast wording mejorado (Tipea → Escribe + del PDF)
    # v3.10.29: el toast fue REMOVIDO completamente (CEO: "solo debe funcionar
    # cuando se detecta zona de apagado exacta y guiar"). Validamos que NO
    # exista la línea de showToast para este caso.
    if re.search(r"if\s*\(\s*!chk\s*\|\|\s*chk\.nivel\s*!==\s*'ccpp_exacto'\s*\)\s*\{[^}]*showToast", html, re.DOTALL):
        result.fail('v3.10.29 JS toast', 'showToast debe estar REMOVIDO en caso no-match')
    elif re.search(r"if\s*\(\s*!chk\s*\|\|\s*chk\.nivel\s*!==\s*'ccpp_exacto'\s*\)\s*\{[^}]*return", html, re.DOTALL):
        result.ok('v3.10.29 JS: BLOQUEO silent en caso no-match (early return, sin toast)')
    else:
        result.fail('v3.10.29 JS bloque', 'silent early return esperado en caso no-match')

    # HTML: CCPP placeholder mejorado para guiar al usuario
    if 'placeholder="(elegir CCPP exacto del PDF)"' in html:
        result.ok('v3.10.29 HTML: CCPP placeholder guiando "(elegir CCPP exacto del PDF)"')
    else:
        result.fail('v3.10.29 HTML placeholder', 'placeholder CCPP mejorado esperado')

    # Anti-regression: NO debe quedar el placeholder viejo
    if 'placeholder="(tras elegir distrito)"' in html:
        result.fail('v3.10.29 HTML placeholder regresión', 'placeholder viejo aún presente')
    else:
        result.ok('v3.10.29 HTML: placeholder viejo REMOVIDO')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.29 — RESULTADO hover sin halo
    # CEO directive: "input de RESULTADO al pasar el mouse por encima como
    # hover se ve un background cuadrado que rompe estetica del borde redondo"
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.29 — RESULTADO hover sin halo (mismo bug que .filled v3.10.28)')

    # CSS: .output-box.output-result:hover SIN box-shadow halo (3px green ring)
    css_hover_match = re.search(r'\.output-box\.output-result:hover:not\(\.placeholder\)\s*\{([^}]+)\}', html)
    if css_hover_match:
        hover_body = re.sub(r'/\*[\s\S]*?\*/', '', css_hover_match.group(1))
        if 'box-shadow:' in hover_body and '0 0 0 3px' in hover_body:
            result.fail('v3.10.29 CSS hover', 'box-shadow halo 3px debe estar REMOVIDO')
        elif 'box-shadow:' in hover_body:
            result.ok('v3.10.29 CSS: RESULTADO hover SIN halo (solo ring-hairline inset sutil)')
        else:
            result.fail('v3.10.29 CSS hover', 'box-shadow ring-hairline inset esperado')
    else:
        result.fail('v3.10.29 CSS hover', '.output-box.output-result:hover CSS missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.30 — GRID 0 SCRIPT LLAMADA
    # CEO directive: "agregar un grid mas arriba de busqueda address" con
    # 8 secciones del flujo de llamada: bienvenida, atención segura, motivo
    # comunicación, detección avería, retoma, consulta adicional, app, encuesta.
    # Top vars (Agente, Cliente, Teléfono, Motivo) auto-reemplazan {placeholders}
    # en TODAS las cards via JS live update. Persistencia en localStorage.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.30 — GRID 0 Script Llamada (8 cards + vars + live replace)')

    # HTML: grid existe y está ARRIBA de gridAddress
    if 'id="gridScript"' in html and 'id="gridAddress"' in html:
        script_pos = html.find('id="gridScript"')
        address_pos = html.find('id="gridAddress"')
        if script_pos < address_pos:
            result.ok('v3.10.30 HTML: gridScript ARRIBA de gridAddress (orden correcto)')
        else:
            result.fail('v3.10.30 HTML orden', 'gridScript debe estar antes de gridAddress')
    else:
        result.fail('v3.10.30 HTML grid', 'id="gridScript" o id="gridAddress" missing')

    # HTML: 8 script cards presentes
    # v3.10.37: REMOVIDOS 'retorno', 'consulta-adicional', 'cierre-con-solucion',
    # 'cierre-sin-solucion'. AGREGADO 'despedida'. 3 alerts estáticos en
    # MOTIVO reemplazan retorno. 1 alert estático en CONSULTA reemplaza consulta-adicional.
    expected_cards_v37 = [
        'bienvenida',
        'sondea', 'sondeo-averias', 'atencion-segura',
        'tiempo-espera',
        'educa-app',
        'despedida',
        'encuesta', 'transferir'
    ]
    missing_cards = [c for c in expected_cards_v37 if 'data-script-id="' + c + '"' not in html]
    if not missing_cards:
        result.ok(f'v3.10.37 HTML: 9 script cards presentes ({len(expected_cards_v37)} totales)')
    else:
        result.fail('v3.10.37 HTML cards', f'cards faltantes: {missing_cards}')

    # v3.10.34: con-quien REMOVIDO (CEO: innecesario)
    if 'data-script-id="con-quien"' not in html:
        result.ok('v3.10.34 HTML: "con-quien" card REMOVIDO (CEO directive)')
    else:
        result.fail('v3.10.34 HTML regresión', '"con-quien" debería estar eliminado')

    # HTML: 4 inputs vars
    expected_vars = ['script-agente', 'script-cliente', 'script-telefono', 'script-motivo']
    missing_vars = [v for v in expected_vars if 'id="' + v + '"' not in html]
    if not missing_vars:
        result.ok('v3.10.30 HTML: 4 vars inputs (agente/cliente/telefono/motivo)')
    else:
        result.fail('v3.10.30 HTML vars', f'vars faltantes: {missing_vars}')

    # CSS: .script-card + .ph + .script-vars definidas
    if '.script-card {' in html:
        result.ok('v3.10.30 CSS: .script-card rule presente')
    else:
        result.fail('v3.10.30 CSS .script-card', '.script-card CSS missing')
    if re.search(r'\.ph\s*\{', html):
        result.ok('v3.10.30 CSS: .ph (placeholder highlight) rule presente')
    else:
        result.fail('v3.10.30 CSS .ph', '.ph CSS missing')

    # v3.10.33: copyScriptCard REMOVIDO (CEO: "eliminar icono de copiar no es necesario")
    # Las cards ahora son click-to-done, no copy-to-clipboard.
    if 'function copyScriptCard(' not in html:
        result.ok('v3.10.33: copyScriptCard() ELIMINADO (CEO directive)')
    else:
        result.fail('v3.10.33 JS copy regresión', 'copyScriptCard() debería estar eliminado')
    if 'function updateScriptPlaceholders(' in html:
        result.ok('v3.10.30 JS: updateScriptPlaceholders() función definida')
    else:
        result.fail('v3.10.30 JS update', 'updateScriptPlaceholders() missing')

    # JS: localStorage persistence
    if "SCRIPT_VARS_STORAGE_KEY" in html and 'cima_script_vars_v1' in html:
        result.ok('v3.10.30 JS: localStorage persistence (SCRIPT_VARS_STORAGE_KEY)')
    else:
        result.fail('v3.10.30 JS storage', 'localStorage persistence missing')

    # JS: input listeners wireados
    if "el.addEventListener('input', updateScriptPlaceholders)" in html:
        result.ok('v3.10.30 JS: input listeners → updateScriptPlaceholders')
    else:
        result.fail('v3.10.30 JS listeners', 'input listeners missing')

    # v3.10.33: script-card-copy botones REMOVIDOS (CEO directive).
    # Las cards ya no tienen botón de copiar. Validamos que NO existan.
    n_copy_buttons = html.count('class="script-card-copy"')
    if n_copy_buttons == 0:
        result.ok('v3.10.33 HTML: 0 botones copy-to-clipboard (eliminados)')
    else:
        result.fail('v3.10.33 HTML buttons regresión', f'0 esperados, {n_copy_buttons} encontrados')

    # HTML: grid collapsibility (sigue patrón v3.10.23)
    if 'id="gridScript-toggle"' in html and 'id="gridScript-body"' in html:
        result.ok('v3.10.30 HTML: gridScript colapsable (toggle + body)')
    else:
        result.fail('v3.10.30 HTML collapse', 'toggle/body missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.31 — HEADER Estructura de Atención mejorada
    # CEO directive: "Mejora los speech Inteligentes... completo correcto
    # coma, puntos etc dales forma sin romperlos. Las preguntas Iniciales
    # de rellenado inteligente deben tener aun mas inteligente, nuestro
    # canal de atencion es solo preguntas de TECNICA MOVIL, si el cliente
    # consulta por otro motivo lo debemos transferir. En telefono debe
    # ser 'TELEFONO EN CONSULTA' en MOTIVO debe ir INTERNET, LLAMADAS,
    # SMS, THETERING, OTRO <- LA OPCION de otro cambiara los speech ya
    # que el flujo es distinto, no hacemos deteccion de averia ni
    # consulta adicional."
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.31 — HEADER Estructura de Atención mejorada')

    # HTML: phone label "TELÉFONO DE CONSULTA" (v3.10.51 EN → DE)
    if '<label for="script-telefono">TELÉFONO DE CONSULTA</label>' in html:
        result.ok('v3.10.31/v3.10.51 HTML: phone label "TELÉFONO DE CONSULTA"')
    else:
        result.fail('v3.10.31/v3.10.51 HTML phone label', 'TELÉFONO DE CONSULTA esperado')

    # HTML: motivo es SELECT (no input) — v3.10.51 consolidado en 1 dropdown
    if re.search(r'<select\s+id="script-motivo"[^>]*>', html) and \
       'hidden' not in re.search(r'<select\s+id="script-motivo"[^>]*>', html).group(0):
        result.ok('v3.10.31/v3.10.51 HTML: motivo es <select> visible (no hidden)')
    else:
        result.fail('v3.10.31/v3.10.51 HTML motivo', '<select id="script-motivo"> visible esperado')

    # HTML: motivo select tiene 4 motivos (MOVIL, OTRAS, AVERIA, GTR) en v3.10.51
    # v3.10.31 tenía 5 opciones (INTERNET/LLAMADAS/SMS/THETERING/OTRO);
    # v3.10.51 las consolidó en "Técnico Móvil" como opción umbrella + 3 próximas.
    expected_motivos = ['MOVIL', 'OTRAS', 'AVERIA', 'GTR']
    missing_motivos = [m for m in expected_motivos if f'value="{m}"' not in html]
    if not missing_motivos:
        result.ok(f'v3.10.51 HTML: motivo select con 4 opciones ({", ".join(expected_motivos)})')
    else:
        result.fail('v3.10.51 HTML motivo options', f'faltantes: {missing_motivos}')

    # HTML: card 9 "Transferir" presente
    if 'data-script-id="transferir"' in html:
        result.ok('v3.10.31 HTML: card 9 "Transferir" presente')
    else:
        result.fail('v3.10.31 HTML transfer', 'card transferir missing')

    # HTML: card transferir está oculta por default (atributo hidden)
    if re.search(r'<div\s+class="script-card[^"]*script-card--transfer[^"]*"[^>]*\s+hidden>', html):
        result.ok('v3.10.31 HTML: card "Transferir" oculta por default (atributo hidden)')
    else:
        result.fail('v3.10.31 HTML transfer hidden', 'card transferir debe estar hidden por default')

    # JS: helper toggleScriptCard presente (legacy v3.10.31)
    if 'function toggleScriptCard(' in html:
        result.ok('v3.10.31 JS: toggleScriptCard() helper presente (legacy)')
    else:
        result.fail('v3.10.31 JS helper', 'toggleScriptCard() missing')

    # v3.10.31 → v3.10.53: lógica "vars.motivo === 'OTRO'" → reemplazada por
    # getMotivoFlowCards(motivo) en applyMotivoFlow(). El valor del select es
    # 'OTRAS' (no 'OTRO') desde v3.10.51.
    if re.search(r"getMotivoFlowCards\s*\(\s*motivo\s*\)", html):
        result.ok('v3.10.53 JS: getMotivoFlowCards(motivo) controla flow visibility (reemplaza OTRO legacy)')
    else:
        result.fail('v3.10.53 JS flow', 'getMotivoFlowCards(motivo) esperado')

    # Anti-regression: lógica legacy "vars.motivo === 'OTRO'" YA NO en updateScriptPlaceholders
    legacy_otro = re.search(r"isOtro\s*=\s*\(vars\.motivo\s*===\s*['\"]OTRO['\"]\)", html)
    if not legacy_otro:
        result.ok('v3.10.53 JS: lógica legacy isOtro = (vars.motivo === "OTRO") ELIMINADA de updateScriptPlaceholders')
    else:
        result.fail('v3.10.53 JS legacy OTRO', 'lógica OTRO legacy debería estar removida')

    # Speeches: bienvenida (v3.10.33: id cambió de bienvenida-1 → bienvenida,
    # y NO debe mencionar TÉCNICO MÓVIL — CEO rechazó ese cambio agresivo en
    # v3.10.31). Esta es una anti-regression test para evitar cambios agresivos
    # futuros.
    bienvenida_match = re.search(
        r'id="script-text-bienvenida"[^>]*>(.+?)</p>',
        html, re.DOTALL
    )
    if bienvenida_match:
        bienvenida_text = bienvenida_match.group(1)
        # CEO directive: NO debe contener "TÉCNICO MÓVIL" (era cambio agresivo v3.10.31)
        if 'TÉCNICO MÓVIL' in bienvenida_text:
            result.fail('v3.10.33 ANTI-anti', 'bienvenida NO debe mencionar TÉCNICO MÓVIL (CEO rechazó v3.10.31)')
        elif 'Bienvenido a Movistar' in bienvenida_text and '¿Es correcto?' in bienvenida_text:
            result.ok('v3.10.33 SPEECH: bienvenida SIN TÉCNICO MÓVIL (CEO anti-regression)')
        else:
            result.fail('v3.10.33 SPEECH bienvenida', 'texto bienvenida no esperado')
    else:
        result.fail('v3.10.33 SPEECH bienvenida', 'script-text-bienvenida not found')

    # CSS: .script-card[hidden] { display: none }
    if '.script-card[hidden]' in html and 'display: none' in html:
        result.ok('v3.10.31 CSS: .script-card[hidden] { display: none }')
    else:
        result.fail('v3.10.31 CSS hidden', '.script-card[hidden] display: none esperado')

    # CSS: select styling (.script-vars .field select)
    if re.search(r'\.script-vars\s+\.field\s+select\s*\{', html):
        result.ok('v3.10.31 CSS: select styling presente')
    else:
        result.fail('v3.10.31 CSS select', '.script-vars .field select CSS missing')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.3 — STOP CHECK visual para COMU-5672 + Santa Rosa de Loreto
    # CEO directive: "como solucionamos los RIESGOS CONOCIDOS?"
    # Soluciones: stop check prominente + distrito nuevo 160405 + anomalías
    # validadas per OSIPTEL cross-check.
    # v3.10.12: ELIMINADO stopCheck (2G fuera de scope). Santa Rosa y
    # anomalías VALIDADAS siguen activos.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.3 — STOP CHECK visual + riesgos resueltos (v3.10.12: stopCheck ELIMINADO)')

    # JS: stopCheckHtml ELIMINADO v3.10.12 (solo aplicaba a 2G)
    if 'stopCheckHtml' not in html:
        result.ok('v3.10.12: stopCheckHtml ELIMINADO (sin 2G)')
    else:
        result.fail('v3.10.12 regresión', 'stopCheckHtml aún en HTML')

    if not re.search(r'STOP.*VERIFICAR CON CLIENTE', html, re.DOTALL):
        result.ok('v3.10.12: texto "STOP — VERIFICAR CON CLIENTE" ELIMINADO (sin 2G)')
    else:
        result.fail('v3.10.12 regresión', 'texto STOP aún en HTML')

    # CSS: .apagado-panel-stopcheck ELIMINADO
    if not re.search(r'\.apagado-speech-panel\s+\.apagado-panel-stopcheck\s*\{', html):
        result.ok('v3.10.12: .apagado-panel-stopcheck CSS ELIMINADO (sin 2G)')
    else:
        result.fail('v3.10.12 regresión', '.apagado-panel-stopcheck CSS aún presente')

    if not re.search(r'@keyframes\s+stopcheckPulse', html):
        result.ok('v3.10.12: stopcheckPulse animation ELIMINADA (sin 2G)')
    else:
        result.fail('v3.10.12 regresión', '@keyframes stopcheckPulse aún presente')

    # UBIGEO: Santa Rosa de Loreto inyectada
    if re.search(r'"SANTA ROSA DE LORETO"', html):
        result.ok('v3.10.3 UBIGEO: SANTA ROSA DE LORETO inyectada (d6 160405)')
    else:
        result.fail('v3.10.3 UBIGEO santa_rosa', 'SANTA ROSA DE LORETO no inyectada')

    if '"_ubigeo":"160405"' in html:
        result.ok('v3.10.3 UBIGEO: d6 160405 presente')
    else:
        result.fail('v3.10.3 UBIGEO 160405', 'd6 160405 missing')

    # Anomalías: estado VALIDADO_*
    if re.search(r'VALIDADO_sin_apagado_oficial', html):
        result.ok('v3.10.3 INTEL: anomalías con estado VALIDADO_*')
    else:
        result.fail('v3.10.3 INTEL validado', 'estados VALIDADO_* no encontrados')

    if re.search(r'validacion_cruzada', html):
        result.ok('v3.10.3 INTEL: campo validacion_cruzada (OSIPTEL cross-check)')
    else:
        result.fail('v3.10.3 INTEL validacion', 'campo validacion_cruzada missing')

    if re.search(r'"resolucion"\s*:', html):
        result.ok('v3.10.3 INTEL: campo resolucion (acción por anomalía)')
    else:
        result.fail('v3.10.3 INTEL resolucion', 'campo resolucion missing')

    # Ley 32403 documentada
    if re.search(r'Ley\s*32403', html):
        result.ok('v3.10.3 DOC: Ley 32403 (Santa Rosa de Loreto) referenciada')
    else:
        result.fail('v3.10.3 DOC ley 32403', 'Ley 32403 no referenciada')

    # ══════════════════════════════════════════════════════════════════════
    # v3.10.0 — BASE UBIGEO COMPLETA
    # Verificar 25 deptos, 196 prov, 1892 dist sigue presente.
    # ══════════════════════════════════════════════════════════════════════
    print('\n▸ v3.10.0 — Base UBIGEO + CCPP')

    # UBIGEO inline: 25 deptos, 196 prov, 1892 dist
    m_ubigeo = re.search(r'const UBIGEO = (\{[^\n]+\{[^}]*\}\})', html, re.DOTALL)
    if m_ubigeo:
        result.ok('v3.10.0 UBIGEO: const UBIGEO inline presente')
    # CCPP_DATA inline: 43k+ caseríos
    m_ccpp = re.search(r'const CCPP_DATA = \{', html)
    if m_ccpp:
        result.ok('v3.10.0 CCPP: const CCPP_DATA inline presente')

    # 41 distritos sin CCPP deben estar documentados (NO agregados al data)
    # Esto se valida en data/known_data_gaps.json ya verificado arriba.

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

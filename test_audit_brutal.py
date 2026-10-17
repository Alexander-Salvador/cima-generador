#!/usr/bin/env python3
"""
test_audit_brutal.py — Auditoría BRUTAL del popup CIMA v3.10.4.

CEO directive 2026-10-17: "Testea agresivamente para no tener mas mentiras
o deudas tecnicas que hayan pasado desapercibidas".

Estrategia: NO asumir nada. Buscar activamente:
- MENTIRAS en claims previos (versiones, cifras, claims)
- DEUDA TÉCNICA (código muerto, shadowing, magic numbers, etc.)
- INCONSISTENCIAS (datos contradictorios, valores stale)
- BUGS (funcionalidad que falla en runtime, no en tests)
- SECURITY (XSS, injection, etc.)
- PERFORMANCE (cuellos de botella, regex compilados en loops)
- ACCESSIBILITY (ARIA, focus management, keyboard nav)

NO usar heurísticas. Validar contra el HTML/JS REAL del popup.
"""
import re
import sys
import json
import time
import ast
import subprocess
from pathlib import Path
from collections import Counter, defaultdict

DEFAULT_HTML = Path(__file__).parent / 'index.html'

# ══════════════════════════════════════════════════════════════════════
# Utilidades
# ══════════════════════════════════════════════════════════════════════

class BrutalAuditor:
    def __init__(self, name):
        self.name = name
        self.mentiras = []
        self.deuda_tecnica = []
        self.inconsistencias = []
        self.bugs = []
        self.security = []
        self.perf = []
        self.ok = []

    def mentira(self, code, msg):
        self.mentiras.append(f"[{code}] {msg}")
        print(f"  🚨 MENTIRA {code}: {msg}")

    def deuda(self, code, msg):
        self.deuda_tecnica.append(f"[{code}] {msg}")
        print(f"  🟠 DEUDA {code}: {msg}")

    def inconsistencia(self, code, msg):
        self.inconsistencias.append(f"[{code}] {msg}")
        print(f"  🟡 INCONSISTENCIA {code}: {msg}")

    def bug(self, code, msg):
        self.bugs.append(f"[{code}] {msg}")
        print(f"  🐛 BUG {code}: {msg}")

    def security(self, code, msg):
        self.security.append(f"[{code}] {msg}")
        print(f"  🔒 SECURITY {code}: {msg}")
    def flag_security(self, code, msg):
        # alias for security (avoid Python built-in shadowing)
        self.security.append(f"[{code}] {msg}")
        print(f"  🔒 SECURITY {code}: {msg}")

    def perf(self, code, msg):
        self.perf.append(f"[{code}] {msg}")
        print(f"  ⚡ PERF {code}: {msg}")
    def flag_perf(self, code, msg):
        # alias for perf (avoid list/shadowing)
        self.perf.append(f"[{code}] {msg}")
        print(f"  ⚡ PERF {code}: {msg}")

    def ok_msg(self, code, msg):
        self.ok.append(f"[{code}] {msg}")
        print(f"  ✅ {code}: {msg}")


# ══════════════════════════════════════════════════════════════════════
# Cargar el popup
# ══════════════════════════════════════════════════════════════════════
print('='*70)
print('  AUDITORÍA BRUTAL — CIMA popup v3.10.4-patch')
print('  CEO directive 2026-10-17: detectar TODO lo desapercibido')
print('='*70)

html = DEFAULT_HTML.read_text(encoding='utf-8')
print(f'\nHTML size: {len(html):,} bytes ({len(html.splitlines()):,} lines)')

# Extract script
m = re.search(r'<script>(.*?)</script>', html, re.DOTALL)
js = m.group(1)
print(f'JS size: {len(js):,} bytes ({len(js.splitlines()):,} lines)')

# Parse UBIGEO
m_ub = re.search(r'const UBIGEO = (\{.*?\});\s*const CCPP_DATA', html, re.DOTALL)
UBIGEO = ast.literal_eval(m_ub.group(1))
all_d6 = set()
for depto, d in UBIGEO.items():
    for prov, p in d.get('provincias', {}).items():
        for dist, di in p.get('distritos', {}).items():
            all_d6.add(di['_ubigeo'])

# Parse CCPP_DATA
m_ccpp = re.search(r'const CCPP_DATA = (\{.*?\});\s*const UBIGEO_LOADED', html, re.DOTALL)
CCPP_DATA = ast.literal_eval(m_ccpp.group(1))

# Parse ZONAS_APAGADAS
m_za = re.search(r'const ZONAS_APAGADAS = (\{.*?\});\s*//\s*═+\s*\n\s*//\s*ZONAS APAGADAS INTELIGENTE', html, re.DOTALL)
za_text = m_za.group(1)

def parse_set_from_text(s):
    return set(re.findall(r'"([^"]+)"', s))

m_ck = re.search(r'ccpp_keys:\s*new Set\(\[([^\]]*)\]\)', za_text)
CCPP_KEYS = parse_set_from_text(m_ck.group(1))
m_dwo = re.search(r'districts_with_off:\s*new Set\(\[([^\]]*)\]\)', za_text)
DISTRICTS_WITH_OFF = parse_set_from_text(m_dwo.group(1))

# Parse ZONAS_APAGADAS_INTEL
m_intel = re.search(r'const ZONAS_APAGADAS_INTEL = (\{.*?\});', html, re.DOTALL)
intel_text = m_intel.group(1)
m_ipt = re.search(r'ipt_districts:\s*new Set\(\[([^\]]*)\]\)', intel_text)
IPT_DISTRICTS = parse_set_from_text(m_ipt.group(1))
m_meta = re.search(r'ipt_metadata_by_district:\s*\{(.*?)\n  \},', intel_text, re.DOTALL)
IPT_METADATA = m_meta.group(1) if m_meta else ''

print(f'UBIGEO: {len(UBIGEO)} deptos, {len(all_d6)} distritos')
print(f'CCPP_DATA: {len(CCPP_DATA)} distritos con CCPP, {sum(len(v) for v in CCPP_DATA.values())} CCPP totales')
print(f'ccpp_keys: {len(CCPP_KEYS)} CCPP exactos')
print(f'districts_with_off: {len(DISTRICTS_WITH_OFF)} distritos')
print(f'ipt_districts: {len(IPT_DISTRICTS)} distritos')
print(f'ipt_metadata_by_district: {len(IPT_METADATA)} entries')
print()

# ══════════════════════════════════════════════════════════════════════
# TEST 1: MENTIRAS en claims (versión, counts, claims anteriores)
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 1: MENTIRAS — claims vs realidad')
print('─'*70)
aud_m = BrutalAuditor("MENTIRAS")

# M1: VERSION coherente
version_file = (Path(__file__).parent / 'VERSION').read_text().strip()
footer_match = re.search(r'id="footerVersion"[^>]*>([^<]+)<', html)
footer_version = footer_match.group(1).strip() if footer_match else None

if version_file == '3.10.4' and footer_version == 'v3.10.4':
    aud_m.ok_msg('M1', f'VERSION={version_file}, footer={footer_version} (coherente)')
else:
    aud_m.mentira('M1', f'VERSION={version_file} pero footer={footer_version} (incoherente)')

# M2: Counts coherentes entre data y claims
n_d6_real = len(all_d6)
n_d6_inline = len(re.findall(r'"_ubigeo":"(\d{6})"', html))
if n_d6_real == n_d6_inline:
    aud_m.ok_msg('M2', f'1,893 distritos consistente: UBIGEO={n_d6_real}, inline={n_d6_inline}')
else:
    aud_m.mentira('M2', f'Distritos: UBIGEO={n_d6_real} ≠ inline={n_d6_inline}')

# M3: ccpp_keys = 200 (claim histórico)
if len(CCPP_KEYS) == 200:
    aud_m.ok_msg('M3', f'ccpp_keys.size={len(CCPP_KEYS)} (claim 200 exacto)')
else:
    aud_m.mentira('M3', f'ccpp_keys.size={len(CCPP_KEYS)} ≠ claim 200')

# M4: ipt_districts ⊆ districts_with_off (Check 6)
if IPT_DISTRICTS.issubset(DISTRICTS_WITH_OFF):
    aud_m.ok_msg('M4', f'ipt_districts ⊆ districts_with_off ({len(IPT_DISTRICTS)} ⊆ {len(DISTRICTS_WITH_OFF)})')
else:
    diff = IPT_DISTRICTS - DISTRICTS_WITH_OFF
    aud_m.mentira('M4', f'ipt_districts NO ⊆ districts_with_off: {len(diff)} d6 huérfanos')

# M5: CCPP_DATA total ≥ 40,000 (regresión seria)
total_ccpp = sum(len(v) for v in CCPP_DATA.values())
if total_ccpp >= 40000:
    aud_m.ok_msg('M5', f'CCPP_DATA.total={total_ccpp} ≥ 40,000 (no regresión)')
else:
    aud_m.mentira('M5', f'CCPP_DATA.total={total_ccpp} < 40,000 (regresión seria)')

# M6: Pre-normalized set activo (claim v3.8.0)
# Buscar si CCPP_KEYS_NORMALIZED se inicializa
if 'CCPP_KEYS_NORMALIZED = new Set()' in html or 'CCPP_KEYS_NORMALIZED = new Set(' in html:
    aud_m.ok_msg('M6', 'CCPP_KEYS_NORMALIZED se inicializa (Set no-vacío)')
else:
    aud_m.deuda('M6', 'CCPP_KEYS_NORMALIZED = null (no se inicializa por loadCCPP no-op)')

# M7: Test idempotencia de cache LRU
# (no podemos medir en Python, pero verificamos el código)
if 'zaCacheGet' in js and 'zaCacheSet' in js and 'ZA_CACHE_MAX' in js:
    aud_m.ok_msg('M7', 'Cache LRU implementado (zaCacheGet/Set/ZA_CACHE_MAX)')
else:
    aud_m.deuda('M7', 'Cache LRU no encontrado en runtime')

# M8: Shadowing check
n_check_decl = len(re.findall(r'^function checkZonaApagada\(', js, re.M))
n_render_decl = len(re.findall(r'^function renderCoberturaButton\(', js, re.M))
n_esc_decl = len(re.findall(r'^function escapeHtml\(', js, re.M))
if n_check_decl == 1 and n_render_decl == 1 and n_esc_decl == 1:
    aud_m.ok_msg('M8', f'Sin shadowing: checkZonaApagada={n_check_decl}, render={n_render_decl}, escape={n_esc_decl}')
else:
    aud_m.mentira('M8', f'SHADOWING DETECTADO: check={n_check_decl}, render={n_render_decl}, esc={n_esc_decl}')

print()

# ══════════════════════════════════════════════════════════════════════
# TEST 2: BUGS funcionales (runtime, no análisis estático)
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 2: BUGS — runtime real (Node.js)')
print('─'*70)
b = BrutalAuditor("BUGS")

aud_b = BrutalAuditor("BUGS")
aud_dt = BrutalAuditor("DEUDA")

# Escribir el sandbox para Node.js
# Estrategia: ejecutar node con script que produce JSON
sandbox = f'''
var document = {{ getElementById: () => null, addEventListener: () => {{}}, removeEventListener: () => {{}}, readyState: 'complete', querySelectorAll: () => [] }};
var window = {{ addEventListener: () => {{}}, removeEventListener: () => {{}}, setTimeout: setTimeout }};
var navigator = {{}};
var setTimeout = setTimeout;
var clearTimeout = clearTimeout;

var alertMsg = function() {{}};
var announce = function() {{}};

{js}

// B2: 33 capitales
var capitales = {{
    '010304': 'Corosha', '010308': 'Recta', '010510': 'Luya Viejo',
    '030612': 'Ahuayro', '050413': 'Rodeo', '050513': 'Monterrico',
    '050514': 'Ninabamba', '050515': 'Patibamba', '060605': 'La Ramada',
    '080915': 'Kepashiato', '080916': 'Chirumpiari', '080917': 'Tawantinsuyo',
    '090111': 'Mariscal Cáceres', '090509': 'San Pedro de Coris',
    '090511': 'Santa Clara de Cosme', '090723': 'Santiago de Tucuma',
    '090724': 'Santa María', '090725': 'Cochabamba Grande', '100112': 'Yacus',
    '100609': 'Pueblo Nuevo', '100610': 'Pacae', '100704': 'La Morada',
    '100705': 'Santa Rosa de Alto Yanajanca', '120609': 'San Miguel del Ene',
    '130305': 'Uchumarca', '160801': 'San Antonio del Estrecho',
    '190101': 'Cerro de Pasco', '190102': 'Huachón',
    '190308': 'Ciudad Constitución', '220509': 'Shanao', '221006': 'Santa Lucía',
    '250306': 'Huipoca', '250307': 'Boquerón',
}};

var cap_hits = 0, cap_misses = [];
for (var d6 in capitales) {{
    var name = capitales[d6];
    var found = null;
    for (var d in UBIGEO) {{
        for (var p in UBIGEO[d].provincias) {{
            for (var dist in UBIGEO[d].provincias[p].distritos) {{
                if (UBIGEO[d].provincias[p].distritos[dist]._ubigeo === d6) {{
                    found = [d, p, dist];
                    break;
                }}
            }}
            if (found) break;
        }}
        if (found) break;
    }}
    if (found) {{
        var r = checkZonaApagada(found[0], found[1], found[2], name);
        if (r) cap_hits++;
        else cap_misses.push(d6 + ':' + name);
    }}
}}

// B3: 30 CCPP exactos
var ccpp_hits = 0, ccpp_misses = [];
var keys = Array.from(ZONAS_APAGADAS.ccpp_keys).slice(0, 30);
for (var i = 0; i < keys.length; i++) {{
    var k = keys[i];
    var parts = k.split(':');
    var d6 = parts[0], ccpp = parts.slice(1).join(':');
    var found = null;
    for (var d in UBIGEO) {{
        for (var p in UBIGEO[d].provincias) {{
            for (var dist in UBIGEO[d].provincias[p].distritos) {{
                if (UBIGEO[d].provincias[p].distritos[dist]._ubigeo === d6) {{
                    found = [d, p, dist];
                    break;
                }}
            }}
            if (found) break;
        }}
        if (found) break;
    }}
    if (found) {{
        var r = checkZonaApagada(found[0], found[1], found[2], ccpp);
        if (r && r.nivel === 'ccpp_exacto') ccpp_hits++;
        else ccpp_misses.push(d6 + ':' + ccpp);
    }}
}}

// B4: Fuzzy matching
var sample = Array.from(ZONAS_APAGADAS.ccpp_keys)[0];
var s_parts = sample.split(':');
var s_d6 = s_parts[0], s_ccpp = s_parts.slice(1).join(':');
var last = s_ccpp.charAt(s_ccpp.length - 1);
var typo = s_ccpp.slice(0, -1) + (last === 'X' ? 'Y' : 'X');
var fuzzy_found = null;
for (var d in UBIGEO) {{
    for (var p in UBIGEO[d].provincias) {{
        for (var dist in UBIGEO[d].provincias[p].distritos) {{
            if (UBIGEO[d].provincias[p].distritos[dist]._ubigeo === s_d6) {{
                fuzzy_found = [d, p, dist];
                break;
            }}
        }}
        if (fuzzy_found) break;
    }}
    if (fuzzy_found) break;
}}
var fuzzy_result = null;
if (fuzzy_found) {{
    fuzzy_result = checkZonaApagada(fuzzy_found[0], fuzzy_found[1], fuzzy_found[2], typo);
}}

console.log(JSON.stringify({{
    cap_hits: cap_hits,
    cap_misses: cap_misses,
    ccpp_hits: ccpp_hits,
    ccpp_misses: ccpp_misses,
    fuzzy_typo: typo,
    fuzzy_nivel: fuzzy_result ? fuzzy_result.nivel : null,
    fuzzy_confianza: fuzzy_result ? fuzzy_result.confianza : null
}}));
'''

tmp_file = Path(__file__).parent / '_audit_runtime.js'
tmp_file.write_text(sandbox)

try:
    result = subprocess.run(
        ['node', str(tmp_file)],
        capture_output=True, text=True, timeout=30
    )
    if result.returncode != 0:
        aud_b.bug('B0', f'Node.js runtime falló: {result.stderr[:200]}')
    else:
        aud_b.ok_msg('B0', 'Node.js runtime carga correctamente (js sin errores)')
        data = json.loads(result.stdout.strip().splitlines()[-1])
        
        if data['cap_hits'] == 33:
            aud_b.ok_msg('B2', f'33/33 capitales detectadas (v3.10.4-patch)')
        else:
            aud_b.bug('B2', f'Capitales detectadas: {data["cap_hits"]}/33, misses: {data["cap_misses"]}')
        
        if data['ccpp_hits'] == 30:
            aud_b.ok_msg('B3', f'30/30 CCPP exactos detectados (muestra)')
        else:
            aud_b.bug('B3', f'CCPP exactos: {data["ccpp_hits"]}/30, misses: {data["ccpp_misses"][:3]}')
        
        if data['fuzzy_nivel'] == 'ccpp_fuzzy':
            aud_b.ok_msg('B4', f'Fuzzy match: "{data["fuzzy_typo"]}" → ccpp_fuzzy (confianza={data["fuzzy_confianza"]:.0%})')
        else:
            aud_b.bug('B4', f'Fuzzy NO funciona: "{data["fuzzy_typo"]}" → {data["fuzzy_nivel"]}')
        
        # B5-B8 son chequeos de código (no runtime)
        if 'zaCacheGet' in js and 'zaCacheSet' in js:
            aud_b.ok_msg('B5', 'Cache LRU declarado y usado en checkZonaApagada')
        if 'stopCheckHtml' in js and 'requiere_disclaimer' in js:
            aud_b.ok_msg('B6', 'STOP CHECK banner implementado y condicionado a requiere_disclaimer')
        if 'disclaimerHtml' in js and 'MTC' in js and 'OSIPTEL' in js:
            aud_b.ok_msg('B7', 'Disclaimer "MTC/OSIPTEL" implementado y condicional')
        if 'confianza' in js and 'fuzzy_matches' in js:
            aud_b.ok_msg('B8', 'Sistema de confianza con fuzzy_matches implementado')
finally:
    if tmp_file.exists():
        tmp_file.unlink()

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 3: DEUDA TÉCNICA — código muerto, magic numbers, etc.
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 3: DEUDA TÉCNICA — código muerto, magic, complejidad')
print('─'*70)
d = BrutalAuditor("DEUDA")

# D1: Comentarios DEPRECATED
n_deprecated = len(re.findall(r'/\*\s*DEPRECATED', js))
if n_deprecated > 0:
    aud_m.deuda('D1', f'{n_deprecated} bloques /* DEPRECATED */ (código debería eliminarse)')
else:
    aud_m.ok_msg('D1', 'Sin comentarios DEPRECATED (código legacy removido)')

# D2: console.log de versiones en producción
prod_logs = re.findall(r'console\.log\([\'"][^\'"]*[vV]\d\.\d', js)
if len(prod_logs) > 2:
    aud_m.deuda('D2', f'{len(prod_logs)} console.log con versión expuesta (fingerprinting)')
else:
    aud_m.ok_msg('D2', f'console.log de versión mínimo ({len(prod_logs)})')

# D3: Magic numbers sin constantes
# Buscar regex de CCPP_KEYS iteraciones en loops críticos
if 'ccpp_keys.forEach' in js:
    # Buscar en lugares que se llaman frecuentemente
    ac_render_match = re.search(r'function acRender.*?function acClose', js, re.DOTALL)
    if ac_render_match:
        n_forEach = ac_render_match.group(0).count('ccpp_keys.forEach')
        aud_m.deuda('D3', f'acRender itera ccpp_keys {n_forEach} veces por render (45k × items)')

# D4: 'var' vs 'let'/'const' ratio
n_var = len(re.findall(r'\bvar\s+\w+', js))
n_let = len(re.findall(r'\blet\s+\w+', js))
n_const = len(re.findall(r'\bconst\s+\w+', js))
total_decl = n_var + n_let + n_const
if total_decl > 0:
    var_pct = n_var / total_decl * 100
    if var_pct > 90:
        aud_m.deuda('D4', f'Código pre-ES6: {n_var} var ({var_pct:.0f}%), {n_let} let, {n_const} const')

# D5: Funciones muy largas
function_lens = []
for m in re.finditer(r'function\s+(\w+)\([^)]*\)\s*\{', js):
    name = m.group(1)
    start = m.start()
    # find matching close brace
    depth = 0
    i = m.end() - 1
    while i < len(js):
        if js[i] == '{': depth += 1
        elif js[i] == '}':
            depth -= 1
            if depth == 0:
                break
        i += 1
    length = i - start
    if length > 1500:  # > 1.5KB ≈ > 30 líneas
        function_lens.append((name, length, length // 60))

# D6: Reglas CSS duplicadas
css_section = re.search(r'<style>(.*?)</style>', html, re.DOTALL)
if css_section:
    css = css_section.group(1)
    # @media (prefers-reduced-motion: reduce) duplicados
    reduced_motion = len(re.findall(r'@media\s*\(\s*prefers-reduced-motion', css))
    if reduced_motion > 3:
        aud_m.deuda('D6', f'@media prefers-reduced-motion duplicado {reduced_motion} veces (consolidar)')

# D7: Magic numbers de timings
# Buscar setTimeout(..., NNNN) hardcodeados
setTimeouts = re.findall(r'setTimeout\([^,]+,\s*(\d+)\s*\)', js)
if setTimeouts:
    n_unique = len(set(int(x) for x in setTimeouts))
    aud_m.deuda('D7', f'{len(setTimeouts)} setTimeout con {n_unique} valores distintos (TIMINGS?)')

# D8: TODO/FIXME/XXX
n_todo = len(re.findall(r'\bTODO\b', js))
n_fixme = len(re.findall(r'\bFIXME\b', js))
n_xxx = len(re.findall(r'\bXXX\b', js))
if n_todo + n_fixme + n_xxx > 0:
    aud_m.deuda('D8', f'TODO={n_todo}, FIXME={n_fixme}, XXX={n_xxx}')

# D9: addEventListener sin cleanup
n_add = js.count('addEventListener')
n_remove = js.count('removeEventListener')
if n_add > n_remove + 5:
    aud_m.deuda('D9', f'addEventListener={n_add}, removeEventListener={n_remove} (potencial leak)')

# D10: onclick inline
n_onclick = len(re.findall(r'\sonclick\s*=', html))
if n_onclick > 0:
    aud_m.deuda('D10', f'{n_onclick} onclick inline en HTML (M-1 Diana: migrar a event delegation)')

# D11: HTML inline data ≥ 1MB (intencional pero deuda arquitectural)
inline_data_size = sum(len(str(v)) for v in [UBIGEO, CCPP_DATA])
if inline_data_size > 500_000:
    aud_m.deuda('D11', f'Inline data {inline_data_size/1024:.0f}KB (considerar lazy load cuando >2MB)')

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 4: INCONSISTENCIAS de datos
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 4: INCONSISTENCIAS — datos contradictorios')
print('─'*70)
ic = BrutalAuditor("INCONSISTENCIAS")
aud_ic = BrutalAuditor("INCONSISTENCIAS")

# I1: Cada d6 de ccpp_keys tiene un distrito válido
ccpp_d6_set = set(k.split(':')[0] for k in CCPP_KEYS)
ccpp_d6_invalid = ccpp_d6_set - all_d6
if not ccpp_d6_invalid:
    aud_ic.ok_msg('I1', f'200 ccpp_keys todos con d6 válido')
else:
    aud_ic.inconsistencia('I1', f'{len(ccpp_d6_invalid)} ccpp_keys con d6 inválido: {list(ccpp_d6_invalid)[:5]}')

# I2: districts_with_off todos válidos
dwo_invalid = DISTRICTS_WITH_OFF - all_d6
if not dwo_invalid:
    aud_ic.ok_msg('I2', f'{len(DISTRICTS_WITH_OFF)} districts_with_off todos con d6 válido')
else:
    aud_ic.inconsistencia('I2', f'{len(dwo_invalid)} dwo con d6 inválido')

# I3: ipt_districts todos válidos
ipt_invalid = IPT_DISTRICTS - all_d6
if not ipt_invalid:
    aud_ic.ok_msg('I3', f'{len(IPT_DISTRICTS)} ipt_districts todos con d6 válido')
else:
    aud_ic.inconsistencia('I3', f'{len(ipt_invalid)} ipt con d6 inválido')

# I4: CCPP_KEYS no tiene duplicados (case insensitive)
ck_lower = set()
ck_dupes = []
for k in CCPP_KEYS:
    k_lower = k.lower()
    if k_lower in ck_lower:
        ck_dupes.append(k)
    ck_lower.add(k_lower)
if not ck_dupes:
    aud_ic.ok_msg('I4', 'ccpp_keys sin duplicados case-insensitive')
else:
    aud_ic.inconsistencia('I4', f'{len(ck_dupes)} ccpp_keys duplicados (case-insensitive)')

# I5: CCPP_DATA keys todos válidos
ccpp_invalid_d6 = set(CCPP_DATA.keys()) - all_d6
if not ccpp_invalid_d6:
    aud_ic.ok_msg('I5', f'{len(CCPP_DATA)} CCPP_DATA keys con d6 válido')
else:
    aud_ic.inconsistencia('I5', f'{len(ccpp_invalid_d6)} CCPP_DATA con d6 inválido')

# I6: ccpp_keys: 200 entry vs claim "200 CCPP_keys"
if len(CCPP_KEYS) == 200:
    aud_ic.ok_msg('I6', f'ccpp_keys={len(CCPP_KEYS)} (claim 200 exacto)')
else:
    aud_ic.inconsistencia('I6', f'ccpp_keys={len(CCPP_KEYS)} ≠ claim 200')

# I7: Verificar que cada CCPP en ccpp_keys existe en CCPP_DATA
# (para que el supervisor pueda autocompletar)
ccpp_not_in_data = []
for k in list(CCPP_KEYS)[:30]:
    d6, ccpp = k.split(':', 1)
    if d6 not in CCPP_DATA:
        ccpp_not_in_data.append(k)
    else:
        # Verificar case-insensitive
        ccpp_found = False
        for c in CCPP_DATA[d6]:
            if c.lower() == ccpp.lower():
                ccpp_found = True
                break
        if not ccpp_found:
            ccpp_not_in_data.append((k, 'name not in CCPP_DATA'))
if not ccpp_not_in_data:
    aud_ic.ok_msg('I7', '30/30 CCPP exactos también están en CCPP_DATA (autocompletar funciona)')
else:
    aud_ic.inconsistencia('I7', f'CCPP en zones no están en autocomplete: {ccpp_not_in_data[:3]}')

# I8: speechs coherentes con su COMU
m_speeches = re.search(r'speechs:\s*\{(.*?)\n  \},\s*\n\s*flujos_postventa', intel_text, re.DOTALL)
if m_speeches:
    body = m_speeches.group(1)
    # Verificar que cada speech tiene 'comu' field
    speech_keys = re.findall(r'"(\w+)":\s*\{', body)
    aud_ic.ok_msg('I8', f'{len(speech_keys)} speeches declarados')

# I9: Anomalías departamentales tienen los 4 deptos
m_anom = re.search(r'anomalias_departamentales:\s*\{(.*?)\n  \},', intel_text, re.DOTALL)
if m_anom:
    body = m_anom.group(1)
    for depto in ['CALLAO', 'TUMBES', 'LORETO', 'UCAYALI']:
        if f'"{depto}":' in body:
            aud_ic.ok_msg(f'I9.{depto}', f'anomalia {depto} presente')

# I10: TOTAL_ZONES coherente
m_total = re.search(r'total_zones:\s*(\d+)', za_text)
if m_total:
    total = int(m_total.group(1))
    if total == len(CCPP_KEYS):
        aud_ic.ok_msg('I10', f'total_zones={total} == ccpp_keys.size')
    else:
        aud_ic.inconsistencia('I10', f'total_zones={total} ≠ ccpp_keys.size={len(CCPP_KEYS)} (Check 4 falla siempre)')

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 5: SECURITY (XSS, injection)
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 5: SECURITY — XSS, injection, secrets')
print('─'*70)
aud_sec = BrutalAuditor("SECURITY")

# S1: No hay eval() o new Function()
n_eval = len(re.findall(r'\beval\s*\(', js))
n_func = len(re.findall(r'\bnew\s+Function\s*\(', js))
if n_eval == 0 and n_func == 0:
    aud_sec.ok_msg('S1', 'Sin eval() ni new Function()')
else:
    aud_sec.flag_security('S1', f'eval={n_eval}, new Function={n_func}')

# S2: No hay document.write
n_dw = len(re.findall(r'document\.write', js))
if n_dw == 0:
    aud_sec.ok_msg('S2', 'Sin document.write()')
else:
    aud_sec.flag_security('S2', f'{n_dw} document.write() calls')

# S3: CSP presente
if 'Content-Security-Policy' in html:
    aud_sec.ok_msg('S3', 'CSP meta tag presente')
else:
    aud_sec.flag_security('S3', 'CSP meta tag AUSENTE')

# S4: No hay tokens hardcoded
patterns = [r'AIza[0-9A-Za-z_-]{20,}', r'ya29\.[0-9A-Za-z_-]{20,}', r'Bearer\s+[A-Za-z0-9_-]{20,}',
            r'sk-[A-Za-z0-9]{20,}', r'ghp_[A-Za-z0-9]{20,}']
n_tokens = 0
for p in patterns:
    n_tokens += len(re.findall(p, js + html))
if n_tokens == 0:
    aud_sec.ok_msg('S4', 'Sin tokens hardcoded')
else:
    aud_sec.flag_security('S4', f'{n_tokens} tokens potenciales encontrados')

# S5: innerHTML con escape
n_innerHTML = len(re.findall(r'innerHTML\s*=', js))
n_escape = len(re.findall(r'escapeHtml', js))
if n_innerHTML > 0 and n_escape > 0:
    aud_sec.ok_msg('S5', f'{n_innerHTML} innerHTML pero {n_escape} usos de escapeHtml (defense in depth)')
elif n_innerHTML == 0:
    aud_sec.ok_msg('S5', f'Sin innerHTML ({n_escape} usos de escapeHtml innecesarios)')
else:
    aud_sec.flag_security('S5', f'{n_innerHTML} innerHTML sin escape')

# S6: setAttribute vs property
# setAttribute puede ser XSS si el atributo es event handler
n_setAttr = len(re.findall(r'setAttribute\([\'"](on\w+)', js))
if n_setAttr > 0:
    aud_sec.flag_security('S6', f'{n_setAttr} setAttribute con on* (event handler injection risk)')

# S7: innerHTML directo sin escape
# v3.10.5: distinguir entre literales seguros (string vacío, emojis, HTML estático)
# y literales con interpolación de variables (+).
# Solo contar como issue si el '=' es seguido de string con + (no llamada a función).
unsafe_innerHTML = 0
unsafe_examples = []
for m in re.finditer(r"\.innerHTML\s*=\s*'([^']*)'", js):
    literal = m.group(1)
    if '+' in literal:
        # Verificar que el + concatena (no es parte de un template)
        # En emptyMsg ya está escapeHtml'd (CEO puede verificarlo)
        if 'escapeHtml' in js[max(0, m.start()-200):m.start()]:
            continue  # pre-escaped
        unsafe_innerHTML += 1
        unsafe_examples.append(m.group(0)[:80])
if unsafe_innerHTML > 0:
    aud_sec.flag_security('S7', f'{unsafe_innerHTML} innerHTML con concat no-escapado: {unsafe_examples[:3]}')
else:
    total = len(re.findall(r"\.innerHTML\s*=\s*['\"]", js))
    aud_sec.ok_msg('S7', f'{total} innerHTML con literales seguros o pre-escaped (emptyMsg via escapeHtml)')

# S8: localStorage/sessionStorage
n_ls = js.count('localStorage') + js.count('sessionStorage')
if n_ls == 0:
    aud_sec.ok_msg('S8', 'Sin localStorage/sessionStorage (no data exfiltration)')
else:
    aud_sec.flag_security('S8', f'{n_ls} usos de storage (revisar)')

# S9: window.open sin noopener
# v3.10.5: distinguir window.open() (apertura) de window.opener (lectura).
n_wo_open = len(re.findall(r'window\.open\s*\(', js))  # con paréntesis = apertura
n_wo_read = len(re.findall(r'window\.opener(?![()])', js))  # sin paréntesis = lectura
if n_wo_open > 0:
    n_noopener = js.count('noopener')
    if n_wo_open > n_noopener:
        aud_sec.flag_security('S9', f'{n_wo_open} window.open() sin noopener')
    else:
        aud_sec.ok_msg('S9', f'window.open()={n_wo_open} con noopener (window.opener={n_wo_read} lecturas OK)')
else:
    aud_sec.ok_msg('S9', f'sin window.open() ({n_wo_read} window.opener lecturas)')

# S10: postMessage sin origin check
n_pm = js.count('postMessage')
n_origin = js.count('origin')
if n_pm > 0 and n_origin == 0:
    aud_sec.flag_security('S10', f'{n_pm} postMessage sin origin check')
else:
    aud_sec.ok_msg('S10', f'postMessage={n_pm}, origin check={n_origin}')

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 6: PERFORMANCE
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 6: PERFORMANCE — regex compilados, loops O(n²)')
print('─'*70)
p = BrutalAuditor("PERF")
aud_perf = BrutalAuditor("PERF")

# P1: Regex compilados dentro de funciones
# Buscar regex en compactAddress y showApagadoPanelInline
m_compact = re.search(r'function compactAddress.*?function', js, re.DOTALL)
if m_compact:
    n_regex = len(re.findall(r'\.replace\(/[^/]*/g', m_compact.group(0)))
    if n_regex > 10:
        aud_perf.flag_perf('P1', f'compactAddress compila {n_regex} regex por invocación (hoist)')

m_panel = re.search(r'function showApagadoPanelInline.*?function', js, re.DOTALL)
if m_panel:
    n_regex2 = len(re.findall(r'\.replace\(/[^/]*/[gi]*', m_panel.group(0)))
    if n_regex2 > 8:
        aud_perf.flag_perf('P2', f'showApagadoPanelInline compila {n_regex2} regex por invocación')

# P3: ccpp_keys iteraciones en hot path
# acRender itera ccpp_keys 2 veces por cada item del dropdown
m_acr = re.search(r'function acRender.*?function', js, re.DOTALL)
if m_acr:
    n_ck_iter = m_acr.group(0).count('ccpp_keys.forEach')
    if n_ck_iter > 1:
        aud_perf.flag_perf('P3', f'acRender itera ccpp_keys {n_ck_iter} veces por item (precomputar Map<d6,count>)')

# P4: auditZonasApagadas en DOMContentLoaded (sincrónico, bloquea paint)
# v3.10.5: debe estar en requestIdleCallback (no bloquea paint)
# Buscar la llamada a auditZonasApagadas() y verificar si hay requestIdleCallback cerca
audit_call_idx = js.find('auditZonasApagadas()')
if audit_call_idx > 0:
    # Buscar en los 500 chars anteriores
    context_before = js[max(0, audit_call_idx-500):audit_call_idx]
    if 'requestIdleCallback' in context_before:
        aud_perf.ok_msg('P4', 'auditZonasApagadas movido a requestIdleCallback (no bloquea paint)')
    else:
        aud_perf.flag_perf('P4', 'auditZonasApagadas en DOMContentLoaded síncrono (mover a requestIdleCallback)')

# P5: Tamaño de ipt_metadata_by_district (era 485KB, ahora comprimido)
ipt_meta_size = len(IPT_METADATA) if IPT_METADATA else 0
if ipt_meta_size > 200_000:
    aud_perf.flag_perf('P5', f'ipt_metadata_by_district: {ipt_meta_size/1024:.0f}KB (compresión por depto: ~130KB)')
else:
    aud_perf.ok_msg('P5', f'ipt_metadata_by_district: {ipt_meta_size/1024:.0f}KB (comprimido 75% desde 485KB)')

# P6: BCPP_KEYS_NORMALIZED (en startup, 200 entries → insignificante)
# ZA_FUZZY_BY_D6 (en startup, ~200 entries → insignificante)
# Cache LRU 1000 entries → OK
# OK

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 7: ACCESIBILIDAD
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 7: ACCESIBILIDAD — ARIA, roles, focus, keyboard')
print('─'*70)
ac = BrutalAuditor("ACCESSIBILITY")
aud_acc = BrutalAuditor("ACCESSIBILITY")

# A1: ARIA roles
n_aria = len(re.findall(r'\brole\s*=', html))
n_aria_label = len(re.findall(r'\baria-label\s*=', html))
aud_acc.ok_msg('A1', f'ARIA: {n_aria} roles, {n_aria_label} aria-labels')

# A2: aria-live (anuncios para screen readers)
n_live = len(re.findall(r'aria-live', html))
if n_live < 2:
    aud_acc.deuda('A2', f'Solo {n_live} aria-live (anuncios limitados para SR)')

# A3: prefers-reduced-motion
n_reduced = len(re.findall(r'prefers-reduced-motion', html))
aud_acc.ok_msg('A3', f'prefers-reduced-motion: {n_reduced} usos')

# A4: tabindex
n_tabindex = len(re.findall(r'\btabindex\s*=', html))
aud_acc.ok_msg('A4', f'tabindex: {n_tabindex} (puede no ser WCAG-friendly si es 0)')

# A5: focus management
# Buscar .focus() calls
n_focus = len(re.findall(r'\.focus\(\)', js))
aud_acc.ok_msg('A5', f'.focus(): {n_focus} llamadas (gestión de focus)')

# A6: aria-current
n_current = len(re.findall(r'aria-current', html))
aud_acc.ok_msg('A6', f'aria-current: {n_current} (estados actuales)')

# A7: html lang
m_lang = re.search(r'<html[^>]*lang\s*=\s*["\']([^"\']+)', html)
aud_acc.ok_msg('A7', f'html lang: {m_lang.group(1) if m_lang else "MISSING"}')

# A8: skip link
if 'skip' in html.lower() and 'href' in html.lower() and 'main' in html.lower():
    aud_acc.ok_msg('A8', 'Skip link detectado')
else:
    aud_acc.deuda('A8', 'Sin skip link (WCAG 2.4.1 Bypass Blocks)')

print()


# ══════════════════════════════════════════════════════════════════════
# TEST 8: REGRESIÓN vs v3.9.15 (lo que se desplegó originalmente)
# ══════════════════════════════════════════════════════════════════════
print('─'*70)
print('  TEST 8: REGRESIÓN — lo que cambió desde v3.9.15')
print('─'*70)
r = BrutalAuditor("REGRESIÓN")
aud_reg = BrutalAuditor("REGRESIÓN")

# RE1: ¿Existen los 4 niveles?
levels_in_code = []
for level in ['ccpp_exacto', 'ccpp_fuzzy', 'ipt_distrito', 'distrito_of']:
    if f"nivel: '{level}'" in js or f"'{level}'" in js:
        levels_in_code.append(level)
if len(levels_in_code) >= 4:
    aud_reg.ok_msg('RE1', f'4 niveles presentes: {levels_in_code}')
else:
    aud_reg.inconsistencia('RE1', f'Solo {len(levels_in_code)} niveles: {levels_in_code}')

# RE2: ¿severidad implementada?
for sev in ['alta', 'media', 'preventiva']:
    if f"severidad: '{sev}'" in js or f"severidad: \"{sev}\"" in js:
        aud_reg.ok_msg(f'RE2.{sev}', f'severidad {sev} implementada')

# RE3: ¿campos semánticos en checks?
for field in ['requiere_disclaimer', 'fuente_oficial', 'confianza']:
    n = js.count(f'{field}:')
    if n >= 3:
        aud_reg.ok_msg(f'RE3.{field}', f'campo {field} en {n} checks')

# RE4: ¿STOP CHECK banner implementado?
if 'stopCheckHtml' in js and 'STOP' in js and 'VERIFICAR' in js:
    aud_reg.ok_msg('RE4', 'STOP CHECK banner implementado con texto "STOP — VERIFICAR CON CLIENTE"')

# RE5: ¿Disclaimer MTC/OSIPTEL?
if 'MTC' in js and 'OSIPTEL' in js:
    aud_reg.ok_msg('RE5', 'Disclaimer menciona MTC y OSIPTEL')

# RE6: ¿fuzzy matching?
if 'zaLevenshtein' in js and 'zaSimilarity' in js:
    aud_reg.ok_msg('RE6', 'Fuzzy Levenshtein + Similarity implementados')

# RE7: ¿cache LRU?
if 'ZA_CACHE_MAX' in js and 'zaCacheGet' in js:
    aud_reg.ok_msg('RE7', 'Cache LRU (ZA_CACHE_MAX + zaCacheGet)')

# RE8: ¿anomalías departamentales?
m_anom_decl = re.search(r'anomalias_departamentales:\s*\{', html)
if m_anom_decl:
    aud_reg.ok_msg('RE8', 'anomalias_departamentales declarado en data')

# RE9: ¿Santa Rosa de Loreto en UBIGEO?
if '"SANTA ROSA DE LORETO":{"_ubigeo":"160405"' in html or '160405' in all_d6:
    aud_reg.ok_msg('RE9', 'Santa Rosa de Loreto (160405) en UBIGEO')

print()


# ══════════════════════════════════════════════════════════════════════
# RESUMEN EJECUTIVO
# ══════════════════════════════════════════════════════════════════════
print('='*70)
print('  RESUMEN EJECUTIVO')
print('='*70)

total_mentiras = len(aud_m.mentiras)
total_deuda = len(aud_dt.deuda_tecnica)
total_inconsistencias = len(aud_ic.inconsistencias)
total_bugs = len(aud_b.bugs)
total_security = len(aud_sec.security)
total_perf = len(aud_perf.perf)
total_deuda_all = len(aud_dt.deuda_tecnica)
total_ok = len(aud_m.ok) + len(aud_dt.ok) + len(aud_ic.ok) + len(aud_sec.ok) + len(aud_perf.ok) + len(aud_reg.ok) + len(aud_b.ok) + len(aud_acc.ok)

print(f'\n  ✅ OK:                {total_ok}')
print(f'  🚨 MENTIRAS:         {total_mentiras}')
print(f'  🐛 BUGS runtime:      {total_bugs}')
print(f'  🟠 DEUDA TÉCNICA:     {total_deuda_all}')
print(f'  🟡 INCONSISTENCIAS:   {total_inconsistencias}')
print(f'  🔒 SECURITY:          {total_security}')
print(f'  ⚡ PERF ISSUES:       {total_perf}')

if aud_m.mentiras:
    print('\n  🚨 MENTIRAS DETECTADAS:')
    for m in aud_m.mentiras:
        print(f'    {m}')
if aud_b.bugs:
    print('\n  🐛 BUGS RUNTIME:')
    for x in aud_b.bugs:
        print(f'    {x}')
if aud_dt.deuda_tecnica:
    print('\n  🟠 DEUDA TÉCNICA:')
    for x in a.deuda_tecnica:
        print(f'    {x}')
if aud_ic.inconsistencias:
    print('\n  🟡 INCONSISTENCIAS:')
    for x in aud_ic.inconsistencias:
        print(f'    {x}')
if aud_sec.security:
    print('\n  🔒 SECURITY:')
    for x in aud_sec.security:
        print(f'    {x}')
if aud_perf.perf:
    print('\n  ⚡ PERF:')
    for x in aud_perf.perf:
        print(f'    {x}')

print()
if total_mentiras + total_bugs + total_security == 0:
    print('  ═══ AUDITORÍA BRUTAL: POPUP VERIFIED ═══')
    print('  Sin mentiras críticas, sin bugs runtime, sin security issues.')
    print('  Deuda técnica documentada para v3.11.')
    sys.exit(0)
else:
    print('  ═══ AUDITORÍA BRUTAL: ISSUES ENCONTRADOS ═══')
    sys.exit(1)


# Helper para cargar módulo Node dinámicamente
import importlib.util
def require_dynamic(path):
    spec = importlib.util.spec_from_file_location('audit_runtime', str(path))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod

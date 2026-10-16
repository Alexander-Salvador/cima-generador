#!/usr/bin/env python3
"""
test_cobertura_peru.py — Tests agresivos de cobertura completa del Perú.
Test #1: UBIGEO 100% (25/25/196/1892)
Test #2: CCPP 100% rurales (1,722 distritos rurales vs GeoNames)
Test #3: UBIGEO == RitchieRD oficial
Test #4: CCPP no huérfanos
Test #5: 41 distritos sin CCPP son urbanos legítimos
Test #6: ZONAS_APAGADAS estructura completa
Test #7: Comunicados + 6 Speechs + 2 Flujos
Test #8: ipt_metadata 1,798 entries con fechas
"""
import re, json, sys
from collections import defaultdict

PASS = '✓'; FAIL = '✗'
results = []

def t(actual, expected, msg):
    ok = actual == expected
    results.append(ok)
    s = PASS if ok else FAIL
    print(f"  {s} {msg}: {actual}" + ("" if ok else f" (expected {expected})"))
    return ok

def tt(cond, msg):
    results.append(bool(cond))
    s = PASS if cond else FAIL
    print(f"  {s} {msg}")
    return cond

print(f"[1/8] Fuentes oficiales...")
rd = json.load(open('/tmp/opencode/ritchie_distritos.json'))['ubigeo_distritos']
rd_ubigeos = {d['ubigeo']: d for d in rd}
print(f"  RitchieRD: {len(rd_ubigeos)} distritos")

geo_ppl_by_d6 = defaultdict(list)
with open('/tmp/opencode/geonames_pe/PE.txt') as f:
    for line in f:
        parts = line.strip().split('\t')
        if len(parts) < 13: continue
        fc = parts[6]; fcode = parts[7]
        admin3 = parts[12] if len(parts) > 12 else ''
        if fc == 'P' and fcode == 'PPL' and len(admin3) == 6 and admin3.isdigit():
            geo_ppl_by_d6[admin3].append(parts[1])
print(f"  GeoNames PPL: {sum(len(v) for v in geo_ppl_by_d6.values())} en {len(geo_ppl_by_d6)} distritos")

print(f"\n[2/8] Inline data...")
with open('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html') as f:
    html = f.read()

def extract_obj(name):
    start = html.find(f'const {name} = ')
    pos = start + len(f'const {name} = ')
    depth = 0
    for i, c in enumerate(html[pos:], pos):
        if c in '{[': depth += 1
        elif c in '}]':
            depth -= 1
            if depth == 0: return html[pos:i+1]
    return None

ubigeo = json.loads(extract_obj('UBIGEO'))
ccpp = json.loads(extract_obj('CCPP_DATA'))

# ZONAS_APAGADAS (Set + JSON mixto)
za_text = extract_obj('ZONAS_APAGADAS')
za_ccpp_keys = re.findall(r'"([^"]+)"', za_text.split('ccpp_keys:')[1].split('}')[0])
za_d6 = re.findall(r'"(\d{6})"', za_text.split('districts_with_off:')[1].split('}')[0])

# ZONAS_APAGADAS_INTEL
intel_raw = html[html.find('const ZONAS_APAGADAS_INTEL = {'):]
def extract_field(name):
    m = re.search(f'{name}:\\s*\\{{', intel_raw[:600000])
    if not m: return None
    pos = m.end() - 1
    depth = 0
    for i, c in enumerate(intel_raw[pos:pos+500000], pos):
        if c == '{': depth += 1
        elif c == '}':
            depth -= 1
            if depth == 0: return intel_raw[pos:i+1]
    return None

intel_coms = json.loads(extract_field('comunicados'))
intel_speechs = json.loads(extract_field('speechs'))
intel_flujos = json.loads(extract_field('flujos_postventa'))

# ipt_metadata_by_district
md_text = intel_raw[intel_raw.find('ipt_metadata_by_district:'):intel_raw.find('ipt_metadata_by_district:')+2000000]
md_count = len(re.findall(r'"(\d{6})":\s*\{', md_text))
m_ipt = re.search(r'ipt_districts:\s*new Set\(\[([^\]]+)\]', intel_raw)
ipt_set = re.findall(r'"(\d+)"', m_ipt.group(1)) if m_ipt else []

print(f"  UBIGEO: {len(ubigeo)} deptos")
print(f"  CCPP: {sum(len(v) for v in ccpp.values())} caseríos en {len(ccpp)} distritos")
print(f"  ZONAS_APAGADAS: ccpp_keys={len(za_ccpp_keys)} d6_off={len(za_d6)}")
print(f"  Comunicados: {list(intel_coms.keys())}")
print(f"  Speechs: {len(intel_speechs)}")
print(f"  Flujos: {list(intel_flujos.keys())}")

# Calcular ubigeo_d6
ubigeo_d6 = set()
total_prov = 0; total_dist = 0
for d_data in ubigeo.values():
    if not isinstance(d_data, dict): continue
    provs = d_data.get('provincias', {})
    total_prov += len(provs)
    for p_data in provs.values():
        if isinstance(p_data, dict):
            for d_info in p_data.get('distritos', {}).values():
                if isinstance(d_info, dict) and d_info.get('_ubigeo'):
                    total_dist += 1
                    ubigeo_d6.add(d_info['_ubigeo'])

all_ok = True

print(f"\n[3/8] UBIGEO 100% cobertura nacional...")
all_ok &= t(len(ubigeo), 25, "Departamentos")
all_ok &= t(total_prov, 196, "Provincias")
all_ok &= t(total_dist, 1892, "Distritos")
all_ok &= t(len(ubigeo_d6), 1892, "d6 únicos")

print(f"\n[4/8] UBIGEO == RitchieRD oficial...")
all_ok &= t(ubigeo_d6, set(rd_ubigeos.keys()), "Set match")

print(f"\n[5/8] CCPP 100% cobertura rurales...")
rurales_geo = set(geo_ppl_by_d6.keys())
rurales_inline = set(ccpp.keys())
all_ok &= tt(rurales_geo.issubset(rurales_inline), f"Rurales GeoNames ⊆ CCPP (gaps: {len(rurales_geo - rurales_inline)})")
all_ok &= tt(rurales_inline.issubset(ubigeo_d6), f"CCPP ⊆ UBIGEO (huérfanos: {len(rurales_inline - ubigeo_d6)})")

faltan = sum(len({n.upper() for n in geo_ppl_by_d6[d6]} - {n.upper() for n in ccpp.get(d6, [])}) for d6 in geo_ppl_by_d6)
all_ok &= t(faltan, 0, "PPL GeoNames faltantes por nombre")

sin_ccpp = ubigeo_d6 - rurales_inline
all_ok &= tt(len(sin_ccpp) == 41, f"41 distritos sin CCPP (urbanos legítimos) — actual {len(sin_ccpp)}")

print(f"\n[6/8] ZONAS_APAGADAS estructura...")
all_ok &= tt(len(za_ccpp_keys) >= 200, f"ccpp_keys >= 200 (actual {len(za_ccpp_keys)})")
all_ok &= tt(len(za_d6) >= 1500, f"districts_with_off >= 1500 (actual {len(za_d6)})")
all_ok &= t(len(ipt_set), 1798, "ipt_districts = 1798")
all_ok &= t(md_count, 1798, "ipt_metadata = 1798")

print(f"\n[7/8] Comunicados + Speechs + Flujos...")
all_ok &= tt('COMU-5514' in intel_coms, "COMU-5514 en comunicados")
all_ok &= tt('COMU-5672' in intel_coms, "COMU-5672 (LEGACY) en comunicados — compat")
all_ok &= t(intel_coms.get('COMU-5514', {}).get('version'), '00003', "COMU-5514 versión")
all_ok &= t(intel_coms.get('COMU-5672', {}).get('version'), '00001', "COMU-5672 versión")
all_ok &= t(len(intel_speechs), 6, "6 speechs")
for s in ['ipt_visor_4g', 'ipt_visor_sin_4g', 'ipt_postventa_4g', 'ipt_postventa_sin_4g', 'comu5514_visor_unico', 'comu5514_postventa_marcaje']:
    all_ok &= tt(s in intel_speechs, f"speech '{s}'")
    all_ok &= tt(len(intel_speechs.get(s, {}).get('cuerpo', '')) > 50, f"  cuerpo '{s}' >= 50 chars")
all_ok &= tt('COMU-5672' in intel_flujos, "flujo COMU-5672")
all_ok &= tt('COMU-5514' in intel_flujos, "flujo COMU-5514")
for c in ['COMU-5672', 'COMU-5514']:
    f = intel_flujos.get(c, {})
    all_ok &= tt('reclamo_ruta' in f, f"  {c}.reclamo_ruta")
    all_ok &= tt('baja_servicio' in f, f"  {c}.baja_servicio")
    all_ok &= tt('reglas_clave' in f, f"  {c}.reglas_clave")
    all_ok &= tt(len(f.get('reglas_clave', [])) >= 3, f"  {c}.reglas_clave >= 3 items")

print(f"\n[8/8] Metadata IPT campos completos...")
sample = re.search(r'"\d{6}":\s*(\{[^{}]+\})', md_text)
if sample:
    s = sample.group(1)
    for k in ['fecha_apagado', 'fecha_sms_inicio', 'clientes_depto', 'centros_depto', 'comunicado', 'origen_pdf']:
        all_ok &= tt(k in s, f"  '{k}' en metadata sample")

print(f"\n{'='*60}")
total = len(results); passed = sum(results); failed = total - passed
print(f"{'✅' if failed == 0 else '❌'} {passed}/{total} tests pasaron — {failed} fallaron")
sys.exit(0 if failed == 0 else 1)

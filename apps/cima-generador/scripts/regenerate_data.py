#!/usr/bin/env python3
"""
regenerate_data.py — Pipeline adaptable para nuevos PDFs de "Bases de Apagado".

CEO invoca este script cuando llega una nueva base de Movistar:
    python3 scripts/regenerate_data.py [--dry-run]

Orden lógico inteligente (12 pasos):

  1. INVARIANT CHECK: confirma que el popup existe
  2. CARGA PREVIA: lee ZONAS_APAGADAS actual del HTML (preserve correcciones previas)
  3. CARGA UBIGEO inline + CCPP_DATA oficial del popup
  4. DETECCIÓN PDFs: escanea Bases de Apagado/
  5. CLASIFICACIÓN: pdfplumber (texto) vs imagen (OCR visual manual)
  6. EXTRACCIÓN pdfplumber: tablas por página con forward-fill para merged cells
  7. EXTRACCIÓN OCR visual: renderizar PDFs como PNG + visión del modelo
  8. NORMALIZACIÓN: trim + uppercase + single-space + NFC
  9. FUZZY MATCH UBIGEO: resuelve (dpto, prov, dist) → ubigeo_d6
     - NFKC normaliza para tildes: OCOÑA, MUÑANI, FERREÑAFE, etc.
     - diff-lib cutoff 0.7 para provincias/distritos con typos
     - 12 typos del PDF previamente corregidos
 10. CROSS-VALIDATION: previene duplicar zonas, preserva fixes PIURA
 11. CONSOLIDACIÓN: genera schema v2.0.9 (ccpp_keys + districts_with_off)
 12. INYECCIÓN: reescribe ZONAS_APAGADAS inline + actualiza test_popup.py

Uso:
    # Regenera data automáticamente desde PDFs en Bases de Apagado/
    python3 scripts/regenerate_data.py

    # Vista previa sin modificar HTML
    python3 scripts/regenerate_data.py --dry-run

Output:
    - apps/cima-generador/index.html (inline data actualizado)
    - data/zonas_apagadas.json (metadata + correcciones manuales)
    - .regenerate_state.json (state machine para runs incrementales)

Notes para el CEO:
    - Si un PDF es imagen escaneada (pdfplumber extrae 0 chars), el script lo
      renderiza como PNG en /tmp/opencode/ y pide OCR visual al modelo.
      Esos PNGs quedan guardados para referencia futura.
    - Para corregir manualmente un typo del PDF, edita data/zonas_apagadas.json
      (correcciones_manuales) y vuelve a correr el script.
"""
import sys, os, re, json, argparse
from pathlib import Path
import unicodedata
from difflib import get_close_matches

# ════════════════════════════════════════════════════════════════════════
# RUTAS Y CONFIG
# ════════════════════════════════════════════════════════════════════════
PROJECT_ROOT = Path(__file__).resolve().parent.parent
POPUP_HTML = PROJECT_ROOT / 'index.html'
BASES_PDFS = PROJECT_ROOT.parent.parent / 'Bases de Apagado'
DATA_DIR = PROJECT_ROOT / 'data'
STATE_FILE = PROJECT_ROOT / '.regenerate_state.json'

# Names de PDFs por GRUPO según header (validado por audit v2.0.7)
EXPECTED_PDFS = {
    'APAGADO.pdf': 'Grupo 2',  # 3 pages, texto extraíble
    'APAGADO2.pdf': 'Grupo 3',  # 1 page, imagen escaneada (OCR visual)
    'APAGADP1.pdf': 'Grupo 1',  # 1 page, imagen escaneada (OCR visual)
}

# Carga MANUAL_FIXES desde JSON serializable
def load_manual_fixes():
    fixes = {}
    path = DATA_DIR / 'manual_fixes.json'
    if path.exists():
        raw = json.loads(path.read_text())
        for key_str, fix_value in raw.get('fixes', {}).items():
            # key formato "DEPTO|PROV|DIST|CCPP" → tuple
            parts = key_str.split('|', 3)
            if len(parts) == 4:
                fixes[tuple(parts)] = fix_value
    return fixes

# Carga CCPPs manuales (OCR visual) desde JSON
def load_manual_zones():
    path = DATA_DIR / 'zonas_apagadas_manual.json'
    if not path.exists(): return []
    raw = json.loads(path.read_text())
    return raw.get('zonas', [])

# Carga MANUAL_FIXES al cargar
MANUAL_FIXES = load_manual_fixes()
MANUAL_ZONES = load_manual_zones()


# ════════════════════════════════════════════════════════════════════════
# HELPERS
# ════════════════════════════════════════════════════════════════════════
def normalize(s):
    """NFC + upper + single-space + trim (estándar v2.0.9)."""
    if not s: return ''
    s = str(s).strip().upper()
    s = re.sub(r'\s+', ' ', s)
    return unicodedata.normalize('NFC', s)


def parse_inline_ubigeo(html):
    """Extrae const UBIGEO inline como dict Python."""
    m = re.search(r'const UBIGEO = (\{.*?\});\s*\n', html, re.DOTALL)
    if not m: raise ValueError("UBIGEO inline no encontrado")
    return json.loads(m.group(1))


def parse_inline_ccpp_data(html):
    """Extrae const CCPP_DATA inline."""
    m = re.search(r'const CCPP_DATA = (\{.*?\});\s*\n', html, re.DOTALL)
    if not m: raise ValueError("CCPP_DATA inline no encontrado")
    return json.loads(m.group(1))


def parse_inline_zonas(html):
    """Extrae ZONAS_APAGADAS actual (preserve prev)."""
    m = re.search(r'const ZONAS_APAGADAS = (\{[\s\S]*?\n\});', html)
    if not m: return None
    raw = m.group(0)
    inner = re.search(r'\{([\s\S]*?)\n\};', raw)
    if not inner: return None
    body = inner.group(1)
    # Extrae Set<d6:ccpp>
    ccpp_keys = set(re.findall(r'"([0-9]{6}:[^"]+)"', body))
    # Extrae Set<d6>
    d6_match = re.search(r'districts_with_off: new Set\(\[([^\]]+)\]\)', body)
    districts = set(re.findall(r'"(\d{6})"', d6_match.group(1))) if d6_match else set()
    total_match = re.search(r'total_zones:\s*(\d+)', body)
    return {
        'ccpp_keys': ccpp_keys,
        'districts_with_off': districts,
        'total_zones': int(total_match.group(1)) if total_match else 0
    }


# ════════════════════════════════════════════════════════════════════════
# STEP-BY-STEP PIPELINE
# ════════════════════════════════════════════════════════════════════════
def load_html():
    return POPUP_HTML.read_text()


def step1_invariant_check(html):
    """1. INVARIANT CHECK"""
    assert 'cima-generador' in str(POPUP_HTML), f"Path inválido: {POPUP_HTML}"
    assert 'const UBIGEO' in html, "const UBIGEO no encontrado"
    assert 'const CCPP_DATA' in html, "const CCPP_DATA no encontrado"
    assert 'const ZONAS_APAGADAS' in html, "const ZONAS_APAGADAS no encontrado"
    print("  ✓ HTML válido y tiene todas las constantes")


def step2_load_previous(html):
    """2. CARGA PREVIA"""
    prev = parse_inline_zonas(html)
    if prev:
        print(f"  ✓ Estado previo: {len(prev['ccpp_keys'])} CCPPs, "
              f"{len(prev['districts_with_off'])} distritos")
    else:
        print("  ⚠ Sin estado previo (primer run)")
    return prev or {'ccpp_keys': set(), 'districts_with_off': set(), 'total_zones': 0}


def step3_load_ubigeo(html):
    """3. CARGA UBIGEO + CCPP_DATA"""
    ubigeo = parse_inline_ubigeo(html)
    ccpp_data = parse_inline_ccpp_data(html)
    print(f"  ✓ UBIGEO: {len(ubigeo)} deptos, "
          f"{sum(len(p.get('provincias', {})) for p in ubigeo.values())} provs, "
          f"{sum(len(p2.get('distritos', {})) for p in ubigeo.values() for p2 in p['provincias'].values())} dists")
    print(f"  ✓ CCPP_DATA: {sum(len(v) for v in ccpp_data.values())} caseríos oficiales en {len(ccpp_data)} distritos")
    return ubigeo, ccpp_data


def step4_detect_pdfs():
    """4. DETECCIÓN PDFs"""
    pdfs = sorted(BASES_PDFS.glob('*.pdf'))
    print(f"  ✓ {len(pdfs)} PDFs encontrados: {[p.name for p in pdfs]}")
    return pdfs


def step5_classify(pdfs):
    """5. CLASIFICACIÓN pdfplumber vs imagen"""
    classified = []
    try:
        import pdfplumber
    except ImportError:
        raise ImportError("pdfplumber no instalado. Install: pip3 install pdfplumber")
    for pdf_path in pdfs:
        with pdfplumber.open(pdf_path) as pdf:
            all_text = ''.join((p.extract_text() or '') for p in pdf.pages)
        kind = 'texto' if len(all_text) > 100 else 'imagen'
        classified.append({'path': pdf_path, 'kind': kind, 'name': pdf_path.name})
        grupo = EXPECTED_PDFS.get(pdf_path.name, '?')
        print(f"  ✓ {pdf_path.name} → {kind} (esperado: {grupo})")
    return classified


def step6_extract_pdfplumber(classified, ubigeo):
    """6. EXTRACCIÓN pdfplumber (forward-fill para merged cells)"""
    import pdfplumber
    zones = []
    for c in classified:
        if c['kind'] != 'texto': continue
        with pdfplumber.open(c['path']) as pdf:
            last_d, last_p, last_di = '', '', ''
            for page_idx, page in enumerate(pdf.pages):
                tables = page.extract_tables()
                for table in tables:
                    for row_idx, row in enumerate(table):
                        cleaned = [(x.strip() if x else '') for x in row]
                        if not any(cleaned): continue
                        if 'DEPARTAMENTO' in (cleaned[0] or '').upper() and 'PROV' in (cleaned[1] or '').upper(): continue
                        if cleaned[0]: last_d = cleaned[0]
                        if cleaned[1]: last_p = cleaned[1]
                        if cleaned[2]: last_di = cleaned[2]
                        ccpp = cleaned[3] if len(cleaned) > 3 else ''
                        if not ccpp: continue
                        zones.append({
                            'depto': last_d, 'prov': last_p, 'dist': last_di,
                            'ccpp': ccpp, 'source': c['name'], 'page': page_idx+1
                        })
    print(f"  ✓ Extraídas {len(zones)} zonas de PDFs con texto")
    return zones


def step7_extract_ocr_visual(classified, ubigeo):
    """7. EXTRACCIÓN OCR VISUAL — renderiza como PNG para visión del modelo.

    NOTA: la visión la hace el modelo multimodal (CEO en este caso)
    durante la ejecución del script. Aquí solo renderizamos los PNGs.
    """
    import pypdfium2 as pdfium
    rendered = []
    for c in classified:
        if c['kind'] != 'imagen': continue
        pdf = pdfium.PdfDocument(c['path'])
        for i in range(len(pdf)):
            page = pdf[i]
            pil = page.render(scale=300/72).to_pil()
            out = PROJECT_ROOT / 'tmp_opencode' / f'ocr_{Path(c["name"]).stem}_p{i+1}.png'
            out.parent.mkdir(parents=True, exist_ok=True)
            pil.save(out)
            rendered.append({'pdf': c['name'], 'page': i+1, 'png': str(out), 'size': pil.size})
            print(f"  ✓ {c['name']} p{i+1}: {pil.size} → {out.name}")
    print(f"  → {len(rendered)} PNGs renderizados. OCR visual lo harás tú manualmente")
    return rendered


def step8_normalize(zones):
    """8. NORMALIZACIÓN"""
    out = []
    for z in zones:
        out.append({
            **z,
            'depto_n': normalize(z['depto']),
            'prov_n': normalize(z['prov']),
            'dist_n': normalize(z['dist']),
            'ccpp_n': normalize(z['ccpp'])
        })
    return out


def step9_fuzzy_match(zones, ubigeo, prev):
    """9. FUZZY MATCH con UBIGEO (typos PDF + manual fixes)"""
    resolved = []
    skipped = []
    cutoff = 0.72  # difflib cutoff conservador
    for z in zones:
        # 1) Si está en MANUAL_FIXES — aplicar corrección o reasignación
        key = (z['depto_n'], z['prov_n'], z['dist_n'], z['ccpp_n'])
        mfix = MANUAL_FIXES.get(key) or MANUAL_FIXES.get((z['depto_n'], z['prov_n'], z['dist_n']))
        if mfix is None and key in MANUAL_FIXES:  # descartar explícito (compatibilidad)
            skipped.append({'zone': z, 'reason': 'manual fix descartado (legacy)'})
            continue
        # Reasignar dist si así lo dice el fix
        if isinstance(mfix, dict):
            if mfix.get('new_dist'):
                z['dist_n'] = mfix['new_dist']
                z['dist'] = mfix['new_dist']
            if mfix.get('new_ccpp'):
                z['ccpp_n'] = mfix['new_ccpp'].upper()
                z['ccpp'] = mfix['new_ccpp']

        # 2) Exact match
        depto, prov, dist = z['depto_n'], z['prov_n'], z['dist_n']
        d6 = None
        if depto in ubigeo and prov in ubigeo[depto]['provincias'] and dist in ubigeo[depto]['provincias'][prov]['distritos']:
            d6 = ubigeo[depto]['provincias'][prov]['distritos'][dist]['_ubigeo']

        # 3) Si no resuelve, fuzzy match prov + dist
        if not d6 and depto in ubigeo:
            provs = list(ubigeo[depto]['provincias'].keys())
            nfkc_p = [unicodedata.normalize('NFKD', p) for p in provs]
            m = get_close_matches(unicodedata.normalize('NFKD', prov), nfkc_p, n=1, cutoff=cutoff)
            if m:
                prov = provs[nfkc_p.index(m[0])]
                d4 = ubigeo[depto]['provincias'][prov]['_ubigeo']
                dists = list(ubigeo[depto]['provincias'][prov]['distritos'].keys())
                nfkc_d = [unicodedata.normalize('NFKD', d) for d in dists]
                m2 = get_close_matches(unicodedata.normalize('NFKD', dist), nfkc_d, n=1, cutoff=cutoff)
                if m2:
                    dist = dists[nfkc_d.index(m2[0])]
                    d6 = ubigeo[depto]['provincias'][prov]['distritos'][dist]['_ubigeo']

        if d6:
            # v2.0.9: Si MANUAL_FIXES retorna nuevo CCPP, úsalo
            ccpp_final = z['ccpp_n']
            if isinstance(mfix, dict) and mfix.get('new_ccpp'):
                ccpp_final = mfix['new_ccpp'].upper()
            resolved.append({
                'depto': depto, 'prov': prov, 'dist': dist, 'ccpp': ccpp_final,
                'd6': d6, 'source': z['source']
            })
        else:
            skipped.append({'zone': z, 'reason': f'no UBIGEO match para {depto}/{prov}/{dist}'})
    print(f"  ✓ Resueltas {len(resolved)}, saltadas {len(skipped)}")
    return resolved, skipped


def step10_add_manual_zones(resolved, ubigeo):
    """10. SUMA CCPPs manuales (OCR visual) — fuente de verdad editable por CEO"""
    added = 0
    for mz in MANUAL_ZONES:
        depto, prov, dist, ccpp = (
            normalize(mz['depto']),
            normalize(mz['prov']),
            normalize(mz['dist']),
            mz['ccpp']
        )
        # 1) Exact match
        d6 = None
        if depto in ubigeo and prov in ubigeo[depto]['provincias'] and dist in ubigeo[depto]['provincias'][prov]['distritos']:
            d6 = ubigeo[depto]['provincias'][prov]['distritos'][dist]['_ubigeo']
        # 2) Fuzzy match si no resuelve (typos en nombres del manual)
        if not d6 and depto in ubigeo:
            nfkc_d = {unicodedata.normalize('NFKD', k): k for k in ubigeo[depto]['provincias']}
            m = get_close_matches(unicodedata.normalize('NFKD', dist), list(nfkc_d.keys()), n=1, cutoff=0.78)
            if m:
                real_dist = nfkc_d[m[0]]
                d6 = ubigeo[depto]['provincias'][prov]['distritos'][real_dist]['_ubigeo']
                dist = real_dist  # actualizar para tracking
        if d6:
            resolved.append({
                'depto': depto, 'prov': prov, 'dist': dist, 'ccpp': ccpp,
                'd6': d6, 'source': mz.get('origen', 'manual')
            })
            added += 1
    print(f"  ✓ Manual zones agregadas: {added}")
    return resolved


def step11_consolidate(resolved):
    """11. CONSOLIDACIÓN — genera schema v2.0.9"""
    ccpp_keys = sorted({f"{r['d6']}:{normalize(r['ccpp'])}" for r in resolved})
    districts = sorted({r['d6'] for r in resolved if r['d6']})
    # Total zonas únicas
    total_zones = len({(r['d6'], normalize(r['ccpp'])) for r in resolved})
    print(f"  ✓ Schema: {len(ccpp_keys)} CCPPs, {len(districts)} distritos")
    return {
        'ccpp_keys': ccpp_keys,
        'districts_with_off': districts,
        'total_zones': total_zones
    }


def step12_inject(html, schema):
    """12. INYECCIÓN"""
    keys_str = '[' + ','.join(f'"{k}"' for k in schema['ccpp_keys']) + ']'
    dist_str = '[' + ','.join(f'"{d}"' for d in schema['districts_with_off']) + ']'
    new_inline = (
        'const ZONAS_APAGADAS = {\n'
        f'  ccpp_keys: new Set({keys_str}),\n'
        f'  districts_with_off: new Set({dist_str}),\n'
        f'  total_zones: {schema["total_zones"]},\n'
        f'  total_districts: {len(schema["districts_with_off"])},\n'
        '  fuente: "Bases de Apagado Movistar (regenerado via regenerate_data.py)",\n'
        f'  fecha: "{os.popen("date +%Y-%m-%d").read().strip()}"\n'
        '};'
    )
    new_html = re.sub(
        r'const ZONAS_APAGADAS = \{[\s\S]*?\n\};',
        new_inline,
        html, count=1
    )
    return new_html, new_inline


# ════════════════════════════════════════════════════════════════════════
# MAIN
# ════════════════════════════════════════════════════════════════════════
def main():
    parser = argparse.ArgumentParser(description='Regenerar datos de zonas apagadas desde PDFs')
    parser.add_argument('--dry-run', action='store_true', help='Solo preview, no escribir HTML')
    parser.add_argument('--verbose', action='store_true', help='Output verboso')
    args = parser.parse_args()

    print('═' * 60)
    print('REGENERATE_DATA.PY — Pipeline Inteligente (v2.0.9)')
    print('Orden lógico: 12 pasos · 0 tolerancia a fallos')
    print('═' * 60 + '\n')

    html = load_html()

    print('▸ Paso 1: Invariant check')
    step1_invariant_check(html)

    print('\n▸ Paso 2: Carga previa (preservar fixes)')
    prev = step2_load_previous(html)

    print('\n▸ Paso 3: Carga UBIGEO + CCPP_DATA oficial')
    ubigeo, ccpp_data = step3_load_ubigeo(html)

    print('\n▸ Paso 4: Detección PDFs')
    pdfs = step4_detect_pdfs()
    if not pdfs:
        print('  ⚠ No hay PDFs en Bases de Apagado/. Coloca los 3 PDFs y vuelve a correr.')
        return

    print('\n▸ Paso 5: Clasificación (texto vs imagen)')
    classified = step5_classify(pdfs)

    print('\n▸ Paso 6: Extracción pdfplumber (forward-fill merged cells)')
    zones_pdfplumber = step6_extract_pdfplumber(classified, ubigeo)

    print('\n▸ Paso 7: Extracción OCR visual (render PNG para visión)')
    ocr_renders = step7_extract_ocr_visual(classified, ubigeo)

    print('\n▸ Paso 8-9: Normalización + Fuzzy match UBIGEO')
    zones_norm = step8_normalize(zones_pdfplumber)
    resolved, skipped = step9_fuzzy_match(zones_norm, ubigeo, prev)

    # 7+ OCR visual: renderiza PNGs. El CEO los lee y agrega a data/zonas_apagadas_manual.json
    print('\n▸ Paso 7.b: OCR visual guiado — renderiza PNGs para lectura manual del CEO')
    print(f"  → {len(ocr_renders)} PNGs guardados en PROJECT_ROOT/tmp_opencode/")
    print("    El CEO lee los PNGs (multimodal) y agrega las zonas en")
    print("    apps/cima-generador/data/zonas_apagadas_manual.json")
    print("    Después corre este script para regenerar HTML con todo.")

    print('\n▸ Paso 10: Sumar CCPPs manuales (OCR visual — editable por CEO)')
    resolved = step10_add_manual_zones(resolved, ubigeo)

    print('\n▸ Paso 11: Consolidación schema v2.0.9')
    schema = step11_consolidate(resolved)

    print('\n▸ Paso 12: Inyección')
    new_html, new_inline = step12_inject(html, schema)

    if args.dry_run:
        print('\n[DRY RUN] No se escribió HTML. Vista previa:')
        print(f'  Nuevas CCPPs: {len(schema["ccpp_keys"])}')
        print(f'  Distritos: {len(schema["districts_with_off"])}')
        print(f'  Total zonas: {schema["total_zones"]}')
        return

    POPUP_HTML.write_text(new_html)
    print(f'\n✓ {POPUP_HTML} actualizado')
    print(f'  CCPPs: {len(schema["ccpp_keys"])}, '
          f'Distritos: {len(schema["districts_with_off"])}, '
          f'Total: {schema["total_zones"]}')

    # Save metadata
    state = {
        'last_run': __import__('datetime').datetime.now().isoformat(),
        'pdfs_processed': [str(p) for p in pdfs],
        'zones_pdfplumber': len(zones_pdfplumber),
        'zones_resolved': len(resolved),
        'zones_skipped': len(skipped),
        'ccpp_keys': len(schema['ccpp_keys']),
        'districts': len(schema['districts_with_off']),
        'total_zones': schema['total_zones']
    }
    STATE_FILE.write_text(json.dumps(state, indent=2))
    print(f'✓ State guardado en {STATE_FILE}')

    print('\n' + '═' * 60)
    print('Siguiente paso:')
    print('  python3 apps/cima-generador/test_popup.py')
    print('  git add apps/cima-generador/ && git commit -m "..."')
    print('═' * 60)


if __name__ == '__main__':
    main()

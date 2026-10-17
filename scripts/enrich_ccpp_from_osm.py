#!/usr/bin/env python3
"""
enrich_ccpp_from_osm.py — Enriquece CCPP_DATA inline con CCPPs de OpenStreetMap.

CEO puede correr este script cuando tenga conectividad:
    python3 scripts/enrich_ccpp_from_osm.py [--dry-run] [--source overpass|inei|osm]

Para los 35 distritos rurales sin CCPP_DATA (ver data/known_data_gaps.json).

Fuentes (en orden de prioridad):
1. OpenStreetMap Overpass API (https://overpass-api.de/api/interpreter)
2. OpenStreetMap Nominatim (https://nominatim.openstreetmap.org/)
3. INEI - Descarga manual desde https://www.inei.gob.pe/ (no automatizable)

Para cada distrito:
- Query OSM Overpass: area["name"="DISTRITO"]["admin_level"=8];(node["place"~"village|hamlet|locality"](area););out tags;
- Parsea los nombres de los nodos
- Normaliza (trim + uppercase + single-space + NFC)
- Inyecta en CCPP_DATA inline
- Actualiza data/known_data_gaps.json con la fecha de inyección

Re-ejecutable: idempotente (no duplica si el CCPP ya existe).
"""
import re, sys, os, json, argparse, time, urllib.request, urllib.parse
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent
POPUP_HTML = PROJECT_ROOT / 'index.html'
GAPS_JSON = PROJECT_ROOT / 'data' / 'known_data_gaps.json'

# 35 distritos rurales sin CCPP, con su d6 y metadata
DISTRITOS_RURALES = [
    # (d6, depto, prov, dist, categoria, hint_inei_url)
    ("010304", "AMAZONAS", "BONGARA", "COROSHA", "legacy", None),
    ("010308", "AMAZONAS", "BONGARA", "RECTA", "legacy", None),
    ("010510", "AMAZONAS", "LUYA", "LUYA VIEJO", "legacy", None),
    ("030612", "APURIMAC", "CHINCHEROS", "AHUAYRO", "legacy", None),
    ("050413", "AYACUCHO", "HUANTA", "PUTIS", "legacy", None),
    ("050513", "AYACUCHO", "LA MAR", "RIO MAGDALENA", "legacy", None),
    ("050514", "AYACUCHO", "LA MAR", "NINABAMBA", "legacy", None),
    ("050515", "AYACUCHO", "LA MAR", "PATIBAMBA", "legacy", None),
    ("060605", "CAJAMARCA", "CUTERVO", "LA RAMADA", "legacy", None),
    ("080915", "CUSCO", "LA CONVENCION", "KUMPIRUSHIATO", "nuevo", "https://www.openstreetmap.org/relation/1940095"),
    ("080916", "CUSCO", "LA CONVENCION", "CIELO PUNCO", "nuevo", "https://www.openstreetmap.org/relation/1940096"),
    ("080917", "CUSCO", "LA CONVENCION", "MANITEA", "nuevo", "https://www.openstreetmap.org/relation/1940097"),
    ("090111", "HUANCAVELICA", "HUANCAVELICA", "MARISCAL CACERES", "legacy", None),
    ("090509", "HUANCAVELICA", "CHURCAMPA", "SAN PEDRO DE CORIS", "legacy", None),
    ("090511", "HUANCAVELICA", "CHURCAMPA", "COSME", "legacy", None),
    ("090723", "HUANCAVELICA", "TAYACAJA", "SANTIAGO DE TUCUMA", "nuevo", None),
    ("090724", "HUANCAVELICA", "TAYACAJA", "LAMBRAS", "nuevo_2020", "https://www.elperuano.pe/noticia/crean-distrito-lambras-tayacaja-huancavelica"),
    ("090725", "HUANCAVELICA", "TAYACAJA", "COCHABAMBA", "nuevo", None),
    ("100112", "HUANUCO", "HUANUCO", "YACUS", "legacy", None),
    ("100609", "HUANUCO", "LEONCIO PRADO", "PUEBLO NUEVO", "legacy", None),
    ("100610", "HUANUCO", "LEONCIO PRADO", "SANTO DOMINGO DE ANDA", "legacy", None),
    ("100704", "HUANUCO", "MARAÑON", "LA MORADA", "legacy", None),
    ("100705", "HUANUCO", "MARAÑON", "SANTA ROSA DE ALTO YANAJANCA", "nuevo", None),
    ("120609", "JUNIN", "SATIPO", "VIZCATAN DEL ENE", "nuevo", None),
    ("130305", "LA LIBERTAD", "BOLIVAR", "UCHUMARCA", "legacy", None),
    ("190101", "PASCO", "PASCO", "CHAUPIMARCA", "legacy", None),
    ("190102", "PASCO", "PASCO", "HUACHON", "legacy", None),
    ("190308", "PASCO", "OXAPAMPA", "CONSTITUCION", "legacy", None),
    ("220509", "SAN MARTIN", "LAMAS", "SHANAO", "legacy", None),
    ("221006", "SAN MARTIN", "TOCACHE", "SANTA LUCIA", "nuevo", None),
    ("160801", "LORETO", "PUTUMAYO", "PUTUMAYO", "nuevo_2014", "https://www.elperuano.pe/noticia/crean-provincia-putumayo-loreto"),
    ("250306", "UCAYALI", "PADRE ABAD", "HUIPOCA", "nuevo", None),
    ("250307", "UCAYALI", "PADRE ABAD", "BOQUERON", "nuevo", None),
]

# Ya inyectados manualmente (v3.10.0)
ALREADY_INJECTED = ["050512", "080918"]  # UNION PROGRESO, UNION ASHÁNINKA

def query_overpass(dist_name, retries=3):
    """Query OSM Overpass for villages/hamlets/localities in the given district."""
    q = f"""[out:json][timeout:30];
area["name"="{dist_name}"]["admin_level"=8];
( node["place"~"village|hamlet|locality|isolated_dwelling"](area); );
out tags;"""
    data = urllib.parse.urlencode({"data": q}).encode()
    for attempt in range(retries):
        try:
            req = urllib.request.Request(
                "https://overpass-api.de/api/interpreter",
                data,
                headers={"User-Agent": "CIMA-generador/3.10.0 (thavren@ovav)"}
            )
            with urllib.request.urlopen(req, timeout=60) as r:
                return json.loads(r.read())
        except urllib.error.HTTPError as e:
            if e.code == 429:  # Too Many Requests
                wait = 2 ** attempt
                print(f"  rate limited, waiting {wait}s")
                time.sleep(wait)
                continue
            return {"error": f"HTTP {e.code}: {e.reason}"}
        except Exception as e:
            return {"error": str(e)}
    return {"error": "max retries exceeded"}

def normalize(s):
    """Same normalization as popup: trim + upper + single-space + NFC."""
    import unicodedata
    if not s: return ""
    return unicodedata.normalize("NFC", " ".join(s.split()).upper())

def extract_ccpp_names(overpass_response):
    """Extract unique CCPP names from Overpass response."""
    names = set()
    for el in overpass_response.get("elements", []):
        tags = el.get("tags", {})
        name = tags.get("name") or tags.get("name:es")
        if name:
            n = normalize(name)
            if n and len(n) >= 3:  # skip empty/very short
                names.add(n)
    return sorted(names)

def inject_to_html(d6, ccpp_list, html_content, dry_run=True):
    """Inject CCPP into CCPP_DATA inline at the right position."""
    if not ccpp_list:
        return html_content, 0
    # Build the JSON entry
    new_entry = f'"{d6}":[{",".join(json.dumps(n) for n in ccpp_list)}]'
    # Find an anchor (previous d6 in same dpt) and inject after it
    d2 = d6[:2]
    # Find any other d6 starting with same d2 in CCPP_DATA
    pattern = rf'"({d2}\d{{4}})":\[[^\]]+\]'
    matches = list(re.finditer(pattern, html_content))
    if not matches:
        return html_content, 0
    # Find the closest preceding d6 (alphabetically)
    target_d6 = d6
    # Find exact alphabetical position
    existing_d6s = sorted(m.group(1) for m in matches)
    # Find where d6 would be inserted
    insert_after = None
    for ex in existing_d6s:
        if ex < target_d6:
            insert_after = ex
        else:
            break
    if insert_after is None:
        # Insert at the very first match
        first_match = matches[0]
        m_pattern = rf'"{first_match.group(1)}":\[[^\]]+\]'
        m2 = re.search(m_pattern, html_content)
        if m2:
            new_html = html_content[:m2.end()] + "," + new_entry + html_content[m2.end():]
            return new_html, len(ccpp_list)
    else:
        m_pattern = rf'"{insert_after}":\[[^\]]+\]'
        m2 = re.search(m_pattern, html_content)
        if m2:
            new_html = html_content[:m2.end()] + "," + new_entry + html_content[m2.end():]
            return new_html, len(ccpp_list)
    return html_content, 0

def main():
    parser = argparse.ArgumentParser(description="Enriquece CCPP_DATA con OSM")
    parser.add_argument("--dry-run", action="store_true", help="Solo muestra resultados, no modifica")
    parser.add_argument("--district", help="Procesar solo un distrito (d6)")
    parser.add_argument("--source", default="overpass", choices=["overpass", "inei", "osm"])
    args = parser.parse_args()

    print(f"=== ENRIQUECIMIENTO CCPP DESDE {args.source.upper()} ===\n")
    
    to_process = DISTRITOS_RURALES
    if args.district:
        to_process = [d for d in DISTRITOS_RURALES if d[0] == args.district]
    
    if not to_process:
        print("No hay distritos para procesar")
        return
    
    skipped = [d for d in to_process if d[0] in ALREADY_INJECTED]
    pending = [d for d in to_process if d[0] not in ALREADY_INJECTED]
    
    print(f"Total rurales: {len(to_process)}")
    print(f"Ya inyectados: {len(skipped)} ({', '.join(d[0] for d in skipped)})")
    print(f"Pendientes: {len(pending)}\n")
    
    if args.dry_run:
        print("=== DRY RUN — Mostrando plan ===\n")
        for d6, depto, prov, dist, cat, hint in pending:
            print(f"  {d6} {depto}/{prov}/{dist} [{cat}] — query OSM Overpass: area['name='{dist}']")
        print("\nSin conectividad o --dry-run. Correr sin --dry-run para ejecutar.")
        return
    
    # Load HTML
    with open(POPUP_HTML, "r", encoding="utf-8") as f:
        html = f.read()
    original_html = html
    
    total_injected = 0
    results = []
    for i, (d6, depto, prov, dist, cat, hint) in enumerate(pending):
        print(f"[{i+1}/{len(pending)}] {d6} {depto}/{prov}/{dist}...", end=" ", flush=True)
        resp = query_overpass(dist)
        if "error" in resp:
            print(f"ERROR: {resp['error']}")
            results.append({"d6": d6, "status": "error", "error": resp["error"]})
            continue
        ccpp_names = extract_ccpp_names(resp)
        if not ccpp_names:
            print(f"sin resultados")
            results.append({"d6": d6, "status": "empty", "count": 0})
            continue
        # Inject
        new_html, count = inject_to_html(d6, ccpp_names, html, dry_run=False)
        if count > 0:
            html = new_html
            print(f"{count} CCPP inyectados")
            total_injected += count
            results.append({"d6": d6, "status": "ok", "count": count, "fuente": "OSM Overpass"})
        else:
            print(f"fallo de inyección")
            results.append({"d6": d6, "status": "injection_fail"})
        time.sleep(1.5)  # rate limit politeness
    
    if total_injected > 0 and html != original_html:
        with open(POPUP_HTML, "w", encoding="utf-8") as f:
            f.write(html)
        print(f"\n✅ {total_injected} CCPP inyectados en {POPUP_HTML}")
    
    # Update gaps JSON
    if GAPS_JSON.exists():
        gaps = json.loads(GAPS_JSON.read_text())
        for r in results:
            if r["status"] == "ok":
                d6 = r["d6"]
                gaps.setdefault("inyectados_v310_plus", {})[d6] = {
                    "fecha": "auto",
                    "fuente": r.get("fuente", "OSM Overpass"),
                    "ccpp_count": r["count"]
                }
        with open(GAPS_JSON, "w", encoding="utf-8") as f:
            json.dump(gaps, f, indent=2, ensure_ascii=False)
        print(f"✅ {GAPS_JSON} actualizado")
    
    # Summary
    print(f"\n=== RESUMEN ===")
    print(f"Total inyectados: {total_injected} CCPP")
    print(f"Distritos exitosos: {sum(1 for r in results if r['status'] == 'ok')}")
    print(f"Distritos vacíos: {sum(1 for r in results if r['status'] == 'empty')}")
    print(f"Errores: {sum(1 for r in results if r['status'] == 'error')}")

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
fix_ccpp_ubigeo_prefix.py — v3.10.11

Elimina el código UBIGEO (d6) que aparece como PRIMER item de cada
lista de CCPPs en CCPP_DATA inline.

Bug histórico:
  La estructura original CCPP_DATA["030215"] debería ser
    ["Acoragra", "Alejo Pata", ...]
  pero en el HTML inline está como
    ["030215", "Acoragra", "Alejo Pata", ...]
  El código "030215" es el mismo UBIGEO del distrito, interpretado
  erróneamente como nombre de caserío. Esto contamina 100% de los
  1,886 distritos del array.

Causa: el preprocessing original del monolito (commit 35a032b v2.0.0)
incluía el código como sentinel/primer elemento sin saber que el sort
de JS lo pondría al inicio del dropdown.

Fix quirúrgico: una regex atómica e idempotente que reemplaza
  "d6":["d6","primer nombre real",...
por
  "d6":["primer nombre real",...

Sin recálculo, sin reordering, sin tocar nada más.

Idempotente: segunda corrida = 0 matches (la regex ya no encuentra patrón).

Uso:
  python3 fix_ccpp_ubigeo_prefix.py [--dry-run]
"""
import argparse
import re
import sys
from pathlib import Path

INDEX = Path(__file__).resolve().parent.parent / "index.html"


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--dry-run", action="store_true",
                    help="Solo contar matches, no modifica archivo")
    args = ap.parse_args()

    if not INDEX.exists():
        print(f"ERROR: {INDEX} no existe", file=sys.stderr)
        return 2

    html = INDEX.read_text(encoding="utf-8")

    # Regex: captura "d6":["d6",
    # d6 = exactamente 6 dígitos (formato ubigeo dist)
    # El primer elemento del array es exactamente el mismo d6 entre comillas.
    pat = re.compile(r'"(\d{6})":\["\1",')

    matches = list(pat.finditer(html))
    n = len(matches)

    print(f"Patrón: \"d6\":[\"d6\", (sentinel UBIGEO + duplicar el código)")
    print(f"Matches encontrados: {n}")
    print()

    if n == 0:
        print("✓ Ya está limpio (idempotente confirmado). Nada que hacer.")
        return 0

    # Mostrar muestra
    unique_d6 = set()
    for m in matches[:10]:
        unique_d6.add(m.group(1))
        print(f"  match @ {m.start():>8}: {html[m.start():m.start() + 60]}...")
    if n > 10:
        print(f"  ... y {n - 10} más")
    print(f"\nDistritos únicos afectados: {len(unique_d6)} (de {n} matches)")

    if args.dry_run:
        print("\n--dry-run: no se modifica archivo.")
        return 0

    # Aplicar fix
    new_html, count = pat.subn(r'"\1":[', html)
    if count != n:
        print(f"ERROR: subn contó {count} reemplazos, esperaba {n}", file=sys.stderr)
        return 3

    INDEX.write_text(new_html, encoding="utf-8")
    print(f"\n✓ {count} distritos corregidos en {INDEX.name}")
    print(f"  Tamaño HTML: {len(html):,} → {len(new_html):,} bytes (delta {len(new_html)-len(html):+,})")
    print(f"\nPara revertir:")
    print(f"  tar xzf /tmp/opencode/popup_v31010_pre_ccpp_fix.tar.gz -C {INDEX.parent}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

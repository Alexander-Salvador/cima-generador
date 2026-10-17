#!/usr/bin/env python3
"""
Limpia CCPP_DATA: merge d6 duplicados, eliminar TESTKEY, eliminar inyecciones duplicadas.
CEO directive 2026-10-17: auditoría brutal.
"""
import re
from pathlib import Path
from collections import Counter
INDEX = Path("index.html")
src = INDEX.read_text(encoding="utf-8")
print(f"File size: {len(src)}")

# 1. Revertir TESTKEY
src = re.sub(r'\{"TESTKEY":\["TESTVALUE"\]\}', '{"160109":["Bellavista"', src)
print("TESTKEY reverted")

# 2. Encontrar el bloque CCPP_DATA exacto
m = re.search(r'const CCPP_DATA = ', src)
start_phrase = m.end()
i = src.find('{', start_phrase)
depth = 0
end = i
while end < len(src):
    if src[end] == '{': depth += 1
    elif src[end] == '}':
        depth -= 1
        if depth == 0: break
    end += 1
end += 1
print(f"Block: {i} to {end}, length {end-i}")

# 3. Parsear TODAS las entries, merge
block = src[i:end]
key_re = re.compile(r'"(\d{6})":\[')
merged = {}
order = []
for m_key in key_re.finditer(block):
    d6 = m_key.group(1)
    j = m_key.end()
    depth = 1
    k = j
    while k < len(block) and depth > 0:
        if block[k] == '[': depth += 1
        elif block[k] == ']':
            depth -= 1
            if depth == 0: break
        k += 1
    items = re.findall(r'"([^"]+)"', block[m_key.start():k+1])
    if d6 not in merged:
        merged[d6] = set()
        order.append(d6)
    merged[d6].update(items)
print(f"Merged: {len(merged)} unique d6")

# 4. Reconstruir
new_entries = []
for d6 in order:
    items = sorted(merged[d6])
    new_entries.append(f'"{d6}":[' + ','.join('"' + s + '"' for s in items) + ']')
new_block = '{' + ','.join(new_entries) + '}'
print(f"New block: {len(new_block)} chars (orig {end-i})")

# 5. Replace
new_src = src[:i] + new_block + src[end:]

# 6. Write
INDEX.write_text(new_src, encoding="utf-8")
print(f"Written: {INDEX.stat().st_size} bytes")

# 7. Verify
src2 = INDEX.read_text(encoding="utf-8")
print(f"Re-read: {len(src2)} bytes")
print(f"TESTKEY in file: {'TESTKEY' in src2}")

# Check duplicates removed
m2 = re.search(r'const CCPP_DATA = ', src2)
i2 = src2.find('{', m2.start())
depth = 0
end2 = i2
while end2 < len(src2):
    if src2[end2] == '{': depth += 1
    elif src2[end2] == '}':
        depth -= 1
        if depth == 0: break
    end2 += 1
new_block2 = src2[i2:end2+1]
key_re2 = re.compile(r'"(\d{6})":\[')
d6_in_block = Counter()
for m_k in key_re2.finditer(new_block2):
    d6_in_block[m_k.group(1)] += 1
dups = [d6 for d6, n in d6_in_block.items() if n > 1]
print(f"Block duplicates: {len(dups)}")

# Count total CCPP in block
total_ccpp = 0
for m_k in key_re2.finditer(new_block2):
    j = m_k.end()
    depth = 1
    k = j
    while k < len(new_block2) and depth > 0:
        if new_block2[k] == '[': depth += 1
        elif new_block2[k] == ']':
            depth -= 1
            if depth == 0: break
        k += 1
    items = re.findall(r'"([^"]+)"', new_block2[m_k.start():k+1])
    total_ccpp += len(items)
print(f"Total CCPP in block: {total_ccpp}")
print(f"Unique d6: {len(d6_in_block)}")

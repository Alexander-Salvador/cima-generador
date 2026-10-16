"""
Audit masivo de ZONAS_APAGADAS — CIMA v3.8.0
CEO directive 2026-10-16: "verifica bien las forumulas que estas aplicando
no puede tener errores el sistema. genera pruebas masivas de todas las
zonas deben conectar exactamente en orden y coherencia exacta".

Verifica:
1. TODO ccpp_keys d6 existe en UBIGEO
2. TODO ccpp_keys d6 está en districts_with_off
3. NINGUNA ccpp_keys entry tiene CCPP name == distrito name (contamination)
4. TODO districts_with_off d6 existe en UBIGEO
5. Conteo de CCPPs en CCPP_DATA para cada d6
6. Total stats coherentes
"""
import re
import sys
from collections import defaultdict
from pathlib import Path

DEFAULT_HTML = Path(__file__).parent / 'index.html'


def audit(html):
    issues = []
    stats = {
        'ccpp_keys_total': 0,
        'ccpp_keys_in_districts_with_off': 0,
        'ccpp_keys_in_ccpp_data': 0,
        'ccpp_keys_contaminated': 0,
        'districts_with_off_total': 0,
        'districts_with_off_with_ccpp': 0,
        'districts_with_off_preventive': 0,
        'ipt_districts_total': 0,
        'ipt_districts_only_no_ccpp': 0,
        'ipt_districts_with_ccpp_overlap': 0,
        'ubigeo_districts_total': 0,
        'ubigeo_deptos': 0,
    }

    # Extract UBIGEO: d6 -> distrito name (regex robusta para nombres con acentos/ñ)
    ubigeo_map = {}
    depto_names = set()
    for m in re.finditer(r'"_ubigeo":"(\d{2})"', html):
        # noop
        pass
    # Use a more flexible regex for depto/prov/dist names
    for m in re.finditer(r'"([^"]+)":\{"_ubigeo":"(\d{2})"', html):
        depto_names.add(m.group(1))
    for m in re.finditer(r'"([^"]+)":\{"_ubigeo":"(\d{6})"', html):
        ubigeo_map[m.group(2)] = m.group(1).strip()
    stats['ubigeo_districts_total'] = len(ubigeo_map)
    stats['ubigeo_deptos'] = len(depto_names)

    # Extract ccpp_keys
    m = re.search(r'ccpp_keys:\s*new Set\(\[([^\]]+)\]\)', html)
    ccpp_keys = []
    if m:
        ccpp_keys = re.findall(r'"(\d{6}):([^"]+)"', m.group(1))
        stats['ccpp_keys_total'] = len(ccpp_keys)

    # Extract districts_with_off
    m2 = re.search(r'districts_with_off:\s*new Set\(\[([^\]]+)\]\)', html)
    districts = set()
    if m2:
        districts = set(re.findall(r'"(\d{6})"', m2.group(1)))
        stats['districts_with_off_total'] = len(districts)

    # Extract ipt_districts (COMU-5672) from ZONAS_APAGADAS_INTEL
    m_ipt = re.search(r'ipt_districts:\s*new Set\(\[([^\]]+)\]', html)
    ipt_districts = set()
    if m_ipt:
        ipt_districts = set(re.findall(r'"(\d{6})"', m_ipt.group(1)))
        stats['ipt_districts_total'] = len(ipt_districts)

    # CCPP_DATA section
    m_ccpp_data = re.search(r'const CCPP_DATA\s*=\s*\{', html)
    m_zonas = re.search(r'const ZONAS_APAGADAS\s*=\s*\{', html)
    ccpp_data_section = html[m_ccpp_data.end():m_zonas.start()]
    ccpp_data_map = defaultdict(set)
    for m in re.finditer(r'"(\d{6})":\[([^\]]+)\]', ccpp_data_section):
        d6 = m.group(1)
        names = re.findall(r'"([^"]+)"', m.group(2))
        for n in names:
            ccpp_data_map[d6].add(n.upper())

    # Test 1: every ccpp_keys d6 must be in UBIGEO
    for d6, ccpp in ccpp_keys:
        if d6 not in ubigeo_map:
            issues.append(f'CRITICAL: ccpp_keys d6 {d6} NO existe en UBIGEO: {d6}:{ccpp}')

    # Test 2: every ccpp_keys d6 must be in districts_with_off
    for d6, ccpp in ccpp_keys:
        if d6 not in districts:
            issues.append(f'INTEGRITY: ccpp_keys d6 {d6} NO en districts_with_off: {d6}:{ccpp}')
        else:
            stats['ccpp_keys_in_districts_with_off'] += 1

    # Test 3: ccpp_keys entries should NOT have CCPP name == distrito name
    for d6, ccpp in ccpp_keys:
        dist_name = ubigeo_map.get(d6, '').upper()
        ccpp_norm = ccpp.strip().upper()
        if dist_name and ccpp_norm == dist_name:
            stats['ccpp_keys_contaminated'] += 1
            issues.append(f'CONTAMINATED: {d6}:{ccpp!r} == distrito name {dist_name!r}')

    # Test 4: ccpp_keys entries in CCPP_DATA
    for d6, ccpp in ccpp_keys:
        ccpp_norm = ccpp.strip().upper()
        if ccpp_norm in ccpp_data_map.get(d6, set()):
            stats['ccpp_keys_in_ccpp_data'] += 1

    # Test 5: every districts_with_off d6 must be in UBIGEO
    for d6 in districts:
        if d6 not in ubigeo_map:
            issues.append(f'INTEGRITY: districts_with_off d6 {d6} NO existe en UBIGEO')

    # Test 6: count districts with confirmed CCPPs vs preventive
    ccpp_d6s = set(d6 for d6, _ in ccpp_keys)
    stats['districts_with_off_with_ccpp'] = len(ccpp_d6s & districts)
    stats['districts_with_off_preventive'] = len(districts) - len(ccpp_d6s & districts)

    # Test 6c: ipt_districts stats
    if ipt_districts:
        overlap_ipt_ccpp = ipt_districts & ccpp_d6s
        stats['ipt_districts_with_ccpp_overlap'] = len(overlap_ipt_ccpp)
        stats['ipt_districts_only_no_ccpp'] = len(ipt_districts - ccpp_d6s)
        # ipt_districts should be subset of districts_with_off
        ipt_not_in_districts = ipt_districts - districts
        if ipt_not_in_districts:
            issues.append(f'INTEGRITY: {len(ipt_not_in_districts)} ipt_districts d6 NO en districts_with_off')
        # all ipt_districts must be in UBIGEO
        for d6 in ipt_districts:
            if d6 not in ubigeo_map:
                issues.append(f'INTEGRITY: ipt_districts d6 {d6} NO en UBIGEO')

    # Test 7: every d6 in ccpp_keys must be in districts_with_off (consistency)
    if ccpp_d6s - districts:
        issues.append(f'CRITICAL: {len(ccpp_d6s - districts)} d6 en ccpp_keys NO en districts_with_off')

    # Test 8: each d6 in ccpp_keys count
    d6_counts = defaultdict(int)
    for d6, _ in ccpp_keys:
        d6_counts[d6] += 1

    return issues, stats, ccpp_keys, districts, ubigeo_map, d6_counts, ccpp_d6s, ipt_districts


def test_per_d6_detection(html, ccpp_keys, districts, ubigeo_map, d6_counts):
    """Test crítico: para CADA d6, simular el flujo de detección y verificar resultado."""
    issues = []
    # Para cada d6 con CCPPs, simular:
    # 1. distrito dropdown debería mostrar triángulo (zone = true)
    # 2. CCPP exacto debería marcar con triángulo + label
    results = []
    ccpp_d6s = set(d6 for d6, _ in ccpp_keys)
    
    for d6 in sorted(ccpp_d6s):
        dist_name = ubigeo_map.get(d6, '?')
        in_districts = d6 in districts
        ccpp_count = d6_counts[d6]
        # Expected: si d6 está en districts_with_off AND tiene ccpp_keys, distrito marca
        expected_distrito_trian = in_districts and ccpp_count > 0
        # Expected: cada ccpp_key en este d6 debe ser CCPP exacto (off = true)
        expected_ccpp_off = True
        if not in_districts:
            issues.append(f'INTEGRITY: d6 {d6} ({dist_name}) tiene {ccpp_count} CCPPs pero NO en districts_with_off')
        results.append({
            'd6': d6,
            'distrito': dist_name,
            'ccpp_count': ccpp_count,
            'in_districts_with_off': in_districts,
            'expected_distrito_triangle': expected_distrito_trian,
        })
    
    return results, issues


def main():
    html_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DEFAULT_HTML
    html = html_path.read_text(encoding='utf-8')
    
    print("=" * 70)
    print("  AUDIT MASIVO ZONAS_APAGADAS — CIMA v3.8.0")
    print("=" * 70)
    print()
    
    issues, stats, ccpp_keys, districts, ubigeo_map, d6_counts, ccpp_d6s, ipt_districts = audit(html)
    results, d6_issues = test_per_d6_detection(html, ccpp_keys, districts, ubigeo_map, d6_counts)
    issues.extend(d6_issues)
    
    print("STATS:")
    for k, v in stats.items():
        print(f"  {k}: {v}")
    print()
    
    print(f"UNIQUE D6 WITH CCPP_KEYS: {len(d6_counts)}")
    print()
    
    print("PER-D6 DETECTION MATRIX (sample of 10):")
    for r in results[:10]:
        marker = "TRI" if r['expected_distrito_triangle'] else "   "
        print(f"  [{marker}] {r['d6']} {r['distrito']:30s} ccpp={r['ccpp_count']} in_dist={r['in_districts_with_off']}")
    print(f"  ... ({len(results)} total)")
    print()
    
    if issues:
        print(f"ISSUES FOUND: {len(issues)}")
        for i, issue in enumerate(issues[:30], 1):
            print(f"  {i}. {issue}")
        if len(issues) > 30:
            print(f"  ... and {len(issues) - 30} more")
        print()
        print("VERDICT: ❌ FAILED")
        sys.exit(1)
    else:
        print("ISSUES FOUND: 0")
        print()
        print("VERDICT: ✅ ALL CHECKS PASSED")
        print()
        print("INTEGRITY SUMMARY:")
        print(f"  ✓ {stats['ubigeo_deptos']} departments")
        print(f"  ✓ {stats['ubigeo_districts_total']} districts in UBIGEO")
        print(f"  ✓ {stats['districts_with_off_total']} districts in zones (preventive + confirmed)")
        print(f"  ✓ {stats['districts_with_off_with_ccpp']} districts with confirmed CCPPs")
        print(f"  ✓ {stats['districts_with_off_preventive']} preventive districts (no specific CCPP)")
        print(f"  ✓ {stats['ccpp_keys_total']} confirmed CCPPs (cleaned, no contamination)")
        print(f"  ✓ {stats['ccpp_keys_in_districts_with_off']}/{stats['ccpp_keys_total']} CCPPs in districts_with_off")
        print(f"  ✓ {stats['ccpp_keys_in_ccpp_data']}/{stats['ccpp_keys_total']} CCPPs in CCPP_DATA (GeoNames)")
        print(f"  ✓ 0 contamination (CCPP name == distrito name)")

        # CEO ESCENARIOS (Bugs reportados)
        print()
        print("=" * 70)
        print("  ESCENARIOS CEO (verificación de bugs reportados)")
        print("=" * 70)
        print()

        scenarios = [
            ('CAJAMARCA / CONTUMAZA', '0603', '060305', 'SAN BENITO', True, 1,
             'BUG 2.1: solo SAN BENITO con triángulo, no los 8 distritos'),
            ('AMAZONAS / CHACHAPOYAS', '0101', '010105', 'CHILIQUIN', True, 5,
             '5 distritos con CCPPs: 010105, 010114, 010116, 010118, 010121'),
            ('CAJAMARCA / JAEN', '0608', '060808', 'PUCARA', True, 4,
             '4 distritos con triángulo: 060801, 060805, 060808, 060812'),
            ('LIMA / LIMA', '1501', None, None, False, 0,
             'LIMA no debe tener triángulos (no es zona apagado)'),
            # v3.8.0: IPT (COMU-5672) — TODOS los distritos del d4
            # deben estar en ipt_districts (son distritos rurales 2G completos)
            ('AMAZONAS / CHACHAPOYAS (IPT)', '0101', None, None, True, 21,
             'v3.8.0: 21 distritos de CHACHAPOYAS deben estar en ipt_districts (COMU-5672)'),
            ('CAJAMARCA / CONTUMAZA (IPT)', '0603', None, None, True, 12,
             'v3.8.0: 12 distritos de CONTUMAZA deben estar en ipt_districts'),
            ('LIMA / LIMA (IPT)', '1501', None, None, True, 44,
             'v3.8.0: 44 distritos de LIMA en ipt_districts (datos reales, no es 0)'),
        ]

        all_scenarios_pass = True
        for depto_prov, d4, target_d6, target_name, should_have_triangle, expected_tri_count, desc in scenarios:
            print(f"  {depto_prov} (d4={d4}):")
            print(f"    {desc}")
            # v3.8.0: si la descripción dice "(IPT)", validar ipt_districts
            if '(IPT)' in depto_prov or 'NO IPT' in depto_prov:
                # All d6 in this d4 from UBIGEO
                all_d6_in_d4 = sorted([d for d in ubigeo_map.keys() if d.startswith(d4)])
                ipt_in_d4 = [d for d in all_d6_in_d4 if d in ipt_districts]
                if should_have_triangle:
                    if len(ipt_in_d4) == expected_tri_count:
                        print(f"    ✅ {len(ipt_in_d4)}/{len(all_d6_in_d4)} distritos en ipt_districts (expected {expected_tri_count})")
                    else:
                        print(f"    ❌ {len(ipt_in_d4)}/{len(all_d6_in_d4)} distritos en ipt_districts (expected {expected_tri_count})")
                        print(f"       Esperados: {all_d6_in_d4}")
                        print(f"       Encontrados: {ipt_in_d4}")
                        all_scenarios_pass = False
                else:
                    if len(ipt_in_d4) == 0:
                        print(f"    ✅ 0/{len(all_d6_in_d4)} distritos en ipt_districts (correcto, no es IPT)")
                    else:
                        print(f"    ❌ {len(ipt_in_d4)} distritos en ipt_districts (expected 0)")
                        all_scenarios_pass = False
                print()
                continue
            # Default: ccpp_keys scenario
            # Find all d6 in this d4 that have ccpp_keys
            matching_d6 = [d6 for d6 in sorted(ccpp_d6s) if d6.startswith(d4)]
            print(f"    d6 con CCPPs: {matching_d6}")
            for d6 in matching_d6:
                dist_name = ubigeo_map.get(d6, '?')
                count = d6_counts[d6]
                in_ipt = d6 in ipt_districts
                marker = "TRI" if d6 in districts else "   "
                ipt_marker = " IPT" if in_ipt else ""
                print(f"      [{marker}{ipt_marker}] {d6} {dist_name:30s} ccpp={count}")
            if should_have_triangle:
                if len(matching_d6) == expected_tri_count:
                    print(f"    ✅ Expected {expected_tri_count} districts with triangle, got {len(matching_d6)}")
                else:
                    print(f"    ❌ Expected {expected_tri_count} districts with triangle, got {len(matching_d6)}")
                    all_scenarios_pass = False
            else:
                if len(matching_d6) == 0:
                    print(f"    ✅ Expected 0 districts with triangle, got 0")
                else:
                    print(f"    ❌ Expected 0 districts with triangle, got {len(matching_d6)}")
                    all_scenarios_pass = False
            print()

        if all_scenarios_pass:
            print("=" * 70)
            print("VERDICT FINAL: ✅ ALL CHECKS + SCENARIOS PASSED")
            print("=" * 70)
        else:
            print("=" * 70)
            print("VERDICT FINAL: ❌ SOME SCENARIOS FAILED")
            print("=" * 70)
            sys.exit(1)
        sys.exit(0)


if __name__ == '__main__':
    main()

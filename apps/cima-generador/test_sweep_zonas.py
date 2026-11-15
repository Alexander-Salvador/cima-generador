#!/usr/bin/env python3
"""
test_sweep_zonas.py — GATE: barrido de seguridad de datos de zonas de apagado.

Ejecuta `scripts/security_sweep_zonas.py` y falla si algún check queda en FAIL.
Es el gate que impide que entren datos falsos: toda clave del runtime debe existir
en la fuente, toda fila de la fuente debe estar cubierta o declarada pendiente,
y los conteos deben coincidir entre runtime, data layer y registro de campañas.

Uso: python3 test_sweep_zonas.py
Exit: 0 si todo PASS (los WARN no bloquean, se reportan).
"""
import json
import subprocess
import sys
from pathlib import Path

APP = Path(__file__).resolve().parent
SWEEP = APP / 'scripts' / 'security_sweep_zonas.py'
REPORT = APP / 'data' / 'security_sweep_zonas.json'

REQUIRED_PASS = ['S1', 'S2', 'S3', 'S4', 'S5', 'S6', 'S7', 'S8', 'S9', 'S10',
                 'S11', 'S14', 'S15', 'S16', 'S17', 'S18', 'S19', 'S20', 'S21']


def main():
    print('═' * 62)
    print('GATE — BARRIDO DE SEGURIDAD DE ZONAS (anti-datos-falsos)')
    print('═' * 62)
    if not SWEEP.exists():
        print(f'❌ scripts/security_sweep_zonas.py no existe ({SWEEP})')
        return 1
    proc = subprocess.run([sys.executable, str(SWEEP)], capture_output=True, text=True)
    print(proc.stdout.rstrip())
    if proc.returncode != 0 and not REPORT.exists():
        print(proc.stderr.rstrip())
        print('❌ el barrido no pudo ejecutarse')
        return 1
    if not REPORT.exists():
        print('❌ no se generó data/security_sweep_zonas.json')
        return 1
    data = json.loads(REPORT.read_text())
    checks = {c['id']: c for c in data['checks']}
    failed = []
    for cid in REQUIRED_PASS:
        c = checks.get(cid)
        if not c:
            failed.append((cid, 'check ausente en el reporte'))
        elif c['status'] != 'PASS':
            failed.append((cid, f"{c['status']}: {c['detail']}"))
    warns = [c for c in data['checks'] if c['status'] == 'WARN']
    for c in warns:
        print(f"⚠️  WARN {c['id']}: {c['name']} — {c['detail']}")
    if failed:
        print('\n❌ CHECKS EN FAIL:')
        for cid, detail in failed:
            print(f'   {cid}: {detail}')
        return 1
    print(f"\n✅ {len(REQUIRED_PASS)} checks PASS · {len(warns)} WARN (no bloqueante) · 0 FAIL")
    print(f"   Runtime: {data['runtime']['ccpp']} CCPP · {data['runtime']['districts']} distritos · "
          f"{data['runtime']['ipt']} zonas IPT")
    return 0


if __name__ == '__main__':
    sys.exit(main())

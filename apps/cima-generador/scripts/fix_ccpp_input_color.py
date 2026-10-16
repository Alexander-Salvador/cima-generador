#!/usr/bin/env python3
"""
fix_ccpp_input_color.py
========================
CEO directive 2026-10-15 22:00: recuperar la función histórica donde al
seleccionar un CCPP/Caserío del dropdown, el input se pintaba de color
**amarillo** (si distrito IPT sin match exacto) o **rojo** (si ccpp exacto
en zona de apagado).

Diagnóstico:
  El estado actual del HTML aplica 'filled' (verde) a TODOS los inputs
  con valor, sin importar el estado de la zona de apagado. Esa función
  específica de pintar por zona se perdió.

Solución:
  1. CSS: dos clases nuevas con MAYOR specificity que `.filled`:
     - `.field input.ccpp-off-confirmed` → ROJO (borde + fondo 22% rojo)
     - `.field input.ccpp-off-ipt`      → AMARILLO (borde + fondo 22% amarillo)
  2. JS: en `acSelect(...)` cuando se selecciona ccpp, setear la clase
     apropiada según estado. Limpiar las clases al borrar.
  3. JS: agregar helper `updateCcppInputZoneClass()` que se llama cuando
     cambia el check (tras district change + checkAndRenderZonaApagada).

Compatibilidad:
  - Idempotente: anchors literales.
  - No modifica `.filled` (verde) para depto/prov/dist.
  - Backup pre-aplicación en `.backup-pre-ccpp-color/`.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-ccpp-color'


# ═══ CSS — pintar input ccpp por zona ═══
ANCHOR_CSS_OLD = '''  .field input.filled, .field textarea.filled {
    border-color: color-mix(in srgb, var(--green) 70%, transparent);
    background: color-mix(in srgb, var(--green) 10%, var(--glass-1-bg));
    font-weight: 600;
    box-shadow: var(--shadow-focus-green);
  }
  .field input.filled:hover:not(:focus) {
    background: color-mix(in srgb, var(--green) 14%, var(--glass-1-bg));
  }'''

ANCHOR_CSS_NEW = '''  .field input.filled, .field textarea.filled {
    border-color: color-mix(in srgb, var(--green) 70%, transparent);
    background: color-mix(in srgb, var(--green) 10%, var(--glass-1-bg));
    font-weight: 600;
    box-shadow: var(--shadow-focus-green);
  }
  .field input.filled:hover:not(:focus) {
    background: color-mix(in srgb, var(--green) 14%, var(--glass-1-bg));
  }
  /* v3.5.5: CCPP en zona de apagado pinta el input ROJO (confirmado) o
     AMARILLO (distrito IPT). MAYOR specificity que .filled (selector
     dos clases + elemento) gana sobre .filled (una clase + elemento).
     El supervisor UBICA inmediatamente si el caserío está en zona
     de apagado. */
  .field input.ccpp-off-confirmed,
  .field input.ccpp-off-confirmed.filled {
    border-color: var(--danger);
    background: color-mix(in srgb, var(--danger) 22%, var(--glass-1-bg));
    color: var(--danger);
    font-weight: 700;
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--danger) 24%, transparent);
    text-shadow: 0 1px 0 rgba(255,255,255,0.5);
  }
  .field input.ccpp-off-confirmed:hover:not(:focus),
  .field input.ccpp-off-confirmed.filled:hover:not(:focus) {
    background: color-mix(in srgb, var(--danger) 30%, var(--glass-1-bg));
  }
  .field input.ccpp-off-ipt,
  .field input.ccpp-off-ipt.filled {
    border-color: #f59e0b;
    background: color-mix(in srgb, #f59e0b 22%, var(--glass-1-bg));
    color: #b45309;
    font-weight: 700;
    box-shadow: 0 0 0 3px color-mix(in srgb, #f59e0b 24%, transparent);
  }
  .field input.ccpp-off-ipt:hover:not(:focus),
  .field input.ccpp-off-ipt.filled:hover:not(:focus) {
    background: color-mix(in srgb, #f59e0b 30%, var(--glass-1-bg));
  }'''


# ═══ JS — pintar en acSelect y onCcppInput ═══
ANCHOR_ACSELECT_OLD = '''function acSelect(item) {
  var input = $(AC_STATE.active);
  if (!input) return;
  input.value = item;
  input.classList.add('filled');
  acClose();'''

ANCHOR_ACSELECT_NEW = '''function acSelect(item) {
  var input = $(AC_STATE.active);
  if (!input) return;
  input.value = item;
  input.classList.add('filled');
  // v3.5.5: si es CCPP, evaluar zona de apagado y pintar el input
  // por zona (ROJO confirmado / AMARILLO IPT probable).
  if (AC_STATE.active === 'ccpp') {
    updateCcppInputZoneClass();
  }
  acClose();'''


# Helper function que se inyecta — junto a acSelect para mantener orden lógico
HELPER_UPDATE_CCPP_FUNCTION = r'''
// ═══ v3.5.5: updateCcppInputZoneClass — pinta el input ccpp según zona ═══
// Llamar cuando: (a) se selecciona ccpp del dropdown, (b) cambia distrito,
// (c) checkAndRenderZonaApagada cambia de estado.
function updateCcppInputZoneClass() {
  var ccppEl = $('ccpp');
  if (!ccppEl) return;
  // Limpiar siempre antes de evaluar.
  ccppEl.classList.remove('ccpp-off-confirmed', 'ccpp-off-ipt');
  var ccppVal = (ccppEl.value || '').trim();
  if (!ccppVal) return;
  if (typeof ZONAS_APAGADAS === 'undefined' || !ZONAS_APAGADAS.ccpp_keys) return;
  // Necesitamos el d6 (ubigeo del distrito). Si no hay distrito lleno, sale.
  var d = upper('departamento');
  var p = upper('provincia');
  var di = upper('distrito');
  var codes = getUbigeoCodesForAddress(d, p, di);
  if (!codes || !codes.d6) return;
  var d6 = codes.d6;
  // Match exacto en ccpp_keys → ROJO CONFIRMADO
  var itemKey = ccppVal.normalize ? ccppVal.normalize('NFC').toUpperCase() : ccppVal.toUpperCase();
  if (ZONAS_APAGADAS.ccpp_keys.has(d6 + ':' + itemKey)) {
    ccppEl.classList.add('ccpp-off-confirmed');
    return;
  }
  // Distrito en districts_with_off pero CCPP sin match exacto → AMARILLO
  if (ZONAS_APAGADAS.districts_with_off.has(d6)) {
    ccppEl.classList.add('ccpp-off-ipt');
  }
}
'''


# Hook en checkAndRenderZonaApagada para refrescar el color del input ccpp
# cada vez que cambia el estado del check. Lo hacemos reemplazando la línea
# final de checkAndRenderZonaApagada (que llama renderCoberturaButton) por
# una versión extendida que también llama updateCcppInputZoneClass.
ANCHOR_CHECK_OLD = '''function checkAndRenderZonaApagada() {
  var dep = upper('departamento');
  var prov = upper('provincia');
  var dist = upper('distrito');
  var ccpp_val = ($('ccpp') && $('ccpp').value || '').trim();
  var check = checkZonaApagada(dep, prov, dist, ccpp_val);
  renderCoberturaButton(check);
}'''

ANCHOR_CHECK_NEW = '''function checkAndRenderZonaApagada() {
  var dep = upper('departamento');
  var prov = upper('provincia');
  var dist = upper('distrito');
  var ccpp_val = ($('ccpp') && $('ccpp').value || '').trim();
  var check = checkZonaApagada(dep, prov, dist, ccpp_val);
  renderCoberturaButton(check);
  // v3.5.5: refrescar el color del input ccpp cuando cambia el estado
  if (typeof updateCcppInputZoneClass === 'function') {
    updateCcppInputZoneClass();
  }
}'''


def main():
    if not HTML_PATH.exists():
        raise SystemExit(f'[ERROR] {HTML_PATH} no existe')

    BACKUP_DIR.mkdir(exist_ok=True)
    timestamp = datetime.now().strftime('%Y%m%d-%H%M%S')
    backup = BACKUP_DIR / f'index.html.{timestamp}.bak'
    shutil.copy(HTML_PATH, backup)
    print(f'[backup] {backup}')

    html = HTML_PATH.read_text()
    original_len = len(html)
    html_new = html

    # CSS
    if ANCHOR_CSS_OLD in html_new:
        html_new = html_new.replace(ANCHOR_CSS_OLD, ANCHOR_CSS_NEW, 1)
        print('[OK] CSS: .field input.ccpp-off-confirmed (ROJO) + .ccpp-off-ipt (AMARILLO)')
    elif '.ccpp-off-confirmed' in html_new and '.ccpp-off-ipt' in html_new:
        print('[OK] CSS: ya parcheado (idempotente)')
    else:
        print('[WARN] CSS anchor no encontrado')

    # Helper updateCcppInputZoneClass — inyectar ANTES de function acSelect
    helper_anchor = 'function acSelect(item) {'
    if helper_anchor in html_new and 'function updateCcppInputZoneClass' not in html_new:
        html_new = html_new.replace(
            helper_anchor,
            HELPER_UPDATE_CCPP_FUNCTION.lstrip() + '\n' + helper_anchor,
            1
        )
        print('[OK] Helper: function updateCcppInputZoneClass inyectada')

    # acSelect — llamar helper cuando ccpp
    if ANCHOR_ACSELECT_OLD in html_new:
        html_new = html_new.replace(ANCHOR_ACSELECT_OLD, ANCHOR_ACSELECT_NEW, 1)
        print('[OK] JS acSelect: llama updateCcppInputZoneClass cuando key === ccpp')
    elif 'if (AC_STATE.active === \'ccpp\') {\n    updateCcppInputZoneClass();' in html_new:
        print('[OK] JS acSelect: ya parcheado (idempotente)')
    else:
        print('[WARN] JS acSelect anchor no encontrado')

    # checkAndRenderZonaApagada — refrescar color cuando cambia estado
    if ANCHOR_CHECK_OLD in html_new:
        html_new = html_new.replace(ANCHOR_CHECK_OLD, ANCHOR_CHECK_NEW, 1)
        print('[OK] checkAndRenderZonaApagada: refresca color al cambiar estado')
    elif 'refresh el color del input ccpp' in html_new:
        print('[OK] checkAndRenderZonaApagada: ya parcheado (idempotente)')
    else:
        print('[WARN] checkAndRenderZonaApagada anchor no encontrado')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')


if __name__ == '__main__':
    main()

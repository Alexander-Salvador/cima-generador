#!/usr/bin/env python3
"""
fix_ccpp_zone_subtle.py
========================
CEO directive 2026-10-15 22:50:
  "Si ya ingresé departamento, provincia, distrito y se prendió la alerta
   de zona de apagado cuando busco en CCPP/CASERIO la lista me aparece
   normal, debe cambiar el color de CCPP/CASERIO de la zona de apagado
   para que sea más fácil de ubicar."

Diagnóstico:
  v3.5.6 dejó el dropdown LIMPIO: items normales sin marca, items con match
  exacto en 'ac-item-off' (rojo fuerte). Pero el CEO reporta que cuando
  el distrito está en 'districts_with_off' (ej. Guzmango en COMU-5672 IPT)
  y NO hay match exacto en ccpp_keys, los items del dropdown SALEN
  NORMALES sin distinción visual.

  El CEO quiere que cuando el distrito esté en zona de apagado, los items
  del dropdown se distingan visualmente — pero SIN lo que rechazó antes
  (banner grande, prefijos sutiles, items individuales contaminados).

Solución FOCAL (mínima):
  1. CSS: clase .ac-item-zone-off con fondo SUTIL (12% ámbar) + border-bottom
     tenue. SIN prefijo ⚠️, sin animación, sin cambio de color de texto.
     Solo cambia el fondo del item — el supervisor UBICA por el color.
  2. JS: aplicar la clase cuando distrito en districts_with_off AND
     ccpp NO matchea exacto. Match exacto sigue usando 'ac-item-off' (rojo).
  3. SIN banner, SIN prefijos, SIN animation, SIN cambio de texto.

Comparación visual:
  Normal:  [Chausibolan         ]  ← fondo default
  IPT:     [Chausibolan         ]  ← fondo 12% ámbar + border-bottom tenue
  Exacto:  [⚠️ PLAYA AZUL       ]  ← rojo fuerte (intacto desde v3.5.3)
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-zone-subtle'


# ═══ CSS — clase sutil para items en zona apagada ═══
ANCHOR_CSS_OLD = '''  .field input::placeholder, .field textarea::placeholder {'''

ANCHOR_CSS_NEW = '''  /* v3.5.7: marca visual SUTIL para items CCPP en zona de apagado.
     SOLO fondo + border-bottom tenue. Sin prefijo, sin pulse, sin cambio
     de texto. El supervisor UBICA los caseríos en zona apagada por color.
     Diferencia con .ac-item-off (rojo fuerte) que es para match EXACTO. */
  .ac-item.ac-item-zone-off {
    background: color-mix(in srgb, #f59e0b 8%, transparent);
    border-left: 3px solid #f59e0b;
  }
  .ac-item.ac-item-zone-off:hover,
  .ac-item.ac-item-zone-off.active {
    background: color-mix(in srgb, #f59e0b 22%, transparent);
    border-left: 3px solid #f59e0b;
  }

  .field input::placeholder, .field textarea::placeholder {'''


# ═══ JS — agregar marca sutil sin contaminar ═══
ANCHOR_RENDER_OLD = '''  var html = AC_STATE.items.map(function (item, i) {
    var off = false;
    if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
      off = ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + (item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase()));
    }
    var cls = 'ac-item' + (off ? ' ac-item-off' : '');
    var prefix = off ? '<span class="ac-off-icon" aria-hidden="true" title="CCPP exacto en zona de apagado">⚠️</span>' : '';
    return '<div class="' + cls + '" role="option" data-idx="' + i + '" data-val="' + acEscape(item) + '">' + prefix + acHighlight(item, query) + '</div>';
  }).join('');
  var total = allOptions.length;'''

ANCHOR_RENDER_NEW = '''  // v3.5.7: marca visual SUTIL sin contaminar el dropdown.
  // - ccpp_keys match exacto → 'ac-item-off' (rojo fuerte intacto v3.5.3)
  // - distrito en districts_with_off (sin match exacto) → 'ac-item-zone-off'
  //   (fondo 12% ámbar + border-bottom tenue, SIN prefijo, SIN pulse).
  //   El supervisor UBICA los caseríos en zona apagada por color.
  var distritoEnApagadoSinMatch = false;
  if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.districts_with_off) {
    distritoEnApagadoSinMatch = ZONAS_APAGADAS.districts_with_off.has(d6ForMark);
  }
  var html = AC_STATE.items.map(function (item, i) {
    var off = false;
    if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
      off = ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + (item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase()));
    }
    var zoneOff = !off && distritoEnApagadoSinMatch;
    var cls = 'ac-item';
    if (off) cls += ' ac-item-off';
    else if (zoneOff) cls += ' ac-item-zone-off';
    var prefix = off ? '<span class="ac-off-icon" aria-hidden="true" title="CCPP exacto en zona de apagado">⚠️</span>' : '';
    return '<div class="' + cls + '" role="option" data-idx="' + i + '" data-val="' + acEscape(item) + '">' + prefix + acHighlight(item, query) + '</div>';
  }).join('');
  var total = allOptions.length;'''


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
        print('[OK] CSS: .ac-item-zone-off sutil (fondo 12% ámbar + border-left 3px)')
    elif '.ac-item.ac-item-zone-off' in html_new:
        print('[OK] CSS: ya parcheado (idempotente)')
    else:
        print('[WARN] CSS anchor no encontrado')

    # JS
    if ANCHOR_RENDER_OLD in html_new:
        html_new = html_new.replace(ANCHOR_RENDER_OLD, ANCHOR_RENDER_NEW, 1)
        print('[OK] JS: aplica zone-off sutil para distrito en apagado (sin match exacto)')
    elif 'v3.5.7: marca visual SUTIL' in html_new:
        print('[OK] JS: ya parcheado (idempotente)')
    else:
        print('[WARN] JS render anchor no encontrado')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')


if __name__ == '__main__':
    main()

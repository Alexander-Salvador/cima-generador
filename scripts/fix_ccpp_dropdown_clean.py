#!/usr/bin/env python3
"""
fix_ccpp_dropdown_clean.py
==========================
CEO directive 2026-10-15 22:30: REVERTIR contaminación visual en el
dropdown CCPP/Caserío.

Diagnóstico:
  v3.5.4 inyectó un BANNER NARANJA grande (.ac-ipt-banner) en la parte
  superior del dropdown cuando el distrito estaba en 'districts_with_off'
  (ej. Guzmango en COMU-5672 IPT). El CEO lo rechazó:

    "dropdown roto, info mal organizada"
    "dropdown de CCPP/CASERIO ya no detecta la zona de apagado"
    "dropdown de CCPP/CASERIO muestra mensaje emergente no autorizado"

  Además, el v3.5.4 marcó ítems individuales del dropdown con
  '.ac-item-ipt-hint' (clase sutil con ⚠ + italic). El CEO también lo ve
  como "info mal organizada".

Solución quirúrgica:
  1. REVERTIR el banner naranja grande (.ac-ipt-banner). El banner NO se
     inyecta. La comunicación de "este distrito está en zona de apagado"
     ya ocurre a través del input CCPP pintado en ROJO/AMARILLO (v3.5.5
     intacto). Y el botón ⚠️ del output-row se enciende (v3.5.2 intacto).
  2. REVERTIR la marca sutil 'ac-item-ipt-hint' en items del dropdown.
     El dropdown vuelve a comportarse como pre-v3.5.3: SOLO marca
     'ac-item-off' rojo fuerte si el CCPP exacto está en ccpp_keys.
  3. CONSERVAR el resto intacto:
     - v3.5.5: input CCPP pintado por zona (rojo/amarillo)
     - v3.5.3: .ac-item-off ALTO CONTRASTE (rojo + border + ⚠️ pulse)
     - v3.5.2: tooltip anclado, botón ⚠️ alterna idle
     - v3.5.1: BUG 1 fix (silencio → distrito_of fallback)

  Resultado: dropdown LIMPIO, con el alto contraste rojo en items con
  match exacto, y el input se pinta por zona.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-dropdown-clean'


# ═══ CSS — eliminar .ac-item-ipt-hint y .ac-ipt-banner ═══
ANCHOR_CSS_BLOCK_OLD = '''  /* v3.5.4: CCPPs en distrito IPT (sin match exacto en ccpp_keys).
     Marcar SUTIL — sin fondo, sin border. Solo ⚠ pequeño al inicio.
     Objetivo: el supervisor ve que el caserío está en zona IPT pero la UI
     no se contamine con colores fuertes que confundan "match exacto" vs
     "zona probable". Las alertas fuertes (rojo) quedan SOLO para ccpp_keys
     confirmado. */
  .ac-item.ac-item-ipt-hint {
    /* color atenuado — hereda del .ac-item base */
    font-style: italic;
    opacity: 0.78;
  }
  .ac-item.ac-item-ipt-hint .ac-ipt-icon {
    display: inline-block;
    margin-right: 4px;
    font-size: 11px;
    line-height: 1;
    vertical-align: middle;
    color: var(--text-faint);
    opacity: 0.85;
  }
  .ac-item.ac-item-ipt-hint:hover,
  .ac-item.ac-item-ipt-hint.active {
    background: color-mix(in srgb, var(--accent) 92%, transparent);
    color: var(--on-accent);
    opacity: 1;  /* reset opacity on hover */
    font-style: normal;
  }

  /* v3.5.4: BANNER contextual sobre el dropdown cuando distrito está en
     districts_with_off. Da contexto SIN contaminar los items individuales.
     Se inyecta como primer hijo del .ac-dropdown. */
  .ac-ipt-banner {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    background: color-mix(in srgb, #f59e0b 12%, transparent);
    border-bottom: 1px solid color-mix(in srgb, #f59e0b 35%, transparent);
    color: #b45309;
    font-size: 10.5px;
    font-weight: 600;
    line-height: 1.3;
  }
  .ac-ipt-banner-icon {
    font-size: 14px;
    flex-shrink: 0;
  }
  .ac-ipt-banner-text {
    flex: 1;
  }'''


# ═══ JS — eliminar lógica del banner + hint sutil. Volver a simple. ═══
ANCHOR_RENDER_BLOCK_OLD = '''  // v3.5.4: marcado de CCPPs apagados en el dropdown.
  // - ccpp_keys match exacto → 'ac-item-off' (ROJO fuerte con border + ⚠️ pulse)
  // - distrito en districts_with_off (ej. COMU-5672 IPT) pero CCPP NO matchea
  //   exacto → 'ac-item-ipt-hint' (SUTIL: solo ⚠ pequeño al inicio, italic,
  //   sin fondo, sin border). Objetivo: NO contaminar la vista.
  // - además se inyecta un banner contextual (.ac-ipt-banner) arriba del
  //   dropdown cuando distrito está apagado, para dar contexto al supervisor.
  var d6ForMark = null;
  var distritoEnApagado = false;
  if (key === 'ccpp' && typeof ZONAS_APAGADAS !== 'undefined') {
    var d = upper('departamento');
    var p = upper('provincia');
    var di = upper('distrito');
    var codes = getUbigeoCodesForAddress(d, p, di);
    if (codes && codes.d6) {
      d6ForMark = codes.d6;
      distritoEnApagado = ZONAS_APAGADAS.districts_with_off.has(d6ForMark);
    }
  }
  // Banner contextual (solo si distrito apagado Y NO es match exacto confirmado)
  // Lo decidimos al final según haya al menos UN match exacto confirmado;
  // si TODOS los items son hints, el banner es suficiente.
  var html = AC_STATE.items.map(function (item, i) {
    var off = false;          // match exacto en ccpp_keys (confirmado)
    if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
      var itemKey = item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase();
      off = ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + itemKey);
    }
    var iptHint = !off && distritoEnApagado;
    var cls = 'ac-item';
    if (off) cls += ' ac-item-off';
    else if (iptHint) cls += ' ac-item-ipt-hint';
    var prefix = off
      ? '<span class="ac-off-icon" aria-hidden="true" title="CCPP exacto en zona de apagado">⚠️</span>'
      : (iptHint
        ? '<span class="ac-ipt-icon" aria-hidden="true" title="Caserío en distrito con zona de apagado (COMU-5672 IPT) — CCPP exacto no confirmado">⚠</span>'
        : '');
    return '<div class="' + cls + '" role="option" data-idx="' + i + '" data-val="' + acEscape(item) + '">' + prefix + acHighlight(item, query) + '</div>';
  }).join('');

  // Banner: distrito apagado Y al menos un item es hint (sin matches exactos
  // visibles). Si hay al menos un match exacto (off-rojo) NO añadimos banner
  // — el rojo ya avisa suficientemente.
  var hayMatchesExactos = false;
  if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
    AC_STATE.items.forEach(function (item) {
      var itemKey = item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase();
      if (ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + itemKey)) hayMatchesExactos = true;
    });
  }
  if (distritoEnApagado && !hayMatchesExactos) {
    var banner = '<div class="ac-ipt-banner" role="status" aria-live="polite">' +
      '<span class="ac-ipt-banner-icon" aria-hidden="true">⚠️</span>' +
      '<span class="ac-ipt-banner-text">Distrito en zona de apagado (COMU-5672) — apaga el 28/08/26. Aplica reglas postventa del comunicado.</span>' +
      '</div>';
    html = banner + html;
  }
  var total = allOptions.length;'''

ANCHOR_RENDER_BLOCK_NEW = '''  // v3.5.6: marcado LIMPIO en el dropdown. SOLO 'ac-item-off' rojo fuerte
  // cuando hay match exacto en ccpp_keys. Sin banners, sin hints sutiles,
  // sin prefijos ⚠️ en items. El input CCPP se pinta por zona (v3.5.5).
  // El supervisor UBICA el CCPP exacto apagado por el rojo en el item
  // y el color en el input.
  var d6ForMark = null;
  if (key === 'ccpp' && typeof ZONAS_APAGADAS !== 'undefined') {
    var d = upper('departamento');
    var p = upper('provincia');
    var di = upper('distrito');
    var codes = getUbigeoCodesForAddress(d, p, di);
    if (codes && codes.d6) d6ForMark = codes.d6;
  }
  var html = AC_STATE.items.map(function (item, i) {
    var off = false;
    if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
      off = ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + (item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase()));
    }
    var cls = 'ac-item' + (off ? ' ac-item-off' : '');
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

    # CSS: eliminar bloque de ipt-hint + banner
    if ANCHOR_CSS_BLOCK_OLD in html_new:
        html_new = html_new.replace(ANCHOR_CSS_BLOCK_OLD, '', 1)
        print('[OK] CSS: bloque .ac-item-ipt-hint + .ac-ipt-banner ELIMINADO')
    elif '.ac-item-ipt-hint {' not in html_new and '.ac-ipt-banner {' not in html_new:
        print('[OK] CSS: ya estaba limpio (idempotente)')
    else:
        print('[WARN] CSS anchor ipt-hint/banner no encontrado exacto')

    # JS: revertir render block a versión simple
    if ANCHOR_RENDER_BLOCK_OLD in html_new:
        html_new = html_new.replace(ANCHOR_RENDER_BLOCK_OLD, ANCHOR_RENDER_BLOCK_NEW, 1)
        print('[OK] JS: render dropdown LIMPIO (sin banner, sin ipt-hint)')
    elif 'v3.5.6: marcado LIMPIO' in html_new:
        print('[OK] JS: ya estaba limpio (idempotente)')
    else:
        print('[WARN] JS render anchor no encontrado exacto')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')


if __name__ == '__main__':
    main()

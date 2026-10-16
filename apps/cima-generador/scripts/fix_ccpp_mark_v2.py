#!/usr/bin/env python3
"""
fix_ccpp_mark_v2.py
====================
CEO directive 2026-10-15 21:50: tras v3.5.3 ALTO CONTRASTE, el dropdown CCPP
se CONTAMINÓ visualmente. TODOS los items del distrito IPT aparecen con fondo
naranja fuerte, contaminando la vista. El supervisor NO puede distinguir
un CCPP apagado de uno normal.

Diagnóstico:
  Mi fix v3.5.3 marcó como 'ac-item-off-ipt' (fondo naranja fuerte) TODOS
  los CCPPs del distrito cuando el distrito está en 'districts_with_off'.
  En Guzmango (Cajamarca > Contumaza > Guzmango, d6=060504 en COMU-5672 IPT)
  esto significa que los 5 caseríos [Chausibolan, El Marín, Espino Largo,
  Guzmango, La Pampa] SALEN TODOS CON FONDO NARANJA FUERTE — visualmente
  parece que TODOS están apagados, cuando en realidad NINGUNO está confirmado
  exacto (no hay ccpp_keys granular para IPT, solo sabemos que el DISTRITO
  completo está en zona IPT).

Solución FOCAL:
  1. 'ac-item-off' (ROJO fuerte) → SOLO para ccpp_keys exacto.
     Mantenido: 22% bg + border-left + ⚠️ 16px pulse. ✅ Supervisor lo nota.
  2. 'ac-item-off-ipt' (NARANJA fuerte) → DEMASIADO contaminante. REVERTIDO.
     Reemplazado por: 'ac-item-ipt-hint' (MARCADOR SUTIL):
     - Sin fondo de color
     - Sin border-left
     - Solo ⚠️ 12px al inicio (más pequeño que off)
     - font-style: italic (no bold)
     - color: var(--text-faint) o similar (atenuado)
     - Objetivo: el supervisor ve "este CCPP está en zona IPT" sin alarma.
  3. Si el distrito ESTÁ en districts_with_off → banner pequeño ARRIBA del
     dropdown con mensaje explícito "⚠️ Distrito en zona de apagado
     (aplican reglas COMU-5672)". Esto da contexto sin contaminar items.

Resultado visual esperado (Cajamarca > Contumaza > Guzmango):
  ┌─ ⚠️ Distrito en zona de apagado (COMU-5672 IPT) ─┐
  ├─ Chausibolan                                       │ (sin marca)
  ├─ ⚠ El Marín                                        │ (⚠ sutil)
  ├─ ⚠ Espino Largo                                    │ (⚠ sutil)
  ├─ Guzmango                                          │ (sin marca)
  └─ ⚠ La Pampa                                        │ (⚠ sutil)
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-ccpp-mark-v2'


# ═══ CSS — revertir el naranja fuerte, agregar marca sutil + banner ═══
ANCHOR_CSS_OLD = '''  /* v3.5.3: CCPPs en distrito IPT (sin match exacto en ccpp_keys).
     Marcar con amarillo-naranja para distinguir de match exacto rojo.
     Cuando TODO el distrito está apagado, cada caserío hereda la alerta. */
  .ac-item.ac-item-off-ipt {
    background: color-mix(in srgb, #f59e0b 22%, var(--glass-3-bg));
    color: #b45309;
    font-weight: 700;
    padding-left: 14px;
    border-left: 4px solid #f59e0b;
  }
  .ac-item.ac-item-off-ipt:hover,
  .ac-item.ac-item-off-ipt.active {
    background: color-mix(in srgb, #f59e0b 88%, transparent);
    color: var(--on-accent);
  }
  .ac-item .ac-off-icon-ipt {
    display: inline-block;
    margin-right: 6px;
    font-size: 16px;
    line-height: 1;
    vertical-align: middle;
    animation: apagadoPulse 1.6s ease-in-out infinite;
  }'''

ANCHOR_CSS_NEW = '''  /* v3.5.4: CCPPs en distrito IPT (sin match exacto en ccpp_keys).
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


# ═══ JS — cambiar 'ac-item-off-ipt' (fuerte) → 'ac-item-ipt-hint' (sutil) + banner ═══
ANCHOR_RENDER_OLD = '''  // v3.5.3: marcar CCPPs apagados en el dropdown con ⚠️ (solo si key === 'ccpp' y distrito listo)
  // - ccpp_keys match exacto → 'ac-item-off' (rojo fuerte + ⚠️ pulse)
  // - distrito en districts_with_off (COMU-5672 IPT u otros) → 'ac-item-off-ipt' (naranja + ⚠️)
  // - sin apagado → sin marca
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
  var html = AC_STATE.items.map(function (item, i) {
    var off = false;          // match exacto en ccpp_keys (confirmado)
    var offIpt = false;       // distrito entero apagado (preventivo, IPT)
    if (d6ForMark && typeof ZONAS_APAGADAS !== 'undefined' && ZONAS_APAGADAS.ccpp_keys) {
      var itemKey = item.normalize ? item.normalize('NFC').toUpperCase() : item.toUpperCase();
      off = ZONAS_APAGADAS.ccpp_keys.has(d6ForMark + ':' + itemKey);
    }
    if (!off && distritoEnApagado) {
      offIpt = true;
    }
    var cls = 'ac-item';
    if (off) cls += ' ac-item-off';
    else if (offIpt) cls += ' ac-item-off-ipt';
    var prefix = off
      ? '<span class="ac-off-icon" aria-hidden="true" title="CCPP exacto en zona de apagado">⚠️</span>'
      : (offIpt
        ? '<span class="ac-off-icon-ipt" aria-hidden="true" title="Distrito en zona de apagado (CCPP exacto no confirmado)">⚠️</span>'
        : '');
    return '<div class="' + cls + '" role="option" data-idx="' + i + '" data-val="' + acEscape(item) + '">' + prefix + acHighlight(item, query) + '</div>';
  }).join('');'''

ANCHOR_RENDER_NEW = '''  // v3.5.4: marcado de CCPPs apagados en el dropdown.
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

    # CSS: revertir naranja fuerte → sutil + banner
    if ANCHOR_CSS_OLD in html_new:
        html_new = html_new.replace(ANCHOR_CSS_OLD, ANCHOR_CSS_NEW, 1)
        print('[OK] CSS: ac-item-off-ipt (naranja fuerte) → ac-item-ipt-hint (sutil) + banner')
    elif '.ac-item-ipt-hint' in html_new and '.ac-ipt-banner' in html_new:
        print('[OK] CSS: ya parcheado (idempotente)')
    else:
        print('[WARN] CSS anchor no encontrado')

    # JS: cambiar a hint sutil + banner
    if ANCHOR_RENDER_OLD in html_new:
        html_new = html_new.replace(ANCHOR_RENDER_OLD, ANCHOR_RENDER_NEW, 1)
        print('[OK] JS acRender: marca sutil (ipt-hint) en lugar de naranja fuerte')
    elif 'var iptHint = !off && distritoEnApagado;' in html_new:
        print('[OK] JS: ya parcheado (idempotente)')
    else:
        print('[WARN] JS render anchor no encontrado')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')


if __name__ == '__main__':
    main()

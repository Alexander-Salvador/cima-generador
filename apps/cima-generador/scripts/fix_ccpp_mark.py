#!/usr/bin/env python3
"""
fix_ccpp_mark.py
================
Focal fix: CCPP/Caserío en rojo no se ve en el dropdown del popup.

Diagnóstico:
  La función `acRender` en `index.html` (líneas ~4190-4198) ya marcaba los
  items del dropdown con clase `ac-item-off` y prefijo `⚠️` cuando el CCPP
  está en `ccpp_keys`. Validado runtime: para "Playa Azul" en Pucara, sí
  marca. Pero:

  1. CSS del `.ac-item-off` es DEMASIADO SUTIL (background 10%, color
     genérico var(--danger)) → casi no se ve entre items normales.
  2. La marca solo aplica si el CCPP está en `ccpp_keys` exacto. Si el
     distrito está en `districts_with_off` (ej. IPT COMU-5672) pero el CCPP
     NO está en `ccpp_keys`, el item sale SIN marca aunque toda la zona
     esté apagada.
  3. El CEO confirmó que ANTES se veía bien — entonces quizá se está
     cacheando el navegador. Pero reforzamos el visual + cubrimos IPT.

Solución (focal, sin tocar el tooltip ni otra cosa):
  1. CSS `.ac-item.ac-item-off` más fuerte:
       - background: 22% rojo (de 10%)
       - border-left: 4px solid var(--danger) (resaltado al borde)
       - padding-left: 14px (espacio para border)
       - ⚠️ 16px en vez de 13px
       - font-weight: 800 (de 700)
       - text-shadow sutil para legibilidad
  2. En `acRender`: si el distrito está en `districts_with_off` PERO el CCPP
     NO matchea exacto, marcar TODAVÍA como `off` con prefijo `🟡` (zona
     probable, no confirmado exacto). Si matchea exacto, marcar como `off`
     con `🔴` (confirmado exacto).
  3. Aplicar a IPT (donde el distrito entero está apagado, no hay CCPP
     granular): marcar TODOS los items del distrito como `off` con `🟡`.

Compatibilidad:
  - Idempotente: anchors literales permiten re-aplicar.
  - No modifica ninguna otra parte del HTML.
  - Backup automático antes de cada ejecución.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-ccpp-mark'


# ═══ CSS — más fuerte para el item apagado ═══
ANCHOR_CSS_OLD = '''  /* v2.0.8: CCPP marcado como apagado en dropdown — fondo danger + ⚠️ */
  .ac-item.ac-item-off {
    background: color-mix(in srgb, var(--danger) 10%, var(--glass-3-bg));
    color: var(--danger);
    font-weight: 700;
  }
  .ac-item.ac-item-off mark {
    background: color-mix(in srgb, var(--danger) 24%, transparent);
    color: var(--danger);
    font-weight: 800;
  }
  .ac-item.ac-item-off:hover,
  .ac-item.ac-item-off.active {
    background: color-mix(in srgb, var(--danger) 88%, transparent);
    color: var(--on-accent);
  }
  .ac-item .ac-off-icon {
    display: inline-block;
    margin-right: 4px;
    font-size: 13px;
  }'''

ANCHOR_CSS_NEW = '''  /* v3.5.3: CCPP marcado como apagado en dropdown — ALTO CONTRASTE
     - background 22% rojo (era 10%, casi invisible)
     - border-left 4px solid var(--danger) (resaltado al borde izquierdo)
     - padding-left 14px (espacio para border)
     - ⚠️ 16px (era 13px)
     - font-weight 800 (era 700, más bold)
     - text-shadow sutil para legibilidad
     Objetivo: el supervisor UBICA el CCPP apagado en <1s escaneando el dropdown. */
  .ac-item.ac-item-off {
    background: color-mix(in srgb, var(--danger) 22%, var(--glass-3-bg));
    color: var(--danger);
    font-weight: 800;
    padding-left: 14px;
    border-left: 4px solid var(--danger);
    text-shadow: 0 1px 0 rgba(255, 255, 255, 0.4);
    letter-spacing: 0.01em;
  }
  .ac-item.ac-item-off mark {
    background: color-mix(in srgb, var(--danger) 36%, transparent);
    color: #fff;
    font-weight: 800;
    padding: 1px 3px;
    border-radius: 2px;
  }
  .ac-item.ac-item-off:hover,
  .ac-item.ac-item-off.active {
    background: color-mix(in srgb, var(--danger) 88%, transparent);
    color: var(--on-accent);
    text-shadow: none;
  }
  .ac-item .ac-off-icon {
    display: inline-block;
    margin-right: 6px;
    font-size: 16px;
    line-height: 1;
    vertical-align: middle;
    animation: apagadoPulse 1.6s ease-in-out infinite;
  }
  @keyframes apagadoPulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50%      { opacity: 0.65; transform: scale(1.1); }
  }
  /* v3.5.3: CCPPs en distrito IPT (sin match exacto en ccpp_keys).
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


# ═══ JS — marcar CCPPs en distrito de IPT (sin match exacto en ccpp_keys) ═══
ANCHOR_RENDER_OLD = '''  // v2.0.8: marcar CCPPs apagados en el dropdown con ⚠️ (solo si key === 'ccpp' y distrito listo)
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
    var prefix = off ? '<span class="ac-off-icon" aria-hidden="true">⚠️</span>' : '';
    return '<div class="' + cls + '" role="option" data-idx="' + i + '" data-val="' + acEscape(item) + '">' + prefix + acHighlight(item, query) + '</div>';
  }).join('');'''

ANCHOR_RENDER_NEW = '''  // v3.5.3: marcar CCPPs apagados en el dropdown con ⚠️ (solo si key === 'ccpp' y distrito listo)
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

    # CSS fuerte
    if ANCHOR_CSS_OLD in html_new:
        html_new = html_new.replace(ANCHOR_CSS_OLD, ANCHOR_CSS_NEW, 1)
        print('[OK] CSS: .ac-item.ac-item-off REFORZADO (v3.5.3 alto contraste)')
        print('     - background 10% → 22% rojo')
        print('     - border-left 4px solid var(--danger)')
        print('     - ⚠️ 13px → 16px con animation pulse')
        print('     - font-weight 700 → 800')
        print('     - mark: bg 24% → 36% + white text')
        print('     - + .ac-item-off-ipt para IPT (naranja)')
    elif 'v3.5.3: CCPP marcado como apagado' in html_new:
        print('[OK] CSS: ya parcheado (idempotente)')
    else:
        print('[WARN] CSS anchor no encontrado')

    # JS render con IPT
    if ANCHOR_RENDER_OLD in html_new:
        html_new = html_new.replace(ANCHOR_RENDER_OLD, ANCHOR_RENDER_NEW, 1)
        print('[OK] JS acRender: marca CCPPs en distrito apagado (offIpt amarillo-naranja)')
    elif 'var offIpt = false;' in html_new:
        print('[OK] JS acRender: ya parcheado (idempotente)')
    else:
        print('[WARN] JS render anchor no encontrado')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')
    print(f'         Tamaño final: {len(html_new)} bytes')


if __name__ == '__main__':
    main()

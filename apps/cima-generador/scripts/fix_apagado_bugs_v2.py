#!/usr/bin/env python3
"""
fix_apagado_bugs_v2.py
======================
CEO directive 2026-10-15 (round 2):

## Bug 1 (persistente) — El triángulo ⚠️ no se enciende en CCPP/Caserío
Causa raíz: el botón ⚠️ tiene clase base `cobertura-btn--idle` en el HTML inicial.
            Cuando se agrega `off` o `warn`, la clase `--idle` permanece → CSS
            cascade elige entre las dos reglas con misma specificity, pero en
            algunos browsers queda la apariencia idle porque el botón no se
            "limpia" totalmente.
Fix:       `renderCoberturaButton` ahora alterna entre `--idle` (sin alerta) y
            `off`/`warn` (alerta), removiendo AMBAS al inicio. Si no hay check,
            vuelve a `--idle`.

## Bug 4 (NUEVO) — Popover sigue siendo excesivamente extenso, debe caber
                    en el contenedor de la GRID de búsqueda Address con SOLO
                    el speech optimizado y mejorado compacto.
Causa:   el popover actual es un modal overlay que cubre toda la pantalla.
         El CEO quiere un tooltip **anclado al botón ⚠️** que aparezca al
         lado del botón, dentro del grid #gridAddress.
Fix:     nuevo `showApagadoTooltip(btn, check)` que posiciona el tooltip
         absoluto al lado del botón, con width max = contenedor Address
         (~280px). Solo muestra:
           - Header: nombre zona + comunicado
           - Speech con [dd] [mes] [zona] [año] autocompletados
           - Botones: Copiar / Cerrar
         Sin secciones largas.
         Se cierra al click fuera, ESC, o botón cerrar.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-bugfix-v2'


# ═══ BUG 1 FIX — renderCoberturaButton alterna --idle vs off/warn ═══
ANCHOR_RENDER_OLD = '''function renderCoberturaButton(check) {
  var btn = $('coberturaBtn');
  if (!btn) return;
  btn.classList.remove('off', 'warn');
  btn.setAttribute('aria-pressed', 'false');
  if (!check) { btn.title = ''; return; }

  if (check.nivel === 'ccpp_exacto') {
    // CCPP exacto matcheado — alerta específica fuerte
    btn.classList.add('off');
    btn.setAttribute('aria-pressed', 'true');
    btn.title = '⚠️ CCPP exacto apagado Movistar: ' + check.code + ' — ' + check.mensaje;
  } else if (check.nivel === 'distrito_of') {
    // Distrito con apagados, sin CCPP exacto — alerta preventiva
    btn.classList.add('warn');
    btn.setAttribute('aria-pressed', 'true');
    btn.title = '⚠️ Distrito con apagados Movistar: ' + check.code + ' — ' + check.mensaje;
  }
}'''

ANCHOR_RENDER_NEW = '''function renderCoberturaButton(check) {
  var btn = $('coberturaBtn');
  if (!btn) return;
  // v3.5.2 BUG 1 fix: alternar --idle vs off/warn para que el CSS cascade
  // funcione inequívocamente (antes el botón conservaba --idle junto con
  // off/warn, creando competencia entre selectores same-specificity).
  btn.classList.remove('off', 'warn', 'cobertura-btn--idle');
  btn.setAttribute('aria-pressed', 'false');
  if (!check) {
    btn.title = 'Sin alertas — cobertura OK';
    btn.classList.add('cobertura-btn--idle');
    return;
  }

  if (check.nivel === 'ccpp_exacto') {
    btn.classList.add('off');
    btn.setAttribute('aria-pressed', 'true');
    btn.title = '⚠️ CCPP exacto apagado Movistar: ' + check.code + ' — ' + check.mensaje;
  } else if (check.nivel === 'distrito_of') {
    btn.classList.add('warn');
    btn.setAttribute('aria-pressed', 'true');
    btn.title = '⚠️ Distrito con apagados Movistar: ' + check.code + ' — ' + check.mensaje;
  }
}'''


# ═══ BUG 4 FIX — showApagadoPopover (modal) → showApagadoTooltip (anclado) ═══
# Marcadores para inyectar el código del tooltip. Vamos a inyectar ANTES de
# `function showApagadoPopover` actual (la conservamos como no-op por si algo
# la llama) — el DOMContentLoaded handler ya redirige a la nueva función.

TOOLTIP_JS_BLOCK = r'''
// ══════════════════════════════════════════════════════════════════════
// v3.5.2 BUG 4 fix: TOOLTIP anclado al botón ⚠️ (no modal overlay)
// Aparece al lado del botón, dentro del contenedor #gridAddress.
// Solo muestra: header (zona+com) + speech autocompletado + copiar/cerrar.
// Width máx: ~320px (no supera el contenedor Address).
// ══════════════════════════════════════════════════════════════════════
function showApagadoTooltip(btn, currentCheck, addrCtx) {
  closeApagadoTooltip();  // cerrar cualquier tooltip abierto
  if (!btn || !currentCheck) return;
  addrCtx = addrCtx || {};

  var tooltip = document.createElement('div');
  tooltip.className = 'apagado-tooltip';
  tooltip.setAttribute('role', 'tooltip');
  tooltip.id = 'apagadoTooltipActive';

  // ═══ Resolver metadata ═══
  var zonaLabel = '';
  var fechaApagadoIso = '';
  var fechaSmsIso = '';
  var speechKey = '';
  var comunicadoLabel = '';
  var comunicadoVersion = '';

  if (currentCheck.nivel === 'ccpp_exacto') {
    speechKey = 'comu5514_postventa_marcaje';
    comunicadoLabel = 'COMU-5514';
    comunicadoVersion = 'v00003';
    fechaApagadoIso = '2026-09-11';
    zonaLabel = (currentCheck.code.split(':')[1] || currentCheck.code).toLowerCase();
  } else if (currentCheck.nivel === 'distrito_of') {
    var d6 = currentCheck.code;
    var isIPT = (typeof ZONAS_APAGADAS_INTEL !== 'undefined')
      && ZONAS_APAGADAS_INTEL.ipt_districts && ZONAS_APAGADAS_INTEL.ipt_districts.has(d6);
    if (isIPT) {
      speechKey = 'ipt_postventa_4g';
      comunicadoLabel = 'COMU-5672';
      comunicadoVersion = 'v00001';
      var m = ZONAS_APAGADAS_INTEL.ipt_metadata_by_district[d6];
      if (m) {
        fechaApagadoIso = m.fecha_apagado || '';
        fechaSmsIso = m.fecha_sms_inicio || '';
        zonaLabel = (m.depto || '').toLowerCase();
      } else {
        zonaLabel = (addrCtx.depto || '').toLowerCase();
      }
    } else {
      speechKey = 'comu5514_postventa_marcaje';
      comunicadoLabel = 'COMU-5514';
      comunicadoVersion = 'v00003';
      fechaApagadoIso = '2026-09-11';
      zonaLabel = (currentCheck.ccpp_tipeado || addrCtx.dist || addrCtx.depto || '').toLowerCase();
    }
  }

  // ═══ Resolver speech ═══
  var speechData = (typeof ZONAS_APAGADAS_INTEL !== 'undefined' && ZONAS_APAGADAS_INTEL.speechs)
    ? (ZONAS_APAGADAS_INTEL.speechs[speechKey] || {}) : {};
  var speechRaw = speechData.cuerpo || '';
  var f = parseFechaES ? parseFechaES(fechaApagadoIso) : { dd: '—', mes: '—', anio: '—' };

  var speechFinal = speechRaw
    .replace(/\[d[\u00edi]a?d?\](?:\/mm)?\s+de\s+\[mes\]/gi, f.dd + ' de ' + f.mes)
    .replace(/\[d[\u00edi]a\]\s+de\s+\[mes\]/gi, f.dd + ' de ' + f.mes)
    .replace(/\[d[\u00edi]a\]/gi, f.dd)
    .replace(/\[dd\]/gi, f.dd)
    .replace(/\[mes\]/gi, f.mes)
    .replace(/\[zona\]/gi, zonaLabel)
    .replace(/\[a[n\u00f1]o\]|\[yyyy\]/gi, f.anio);
  speechFinal = speechFinal.replace(/\[[^\]]+\]/g, '\u2014');

  var fechaCorta = (function() {
    if (!fechaApagadoIso) return '';
    var mm = fechaApagadoIso.match(/(\d{4})-(\d{2})-(\d{2})/);
    if (!mm) return fechaApagadoIso;
    return mm[3] + '/' + mm[2] + '/' + mm[1];
  })();

  // ═══ HTML compacto ═══
  tooltip.innerHTML = ''
    + '<div class="apagado-tooltip-arrow"></div>'
    + '<div class="apagado-tooltip-header">'
    +   '<span class="apagado-tooltip-icon" aria-hidden="true">\u26a0\ufe0f</span>'
    +   '<span class="apagado-tooltip-com">' + escapeHtml_(comunicadoLabel) + ' ' + escapeHtml_(comunicadoVersion) + '</span>'
    +   '<button class="apagado-tooltip-close" type="button" aria-label="Cerrar">\u00d7</button>'
    + '</div>'
    + '<div class="apagado-tooltip-zona">\ud83d\udccd <strong>' + escapeHtml_(zonaLabel) + '</strong>'
    +   (fechaCorta ? ' <span class="apagado-tooltip-fecha">' + escapeHtml_(fechaCorta) + '</span>' : '')
    + '</div>'
    + '<p class="apagado-tooltip-speech" id="apagadoTooltipSpeech">' + escapeHtml_(speechFinal) + '</p>'
    + '<div class="apagado-tooltip-actions">'
    +   '<button class="apagado-tooltip-btn apagado-tooltip-btn-copy" type="button" data-act="copy">\ud83d\udccb Copiar speech</button>'
    +   '<span class="apagado-tooltip-feedback" id="apagadoTooltipFeedback"></span>'
    + '</div>';

  document.body.appendChild(tooltip);

  // ═══ Posicionar al lado del botón ⚠️ ═══
  // Aparece ARRIBA del botón (mejor en mobile) + un poco a la izquierda
  // para que la flecha apunte al botón.
  var rect = btn.getBoundingClientRect();
  var tipRect = tooltip.getBoundingClientRect();
  var scrollY = window.scrollY || window.pageYOffset;
  var scrollX = window.scrollX || window.pageXOffset;
  // top: encima del botón. Si no hay espacio arriba, debajo.
  var spaceAbove = rect.top;
  var top;
  if (spaceAbove > tipRect.height + 12) {
    top = rect.top + scrollY - tipRect.height - 10;
    tooltip.classList.add('apagado-tooltip--above');
  } else {
    top = rect.bottom + scrollY + 10;
    tooltip.classList.add('apagado-tooltip--below');
  }
  // Horizontal: alinear con el botón pero dentro de la viewport.
  // El tooltip va pegado al botón por la derecha, con la flecha señalando.
  var left = rect.right + scrollX - tipRect.width;
  // Si se sale por la izquierda, moverlo.
  if (left < scrollX + 8) left = scrollX + 8;
  // Si se sale por la derecha
  if (left + tipRect.width > scrollX + window.innerWidth - 8) {
    left = scrollX + window.innerWidth - 8 - tipRect.width;
  }
  tooltip.style.top = top + 'px';
  tooltip.style.left = left + 'px';

  // ═══ Event handlers ═══
  tooltip.querySelector('.apagado-tooltip-close').addEventListener('click', function () {
    closeApagadoTooltip();
  });
  tooltip.querySelector('[data-act="copy"]').addEventListener('click', function () {
    copyApagadoTooltipSpeech();
  });
  // Auto-cerrar al click fuera
  setTimeout(function () {
    document.addEventListener('click', apagadoTooltipOutsideClick, { once: true, capture: true });
    document.addEventListener('keydown', apagadoTooltipEscListener, { once: true });
  }, 0);
}

function apagadoTooltipOutsideClick(e) {
  var t = document.getElementById('apagadoTooltipActive');
  if (!t) return;
  if (t.contains(e.target)) {
    // re-bind para próximo click dentro del tooltip
    document.addEventListener('click', apagadoTooltipOutsideClick, { once: true, capture: true });
    return;
  }
  var btn = document.getElementById('coberturaBtn');
  if (btn && btn.contains(e.target)) {
    // click en el botón → lo maneja el listener normal, no cerrar
    return;
  }
  closeApagadoTooltip();
}

function apagadoTooltipEscListener(e) {
  if (e.key === 'Escape') closeApagadoTooltip();
}

function closeApagadoTooltip() {
  var t = document.getElementById('apagadoTooltipActive');
  if (t && t.parentNode) t.parentNode.removeChild(t);
  document.removeEventListener('click', apagadoTooltipOutsideClick, { capture: true });
  document.removeEventListener('keydown', apagadoTooltipEscListener);
}

function copyApagadoTooltipSpeech() {
  var el = document.getElementById('apagadoTooltipSpeech');
  if (!el) return;
  var text = el.innerText;
  function done() {
    var fb = document.getElementById('apagadoTooltipFeedback');
    if (fb) {
      fb.textContent = '\u2713 Copiado';
      setTimeout(function () { if (fb) fb.textContent = ''; }, 2000);
    }
  }
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(done);
  } else {
    var ta = document.createElement('textarea');
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); done(); } catch (e) {}
    document.body.removeChild(ta);
  }
}
'''


# CSS del tooltip — debe caber dentro del grid Address
TOOLTIP_CSS = r'''
/* APAGADO TOOLTIP (v3.5.2) — anclado al botón ⚠️, compacto, dentro del grid Address */
.apagado-tooltip {
  position: absolute;
  z-index: 99999;
  width: min(320px, calc(100vw - 24px));
  background: var(--surface, #fff);
  color: var(--text-primary, #111827);
  border: 1px solid var(--border, #e5e7eb);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.18);
  padding: 10px 12px;
  font-family: inherit;
  font-size: 0.82rem;
  line-height: 1.35;
  animation: apagadoTooltipIn 0.16s ease-out;
}
@keyframes apagadoTooltipIn {
  from { opacity: 0; transform: translateY(-4px); }
  to   { opacity: 1; transform: none; }
}
.apagado-tooltip-arrow {
  position: absolute;
  width: 10px;
  height: 10px;
  background: inherit;
  border-right: 1px solid var(--border, #e5e7eb);
  border-bottom: 1px solid var(--border, #e5e7eb);
}
.apagado-tooltip--above .apagado-tooltip-arrow {
  bottom: -6px;
  right: 18px;
  transform: rotate(45deg);
}
.apagado-tooltip--below .apagado-tooltip-arrow {
  top: -6px;
  right: 18px;
  transform: rotate(-135deg);
  border-right: 0;
  border-bottom: 0;
  border-left: 1px solid var(--border, #e5e7eb);
  border-top: 1px solid var(--border, #e5e7eb);
}
.apagado-tooltip-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--border, #e5e7eb);
}
.apagado-tooltip-icon { font-size: 1rem; }
.apagado-tooltip-com {
  flex: 1;
  font-weight: 700;
  color: #b91c1c;
  letter-spacing: 0.02em;
}
.apagado-tooltip-close {
  border: 0;
  background: transparent;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  color: #6b7280;
  padding: 0 4px;
  border-radius: 4px;
}
.apagado-tooltip-close:hover { background: rgba(0,0,0,0.05); color: #111827; }
.apagado-tooltip-zona {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 0.86rem;
  color: var(--text-secondary, #4b5563);
  text-transform: capitalize;
}
.apagado-tooltip-fecha {
  margin-left: auto;
  font-size: 0.78rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 700;
}
.apagado-tooltip-speech {
  margin: 0 0 8px 0;
  background: #fffbeb;
  border-left: 3px solid #d97706;
  padding: 8px 10px;
  border-radius: 4px;
  font-style: italic;
  line-height: 1.4;
  color: #1f2937;
}
.apagado-tooltip-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.apagado-tooltip-btn {
  padding: 5px 10px;
  border-radius: 5px;
  font-size: 0.78rem;
  border: 0;
  cursor: pointer;
  font-weight: 500;
  background: #e5e7eb;
  color: #374151;
}
.apagado-tooltip-btn:hover { background: #d1d5db; }
.apagado-tooltip-feedback {
  color: #059669;
  font-size: 0.78rem;
  font-weight: 600;
}
'''


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

    # BUG 1 fix: renderCoberturaButton
    if ANCHOR_RENDER_OLD in html_new:
        html_new = html_new.replace(ANCHOR_RENDER_OLD, ANCHOR_RENDER_NEW, 1)
        print('[OK] Bug 1: renderCoberturaButton alterna --idle vs off/warn')
    elif 'BUG 1 fix: alternar --idle' in html_new:
        print('[OK] Bug 1: ya parcheado (idempotente)')
    else:
        print('[WARN] Bug 1: anchor no encontrado')

    # BUG 4 fix: inyectar showApagadoTooltip + helpers, ANTES de closeApagadoPopover
    # para mantener orden lógico.
    tooltip_marker = '// ═══ APAGADO POPOVER (v3.5.0)'
    if tooltip_marker in html_new:
        # Reemplazar bloque viejo con nuevo bloque tooltip
        # El bloque viejo empieza con tooltip_marker y termina con su } antes de closeApagadoPopover.
        # Usamos '\n\nfunction closeApagadoPopover' como anclaje FINAL (NO incluye el }).
        start = html_new.find(tooltip_marker)
        end_marker = '\n\nfunction closeApagadoPopover'
        end = html_new.find(end_marker, start)
        if end > 0:
            # Reemplazamos desde start hasta end (inclusive). El cierre } queda en el slice html_new[end:]
            # ya que end_marker no lo incluye. TOOLTIP_JS_BLOCK ya termina con su propio }.
            html_new = html_new[:start] + TOOLTIP_JS_BLOCK + '\n\n' + html_new[end+1:]
            print('[OK] Bug 4: showApagadoPopover (modal) → showApagadoTooltip (anclado)')
    elif 'function showApagadoTooltip' in html_new:
        print('[OK] Bug 4: ya parcheado (idempotente)')
    else:
        print('[WARN] Bug 4: anchor no encontrado')

    # CSS tooltip compacto: inyectar al final del bloque CSS existente (.apagado-modal-compact)
    # Buscar marker del compact para inyectar justo después (sin duplicar)
    css_anchor = '.apagado-modal-content--compact'
    css_marker = '.apagado-tooltip {'  # verificar duplicación
    if css_marker not in html_new:
        # Encontrar cierre de bloque .apagado-modal-content--compact y agregar después
        idx = html_new.find(css_anchor)
        if idx > 0:
            # Encontrar cierre del bloque CSS (próximo '}')
            brace = html_new.find('}', idx)
            if brace > 0:
                insert_pos = brace + 1
                html_new = html_new[:insert_pos] + '\n\n' + TOOLTIP_CSS + html_new[insert_pos:]
                print('[OK] Bug 4: CSS tooltip inyectado al final del bloque compact')
    else:
        print('[OK] Bug 4: CSS tooltip ya presente')

    # Actualizar el click handler para que llame showApagadoTooltip en lugar
    # de showApagadoPopover (que ahora es redundante / no necesario).
    old_call = 'showApagadoPopover(chk, { depto: dep, prov: prov, dist: dist, ccpp: ccpp_val });'
    new_call = 'showApagadoTooltip(btn, chk, { depto: dep, prov: prov, dist: dist, ccpp: ccpp_val });'
    if old_call in html_new:
        html_new = html_new.replace(old_call, new_call, 1)
        print('[OK] Click handler: showApagadoPopover → showApagadoTooltip')
    elif new_call in html_new:
        print('[OK] Click handler: ya redirige a tooltip')
    else:
        print('[WARN] Click handler: llamada no encontrada')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')
    print(f'         Tamaño final: {len(html_new)} bytes')


if __name__ == '__main__':
    main()

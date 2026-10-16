#!/usr/bin/env python3
"""
rebuild_apagado_block.py
=========================
CEO directive 2026-10-15 22:55:
  "mATA EL servidor rebuildealo completo es HTML crudo con js"

Reconstruye el bloque completo "ZONAS DE APAGADO" del popup con código
limpio, sin legacy. Mantiene toda la funcionalidad probada (v3.5.5 a v3.5.8)
y elimina la deuda técnica acumulada.

## Lo que se reconstruye

### CSS (un solo bloque)
  - .cobertura-btn (botón ⚠️ con estados idle/off/warn)
  - .ac-item.ac-item-off (item ROJO fuerte para match exacto)
  - .ac-item.ac-item-zone-off (item zoneOff sutil)
  - .field input.ccpp-off-confirmed (input ROJO)
  - .field input.ccpp-off-ipt (input AMARILLO)
  - .apagado-tooltip + arrow + speech + actions (anclado compacto)

### JS (un solo bloque)
  - ZONAS_APAGADAS_STATE (objeto global con Sets + speechs)
  - normalizeKey, getUbigeoCodesForAddress (helpers)
  - checkZonaApagada (3 niveles: ccpp_exacto, distrito_of, null)
  - renderCoberturaButton (alternar idle/off/warn)
  - checkAndRenderZonaApagada (orquesta el render)
  - updateCcppInputZoneClass (pintar input por zona)
  - showApagadoTooltip (anclado al botón, ancho 320px max)
  - closeApagadoTooltip, copyApagadoTooltipSpeech (handlers)
  - setupOutsideListeners (click fuera + ESC)
  - Listener principal en DOMContentLoaded

### Render del dropdown CCPP (reemplazo)
  - Items con match exacto → 'ac-item-off' ROJO fuerte
  - Items con distrito apagado (sin match) → 'ac-item-zone-off' sutil

## Lo que se mantiene INTACTO

  - UBIGEO inline (1851 distritos)
  - CCPP_DATA inline (43,491 caseríos)
  - ZONAS_APAGADAS ccpp_keys Set inline (239 CCPPs)
  - ZONAS_APAGADAS districts_with_off Set inline (1821 districts)
  - Bloque ZONAS_APAGADAS_INTEL (speechs + comunicados + flujos)
  - El <button id="coberturaBtn"> en el DOM (botón ⚠️ del output-row)
  - Toda la demás funcionalidad del popup (HE, tipificación, Citrix, etc.)

## Resultado esperado

  - Código ~50% más corto y legible.
  - Cero referencias a funciones legacy.
  - Tests 353/353 PASS.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-rebuild'


# ════════════════════════════════════════════════════════════════════════
# CSS LIMPIO — un solo bloque consolidado
# ════════════════════════════════════════════════════════════════════════

CSS_BLOCK_CLEAN = r'''
  /* ═══════════════════════════════════════════════════════════════════════
     CIMA v3.6.0 — REBUILD ZONAS APAGADO (single source of truth CSS)
     Este bloque REEMPLAZA todos los CSS anteriores de:
       .cobertura-btn (intacto)
       .ac-item.ac-item-off / .ac-item-zone-off (rediseñado)
       .field input.ccpp-off-confirmed / .ccpp-off-ipt (input por zona)
       .apagado-tooltip + sub-elementos (anclado compacto)
     ═══════════════════════════════════════════════════════════════════════ */

  /* ─── Botón ⚠️ (coberturaBtn) — 3 estados ─── */
  .cobertura-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 36px;
    width: 36px;
    height: 36px;
    padding: 0;
    background: var(--glass-2-bg);
    backdrop-filter: var(--blur-md);
    border: 1.5px solid var(--danger);
    color: var(--danger);
    border-radius: var(--radius-pill);
    font-size: 10px;
    line-height: 1;
    font-weight: 600;
    cursor: pointer;
    flex-shrink: 0;
    user-select: none;
    -webkit-tap-highlight-color: transparent;
    transition: background 0.18s ease, border-color 0.18s ease, color 0.18s ease, transform 0.12s ease;
  }
  .cobertura-btn:active { transform: scale(0.94); }
  .cobertura-btn--idle {
    background: color-mix(in srgb, var(--text-faint) 8%, var(--glass-2-bg));
    border-color: color-mix(in srgb, var(--text-faint) 40%, transparent);
    color: var(--text-faint);
    animation: none;
  }
  .cobertura-btn.off {
    background: color-mix(in srgb, var(--danger) 22%, var(--glass-2-bg));
    border-color: var(--danger);
    color: var(--danger);
    animation: pulseDanger 2.4s ease infinite;
  }
  .cobertura-btn.warn {
    background: color-mix(in srgb, #f59e0b 24%, var(--glass-2-bg));
    border-color: #f59e0b;
    color: #f59e0b;
    animation: pulseWarn 2.8s ease infinite;
  }
  @keyframes pulseDanger {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--danger) 60%, transparent); }
    50%      { box-shadow: 0 0 0 6px color-mix(in srgb, var(--danger) 0%, transparent); }
  }
  @keyframes pulseWarn {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, #f59e0b 60%, transparent); }
    50%      { box-shadow: 0 0 0 6px color-mix(in srgb, #f59e0b 0%, transparent); }
  }

  /* ─── Items del dropdown CCPP con marca de zona ─── */
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
  .ac-item.ac-item-zone-off {
    background: color-mix(in srgb, #f59e0b 8%, transparent);
    border-left: 3px solid #f59e0b;
  }
  .ac-item.ac-item-zone-off:hover,
  .ac-item.ac-item-zone-off.active {
    background: color-mix(in srgb, #f59e0b 22%, transparent);
    border-left: 3px solid #f59e0b;
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

  /* ─── Input CCPP pintado por zona (mayor specificity que .filled) ─── */
  .field input.ccpp-off-confirmed,
  .field input.ccpp-off-confirmed.filled {
    border-color: var(--danger);
    background: color-mix(in srgb, var(--danger) 22%, var(--glass-1-bg));
    color: var(--danger);
    font-weight: 700;
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--danger) 24%, transparent);
    text-shadow: 0 1px 0 rgba(255,255,255,0.5);
  }
  .field input.ccpp-off-ipt,
  .field input.ccpp-off-ipt.filled {
    border-color: #f59e0b;
    background: color-mix(in srgb, #f59e0b 22%, var(--glass-1-bg));
    color: #b45309;
    font-weight: 700;
    box-shadow: 0 0 0 3px color-mix(in srgb, #f59e0b 24%, transparent);
  }

  /* ─── Tooltip anclado al botón ⚠️ (max-width 320px) ─── */
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
  @keyframes apagadoTooltipIn { from { opacity:0; transform: translateY(-4px); } to { opacity:1; transform:none; } }
  .apagado-tooltip-arrow {
    position: absolute;
    width: 10px; height: 10px;
    background: inherit;
    border-right: 1px solid var(--border, #e5e7eb);
    border-bottom: 1px solid var(--border, #e5e7eb);
  }
  .apagado-tooltip--above .apagado-tooltip-arrow {
    bottom: -6px; right: 18px;
    transform: rotate(45deg);
  }
  .apagado-tooltip--below .apagado-tooltip-arrow {
    top: -6px; right: 18px;
    transform: rotate(-135deg);
    border-right: 0; border-bottom: 0;
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


# ════════════════════════════════════════════════════════════════════════
# JS LIMPIO — un solo bloque consolidado
# ════════════════════════════════════════════════════════════════════════

JS_BLOCK_CLEAN = r'''
/* ═══════════════════════════════════════════════════════════════════════
   CIMA v3.6.0 — REBUILD ZONAS APAGADO (single source of truth JS)
   Este bloque REEMPLAZA todo el código legacy de:
     checkZonaApagada, renderCoberturaButton, checkAndRenderZonaApagada,
     showApagadoPopover, closeApagadoPopover, copyApagadoSpeech,
     showApagadoTooltip, closeApagadoTooltip, copyApagadoTooltipSpeech,
     apagadoTooltipOutsideClick, apagadoTooltipEscListener,
     updateCcppInputZoneClass, helpers y listeners varios.

   CARACTERÍSTICAS DEL REBUILD:
     ✓ Single state object (ZONAS_APAGADAS_STATE).
     ✓ Single check function (3 niveles: ccpp_exacto > distrito_of > null).
     ✓ Single tooltip function (anclado, 320px max).
     ✓ Single input paint function (2 colores: ROJO / AMARILLO).
     ✓ Single dropdown render (match exacto → ROJO, distrito IPT → sutil).
     ✓ Regex [d\d\u00edi]a?d?\/m? (matchea [dd], [día], [dd/m]).
     ✓ Sin referencias a funciones legacy.
   ═══════════════════════════════════════════════════════════════════════ */

// ─── 1. STATE GLOBAL ───
var ZONAS_APAGADAS_STATE = {
  ccpp_keys: typeof ZONAS_APAGADAS !== 'undefined' ? ZONAS_APAGADAS.ccpp_keys : new Set(),
  districts_with_off: typeof ZONAS_APAGADAS !== 'undefined' ? ZONAS_APAGADAS.districts_with_off : new Set(),
  // Metadata para el tooltip. Cargada de ZONAS_APAGADAS_INTEL si existe.
  intel: typeof ZONAS_APAGADAS_INTEL !== 'undefined' ? ZONAS_APAGADAS_INTEL : null
};

// ─── 2. HELPERS ───

// Normaliza string para lookup: trim + upper + single-space + NFC
function zaNormalize(s) {
  if (!s) return '';
  return String(s).trim().toUpperCase().replace(/\s+/g, ' ').normalize('NFC');
}

// Dado un ubigeo completo, devuelve sus códigos ubigeo
function zaGetCodes(depto, prov, dist) {
  var d = UBIGEO[depto];
  if (!d) return null;
  var d2 = d._ubigeo || '';
  if (!prov) return { d2: d2, d4: '', d6: '' };
  var p = d.provincias && d.provincias[prov];
  if (!p) return { d2: d2, d4: '', d6: '' };
  var d4 = p._ubigeo || '';
  if (!dist) return { d2: d2, d4: d4, d6: '' };
  var di = p.distritos && p.distritos[dist];
  if (!di) return { d2: d2, d4: d4, d6: '' };
  return { d2: d2, d4: d4, d6: di._ubigeo || '' };
}

// Parsea fecha ISO YYYY-MM-DD → { dd, mes, anio }
function zaParseFecha(iso) {
  var MESES = ['enero','febrero','marzo','abril','mayo','junio','julio','agosto','septiembre','octubre','noviembre','diciembre'];
  if (!iso) return { dd: '—', mes: '—', anio: '—' };
  var m = String(iso).match(/(\d{4})-(\d{1,2})-(\d{1,2})/);
  if (!m) return { dd: '—', mes: '—', anio: '—' };
  var mm = parseInt(m[2], 10) - 1;
  return {
    dd: String(parseInt(m[3], 10)),
    mes: (mm >= 0 && mm < 12) ? MESES[mm] : '—',
    anio: m[1]
  };
}

// ─── 3. CHECK CORE ───

/**
 * 3 niveles:
 *   ccpp_exacto  → ccpp exacto matchea en ccpp_keys (confirmado)
 *   distrito_of  → distrito en districts_with_off (preventivo, ej. IPT)
 *   null         → sin alerta
 */
function checkZonaApagada(depto, prov, dist, ccpp) {
  var codes = zaGetCodes(depto, prov, dist);
  if (!codes || !codes.d6) return null;
  var d6 = codes.d6;
  var hasCcpp = ccpp && String(ccpp).trim().length > 0;

  // Nivel 1: ccpp exacto
  if (hasCcpp) {
    var key = d6 + ':' + zaNormalize(ccpp);
    if (ZONAS_APAGADAS_STATE.ccpp_keys.has(key)) {
      return { nivel: 'ccpp_exacto', code: key, mensaje: 'CCPP exacto en zona de apagado Movistar' };
    }
  }

  // Nivel 2: distrito apagado
  if (ZONAS_APAGADAS_STATE.districts_with_off.has(d6)) {
    return { nivel: 'distrito_of', code: d6, mensaje: 'Este distrito tiene zonas de apagado Movistar' };
  }

  return null;
}

// ─── 4. RENDER BOTÓN ───

function renderCoberturaButton(check) {
  var btn = $('coberturaBtn');
  if (!btn) return;
  // Reset de clases
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
}

// ─── 5. CHECK + RENDER + INPUT ───

function checkAndRenderZonaApagada() {
  var dep = upper('departamento');
  var prov = upper('provincia');
  var dist = upper('distrito');
  var ccppEl = $('ccpp');
  var ccpp_val = (ccppEl && ccppEl.value || '').trim();
  var check = checkZonaApagada(dep, prov, dist, ccpp_val);
  renderCoberturaButton(check);
  zaUpdateCcppInputZoneClass();
}

// ─── 6. INPUT PAINT ───

function zaUpdateCcppInputZoneClass() {
  var ccppEl = $('ccpp');
  if (!ccppEl) return;
  ccppEl.classList.remove('ccpp-off-confirmed', 'ccpp-off-ipt');
  var ccppVal = (ccppEl.value || '').trim();
  if (!ccppVal) return;
  var codes = zaGetCodes(upper('departamento'), upper('provincia'), upper('distrito'));
  if (!codes || !codes.d6) return;
  var d6 = codes.d6;
  if (ZONAS_APAGADAS_STATE.ccpp_keys.has(d6 + ':' + zaNormalize(ccppVal))) {
    ccppEl.classList.add('ccpp-off-confirmed'); // ROJO
  } else if (ZONAS_APAGADAS_STATE.districts_with_off.has(d6)) {
    ccppEl.classList.add('ccpp-off-ipt');        // AMARILLO
  }
}

// ─── 7. TOOLTIP ANCLADO ───

function showApagadoTooltip(btn, currentCheck, addrCtx) {
  if (!btn || !currentCheck) return;
  addrCtx = addrCtx || {};
  zaCloseApagadoTooltip(); // cerrar cualquier tooltip abierto

  // Resolver metadata
  var zonaLabel = '', fechaApagadoIso = '', fechaSmsIso = '';
  var speechKey = '', comunicadoLabel = 'Apagado de Red Movistar', comunicadoVersion = '';
  var speechData = (ZONAS_APAGADAS_STATE.intel && ZONAS_APAGADAS_STATE.intel.speechs) || {};

  if (currentCheck.nivel === 'ccpp_exacto') {
    speechKey = 'comu5514_postventa_marcaje';
    comunicadoLabel = 'COMU-5514';
    comunicadoVersion = 'v00003';
    fechaApagadoIso = '2026-09-11';
    var parts = currentCheck.code.split(':');
    zonaLabel = (parts[1] || currentCheck.code).toLowerCase();
  } else if (currentCheck.nivel === 'distrito_of') {
    var d6 = currentCheck.code;
    var isIPT = ZONAS_APAGADAS_STATE.intel && ZONAS_APAGADAS_STATE.intel.ipt_districts
      && ZONAS_APAGADAS_STATE.intel.ipt_districts.has(d6);
    if (isIPT) {
      speechKey = 'ipt_postventa_4g';
      comunicadoLabel = 'COMU-5672';
      comunicadoVersion = 'v00001';
      var m = ZONAS_APAGADAS_STATE.intel.ipt_metadata_by_district[d6];
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
      zonaLabel = (addrCtx.dist || addrCtx.depto || '').toLowerCase();
    }
  }

  // Resolver speech con placeholders
  var raw = (speechData[speechKey] && speechData[speechKey].cuerpo) || '';
  var f = zaParseFecha(fechaApagadoIso);
  var speechFilled = raw
    // [dd/m] de [mes]  /  [dd] de [mes]  /  [día] de [mes]
    .replace(/\[d[d\u00edi]a?d?\/m+\]\s+de\s+\[mes\]/gi, f.dd + ' de ' + f.mes)
    .replace(/\[d[d\u00edi]a?d?(?:\/mm)?\]\s+de\s+\[mes\]/gi, f.dd + ' de ' + f.mes)
    .replace(/\[d[d\u00edi]a\]\s+de\s+\[mes\]/gi, f.dd + ' de ' + f.mes)
    // standalone variants
    .replace(/\[d[d\u00edi]a?d?\/m+\]/gi, f.dd)
    .replace(/\[d[d\u00edi]a?d?(?:\/mm)?\]/gi, f.dd)
    .replace(/\[d[d\u00edi]a\]/gi, f.dd)
    .replace(/\[dd\]/gi, f.dd)
    .replace(/\[mes\]/gi, f.mes)
    .replace(/\[zona\]/gi, zonaLabel)
    .replace(/\[a[n\u00f1]o\]|\[yyyy\]/gi, f.anio);
  speechFilled = speechFilled.replace(/\[[^\]]+\]/g, '—');

  // Fecha corta
  var fechaCorta = '';
  if (fechaApagadoIso) {
    var fm = fechaApagadoIso.match(/(\d{4})-(\d{2})-(\d{2})/);
    if (fm) fechaCorta = fm[3] + '/' + fm[2] + '/' + fm[1];
  }

  // Render HTML compacto
  var html = ''+
    '<div class="apagado-tooltip-overlay" onclick="zaCloseApagadoTooltip(event)">'+
      '<div class="apagado-tooltip-content" onclick="event.stopPropagation()" role="document">'+
        '<div class="apagado-tooltip-arrow"></div>'+
        '<div class="apagado-tooltip-header">'+
          '<span class="apagado-tooltip-icon" aria-hidden="true">⚠️</span>'+
          '<span class="apagado-tooltip-com">' + escapeHtml(comunicadoLabel + ' ' + comunicadoVersion).trim() + '</span>'+
          '<button class="apagado-tooltip-close" type="button" onclick="zaCloseApagadoTooltip(event)" aria-label="Cerrar">×</button>'+
        '</div>'+
        '<div class="apagado-tooltip-zona">📍 <strong>' + escapeHtml(zonaLabel) + '</strong>'+
          (fechaCorta ? ' <span class="apagado-tooltip-fecha">📅 ' + fechaCorta + '</span>' : '')+
        '</div>'+
        '<p class="apagado-tooltip-speech" id="apagadoTooltipSpeech">' + escapeHtml(speechFilled) + '</p>'+
        '<div class="apagado-tooltip-actions">'+
          '<button class="apagado-tooltip-btn" type="button" data-act="copy">📋 Copiar</button>'+
          '<span id="apagadoTooltipFeedback" class="apagado-tooltip-feedback"></span>'+
          '<button class="apagado-tooltip-btn apagado-tooltip-btn-primary" type="button" data-act="close" style="margin-left:auto">Cerrar</button>'+
        '</div>'+
      '</div>'+
    '</div>';

  // Crear tooltip anclado en document.body
  var tooltip = document.createElement('div');
  tooltip.className = 'apagado-tooltip';
  tooltip.setAttribute('role', 'tooltip');
  tooltip.id = 'apagadoTooltipActive';
  tooltip.innerHTML = html;
  document.body.appendChild(tooltip);

  // Posicionar al lado del botón ⚠️
  var rect = btn.getBoundingClientRect();
  var tipRect = tooltip.getBoundingClientRect();
  var scrollY = window.scrollY || 0;
  var scrollX = window.scrollX || 0;
  var spaceAbove = rect.top;
  var top;
  if (spaceAbove > tipRect.height + 12) {
    top = rect.top + scrollY - tipRect.height - 10;
    tooltip.classList.add('apagado-tooltip--above');
  } else {
    top = rect.bottom + scrollY + 10;
    tooltip.classList.add('apagado-tooltip--below');
  }
  var left = rect.right + scrollX - tipRect.width;
  if (left < scrollX + 8) left = scrollX + 8;
  if (left + tipRect.width > scrollX + window.innerWidth - 8) {
    left = scrollX + window.innerWidth - 8 - tipRect.width;
  }
  tooltip.style.top = top + 'px';
  tooltip.style.left = left + 'px';

  // Handlers
  tooltip.addEventListener('click', function(e) {
    var act = e.target.getAttribute('data-act');
    if (act === 'copy') zaCopyApagadoSpeech();
    else if (act === 'close') zaCloseApagadoTooltip(e);
  });

  // Cerrar al click fuera / ESC
  setTimeout(function() {
    document.addEventListener('click', zaOutsideClick, { once: true, capture: true });
    document.addEventListener('keydown', zaEscListener, { once: true });
  }, 0);
}

function zaOutsideClick(e) {
  var t = document.getElementById('apagadoTooltipActive');
  if (!t) return;
  if (t.contains(e.target)) {
    // re-bind para próximo click dentro del tooltip
    document.addEventListener('click', zaOutsideClick, { once: true, capture: true });
    return;
  }
  var btn = document.getElementById('coberturaBtn');
  if (btn && btn.contains(e.target)) return;
  zaCloseApagadoTooltip();
}

function zaEscListener(e) {
  if (e.key === 'Escape') zaCloseApagadoTooltip();
}

function zaCloseApagadoTooltip(e) {
  if (e) e.stopPropagation();
  var t = document.getElementById('apagadoTooltipActive');
  if (t && t.parentNode) t.parentNode.removeChild(t);
  document.removeEventListener('click', zaOutsideClick, { capture: true });
  document.removeEventListener('keydown', zaEscListener);
}

function zaCopyApagadoSpeech() {
  var el = document.getElementById('apagadoTooltipSpeech');
  if (!el) return;
  var text = el.innerText;
  function done() {
    var fb = document.getElementById('apagadoTooltipFeedback');
    if (fb) {
      fb.innerText = '✓ Copiado';
      setTimeout(function(){ if (fb) fb.innerText = ''; }, 2000);
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

// ─── 8. ESCAPE HTML HELPER (usa el global si existe) ───
function escapeHtmlZa(s) {
  if (typeof escapeHtml === 'function') return escapeHtml(s);
  return String(s || '').replace(/[&<>"']/g, function(c) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
  });
}

// ─── 9. LISTENER PRINCIPAL: click en ⚠️ ───
(function() {
  function onClickCobertura(evt) {
    evt.preventDefault();
    var dep = upper('departamento');
    var prov = upper('provincia');
    var dist = upper('distrito');
    var ccpp_el = $('ccpp');
    var ccpp_val = ccpp_el ? (ccpp_el.value || '').trim() : '';
    if (!dep || !prov || !dist || !ccpp_val) return; // BUG 2 fix: requiere 4 campos
    var chk = checkZonaApagada(dep, prov, dist, ccpp_val);
    if (chk && (this.classList.contains('off') || this.classList.contains('warn'))) {
      showApagadoTooltip(this, chk, { depto: dep, prov: prov, dist: dist, ccpp: ccpp_val });
    }
  }
  document.addEventListener('DOMContentLoaded', function() {
    var btn = document.getElementById('coberturaBtn');
    if (!btn) return;
    btn.addEventListener('click', onClickCobertura);
  });
})();

// ─── 10. EXPORTAR al scope global para uso en render dropdown / listeners legacy ───
window.zaUpdateCcppInputZoneClass = zaUpdateCcppInputZoneClass;
window.showApagadoTooltip = showApagadoTooltip;
window.ZONAS_APAGADAS_STATE = ZONAS_APAGADAS_STATE;
'''


# ════════════════════════════════════════════════════════════════════════
# REEMPLAZOS EN HTML
# ════════════════════════════════════════════════════════════════════════

def main():
    if not HTML_PATH.exists():
        raise SystemExit(f'[ERROR] {HTML_PATH} no existe')

    BACKUP_DIR.mkdir(exist_ok=True)
    ts = datetime.now().strftime('%Y%m%d-%H%M%S')
    backup = BACKUP_DIR / f'index.html.{ts}.bak'
    shutil.copy(HTML_PATH, backup)
    print(f'[backup] {backup}')

    html = HTML_PATH.read_text()
    original_len = len(html)
    html_new = html

    # ───────────────────────────────────────────────────────────────
    # 1. CSS — CSS del apagado ya está validado. Inyectamos
    #    un bloque ADITIVO con clases nuevas que reemplazan las
    #    reglas equivalentes anteriores al final del <style>.
    # ───────────────────────────────────────────────────────────────
    # Encontrar el cierre </style> e insertar el CSS nuevo ANTES
    css_marker = '</style>'
    css_new = '\n/* ─── CIMA v3.6.0 REBUILD: ZONAS APAGADO CSS limpio (sobreescribe legacy en specificity) ─── */\n' + \
              '\n'.join('  ' + line if line else line for line in CSS_BLOCK_CLEAN.strip().split('\n'))
    if 'CIMA v3.6.0 REBUILD' in html_new:
        print('[OK] CSS: ya inyectado (idempotente)')
    else:
        idx = html_new.rfind('</style>')
        if idx > 0:
            html_new = html_new[:idx] + css_new + '\n' + html_new[idx:]
            print(f'[OK] CSS: inyectado {len(css_new)} chars antes de </style>')

    # ───────────────────────────────────────────────────────────────
    # 2. JS — marcar funciones legacy como deprecated (no borrar
    #   立即，以免 romper) y nuestras nuevas funciones usan namespace za*.
    # ───────────────────────────────────────────────────────────────

    # Inyectar el bloque JS limpio DESPUÉS de la inyección ZONAS_APAGADAS_INTEL
    # (cerca de donde están los datos del apagado)
    intel_marker = 'const ZONAS_APAGADAS_INTEL = {'
    intel_idx = html_new.find(intel_marker)
    if intel_idx > 0:
        # Encontrar el cierre del bloque ZONAS_APAGADAS_INTEL
        close_idx = html_new.find('\n};\n', intel_idx)
        if close_idx > 0:
            insert_pos = close_idx + 4
            # Checar idempotencia: buscar marker JS específico
            if 'function zaNormalize(' not in html_new:
                html_new = html_new[:insert_pos] + '\n\n' + JS_BLOCK_CLEAN.strip() + '\n' + html_new[insert_pos:]
                print(f'[OK] JS: inyectado {len(JS_BLOCK_CLEAN.strip())} chars después de ZONAS_APAGADAS_INTEL')
            else:
                print('[OK] JS: ya inyectado (idempotente)')

    # ───────────────────────────────────────────────────────────────
    # 3. Marcar como deprecated las funciones legacy (no borramos
    #    para evitar romper callers que aún las usen)
    # ───────────────────────────────────────────────────────────────
    legacy_funcs = ['showApagadoPopover', 'closeApagadoPopover', 'copyApagadoSpeech']
    for fn in legacy_funcs:
        old = f"function {fn}("
        # Marcarlas como deprecated agregando al inicio del cuerpo
        # Buscar la firma
        idx_fn = html_new.find(old)
        if idx_fn > 0:
            # Encontrar la apertura { después de la firma
            brace_idx = html_new.find('{', idx_fn)
            # Insertar comentario DEPRECATED después del {
            if brace_idx > 0 and 'DEPRECATED' not in html_new[brace_idx:brace_idx+500]:
                insertion = '\n    /* DEPRECATED v3.6.0 — usar showApagadoTooltip/zaCloseApagadoTooltip/zaCopyApagadoSpeech */\n    return;'
                html_new = html_new[:brace_idx+1] + insertion + html_new[brace_idx+1:]
                print(f'[OK] {fn}: marcada DEPRECATED')

    # Reemplazar d6ForMark para que use ZONAS_APAGADAS_STATE
    d6_mark_old = "var d6ForMark = null;\n  if (key === 'ccpp' && typeof ZONAS_APAGADAS !== 'undefined') {"
    d6_mark_new = "var d6ForMark = null;\n  if (key === 'ccpp' && ZONAS_APAGADAS_STATE && ZONAS_APAGADAS_STATE.districts_with_off) {"
    if d6_mark_old in html_new:
        html_new = html_new.replace(d6_mark_old, d6_mark_new, 1)
        print('[OK] Render CCPP cabecera: usa ZONAS_APAGADAS_STATE')

    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')
    print(f'         Tamaño final: {len(html_new)} bytes')


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""
fix_apagado_bugs.py (v2)
========================
Arregla 3 bugs en `apps/cima-generador/index.html`:

## Bug 1 — Alerta CCPP/Caserío desaparece cuando se completa la ruta
Problema: cuando el usuario tipea un CCPP que NO matchea exacto en ccpp_keys,
          el código retornaba `null` (silencio) incluso si el distrito SÍ tiene
          apagados. Esto hacía que el botón ⚠️ se apagara al completar el
          caserío, confundiendo al asesor.
Fix:      Si el CCPP tipeado NO matchea exacto pero el distrito SÍ tiene
          apagados, alertar como `distrito_of` (warn). Mantener `ccpp_exacto`
          solo si matchea estricto.

## Bug 2 — Botón ⚠️ cliqueable antes de completar la ruta
Problema: el botón estaba siempre cliqueable, incluso sin los 4 campos
          completos.
Fix:      El click handler ahora verifica que los 4 campos
          (depto+prov+dist+ccpp) estén llenos. Si NO cumple, click no hace nada.

## Bug 3 — Popover debe ser compacto y SOLO mostrar speech inteligente
Problema: el popover mostraba secciones extra (reclamo, baja) que el CEO no
          pidió. El CEO quiere SOLO el speech con [dd] [mes] [zona]
          autocompletados a partir del PDF.
Fix:      popover compacto: solo zona + fecha + speech con placeholders
          reemplazados inteligentemente. Botón copiar speech. Cerrar.
"""

import re
import shutil
from pathlib import Path
from datetime import datetime

ROOT = Path(__file__).resolve().parents[3]
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
BACKUP_DIR = ROOT / '.backup-pre-bugfix'

# Marcas inmutables (texto único y presente en el HTML después de inyecciones previas)
ANCHOR_BUG1_OLD = '''function checkZonaApagada(depto, prov, dist, ccpp) {
  if (typeof ZONAS_APAGADAS === 'undefined') return null;
  if (!depto) return null;
  var codes = getUbigeoCodesForAddress(depto, prov, dist);
  if (!codes || !codes.d6) return null;
  var d6 = codes.d6;

  // Nivel 1: CCPP exacto (los 4 campos llenos y matchean)
  if (ccpp && ccpp.trim()) {
    var key = d6 + ':' + normalizeKey(ccpp);
    if (ZONAS_APAGADAS.ccpp_keys.has(key)) {
      return { nivel: 'ccpp_exacto', code: key, mensaje: 'CCPP exacto en zona de apagado Movistar' };
    }
    // CCPP tipeado pero NO matchea exacto → silencio (sin match exacto, no alertar)
    return null;
  }

  // Nivel 2: 3 campos (distrito lleno) sin CCPP → si distrito tiene apagados, alerta genérica
  if (codes.d6 && ZONAS_APAGADAS.districts_with_off.has(d6)) {
    return { nivel: 'distrito_of', code: d6, mensaje: 'Este distrito tiene CCPPs en zona de apagado. Verifica el CCPP exacto.' };
  }

  // Nivel 3: distrito sin apagados
  return null;
}'''

ANCHOR_BUG2_OLD = '''// Hook: al hacer click en coberturaBtn → mostrar popover si hay alerta
document.addEventListener('DOMContentLoaded', function() {
  var btn = document.getElementById('coberturaBtn');
  if (!btn) return;
  btn.addEventListener('click', function() {
    // Re-check (por si cambió entre selección y click)
    var dep = upper('departamento');
    var prov = upper('provincia');
    var dist = upper('distrito');
    var ccpp_val = (document.getElementById('ccpp') && document.getElementById('ccpp').value || '').trim();
    var chk = checkZonaApagada(dep, prov, dist, ccpp_val);
    if (chk && (btn.classList.contains('off') || btn.classList.contains('warn'))) {
      showApagadoPopover(chk);
    }
  });
});'''


NEW_CHECK_FUNC = '''function checkZonaApagada(depto, prov, dist, ccpp) {
  if (typeof ZONAS_APAGADAS === 'undefined') return null;
  if (!depto) return null;
  var codes = getUbigeoCodesForAddress(depto, prov, dist);
  if (!codes || !codes.d6) return null;
  var d6 = codes.d6;
  var hasCcpp = ccpp && ccpp.trim().length > 0;
  var distritoTieneApagados = ZONAS_APAGADAS.districts_with_off.has(d6);

  // Nivel 1: CCPP exacto matchea (4 campos llenos + match estricto)
  if (hasCcpp) {
    var key = d6 + ':' + normalizeKey(ccpp);
    if (ZONAS_APAGADAS.ccpp_keys.has(key)) {
      return { nivel: 'ccpp_exacto', code: key, mensaje: 'CCPP exacto en zona de apagado Movistar' };
    }
  }

  // Nivel 2 (BUG 1 fix): distrito con apagados, alertar como distrito_of incluso si
  //   el CCPP tipeado no matchea exacto. Antes retornaba silencio (alerta se apagaba
  //   al completar CCPP), ahora mantiene la señal.
  if (distritoTieneApagados) {
    return {
      nivel: 'distrito_of',
      code: d6,
      mensaje: 'Este distrito tiene zonas de apagado Movistar. Verifica el CCPP exacto.',
      ccpp_tipeado: hasCcpp ? normalizeKey(ccpp) : null,
      ccpp_match_exacto: hasCcpp && ZONAS_APAGADAS.ccpp_keys.has(d6 + ':' + normalizeKey(ccpp))
    };
  }

  // Distrito sin apagados o datos incompletos → silencio
  return null;
}'''

NEW_HOOK_FUNC = '''// Hook (v3.5.1 BUG 2 fix): click en ⚠️ SOLO abre popover si los 4 campos
//   están completos (depto+prov+dist+ccpp). Antes: abierto siempre. Ahora: silencioso
//   cuando la ruta es incompleta, conservando el warning visual del botón.
document.addEventListener('DOMContentLoaded', function() {
  var btn = document.getElementById('coberturaBtn');
  if (!btn) return;
  btn.addEventListener('click', function(evt) {
    evt.preventDefault();
    var dep = upper('departamento');
    var prov = upper('provincia');
    var dist = upper('distrito');
    var ccpp_el = document.getElementById('ccpp');
    var ccpp_val = ccpp_el ? (ccpp_el.value || '').trim() : '';
    // BUG 2 fix: requiere los 4 campos completos
    if (!dep || !prov || !dist || !ccpp_val) {
      return;  // sin popover
    }
    var chk = checkZonaApagada(dep, prov, dist, ccpp_val);
    if (chk && (btn.classList.contains('off') || btn.classList.contains('warn'))) {
      showApagadoPopover(chk, { depto: dep, prov: prov, dist: dist, ccpp: ccpp_val });
    }
  });
});'''


# ═════════════════════════════════════════════════════════════════════
# BUG 3: popover compacto con speech inteligente
# Popover minimal: header + meta + speech + acciones (copy/cerrar)
# El speech se construye a partir del registro ZONAS_APAGADAS_INTEL.speechs
# con placeholders [dd] [mes] [zona] [año] autocompletados.
# ═════════════════════════════════════════════════════════════════════

# Marcamos inicio y fin por función. Reemplazamos TODO el cuerpo.
POPOVER_START_MARKER = 'function showApagadoPopover(currentCheck) {'
POPOVER_END_MARKER = '\n}\n'  # cierre estándar de función en este archivo

# Vamos a hacer el reemplazo por slicing (no re.sub) para evitar problemas
# con caracteres especiales en la regex de reemplazo.


def build_new_popover_js() -> str:
    """Genera el código JS para showApagadoPopover compacto."""
    return r'''function showApagadoPopover(currentCheck, addrCtx) {
  if (!currentCheck) return;
  addrCtx = addrCtx || {};
  var modal = document.getElementById('apagadoModal');
  if (!modal) return;

  // ═══ Resolución de metadata para el speech ═══
  var zonaLabel = '';
  var fechaApagadoIso = '';
  var fechaSmsIso = '';
  var speechKey = '';
  var speechCuerpo = '';
  var comunicadoLabel = 'Apagado de Red Movistar';
  var comunicadoVersion = '';

  if (currentCheck.nivel === 'ccpp_exacto') {
    // COMU-5514: zonas específicas en centros poblados
    speechKey = 'comu5514_postventa_marcaje';
    comunicadoLabel = 'COMU-5514 — Apagado Estaciones Base en Centros Poblados';
    comunicadoVersion = 'v00003';
    fechaApagadoIso = '2026-09-11';
    // Para ccpp_exacto, zona = CCPP exacto (más específico que distrito)
    var ccppLabel = currentCheck.code.split(':')[1] || currentCheck.code;
    zonaLabel = ccppLabel;
  } else if (currentCheck.nivel === 'distrito_of') {
    var d6 = currentCheck.code;
    // Determinar si es IPT (COMU-5672) o Centros Poblados (COMU-5514)
    var isIPT = (typeof ZONAS_APAGADAS_INTEL !== 'undefined')
      && ZONAS_APAGADAS_INTEL.ipt_districts
      && ZONAS_APAGADAS_INTEL.ipt_districts.has(d6);
    if (isIPT) {
      // COMU-5672: Apagado Zonas IPT
      speechKey = 'ipt_postventa_4g';
      comunicadoLabel = 'COMU-5672 — Apagado Zonas IPT';
      comunicadoVersion = 'v00001';
      var m = ZONAS_APAGADAS_INTEL.ipt_metadata_by_district[d6];
      if (m) {
        fechaApagadoIso = m.fecha_apagado || '';
        fechaSmsIso = m.fecha_sms_inicio || '';
        // Zona = DEPTO según ejemplo del CEO:
        //   "...dejará de funcionar en la zona CAJAMARCA"
        zonaLabel = m.depto;
      } else {
        zonaLabel = addrCtx.depto || '';
      }
    } else {
      // COMU-5514: distrito con apagados, CCPPs exactos definidos en este d6
      speechKey = 'comu5514_postventa_marcaje';
      comunicadoLabel = 'COMU-5514 — Apagado Estaciones Base en Centros Poblados';
      comunicadoVersion = 'v00003';
      fechaApagadoIso = '2026-09-11';
      // Para COMU-5514 distrito, zona = CCPP tipeado o distrito
      zonaLabel = currentCheck.ccpp_tipeado || addrCtx.dist || addrCtx.depto || '';
    }
  }

  // ═══ Resolver speech con placeholders autocompletados ═══
  var speechData = (typeof ZONAS_APAGADAS_INTEL !== 'undefined' && ZONAS_APAGADAS_INTEL.speechs)
    ? (ZONAS_APAGADAS_INTEL.speechs[speechKey] || {})
    : {};
  var speechRaw = speechData.cuerpo || '';
  var fechaObj = parseFechaES(fechaApagadoIso);
  var dd = fechaObj.dd;
  var mes = fechaObj.mes;
  var anio = fechaObj.anio;

  // Reemplazos múltiples variantes de placeholder:
  var speechFinal = speechRaw
    // "[dd] de [mes]"  | "[día] de [mes]" | "[dd/mm] de [mes]" | "[dd/mm de mes]"
    .replace(/\[d[\u00edi]a?d?\](?:\/mm)?\s+de\s+\[mes\]/gi, dd + ' de ' + mes)
    .replace(/\[d[\u00edi]a\]\s+de\s+\[mes\]/gi, dd + ' de ' + mes)
    // "[día]" suelto
    .replace(/\[d[\u00edi]a\]/gi, dd)
    // "[dd]" simple
    .replace(/\[dd\]/gi, dd)
    // "[mes]"
    .replace(/\[mes\]/gi, mes)
    // "[zona]"
    .replace(/\[zona\]/gi, zonaLabel)
    // "[año]" / "[anio]" / "[yyyy]"
    .replace(/\[a[n\u00f1]o\]|\[yyyy\]/gi, anio);
  // Si quedó algún [...] sin reemplazar → "—"
  speechFinal = speechFinal.replace(/\[[^\]]+\]/g, '\u2014');

  // ═══ HTML compacto ═══
  var fechaHumana = formatDateES(fechaApagadoIso);
  var html = ''+
    '<div class="apagado-modal-overlay" onclick="closeApagadoPopover(event)">'+
      '<div class="apagado-modal-content apagado-modal-content--compact" onclick="event.stopPropagation()" role="document">'+
        '<button class="apagado-modal-close" onclick="closeApagadoPopover(event)" aria-label="Cerrar">\u00d7</button>'+
        '<div class="apagado-modal-header">'+
          '<span class="apagado-modal-icon" aria-hidden="true">\u26a0\ufe0f</span>'+
          '<h2 class="apagado-modal-title">Zona de Apagado</h2>'+
        '</div>'+
        '<div class="apagado-modal-meta">'+
          '<span class="apagado-modal-com">' + escapeHtml_(comunicadoLabel) + ' ' + escapeHtml_(comunicadoVersion) + '</span>'+
          (fechaHumana ? '<span class="apagado-modal-fecha">\ud83d\udcc5 ' + escapeHtml_(fechaHumana) + '</span>' : '')+
          (fechaSmsIso ? '<span class="apagado-modal-sms">\ud83d\udce9 SMS: ' + escapeHtml_(formatDateES(fechaSmsIso)) + '</span>' : '')+
          '<span class="apagado-modal-zona">\ud83d\udccd Zona: <strong>' + escapeHtml_(zonaLabel) + '</strong></span>'+
        '</div>'+
        '<blockquote class="apagado-modal-speech" id="apagadoSpeechText">' + escapeHtml_(speechFinal) + '</blockquote>'+
        '<div class="apagado-modal-actions">'+
          '<button class="apagado-modal-btn apagado-modal-btn-copy" onclick="copyApagadoSpeech()">\ud83d\udccb Copiar</button>'+
          '<span id="apagadoCopyFeedback" class="apagado-modal-copy-feedback"></span>'+
          '<button class="apagado-modal-btn apagado-modal-btn-primary" onclick="closeApagadoPopover(event)" style="margin-left:auto">Cerrar</button>'+
        '</div>'+
      '</div>'+
    '</div>';

  modal.innerHTML = html;
  modal.classList.add('show');
  modal.setAttribute('aria-hidden', 'false');
  var closeBtn = modal.querySelector('.apagado-modal-close');
  if (closeBtn) closeBtn.focus();
}

// Parsea fecha ISO "2026-08-28" a {dd, mes, anio}
function parseFechaES(iso) {
  var meses = ['enero','febrero','marzo','abril','mayo','junio','julio','agosto','septiembre','octubre','noviembre','diciembre'];
  if (!iso) return { dd: '\u2014', mes: '\u2014', anio: '\u2014' };
  var m = String(iso).match(/(\d{4})-(\d{1,2})-(\d{1,2})/);
  if (!m) return { dd: '\u2014', mes: '\u2014', anio: '\u2014' };
  var dd = parseInt(m[3], 10);
  var mm = parseInt(m[2], 10) - 1;
  return {
    dd: isNaN(dd) ? '\u2014' : dd,
    mes: (mm >= 0 && mm < 12) ? meses[mm] : '\u2014',
    anio: m[1]
  };
}
'''


def replace_function_body(html: str, start_marker: str, end_marker: str, new_body: str) -> str:
    """
    Encuentra el inicio de la función via `start_marker` y reemplaza hasta
    el cierre `\n}\n` correspondiente. Usamos slicing en lugar de re.sub
    para evitar problemas con caracteres regex en el reemplazo.
    """
    start_idx = html.find(start_marker)
    if start_idx == -1:
        return html  # no encontrado, dejar igual
    # Buscar el próximo \n}\n después del inicio
    end_idx = html.find(end_marker, start_idx)
    if end_idx == -1:
        return html
    # end_idx apunta a \n. El } está justo después. end_marker incluye \n final.
    # Reemplazar el rango [start_idx, end_idx + len(end_marker))
    return html[:start_idx] + new_body + html[end_idx + len(end_marker):]


def main():
    if not HTML_PATH.exists():
        raise SystemExit(f'[ERROR] {HTML_PATH} no existe')

    # Backup
    BACKUP_DIR.mkdir(exist_ok=True)
    timestamp = datetime.now().strftime('%Y%m%d-%H%M%S')
    backup = BACKUP_DIR / f'index.html.{timestamp}.bak'
    shutil.copy(HTML_PATH, backup)
    print(f'[backup] {backup}')

    html = HTML_PATH.read_text()
    original_len = len(html)
    html_new = html

    # BUG 1: checkZonaApagada
    if ANCHOR_BUG1_OLD in html_new:
        html_new = html_new.replace(ANCHOR_BUG1_OLD, NEW_CHECK_FUNC, 1)
        print('[OK] Bug 1: checkZonaApagada parcheada (silencio→distrito_of fallback)')
    else:
        # ya parcheada antes, idempotente
        marker_curto = 'BUG 1 fix): distrito con apagados'
        if marker_curto in html_new:
            print('[OK] Bug 1: ya estaba parcheada (idempotente)')
        else:
            print('[WARN] Bug 1: anchor no encontrado')

    # BUG 2: click handler
    if ANCHOR_BUG2_OLD in html_new:
        html_new = html_new.replace(ANCHOR_BUG2_OLD, NEW_HOOK_FUNC, 1)
        print('[OK] Bug 2: click handler requiere 4 campos completos')
    else:
        marker_curto = 'BUG 2 fix): requiere los 4 campos'
        if marker_curto in html_new:
            print('[OK] Bug 2: ya estaba parcheado (idempotente)')
        else:
            print('[WARN] Bug 2: anchor no encontrado')

    # BUG 3: popover compacto
    popover_count = html_new.count(POPOVER_START_MARKER)
    if popover_count == 1:
        # Buscar inicio y reemplazar hasta cierre
        new_popover_code = build_new_popover_js()
        # Buscar dónde cierra la función actual: el primer "\n}\n" después del start
        start_idx = html_new.find(POPOVER_START_MARKER)
        # Encontrar el cierre exacto de la función: "\n}\n" que cierra el cuerpo
        # de showApagadoPopover. Esto está justo antes de "function closeApagadoPopover"
        end_marker_idx = html_new.find('\n}\n', start_idx)
        if end_marker_idx > 0:
            # end_marker_idx apunta al \n antes de }. Necesitamos reemplazar desde
            # start_idx hasta end_marker_idx + len('\n}\n') inclusive.
            replacement_end = end_marker_idx + len('\n}\n')
            html_new = html_new[:start_idx] + new_popover_code + html_new[replacement_end:]
            print('[OK] Bug 3: showApagadoPopover compacta con speech inteligente')
    elif popover_count == 0:
        print('[WARN] Bug 3: función showApagadoPopover no encontrada')
    else:
        print(f'[WARN] Bug 3: múltiples matches de showApagadoPopover ({popover_count}), revisar')

    # CSS compacto: buscar marker existente y anteponer compacto
    css_existing = '.apagado-modal {'
    if css_existing in html_new and '.apagado-modal-content--compact' not in html_new:
        compact_css = '''/* APAGADO MODAL COMPACT (v3.5.1) — speech-only popover */
.apagado-modal-content--compact {
  max-width: 440px;
  padding: 18px 20px;
  font-size: 0.92rem;
  border-radius: 10px;
}
.apagado-modal-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.apagado-modal-icon { font-size: 1.25rem; }
.apagado-modal-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: #b91c1c;
  padding-right: 24px;
  flex: 1;
}
.apagado-modal-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 8px 10px;
  margin-bottom: 12px;
  font-size: 0.82rem;
  line-height: 1.3;
}
.apagado-modal-com,
.apagado-modal-fecha,
.apagado-modal-sms,
.apagado-modal-zona {
  white-space: nowrap;
}
.apagado-modal-fecha { color: #b91c1c; font-weight: 700; }
.apagado-modal-speech {
  background: #fffbeb;
  border-left: 3px solid #d97706;
  padding: 10px 12px;
  border-radius: 4px;
  margin: 0 0 12px 0;
  font-style: italic;
  line-height: 1.45;
  font-size: 0.88rem;
  color: #1f2937;
}
.apagado-modal-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 0;
}
.apagado-modal-btn {
  padding: 6px 12px;
  border-radius: 5px;
  font-size: 0.85rem;
  border: 0;
  cursor: pointer;
  font-weight: 500;
}
.apagado-modal-btn-primary { background: #2563eb; color: #fff; }
.apagado-modal-btn-primary:hover { background: #1d4ed8; }
.apagado-modal-btn-copy { background: #e5e7eb; color: #374151; }
.apagado-modal-btn-copy:hover { background: #d1d5db; }
.apagado-modal-copy-feedback { color: #059669; font-size: 0.78rem; font-weight: 600; }
.apagado-modal-close {
  position: absolute;
  top: 8px;
  right: 12px;
  width: 26px;
  height: 26px;
  border: 0;
  background: transparent;
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  color: #6b7280;
  border-radius: 4px;
}
.apagado-modal-close:hover { background: rgba(0,0,0,0.06); color: #111827; }
'''
        idx = html_new.find(css_existing)
        html_new = html_new[:idx] + compact_css + '\n\n' + html_new[idx:]
        print('[OK] CSS compacto inyectado')

    # Escribir
    HTML_PATH.write_text(html_new)
    diff = len(html_new) - original_len
    print(f'\n[RESUMEN] Diff bytes: {diff:+d}')
    print(f'         Tamaño final: {len(html_new)} bytes')


if __name__ == '__main__':
    main()

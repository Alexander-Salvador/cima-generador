#!/usr/bin/env python3
"""
inject_intel_apagado.py
=======================
Modifica `apps/cima-generador/index.html` de forma ADITIVA:

- NO reemplaza `ZONAS_APAGADAS.ccpp_keys` existente (239 CCPPs validados en prod)
- AGREGA distritos IPT al `districts_with_off` Set (de 21 deptos COMU-5672)
- AGREGA bloque nuevo `ZONAS_APAGADAS_INTEL` con:
    * comunicados (COMU-5514 / COMU-5672)
    * speechs (6 entries con texto completo)
    * flujos_postventa
    * ipt_districts (Set<d6>) para detectar alertas IPT
    * ipt_metadata_by_district (dict por d6)

Agrega la función `showApagadoPopover(check)` que arma un modal con:
- Zona (depto, prov, dist, ccpp)
- Comunicado y versión
- Fecha de apagado
- SMS enviado
- Speech completo (con copia al clipboard)
- Acciones sugeridas (Call Center 1ra línea, Tiendas, Canales Escritos)
- Botón cerrar

Modifica `checkZonaApagada()` para que también marque alertas IPT.

Idempotente: si ya inyectó, reemplaza el bloque.
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
DATA = ROOT / 'apps' / 'cima-generador' / 'data'
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
JSON_PATH = DATA / 'zonas_apagadas.json'


def js_str(s):
    return json.dumps(s, ensure_ascii=False)


def load_ipt_data():
    """Lee JSON v2.0.0 y devuelve distritos IPT + metadata."""
    with open(JSON_PATH) as f:
        d = json.load(f)
    ipt_districts = set()
    ipt_meta = {}  # d6 -> {depto, prov, dist, fecha_apagado, sms, clientes, centros, speech_key}
    for z in d['zones']:
        if z.get('comunicado') != 'COMU-5672':
            continue
        d6 = z.get('ubigeo')
        if not d6:
            continue
        ipt_districts.add(d6)
        # El primero por depto gana para metadata (mismo depto, misma fecha)
        if d6 not in ipt_meta:
            ipt_meta[d6] = {
                'depto': z.get('depto'),
                'prov': z.get('prov'),
                'dist': z.get('dist'),
                'fecha_apagado': z.get('fecha_apagado'),
                'fecha_sms_inicio': z.get('fecha_sms_inicio'),
                'clientes_depto': z.get('clientes_depto'),
                'centros_depto': z.get('centros_depto'),
                'comunicado': z.get('comunicado'),
                'origen_pdf': z.get('origen_pdf'),
                'amplitud': z.get('amplitud'),
            }
    return ipt_districts, ipt_meta, d


def inject_in_html(html: str, ipt_districts: set, ipt_meta: dict, json_data: dict) -> str:
    """
    Inyecta:
    - IPT districts al `districts_with_off` Set existente (preservando entradas originales)
    - Bloque `ZONAS_APAGADAS_INTEL` después del bloque `ZONAS_APAGADAS`
    - Función `showApagadoPopover` antes del IIFE de inicialización
    - Handler `coberturaBtn.onclick` (idempotente)
    """
    # 1) MERGE districts_with_off — agregar IPT districts sin duplicar
    block_pat = re.compile(r'(const ZONAS_APAGADAS = \{.*?^\};)', re.MULTILINE | re.DOTALL)
    m = block_pat.search(html)
    if not m:
        print('[ERROR] bloque ZONAS_APAGADAS no encontrado', file=sys.stderr)
        sys.exit(1)
    old_block = m.group(1)
    # Extraer distritos actuales
    m_dist = re.search(r'districts_with_off: new Set\(\[([^\]]*)\]\)', old_block)
    if not m_dist:
        print('[ERROR] districts_with_off Set no encontrado', file=sys.stderr)
        sys.exit(1)
    existing = re.findall(r'"([0-9]{6})"', m_dist.group(1))
    merged = sorted(set(existing) | ipt_districts)
    # Reemplazar el Set
    new_dist_block = 'districts_with_off: new Set(' + js_str(merged) + ')'
    new_block = re.sub(
        r'districts_with_off: new Set\(\[[^\]]*\]\)',
        new_dist_block,
        old_block,
        count=1
    )
    # Actualizar total_districts si aparece
    new_block = re.sub(
        r'total_districts: \d+',
        f'total_districts: {len(merged)}',
        new_block,
        count=1
    )
    # Actualizar fecha + fuente
    new_block = re.sub(
        r'fecha: "[^"]*"',
        f'fecha: "2026-10-15"',
        new_block,
        count=1
    )
    new_block = re.sub(
        r'fuente: "[^"]*"',
        f'fuente: "Bases de Apagado Movistar + COMU-5672 IPT (regenerado 2026-10-15)"',
        new_block,
        count=1
    )
    # Update total_zones: agregar IPT districts como zones (sin CCPP)
    ipt_count = len(ipt_districts)
    m_total = re.search(r'total_zones: (\d+)', new_block)
    if m_total:
        old_total = int(m_total.group(1))
        new_total = old_total + ipt_count
        new_block = re.sub(
            r'total_zones: \d+',
            f'total_zones: {new_total}',
            new_block,
            count=1
        )
    html_new = html.replace(old_block, new_block, 1)

    # 2) Inyectar bloque ZONAS_APAGADAS_INTEL después de la línea "};" que cierra ZONAS_APAGADAS
    intel_block_marker = 'const ZONAS_APAGADAS_INTEL'
    if intel_block_marker in html_new:
        # Reemplazar bloque existente (idempotencia)
        pat_intel = re.compile(
            r'const ZONAS_APAGADAS_INTEL = \{.*?^\};',
            re.MULTILINE | re.DOTALL
        )
        html_new = pat_intel.sub(build_intel_block(ipt_districts, ipt_meta, json_data), html_new, count=1)
    else:
        # Insertar después del bloque ZONAS_APAGADAS modificado
        pat_after = re.compile(r'(const ZONAS_APAGADAS = \{.*?^\};)', re.MULTILINE | re.DOTALL)
        m_after = pat_after.search(html_new)
        if not m_after:
            print('[ERROR] no se encontró ZONAS_APAGADAS actualizado', file=sys.stderr)
            sys.exit(1)
        html_new = html_new.replace(
            m_after.group(1),
            m_after.group(1) + '\n\n' + build_intel_block(ipt_districts, ipt_meta, json_data),
            1
        )

    # 3) Inyectar showApagadoPopover antes de renderCoberturaButton (o equivalente)
    popover_marker = '// ═══ APAGADO POPOVER (v3.5.0)'
    if popover_marker not in html_new:
        # Insertar antes de "function renderCoberturaButton"
        anchor = '// COBERTURA BUTTON (v2.0.8) — 3 estados según nivel de alerta'
        if anchor in html_new:
            html_new = html_new.replace(
                anchor,
                build_popover_js() + '\n\n' + anchor,
                1
            )
        else:
            print('[WARN] anchor de coberturaBtn no encontrado, popover no inyectado', file=sys.stderr)

    # 4) Inyectar DOM container para el modal (id="apagadoModal")
    modal_marker = '<div id="apagadoModal" class="apagado-modal"'
    if modal_marker not in html_new:
        # Insertar después de <body>
        anchor_body = '<body>'
        if anchor_body in html_new:
            html_new = html_new.replace(
                anchor_body,
                anchor_body + '\n' + build_modal_dom(),
                1
            )

    # 5) CSS para el modal (idempotente)
    if '.apagado-modal {' not in html_new:
        anchor_css_candidates = [
            '/* BOTÓN COBERTURA — Liquid Glass pill (v2.1.4 icon-only) */',
            '/* ══════════════════════════════════════════════════════════════════════\n     BOTÓN COBERTURA — Liquid Glass pill (v2.1.4 icon-only)',
        ]
        css_anchor_found = None
        for cand in anchor_css_candidates:
            if cand in html_new:
                css_anchor_found = cand
                break
        if css_anchor_found:
            html_new = html_new.replace(
                css_anchor_found,
                build_modal_css() + '\n\n' + css_anchor_found,
                1
            )

    return html_new


def build_intel_block(ipt_districts, ipt_meta, json_data):
    speechs = json_data.get('speechs', {})
    comunicados = json_data.get('comunicados', {})
    flujos = json_data.get('flujos_postventa', {})
    ipt_meta_serializable = {
        d6: meta for d6, meta in ipt_meta.items()
    }
    parts = []
    parts.append('// ═══════════════════════════════════════════════════════════════════════')
    parts.append('// ZONAS APAGADAS INTELIGENTE (v3.5.0) — datos extendidos para popover')
    parts.append('// - ipt_districts: Set<d6> para detectar alertas Apagado Zonas IPT (COMU-5672)')
    parts.append('// - ipt_metadata_by_district: { d6: { depto, fecha_apagado, sms, ... } }')
    parts.append('// - comunicados: info de COMU-5514 / COMU-5672 (versión, fecha, q clientes)')
    parts.append('// - speechs: textos completos indexados por key')
    parts.append('// - flujos_postventa: rutas de reclamo, tipificación, baja servicio')
    parts.append('// ═══════════════════════════════════════════════════════════════════════')
    parts.append('const ZONAS_APAGADAS_INTEL = {')
    parts.append('  version: "3.5.0",')
    parts.append('  fecha: "2026-10-15",')
    parts.append('  ipt_districts: new Set(' + js_str(sorted(ipt_districts)) + '),')
    parts.append('  ipt_metadata_by_district: ' + js_str(ipt_meta_serializable) + ',')
    parts.append('  comunicados: ' + js_str(comunicados) + ',')
    parts.append('  speechs: ' + js_str(speechs) + ',')
    parts.append('  flujos_postventa: ' + js_str(flujos) + ',')
    parts.append('};')
    parts.append('')
    return '\n'.join(parts)


def build_popover_js():
    """Función JS para mostrar popover al hacer click en ⚠️."""
    return '''// ═══ APAGADO POPOVER (v3.5.0) — click en ⚠️ → modal inteligente ═══
function showApagadoPopover(currentCheck) {
  if (!currentCheck) return;
  var modal = document.getElementById('apagadoModal');
  if (!modal) return;

  // Lookup metadata
  var meta = null;
  var tipo = currentCheck.nivel;
  var zonaLabel = '';
  var fechaApagado = '';
  var speechKey = '';
  var speechTitulo = '';
  var speechCuerpo = '';
  var comunicadoLabel = '';
  var extraInfo = '';

  if (tipo === 'ccpp_exacto') {
    // Buscar en ZONAS_APAGADAS metadata o rebuild manual
    var code = currentCheck.code;
    zonaLabel = 'CCPP exacto: ' + code;
    speechKey = 'comu5514_postventa_marcaje';
  } else if (tipo === 'distrito_of') {
    // distrito_of puede ser COMU-5514 (CCPPs en distrito) o COMU-5672 (IPT)
    var d6 = currentCheck.code;
    if (typeof ZONAS_APAGADAS_INTEL !== 'undefined' && ZONAS_APAGADAS_INTEL.ipt_districts.has(d6)) {
      speechKey = 'ipt_postventa_4g';
      meta = ZONAS_APAGADAS_INTEL.ipt_metadata_by_district[d6] || null;
      if (meta) {
        zonaLabel = meta.depto + ' (Zona IPT)';
        fechaApagado = meta.fecha_apagado;
        comunicadoLabel = 'COMU-5672 — Apagado Zonas IPT';
        extraInfo = 'SMS enviado: ' + meta.fecha_sms_inicio + ' · Clientes impactados (depto): ' + (meta.clientes_depto||'?') + ' · Centros poblados (depto): ' + (meta.centros_depto||'?');
        speechTitulo = 'Apagado Zonas IPT — ' + meta.depto;
      }
    } else {
      speechKey = 'comu5514_postventa_marcaje';
      zonaLabel = 'Distrito con CCPPs en zonas apagadas (ubigeo: ' + d6 + ')';
      comunicadoLabel = 'COMU-5514 — Apagado Estaciones Base en Centros Poblados';
      fechaApagado = '2026-09-11';
    }
  }

  // Speech con placeholders reemplazados
  var speechData = (typeof ZONAS_APAGADAS_INTEL !== 'undefined' && ZONAS_APAGADAS_INTEL.speechs) ? (ZONAS_APAGADAS_INTEL.speechs[speechKey] || {}) : {};
  var speechBody = speechData.cuerpo || speechData.titulo || 'No hay speech configurado para esta clave.';
  var ddmmMatch = (fechaApagado || '').match(/(\\d{4})-(\\d{2})-(\\d{2})/);
  var dd = ddmmMatch ? ddmmMatch[3] : 'dd';
  var mm = ddmmMatch ? (
    ['enero','febrero','marzo','abril','mayo','junio','julio','agosto','septiembre','octubre','noviembre','diciembre'][parseInt(ddmmMatch[2],10)-1] || 'mes'
  ) : 'mes';
  var zonaNombre = (meta && meta.depto) || zonaLabel;
  var speechFilled = speechBody
    .replace(/\\[díad]+(?:\\/mm)?\\] de \\[mes\\]/g, dd + ' de ' + mm)
    .replace(/\\[d+\\/m+\\] de \\[mes\\]/g, dd + ' de ' + mm)
    .replace(/\\[zona\\]/g, zonaNombre)
    .replace(/\\[día\\] de \\[mes\\]/g, dd + ' de ' + mm);

  speechTitulo = speechTitulo || speechData.titulo || '';

  // Flujo postventa
  var flujo = (typeof ZONAS_APAGADAS_INTEL !== 'undefined' && ZONAS_APAGADAS_INTEL.flujos_postventa) ? (ZONAS_APAGADAS_INTEL.flujos_postventa[speechKey.startsWith('ipt')?'COMU-5672':'COMU-5514'] || {}) : {};

  var html = ''+
    '<div class="apagado-modal-overlay" onclick="closeApagadoPopover(event)">'+
      '<div class="apagado-modal-content" onclick="event.stopPropagation()">'+
        '<button class="apagado-modal-close" onclick="closeApagadoPopover(event)" aria-label="Cerrar">×</button>'+
        '<h2 class="apagado-modal-title">⚠️ Zona de Apagado Movistar</h2>'+
        '<div class="apagado-modal-meta">'+
          '<p><strong>Comunicado:</strong> ' + escapeHtml_(comunicadoLabel || '—') + '</p>'+
          '<p><strong>Zona:</strong> ' + escapeHtml_(zonaLabel) + '</p>'+
          (fechaApagado ? '<p><strong>Fecha de apagado:</strong> <span class="badge badge-danger">' + escapeHtml_(formatDateES(fechaApagado)) + '</span></p>' : '')+
          (extraInfo ? '<p class="apagado-modal-extra">' + escapeHtml_(extraInfo) + '</p>' : '')+
        '</div>'+
        '<div class="apagado-modal-section">'+
          '<h3>📢 Speech del comunicado</h3>'+
          (speechTitulo ? '<p class="apagado-modal-speech-title"><em>' + escapeHtml_(speechTitulo) + '</em></p>' : '')+
          '<blockquote class="apagado-modal-speech" id="apagadoSpeechText">' + escapeHtml_(speechFilled) + '</blockquote>'+
          '<button class="apagado-modal-btn apagado-modal-btn-copy" onclick="copyApagadoSpeech()">📋 Copiar speech</button>'+
          '<span id="apagadoCopyFeedback" class="apagado-modal-copy-feedback"></span>'+
        '</div>'+
        (flujo.reclamo_ruta ? '<div class="apagado-modal-section">'+
          '<h3>🛠️ Si cliente reclama</h3>'+
          '<ul class="apagado-modal-list">'+
            '<li><strong>Detalle reclamo:</strong> ' + escapeHtml_(flujo.reclamo_ruta.detalle||'—') + '</li>'+
            '<li><strong>Motivo:</strong> ' + escapeHtml_(flujo.reclamo_ruta.motivo||'—') + '</li>'+
            '<li><strong>SubMotivo:</strong> ' + escapeHtml_(flujo.reclamo_ruta.submotivo||'—') + '</li>'+
            '<li><strong>Tipificación CRM:</strong> ' + escapeHtml_(flujo.tipificacion_crm?(flujo.tipificacion_crm.motivo+' / '+flujo.tipificacion_crm.submotivo):'—') + '</li>'+
            '<li><strong>Canales Escritos:</strong> <a href="' + escapeHtml_(flujo.reclamo_ruta.canales_escritos_url||'#') + '" target="_blank">Derivar a web</a></li>'+
          '</ul>'+
        '</div>' : '')+
        (flujo.baja_servicio ? '<div class="apagado-modal-section">'+
          '<h3>📞 Si cliente pide BAJA del servicio</h3>'+
          '<ul class="apagado-modal-list">'+
            '<li><strong>Call Center 1ra línea:</strong> ' + escapeHtml_(flujo.baja_servicio.call_center_1ra_linea||'—') + '</li>'+
            '<li><strong>Tiendas:</strong> ' + escapeHtml_(flujo.baja_servicio.tiendas||'—') + '</li>'+
            '<li><strong>Canales Escritos:</strong> ' + escapeHtml_(flujo.baja_servicio.canales_escritos||'—') + '</li>'+
          '</ul>'+
        '</div>' : '')+
        '<div class="apagado-modal-actions">'+
          '<button class="apagado-modal-btn apagado-modal-btn-primary" onclick="closeApagadoPopover(event)">Cerrar</button>'+
        '</div>'+
      '</div>'+
    '</div>';

  modal.innerHTML = html;
  modal.classList.add('show');
  modal.setAttribute('aria-hidden', 'false');
  // Focus trap simple
  var closeBtn = modal.querySelector('.apagado-modal-close');
  if (closeBtn) closeBtn.focus();
}

function closeApagadoPopover(event) {
  if (event) event.stopPropagation();
  var modal = document.getElementById('apagadoModal');
  if (modal) {
    modal.classList.remove('show');
    modal.setAttribute('aria-hidden', 'true');
  }
}

function copyApagadoSpeech() {
  var txt = document.getElementById('apagadoSpeechText');
  if (!txt) return;
  var text = txt.innerText;
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(function() {
      var fb = document.getElementById('apagadoCopyFeedback');
      if (fb) { fb.innerText = '✓ Copiado'; setTimeout(function(){ fb.innerText = ''; }, 2000); }
    });
  } else {
    // Fallback
    var t = document.createElement('textarea');
    t.value = text;
    document.body.appendChild(t);
    t.select();
    document.execCommand('copy');
    document.body.removeChild(t);
    var fb = document.getElementById('apagadoCopyFeedback');
    if (fb) { fb.innerText = '✓ Copiado'; setTimeout(function(){ fb.innerText = ''; }, 2000); }
  }
}

function escapeHtml_(s) {
  if (typeof escapeHtml === 'function') return escapeHtml(s);
  return String(s||'').replace(/[&<>"']/g, function(c) {
    return { '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;' }[c];
  });
}

function formatDateES(iso) {
  if (!iso) return '';
  var m = iso.match(/(\\d{4})-(\\d{2})-(\\d{2})/);
  if (!m) return iso;
  var meses = ['enero','febrero','marzo','abril','mayo','junio','julio','agosto','septiembre','octubre','noviembre','diciembre'];
  return parseInt(m[3],10) + ' de ' + meses[parseInt(m[2],10)-1] + ' de ' + m[1];
}

// Hook: al hacer click en coberturaBtn → mostrar popover si hay alerta
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
});
'''


def build_modal_dom():
    """DOM container para el modal."""
    return '''<!-- APAGADO MODAL (v3.5.0) — contenedor del popover inteligente -->
<div id="apagadoModal" class="apagado-modal" aria-hidden="true" role="dialog" aria-modal="true" aria-labelledby="apagadoModalTitle"></div>
'''


def build_modal_css():
    """CSS para el modal (modal overlay + content + estilos)."""
    return '''/* APAGADO MODAL (v3.5.0) — popover inteligente al click en ⚠️ */
.apagado-modal {
  display: none;
  position: fixed;
  inset: 0;
  z-index: 99999;
}
.apagado-modal.show {
  display: block;
}
.apagado-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  animation: apagadoFadeIn 0.18s ease-out;
}
@keyframes apagadoFadeIn { from {opacity:0} to {opacity:1} }
.apagado-modal-content {
  background: var(--bg-card, #fff);
  color: var(--text-primary, #111827);
  border-radius: 12px;
  max-width: 640px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  padding: 24px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.35);
  position: relative;
  animation: apagadoSlideUp 0.22s ease-out;
}
@keyframes apagadoSlideUp { from {transform: translateY(20px); opacity:0} to {transform:none; opacity:1} }
.apagado-modal-close {
  position: absolute;
  top: 12px;
  right: 16px;
  width: 32px;
  height: 32px;
  border: 0;
  background: transparent;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  color: #6b7280;
  border-radius: 6px;
  transition: background 0.15s;
}
.apagado-modal-close:hover { background: rgba(0,0,0,0.06); color: #111827; }
.apagado-modal-title {
  margin: 0 0 16px 0;
  font-size: 1.35rem;
  font-weight: 700;
  color: #b91c1c;
  padding-right: 32px;
}
.apagado-modal-meta {
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  padding: 12px 14px;
  margin-bottom: 16px;
}
.apagado-modal-meta p { margin: 4px 0; font-size: 0.92rem; }
.apagado-modal-extra {
  color: #6b7280;
  font-size: 0.85rem !important;
  font-style: italic;
}
.apagado-modal-section {
  margin-bottom: 18px;
}
.apagado-modal-section h3 {
  margin: 0 0 8px 0;
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
}
.apagado-modal-speech-title {
  margin: 0 0 6px 0;
  color: #6b7280;
  font-size: 0.88rem;
}
.apagado-modal-speech {
  background: #f9fafb;
  border-left: 4px solid #b91c1c;
  padding: 12px 14px;
  border-radius: 6px;
  margin: 8px 0;
  font-style: italic;
  line-height: 1.5;
  font-size: 0.94rem;
  color: #1f2937;
}
.apagado-modal-list {
  margin: 0;
  padding-left: 18px;
  font-size: 0.92rem;
  line-height: 1.6;
}
.apagado-modal-list li { margin-bottom: 4px; }
.apagado-modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  justify-content: flex-end;
}
.apagado-modal-btn {
  padding: 9px 18px;
  border-radius: 6px;
  font-size: 0.92rem;
  cursor: pointer;
  border: 0;
  font-weight: 500;
  transition: all 0.15s;
}
.apagado-modal-btn-primary {
  background: #2563eb;
  color: #fff;
}
.apagado-modal-btn-primary:hover { background: #1d4ed8; }
.apagado-modal-btn-copy {
  background: #e5e7eb;
  color: #374151;
  margin-top: 8px;
}
.apagado-modal-btn-copy:hover { background: #d1d5db; }
.apagado-modal-copy-feedback {
  color: #059669;
  font-size: 0.85rem;
  margin-left: 10px;
  font-weight: 600;
}
.badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.85rem; font-weight: 600; }
.badge-danger { background: #b91c1c; color: #fff; }
@media (max-width: 480px) {
  .apagado-modal-content { padding: 16px; max-height: 95vh; }
  .apagado-modal-title { font-size: 1.15rem; }
}
'''


def main():
    if not JSON_PATH.exists():
        print(f'[ERROR] {JSON_PATH} no existe. Corre primero regenerate_zonas_apagadas.py', file=sys.stderr)
        sys.exit(1)
    ipt_districts, ipt_meta, json_data = load_ipt_data()
    with open(HTML_PATH) as f:
        html = f.read()
    new_html = inject_in_html(html, ipt_districts, ipt_meta, json_data)
    if new_html == html:
        print('[WARN] ningún cambio aplicado', file=sys.stderr)
    with open(HTML_PATH, 'w') as f:
        f.write(new_html)
    print(f'[OK] {HTML_PATH}:')
    print(f'     IPT districts agregados: {len(ipt_districts)}')
    print(f'     Metadata por district: {len(ipt_meta)}')
    print(f'     Diff bytes: {len(new_html) - len(html):+d}')


if __name__ == '__main__':
    main()

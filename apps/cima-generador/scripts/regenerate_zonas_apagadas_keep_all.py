#!/usr/bin/env python3
"""
regenerate_zonas_apagadas_keep_all.py
======================================
Regenera zonas_apagadas.json v3.0.0 CONSERVANDO los 239 CCPPs
originales del HTML inline (que vienen del regenerate_data.py + manual).

A diferencia de la versión anterior, este script:
  1. Lee los ccpp_keys existentes del HTML inline (239 CCPPs originales)
  2. Lee los manual_fixes
  3. Consolida con IPTs nuevas (21 deptos × sus distritos)
  4. Genera un JSON v3.0.0 COMPLETO con TODOS los CCPPs

Uso:
    python3 scripts/regenerate_zonas_apagadas_keep_all.py

Output:
    apps/cima-generador/data/zonas_apagadas.json
    schema: { version, fecha, fuente, comunicados, speechs, flujos_postventa,
              zones (CCPPs detallados), ccpp_keys, by_ubigeo, by_dist, by_prov,
              by_depto, cobertura }

NO toca: apps/cima-generador/index.html (los datos inline se mantienen)
"""

import json
import re
import unicodedata
from pathlib import Path
from collections import OrderedDict

ROOT = Path(__file__).resolve().parents[3]
DATA = ROOT / 'apps' / 'cima-generador' / 'data'
HTML_PATH = ROOT / 'apps' / 'cima-generador' / 'index.html'
JSON_PATH = DATA / 'zonas_apagadas.json'
UBIGEO_PATH = DATA / 'ubigeo_consolidated.json'
MANUAL_PATH = DATA / 'zonas_apagadas_manual.json'


# ============================================================================
# Constantes (fechas / comunicados)
# ============================================================================

COMU_5514 = {
    'nombre': 'Apagado Estaciones Base en Centros Poblados',
    'comunicado': 'COMU-5514',
    'version': '00003',
    'actualizado': '2026-09-18',
    'fecha_apagado_unica': '2026-09-11',
    'fecha_sms_inicio': '2026-09-04',
    'q_clientes_g1': 1700,
    'q_clientes_g2': 15000,
    'q_clientes_g3': 1700,
    'a_quien_aplica': 'Postpago + Prepago en centros poblados',
}

COMU_5672 = {
    'nombre': 'Apagado de Red en Centros Poblados (Apagados Zonas IPT)',
    'comunicado': 'COMU-5672',
    'version': '00001',
    'fecha_apagado_unica': '2026-08-28',
    'q_clientes': 39500,
    'a_quien_aplica': 'Postpago + Prepago con señal Movistar 2G en zonas rurales',
}

IPT_POR_DEPTO = {
    'AMAZONAS':     {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 1985, 'centros': 9},
    'ANCASH':       {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 657,  'centros': 3},
    'APURIMAC':     {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 6419, 'centros': 33},
    'AREQUIPA':     {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 1032, 'centros': 5},
    'AYACUCHO':     {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 6052, 'centros': 38},
    'CAJAMARCA':    {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 772,  'centros': 7},
    'CUSCO':        {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 8335, 'centros': 40},
    'HUANCAVELICA': {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 7709, 'centros': 50},
    'HUANUCO':      {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 1183, 'centros': 7},
    'ICA':          {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 126,  'centros': 2},
    'JUNIN':        {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 15,   'centros': 2},
    'LA LIBERTAD':  {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 1306, 'centros': 6},
    'LAMBAYEQUE':   {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 221,  'centros': 2},
    'LIMA':         {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 790,  'centros': 8},
    'MADRE DE DIOS':{'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 131,  'centros': 6},
    'MOQUEGUA':     {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 393,  'centros': 3},
    'PASCO':        {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 36,   'centros': 2},
    'PIURA':        {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 2592, 'centros': 18},
    'PUNO':         {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 2051, 'centros': 11},
    'SAN MARTIN':   {'fecha_apagado': '2026-09-03', 'fecha_sms_inicio': '2026-08-18', 'clientes': 1677, 'centros': 8},
    'TACNA':        {'fecha_apagado': '2026-08-28', 'fecha_sms_inicio': '2026-08-17', 'clientes': 246,  'centros': 3},
}


# ============================================================================
# Speechs — los 6 completos con placeholders
# ============================================================================

SPEECHS = {
    'ipt_visor_4g': {
        'titulo': 'APAGADOS ZONAS IPT: Clientes con equipo 4G (CONFIGURACIÓN DE EQUIPO)',
        'cuerpo': ("Queremos informarte que, a partir del [dd] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona]. Estamos implementando mejoras en la "
                   "tecnología 4G para brindarte una mejor experiencia en tu navegación, por lo "
                   "que es necesario que tu equipo esté configurado adecuadamente para soportar "
                   "esta nueva red. ¡No permitas quedarte sin servicio! Configura tu equipo y "
                   "sigue navegando con normalidad."),
        'comunicado': 'COMU-5672',
        'version': '00001',
    },
    'ipt_visor_sin_4g': {
        'titulo': 'APAGADOS ZONAS IPT: Clientes SIN equipo 4G (SUGERIR RENOVACIÓN DE EQUIPO)',
        'cuerpo': ("Queremos informarte que, a partir del [dd/m] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona], por lo que ya no podrás contar con "
                   "equipo 4G compatible con esta red. A partir de la fecha mencionada, tu "
                   "servicio de datos de Movistar quedará limitado a la velocidad de la "
                   "tecnología 2G. Si aún no dispones de un equipo 4G, te invitamos a adquirirlo "
                   "en nuestros puntos de venta Movistar para seguir disfrutando de los beneficios "
                   "de nuestra red."),
        'comunicado': 'COMU-5672',
    },
    'ipt_postventa_4g': {
        'titulo': 'Postventa IPT — Cliente con equipo 4G',
        'cuerpo': ("APAGADOS ZONAS IPT: Clientes con equipo 4G (CONFIGURACIÓN DE EQUIPO): "
                   "Queremos informarte que, a partir del [dd] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona]. Estamos implementando mejoras en la "
                   "tecnología 4G para brindarte una mejor experiencia en tu navegación, por lo "
                   "que es necesario que tu equipo esté configurado adecuadamente para soportar "
                   "esta nueva red. ¡No permitas quedarte sin servicio! Configura tu equipo y "
                   "sigue navegando con normalidad."),
        'comunicado': 'COMU-5672',
    },
    'ipt_postventa_sin_4g': {
        'titulo': 'Postventa IPT — Cliente SIN equipo 4G',
        'cuerpo': ("APAGADOS ZONAS IPT: Clientes SIN equipo 4G (SUGERIR RENOVACIÓN DE EQUIPO): "
                   "Queremos informarte que, a partir del [dd/m] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona], por lo que ya no podrás contar con "
                   "equipo 4G compatible con esta red. A partir de la fecha mencionada, tu "
                   "servicio de datos de Movistar quedará limitado a la velocidad de la "
                   "tecnología 2G. Si aún no dispones de un equipo 4G, te invitamos a adquirirlo "
                   "en nuestros puntos de venta Movistar para seguir disfrutando de los beneficios "
                   "de nuestra red."),
        'comunicado': 'COMU-5672',
    },
    'comu5514_visor_unico': {
        'titulo': 'APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS — Speech único',
        'cuerpo': ("Queremos informarte que, a partir del [día] de [mes], el servicio Movistar "
                   "dejará de funcionar en la zona [zona]. En otras zonas, el servicio seguirá "
                   "disponible con normalidad. Lamentamos las molestias y agradecemos tu "
                   "comprensión."),
        'comunicado': 'COMU-5514',
        'version': '00003',
    },
    'comu5514_postventa_marcaje': {
        'titulo': 'Postventa COMU-5514 — Marcaje',
        'cuerpo': ("APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS: Queremos informarte que, a "
                   "partir del [día] de [mes], el servicio Movistar dejará de funcionar en la "
                   "zona [zona]. En otras zonas, el servicio seguirá disponible con normalidad. "
                   "Lamentamos las molestias y agradecemos tu comprensión."),
        'comunicado': 'COMU-5514',
    },
}


# ============================================================================
# Flujos postventa
# ============================================================================

FLUJOS_POSTVENTA = {
    'COMU-5672': {
        'canal_inicial': 'Call Center 1ra línea, Tiendas Postventa, Canales Escritos',
        'tipificacion_crm': {
            'motivo': 'Internet No Navega',
            'submotivo': 'Configuración equipo Datos 2G/3G / No datos',
            'moviles_no_apn': 'Movistar.pe',
        },
        'reclamo_ruta': {
            'detalle': 'Otros Reclamos Soluciones Móvil',
            'motivo': 'Calidad',
            'submotivo': 'Problemas de cobertura y señal',
            'color_observaciones': 'Color en observaciones — Apagado de zona rural',
            'canales_escritos_url': 'https://www.movistar.com.pe/atencion-al-cliente/reclamos',
        },
        'baja_servicio': {
            'call_center_1ra_linea': 'Transferir a Pool Retenciones',
            'tiendas': 'Atiende según flujo regular de retenciones',
            'canales_escritos': 'Derivar al 104 o tienda',
        },
        'reglas_clave': [
            'Si cliente tiene equipo 4G en zona IPT → speech ipt_visor_4g + orientar configuración APN/equipo',
            'Si cliente NO tiene equipo 4G en zona IPT → speech ipt_visor_sin_4g + sugerir renovación equipo / venta pta',
            'Si cliente NO contacta y está en zona → atender flujo regular (VOZ → CEE → DATOS)',
        ],
    },
    'COMU-5514': {
        'canal_inicial': 'Call Center 1ra línea, Tiendas Postventa, Canales Escritos',
        'reclamo_ruta': {
            'detalle': 'Otros Reclamos Soluciones Móvil',
            'motivo': 'Calidad',
            'submotivo': 'Problemas de cobertura y señal',
            'color_observaciones': 'Color en observaciones — Apagado de zona rural',
            'canales_escritos_url': 'https://www.movistar.com.pe/atencion-al-cliente/reclamos',
        },
        'baja_servicio': {
            'call_center_1ra_linea': 'Transferir a Pool Retenciones',
            'tiendas': 'Atiende según flujo regular de retenciones',
            'canales_escritos': 'Derivar al 104 o tienda',
        },
        'reglas_clave': [
            'Si cliente consulta por comunicado recibido → speech comu5514_visor_unico (con o sin marcaje)',
            'Si cliente NO contacta con señal en zona → Call Center: VOZ → CEE → DATOS',
            'Si cliente pregunta "no hay señal en mi línea" → responder que solo en esa zona no tendrá, otras zonas sí',
            'Si cliente pregunta "puedo navegar con WiFi" → sí, con WiFi navega',
        ],
    },
}


# ============================================================================
# Helpers
# ============================================================================

def normalize(s):
    if not s: return ''
    s = str(s).strip().upper()
    s = re.sub(r'\s+', ' ', s)
    return unicodedata.normalize('NFC', s)


def js_str(s):
    return json.dumps(s, ensure_ascii=False)


def load_ubigeo():
    """Devuelve mapa {depto: {prov: {dist: d6}}}."""
    with open(UBIGEO_PATH) as f:
        u = json.load(f)
    res = {}
    for depto, dv in u.items():
        if not isinstance(dv, dict) or 'provincias' not in dv:
            continue
        res[depto.upper()] = {}
        for prov, pv in dv['provincias'].items():
            if not isinstance(pv, dict) or 'distritos' not in pv:
                continue
            res[depto.upper()][prov.upper()] = {}
            for dist, d_info in pv['distritos'].items():
                d6 = d_info.get('_ubigeo') if isinstance(d_info, dict) else None
                res[depto.upper()][prov.upper()][dist.upper()] = d6
    return res


def extract_inline_zones(html):
    """
    Extrae las zones detalladas del bloque ZONAS_APAGADAS — para CCPPs
    de COMU-5514 conserva los 239 (originado por regenerate_data.py).

    Devuelve:
      - zones: lista de dicts (CCPPs COMU-5514 con grupo + d6 + ccpp)
      - ccpp_keys_set: set de keys "d6:CCPP" exactos
    """
    # Inicializar el runtime del JS
    scope = {}
    ubigeo_src = UBIGEO_PATH.read_text() if UBIGEO_PATH.exists() else ''
    # En vez de evaluar el JS gigante, parseamos regex sobre el HTML
    m = re.search(r'ccpp_keys:\s*new Set\(\[([\s\S]*?)\]\)', html)
    if not m:
        return [], set()
    body = m.group(1)
    items = re.findall(r'"([^"]+)"', body)
    ccpp_keys_set = set(items)
    return [], ccpp_keys_set


# ============================================================================
# Main
# ============================================================================

def main():
    if not JSON_PATH.exists():
        raise SystemExit(f'[ERROR] {JSON_PATH} no existe')

    # Backup
    from shutil import copy
    from datetime import datetime
    backup_dir = ROOT / '.backup-pre-v3'
    backup_dir.mkdir(exist_ok=True)
    timestamp = datetime.now().strftime('%Y%m%d-%H%M%S')
    backup = backup_dir / f'zonas_apagadas.{timestamp}.bak'
    copy(JSON_PATH, backup)
    print(f'[backup] {backup}')

    # Cargar HTML para extraer ccpp_keys originales (239)
    html = HTML_PATH.read_text()
    _, ccpp_keys_set = extract_inline_zones(html)
    print(f'[inline] ccpp_keys originales: {len(ccpp_keys_set)}')

    # Cargar ubigeo para mapear CCPPs de los 239 a {depto, prov, dist}
    ubigeo = load_ubigeo()
    print(f'[ubigeo] deptos: {len(ubigeo)}')

    # Construir mapa d6 → (depto, prov, dist)
    d6_to_names = {}
    for depto, provs in ubigeo.items():
        for prov, dists in provs.items():
            for dist, d6 in dists.items():
                if d6:
                    d6_to_names[d6] = (depto, prov, dist)

    # Cargar manual JSON (correcciones)
    manual_zones = []
    if MANUAL_PATH.exists():
        with open(MANUAL_PATH) as f:
            m_data = json.load(f)
        manual_zones = m_data.get('zonas', [])

    # Construir zones COMU-5514 desde los 239 ccpp_keys
    zones_comu5514 = []
    seen_keys = set()
    for key in ccpp_keys_set:
        if ':' not in key:
            continue
        d6, ccpp = key.split(':', 1)
        if (d6, ccpp) in seen_keys:
            continue
        seen_keys.add((d6, ccpp))
        names = d6_to_names.get(d6, (None, None, None))
        zones_comu5514.append({
            'depto': names[0] or 'UNKNOWN',
            'prov': names[1] or 'UNKNOWN',
            'dist': names[2] or 'UNKNOWN',
            'ccpp': ccpp,
            'ubigeo': d6,
            'comunicado': 'COMU-5514',
            'fecha_apagado': COMU_5514['fecha_apagado_unica'],
            'fecha_sms_inicio': COMU_5514['fecha_sms_inicio'],
            'speech_key': 'comu5514_visor_unico',
            'postventa_speech_key': 'comu5514_postventa_marcaje',
            'zona_nombre': f"{names[2]}, {names[1]}, {names[0]}",
            'origen_pdf': 'COMU-5514 (Bases de Apagado PDFs + manual)',
            'nivel': 'ccpp',
            'grupo': 1 if d6.startswith('06') and '01' <= d6[2:4] <= '12' else (
                     2 if '040' <= d6[:3] <= '060' else (
                     3 if '080' <= d6[:3] <= '100' else 1)),
        })
    print(f'[zones COMU-5514]: {len(zones_comu5514)}')

    # Construir zones COMU-5672 IPT (por depto × distritos)
    zones_comu5672 = []
    for depto, info in IPT_POR_DEPTO.items():
        depto_norm = depto.upper()
        if depto_norm not in ubigeo:
            print(f'[WARN] depto IPT no en ubigeo: {depto}')
            continue
        for prov, dists in ubigeo[depto_norm].items():
            for dist, d6 in dists.items():
                if not d6:
                    continue
                zones_comu5672.append({
                    'depto': depto_norm,
                    'prov': prov,
                    'dist': dist,
                    'ccpp': None,
                    'ubigeo': d6,
                    'grupo': 0,
                    'nivel': 'departamento_ipt',
                    'comunicado': 'COMU-5672',
                    'fecha_apagado': info['fecha_apagado'],
                    'fecha_sms_inicio': info['fecha_sms_inicio'],
                    'clientes_depto': info['clientes'],
                    'centros_depto': info['centros'],
                    'speech_key': 'ipt_visor_4g',
                    'postventa_speech_key': 'ipt_postventa_4g',
                    'zona_nombre': f"{dist}, {prov}, {depto_norm} (Zona IPT)",
                    'origen_pdf': 'Apagado Zonas IPT.pdf',
                    'origen_datos': 'COMU-5672 Zonas IPT — tabular por depto',
                    'amplitud': 'departamento',
                })
    print(f'[zones COMU-5672]: {len(zones_comu5672)}')

    all_zones = zones_comu5514 + zones_comu5672

    # Cobertura
    by_ubigeo = sorted({z['ubigeo'] for z in all_zones if z.get('ubigeo')})
    by_dist = sorted({f"{z['depto']}|{z['prov']}|{z['dist']}" for z in all_zones})
    by_prov = sorted({f"{z['depto']}|{z['prov']}" for z in all_zones})
    by_depto = sorted({z['depto'] for z in all_zones if z.get('depto')})

    out = {
        'version': '3.0.0',
        'fecha_generacion': '2026-10-15',
        'fuente': [
            'regenerate_data.py (PDFs Bases Apagado + manual_fixes) — 239 CCPPs COMU-5514',
            'Apagado Zonas IPT.pdf — 21 deptos × distritos (COMU-5672)',
        ],
        'total_zones': len(all_zones),
        'cobertura': {
            'ubigeo_6d': len(by_ubigeo),
            'distritos': len(by_dist),
            'provincias': len(by_prov),
            'departamentos': len(by_depto),
        },
        'comunicados': {
            'COMU-5514': COMU_5514,
            'COMU-5672': COMU_5672,
        },
        'speechs': SPEECHS,
        'flujos_postventa': FLUJOS_POSTVENTA,
        'by_ubigeo': by_ubigeo,
        'by_dist': by_dist,
        'by_prov': by_prov,
        'by_depto': by_depto,
        'ccpp_keys': sorted(ccpp_keys_set),
        'districts_with_off': by_ubigeo,
        'zones': all_zones,
    }

    JSON_PATH.write_text(json.dumps(out, ensure_ascii=False, indent=2))
    print(f'\n[OK] {JSON_PATH}')
    print(f'     version: v3.0.0')
    print(f'     total zones: {len(all_zones)}')
    print(f'     ccpp_keys: {len(ccpp_keys_set)}')
    print(f'     coverage: {out["cobertura"]}')


if __name__ == '__main__':
    main()

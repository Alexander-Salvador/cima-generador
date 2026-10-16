#!/usr/bin/env python3
"""
regenerate_zonas_apagadas.py
============================
Regenera `apps/cima-generador/data/zonas_apagadas.json` con:

- Campos enriquecidos por zone (`comunicado`, `fecha_apagado`, `fecha_sms_inicio`,
  `speech_key`, `postventa_speech_key`, `zona_nombre`, `origen_pdf`)
- Speechs completos por comunicado en bloque `speechs`
- Zonas IPT (COMU-5672) nuevas: 21 deptos × distritos capitales provinciales
- Mantiene estructura backward-compat (`by_ubigeo`, `by_dist`, `by_prov`,
  `by_depto`, `zones`, `ccpp_keys` que el JS espera).

Uso:
    python3 scripts/regenerate_zonas_apagadas.py

Lee:
    apps/cima-generador/data/zonas_apagadas.json   (anterior)
    apps/cima-generador/data/ubigeo_consolidated.json

Escribe:
    apps/cima-generador/data/zonas_apagadas.json
"""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
DATA_DIR = ROOT / 'apps' / 'cima-generador' / 'data'
JSON_IN = DATA_DIR / 'zonas_apagadas.json'
UBIGEO = DATA_DIR / 'ubigeo_consolidated.json'
JSON_OUT = JSON_IN

# -----------------------------------------------------------------------------
# DATOS DE PDFs (extraídos visualmente)
# -----------------------------------------------------------------------------

# COMU-5514 — Apagado Estaciones Base Centros Poblados
COMU_5514 = {
    'nombre': 'Apagado Estaciones Base en Centros Poblados',
    'comunicado': 'COMU-5514',
    'version': '00003',
    'actualizado': '2026-09-18',
    'fecha_apagado_unica': '2026-09-11',  # inicio campaña
    'fecha_sms_inicio': '2026-09-04',     # ~1 semana antes
    'q_clientes': {'grupo1': 1700, 'grupo2': 15000, 'grupo3': 1700},
    'a_quien_aplica': 'Postpago + Prepago en centros poblados',
}

# COMU-5672 — Apagado Zonas IPT (Residencial Postpago)
COMU_5672 = {
    'nombre': 'Apagado de Red en Centros Poblados (Apagados Zonas IPT)',
    'comunicado': 'COMU-5672',
    'version': '00001',
    'fecha_apagado_unica': '2026-08-28',
    'q_clientes': 39500,
    'a_quien_aplica': 'Postpago + Prepago con señal Movistar 2G en zonas rurales',
}
# Por depto COMU-5672 (fecha apagado + fecha SMS inicio)
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

# -----------------------------------------------------------------------------
# SPEECHS
# -----------------------------------------------------------------------------
SPEECHS = {
    'ipt_visor_4g': {
        'titulo': 'APAGADOS ZONAS IPT: Clientes con equipo 4G (CONFIGURACIÓN DE EQUIPO)',
        'cuerpo': ("Queremos informarte que, a partir del [dd] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona]. Estamos implementando mejoras en la "
                   "tecnología 4G para brindarte una mejor experiencia en tu navegación, por lo "
                   "que es necesario que tu equipo esté configurado adecuadamente para soportar "
                   "esta nueva red. ¡No permitas quedarte sin servicio! Configura tu equipo y "
                   "sigue navegando con normalidad."),
        'nota': 'Orientar en la configuración del equipo en red 4G — REVISAR COMUNICADO APAGADOS ZONAS IPT',
        'aplicable_a': 'Cliente con equipo 4G en zona IPT — speech visor (4.1)',
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
        'nota': 'Cliente no podrá navegar — REVISAR COMUNICADO APAGADOS ZONAS IPT',
        'aplicable_a': 'Cliente SIN equipo 4G en zona IPT — speech visor (4.2)',
        'comunicado': 'COMU-5672',
        'version': '00001',
    },
    'ipt_postventa_4g': {
        'titulo': 'Postventa IPT — Cliente con equipo 4G (marcaje 4.1)',
        'cuerpo': ("APAGADOS ZONAS IPT: Clientes con equipo 4G (CONFIGURACIÓN DE EQUIPO): "
                   "Queremos informarte que, a partir del [dd] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona]. Estamos implementando mejoras en la "
                   "tecnología 4G para brindarte una mejor experiencia en tu navegación, por lo "
                   "que es necesario que tu equipo esté configurado adecuadamente para soportar "
                   "esta nueva red. ¡No permitas quedarte sin servicio! Configura tu equipo y "
                   "sigue navegando con normalidad."),
        'aplicable_a': 'Postventa: cliente con marcaje visor (4.1) en zona IPT',
        'comunicado': 'COMU-5672',
    },
    'ipt_postventa_sin_4g': {
        'titulo': 'Postventa IPT — Cliente SIN equipo 4G (marcaje 4.2)',
        'cuerpo': ("APAGADOS ZONAS IPT: Clientes SIN equipo 4G (SUGERIR RENOVACIÓN DE EQUIPO): "
                   "Queremos informarte que, a partir del [dd/m] de [mes], el servicio Movistar 2G "
                   "dejará de funcionar en la zona [zona], por lo que ya no podrás contar con "
                   "equipo 4G compatible con esta red. A partir de la fecha mencionada, tu "
                   "servicio de datos de Movistar quedará limitado a la velocidad de la "
                   "tecnología 2G. Si aún no dispones de un equipo 4G, te invitamos a adquirirlo "
                   "en nuestros puntos de venta Movistar para seguir disfrutando de los beneficios "
                   "de nuestra red."),
        'aplicable_a': 'Postventa: cliente con marcaje visor (4.2) en zona IPT — sugerir renovación',
        'comunicado': 'COMU-5672',
    },
    'comu5514_visor_unico': {
        'titulo': 'APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS — Speech único visor',
        'cuerpo': ("Queremos informarte que, a partir del [día] de [mes], el servicio Movistar "
                   "dejará de funcionar en la zona [zona]. En otras zonas, el servicio seguirá "
                   "disponible con normalidad. Lamentamos las molestias y agradecemos tu "
                   "comprensión."),
        'nota': 'REVISAR COMUNICADO APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS',
        'aplicable_a': 'Consulta por comunicación recibida (con o sin marcaje) — Grupos 1/2/3',
        'comunicado': 'COMU-5514',
        'version': '00003',
        'actualizado': '2026-09-18',
    },
    'comu5514_postventa_marcaje': {
        'titulo': 'Postventa COMU-5514 — Cliente con marcaje',
        'cuerpo': ("APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS: Queremos informarte que, a "
                   "partir del [día] de [mes], el servicio Movistar dejará de funcionar en la "
                   "zona [zona]. En otras zonas, el servicio seguirá disponible con normalidad. "
                   "Lamentamos las molestias y agradecemos tu comprensión."),
        'aplicable_a': 'Postventa: cliente con marcaje visor en zona centros poblados apagados',
        'comunicado': 'COMU-5514',
    },
}

# -----------------------------------------------------------------------------
# FLUJOS POSTVENTA — guiados para el asesor
# -----------------------------------------------------------------------------
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


def load_ubigeo_distritos():
    """Devuelve {DEPTO: {prov: {dist: d6}}} para lookup de IPT."""
    with open(UBIGEO) as f:
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


def main():
    # Cargar consolidado previo
    with open(JSON_IN) as f:
        prev = json.load(f)

    # 1. Enriquecer zones existentes (COMU-5514)
    zones_enriq = []
    for z in prev['zones']:
        new_z = dict(z)
        new_z['comunicado'] = 'COMU-5514'
        new_z['fecha_apagado'] = COMU_5514['fecha_apagado_unica']
        new_z['fecha_sms_inicio'] = COMU_5514['fecha_sms_inicio']
        new_z['speech_key'] = 'comu5514_visor_unico'
        new_z['postventa_speech_key'] = 'comu5514_postventa_marcaje'
        new_z['zona_nombre'] = f"{z['dist']}, {z['prov']}, {z['depto']}"
        # v3.6.0+L-051: Mapeo PDF→Grupo corregido.
        # APAGADO.pdf (3 pág texto) = Grupo 2 (23 deptos, 200+ CCPPs).
        # APAGADP1.pdf (1 pág imagen) = Grupo 1 (6 deptos, 16 CCPPs).
        # APAGADO2.pdf (1 pág imagen) = Grupo 3 (5 deptos, 21 CCPPs).
        pdf_map = {1: 'APAGADP1.pdf', 2: 'APAGADO.pdf', 3: 'APAGADO2.pdf'}
        new_z['origen_pdf'] = pdf_map.get(z.get('grupo'), 'desconocido')
        zones_enriq.append(new_z)

    # 2. Agregar zones IPT (COMU-5672) — TODOS los distritos del depto.
    # El PDF COMU-5672 dice: "Aplica para todos los clientes móvil que abarquen
    # las zonas rurales que brinda señal a través de terceros" — Tabular por depto.
    # Estrategia: marcar TODOS los distritos del depto para que cualquier consulta
    # desde un distrito de un depto con IPT alerte correctamente.
    ubigeo = load_ubigeo_distritos()
    ipt_added = 0
    for depto, info in IPT_POR_DEPTO.items():
        depto_norm = depto.upper()
        if depto_norm not in ubigeo:
            print(f'[WARN] depto IPT no encontrado en ubigeo: {depto}')
            continue
        for prov, dists in ubigeo[depto_norm].items():
            for dist, d6 in dists.items():
                if not d6:
                    continue
                zone = {
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
                    'origen_datos': 'COMU-5672 — tabular por depto (PDF sin CCPP granular)',
                    'amplitud': 'departamento',
                }
                zones_enriq.append(zone)
                ipt_added += 1
    print(f'[OK] {ipt_added} zonas IPT (COMU-5672) agregadas (todos los distritos del depto)')

    # 3. Recalcular coverage / by_* a partir de zones_enriq
    by_ubigeo = sorted({z['ubigeo'] for z in zones_enriq if z.get('ubigeo')})
    by_dist_set = set()
    by_prov_set = set()
    by_depto_set = set()
    for z in zones_enriq:
        if z.get('depto') and z.get('dist'):
            by_dist_set.add(f"{z['depto']}|{z['prov']}|{z['dist']}")
        if z.get('depto') and z.get('prov'):
            by_prov_set.add(f"{z['depto']}|{z['prov']}")
        if z.get('depto'):
            by_depto_set.add(z['depto'])

    # 4. ccpp_keys (para lookup JS) — incluye CCPPs reales + distrito-level IPT
    ccpp_keys = set()
    for z in zones_enriq:
        d6 = z.get('ubigeo')
        ccpp = z.get('ccpp')
        if d6 and ccpp:
            ccpp_keys.add(f"{d6}:{ccpp}".upper().replace('  ', ' '))

    out = {
        'version': '2.0.0',
        'fecha_generacion': '2026-10-15',
        'fuente': [
            'PDFs Apagado Estaciones Base Centros Poblados (APAGADO.pdf, APAGADO2.pdf, APAGADP1.pdf)',
            'PDFs Apagado Zonas IPT (Apagado Zonas IPT.pdf)',
            'COMU-5514 v00003 (actualizado 2026-09-18)',
            'COMU-5672 v00001',
        ],
        'total_zones': len(zones_enriq),
        'cobertura': {
            'ubigeo_6d': len(by_ubigeo),
            'distritos': len(by_dist_set),
            'provincias': len(by_prov_set),
            'departamentos': len(by_depto_set),
        },
        'comunicados': {
            'COMU-5514': COMU_5514,
            'COMU-5672': COMU_5672,
        },
        'speechs': SPEECHS,
        'flujos_postventa': FLUJOS_POSTVENTA,
        'by_ubigeo': by_ubigeo,
        'by_dist': sorted(by_dist_set),
        'by_prov': sorted(by_prov_set),
        'by_depto': sorted(by_depto_set),
        'ccpp_keys': sorted(ccpp_keys),
        'zones': zones_enriq,
    }

    # Escribir (preservar campos previos si no se sobreescribieron)
    with open(JSON_OUT, 'w') as f:
        json.dump(out, f, ensure_ascii=False, indent=2)

    print(f'[OK] {JSON_OUT} escrito:')
    print(f'     version={out["version"]}')
    print(f'     total_zones={out["total_zones"]}')
    print(f'     cobertura={out["cobertura"]}')
    print(f'     comunicados={list(out["comunicados"].keys())}')
    print(f'     speechs={len(out["speechs"])}')
    print(f'     flujos_postventa={list(out["flujos_postventa"].keys())}')


if __name__ == '__main__':
    main()

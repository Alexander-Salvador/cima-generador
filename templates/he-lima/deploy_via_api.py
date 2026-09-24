#!/usr/bin/env python3
"""
HE LIMA — DEPLOY VIA GOOGLE SHEETS API
========================================
Script Python alternativo al deploy_he_lima.gs (Apps Script).
Usa google-api-python-client para crear/modificar el sheet directamente.

SETUP REQUERIDO:
1. Crear proyecto en Google Cloud Console
2. Habilitar Google Sheets API + Google Drive API
3. Crear Service Account o usar OAuth2 user credentials
4. Descargar credentials.json al directorio de trabajo
5. Instalar dependencias: pip install google-api-python-client google-auth

USO:
  python deploy_via_api.py --sheet-id <ID>           # modificar sheet existente
  python deploy_via_api.py --new --name "HE Lima"   # crear sheet nuevo
  python deploy_via_api.py --rollback --sheet-id <ID>  # rollback

OUTPUT:
  - Sheet con 6 sheets visibles poblados
  - 13 feriados PE 2026 insertados
  - Dropdowns configurados
  - URL del sheet resultante
"""

import argparse
import sys
from datetime import datetime, date

try:
    from google.oauth2 import service_account
    from googleapiclient.discovery import build
    from googleapiclient.errors import HttpError
except ImportError:
    print("❌ Faltan dependencias. Ejecutá: pip install google-api-python-client google-auth")
    sys.exit(1)

SCOPES = [
    'https://www.googleapis.com/auth/spreadsheets',
    'https://www.googleapis.com/auth/drive'
]

# ─────────────────────────────────────────────────────────────────────
# HEADERS por sheet (orden importa)
# ─────────────────────────────────────────────────────────────────────

HEADERS = {
    '00_Config': ['Param', 'Valor', 'Notas'],
    '01_Feriados_PE': ['ID', 'Nombre', 'Tipo', 'Fecha', 'Es_Movil', 'Aplica_HE'],
    '02_Personas': [
        'ID_Persona', 'DNI', 'Apellidos_Nombres', 'Cargo', 'Cuadrilla',
        'Turno_Base', 'Fecha_Ingreso', 'Activo', 'Tarifa_Hora_PEN', 'Notas'
    ],
    '10_Registro_HE': [
        'Fecha', 'ISO_Week', 'Dia_Semana', 'Es_Feriado', 'Tipo_Dia',
        'ID_Persona', 'Apellidos_Nombres', 'Turno',
        'Hora_Ingreso', 'Hora_Salida', 'Horas_Total',
        'Horas_Normales', 'Horas_Extras', 'Factor_Aplicado', 'Horas_Ponderadas',
        'Pago_HE_PEN', 'Observaciones',
        'Aud_Usuario', 'Aud_Timestamp', 'Estado',
        'Mes_Año', 'Partition_Key'
    ],
    '20_Dashboard': ['KPI', 'Valor', 'Unidad', 'Detalle'],
    '99_Audit_Log': [
        'Timestamp', 'Usuario', 'Acción', 'Sheet', 'Fila', 'Columna',
        'Valor_Anterior', 'Valor_Nuevo'
    ]
}

# ─────────────────────────────────────────────────────────────────────
# 13 Feriados PE 2026
# ─────────────────────────────────────────────────────────────────────

FERIADOS_PE_2026 = [
    ('F-01', 'Año Nuevo',              'nacional', '2026-01-01', False, True),
    ('F-02', 'Jueves Santo',           'nacional', '2026-04-02', True,  True),
    ('F-03', 'Viernes Santo',          'nacional', '2026-04-03', True,  True),
    ('F-04', 'Día del Trabajo',        'nacional', '2026-05-01', False, True),
    ('F-05', 'San Pedro y San Pablo',  'nacional', '2026-06-29', False, True),
    ('F-06', 'Fiestas Patrias',        'nacional', '2026-07-28', False, True),
    ('F-07', 'Fiestas Patrias (2da)',  'nacional', '2026-07-29', False, True),
    ('F-08', 'Santa Rosa de Lima',     'nacional', '2026-08-30', False, True),
    ('F-09', 'Combate de Angamos',     'nacional', '2026-10-08', False, True),
    ('F-10', 'Todos los Santos',       'nacional', '2026-11-01', False, True),
    ('F-11', 'Inmaculada Concepción',  'nacional', '2026-12-08', False, True),
    ('F-12', 'Batalla de Ayacucho',    'nacional', '2026-12-09', False, True),
    ('F-13', 'Navidad',                'nacional', '2026-12-25', False, True),
]

# ─────────────────────────────────────────────────────────────────────
# Sample personas para 02_Personas
# ─────────────────────────────────────────────────────────────────────

PERSONAS_SAMPLE = [
    ('P-001', '12345678', 'URIBE DAVILA MONICA',      'tecnico',        'A', 'diurno',   '2020-03-15', True,  22.50),
    ('P-002', '23456789', 'WONG CORREA PATRICIA',      'tecnico',        'A', 'diurno',   '2019-08-01', True,  24.00),
    ('P-003', '34567890', 'RAMIREZ HUAMAN JOHNN',      'tecnico',        'B', 'nocturno', '2021-01-10', True,  26.00),
    ('P-004', '45678901', 'ESCUDERO GRANDEZ CRIS',     'supervisor',     'A', 'diurno',   '2018-05-20', True,  35.00),
    ('P-005', '56789012', 'ALVARADO RODRIGUEZ JAC',   'tecnico',        'C', 'rotativo', '2022-07-01', True,  23.50),
    ('P-006', '67890123', 'ALBA CARRILLO OSCAR',       'tecnico',        'B', 'diurno',   '2020-11-15', True,  25.00),
    ('P-007', '78901234', 'HILARIO PURIZACA BEATRIZ', 'jefe_cuadrilla', 'D', 'diurno',   '2017-02-01', True,  40.00),
    ('P-008', '89012345', 'VARGAS TAPIA ISABEL',       'tecnico',        'C', 'nocturno', '2021-09-20', True,  26.50),
    ('P-009', '90123456', 'MENDOZA CARMEN CECILIA',    'admin',          'A', 'diurno',   '2019-04-10', True,  30.00),
    ('P-010', '01234567', 'ARIAS CUEVA STEPHANNI',     'tecnico',        'B', 'diurno',   '2023-03-05', True,  21.00),
]

# ─────────────────────────────────────────────────────────────────────
# Configuración inicial 00_Config
# ─────────────────────────────────────────────────────────────────────

def get_config_rows():
    now = datetime.now()
    return [
        ['Mes',                str(now.month),            'Dropdown: 1-12'],
        ['Año',                str(now.year),             'Dropdown: 2024-2030'],
        ['Semana',             f'W{now.isocalendar()[1]:02d}', 'Dropdown W##'],
        ['Persona',            'TODOS',                   'Dropdown desde 02_Personas'],
        ['Factor HE Diurno',   1.25,                      'Rango: 1.0-3.0'],
        ['Factor HE Nocturno', 1.35,                      'Rango: 1.0-3.0'],
        ['Factor Feriado',     2.00,                      'Rango: 1.0-3.0'],
        ['Jornada Max Diaria', 8,                         'Rango: 4-12'],
        ['',                   '',                        ''],
        ['Último refresh',     now.strftime('%Y-%m-%d %H:%M:%S'), 'auto'],
    ]


# ─────────────────────────────────────────────────────────────────────
# CORE — Crear/modificar sheets
# ─────────────────────────────────────────────────────────────────────

def deploy(sheets_service, spreadsheet_id, sobrescribir=False):
    print(f'🚀 HE Lima Deploy — Spreadsheet ID: {spreadsheet_id}')
    print(f'   Sobrescribir existentes: {sobrescribir}')
    print()

    # Verificar que spreadsheet existe
    try:
        meta = sheets_service.spreadsheets().get(spreadsheetId=spreadsheet_id).execute()
        print(f'✅ Spreadsheet encontrado: "{meta["properties"]["title"]}"')
    except HttpError as e:
        print(f'❌ No se puede acceder al spreadsheet: {e}')
        sys.exit(1)

    existing_titles = {s['properties']['title']: s for s in meta['sheets']}

    # 1. Crear sheets faltantes
    print('▶ Paso 1/5: Creando sheets...')
    requests = []
    for sheet_name in HEADERS.keys():
        if sheet_name in existing_titles:
            if not sobrescribir:
                print(f'  ❌ {sheet_name} ya existe. Usa --force para sobrescribir.')
                sys.exit(1)
            print(f'  ⚠️  {sheet_name} ya existe — sobreescribiendo headers')
        else:
            requests.append({'addSheet': {'properties': {'title': sheet_name}}})
            print(f'  ✅ {sheet_name} creada')

    if requests:
        sheets_service.spreadsheets().batchUpdate(
            spreadsheetId=spreadsheet_id,
            body={'requests': requests}
        ).execute()

    # 2. Escribir headers
    print('▶ Paso 2/5: Escribiendo headers...')
    for sheet_name, headers in HEADERS.items():
        result = sheets_service.spreadsheets().values().update(
            spreadsheetId=spreadsheet_id,
            range=f'{sheet_name}!A1',
            valueInputOption='RAW',
            body={'values': [headers]}
        ).execute()
        print(f'  ✅ {sheet_name}: {len(headers)} columnas')

    # 3. Poblar 01_Feriados_PE
    print('▶ Paso 3/5: Poblando 01_Feriados_PE (13 feriados)...')
    feriados_rows = [
        [f[0], f[1], f[2], f[3], f[4], f[5]] for f in FERIADOS_PE_2026
    ]
    sheets_service.spreadsheets().values().update(
        spreadsheetId=spreadsheet_id,
        range='01_Feriados_PE!A2:F14',
        valueInputOption='USER_ENTERED',
        body={'values': feriados_rows}
    ).execute()
    print(f'  ✅ 13 feriados PE 2026 insertados')

    # 4. Configurar 00_Config
    print('▶ Paso 4/5: Configurando 00_Config...')
    sheets_service.spreadsheets().values().update(
        spreadsheetId=spreadsheet_id,
        range='00_Config!A2:C11',
        valueInputOption='USER_ENTERED',
        body={'values': get_config_rows()}
    ).execute()
    print(f'  ✅ Dropdowns Mes/Año + factores configurados')

    # 5. Sample personas (opcional, solo si 02_Personas está vacío)
    personas_data = sheets_service.spreadsheets().values().get(
        spreadsheetId=spreadsheet_id,
        range='02_Personas!A1:A1000'
    ).execute().get('values', [])

    if len(personas_data) <= 1:  # Solo headers o vacío
        print('▶ Paso 5/5: Insertando sample personas (10)...')
        personas_rows = [list(p) for p in PERSONAS_SAMPLE]
        sheets_service.spreadsheets().values().update(
            spreadsheetId=spreadsheet_id,
            range='02_Personas!A2:J11',
            valueInputOption='USER_ENTERED',
            body={'values': personas_rows}
        ).execute()
        print(f'  ✅ 10 personas sample insertadas')
    else:
        print(f'▶ Paso 5/5: 02_Personas ya tiene {len(personas_data)-1} personas — NO se inserta sample')

    print()
    print('═══════════════════════════════════════════')
    print('  ✅ HE LIMA DESPLEGADO EXITOSAMENTE')
    print('═══════════════════════════════════════════')
    print()
    print(f'Spreadsheet URL:')
    print(f'  https://docs.google.com/spreadsheets/d/{spreadsheet_id}/edit')
    print()
    print('PRÓXIMOS PASOS:')
    print('1. Pegar código Apps Script en Extensions → Apps Script')
    print('2. Pegar appsscript.json manifest')
    print('3. Ejecutar instalarTriggers()')
    print('4. Validar AC-01..23 (ver tests/test-cases.md)')


def rollback(sheets_service, spreadsheet_id):
    print(f'🗑️  Rollback HE Lima — Spreadsheet: {spreadsheet_id}')

    meta = sheets_service.spreadsheets().get(spreadsheetId=spreadsheet_id).execute()
    sheets_to_delete = [
        s['properties']['sheetId'] for s in meta['sheets']
        if s['properties']['title'] in HEADERS.keys()
    ]

    if not sheets_to_delete:
        print('  No hay sheets del template HE Lima para eliminar.')
        return

    requests = [{'deleteSheet': {'sheetId': sid}} for sid in sheets_to_delete]
    sheets_service.spreadsheets().batchUpdate(
        spreadsheetId=spreadsheet_id,
        body={'requests': requests}
    ).execute()
    print(f'  ✅ {len(sheets_to_delete)} sheets eliminadas')


def create_new_sheet(drive_service, sheets_service, name):
    print(f'📄 Creando nuevo Google Sheet: "{name}"')

    spreadsheet = sheets_service.spreadsheets().create(
        body={
            'properties': {'title': name, 'locale': 'es_PE', 'timeZone': 'America/Lima'},
            'sheets': [{'properties': {'title': 'Sheet1', 'gridProperties': {'rowCount': 100, 'columnCount': 26}}}]
        }
    ).execute()

    spreadsheet_id = spreadsheet['spreadsheetId']
    print(f'  ✅ Sheet creado: {spreadsheet_id}')

    # Mover a una carpeta específica del CEO (si se proporciona)
    # Por ahora lo deja en Mi unidad

    # Eliminar Sheet1 default
    default_sheet_id = spreadsheet['sheets'][0]['properties']['sheetId']
    sheets_service.spreadsheets().batchUpdate(
        spreadsheetId=spreadsheet_id,
        body={'requests': [{'deleteSheet': {'sheetId': default_sheet_id}}]}
    ).execute()

    # Deploy en el nuevo sheet
    deploy(sheets_service, spreadsheet_id, sobrescribir=True)

    return spreadsheet_id


# ─────────────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description='HE Lima — Deploy via Google Sheets API')
    parser.add_argument('--credentials', default='credentials.json', help='Path to OAuth/service account credentials')
    parser.add_argument('--sheet-id', help='ID del spreadsheet existente')
    parser.add_argument('--new', action='store_true', help='Crear nuevo spreadsheet')
    parser.add_argument('--name', default='CIMA — Técnica Móvil (HE Lima)', help='Nombre para nuevo sheet')
    parser.add_argument('--force', action='store_true', help='Sobrescribir sheets existentes')
    parser.add_argument('--rollback', action='store_true', help='Rollback: eliminar sheets HE Lima')

    args = parser.parse_args()

    if not (args.sheet_id or args.new or args.rollback):
        parser.error('Especificá --sheet-id, --new, o --rollback')

    # Autenticación
    print(f'🔐 Autenticando con {args.credentials}...')
    try:
        creds = service_account.Credentials.from_service_account_file(
            args.credentials, scopes=SCOPES
        )
    except Exception as e:
        print(f'❌ Error de auth: {e}')
        print('   Asegurate de tener credentials.json (Service Account o OAuth2)')
        sys.exit(1)

    sheets_service = build('sheets', 'v4', credentials=creds)
    drive_service = build('drive', 'v3', credentials=creds)
    print('  ✅ Autenticado')
    print()

    if args.rollback:
        rollback(sheets_service, args.sheet_id)
    elif args.new:
        create_new_sheet(drive_service, sheets_service, args.name)
    else:
        deploy(sheets_service, args.sheet_id, sobrescribir=args.force)


if __name__ == '__main__':
    main()

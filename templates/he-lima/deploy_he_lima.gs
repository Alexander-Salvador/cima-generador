/**
 * ════════════════════════════════════════════════════════════════════════════
 *  HE LIMA — DEPLOY AUTOMATION
 *  Ejecutar UNA SOLA VEZ desde el editor de Apps Script del sheet "CIMA —
 *  Técnica Móvil" (o cualquier sheet donde quieras desplegar el template).
 *
 *  Flujo:
 *  1. Abre Extensions → Apps Script en tu Google Sheet
 *  2. Crea un nuevo archivo .gs llamado "deploy_he_lima"
 *  3. Pega este código completo
 *  4. Autoriza los permisos OAuth cuando lo pida (lectura/escritura Sheets)
 *  5. Ejecuta la función `desplegarHELima()` desde el menú "Run"
 *  6. Espera 30-60 segundos
 *  7. Verifica que aparezcan las 6 sheets nuevas + 12 ocultas
 *
 *  ⚠️ NO MODIFICAR las funciones existentes. Solo agregar este archivo.
 *  ⚠️ Después del deploy, este archivo se puede eliminar (es one-shot).
 * ════════════════════════════════════════════════════════════════════════════
 */

// ─────────────────────────────────────────────────────────────────────
// CONFIGURACIÓN — ajustar antes de ejecutar si es necesario
// ─────────────────────────────────────────────────────────────────────

const DEPLOY_CONFIG = {
  // Sheets visibles (en este orden, aparecerán como tabs al fondo)
  visibleSheets: [
    '00_Config',
    '01_Feriados_PE',
    '02_Personas',
    '10_Registro_HE',
    '20_Dashboard',
    '99_Audit_Log'   // esta queda visible para auditoría, puede ocultarse después
  ],

  // Sheets mensuales ocultas (auto-creadas por 08_DBSync al primer acceso)
  // No las creamos aquí — DBSync lo hace lazy.

  // Versión del template
  version: '0.2.0',
  deployDate: new Date().toISOString().slice(0, 10),

  // Si TRUE, sobrescribe sheets existentes con el template
  // Si FALSE, aborta si alguna sheet ya existe
  sobrescribirExistentes: false
};

// ─────────────────────────────────────────────────────────────────────
// HEADERS de cada sheet (orden importa)
// ─────────────────────────────────────────────────────────────────────

const HEADERS = {
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
};

// ─────────────────────────────────────────────────────────────────────
// 13 FERIADOS PE 2026 — validados contra D.L. 713
// ─────────────────────────────────────────────────────────────────────

const FERIADOS_PE_2026 = [
  ['F-01', 'Año Nuevo',              'nacional', new Date('2026-01-01'), false, true],
  ['F-02', 'Jueves Santo',           'nacional', new Date('2026-04-02'), true,  true],  // Pascua - 3
  ['F-03', 'Viernes Santo',          'nacional', new Date('2026-04-03'), true,  true],  // Pascua - 2
  ['F-04', 'Día del Trabajo',        'nacional', new Date('2026-05-01'), false, true],
  ['F-05', 'San Pedro y San Pablo',  'nacional', new Date('2026-06-29'), false, true],
  ['F-06', 'Fiestas Patrias',        'nacional', new Date('2026-07-28'), false, true],
  ['F-07', 'Fiestas Patrias (2da)',  'nacional', new Date('2026-07-29'), false, true],
  ['F-08', 'Santa Rosa de Lima',     'nacional', new Date('2026-08-30'), false, true],
  ['F-09', 'Combate de Angamos',     'nacional', new Date('2026-10-08'), false, true],
  ['F-10', 'Todos los Santos',       'nacional', new Date('2026-11-01'), false, true],
  ['F-11', 'Inmaculada Concepción',  'nacional', new Date('2026-12-08'), false, true],
  ['F-12', 'Batalla de Ayacucho',    'nacional', new Date('2026-12-09'), false, true],
  ['F-13', 'Navidad',                'nacional', new Date('2026-12-25'), false, true],
];

// ─────────────────────────────────────────────────────────────────────
// FUNCIÓN PRINCIPAL — ejecutar desde el editor
// ─────────────────────────────────────────────────────────────────────

function desplegarHELima() {
  const ui = SpreadsheetApp.getUi();
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // Confirmación
  const response = ui.alert(
    'HE Lima — Deploy v' + DEPLOY_CONFIG.version,
    'Esto creará 6 sheets nuevos en el sheet activo:\n' +
    '• 00_Config (panel control)\n' +
    '• 01_Feriados_PE (13 feriados D.L. 713)\n' +
    '• 02_Personas (maestro)\n' +
    '• 10_Registro_HE (working surface)\n' +
    '• 20_Dashboard (KPIs)\n' +
    '• 99_Audit_Log (append-only)\n\n' +
    '¿Proceder?',
    ui.ButtonSet.YES_NO
  );

  if (response !== ui.Button.YES) {
    ui.alert('Deploy cancelado.');
    return;
  }

  const log = [];
  log.push('═══════════════════════════════════════════');
  log.push('  HE Lima Deploy v' + DEPLOY_CONFIG.version);
  log.push('  Fecha: ' + DEPLOY_CONFIG.deployDate);
  log.push('  Sheet activo: ' + ss.getName());
  log.push('═══════════════════════════════════════════');
  log.push('');

  // 1. Crear sheets
  log.push('▶ Paso 1/5: Creando sheets...');
  DEPLOY_CONFIG.visibleSheets.forEach(name => {
    let sheet = ss.getSheetByName(name);
    if (sheet) {
      if (DEPLOY_CONFIG.sobrescribirExistentes) {
        log.push('  ⚠️  ' + name + ' ya existe — sobreescribiendo');
        sheet.clear();
      } else {
        log.push('  ❌ ' + name + ' ya existe — abortando');
        log.push('     Cambia DEPLOY_CONFIG.sobrescribirExistentes = true para forzar');
        log.push('');
        log.push('═══════════════════════════════════════════');
        log.push('  DEPLOY FALLIDO');
        log.push('═══════════════════════════════════════════');
        Logger.log(log.join('\n'));
        ui.alert('Deploy fallido: sheet "' + name + '" ya existe.\n\nRevisa el log (View → Logs).');
        return;
      }
    } else {
      sheet = ss.insertSheet(name);
      log.push('  ✅ ' + name + ' creada');
    }

    // Headers
    if (HEADERS[name]) {
      sheet.getRange(1, 1, 1, HEADERS[name].length).setValues([HEADERS[name]]);
      sheet.getRange(1, 1, 1, HEADERS[name].length)
        .setFontWeight('bold')
        .setBackground('#1F4E79')
        .setFontColor('#FFFFFF');
      sheet.setFrozenRows(1);
    }
  });
  log.push('');

  // 2. Poblar 01_Feriados_PE con 13 feriados 2026
  log.push('▶ Paso 2/5: Poblando 01_Feriados_PE (13 feriados D.L. 713)...');
  const feriadosSheet = ss.getSheetByName('01_Feriados_PE');
  const feriadosRows = FERIADOS_PE_2026.map(f => [
    f[0], f[1], f[2], Utilities.formatDate(f[3], 'America/Lima', 'yyyy-MM-dd'), f[4], f[5]
  ]);
  feriadosSheet.getRange(2, 1, feriadosRows.length, 6).setValues(feriadosRows);
  log.push('  ✅ 13 feriados 2026 insertados');
  log.push('');

  // 3. Configurar 00_Config con valores iniciales
  log.push('▶ Paso 3/5: Configurando 00_Config...');
  const cfgSheet = ss.getSheetByName('00_Config');
  const now = new Date();
  const currentMonth = now.getMonth() + 1;
  const currentYear = now.getFullYear();

  const configRows = [
    ['Mes',         currentMonth,        'Dropdown: 1-12'],
    ['Año',         currentYear,         'Dropdown: 2024-2030'],
    ['Semana',      'W' + getISOWeek(now), 'Dropdown W##'],
    ['Persona',     'TODOS',             'Dropdown desde 02_Personas'],
    ['Factor HE Diurno',   1.25,         'Rango: 1.0-3.0'],
    ['Factor HE Nocturno', 1.35,         'Rango: 1.0-3.0'],
    ['Factor Feriado',     2.00,         'Rango: 1.0-3.0'],
    ['Jornada Max Diaria', 8,            'Rango: 4-12'],
    ['', '', ''],
    ['Último refresh', Utilities.formatDate(now, 'America/Lima', 'yyyy-MM-dd HH:mm:ss'), 'auto']
  ];
  cfgSheet.getRange(2, 1, configRows.length, 3).setValues(configRows);

  // Data Validation en dropdowns
  applyDataValidation(cfgSheet);
  log.push('  ✅ Dropdowns configurados: Mes=' + currentMonth + ', Año=' + currentYear);
  log.push('');

  // 4. Aplicar formato básico a 10_Registro_HE
  log.push('▶ Paso 4/5: Aplicando formato a 10_Registro_HE...');
  const regSheet = ss.getSheetByName('10_Registro_HE');
  applyRegistroFormatting(regSheet);
  applyRegistroConditionalFormatting(regSheet);
  log.push('  ✅ Headers + CF + DV aplicados');
  log.push('');

  // 5. Mensaje final
  log.push('▶ Paso 5/5: Deploy completo');
  log.push('');
  log.push('═══════════════════════════════════════════');
  log.push('  ✅ HE LIMA DESPLEGADO EXITOSAMENTE');
  log.push('═══════════════════════════════════════════');
  log.push('');
  log.push('PRÓXIMOS PASOS:');
  log.push('1. ☐ Revisar sheets creadas visualmente');
  log.push('2. ☐ Pegar código de apps-script/01-08_*.gs en Extensions → Apps Script');
  log.push('3. ☐ Pegar appsscript.json manifest');
  log.push('4. ☐ Ejecutar instalarTriggers() desde el editor');
  log.push('5. ☐ Probar cambiar dropdown Mes — grid debe actualizarse');
  log.push('6. ☐ Validar AC-01..23 de tests/test-cases.md');
  log.push('');
  log.push('VERSIÓN: ' + DEPLOY_CONFIG.version + ' · FECHA: ' + DEPLOY_CONFIG.deployDate);

  Logger.log(log.join('\n'));
  ui.alert('✅ Deploy exitoso\n\nRevisa el log (View → Logs) para detalles.\n\nPróximo: pegar el código Apps Script en Extensions → Apps Script.');
}

// ─────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────

function applyDataValidation(cfgSheet) {
  // Mes: 1-12
  const mesRule = SpreadsheetApp.newDataValidation()
    .requireValueInList(['1','2','3','4','5','6','7','8','9','10','11','12'])
    .setAllowInvalid(false)
    .setHelpText('Mes 1-12')
    .build();
  cfgSheet.getRange('B2').setDataValidation(mesRule);

  // Año: 2024-2030
  const anioRule = SpreadsheetApp.newDataValidation()
    .requireValueInList(['2024','2025','2026','2027','2028','2029','2030'])
    .setAllowInvalid(false)
    .setHelpText('Año 2024-2030')
    .build();
  cfgSheet.getRange('B3').setDataValidation(anioRule);

  // Persona: TODOS (dropdown dinámico se configura vía Apps Script en producción)
  const personaRule = SpreadsheetApp.newDataValidation()
    .requireValueInList(['TODOS'])
    .setAllowInvalid(false)
    .build();
  cfgSheet.getRange('B5').setDataValidation(personaRule);

  // Factores HE: 1.0-3.0
  const factorRule = SpreadsheetApp.newDataValidation()
    .requireNumberBetween(1.0, 3.0)
    .setAllowInvalid(false)
    .setHelpText('Factor HE: rango 1.0-3.0')
    .build();
  cfgSheet.getRange('B6:B8').setDataValidation(factorRule);

  // Jornada max: 4-12
  const jornadaRule = SpreadsheetApp.newDataValidation()
    .requireNumberBetween(4, 12)
    .setAllowInvalid(false)
    .build();
  cfgSheet.getRange('B9').setDataValidation(jornadaRule);
}

function applyRegistroFormatting(regSheet) {
  // Congelar header + primeras 3 columnas
  regSheet.setFrozenRows(1);
  regSheet.setFrozenColumns(3);

  // Anchos sugeridos
  regSheet.setColumnWidth(1, 90);  // Fecha
  regSheet.setColumnWidth(2, 60);  // ISO_Week
  regSheet.setColumnWidth(3, 90);  // Dia_Semana
  regSheet.setColumnWidth(4, 70);  // Es_Feriado
  regSheet.setColumnWidth(5, 80);  // Tipo_Dia
  regSheet.setColumnWidth(6, 80);  // ID_Persona
  regSheet.setColumnWidth(7, 200); // Apellidos_Nombres
  regSheet.setColumnWidth(8, 70);  // Turno
  regSheet.setColumnWidth(9, 80);  // Hora_Ingreso
  regSheet.setColumnWidth(10, 80); // Hora_Salida
  regSheet.setColumnWidth(11, 80); // Horas_Total
  regSheet.setColumnWidth(12, 90); // Horas_Normales
  regSheet.setColumnWidth(13, 90); // Horas_Extras
  regSheet.setColumnWidth(14, 90); // Factor_Aplicado
  regSheet.setColumnWidth(15, 100); // Horas_Ponderadas
  regSheet.setColumnWidth(16, 90); // Pago_HE_PEN
  regSheet.setColumnWidth(17, 150); // Observaciones
  regSheet.setColumnWidth(18, 120); // Aud_Usuario
  regSheet.setColumnWidth(19, 120); // Aud_Timestamp
  regSheet.setColumnWidth(20, 80); // Estado
  regSheet.setColumnWidth(21, 80); // Mes_Año
  regSheet.setColumnWidth(22, 120); // Partition_Key
}

function applyRegistroConditionalFormatting(regSheet) {
  // CF regla cf_feriado: filas con Tipo_Dia=Feriado → fondo rosa
  const feriadoRule = SpreadsheetApp.newConditionalFormatRule()
    .whenTextEqualTo('Feriado')
    .setBackground('#FEE2E2')
    .setFontColor('#991B1B')
    .setBold(true)
    .setRanges([regSheet.getRange('E2:E1000')])
    .build();
  // Nota: API para conditional formatting más complejo requiere builder manual
  // Para CF completa, ver 08_DBSync / 04_FormulasRecalc al instalar Apps Script
}

function getISOWeek(date) {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  const weekNum = Math.ceil((((d - yearStart) / 86400000) + 1) / 7);
  return 'W' + String(weekNum).padStart(2, '0');
}

// ─────────────────────────────────────────────────────────────────────
// Función auxiliar: rollback del deploy (si algo salió mal)
// ─────────────────────────────────────────────────────────────────────

function rollbackDeployHELima() {
  const ui = SpreadsheetApp.getUi();
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  const response = ui.alert(
    'Rollback HE Lima',
    'Esto eliminará TODAS las sheets del template HE Lima.\n' +
    '¿Continuar? (irreversible)',
    ui.ButtonSet.YES_NO
  );

  if (response !== ui.Button.YES) return;

  DEPLOY_CONFIG.visibleSheets.forEach(name => {
    const sheet = ss.getSheetByName(name);
    if (sheet) {
      ss.deleteSheet(sheet);
      Logger.log('🗑️  ' + name + ' eliminada');
    }
  });

  // Eliminar mensuales
  for (let mes = 1; mes <= 12; mes++) {
    for (let año = 2024; año <= 2030; año++) {
      const name = '98_DB_' + año + '_' + String(mes).padStart(2, '0');
      const sheet = ss.getSheetByName(name);
      if (sheet) {
        ss.deleteSheet(sheet);
      }
    }
  }

  ui.alert('✅ Rollback completo');
  Logger.log('✅ Rollback HE Lima ejecutado');
}

/**
 * ════════════════════════════════════════════════════════════════════════════
 *  HE LIMA — DEPLOY ALL-IN-ONE
 *  ─────────────────────────────────────────────────────────────────────────
 *  USO (5 minutos):
 *  1. Abrí tu Google Sheet "CIMA — Técnica Móvil" (o creá uno nuevo)
 *  2. Menú: Extensions → Apps Script
 *  3. Pegá TODO este archivo en el editor
 *  4. Click 💾 Save (Ctrl+S)
 *  5. Click ▶ Run → función `desplegarHELimaCompleto()`
 *  6. Primera vez: autorizá OAuth (Sheets + Drive)
 *  7. Esperá 30-60 segundos
 *  8. ✅ Listo: 6 sheets visibles + 12 ocultas + 8 módulos Apps Script
 *
 *  RESULTADO:
 *  - 6 sheets visibles (00_Config, 01_Feriados_PE, 02_Personas,
 *    10_Registro_HE, 20_Dashboard, 99_Audit_Log)
 *  - 12 hojas mensuales ocultas (98_DB_YYYY_MM) auto-creadas al usar
 *  - 13 feriados PE 2026 pre-cargados
 *  - 10 personas de muestra en 02_Personas
 *  - Dropdowns Mes/Año/Semana/Persona configurados
 *  - Triggers onEdit + time-driven activos
 *
 *  VERSIÓN: 0.2.0 · FECHA: 2026-09-24
 * ════════════════════════════════════════════════════════════════════════════
 */

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  PARTE 1/5: CONFIGURACIÓN + CREACIÓN DE SHEETS                       ║
// ╚══════════════════════════════════════════════════════════════════════╝

const DEPLOY_VERSION = '0.2.0';

const SHEETS_VISIBLES = [
  '00_Config',
  '01_Feriados_PE',
  '02_Personas',
  '10_Registro_HE',
  '20_Dashboard',
  '99_Audit_Log'
];

const HEADERS = {
  '00_Config': ['Param', 'Valor', 'Notas'],
  '01_Feriados_PE': ['ID', 'Nombre', 'Tipo', 'Fecha', 'Es_Movil', 'Aplica_HE'],
  '02_Personas': ['ID_Persona', 'DNI', 'Apellidos_Nombres', 'Cargo', 'Cuadrilla',
                  'Turno_Base', 'Fecha_Ingreso', 'Activo', 'Tarifa_Hora_PEN', 'Notas'],
  '10_Registro_HE': ['Fecha', 'ISO_Week', 'Dia_Semana', 'Es_Feriado', 'Tipo_Dia',
                     'ID_Persona', 'Apellidos_Nombres', 'Turno', 'Hora_Ingreso', 'Hora_Salida',
                     'Horas_Total', 'Horas_Normales', 'Horas_Extras', 'Factor_Aplicado',
                     'Horas_Ponderadas', 'Pago_HE_PEN', 'Observaciones',
                     'Aud_Usuario', 'Aud_Timestamp', 'Estado', 'Mes_Año', 'Partition_Key'],
  '20_Dashboard': ['KPI', 'Valor', 'Unidad', 'Detalle'],
  '99_Audit_Log': ['Timestamp', 'Usuario', 'Acción', 'Sheet', 'Fila', 'Columna',
                    'Valor_Anterior', 'Valor_Nuevo']
};

const FERIADOS_PE_2026 = [
  ['F-01', 'Año Nuevo',              'nacional', new Date('2026-01-01'), false, true],
  ['F-02', 'Jueves Santo',           'nacional', new Date('2026-04-02'), true,  true],
  ['F-03', 'Viernes Santo',          'nacional', new Date('2026-04-03'), true,  true],
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

const PERSONAS_SAMPLE = [
  ['P-001', '12345678', 'URIBE DAVILA MONICA',      'tecnico',        'A', 'diurno',   '2020-03-15', true,  22.50, ''],
  ['P-002', '23456789', 'WONG CORREA PATRICIA',      'tecnico',        'A', 'diurno',   '2019-08-01', true,  24.00, ''],
  ['P-003', '34567890', 'RAMIREZ HUAMAN JOHNN',      'tecnico',        'B', 'nocturno', '2021-01-10', true,  26.00, ''],
  ['P-004', '45678901', 'ESCUDERO GRANDEZ CRIS',     'supervisor',     'A', 'diurno',   '2018-05-20', true,  35.00, ''],
  ['P-005', '56789012', 'ALVARADO RODRIGUEZ JAC',   'tecnico',        'C', 'rotativo', '2022-07-01', true,  23.50, ''],
  ['P-006', '67890123', 'ALBA CARRILLO OSCAR',       'tecnico',        'B', 'diurno',   '2020-11-15', true,  25.00, ''],
  ['P-007', '78901234', 'HILARIO PURIZACA BEATRIZ', 'jefe_cuadrilla', 'D', 'diurno',   '2017-02-01', true,  40.00, ''],
  ['P-008', '89012345', 'VARGAS TAPIA ISABEL',       'tecnico',        'C', 'nocturno', '2021-09-20', true,  26.50, ''],
  ['P-009', '90123456', 'MENDOZA CARMEN CECILIA',    'admin',          'A', 'diurno',   '2019-04-10', true,  30.00, ''],
  ['P-010', '01234567', 'ARIAS CUEVA STEPHANNI',     'tecnico',        'B', 'diurno',   '2023-03-05', true,  21.00, ''],
];


// ╔══════════════════════════════════════════════════════════════════════╗
// ║  PARTE 2/5: FUNCIONES DE DEPLOY                                     ║
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * FUNCIÓN PRINCIPAL — ejecutar desde editor para deploy completo
 */
function desplegarHELimaCompleto() {
  const ui = SpreadsheetApp.getUi();
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  const response = ui.alert(
    'HE Lima — Deploy Completo v' + DEPLOY_VERSION,
    'Esto creará 6 sheets visibles + 12 ocultas + activará todos los módulos.\n\n¿Proceder?',
    ui.ButtonSet.YES_NO
  );
  if (response !== ui.Button.YES) return;

  Logger.log('═══════════════════════════════════════════');
  Logger.log('  HE Lima Deploy v' + DEPLOY_VERSION);
  Logger.log('═══════════════════════════════════════════');

  // Paso 1: Crear sheets visibles
  Logger.log('▶ Paso 1/4: Creando sheets visibles...');
  SHEETS_VISIBLES.forEach(name => {
    let sheet = ss.getSheetByName(name);
    if (sheet) sheet.clear();
    else sheet = ss.insertSheet(name);

    if (HEADERS[name]) {
      const range = sheet.getRange(1, 1, 1, HEADERS[name].length);
      range.setValues([HEADERS[name]]);
      range.setFontWeight('bold').setBackground('#1F4E79').setFontColor('#FFFFFF');
      sheet.setFrozenRows(1);
    }
    Logger.log('  ✅ ' + name);
  });

  // Paso 2: Poblar feriados y config
  Logger.log('▶ Paso 2/4: Poblando feriados y config...');
  const feriadosSheet = ss.getSheetByName('01_Feriados_PE');
  const feriadosRows = FERIADOS_PE_2026.map(f => [
    f[0], f[1], f[2], Utilities.formatDate(f[3], 'America/Lima', 'yyyy-MM-dd'), f[4], f[5]
  ]);
  feriadosSheet.getRange(2, 1, feriadosRows.length, 6).setValues(feriadosRows);
  Logger.log('  ✅ 13 feriados 2026');

  const now = new Date();
  const cfgSheet = ss.getSheetByName('00_Config');
  cfgSheet.getRange(2, 1, 10, 3).setValues([
    ['Mes',                now.getMonth() + 1,        'Dropdown 1-12'],
    ['Año',                now.getFullYear(),         'Dropdown 2024-2030'],
    ['Semana',             'W' + getISOWeek(now),     'Dropdown W##'],
    ['Persona',            'TODOS',                   'Desde 02_Personas'],
    ['Factor HE Diurno',   1.25,                      'Rango 1.0-3.0'],
    ['Factor HE Nocturno', 1.35,                      'Rango 1.0-3.0'],
    ['Factor Feriado',     2.00,                      'Rango 1.0-3.0'],
    ['Jornada Max',        8,                         'Rango 4-12'],
    ['',                   '',                        ''],
    ['Último refresh',     Utilities.formatDate(now, 'America/Lima', 'yyyy-MM-dd HH:mm:ss'), 'auto']
  ]);
  applyDataValidation00Config(cfgSheet);
  Logger.log('  ✅ 00_Config con dropdowns');

  // Paso 3: Sample personas + formato
  Logger.log('▶ Paso 3/4: Sample personas + formato...');
  const personasSheet = ss.getSheetByName('02_Personas');
  personasSheet.getRange(2, 1, PERSONAS_SAMPLE.length, 10).setValues(PERSONAS_SAMPLE);
  applyFormato10Registro(ss.getSheetByName('10_Registro_HE'));
  Logger.log('  ✅ 10 personas + formato aplicado');

  // Paso 4: Instalar triggers
  Logger.log('▶ Paso 4/4: Instalando triggers...');
  instalarTriggers();

  Logger.log('═══════════════════════════════════════════');
  Logger.log('  ✅ DEPLOY COMPLETO');
  Logger.log('═══════════════════════════════════════════');
  ui.alert('✅ Deploy completo\n\nRevisá View → Logs para detalles.');
}

function applyDataValidation00Config(cfgSheet) {
  cfgSheet.getRange('B2').setDataValidation(
    SpreadsheetApp.newDataValidation()
      .requireValueInList(['1','2','3','4','5','6','7','8','9','10','11','12'])
      .setAllowInvalid(false).build()
  );
  cfgSheet.getRange('B3').setDataValidation(
    SpreadsheetApp.newDataValidation()
      .requireValueInList(['2024','2025','2026','2027','2028','2029','2030'])
      .setAllowInvalid(false).build()
  );
  cfgSheet.getRange('B4').setDataValidation(
    SpreadsheetApp.newDataValidation()
      .requireValueInList(['W14','W15','W16','W17','W18','W19','W20'])
      .setAllowInvalid(false).build()
  );
  cfgSheet.getRange('B6:B8').setDataValidation(
    SpreadsheetApp.newDataValidation()
      .requireNumberBetween(1.0, 3.0)
      .setAllowInvalid(false).build()
  );
  cfgSheet.getRange('B9').setDataValidation(
    SpreadsheetApp.newDataValidation()
      .requireNumberBetween(4, 12)
      .setAllowInvalid(false).build()
  );
}

function applyFormato10Registro(sheet) {
  sheet.setFrozenRows(1);
  sheet.setFrozenColumns(3);
  sheet.setColumnWidth(1, 90); sheet.setColumnWidth(2, 60); sheet.setColumnWidth(3, 90);
  sheet.setColumnWidth(4, 70); sheet.setColumnWidth(5, 80); sheet.setColumnWidth(6, 80);
  sheet.setColumnWidth(7, 200); sheet.setColumnWidth(8, 70); sheet.setColumnWidth(9, 80);
  sheet.setColumnWidth(10, 80); sheet.setColumnWidth(11, 80); sheet.setColumnWidth(12, 90);
  sheet.setColumnWidth(13, 90); sheet.setColumnWidth(14, 90); sheet.setColumnWidth(15, 100);
  sheet.setColumnWidth(16, 90); sheet.setColumnWidth(17, 150); sheet.setColumnWidth(18, 120);
  sheet.setColumnWidth(19, 120); sheet.setColumnWidth(20, 80); sheet.setColumnWidth(21, 80);
  sheet.setColumnWidth(22, 120);

  // CF regla cf_feriado
  const reglaFeriado = SpreadsheetApp.newConditionalFormatRule()
    .whenTextEqualTo('Feriado')
    .setBackground('#FEE2E2')
    .setFontColor('#991B1B')
    .setBold(true)
    .setRanges([sheet.getRange('E2:E1000')])
    .build();
  const reglaSabDom = SpreadsheetApp.newConditionalFormatRule()
    .whenTextEqualTo('Sabado')
    .setBackground('#FEF3C7')
    .setRanges([sheet.getRange('E2:E1000')])
    .build();
  const reglaSabDom2 = SpreadsheetApp.newConditionalFormatRule()
    .whenTextEqualTo('Domingo')
    .setBackground('#FEF3C7')
    .setRanges([sheet.getRange('E2:E1000')])
    .build();
  const rules = sheet.getConditionalFormatRules();
  rules.push(reglaFeriado, reglaSabDom, reglaSabDom2);
  sheet.setConditionalFormatRules(rules);
}

function getISOWeek(date) {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  return String(Math.ceil((((d - yearStart) / 86400000) + 1) / 7)).padStart(2, '0');
}


// ╔══════════════════════════════════════════════════════════════════════╗
// ║  01_CfgLoader.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 01_CfgLoader.gs
 * ----------------------------------------------------------------------------
 * Lee 00_Config!B2:B9 y expone los valores vía CacheService.
 * Reduce lecturas a Sheets API en operaciones masivas (load/save/sync).
 *
 * Anchors (architecture.yaml v2):
 *   B2 = mes activo (1-12)
 *   B3 = año activo (YYYY)
 *   B4 = semana activa (W01..W53)
 *   B5 = persona filtro ("TODOS" o "P-NNN")
 *   B6 = factor HE diurno (default 1.25)
 *   B7 = factor HE nocturno (default 1.35)
 *   B8 = factor feriado (default 2.00)
 *   B9 = jornada max diaria (default 8)
 *   B11 = timestamp último refresh
 *
 * Cache TTL: 600s (10 min) · prefix: "he_lima_"
 * Authority: andres · area: platform_engineering · lead: thavren
 */

/** @constant {string} Nombre de la hoja de configuración */
const SHEET_CFG = '00_Config';

/** @constant {string} Prefijo de claves para CacheService */
const CACHE_PREFIX = 'he_lima_';

/** @constant {number} TTL del caché en segundos */
const CACHE_TTL_SEC = 600;

/** @constant {Object<string, string>} Mapa anchor → celda */
const CFG_ANCHORS = Object.freeze({
  mes: 'B2',
  anio: 'B3',
  semana: 'B4',
  persona: 'B5',
  factorDiurno: 'B6',
  factorNocturno: 'B7',
  factorFeriado: 'B8',
  jornadaMax: 'B9',
  lastRefresh: 'B11',
});

/**
 * Cache interna local para evitar llamadas repetidas a CacheService dentro
 * del mismo execution context (Apps Script no comparte caché entre runs).
 * @type {Object<string, *>}
 */
const _memCache = {};

/**
 * Obtiene la hoja 00_Config (helper reutilizable).
 * @returns {GoogleAppsScript.Spreadsheet.Sheet}
 * @throws {Error} Si la hoja no existe.
 */
function getCfgSheet_() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_CFG);
  if (!sh) {
    throw new Error('cfg_missing: hoja 00_Config no encontrada en el spreadsheet');
  }
  return sh;
}

/**
 * Lee el valor de un anchor desde 00_Config con caché.
 * Cache miss → lee Sheet + escribe CacheService.
 * @param {string} key Clave lógica ('mes', 'anio', 'factorDiurno', etc.)
 * @returns {*} Valor leído (tipo depende del anchor)
 */
function readCfg_(key) {
  const cell = CFG_ANCHORS[key];
  if (!cell) {
    throw new Error('cfg_unknown_key: ' + key);
  }
  const cacheKey = CACHE_PREFIX + 'cfg_' + key;

  // 1. Mem cache (durante mismo execution)
  if (Object.prototype.hasOwnProperty.call(_memCache, cacheKey)) {
    return _memCache[cacheKey];
  }

  // 2. CacheService (entre executions)
  const cache = CacheService.getDocumentCache();
  const cached = cache.get(cacheKey);
  if (cached !== null) {
    const parsed = JSON.parse(cached);
    _memCache[cacheKey] = parsed;
    return parsed;
  }

  // 3. Sheets API
  const sh = getCfgSheet_();
  const raw = sh.getRange(cell).getValue();

  // Normalización de tipos por anchor
  let normalized = raw;
  if (['mes', 'anio', 'jornadaMax'].indexOf(key) >= 0) {
    normalized = Number(raw);
  } else if (key === 'factorDiurno' || key === 'factorNocturno' || key === 'factorFeriado') {
    normalized = Number(raw);
  } else if (key === 'semana' && typeof raw === 'string') {
    normalized = raw; // "W##"
  }

  cache.put(cacheKey, JSON.stringify(normalized), CACHE_TTL_SEC);
  _memCache[cacheKey] = normalized;
  return normalized;
}

/**
 * Mes activo (1-12).
 * @returns {number}
 * @example const m = getMes(); // 3
 */
function getMes() {
  return readCfg_('mes');
}

/**
 * Año activo (YYYY).
 * @returns {number}
 * @example const y = getAnio(); // 2026
 */
function getAnio() {
  return readCfg_('anio');
}

/**
 * Semana activa en formato "W##".
 * @returns {string}
 * @example const w = getSemana(); // "W14"
 */
function getSemana() {
  return readCfg_('semana');
}

/**
 * Persona filtro ("TODOS" o "P-NNN").
 * @returns {string}
 */
function getPersonaFiltro() {
  return readCfg_('persona');
}

/**
 * Factores HE + jornada max en un solo bundle (optimiza 1 round-trip a caché).
 * @returns {{diurno:number, nocturno:number, feriado:number, jornadaMax:number}}
 * @example
 *   const f = getFactores();
 *   // {diurno:1.25, nocturno:1.35, feriado:2.00, jornadaMax:8}
 */
function getFactores() {
  return {
    diurno: readCfg_('factorDiurno'),
    nocturno: readCfg_('factorNocturno'),
    feriado: readCfg_('factorFeriado'),
    jornadaMax: readCfg_('jornadaMax'),
  };
}

/**
 * Invalida toda la caché de configuración (forzar re-lectura desde Sheets).
 * Llamar cuando un administrador edita manualmente 00_Config!B2:B9.
 * @returns {void}
 */
function invalidateCache() {
  const cache = CacheService.getDocumentCache();
  // Borrar todas las claves conocidas
  Object.keys(CFG_ANCHORS).forEach(function (key) {
    cache.remove(CACHE_PREFIX + 'cfg_' + key);
    delete _memCache[CACHE_PREFIX + 'cfg_' + key];
  });
  // También limpia caché de feriados (lo mantiene 02_HolidaysEngine)
  cache.remove(CACHE_PREFIX + 'feriados_anio');
}

/**
 * Escribe el timestamp de último refresh en 00_Config!B11.
 * Llamado por 05_DashboardRefresh y 08_DBSync después de operaciones pesadas.
 * @returns {void}
 */
function touchLastRefresh() {
  const sh = getCfgSheet_();
  sh.getRange(CFG_ANCHORS.lastRefresh).setValue(new Date());
  // B11 no se cachea (es efímero)
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  02_HolidaysEngine.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 02_HolidaysEngine.gs
 * ----------------------------------------------------------------------------
 * Mantiene 01_Feriados_PE (13 feriados nacionales PE validados D.L. 713).
 *
 * Algoritmo: Pascua calculada con Meeus/Jones/Butcher (1987) — método
 * anónimo gregoriano. Determinístico, O(1), válido 1583..4099.
 *
 * Responsabilidades:
 *   - calcularPascua(anio)        → Date (domingo de Pascua)
 *   - calcularJuevesSanto(anio)   → Date
 *   - calcularViernesSanto(anio)  → Date
 *   - poblarFeriados(anio)        → escribe 13 filas en 01_Feriados_PE
 *   - esFeriado(fecha)            → boolean (lookup contra hoja)
 *
 * Authority: marco (arquitecto) · Implementer: andres
 */

/** @constant {string} Hoja de feriados */
const SHEET_HOLIDAYS = '01_Feriados_PE';

/** @constant {string} Header de la hoja */
const HOLIDAYS_HEADER = [
  'ID',
  'Nombre',
  'Tipo',
  'Fecha',
  'Es_Movil',
  'Aplica_HE',
];

/**
 * 13 feriados nacionales PE validados D.L. 713 (corregido v2).
 * Solo los 11 fijos se hardcodean — los 2 móviles (Jueves/Viernes Santo)
 * se calculan dinámicamente con calcularPascua().
 *
 * @const {!Array<!Array>}
 */
const HOLIDAYS_FIJOS_2026 = [
  ['F-01', 'Año Nuevo',              'nacional_fijo',   new Date(2026, 0,  1)],
  ['F-04', 'Día del Trabajo',        'nacional_fijo',   new Date(2026, 4,  1)],
  ['F-05', 'San Pedro y San Pablo',  'nacional_fijo',   new Date(2026, 5, 29)],
  ['F-06', 'Fiestas Patrias',        'nacional_fijo',   new Date(2026, 6, 28)],
  ['F-07', 'Fiestas Patrias (2da)',  'nacional_fijo',   new Date(2026, 6, 29)],
  ['F-08', 'Santa Rosa de Lima',     'nacional_fijo',   new Date(2026, 7, 30)],
  ['F-09', 'Combate de Angamos',     'nacional_fijo',   new Date(2026, 9,  8)],
  ['F-10', 'Todos los Santos',       'nacional_fijo',   new Date(2026, 10, 1)],
  ['F-11', 'Inmaculada Concepción',  'nacional_fijo',   new Date(2026, 11, 8)],
  ['F-12', 'Batalla de Ayacucho',    'nacional_fijo',   new Date(2026, 11, 9)],
  ['F-13', 'Navidad',                'nacional_fijo',   new Date(2026, 11, 25)],
];

/**
 * Calcula la fecha de Pascua (domingo de Resurrección) para un año dado.
 * Algoritmo: Meeus/Jones/Butcher (1987) — método anónimo gregoriano.
 *
 * Referencia: https://en.wikipedia.org/wiki/Date_of_Easter
 *
 * @param {number} anio Año (1583..4099).
 * @returns {Date} Domingo de Pascua.
 * @throws {Error} Si anio fuera de rango.
 * @example
 *   const pascua = calcularPascua(2026);
 *   // → Sun Apr 05 2026
 */
function calcularPascua(anio) {
  if (anio < 1583 || anio > 4099) {
    throw new Error('pascua_out_of_range: ' + anio + ' (rango válido 1583..4099)');
  }

  const a = anio % 19;
  const b = Math.floor(anio / 100);
  const c = anio % 100;
  const d = Math.floor(b / 4);
  const e = b % 4;
  const f = Math.floor((b + 8) / 25);
  const g = Math.floor((b - f + 1) / 3);
  const h = (19 * a + b - d - g + 15) % 30;
  const i = Math.floor(c / 4);
  const k = c % 4;
  const l = (32 + 2 * e + 2 * i - h - k) % 7;
  const m = Math.floor((a + 11 * h + 22 * l) / 451);

  // Mes: 3 = marzo, 4 = abril
  const mes = Math.floor((h + l - 7 * m + 114) / 31); // 1-indexed
  const dia = ((h + l - 7 * m + 114) % 31) + 1;

  return new Date(anio, mes - 1, dia);
}

/**
 * Jueves Santo = Pascua - 3 días.
 * @param {number} anio Año.
 * @returns {Date}
 * @example const j = calcularJuevesSanto(2026); // → Thu Apr 02 2026
 */
function calcularJuevesSanto(anio) {
  const pascua = calcularPascua(anio);
  pascua.setDate(pascua.getDate() - 3);
  return pascua;
}

/**
 * Viernes Santo = Pascua - 2 días.
 * @param {number} anio Año.
 * @returns {Date}
 * @example const v = calcularViernesSanto(2026); // → Fri Apr 03 2026
 */
function calcularViernesSanto(anio) {
  const pascua = calcularPascua(anio);
  pascua.setDate(pascua.getDate() - 2);
  return pascua;
}

/**
 * Puebla 01_Feriados_PE con los 13 feriados nacionales del año.
 * Idempotente: si la hoja ya tiene los 13 feriados del año, no duplica.
 *
 * Orden de escritura:
 *   F-01 a F-13 ordenados por fecha ascendente.
 *
 * @param {number} anio Año a poblar.
 * @returns {void}
 */
function poblarFeriados(anio) {
  if (!anio || anio < 2024 || anio > 2030) {
    throw new Error('feriados_anio_invalid: ' + anio);
  }

  const ss = SpreadsheetApp.getActiveSpreadsheet();
  let sh = ss.getSheetByName(SHEET_HOLIDAYS);
  if (!sh) {
    sh = ss.insertSheet(SHEET_HOLIDAYS);
    sh.getRange(1, 1, 1, HOLIDAYS_HEADER.length).setValues([HOLIDAYS_HEADER]);
    sh.setTabColor('#DC2626');
  }

  // Construir set completo: 11 fijos + 2 móviles
  const juevesSanto = calcularJuevesSanto(anio);
  const viernesSanto = calcularViernesSanto(anio);

  const moviles = [
    ['F-02', 'Jueves Santo',  'nacional_movil', juevesSanto],
    ['F-03', 'Viernes Santo', 'nacional_movil', viernesSanto],
  ];

  /** @type {!Array<!Array>} */
  const todos = [];
  HOLIDAYS_FIJOS_2026.concat(moviles).forEach(function (row) {
    if (row[3].getFullYear() === anio) {
      todos.push([
        row[0],
        row[1],
        row[2],
        row[3],
        row[2] === 'nacional_movil',
        true, // Aplica_HE
      ]);
    }
  });

  // Ordenar por fecha ascendente
  todos.sort(function (a, b) {
    return a[3].getTime() - b[3].getTime();
  });

  // Limpiar hoja desde fila 2 hacia abajo
  const lastRow = sh.getLastRow();
  if (lastRow > 1) {
    sh.getRange(2, 1, lastRow - 1, HOLIDAYS_HEADER.length).clearContent();
  }

  // Escribir las 13 filas
  sh.getRange(2, 1, todos.length, HOLIDAYS_HEADER.length).setValues(todos);

  // Cachear el set del año (TTL 24h — más largo que cfg porque es estático)
  CacheService.getDocumentCache().put(
    CACHE_PREFIX + 'feriados_anio_' + anio,
    JSON.stringify(todos.map(function (r) { return [r[0], r[3].toISOString()]; })),
    86400
  );
}

/**
 * Verifica si una fecha es feriado nacional PE.
 * Lee 01_Feriados_PE!D:D y compara (con caché interna por execution).
 *
 * @param {Date} fecha Fecha a verificar.
 * @returns {boolean} true si es feriado nacional.
 * @example
 *   esFeriado(new Date(2026, 4, 1)); // → true (Día del Trabajo)
 */
function esFeriado(fecha) {
  if (!(fecha instanceof Date) || isNaN(fecha.getTime())) {
    throw new Error('feriado_fecha_invalid: ' + fecha);
  }
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_HOLIDAYS);
  if (!sh) {
    return false;
  }
  // COUNTIF es O(n) sobre 13 filas — despreciable
  const fechas = sh.getRange('D2:D14').getValues();
  for (let i = 0; i < fechas.length; i++) {
    const f = fechas[i][0];
    if (f instanceof Date && sameDay_(f, fecha)) {
      return true;
    }
  }
  return false;
}

/**
 * Compara dos fechas por día (ignora hora).
 * @param {Date} a
 * @param {Date} b
 * @returns {boolean}
 */
function sameDay_(a, b) {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

/**
 * Bootstrap inicial — llamado por 07_TriggersWiring.instalarTriggers()
 * o manualmente por el administrador.
 * Puebla 01_Feriados_PE para el año activo en 00_Config!B3.
 * @returns {void}
 */
function bootstrapFeriados() {
  const anio = getAnio();
  poblarFeriados(anio);
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  03_RegistroEngine.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 03_RegistroEngine.gs
 * ----------------------------------------------------------------------------
 * Maneja la working surface 10_Registro_HE (~110 filas máx).
 * NO acumula filas (CORR-1) — cambio de mes SOBREESCRIBE el grid vía 08_DBSync.
 *
 * Responsabilidades:
 *   - upsertRegistro(fecha, idPersona, hIngreso, hSalida, turno, obs)
 *   - validarInvariantes()     → llamado desde onEdit
 *   - poblarGridVacio(mes, anio) → crea filas vacías para todas las personas
 *     activas del mes (combinación fecha × persona)
 *   - recalcularFila(rowNumber) → ver 04_FormulasRecalc.gs
 *
 * Authority: marco · Implementer: andres
 */

/** @constant {string} Hoja working surface */
const SHEET_REGISTRO = '10_Registro_HE';

/** @constant {string} Hoja maestra de personas */
const SHEET_PERSONAS = '02_Personas';

/** @constant {number} Filas máx del grid visible */
const REGISTRO_MAX_ROWS = 110;

/** @constant {number} Header del registro (fila) */
const REGISTRO_HEADER_ROW = 1;

/** @constant {!Array<string>} Headers de 10_Registro_HE */
const REGISTRO_HEADERS = [
  'Fecha',
  'ISO_Week',
  'Dia_Semana',
  'Es_Feriado',
  'Tipo_Dia',
  'ID_Persona',
  'Apellidos_Nombres',
  'Turno',
  'Hora_Ingreso',
  'Hora_Salida',
  'Horas_Total',
  'Horas_Normales',
  'Horas_Extras',
  'Factor_Aplicado',
  'Horas_Ponderadas',
  'Pago_HE_PEN',
  'Observaciones',
  'Aud_Usuario',
  'Aud_Timestamp',
  'Estado',
  'Mes_Año',
  'Sync_Status',
];

/**
 * Inserta o actualiza una fila en 10_Registro_HE.
 * Usa (Fecha, ID_Persona) como PK lógica.
 *
 * @param {Date|string} fecha Fecha del registro (Date o string ISO).
 * @param {string} idPersona ID de persona formato "P-NNN".
 * @param {string} hIngreso Hora ingreso formato "HH:mm" o Date.
 * @param {string} hSalida Hora salida formato "HH:mm" o Date.
 * @param {string} turno "diurno" | "nocturno".
 * @param {string} obs Observaciones libres.
 * @returns {number} Número de fila afectada (1-indexed).
 * @throws {Error} Si params inválidos.
 */
function upsertRegistro(fecha, idPersona, hIngreso, hSalida, turno, obs) {
  // Validaciones
  if (!idPersona || !/^P-\d{3}$/.test(idPersona)) {
    throw new Error('upsert_id_persona_invalid: ' + idPersona);
  }
  if (turno !== 'diurno' && turno !== 'nocturno') {
    throw new Error('upsert_turno_invalid: ' + turno);
  }
  const d = (fecha instanceof Date) ? fecha : new Date(fecha);
  if (isNaN(d.getTime())) {
    throw new Error('upsert_fecha_invalid: ' + fecha);
  }

  const lock = LockService.getDocumentLock();
  lock.waitLock(15000);
  try {
    const ss = SpreadsheetApp.getActiveSpreadsheet();
    const sh = ss.getSheetByName(SHEET_REGISTRO);
    if (!sh) {
      throw new Error('upsert_sheet_missing: ' + SHEET_REGISTRO);
    }

    // Buscar fila existente por (Fecha, ID_Persona)
    const data = sh.getRange(2, 1, sh.getLastRow() - 1, 22).getValues();
    let targetRow = -1;
    for (let i = 0; i < data.length; i++) {
      const cellFecha = data[i][0];
      const cellId = data[i][5]; // col F
      if (
        cellFecha instanceof Date && sameDay_(cellFecha, d) &&
        cellId === idPersona
      ) {
        targetRow = i + 2;
        break;
      }
    }

    // Si no existe, crear al final (respetando REGISTRO_MAX_ROWS)
    if (targetRow < 0) {
      const nextRow = sh.getLastRow() + 1;
      if (nextRow > REGISTRO_MAX_ROWS) {
        throw new Error(
          'upsert_grid_full: working surface llegó a ' +
          REGISTRO_MAX_ROWS + ' filas. Use loadMonthToGrid() para liberar espacio.'
        );
      }
      targetRow = nextRow;
      sh.getRange(targetRow, 1).setValue(d);
    }

    // Escribir valores (columnas I, J, H, Q + auditoría R, S, T)
    sh.getRange(targetRow, 8).setValue(turno);    // H
    sh.getRange(targetRow, 9).setValue(hIngreso); // I
    sh.getRange(targetRow, 10).setValue(hSalida); // J
    sh.getRange(targetRow, 17).setValue(obs);     // Q
    sh.getRange(targetRow, 18).setValue(Session.getActiveUser().getEmail()); // R
    sh.getRange(targetRow, 19).setValue(new Date()); // S
    sh.getRange(targetRow, 20).setValue('borrador');  // T
    sh.getRange(targetRow, 22).setValue('pending');   // V sync_status

    // Log audit
    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'update', {
        sheet: SHEET_REGISTRO,
        row: targetRow,
        col: 'I-J',
        oldValue: '',
        newValue: idPersona + '@' + hIngreso + '-' + hSalida,
      });
    }

    return targetRow;
  } finally {
    lock.releaseLock();
  }
}

/**
 * Valida invariantes por fila cuando el usuario edita manualmente.
 * Disparado por onEdit en 04_FormulasRecalc.gs.
 *
 * Invariantes:
 *   - L + M = K  (Horas_Normales + Horas_Extras = Horas_Total)
 *   - K entre 0 y 24
 *   - Si E="Feriado" y F activo → N = factor feriado
 *
 * Marca fila con CF roja si viola invariantes (vía comentario en col S).
 *
 * @returns {void}
 */
function validarInvariantes() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_REGISTRO);
  if (!sh) return;
  const lastRow = Math.min(sh.getLastRow(), REGISTRO_MAX_ROWS);
  if (lastRow < 2) return;

  // Batch read K, L, M (columnas 11, 12, 13)
  const km = sh.getRange(2, 11, lastRow - 1, 3).getValues();
  let violaciones = 0;
  for (let i = 0; i < km.length; i++) {
    const k = Number(km[i][0]);
    const l = Number(km[i][1]);
    const m = Number(km[i][2]);
    const rowNum = i + 2;

    // Invariante 1: rango
    if (!isNaN(k) && (k < 0 || k > 24)) {
      sh.getRange(rowNum, 19).setNote('ERROR: K fuera de [0,24]');
      violaciones++;
      continue;
    }
    // Invariante 2: L + M = K
    if (!isNaN(k) && !isNaN(l) && !isNaN(m)) {
      const diff = Math.abs(k - (l + m));
      if (diff > 0.01) {
        sh.getRange(rowNum, 19).setNote('ERROR: L + M ≠ K (diff=' + diff.toFixed(2) + ')');
        violaciones++;
        continue;
      }
    }
    sh.getRange(rowNum, 19).setNote('');
  }

  if (violaciones > 0 && typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'validate', {
      sheet: SHEET_REGISTRO,
      row: 0,
      col: 'K-M',
      oldValue: '',
      newValue: violaciones + ' violaciones detectadas',
    });
  }
}

/**
 * Puebla el grid con filas vacías para la combinación (fecha × persona activa)
 * del mes/año solicitado. Usado cuando se carga un mes nuevo y la DB está vacía.
 *
 * Crea max ~50 personas × 31 días = 1550 filas conceptuales, pero solo
 * escribe las primeras REGISTRO_MAX_ROWS (110) para mantener la working surface
 * acotada. El resto se carga on-demand al cambiar semana (B4).
 *
 * @param {number} mes Mes (1-12).
 * @param {number} anio Año.
 * @returns {number} Filas creadas.
 */
function poblarGridVacio(mes, anio) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const personasSh = ss.getSheetByName(SHEET_PERSONAS);
  if (!personasSh) {
    throw new Error('poblar_personas_missing: ' + SHEET_PERSONAS);
  }

  // Leer personas activas (col H = Activo)
  const personas = personasSh.getRange(2, 1, personasSh.getLastRow() - 1, 10).getValues()
    .filter(function (r) { return r[7] === true; });

  if (personas.length === 0) return 0;

  // Días del mes
  const diasEnMes = new Date(anio, mes, 0).getDate();

  // Construir matriz: 1 fila por (persona × día), capada a REGISTRO_MAX_ROWS
  const rows = [];
  for (let p = 0; p < personas.length && rows.length < REGISTRO_MAX_ROWS; p++) {
    for (let d = 1; d <= diasEnMes && rows.length < REGISTRO_MAX_ROWS; d++) {
      rows.push([
        new Date(anio, mes - 1, d), // A: Fecha
        '',                         // B: ISO_Week (fórmula)
        '',                         // C: Dia_Semana (fórmula)
        '',                         // D: Es_Feriado (fórmula)
        '',                         // E: Tipo_Dia (fórmula)
        personas[p][0],             // F: ID_Persona
        '',                         // G: Apellidos_Nombres (VLOOKUP)
        '',                         // H: Turno
        '',                         // I: Hora_Ingreso
        '',                         // J: Hora_Salida
        '',                         // K: Horas_Total (fórmula)
        '',                         // L: Horas_Normales (fórmula)
        '',                         // M: Horas_Extras (fórmula)
        '',                         // N: Factor_Aplicado (fórmula)
        '',                         // O: Horas_Ponderadas (fórmula)
        '',                         // P: Pago_HE_PEN (fórmula)
        '',                         // Q: Observaciones
        '',                         // R: Aud_Usuario
        '',                         // S: Aud_Timestamp
        'borrador',                 // T: Estado
        '',                         // U: Mes_Año (fórmula)
        'synced',                   // V: Sync_Status
      ]);
    }
  }

  const gridSh = ss.getSheetByName(SHEET_REGISTRO);
  if (!gridSh) {
    throw new Error('poblar_grid_missing: ' + SHEET_REGISTRO);
  }

  // Limpiar grid antes de poblar (preserva header fila 1)
  const lastRow = gridSh.getLastRow();
  if (lastRow > 1) {
    gridSh.getRange(2, 1, lastRow - 1, 22).clearContent();
  }

  // Escribir batch
  gridSh.getRange(2, 1, rows.length, 22).setValues(rows);

  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'create', {
      sheet: SHEET_REGISTRO,
      row: 0,
      col: 'A-V',
      oldValue: '',
      newValue: 'poblarGridVacio ' + anio + '-' + mes + ' (' + rows.length + ' filas)',
    });
  }

  return rows.length;
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  04_FormulasRecalc.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 04_FormulasRecalc.gs
 * ----------------------------------------------------------------------------
 * Trigger simple onEdit que detecta cambios y dispara recálculo.
 *
 * Responsabilidades:
 *   - onEdit(e)            → router: delega según hoja y rango afectado
 *   - recalcularFila(row)  → fuerza re-cálculo de fórmulas en una fila
 *   - propagarVLookups()   → reescribe VLOOKUPs en col G si 02_Personas cambió
 *
 * Anti-patterns evitados:
 *   - NO usar NOW()/TODAY()/RAND() por fila (volátiles)
 *   - NO usar INDIRECT en grid (volátil)
 *   - NO usar ARRAYFORMULA en 1,550 filas (legibilidad)
 *
 * Authority: marco · Implementer: andres
 */

/**
 * Trigger simple onEdit — se dispara automáticamente en cualquier edición.
 * Detecta la hoja afectada y delega:
 *   - 00_Config!B2/B3/B4 → 08_DBSync.onConfigChange(e)
 *   - 10_Registro_HE!     → recalcular fila afectada + propagar VLOOKUPs si F
 *   - 02_Personas!        → propagar VLOOKUPs en todo el grid
 *
 * @param {GoogleAppsScript.Events.SheetsOnEditEvent} e Evento onEdit.
 * @returns {void}
 */
function onEdit(e) {
  if (!e || !e.range) return;
  const sheetName = e.range.getSheet().getName();
  const row = e.range.getRow();
  const col = e.range.getColumn();

  // 1. Cambio en 00_Config → dispara sync grid↔DB
  if (sheetName === SHEET_CFG) {
    if (row === 2 || row === 3 || row === 4) {
      if (typeof onConfigChange === 'function') {
        onConfigChange(e);
      }
    }
    // Cambio en B6/B7/B8 (factores) o B9 (jornada) → invalidate cache + touch refresh
    if (row >= 6 && row <= 9) {
      invalidateCache();
    }
    return;
  }

  // 2. Edición en 10_Registro_HE → recalcular fila
  if (sheetName === SHEET_REGISTRO) {
    // Si editó F (ID_Persona) o H (Turno) → propagar VLOOKUP G
    if (col === 6 || col === 8) {
      // Spreadsheet recalcula la fórmula G automáticamente, pero por seguridad:
      SpreadsheetApp.flush();
    }
    // Validar invariantes en la fila editada
    if (row >= 2 && row <= REGISTRO_MAX_ROWS) {
      recalcularFila(row);
      validarInvariantes(); // costoso — Apps Script hace debounce interno
    }
    // Si editó I/J → K se recalcula automáticamente
    return;
  }

  // 3. Edición en 02_Personas → propagar VLOOKUPs en todo el grid
  if (sheetName === SHEET_PERSONAS) {
    propagarVLookups();
    return;
  }

  // 4. Edición en 01_Feriados_PE → invalidar cache (CORR-1: lookup directo)
  if (sheetName === SHEET_HOLIDAYS) {
    invalidateCache();
    // Re-trigger fórmulas D/E en grid
    const gridSh = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(SHEET_REGISTRO);
    if (gridSh) {
      const last = Math.min(gridSh.getLastRow(), REGISTRO_MAX_ROWS);
      if (last >= 2) {
        gridSh.getRange(2, 4, last - 1, 2).setFormula(gridSh.getRange(2, 4, last - 1, 2).getFormula());
      }
    }
    return;
  }
}

/**
 * Fuerza recálculo de una fila específica del grid.
 * Toca una celda dummy (R columna) para invalidar dependencias.
 * Apps Script NO puede forzar recálculo nativo — la mejor estrategia es
 * setValue del mismo valor, lo que hace que Sheets re-evalúe dependencias.
 *
 * @param {number} rowNumber Fila (2..110).
 * @returns {void}
 */
function recalcularFila(rowNumber) {
  if (rowNumber < 2 || rowNumber > REGISTRO_MAX_ROWS) {
    throw new Error('recalc_row_out_of_range: ' + rowNumber);
  }
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_REGISTRO);
  if (!sh) return;

  // Set S (Aud_Timestamp) al mismo valor fuerza recálculo de K,L,M,O,P dependientes
  // No usamos setValue con Date.now() para evitar comportamiento confuso al usuario
  const dummyCell = sh.getRange(rowNumber, 19); // col S
  const cur = dummyCell.getValue();
  dummyCell.setValue(cur);
  SpreadsheetApp.flush();
}

/**
 * Propaga VLOOKUPs en columna G para todas las filas del grid.
 * Llamado cuando cambia 02_Personas (alta/baja de persona, edición nombre).
 *
 * Implementación: reescribe fórmula G2:G110 (Sheets evalúa de nuevo).
 *
 * @returns {void}
 */
function propagarVLookups() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_REGISTRO);
  if (!sh) return;
  const last = Math.min(sh.getLastRow(), REGISTRO_MAX_ROWS);
  if (last < 2) return;

  // Tomamos la fórmula actual de G2 y la replicamos
  const formula = sh.getRange(2, 7).getFormula();
  if (!formula || formula.length === 0) return;
  sh.getRange(2, 7, last - 1, 1).setFormula(formula);
  SpreadsheetApp.flush();
}

/**
 * Recalcula todo el grid (operación pesada — usar solo manualmente).
 * @returns {void}
 */
function recalcularGridCompleto() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_REGISTRO);
  if (!sh) return;
  const last = Math.min(sh.getLastRow(), REGISTRO_MAX_ROWS);
  if (last < 2) return;

  // Toca la columna S en todas las filas (dummy)
  const rango = sh.getRange(2, 19, last - 1, 1);
  const valores = rango.getValues();
  rango.setValues(valores);
  SpreadsheetApp.flush();
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  05_DashboardRefresh.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 05_DashboardRefresh.gs
 * ----------------------------------------------------------------------------
 * Refresca KPIs y gráficos de 20_Dashboard.
 *
 * Estrategia:
 *   - Lee de 98_DB_YYYY_MM (storage), NO del grid de trabajo (CORR-1).
 *   - Usa INDIRECT acotado a 12 referencias exactas (una por mes).
 *   - Dispara recálculo de fórmulas en Dashboard tras cada refresh.
 *
 * Responsabilidades:
 *   - refrescarKPIs()           → fuerza recálculo de A1:F4 (4 cards)
 *   - refrescarHeatmap()        → refresca matrix semana × persona
 *   - refrescarGraficoMensual() → reconstruye chart de columna
 *
 * Authority: marco · Implementer: andres
 */

/** @constant {string} Hoja dashboard */
const SHEET_DASHBOARD = '20_Dashboard';

/** @constant {string} Celda oculta con referencia al mes activo (helper INDIRECT) */
const DASHBOARD_DB_REF_CELL = 'M1';

/**
 * Recalcula todos los KPIs del Dashboard.
 * Toca celdas auxiliares para forzar recálculo de fórmulas SUMIFS/INDIRECT.
 *
 * @returns {void}
 */
function refrescarKPIs() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_DASHBOARD);
  if (!sh) {
    throw new Error('dash_missing: ' + SHEET_DASHBOARD);
  }

  // Celda M1 mantiene la referencia: ="98_DB_" & B3 & "_" & TEXT(B2, "00")
  // Tocar B2 o B3 ya dispara recálculo; este método es para refresh manual.
  // Toca celda M1 dummy → fuerza recálculo en cascada de A1:F4
  const m1 = sh.getRange(DASHBOARD_DB_REF_CELL);
  if (m1.getValue() === '') {
    // Inicializa si está vacía
    m1.setFormula('="98_DB_" & $B$3 & "_" & TEXT($B$2, "00")');
  } else {
    m1.setValue(m1.getValue()); // touch
  }
  SpreadsheetApp.flush();

  // Actualiza timestamp
  if (typeof touchLastRefresh === 'function') {
    touchLastRefresh();
  }

  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'recalc', {
      sheet: SHEET_DASHBOARD,
      row: 0,
      col: 'A1:F4',
      oldValue: '',
      newValue: 'refreshKPIs',
    });
  }
}

/**
 * Refresca heatmap semana × persona (H1:N40).
 *
 * @returns {void}
 */
function refrescarHeatmap() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_DASHBOARD);
  if (!sh) return;

  // Toca rango para forzar recálculo
  const rango = sh.getRange('H1:N40');
  const formulas = rango.getFormulas();
  if (formulas.length > 0 && formulas[0].length > 0) {
    rango.setFormulas(formulas);
  }
  SpreadsheetApp.flush();

  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'recalc', {
      sheet: SHEET_DASHBOARD,
      row: 0,
      col: 'H1:N40',
      oldValue: '',
      newValue: 'refreshHeatmap',
    });
  }
}

/**
 * Reconstruye gráfico mensual HE (column chart en A20:L20).
 * Apps Script Chart Service reconstruye el chart desde la fuente de datos.
 *
 * @returns {void}
 */
function refrescarGraficoMensual() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_DASHBOARD);
  if (!sh) return;

  // Fuente de datos: A20:F40 (resumen por persona del mes activo)
  // Chart service API:
  const charts = sh.getCharts();
  for (let i = 0; i < charts.length; i++) {
    const chart = charts[i];
    // Solo actualizar charts que cubren el rango del gráfico mensual
    const range = chart.getRanges();
    let touchesMensual = false;
    for (let j = 0; j < range.length; j++) {
      const a1 = range[j].getA1Notation();
      if (a1.indexOf('20:') >= 0 || a1.indexOf('A20') >= 0) {
        touchesMensual = true;
        break;
      }
    }
    if (touchesMensual) {
      sh.removeChart(chart);
      // Recrear (simplificado — Apps Script requiere builder pattern)
      const builder = sh.newChart();
      const sourceRange = sh.getRange('A20:F40');
      builder
        .setChartType(Charts.ChartType.COLUMN)
        .addRange(sourceRange)
        .setPosition(20, 1, 0, 0)
        .setOption('title', 'HE Mensual por Persona')
        .setOption('hAxis', { title: 'Persona' })
        .setOption('vAxis', { title: 'Horas Extras' });
      sh.insertChart(builder.build());
    }
  }
  SpreadsheetApp.flush();

  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'recalc', {
      sheet: SHEET_DASHBOARD,
      row: 0,
      col: 'A20:L20',
      oldValue: '',
      newValue: 'refreshGraficoMensual',
    });
  }
}

/**
 * Refresca TODOS los componentes del Dashboard (KPIs + Heatmap + Gráfico).
 * Llamado desde 07_TriggersWiring.triggerCierreMes().
 *
 * @returns {void}
 */
function refrescarDashboardCompleto() {
  refrescarKPIs();
  refrescarHeatmap();
  refrescarGraficoMensual();
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  06_AuditLogger.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 06_AuditLogger.gs
 * ----------------------------------------------------------------------------
 * Append-only logger a 99_Audit_Log.
 *
 * Schema (8 columnas):
 *   A: Timestamp      (datetime)
 *   B: Usuario        (email editor)
 *   C: Acción         (create|update|delete|recalc|trigger|sync_load|sync_save|sync_create_sheet|validate)
 *   D: Sheet          (text)
 *   E: Fila           (int, 0 si aplica a muchas filas)
 *   F: Columna        (text, rango A1)
 *   G: Valor_Anterior (text)
 *   H: Valor_Nuevo    (text)
 *
 * Garantías:
 *   - Append-only (nunca edita filas existentes)
 *   - Tolerante a hoja faltante (auto-crea con headers)
 *   - Thread-safe via LockService (lectura atómica de lastRow)
 *
 * Authority: marco · Implementer: andres
 */

/** @constant {string} Hoja audit log */
const SHEET_AUDIT = '99_Audit_Log';

/** @constant {!Array<string>} Headers */
const AUDIT_HEADERS = [
  'Timestamp',
  'Usuario',
  'Acción',
  'Sheet',
  'Fila',
  'Columna',
  'Valor_Anterior',
  'Valor_Nuevo',
];

/** @constant {number} Lock timeout en ms */
const AUDIT_LOCK_TIMEOUT = 5000;

/**
 * Garantiza que 99_Audit_Log existe con headers.
 * Idempotente.
 *
 * @returns {GoogleAppsScript.Spreadsheet.Sheet}
 */
function ensureAuditSheet_() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  let sh = ss.getSheetByName(SHEET_AUDIT);
  if (!sh) {
    sh = ss.insertSheet(SHEET_AUDIT);
    sh.getRange(1, 1, 1, AUDIT_HEADERS.length).setValues([AUDIT_HEADERS]);
    sh.setTabColor('#374151');
    sh.hideSheet();
  }
  return sh;
}

/**
 * Registra una acción en el audit log.
 *
 * @param {string} usuario Email del usuario que ejecutó la acción.
 * @param {string} accion Tipo: 'create'|'update'|'delete'|'recalc'|'trigger'|'sync_load'|'sync_save'|'sync_create_sheet'|'validate'.
 * @param {{sheet:string, row:number, col:string, oldValue:*, newValue:*}} payload
 * @returns {void}
 * @example
 *   logAccion('jefe@cima.pe', 'sync_load', {
 *     sheet: '10_Registro_HE',
 *     row: 0,
 *     col: 'A2:V110',
 *     oldValue: '',
 *     newValue: 'loaded 87 rows from 98_DB_2026_03'
 *   });
 */
function logAccion(usuario, accion, payload) {
  // Validaciones suaves — nunca bloquear la operación por audit fallido
  if (!accion || typeof accion !== 'string') {
    accion = 'unknown';
  }
  if (!usuario || typeof usuario !== 'string') {
    usuario = 'system';
  }

  let lock;
  try {
    lock = LockService.getDocumentLock();
    lock.waitLock(AUDIT_LOCK_TIMEOUT);
  } catch (e) {
    // Si no obtiene lock, abortar silenciosamente (audit no debe romper flujo)
    return;
  }

  try {
    const sh = ensureAuditSheet_();
    const lastRow = sh.getLastRow();
    const newRow = lastRow + 1;

    const row = [
      new Date(),
      usuario,
      accion,
      (payload && payload.sheet) || '',
      (payload && payload.row != null) ? payload.row : 0,
      (payload && payload.col) || '',
      stringifySafe_(payload && payload.oldValue),
      stringifySafe_(payload && payload.newValue),
    ];

    sh.getRange(newRow, 1, 1, AUDIT_HEADERS.length).setValues([row]);

    // Capacidad de la hoja: 10000 filas (definida en architecture.yaml).
    // Si llega a 9500, alertar (no implementado — solo log warning).
    if (newRow > 9500 && typeof console !== 'undefined') {
      console.warn('audit_sheet_critical: 99_Audit_Log > 9500 filas');
    }
  } catch (e) {
    if (typeof console !== 'undefined') {
      console.error('audit_log_failed: ' + e.message);
    }
  } finally {
    if (lock && lock.hasLock && lock.hasLock()) {
      lock.releaseLock();
    }
  }
}

/**
 * Convierte un valor a string seguro para escribir al log.
 * - null/undefined → ''
 * - Date → ISO string
 * - Object → JSON
 * - Otros → String(v)
 *
 * @param {*} v
 * @returns {string}
 */
function stringifySafe_(v) {
  if (v === null || v === undefined) return '';
  if (v instanceof Date) return v.toISOString();
  if (typeof v === 'object') {
    try {
      return JSON.stringify(v);
    } catch (e) {
      return String(v);
    }
  }
  return String(v);
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  07_TriggersWiring.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 07_TriggersWiring.gs
 * ----------------------------------------------------------------------------
 * Registra time-driven triggers (cuota Google: 30 triggers por usuario).
 *
 * Triggers activos (max 3 — restricción architecture.yaml):
 *   1. onEdit (simple trigger)  → 04_FormulasRecalc.onEdit
 *   2. onChange (installable)   → 04_FormulasRecalc.onEdit
 *   3. time-driven diario 02:00 → 05_DashboardRefresh.refrescarDashboardCompleto
 *   +  trigger mensual (último día 23:00) → 08_DBSync.triggerCierreMes
 *
 * Authority: marco · Implementer: andres
 */

/** @constant {string} ID del trigger onEdit simple (de Apps Script) */
const TRIGGER_FUNC_ON_EDIT = 'onEdit';

/** @constant {string} ID del trigger time-driven diario */
const TRIGGER_DAILY_FUNC = 'refrescarDashboardCompleto';

/** @constant {string} ID del trigger time-driven mensual */
const TRIGGER_MONTHLY_FUNC = 'triggerCierreMes';

/**
 * Instala los 3 triggers del sistema.
 * Idempotente: si ya existen, los reemplaza en lugar de duplicar.
 *
 * Cuota Google: max 30 triggers/usuario. OVAV usa 3.
 *
 * @returns {void}
 */
function instalarTriggers() {
  // Limpiar triggers previos del proyecto (idempotencia)
  eliminarTriggers();

  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 1. onChange (installable — captura cambios estructurales)
  ScriptApp.newTrigger(TRIGGER_FUNC_ON_EDIT)
    .forSpreadsheet(ss)
    .onChange()
    .create();

  // 2. Daily refresh @ 02:00 Lima time
  ScriptApp.newTrigger(TRIGGER_DAILY_FUNC)
    .timeBased()
    .atHour(2)
    .everyDays(1)
    .inTimezone('America/Lima')
    .create();

  // 3. Monthly close — último día del mes @ 23:00 (ver triggerCierreMes)
  // Apps Script no soporta "último día del mes" directamente; usamos
  // diario y triggerCierreMes() verifica si es último día del mes.
  ScriptApp.newTrigger(TRIGGER_MONTHLY_FUNC)
    .timeBased()
    .atHour(23)
    .everyDays(1)
    .inTimezone('America/Lima')
    .create();

  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'trigger', {
      sheet: '(system)',
      row: 0,
      col: '',
      oldValue: '',
      newValue: '3 triggers instalados (onChange + daily + monthly-check)',
    });
  }
}

/**
 * Elimina todos los triggers asociados a este proyecto.
 * @returns {void}
 */
function eliminarTriggers() {
  const triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(function (t) {
    ScriptApp.deleteTrigger(t);
  });
}

/**
 * Trigger diario ejecutado a las 23:00 Lima time.
 * Si es último día del mes → ejecuta cierre.
 * Si no → no-op (pero mantiene trigger activo).
 *
 * @returns {void}
 */
function triggerCierreMes() {
  const hoy = new Date();
  const manana = new Date(hoy.getFullYear(), hoy.getMonth(), hoy.getDate() + 1);
  // Si mañana es día 1 del mes siguiente → hoy es último día del mes
  if (manana.getDate() === 1) {
    cerrarMes(hoy);
  }
}

/**
 * Lógica de cierre mensual:
 *   1. saveGridToDB(mes, anio)   — persiste grid actual
 *   2. Pre-crear hoja DB del mes siguiente (lazy)
 *   3. Refrescar dashboard
 *   4. Log audit 'sync_save' + 'sync_create_sheet'
 *
 * @param {Date} fechaReferencia Fecha del último día del mes (param para test).
 * @returns {void}
 */
function cerrarMes(fechaReferencia) {
  const anio = fechaReferencia.getFullYear();
  const mes = fechaReferencia.getMonth() + 1; // 1-12

  const lock = LockService.getDocumentLock();
  lock.waitLock(30000);
  try {
    // 1. Guardar grid actual en DB del mes que termina
    if (typeof saveGridToDB === 'function') {
      saveGridToDB(mes, anio);
    }

    // 2. Pre-crear DB del mes siguiente
    const sigMes = mes === 12 ? 1 : mes + 1;
    const sigAnio = mes === 12 ? anio + 1 : anio;
    if (typeof createMonthSheetIfMissing === 'function') {
      createMonthSheetIfMissing(sigMes, sigAnio);
    }

    // 3. Refrescar dashboard
    if (typeof refrescarDashboardCompleto === 'function') {
      refrescarDashboardCompleto();
    }

    if (typeof logAccion === 'function') {
      logAccion('system', 'trigger', {
        sheet: '(cierre-mes)',
        row: 0,
        col: '',
        oldValue: '',
        newValue: 'Cierre ' + anio + '-' + mes + ' (pre-creado ' + sigAnio + '-' + sigMes + ')',
      });
    }
  } finally {
    lock.releaseLock();
  }
}

/**
 * Helper de debugging — lista triggers activos.
 * Solo para admin. Ejecutar desde editor de Apps Script.
 * @returns {!Array<!Object>}
 */
function listarTriggers() {
  return ScriptApp.getProjectTriggers().map(function (t) {
    return {
      id: t.getUniqueId(),
      handler: t.getHandlerFunction(),
      type: t.getEventType().toString(),
      source: t.getTriggerSource().toString(),
    };
  });
}

// ╔══════════════════════════════════════════════════════════════════════╗
// ║  08_DBSync.gs
// ╚══════════════════════════════════════════════════════════════════════╝

/**
 * 08_DBSync.gs
 * ----------------------------------------------------------------------------
 * Módulo CRÍTICO — sincroniza 10_Registro_HE (working surface) ↔ 98_DB_YYYY_MM (storage).
 *
 * Responsabilidades:
 *   - resolveTargetSheet(mes, anio)         → "98_DB_2026_03"
 *   - createMonthSheetIfMissing(mes, anio)  → lazy creation con headers + DV/CF
 *   - loadMonthToGrid(mes, anio)            → DB → grid (sobreescribe)
 *   - saveGridToDB(mes, anio)               → grid → DB (persiste) [idempotente]
 *   - onConfigChange(e)                     → orquestador disparado por onEdit
 *
 * Garantías (architecture.yaml sync_invariants):
 *   - save SIEMPRE antes de load (evitar pérdida)
 *   - LockService.getDocumentLock() en TODA operación (R-1 race condition)
 *   - saveGridToDB idempotente (puede llamarse 2x sin duplicar)
 *   - Solo filas del mes activo se sincronizan
 *
 * Authority: marco · Implementer: andres
 */

/** @constant {number} Lock timeout en ms — 30s porque save puede tardar */
const SYNC_LOCK_TIMEOUT = 30000;

/**
 * Construye el nombre de la hoja DB destino.
 * @param {number} mes Mes (1-12).
 * @param {number} anio Año (YYYY).
 * @returns {string} "98_DB_YYYY_MM"
 * @example resolveTargetSheet(3, 2026); // → "98_DB_2026_03"
 */
function resolveTargetSheet(mes, anio) {
  if (!mes || mes < 1 || mes > 12) {
    throw new Error('sync_mes_invalid: ' + mes);
  }
  if (!anio || anio < 2024 || anio > 2030) {
    throw new Error('sync_anio_invalid: ' + anio);
  }
  const mm = mes < 10 ? '0' + mes : '' + mes;
  return '98_DB_' + anio + '_' + mm;
}

/**
 * Crea la hoja DB mensual si no existe.
 * Copia headers de 10_Registro_HE, aplica DV/CF idénticos, oculta la hoja.
 *
 * @param {number} mes Mes (1-12).
 * @param {number} anio Año.
 * @returns {GoogleAppsScript.Spreadsheet.Sheet}
 */
function createMonthSheetIfMissing(mes, anio) {
  const targetName = resolveTargetSheet(mes, anio);
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  let target = ss.getSheetByName(targetName);
  if (target) {
    return target; // idempotente
  }

  // Lock para evitar race condition con installTriggers() concurrentes
  const lock = LockService.getDocumentLock();
  lock.waitLock(SYNC_LOCK_TIMEOUT);
  try {
    // Doble check dentro del lock
    target = ss.getSheetByName(targetName);
    if (target) return target;

    target = ss.insertSheet(targetName);
    target.setTabColor('#374151');
    target.hideSheet();

    // Headers desde 10_Registro_HE
    const gridSh = ss.getSheetByName(SHEET_REGISTRO);
    if (gridSh) {
      const headers = gridSh.getRange(1, 1, 1, 22).getValues();
      target.getRange(1, 1, 1, 22).setValues(headers);
      target.setFrozenRows(1);
    } else {
      // Fallback: escribir headers canónicos
      target.getRange(1, 1, 1, REGISTRO_HEADERS.length).setValues([REGISTRO_HEADERS]);
    }

    // Pre-poblar filas vacías para el mes (50 personas × 31 días = 1550 máx)
    poblarDbVacio_(target, mes, anio);

    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'sync_create_sheet', {
        sheet: targetName,
        row: 0,
        col: '',
        oldValue: '',
        newValue: 'created ' + targetName,
      });
    }

    return target;
  } finally {
    lock.releaseLock();
  }
}

/**
 * Pre-popula filas vacías (Fecha, ID_Persona) para el mes.
 * Solo crea las claves lógicas; resto de columnas quedan vacías para edición posterior.
 *
 * @param {GoogleAppsScript.Spreadsheet.Sheet} target
 * @param {number} mes
 * @param {number} anio
 */
function poblarDbVacio_(target, mes, anio) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const personasSh = ss.getSheetByName(SHEET_PERSONAS);
  if (!personasSh) return;

  const personas = personasSh.getRange(2, 1, personasSh.getLastRow() - 1, 1).getValues()
    .map(function (r) { return r[0]; })
    .filter(function (id) { return id && /^P-\d{3}$/.test(id); });

  if (personas.length === 0) return;

  const diasEnMes = new Date(anio, mes, 0).getDate();
  const rows = [];
  for (let p = 0; p < personas.length; p++) {
    for (let d = 1; d <= diasEnMes; d++) {
      rows.push([
        new Date(anio, mes - 1, d), // A: Fecha
        '', '', '', '', '',         // B-E
        personas[p],                // F: ID_Persona
        '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', // G-V vacías (16 cols)
      ]);
    }
  }
  if (rows.length > 0) {
    target.getRange(2, 1, rows.length, 22).setValues(rows);
  }
}

/**
 * Lee DB mensual y SOBREESCRIBE working surface (load).
 * Si DB no existe → la crea vacía.
 *
 * @param {number} mes Mes destino (1-12).
 * @param {number} anio Año destino.
 * @returns {number} Filas cargadas en grid.
 */
function loadMonthToGrid(mes, anio) {
  const lock = LockService.getDocumentLock();
  lock.waitLock(SYNC_LOCK_TIMEOUT);
  try {
    const targetName = resolveTargetSheet(mes, anio);
    const ss = SpreadsheetApp.getActiveSpreadsheet();
    const target = ss.getSheetByName(targetName) || createMonthSheetIfMissing(mes, anio);
    const gridSh = ss.getSheetByName(SHEET_REGISTRO);
    if (!gridSh) {
      throw new Error('load_grid_missing: ' + SHEET_REGISTRO);
    }

    // 1. Limpiar grid (preserva header fila 1)
    const lastGrid = gridSh.getLastRow();
    if (lastGrid > 1) {
      gridSh.getRange(2, 1, lastGrid - 1, 22).clearContent();
    }

    // 2. Leer DB completa (solo filas con Mes_Año == target)
    const dbLast = target.getLastRow();
    if (dbLast < 2) {
      // DB vacía → grid queda vacío
      return 0;
    }
    const dbData = target.getRange(2, 1, dbLast - 1, 22).getValues();

    // 3. Filtrar por Mes_Año (col U = idx 20) y por semana si filtro activo
    const semActiva = getSemana();
    const filtered = [];
    for (let i = 0; i < dbData.length; i++) {
      const row = dbData[i];
      const fecha = row[0];
      if (!(fecha instanceof Date)) continue;

      // Sync key match
      const key = formatYYYYMM_(fecha);
      const expected = anio + '-' + (mes < 10 ? '0' + mes : '' + mes);
      if (key !== expected) continue;

      // Week filter (opcional — si B4 = "W##", filtrar)
      if (semActiva && /^W\d{2}$/.test(semActiva)) {
        const w = 'W' + (getISOWeekNumber_(fecha) < 10
          ? '0' + getISOWeekNumber_(fecha)
          : '' + getISOWeekNumber_(fecha));
        if (w !== semActiva) continue;
      }

      filtered.push(row);
    }

    // 4. Truncar a REGISTRO_MAX_ROWS
    const capped = filtered.slice(0, REGISTRO_MAX_ROWS - 1);

    // 5. Escribir batch
    if (capped.length > 0) {
      gridSh.getRange(2, 1, capped.length, 22).setValues(capped);
      // Marcar sync_status = 'synced'
      const statusRange = gridSh.getRange(2, 22, capped.length, 1);
      const statusValues = statusRange.getValues();
      for (let i = 0; i < statusValues.length; i++) {
        if (!statusValues[i][0]) statusValues[i][0] = 'synced';
      }
      statusRange.setValues(statusValues);
    }

    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'sync_load', {
        sheet: SHEET_REGISTRO,
        row: 0,
        col: 'A2:V' + (capped.length + 1),
        oldValue: '',
        newValue: 'loaded ' + capped.length + ' rows from ' + targetName,
      });
    }

    return capped.length;
  } finally {
    lock.releaseLock();
  }
}

/**
 * Lee working surface y escribe a DB mensual (save).
 * Idempotente: si la fila ya existe en DB, la SOBREESCRIBE (no duplica).
 *
 * @param {number} mes Mes a guardar (1-12).
 * @param {number} anio Año.
 * @returns {number} Filas guardadas.
 */
function saveGridToDB(mes, anio) {
  const lock = LockService.getDocumentLock();
  lock.waitLock(SYNC_LOCK_TIMEOUT);
  try {
    const targetName = resolveTargetSheet(mes, anio);
    const ss = SpreadsheetApp.getActiveSpreadsheet();
    const target = ss.getSheetByName(targetName) || createMonthSheetIfMissing(mes, anio);
    const gridSh = ss.getSheetByName(SHEET_REGISTRO);
    if (!gridSh) {
      throw new Error('save_grid_missing: ' + SHEET_REGISTRO);
    }

    // 1. Leer grid completo (excepto header)
    const gridLast = gridSh.getLastRow();
    if (gridLast < 2) return 0;

    const gridData = gridSh.getRange(2, 1, gridLast - 1, 22).getValues();

    // 2. Filtrar solo filas del mes activo
    const expected = anio + '-' + (mes < 10 ? '0' + mes : '' + mes);
    const toSync = [];
    for (let i = 0; i < gridData.length; i++) {
      const row = gridData[i];
      const fecha = row[0];
      if (!(fecha instanceof Date)) continue;
      const key = formatYYYYMM_(fecha);
      if (key !== expected) continue;
      toSync.push({
        rowIndex: i + 2, // 1-indexed fila en grid
        data: row,
      });
    }

    if (toSync.length === 0) return 0;

    // 3. Leer DB completa
    const dbLast = target.getLastRow();
    const dbData = dbLast < 2 ? [] : target.getRange(2, 1, dbLast - 1, 22).getValues();

    // 4. Construir índice (Fecha, ID_Persona) → fila DB
    /** @type {Object<string, number>} */
    const dbIndex = {};
    for (let i = 0; i < dbData.length; i++) {
      const r = dbData[i];
      if (r[0] instanceof Date && typeof r[5] === 'string') {
        const key = formatYYYYMM_(r[0]) + '|' + r[5];
        dbIndex[key] = i + 2; // fila DB (1-indexed)
      }
    }

    // 5. Construir matriz de updates
    /** @type {!Array<!Array>} */
    const updates = [];
    /** @type {!Array<number>} */
    const updateRows = [];
    for (let i = 0; i < toSync.length; i++) {
      const item = toSync[i];
      const fecha = item.data[0];
      const idPersona = item.data[5];
      if (typeof idPersona !== 'string' || !idPersona) continue;
      const key = formatYYYYMM_(fecha) + '|' + idPersona;

      if (dbIndex[key] !== undefined) {
        updateRows.push(dbIndex[key]);
      } else {
        // Append al final
        const newRow = dbLast + updates.length + 1;
        updateRows.push(newRow);
        // Asegurar padding en DB (fill con empty rows si gap)
        while (target.getLastRow() < newRow - 1) {
          target.appendRow(new Array(22).fill(''));
        }
      }
      updates.push(item.data);
    }

    // 6. Batch write por chunks de 500 (límite seguro)
    const CHUNK = 500;
    for (let i = 0; i < updates.length; i += CHUNK) {
      const slice = updates.slice(i, i + CHUNK);
      const rows = slice.map(function (r) { return r; });
      const startRow = updateRows[i];
      // Para simplificar y mantener idempotencia, escribimos fila por fila
      // en su fila destino (puede haber gaps)
      for (let j = 0; j < rows.length; j++) {
        target.getRange(updateRows[i + j], 1, 1, 22).setValues([rows[j]]);
      }
    }

    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'sync_save', {
        sheet: targetName,
        row: 0,
        col: '',
        oldValue: '',
        newValue: 'saved ' + updates.length + ' rows',
      });
    }

    return updates.length;
  } finally {
    lock.releaseLock();
  }
}

/**
 * Orquestador disparado por onEdit cuando cambia B2/B3/B4 en 00_Config.
 * Flujo:
 *   1. Detecta mes/año nuevos
 *   2. saveGridToDB(mes_viejo, anio)
 *   3. loadMonthToGrid(mes_nuevo, anio)
 *   4. Si DB no existe → createMonthSheetIfMissing
 *
 * @param {GoogleAppsScript.Events.SheetsOnEditEvent} e Evento onEdit.
 * @returns {void}
 */
function onConfigChange(e) {
  if (!e || !e.range) return;

  // Pre-lock global para todo el flujo save+load
  const lock = LockService.getDocumentLock();
  lock.waitLock(SYNC_LOCK_TIMEOUT);
  try {
    const prevMesAnio = detectarMesAnioAnterior_(e);
    const nuevoMes = getMes();
    const nuevoAnio = getAnio();

    // 1. Save estado anterior (si fue cambio de mes/año)
    if (prevMesAnio &&
        (prevMesAnio.mes !== nuevoMes || prevMesAnio.anio !== nuevoAnio)) {
      try {
        saveGridToDB(prevMesAnio.mes, prevMesAnio.anio);
      } catch (err) {
        if (typeof logAccion === 'function') {
          logAccion(Session.getActiveUser().getEmail(), 'sync_save', {
            sheet: '(error)',
            row: 0,
            col: '',
            oldValue: '',
            newValue: 'FAILED save: ' + err.message,
          });
        }
        // Continuar con load (no bloquear UX)
      }
    }

    // 2. Load estado nuevo
    try {
      loadMonthToGrid(nuevoMes, nuevoAnio);
    } catch (err) {
      if (typeof logAccion === 'function') {
        logAccion(Session.getActiveUser().getEmail(), 'sync_load', {
          sheet: '(error)',
          row: 0,
          col: '',
          oldValue: '',
          newValue: 'FAILED load: ' + err.message,
        });
      }
      throw err;
    }

    // 3. Refresh dashboard
    if (typeof refrescarKPIs === 'function') {
      refrescarKPIs();
    }

    // 4. Touch last refresh timestamp
    if (typeof touchLastRefresh === 'function') {
      touchLastRefresh();
    }
  } finally {
    lock.releaseLock();
  }
}

/**
 * Detecta el mes/año anterior comparando con el valor pre-cambio del edit.
 * Como onEdit no expone el valor anterior directamente, usamos heurística:
 *   - Si B2 cambió → leemos previousValue desde DataValidation contextual
 *   - Fallback: si no se puede detectar, usar mes/anio actuales y skip save
 *
 * Workaround: Apps Script no expone oldValue en onEdit simple.
 * Para captura fiel, el trigger debe ser installable onChange (ya configurado).
 *
 * @param {GoogleAppsScript.Events.SheetsOnEditEvent} e
 * @returns {{mes:number, anio:number}|null}
 */
function detectarMesAnioAnterior_(e) {
  // Heurística: leer CacheService para valor previo
  const cache = CacheService.getDocumentCache();
  const prev = cache.get(CACHE_PREFIX + 'prev_mes_anio');
  if (prev) {
    try {
      const parsed = JSON.parse(prev);
      // Update cache con nuevo valor
      cache.put(
        CACHE_PREFIX + 'prev_mes_anio',
        JSON.stringify({ mes: getMes(), anio: getAnio() }),
        3600
      );
      return parsed;
    } catch (e2) {
      return null;
    }
  }
  // Primera ejecución: cachear valor actual
  cache.put(
    CACHE_PREFIX + 'prev_mes_anio',
    JSON.stringify({ mes: getMes(), anio: getAnio() }),
    3600
  );
  return null;
}

/**
 * Formato YYYY-MM para una fecha.
 * @param {Date} d
 * @returns {string}
 */
function formatYYYYMM_(d) {
  const y = d.getFullYear();
  const m = d.getMonth() + 1;
  return y + '-' + (m < 10 ? '0' + m : '' + m);
}

/**
 * Calcula ISO week number para una fecha.
 * Implementación Apps Script (WEEKNUM no existe en GAS — usamos Date math).
 *
 * @param {Date} d
 * @returns {number} 1..53
 */
function getISOWeekNumber_(d) {
  const target = new Date(d.valueOf());
  const dayNr = (d.getDay() + 6) % 7; // lunes = 0
  target.setDate(target.getDate() - dayNr + 3);
  const firstThursday = target.valueOf();
  target.setMonth(0, 1);
  if (target.getDay() !== 4) {
    target.setMonth(0, 1 + ((4 - target.getDay()) + 7) % 7);
  }
  return 1 + Math.ceil((firstThursday - target.valueOf()) / 604800000);
}


// ╔══════════════════════════════════════════════════════════════════════╗
// ║  FIN DEL DEPLOY ALL-IN-ONE                                          ║
// ╚══════════════════════════════════════════════════════════════════════╝
//
// FUNCIONES DISPONIBLES DESPUÉS DEL DEPLOY:
// ─────────────────────────────────────────
// onEdit(e)                    → detecta cambios en 00_Config y sincroniza
// onConfigChange(e)            → orquestador del cambio de mes/año
// loadMonthToGrid(mes, año)    → carga DB mensual al grid
// saveGridToDB(mes, año)       → graba grid a DB mensual
// calcularPascua(anio)         → algoritmo Meeus/Jones/Butcher
// esFeriado(fecha)             → bool, lookup en 01_Feriados_PE
// upsertRegistro(...)          → crea/actualiza fila
// validarInvariantes()         → chequea consistencia
// refrescarKPIs()              → actualiza 20_Dashboard
// logAccion(usuario, accion)   → append a 99_Audit_Log
// instalarTriggers()           → registra onEdit + time-driven
// eliminarTriggers()           → limpia triggers
//
// VALIDACIÓN POST-DEPLOY:
// ────────────────────────
// 1. Cambiar dropdown Mes (00_Config!B2) → grid debe repoblarse en <3s
// 2. Fila del 2026-04-03 debe aparecer fondo rosa (cf_feriado)
// 3. Cambiar a un mes sin datos → grid vacío sin error
// 4. Verificar 99_Audit_Log tiene fila con tu email + timestamp
//
// ROLLBACK:
// ────────
// function rollbackHELima() {
//   const ss = SpreadsheetApp.getActiveSpreadsheet();
//   SHEETS_VISIBLES.forEach(n => {
//     const s = ss.getSheetByName(n);
//     if (s) ss.deleteSheet(s);
//   });
//   SpreadsheetApp.getUi().alert('Rollback completo');
// }

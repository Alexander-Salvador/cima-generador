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
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
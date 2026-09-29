/**
 * 04_FormulasRecalc.gs
 * ----------------------------------------------------------------------------
 * Trigger simple onEdit que detecta cambios y dispara recálculo.
 *
 * Responsabilidades:
 *   - onEdit(e)            → router: delega según hoja y rango afectado
 *   - recalcularFila(row)  → fuerza re-cálculo de fórmulas en una fila
 *   - propagarVLookups()   → reescribe VLOOKUPs en col G si 02_Personas cambió
 *   - alertasEnOnEdit()    → detecta anomalías (H>14, inactivo, conflict)
 *
 * Anti-patterns evitados:
 *   - NO usar NOW()/TODAY()/RAND() por fila (volátiles)
 *   - NO usar INDIRECT en grid (volátil)
 *   - NO usar ARRAYFORMULA en 1,550 filas (legibilidad)
 *
 * Authority: marco · Implementer: andres · alerts: thavren
 * Updated: 2026-09-29 — sistema de alertas automáticas
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
      alertasEnOnEdit_(row, col);
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

/**
 * Detecta anomalías en la fila editada y dispara alertas/al log.
 * Reglas:
 *   - H_total > 14 → sospecha fraude o error data entry
 *   - ID_Persona inactiva (02_Personas!H=FALSE) → inconsistencia
 *   - Sync_Status = 'conflict' → requiere intervención manual
 *
 * Estrategia: silencioso por defecto. Solo log a 99_Audit_Log.
 * UI alert se agrega si usuario habilita explícitamente via menú.
 *
 * @param {number} row Fila editada (2..110)
 * @param {number} col Columna editada (1..22)
 * @private
 */
function alertasEnOnEdit_(row, col) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_REGISTRO);
  if (!sh) return;

  // Solo evaluar cuando cambió K (H_total), F (ID_Persona) o V (Sync_Status)
  const colEvaluar = [6, 11, 22]; // F, K, V
  if (colEvaluar.indexOf(col) === -1) return;

  try {
    const k = sh.getRange(row, 11).getValue(); // H_total
    const f = sh.getRange(row, 6).getValue();  // ID_Persona
    const v = sh.getRange(row, 22).getValue(); // Sync_Status

    // Regla 1: H_total > 14
    if (typeof k === 'number' && k > 14) {
      logAlerta_('horas_excesivas', row, 'H_total=' + k + ' > 14h umbral', f);
    }

    // Regla 2: ID_Persona inactiva
    if (typeof f === 'string' && /^P-\d{3}$/.test(f)) {
      const personasSh = ss.getSheetByName(SHEET_PERSONAS);
      if (personasSh) {
        const filaPersona = buscarFilaPersona_(personasSh, f);
        if (filaPersona > 0) {
          const activo = personasSh.getRange(filaPersona, 8).getValue();
          if (activo === false || activo === 'FALSE' || activo === 0) {
            logAlerta_('persona_inactiva_con_horas', row, f + ' está Activo=FALSE pero recibe horas', f);
          }
        }
      }
    }

    // Regla 3: Sync_Status = conflict
    if (v === 'conflict') {
      logAlerta_('sync_conflict', row, 'Sync_Status=conflict — requiere intervención manual', f);
    }
  } catch (err) {
    console.warn('alertas_onedit_failed row=' + row + ': ' + err.message);
  }
}

/**
 * Log estructurado de alerta — usa logAccion con prefijo 'alert:'.
 * Categorías: 'horas_excesivas' | 'persona_inactiva_con_horas' | 'sync_conflict'
 *
 * @param {string} tipo
 * @param {number} row
 * @param {string} detalle
 * @param {string} idPersona
 * @private
 */
function logAlerta_(tipo, row, detalle, idPersona) {
  if (typeof logAccion !== 'function') return;
  try {
    logAccion(
      Session.getActiveUser().getEmail(),
      'alert:' + tipo,
      {
        sheet: SHEET_REGISTRO,
        row: row,
        col: 'K',
        oldValue: '',
        newValue: detalle + ' (persona=' + (idPersona || 'unknown') + ')',
      }
    );
  } catch (e) {
    console.warn('log_alerta_failed tipo=' + tipo + ': ' + e.message);
  }
}

/**
 * Busca fila de persona por ID_Persona en 02_Personas.
 * Retrecó 0 si no encontrada.
 *
 * @param {GoogleAppsScript.Spreadsheet.Sheet} personasSh
 * @param {string} idPersona ej: P-007
 * @returns {number} fila (1-based, 0 si no encontrada)
 * @private
 */
function buscarFilaPersona_(personasSh, idPersona) {
  const last = personasSh.getLastRow();
  if (last < 2) return 0;
  const ids = personasSh.getRange(2, 1, last - 1, 1).getValues();
  for (let i = 0; i < ids.length; i++) {
    if (ids[i][0] === idPersona) return i + 2;
  }
  return 0;
}
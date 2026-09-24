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
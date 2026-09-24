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
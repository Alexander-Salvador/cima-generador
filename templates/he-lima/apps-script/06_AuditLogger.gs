/**
 * 06_AuditLogger.gs
 * ----------------------------------------------------------------------------
 * Append-only logger a 99_Audit_Log.
 *
 * Schema (8 columnas, ISO 8601 estricto):
 *   A: Timestamp_ISO  (text — ISO 8601 string: 2026-09-29T15:32:00.123-05:00)
 *   B: Usuario        (email editor)
 *   C: Acción         (create|update|delete|recalc|trigger|sync_load|sync_save|sync_create_sheet|validate|menu_*)
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
 *   - Auto-archivado cuando se acerca al límite (9500 → 8500)
 *   - ISO 8601 estricto (string, no Date object — portable)
 *
 * Authority: marco · Implementer: andres · stabilize: thavren
 * Updated: 2026-09-29 — ISO 8601 + auto-archive
 */

/** @constant {string} Hoja audit log */
const SHEET_AUDIT = '99_Audit_Log';

/** @constant {!Array<string>} Headers */
const AUDIT_HEADERS = [
  'Timestamp_ISO',
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

/** @constant {number} Capacidad máxima antes de archivado */
const AUDIT_MAX_ROWS = 9500;

/** @constant {number} Umbral para disparar archivado (rotar antes del límite) */
const AUDIT_ARCHIVE_THRESHOLD = 8500;

/** @constant {number} Filas a mantener tras archivado */
const AUDIT_KEEP_ROWS = 5000;

/** @constant {string} Hoja destino del archivo (sufijo _archive_YYYY_MM) */
const AUDIT_ARCHIVE_SHEET_PREFIX = '99_Audit_Log_archive_';

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
      nowIso_(),
      usuario,
      accion,
      (payload && payload.sheet) || '',
      (payload && payload.row != null) ? payload.row : 0,
      (payload && payload.col) || '',
      stringifySafe_(payload && payload.oldValue),
      stringifySafe_(payload && payload.newValue),
    ];

    sh.getRange(newRow, 1, 1, AUDIT_HEADERS.length).setValues([row]);

    // Auto-archivado cuando supera el umbral — no esperar al límite
    if (newRow >= AUDIT_ARCHIVE_THRESHOLD) {
      archiveAuditLogIfNeeded_(sh);
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

/**
 * Devuelve timestamp actual en formato ISO 8601 estricto con timezone.
 * Formato: 2026-09-29T15:32:00.123-05:00
 * @private
 * @returns {string}
 */
function nowIso_() {
  const tz = Session.getScriptTimeZone() || 'America/Lima';
  return Utilities.formatDate(new Date(), tz, "yyyy-MM-dd'T'HH:mm:ss.SSSXXX");
}

/**
 * Archiva las filas más antiguas cuando se acerca al límite.
 * Mantiene AUDIT_KEEP_ROWS más recientes; mueve el resto a una hoja
 * oculta con sufijo _archive_YYYY_MM.
 *
 * Idempotente: si ya hay archivo del mes actual, añade al existente.
 *
 * @param {GoogleAppsScript.Spreadsheet.Sheet} sh Hoja 99_Audit_Log
 * @private
 */
function archiveAuditLogIfNeeded_(sh) {
  try {
    const lastRow = sh.getLastRow();
    if (lastRow < AUDIT_ARCHIVE_THRESHOLD) return;

    const ss = SpreadsheetApp.getActiveSpreadsheet();
    const tz = Session.getScriptTimeZone() || 'America/Lima';
    const stamp = Utilities.formatDate(new Date(), tz, 'yyyy_MM');
    const archiveName = AUDIT_ARCHIVE_SHEET_PREFIX + stamp;

    // 1. Crear/obtener hoja de archivo
    let archiveSh = ss.getSheetByName(archiveName);
    if (!archiveSh) {
      archiveSh = ss.insertSheet(archiveName);
      archiveSh.getRange(1, 1, 1, AUDIT_HEADERS.length).setValues([AUDIT_HEADERS]);
      archiveSh.setTabColor('#1F2937');
      archiveSh.hideSheet();
    }

    // 2. Determinar rango a mover (filas 2 a lastRow - AUDIT_KEEP_ROWS)
    const archiveLast = lastRow - AUDIT_KEEP_ROWS;
    if (archiveLast < 2) return;

    const archiveStartArchiveRow = archiveSh.getLastRow() + 1;
    const sourceData = sh.getRange(2, 1, archiveLast - 1, AUDIT_HEADERS.length).getValues();
    archiveSh.getRange(archiveStartArchiveRow, 1, sourceData.length, AUDIT_HEADERS.length)
      .setValues(sourceData);

    // 3. Limpiar 99_Audit_Log: borrar filas 2 a archiveLast, dejar header + AUDIT_KEEP_ROWS
    sh.deleteRows(2, archiveLast - 1);

    console.info('audit_log_archived: moved ' + sourceData.length +
      ' rows to ' + archiveName + ', kept ' + AUDIT_KEEP_ROWS + ' recent rows');
  } catch (e) {
    console.error('audit_archive_failed: ' + e.message);
  }
}
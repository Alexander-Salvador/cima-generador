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
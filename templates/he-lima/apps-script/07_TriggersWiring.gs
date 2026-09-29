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

/** @constant {number} Lock timeout en ms — instalación de triggers puede tardar */
const TRIGGER_LOCK_TIMEOUT = 15000;

/**
 * Instala los 3 triggers del sistema.
 * Idempotente: si ya existen, los reemplaza en lugar de duplicar.
 *
 * Cuota Google: max 30 triggers/usuario. OVAV usa 3.
 *
 * Concurrencia: protegido con LockService.getDocumentLock() para evitar
 * duplicación cuando 2 admins ejecutan simultáneamente.
 *
 * @returns {void}
 */
function instalarTriggers() {
  const lock = LockService.getDocumentLock();
  lock.waitLock(TRIGGER_LOCK_TIMEOUT);
  try {
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
  } finally {
    if (lock && lock.hasLock && lock.hasLock()) {
      lock.releaseLock();
    }
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
/**
 * 00_AddOnMenu.gs
 * ----------------------------------------------------------------------------
 * Entry point del add-on HE Lima — CIMA.
 * Manifest declara addOns.homepageTrigger.runFunction = "onOpen",
 * por lo que esta función DEBE existir (runtime failure si falta).
 *
 * Estrategia 2026 (sin HTML):
 *   - Menú 100% nativo (Ui.createAddonMenu) — cero HtmlService
 *   - Bootstrap de dropdowns al abrir el sheet
 *   - Warm-up de caché para evitar primera lectura lenta
 *   - Auto-verificación de estructura (alerta si falta sheets)
 *
 * Authority: thavren (lead) · implementer: andres · review: marco
 * Status: stable · 2026-09-29
 */

/** @constant {string} Versión del add-on (visible en menú) */
const ADDON_VERSION = 'v0.3.0-stable';
const ADDON_NAME = 'CIMA HE Lima';

/**
 * onOpen(e) — entry point requerido por addOns.homepageTrigger en manifest.
 * Se ejecuta automáticamente al abrir el spreadsheet que contiene este add-on.
 *
 * @param {GoogleAppsScript.Events.SheetsOnOpenEvent} e Evento de apertura
 * @returns {void}
 */
function onOpen(e) {
  const ui = (e && e.authMode !== ScriptApp.AuthMode.NONE)
    ? SpreadsheetApp.getUi()
    : SpreadsheetApp.getUi();

  ui.createAddonMenu()
    .addItem('🔄 Refrescar Config (invalidar caché)', 'menuRefrescarConfig')
    .addItem('💾 Sync Manual (forzar save→load)', 'menuSyncManual')
    .addSeparator()
    .addSubMenu(ui.createMenu('📊 Dashboard')
      .addItem('Refrescar KPIs', 'menuRefrescarKPIs')
      .addItem('Refrescar Heatmap', 'menuRefrescarHeatmap'))
    .addSubMenu(ui.createMenu('🔍 Diagnóstico')
      .addItem('Verificar estructura (18 sheets)', 'menuVerificarEstructura')
      .addItem('Ver audit log (últimas 50)', 'menuVerAuditLog')
      .addItem('Probar conexión CacheService', 'menuTestCache'))
    .addSubMenu(ui.createMenu('⚙️ Configuración')
      .addItem('Poblar dropdowns si están vacíos', 'menuPoblarDropdowns')
      .addItem('Instalar/Reparar triggers', 'menuInstalarTriggers')
      .addItem('Quitar triggers', 'menuQuitarTriggers'))
    .addSeparator()
    .addItem('📖 Acerca de ' + ADDON_NAME + ' ' + ADDON_VERSION, 'menuAcercaDe')
    .addToUi();

  // Bootstrap no bloqueante: solo si el usuario es dueño o editor
  try {
    bootstrapOnOpen_();
  } catch (err) {
    // No mostrar error al usuario — solo log en consola para diagnóstico
    console.warn('onOpen bootstrap warning: ' + err.message);
  }
}

/**
 * Bootstrap ejecutado al abrir el sheet.
 * - Invalida caché para forzar re-lectura de config
 * - Puebla dropdowns si están vacíos
 * - Verifica estructura y alerta si falta sheets
 * @private
 */
function bootstrapOnOpen_() {
  // 1. Invalidar caché (lectura fresca al primer uso)
  if (typeof invalidateCache === 'function') {
    invalidateCache();
  }

  // 2. Poblar dropdowns si están vacíos
  poblarDropdownsSiVacio_();

  // 3. Verificación silenciosa de sheets (solo alerta si faltan)
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const resultado = verificarEstructuraSilencioso_(ss);

  if (!resultado.ok) {
    const ui = SpreadsheetApp.getUi();
    const sheetsFaltantes = resultado.faltantes.join(', ');
    const respuesta = ui.alert(
      '⚠️ Estructura incompleta',
      'Faltan ' + resultado.faltantes.length + ' sheets:\n\n' +
      sheetsFaltantes + '\n\n' +
      '¿Ejecutar deploy completo ahora? (recomendado)',
      ui.ButtonSet.YES_NO
    );
    if (respuesta === ui.Button.YES && typeof ejecutarDeployCompleto_ === 'function') {
      ejecutarDeployCompleto_();
    }
  }
}

/**
 * Puebla dropdowns de 00_Config si están vacíos.
 * - B2 (mes): default = mes actual
 * - B3 (año): default = año actual
 * - B4 (semana): default = W01
 * - B5 (persona): default = "TODOS"
 * - B6-B9 (factores): defaults 1.25/1.35/2.00/8
 * @private
 */
function poblarDropdownsSiVacio_() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_CFG || '00_Config');
  if (!sh) return;

  const ahora = new Date();
  const mesActual = ahora.getMonth() + 1; // 1-12
  const anioActual = ahora.getFullYear();

  // B2 mes
  const b2 = sh.getRange('B2').getValue();
  if (b2 === '' || b2 === null || b2 === undefined) {
    sh.getRange('B2').setValue(mesActual);
  }

  // B3 año
  const b3 = sh.getRange('B3').getValue();
  if (b3 === '' || b3 === null || b3 === undefined) {
    sh.getRange('B3').setValue(anioActual);
  }

  // B4 semana (W01-W53)
  const b4 = sh.getRange('B4').getValue();
  if (b4 === '' || b4 === null || b4 === undefined) {
    sh.getRange('B4').setValue('W01');
  }

  // B5 persona filtro
  const b5 = sh.getRange('B5').getValue();
  if (b5 === '' || b5 === null || b5 === undefined) {
    sh.getRange('B5').setValue('TODOS');
  }

  // B6-B9 factores (solo si están vacíos)
  const defaults = { B6: 1.25, B7: 1.35, B8: 2.00, B9: 8 };
  Object.keys(defaults).forEach(function (cell) {
    const val = sh.getRange(cell).getValue();
    if (val === '' || val === null || val === undefined) {
      sh.getRange(cell).setValue(defaults[cell]);
    }
  });

  // B11 timestamp
  if (typeof touchLastRefresh === 'function') {
    touchLastRefresh();
  }
}

// ─────────────────────────────────────────────────────────────────
// MENÚ HANDLERS — funciones invocadas desde items del menú
// ─────────────────────────────────────────────────────────────────

/**
 * 🔄 Refrescar Config — invalida caché y re-lee 00_Config.
 */
function menuRefrescarConfig() {
  if (typeof invalidateCache === 'function') {
    invalidateCache();
  }
  SpreadsheetApp.getUi().alert(
    '✅ Caché invalidada',
    'La próxima lectura de 00_Config!B2:B9 será desde la hoja.',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
  if (typeof logAccion === 'function') {
    logAccion(Session.getActiveUser().getEmail(), 'menu_refrescar_config', {
      sheet: SHEET_CFG || '00_Config',
      row: 0,
      col: '',
      oldValue: '',
      newValue: 'cache_invalidated',
    });
  }
}

/**
 * 💾 Sync Manual — fuerza save del mes activo y luego load.
 * Útil cuando onEdit no disparó (ej: edición via API externa).
 */
function menuSyncManual() {
  const ui = SpreadsheetApp.getUi();
  if (typeof getMes !== 'function' || typeof getAnio !== 'function') {
    ui.alert('❌ Error', 'Módulo 01_CfgLoader no disponible.', ui.ButtonSet.OK);
    return;
  }

  const mes = getMes();
  const anio = getAnio();

  if (mes === null || mes === undefined || anio === null || anio === undefined) {
    ui.alert(
      '⚠️ Config incompleta',
      'Define mes (B2) y año (B3) en 00_Config antes de sincronizar.',
      ui.ButtonSet.OK
    );
    return;
  }

  try {
    if (typeof saveGridToDB === 'function') {
      saveGridToDB(mes, anio);
    }
    if (typeof loadMonthToGrid === 'function') {
      loadMonthToGrid(mes, anio);
    }
    if (typeof touchLastRefresh === 'function') {
      touchLastRefresh();
    }
    ui.alert(
      '✅ Sync completo',
      'Mes ' + mes + '/' + anio + ' — save y load ejecutados sin error.',
      ui.ButtonSet.OK
    );
    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'menu_sync_manual', {
        sheet: '10_Registro_HE',
        row: 0,
        col: '',
        oldValue: '',
        newValue: 'sync_ok_' + anio + '_' + (mes < 10 ? '0' + mes : mes),
      });
    }
  } catch (err) {
    ui.alert('❌ Error en sync', err.message, ui.ButtonSet.OK);
    if (typeof logAccion === 'function') {
      logAccion(Session.getActiveUser().getEmail(), 'menu_sync_manual_error', {
        sheet: '10_Registro_HE',
        row: 0,
        col: '',
        oldValue: '',
        newValue: 'sync_error: ' + err.message,
      });
    }
  }
}

/**
 * 📊 Refrescar KPIs del Dashboard.
 */
function menuRefrescarKPIs() {
  if (typeof refrescarKPIs === 'function') {
    refrescarKPIs();
    SpreadsheetApp.getUi().alert('✅ Dashboard refrescado', '', SpreadsheetApp.getUi().ButtonSet.OK);
  } else {
    SpreadsheetApp.getUi().alert(
      '⚠️ Módulo no disponible',
      '05_DashboardRefresh no está cargado. Revisa el deploy.',
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  }
}

/**
 * 📊 Refrescar Heatmap del Dashboard.
 */
function menuRefrescarHeatmap() {
  if (typeof refrescarHeatmap === 'function') {
    refrescarHeatmap();
  } else {
    SpreadsheetApp.getUi().alert(
      '⚠️ Heatmap no implementado aún',
      'Pendiente: sprint Day 4.',
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  }
}

/**
 * 🔍 Verificar estructura — confirma que las 18 sheets existen y tienen headers.
 */
function menuVerificarEstructura() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const resultado = verificarEstructuraCompleto_(ss);

  if (resultado.ok) {
    SpreadsheetApp.getUi().alert(
      '✅ Estructura correcta',
      'Las 18 sheets están presentes:\n\n' +
      '• 6 base: 00_Config, 01_Feriados_PE, 02_Personas, 10_Registro_HE, 20_Dashboard, 99_Audit_Log\n' +
      '• 12 mensuales (lazy): 98_DB_2026_01..12\n\n' +
      'Hojas mensuales existentes: ' + resultado.dbHojasExistentes.join(', '),
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  } else {
    SpreadsheetApp.getUi().alert(
      '❌ Estructura incompleta',
      'Faltan ' + resultado.faltantes.length + ' sheets:\n\n' +
      resultado.faltantes.join('\n') +
      '\n\nEjecuta el deploy completo desde el editor para reparar.',
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  }
}

/**
 * 🔍 Ver audit log (últimas 50 entradas).
 */
function menuVerAuditLog() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName('99_Audit_Log');
  if (!sh) {
    SpreadsheetApp.getUi().alert('❌ 99_Audit_Log no existe.', '', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  const lastRow = sh.getLastRow();
  const startRow = Math.max(2, lastRow - 49);
  const numRows = lastRow - startRow + 1;
  const data = sh.getRange(startRow, 1, numRows, 8).getValues();

  let msg = 'Últimas ' + numRows + ' acciones:\n\n';
  data.reverse().forEach(function (row) {
    msg += row[0] + ' | ' + row[1] + ' | ' + row[2] + ' | ' + (row[7] || '') + '\n';
  });

  SpreadsheetApp.getUi().alert('📋 Audit Log', msg, SpreadsheetApp.getUi().ButtonSet.OK);
}

/**
 * 🔍 Probar CacheService.
 */
function menuTestCache() {
  try {
    const cache = CacheService.getDocumentCache();
    const testKey = 'he_lima_test_' + Date.now();
    cache.put(testKey, 'OK', 60);
    const val = cache.get(testKey);
    cache.remove(testKey);
    SpreadsheetApp.getUi().alert(
      val === 'OK' ? '✅ CacheService funcional' : '⚠️ Cache retornó valor inesperado',
      'Put/get/remove ejecutado correctamente.',
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  } catch (err) {
    SpreadsheetApp.getUi().alert('❌ CacheService error', err.message, SpreadsheetApp.getUi().ButtonSet.OK);
  }
}

/**
 * ⚙️ Poblar dropdowns (forzar — útil si quedó algún B vacío).
 */
function menuPoblarDropdowns() {
  poblarDropdownsSiVacio_();
  SpreadsheetApp.getUi().alert('✅ Dropdowns revisados', '', SpreadsheetApp.getUi().ButtonSet.OK);
}

/**
 * ⚙️ Instalar/Reparar triggers.
 */
function menuInstalarTriggers() {
  if (typeof instalarTriggers === 'function') {
    instalarTriggers();
    SpreadsheetApp.getUi().alert('✅ Triggers instalados', '', SpreadsheetApp.getUi().ButtonSet.OK);
  } else {
    SpreadsheetApp.getUi().alert('⚠️ 07_TriggersWiring no cargado.', '', SpreadsheetApp.getUi().ButtonSet.OK);
  }
}

/**
 * ⚙️ Quitar triggers.
 */
function menuQuitarTriggers() {
  if (typeof eliminarTriggers === 'function') {
    eliminarTriggers();
    SpreadsheetApp.getUi().alert('✅ Triggers removidos', '', SpreadsheetApp.getUi().ButtonSet.OK);
  }
}

/**
 * 📖 Acerca de — info del add-on.
 */
function menuAcercaDe() {
  SpreadsheetApp.getUi().alert(
    '📖 ' + ADDON_NAME + ' ' + ADDON_VERSION,
    'Add-on de tracking de Horas Extras para CIMA — Técnica Móvil Lima.\n\n' +
    'Arquitectura: working surface + storage particionado mensual\n' +
    'Sheets: 18 (6 base + 12 mensuales lazy)\n' +
    'Apps Script: 9 módulos (00..08)\n' +
    'Runtime: V8\n' +
    'Sin HTML — 100% nativo Google Sheets',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

// ─────────────────────────────────────────────────────────────────
// VERIFICACIÓN DE ESTRUCTURA — helpers
// ─────────────────────────────────────────────────────────────────

/**
 * Verificación silenciosa — solo retorna resultado, no alerta.
 * @param {GoogleAppsScript.Spreadsheet.Spreadsheet} ss
 * @returns {{ok:boolean, faltantes:string[], dbHojasExistentes:string[]}}
 * @private
 */
function verificarEstructuraSilencioso_(ss) {
  const SHEETS_REQUERIDAS = [
    '00_Config', '01_Feriados_PE', '02_Personas',
    '10_Registro_HE', '20_Dashboard', '99_Audit_Log'
  ];

  const faltantes = [];
  SHEETS_REQUERIDAS.forEach(function (name) {
    if (!ss.getSheetByName(name)) faltantes.push(name);
  });

  // DB mensual: solo reportar cuáles existen (las 12 son lazy)
  const dbHojasExistentes = [];
  for (let mes = 1; mes <= 12; mes++) {
    const mm = mes < 10 ? '0' + mes : '' + mes;
    const name = '98_DB_2026_' + mm;
    if (ss.getSheetByName(name)) dbHojasExistentes.push(name);
  }

  return {
    ok: faltantes.length === 0,
    faltantes: faltantes,
    dbHojasExistentes: dbHojasExistentes
  };
}

/**
 * Verificación completa — retorna detalle para diagnóstico.
 * @param {GoogleAppsScript.Spreadsheet.Spreadsheet} ss
 * @returns {{ok:boolean, faltantes:string[], dbHojasExistentes:string[], warnings:string[]}}
 * @private
 */
function verificarEstructuraCompleto_(ss) {
  const basico = verificarEstructuraSilencioso_(ss);
  const warnings = [];

  // Validar headers de sheets existentes
  const HEADERS_ESPERADOS = {
    '00_Config': ['Param', 'Valor', 'Notas'],
    '01_Feriados_PE': ['ID', 'Nombre', 'Tipo', 'Fecha', 'Es_Movil', 'Aplica_HE'],
    '02_Personas': ['ID_Persona', 'DNI', 'Apellidos_Nombres'],
    '10_Registro_HE': ['Fecha', 'ID_Persona', 'Horas_Total', 'Sync_Status'],
    '20_Dashboard': ['KPI', 'Valor', 'Unidad', 'Detalle'],
    '99_Audit_Log': ['Timestamp_ISO', 'Usuario', 'Acción']
  };

  Object.keys(HEADERS_ESPERADOS).forEach(function (sheetName) {
    const sh = ss.getSheetByName(sheetName);
    if (!sh) return;
    const headers = sh.getRange(1, 1, 1, HEADERS_ESPERADOS[sheetName].length).getValues()[0];
    HEADERS_ESPERADOS[sheetName].forEach(function (expected, idx) {
      if (headers[idx] !== expected) {
        warnings.push(sheetName + ' col[' + idx + '] esperado "' + expected + '" got "' + headers[idx] + '"');
      }
    });
  });

  return Object.assign({}, basico, { warnings: warnings });
}

/**
 * Ejecutar deploy completo — placeholder para bootstrap.
 * Llama a la función de deploy del script principal si existe.
 * @private
 */
function ejecutarDeployCompleto_() {
  if (typeof desplegarHELima === 'function') {
    desplegarHELima();
  } else if (typeof setupAllSheets === 'function') {
    setupAllSheets();
  }
}
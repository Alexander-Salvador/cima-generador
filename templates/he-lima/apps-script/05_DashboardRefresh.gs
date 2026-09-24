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
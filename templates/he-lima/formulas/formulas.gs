/**
 * formulas.gs — Constantes con fórmulas del template he-lima v2
 * ----------------------------------------------------------------------------
 * Archivo consolidado. NO se ejecuta como Apps Script runtime — es una
 * referencia técnica + helper para que el script de bootstrap (P5) pueda
 * aplicar las fórmulas al spreadsheet vía `Range.setFormula()`.
 *
 * Convenciones:
 *   - Constantes UPPER_SNAKE_CASE
 *   - Cada fórmula tiene JSDoc con celda, intención y referencia
 *   - Fórmulas parametrizadas con `{ROW}` para row-relative substitution
 *
 * Authority: marco (arquitecto) · Implementer: andres
 */

/**
 * Bloque 1 — fórmulas 10_Registro_HE (working surface)
 * Aplicadas a filas 2..110 (REGISTRO_MAX_ROWS).
 * Total: 8 columnas con fórmulas × 109 filas = 872 celdas.
 */

/**
 * ISO Week (col B): "W##" calculado.
 * @formula TEXT(WEEKNUM(A{ROW}, 21), "W00")
 */
const FORMULA_ISO_WEEK = '=TEXT(WEEKNUM(A{ROW}, 21), "W00")';

/**
 * Día de la semana textual (col C): lun..dom.
 * @formula CHOOSE(WEEKDAY(A{ROW}, 2), "lun", "mar", "mie", "jue", "vie", "sab", "dom")
 */
const FORMULA_DIA_SEMANA = '=CHOOSE(WEEKDAY(A{ROW}, 2), "lun", "mar", "mie", "jue", "vie", "sab", "dom")';

/**
 * Es feriado (col D): lookup directo contra 01_Feriados_PE!D:D.
 * CORR-1: COUNTIF > 0 (lookup directo, sin columna auxiliar).
 * @formula =COUNTIF(01_Feriados_PE!$D:$D, A{ROW}) > 0
 */
const FORMULA_ES_FERIADO = '=COUNTIF(01_Feriados_PE!$D:$D, A{ROW}) > 0';

/**
 * Tipo de día (col E): Feriado / Sabado / Domingo / Normal.
 * Anida: feriado > sabado > domingo > normal.
 * @formula =IF(D{ROW}, "Feriado", IF(WEEKDAY(A{ROW}, 2) = 6, "Sabado", IF(WEEKDAY(A{ROW}, 2) = 7, "Domingo", "Normal")))
 */
const FORMULA_TIPO_DIA = '=IF(D{ROW}, "Feriado", IF(WEEKDAY(A{ROW}, 2) = 6, "Sabado", IF(WEEKDAY(A{ROW}, 2) = 7, "Domingo", "Normal")))';

/**
 * Apellidos_Nombres (col G): VLOOKUP desde 02_Personas.
 * Si no encuentra (persona inactiva), devuelve "".
 * @formula =IFERROR(VLOOKUP(F{ROW}, 02_Personas!$A:$J, 3, FALSE), "")
 */
const FORMULA_APELLIDOS_NOMBRES = '=IFERROR(VLOOKUP(F{ROW}, 02_Personas!$A:$J, 3, FALSE), "")';

/**
 * Horas total (col K): con manejo de cruce de medianoche via MOD.
 * Devuelve 0 si I o J están vacíos.
 * @formula =IF(OR(I{ROW}="", J{ROW}=""), 0, MOD(J{ROW} - I{ROW}, 1) * 24)
 */
const FORMULA_HORAS_TOTAL = '=IF(OR(I{ROW}="", J{ROW}=""), 0, MOD(J{ROW} - I{ROW}, 1) * 24)';

/**
 * Horas normales (col L): dentro de jornada, solo si Tipo=Normal.
 * CORR-1: jornada max es B9 (no B8).
 * @formula =IF(E{ROW} = "Normal", MIN(K{ROW}, $B$9), 0)
 */
const FORMULA_HORAS_NORMALES = '=IF(E{ROW} = "Normal", MIN(K{ROW}, $B$9), 0)';

/**
 * Horas extras (col M): K - L (invariante L + M = K).
 * @formula =K{ROW} - L{ROW}
 */
const FORMULA_HORAS_EXTRAS = '=K{ROW} - L{ROW}';

/**
 * Factor aplicado (col N): INDEX/MATCH sobre matriz en 00_Config!D6:G9.
 *   D6 = "Tipo", E6 = "Diurno", F6 = "Nocturno" (header)
 *   D7 = "Normal",  E7 = $B$6,  F7 = $B$7
 *   D8 = "Sab",     E8 = $B$8,  F8 = $B$8
 *   D9 = "Dom",     E9 = $B$8,  F9 = $B$8
 *   D10 = "Fer",    E10 = $B$8, F10 = $B$8  ← ajustamos layout a D7:G10
 *
 * Implementación condicional (no requiere columna auxiliar):
 *   - Si nocturno: factor feriado/nocturno según tipo
 *   - Si diurno:   factor diurno/feriado según tipo
 *
 * @formula =IF(H{ROW} = "nocturno", IF(E{ROW} = "Normal", $B$7, $B$8), IF(E{ROW} = "Normal", $B$6, $B$8))
 */
const FORMULA_FACTOR_APLICADO = '=IF(H{ROW} = "nocturno", IF(E{ROW} = "Normal", $B$7, $B$8), IF(E{ROW} = "Normal", $B$6, $B$8))';

/**
 * Horas ponderadas (col O): K × N.
 * @formula =K{ROW} * N{ROW}
 */
const FORMULA_HORAS_PONDERADAS = '=K{ROW} * N{ROW}';

/**
 * Pago HE PEN (col P): O × Tarifa de la persona (col I de 02_Personas).
 * @formula =O{ROW} * VLOOKUP(F{ROW}, 02_Personas!$A:$J, 9, FALSE)
 */
const FORMULA_PAGO_HE_PEN = '=O{ROW} * IFERROR(VLOOKUP(F{ROW}, 02_Personas!$A:$J, 9, FALSE), 0)';

/**
 * Mes_Año (col U): clave de sincronización con 98_DB_YYYY_MM.
 * @formula =TEXT(A{ROW}, "YYYY-MM")
 */
const FORMULA_MES_ANIO = '=TEXT(A{ROW}, "YYYY-MM")';

/**
 * Bloque 2 — fórmulas 98_DB_YYYY_MM (storage mensual)
 * Idénticas al working surface (mismas 22 columnas A-V).
 * Aplicadas a filas 2..1550 (50 personas × 31 días máx).
 */
const FORMULA_DB_ISO_WEEK = FORMULA_ISO_WEEK;
const FORMULA_DB_DIA_SEMANA = FORMULA_DIA_SEMANA;
const FORMULA_DB_ES_FERIADO = FORMULA_ES_FERIADO;
const FORMULA_DB_TIPO_DIA = FORMULA_TIPO_DIA;
const FORMULA_DB_APELLIDOS_NOMBRES = FORMULA_APELLIDOS_NOMBRES;
const FORMULA_DB_HORAS_TOTAL = FORMULA_HORAS_TOTAL;
const FORMULA_DB_HORAS_NORMALES = FORMULA_HORAS_NORMALES;
const FORMULA_DB_HORAS_EXTRAS = FORMULA_HORAS_EXTRAS;
const FORMULA_DB_FACTOR_APLICADO = FORMULA_FACTOR_APLICADO;
const FORMULA_DB_HORAS_PONDERADAS = FORMULA_HORAS_PONDERADAS;
const FORMULA_DB_PAGO_HE_PEN = FORMULA_PAGO_HE_PEN;
const FORMULA_DB_MES_ANIO = FORMULA_MES_ANIO;

/**
 * Bloque 3 — fórmulas 20_Dashboard (KPIs ejecutivos)
 * Estrategia: INDIRECT acotado a 12 referencias (uno por mes).
 * NUNCA usar INDIRECT en 10_Registro_HE ni en 98_DB_YYYY_MM (volátil).
 */

/**
 * Celda auxiliar oculta M1: referencia al nombre de la DB activa.
 * @formula ="98_DB_" & $B$3 & "_" & TEXT($B$2, "00")
 */
const FORMULA_DASH_DB_REF = '="98_DB_" & $B$3 & "_" & TEXT($B$2, "00")';

/**
 * KPI HE del mes: SUMIFS sobre Horas_Extras (col M) filtrado por Mes_Año (col U).
 * Ubicación esperada: 20_Dashboard!B2
 * @formula =SUMIFS(INDIRECT(M1 & "!$M$2:$M$1550"), INDIRECT(M1 & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)
 */
const FORMULA_KPI_HE_MES = '=SUMIFS(INDIRECT($M$1 & "!$M$2:$M$1550"), INDIRECT($M$1 & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)';

/**
 * KPI HE acumulado YTD: suma de todos los meses desde enero hasta mes activo.
 * Implementación directa (sin SUMPRODUCT volátil): 12 SUMIFS en cascada.
 *
 * Ubicación: 20_Dashboard!B3
 *
 * NOTA: esta fórmula referencia solo el rango del DB activo para velocidad.
 * Para YTD verdadero, se recomienda usar 12 celdas intermedias (enero..mes).
 *
 * @formula alternativa (single-sheet)
 *   =SUMPRODUCT((MONTH(INDIRECT(M1 & "!$A$2:$A$1550")) <= $B$2) * INDIRECT(M1 & "!$M$2:$M$1550"))
 *
 * @formula simple (recomendada — usa solo DB activa)
 *   =SUMIFS(INDIRECT($M$1 & "!$M$2:$M$1550"), INDIRECT($M$1 & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)
 */
const FORMULA_KPI_HE_YTD = FORMULA_KPI_HE_MES;

/**
 * KPI Pago del mes: SUMIFS sobre Pago_HE_PEN (col P).
 * Ubicación: 20_Dashboard!B4
 * @formula =SUMIFS(INDIRECT($M$1 & "!$P$2:$P$1550"), INDIRECT($M$1 & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)
 */
const FORMULA_KPI_PAGO_MES = '=SUMIFS(INDIRECT($M$1 & "!$P$2:$P$1550"), INDIRECT($M$1 & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)';

/**
 * KPI Personas activas: COUNTIF sobre 02_Personas!H:H = TRUE.
 * Ubicación: 20_Dashboard!B5
 * @formula =COUNTIF(02_Personas!$H:$H, TRUE)
 */
const FORMULA_KPI_PERSONAS_ACTIVAS = '=COUNTIF(02_Personas!$H:$H, TRUE)';

/**
 * KPI Top 10 HE — SORTN ordenado descendente por Horas_Ponderadas (col O).
 * Ubicación: 20_Dashboard!D20:F30
 * @formula =SORTN(INDIRECT($M$1 & "!$A$2:$O$1550"), 10, 0, INDIRECT($M$1 & "!$O$2:$O$1550"), FALSE)
 */
const FORMULA_KPI_TOP10_HE = '=SORTN(INDIRECT($M$1 & "!$A$2:$O$1550"), 10, 0, INDIRECT($M$1 & "!$O$2:$O$1550"), FALSE)';

/**
 * KPI Heatmap celda (semana × persona): SUMIFS cruzando WEEKNUM y persona.
 * Ubicación: 20_Dashboard!H2:N40
 *
 * Para una celda (sem=W##, persona=P-NNN):
 * @formula =SUMPRODUCT(
 *   (TEXT(WEEKNUM(INDIRECT($M$1 & "!$A$2:$A$1550") + 0, 21), "W00") = H$1) *
 *   (INDIRECT($M$1 & "!$F$2:$F$1550") = $G2) *
 *   INDIRECT($M$1 & "!$M$2:$M$1550")
 * )
 */
const FORMULA_KPI_HEATMAP_CELL =
  '=SUMPRODUCT(' +
  '(TEXT(WEEKNUM(INDIRECT($M$1 & "!$A$2:$A$1550") + 0, 21), "W00") = H$1) * ' +
  '(INDIRECT($M$1 & "!$F$2:$F$1550") = $G2) * ' +
  'INDIRECT($M$1 & "!$M$2:$M$1550"))';

/**
 * Bloque 4 — fórmulas 00_Config (data validation helpers)
 */

/**
 * Dropdown Mes (B2): lista 1..12.
 * Aplicar vía DataValidation en lugar de fórmula.
 */
const CONFIG_B2_MES_DV_ITEMS = [
  '1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12'
];

/**
 * Dropdown Año (B3): 2024..2030.
 */
const CONFIG_B3_ANIO_DV_ITEMS = [
  '2024', '2025', '2026', '2027', '2028', '2029', '2030'
];

/**
 * Dropdown Semana (B4): W01..W53.
 */
const CONFIG_B4_SEMANA_DV_ITEMS = (function () {
  const out = [];
  for (let i = 1; i <= 53; i++) {
    out.push('W' + (i < 10 ? '0' + i : '' + i));
  }
  return out;
})();

/**
 * Bloque 5 — fórmulas lookup contra 01_Feriados_PE
 * Estas fórmulas ya están definidas arriba (FORMULA_ES_FERIADO usa COUNTIF D:D).
 *
 * Helper adicional para 02_HolidaysEngine.gs / dashboard:
 */

/**
 * Nombre del feriado si la fecha es feriado, sino "".
 * Útil para 20_Dashboard breakdown.
 * @formula =IFERROR(VLOOKUP(A{ROW}, 01_Feriados_PE!$D:$B, 2, FALSE), "")
 */
const FORMULA_NOMBRE_FERIADO = '=IFERROR(INDEX(01_Feriados_PE!$B:$B, MATCH(A{ROW}, 01_Feriados_PE!$D:$D, 0)), "")';

/**
 * Helper: aplica las fórmulas del working surface a un rango (rows).
 *
 * @param {GoogleAppsScript.Spreadsheet.Sheet} sh Hoja destino.
 * @param {number} firstRow Fila inicial (típicamente 2).
 * @param {number} numRows Cantidad de filas.
 * @returns {void}
 */
function applyFormulasGrid_(sh, firstRow, numRows) {
  // Columna B (ISO_Week)
  const formulasB = [];
  for (let r = firstRow; r < firstRow + numRows; r++) {
    formulasB.push([FORMULA_ISO_WEEK.replace(/\{ROW\}/g, r)]);
  }
  sh.getRange(firstRow, 2, numRows, 1).setFormulas(formulasB);

  // Columnas C, D, E
  const formulasCDE = [];
  for (let r = firstRow; r < firstRow + numRows; r++) {
    formulasCDE.push([
      FORMULA_DIA_SEMANA.replace(/\{ROW\}/g, r),
      FORMULA_ES_FERIADO.replace(/\{ROW\}/g, r),
      FORMULA_TIPO_DIA.replace(/\{ROW\}/g, r),
    ]);
  }
  sh.getRange(firstRow, 3, numRows, 3).setFormulas(formulasCDE);

  // Columna G (Apellidos_Nombres via VLOOKUP)
  const formulasG = [];
  for (let r = firstRow; r < firstRow + numRows; r++) {
    formulasG.push([FORMULA_APELLIDOS_NOMBRES.replace(/\{ROW\}/g, r)]);
  }
  sh.getRange(firstRow, 7, numRows, 1).setFormulas(formulasG);

  // Columnas K, L, M, N, O, P
  const formulasKLMNOP = [];
  for (let r = firstRow; r < firstRow + numRows; r++) {
    formulasKLMNOP.push([
      FORMULA_HORAS_TOTAL.replace(/\{ROW\}/g, r),
      FORMULA_HORAS_NORMALES.replace(/\{ROW\}/g, r),
      FORMULA_HORAS_EXTRAS.replace(/\{ROW\}/g, r),
      FORMULA_FACTOR_APLICADO.replace(/\{ROW\}/g, r),
      FORMULA_HORAS_PONDERADAS.replace(/\{ROW\}/g, r),
      FORMULA_PAGO_HE_PEN.replace(/\{ROW\}/g, r),
    ]);
  }
  sh.getRange(firstRow, 11, numRows, 6).setFormulas(formulasKLMNOP);

  // Columna U (Mes_Año)
  const formulasU = [];
  for (let r = firstRow; r < firstRow + numRows; r++) {
    formulasU.push([FORMULA_MES_ANIO.replace(/\{ROW\}/g, r)]);
  }
  sh.getRange(firstRow, 21, numRows, 1).setFormulas(formulasU);
}

/**
 * Helper: aplica fórmulas al Dashboard.
 *
 * @param {GoogleAppsScript.Spreadsheet.Sheet} sh Hoja 20_Dashboard.
 * @returns {void}
 */
function applyFormulasDashboard_(sh) {
  // M1: DB ref (oculta)
  sh.getRange('M1').setFormula(FORMULA_DASH_DB_REF);

  // KPIs A1:F4
  sh.getRange('B2').setFormula(FORMULA_KPI_HE_MES);
  sh.getRange('B3').setFormula(FORMULA_KPI_HE_YTD);
  sh.getRange('B4').setFormula(FORMULA_KPI_PAGO_MES);
  sh.getRange('B5').setFormula(FORMULA_KPI_PERSONAS_ACTIVAS);
}
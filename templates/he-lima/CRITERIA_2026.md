# CIMA 2026 — Criterios para Potencia Real en Google Sheets
**Análisis razonado · 2026-09-25 · Thavren / OVAV Platform Engineering**
**Fuentes:** workspaceupdates.googleblog.com, developers.google.com, Apps Script docs, Excel desktop analizado

---

## 🎯 Pregunta CEO

> "¿Qué necesitamos en 2026 para obtener potencia real en Google Sheets e implementar lo que se puede hacer en Excel desktop?"

**Caso de uso:** Registro de HE con detección de feriados, cambio por año/mes, valorización 25%/35%.

---

## 🔬 Investigación: Estado del Arte Google Sheets 2026

### Apps Script 2026 — Ya es **Core Service**

| Hito | Fecha | Implicación |
|---|---|---|
| **Apps Script = Workspace core service** | Jun 2026 | Mismas garantías de protección de datos, soporte técnico enterprise. Sin restricciones legacy. |
| **Data regions para Apps Script** | Sep 2026 | Compliance geográfico (US/EU/etc). CIMA puede correr en región Latam. |
| **Granular OAuth consent** | Ene 2025 | Sub-scopes (Sheets sí, Forms no). Menos fricción de re-auth. |
| **Project History UI** | Ago 2023 | Diff entre versiones desplegadas del script. |
| **IDE rediseñado + Script Properties + Time Zone** | Abr 2022 | Funcionalidad parity con legacy IDE. |
| **V8 runtime (default)** | Feb 2020+ | Modern JavaScript, no Rhino. |

### Sheets Nativo 2026 — Capacidades Maduras

| Capacidad nativa | GA desde | Uso CIMA |
|---|---|---|
| **`ARRAYFORMULA`** | 2017 | 1 fórmula → N filas. HE Lima entero sin copy-paste. |
| **`QUERY`** (SQL-like) | 2014 | `SELECT nombre, SUM(horas) WHERE mes=X GROUP BY` |
| **`LET`** (variables) | 2022 | Fórmulas legibles, 30-50% más rápidas. |
| **`LAMBDA`** (funciones custom) | 2022 | Helper reutilizable: `JORNADA_MAX()`, `FACTOR_HORAS_EXTRA()`. |
| **`SEQUENCE`** | 2019 | Generar rango de fechas del mes sin helper externo. |
| **`LAMBDA` + `MAP`/`REDUCE`** | 2022 | Programación funcional en celdas. |
| **`TEXTBEFORE`/`TAFTER`** | 2022 | Parsing de strings (sin regex manual). |
| **`XLOOKUP`** | 2022 | Reemplazo moderno de VLOOKUP. |
| **`COUNTIFS`/`SUMIFS`** | GA | Filtros multi-condición. |
| **`Data Validation` con dropdowns** | GA | Mes/Area/Empleado dropdown. |
| **`Conditional Formatting` avanzado** | GA | Escala de color, iconos, fórmulas custom. |
| **`Protected ranges`** | GA | Config inmutable. |
| **`Named ranges`** | GA | `=tarifa_hora` en vez de `$B$3`. |
| **`IMPORTRANGE`** | GA | Sync entre spreadsheets sin OAuth. |
| **`IMPORTDATA`** | GA | Traer CSVs públicos (URL fija). |
| **`Sparkline`** | 2018 | Mini-gráficos en celda. |
| **`Version History`** (built-in) | GA | Audit log sin código, 30-100 versiones. |
| **Tab color/icon** | GA | Organización visual. |
| **`Slicers`** en pivot tables | 2021 | Filtros visuales clickables. |
| **`Pivot tables` con calculated fields** | 2021 | Custom aggregations. |

### Integraciones Enterprise 2026

| Integración | Estado | Uso CIMA |
|---|---|---|
| **Connected Sheets (BigQuery)** | GA 2020+ | Si crece >100K filas, BigQuery detrás de Sheets sin tocar UX. |
| **CSV → BigQuery directo en Sheets UI** | Nov 2025 | Importar >10K filas sin Dataflow. |
| **Looker Studio (BI dashboards)** | GA | Dashboards ejecutivos desde Sheets. |
| **AppSheet (no-code apps)** | GA | Mobile data entry desde Sheets. |
| **Workspace add-ons framework** | Jul 2025 | Build-once run-en-C Sheets, Chat, Calendar, Drive. |
| **Gemini in Workspace** | GA 2024 | AI para fórmulas: "=GEMINI()". |

### Apps Script 2026 — Cuándo SÍ Usar

| Caso | Apps Script | Razón |
|---|---|---|
| Email programado al supervisor | ✅ Sí | No nativo |
| Auto-archivar mes al cerrar | ✅ Sí | Trigger onEdit/time-based |
| API externa (RENIEC, SUNAT) | ✅ Sí | UrlFetch + parsing |
| PDF generation | ✅ Sí | No nativo |
| Webhook Slack/Teams | ✅ Sí | UrlFetch |
| Apps para mobile (UI propia) | ✅ Sí | HtmlService |
| **Recálculo de fórmulas onEdit** | ❌ NO | Fórmulas recalculan en tiempo real (cada 5 min forzado, o instant en Sheets) |
| **Generación de fechas de mes** | ❌ NO | `=SEQUENCE(1, DAY(EOMONTH(...)), DATE(...), 1)` nativo |
| **Detección de feriados** | ❌ NO | `=COUNTIF(Feriados!A:A, fecha)` |
| **Audit log** | ❌ NO | Version History built-in |
| **Protección de celdas** | ❌ NO | Protected ranges nativo |

---

## ⚖️ Gap Analysis: Excel Desktop vs Google Sheets 2026

### ✅ FÁCIL en Sheets (1:1 o mejor)

| Capacidad Excel | Sheets 2026 | Notas |
|---|---|---|
| `DATE()` / `EOMONTH()` | ✅ Idéntico | |
| `VLOOKUP` | ✅ + `XLOOKUP` | XLOOKUP es estrictamente superior |
| `SUM` / `SUMIFS` / `COUNTIFS` | ✅ Idéntico | |
| `IF` / `AND` / `OR` | ✅ Idéntico | |
| Data Validation (dropdowns) | ✅ Idéntico | |
| Conditional Formatting | ✅ Idéntico + icon set | |
| Filtros / Sort | ✅ Idéntico | |
| Pivot tables | ✅ Idéntico | |
| Macros grabados | ✅ Apps Script macro | |
| Multiple sheets | ✅ Idéntico | |
| **LAMBDA custom functions** | ✅ Nativo | Sin VBA |
| **Gráficos** | ✅ Idéntico + sparkline | |
| **Power Query (ETL)** | ⚠️ Limitado | Usar BigQuery o QUERY |
| **Power Pivot (DAX)** | ❌ No | Solo SQL-like QUERY |

### ⚠️ DIFICULTAD MEDIA en Sheets

| Capacidad Excel | Sheets 2026 | Mitigación |
|---|---|---|
| **Macros VBA complejas** | Apps Script + grabación | Plan B: reescribir como fórmulas |
| **Tablas dinámicas con jerarquías** | Pivot tables nativos | Menos flexible que Excel |
| **Power Query / Get & Transform** | QUERY function | Menos poderoso, sin GUI |
| **100K+ filas** | Limit 10M cells | Usar Connected Sheets → BigQuery |
| **Imágenes en celda** | `IMAGE()` function | Sí, pero menos control |
| **Comentarios por celda** | Sí nativos | |
| **Sparklines nativas** | `SPARKLINE()` function | |

### ❌ LIMITACIONES Sheets 2026 (no tienen equivalente)

| Excel | Sheets 2026 | Workaround |
|---|---|---|
| **VBA puro (UDF, eventos)** | Apps Script (no VBA) | Reescribir en JS |
| **Power BI nativo** | Looker Studio (separado) | |
| **Power Query en desktop offline** | No | Sheets siempre online |
| **Análisis de hipótesis / Solver** | LinearOptimizationService (Apps Script) | Solo LP, no IP/NLP |
| **Formato condicional avanzado (barra datos)** | Sí pero limitado | |
| **Macros offline** | No | Sheets es cloud-first |
| **CSV import wizard** | Sí, básico | Para avanzado → BigQuery |

---

## 📐 Patrón del Excel compartido: Mapeo a Sheets 2026

### Excel tiene:
- **Sheet `Datos`**: 17 empleados, 10 cargos, 4 áreas, salarios
- **3 sheets mensuales** (`Registro`, `Hoja1`, `registro.`): feb, abr, ago 2025
- **Headers dinámicos**: Mes (dropdown), Año (input), Área, Período
- **Fechas auto-generadas**: D-AH (28/29/30/31 cols)
- **Día semana derivado**: vie/sáb/dom/lun...
- **Input horas**: por empleado × por día
- **Cálculos sumarios**: H_Normal (≤8h), H_Extras (>8h), H_25% (laborable), H_35% (feriado/domingo)
- **Valorización**: `$_normal = H × tarifa/8`, `$_extra = H_25% × 1.25 + H_35% × 1.35`
- **VLOOKUP auto-poblar**: "Totales" row al final

### Mapeo Sheets 2026 (todo nativo):

| Excel | Sheets 2026 equivalente |
|---|---|
| 3 sheets mensuales | **1 sheet anual** con `SEQUENCE` por día del mes |
| Mes hardcodeado en celda | **Dropdown validado** → `CONFIG!B2` |
| Día semana manual | `=TEXT(fecha, "DDD")` (auto) |
| H_Normal ≤8h | `=MIN(D5, CONFIG!$B$5)` |
| H_Extras >8h | `=MAX(D5 - CONFIG!$B$5, 0)` |
| H_25% (laborable) | `=IF(AND(NO feriado, WEEKDAY(fecha,2)<6), extra_dia, 0)` |
| H_35% (feriado/domingo) | `=IF(OR(feriado, WEEKDAY(fecha,2)>=6), extra_dia, 0)` |
| Detección feriado | `=IF(COUNTIF(Feriados!A:A, fecha)>0, "FERIADO", "")` |
| VLOOKUP "Totales" | `=QUERY(HE!data, "SELECT SUM(J) LABEL SUM(J) ''", 1)` |
| `$*` valorización | `=LET(horas, D5, jornada, $B$5, tarifa, $B$6, IF(horas<=jornada, horas*tarifa/jornada, (horas-jornada)*tarifa*factor/jornada))` |
| Catálogo empleados | Sheet `PERSONAL` con Data Validation |
| Catálogo cargos/áreas | Named ranges con dropdowns |

---

## 🎯 Criterios para "Potencia Real" 2026

### **CRITERIO 1: Nativo sobre Script por defecto**
> Antes de escribir Apps Script, ¿se puede hacer con fórmulas + named ranges + data validation?

**Test:** Cualquier cálculo < 1000 caracteres en 1 fórmula `LET` → nativo. Si requiere loop > 10 iteraciones o web request → Apps Script.

### **CRITERIO 2: Una sola fuente de verdad por concepto**
> Personal, Config, Feriados viven en sheets dedicadas. Otras sheets leen vía `VLOOKUP/XLOOKUP`/`IMPORTRANGE`.

**Anti-patrón:** Copiar datos entre sheets. Cada cambio debe propagar.

### **CRITERIO 3: Inputs son celdas nombradas, NO macros**
> `Mes`, `Año`, `JornadaMax`, `Factor25`, `Factor35` = named ranges en `CONFIG`.

**Beneficio:** Cambiar 1 celda → toda la sheet reacciona. Sin recálculo manual.

### **CRITERIO 4: Cambiar mes/año = cambiar 2 celdas, no rehacer fórmulas**
> Toda la estructura de fechas se genera con `SEQUENCE(1, DAY(EOMONTH(...)), ...)`. No se hardcodean columnas.

**Test:** Cambiar CONFIG!B2 (Mes) → toda la vista se ajusta automáticamente.

### **CRITERIO 5: Detección de feriados sin Apps Script**
> Catálogo multi-año `01_FERIADOS_PE` con Pascua calculada (`=DATE(...)`). Marcar feriados vía `COUNTIF` + conditional formatting.

**Beneficio:** Actualizar catálogo 1 vez al año, no tocar fórmulas.

### **CRITERIO 6: Audit por versión, no por script**
> Usar File → Version History (built-in). Apps Script `onChange` solo si se necesita notificar.

**Beneficio:** Cero código para audit. 30-100 versiones automáticas.

### **CRITERIO 7: Validaciones declarativas en capa de datos**
> `Data Validation` con `ONE_OF_LIST` o `CUSTOM_FORMULA`. Inputs siempre validados antes de cálculo.

**Beneficio:** No se puede tipear "25 de febrero" si no existe esa fecha.

### **CRITERIO 8: SUMARIOS via QUERY, no via VLOOKUP anidados**
> `=QUERY(HE!data, "SELECT nombre, SUM(horas), SUM(_) WHERE mes=$B$2 GROUP BY nombre")`.

**Beneficio:** Una fórmula vs 20+. Refactoriza con 1 cambio.

### **CRITERIO 9: Apps Script SOLO si la capacidad no es nativa**
> Regla de los 6 casos: email programado, auto-archivo, API externa, PDF, webhook, mobile app.

**Anti-patrón:** Apps Script que solo llama `SpreadsheetApp.getActiveSpreadsheet().getSheetByName(...)` — eso es nativo.

### **CRITERIO 10: Schema versionado en git**
> Cada cambio de estructura (nueva sheet, nuevo campo, nueva fórmula) → commit `feat(sheets-cima): schema v1.1 — +column X`.

**Beneficio:** Rollback documentado. Migración testeable.

### **CRITERIO 11: Performance budget = < 2s recalc**
> Máximo 30 empleados × 31 días × 5 fórmulas = 4,650 celdas activas. Sheets aguanta 100x más.

**Test:** Cualquier sheet con fórmulas debe recalcular en < 2s (test manual en browser).

### **CRITERIO 12: Documentación operativa, no técnica**
> `RUNBOOK.md` para CEO: "Cómo agregar empleado", "Cómo cambiar factor 25%". Sin jerga Apps Script.

**Beneficio:** CEO puede mantener sin depender de Thavren.

---

## 🛣️ Roadmap Implementación

### **Fase 0: Decisión Arquitectónica** (1 sesión, no tocar sheet)
- [ ] CEO aprueba los 12 criterios
- [ ] Decidir: ¿Apps Script ON o OFF?
  - **Default:** OFF. Solo se prende si CEO quiere email programado o auto-archive
- [ ] Decidir estructura: 4 sheets vs monolítica vs anual

### **Fase 1: Schema Canónico** (2-3 sesiones)
- [ ] Diseñar `00_CONFIG` con named ranges
- [ ] Diseñar `01_FERIADOS_PE` con Pascua calculada para 2025-2030
- [ ] Sincronizar `PERSONAL` actual (21 agentes) con catálogo del Excel si aplica
- [ ] Diseñar `10_HE` (vista única anual con SEQUENCE)
- [ ] Diseño de `20_DASHBOARD` con KPIs via QUERY

### **Fase 2: Fórmulas + Validaciones** (3-4 sesiones)
- [ ] Implementar detección feriados con COUNTIF + CF
- [ ] Implementar valorización 25%/35% con LET
- [ ] Data Validation: Mes, Año, Empleado, TipoHE
- [ ] Protected ranges: CONFIG, FERIADOS
- [ ] Named ranges: `jornada_max`, `factor_25`, `factor_35`, `tarifa_*`

### **Fase 3: Dashboards** (2 sesiones)
- [ ] KPI cards: total H, $ total, top excedidos, ranking áreas
- [ ] Charts: tendencia mensual, heatmap día × empleado
- [ ] Sparklines por agente

### **Fase 4: Hardening** (2-3 sesiones)
- [ ] Tests con datos de muestra (los 3 meses del Excel compartido)
- [ ] Snapshot pre-deploy
- [ ] Rollback playbook
- [ ] RUNBOOK.md

### **Fase 5: Producción** (1 sesión)
- [ ] Limpieza celdas de muestra
- [ ] Datos reales 2026-Q4 desde PERSONAL
- [ ] Freeze panes, tabs color, iconografía

**Total estimado:** 12-15 sesiones (~4 semanas working lento).

---

## ⚖️ Trade-offs Reconocibles

| Decisión | Pro | Con |
|---|---|---|
| **Solo nativo (sin Apps Script)** | Simple, mantenible, sin re-auth | Sin notificaciones, sin auto-archivo |
| **1 vista anual vs 12 mensuales** | Sin duplicación, sin caos | 1 sheet con 365+ cols (legible pero denso) |
| **Feriados en sheet dedicada** | Editable sin tocar código | Hay que actualizar 1 vez al año |
| **Version History en vez de audit log Apps Script** | Cero código | Solo cambia si guardas manualmente cada cambio importante |
| **PERSONAL como catálogo (no import)** | Single source of truth | Si editan PERSONAL, cambia toda la sheet |

---

## 🎬 Recomendación Final

> **Google Sheets 2026 puede hacer el 100% del Excel compartido, mejor, con menos código, y de forma colaborativa.**

**Stack ganador para CIMA 2026:**

```
Stack CIMA 2026:
├── Sheets (4 sheets: CONFIG, FERIADOS, PERSONAL, HE + Dashboard)
├── Fórmulas nativas (LET, LAMBDA, QUERY, XLOOKUP, ARRAYFORMULA)
├── Data Validation (dropdowns + custom formulas)
├── Protected Ranges + Named Ranges
├── Conditional Formatting
├── Version History (audit built-in)
├── Apps Script (opcional, solo si CEO pide email/auto-archive)
└── Connected Sheets / BigQuery (Fase 2+ si crece >100K filas)
```

**Próximo paso CEO:** Aprobar los 12 criterios + elegir Fase 0 (decisión arquitectónica sin tocar el sheet).

---

*Documento de análisis razonado · CRIT-009 (línea base avanzado+, no funcional) · CRIT-002 (observable result = contract)*
*OVAV Platform Engineering · Thavren · 2026-09-25*
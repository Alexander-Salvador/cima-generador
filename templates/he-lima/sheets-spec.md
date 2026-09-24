# Sheets Spec — he-lima (v2 — design-revised)

> **Versión 2** — 2026-09-24 · Authority: marco
> **Cambios clave vs v1:**
> - 18 sheets totales (6 base + 12 mensuales)
> - `10_Registro_HE`不再是 fact table → **working surface** (~110 filas)
> - 12 hojas mensuales `98_DB_YYYY_MM` ocultas, auto-creadas lazy
> - `01_Feriados_PE` ahora **visible** (CORR-4)
> - Anchors `00_Config` reorganizados: B2 mes, B3 año, B4 semana, B5 persona

---

## Índice de sheets (18 total)

| Sheet ID | Nombre | Propósito | Rango activo | Filas header | Color tab | Freeze | Visibilidad |
|---|---|---|---|---|---|---|---|
| `cfg` | `00_Config` | Panel de control + anclas globales | `A1:C60` | fila 1 | `#1F2937` | `A2` | protegida |
| `holidays` | `01_Feriados_PE` | Catálogo feriados PE (13 nacionales) | `A1:F25` | fila 1 | `#DC2626` | `A2` | **visible** (CORR-4) |
| `personas` | `02_Personas` | Maestro de personas (FK) | `A1:J60` | fila 1 | `#2563EB` | `A2` | visible |
| `registro` | `10_Registro_HE` | **Working surface** — grid editable | `A1:V110` | fila 1 | `#059669` | `A2:C2` | visible |
| `db_2026_01` | `98_DB_2026_01` | Storage Enero 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_02` | `98_DB_2026_02` | Storage Febrero 2026 | `A1:V1450` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_03` | `98_DB_2026_03` | Storage Marzo 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_04` | `98_DB_2026_04` | Storage Abril 2026 | `A1:V1500` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_05` | `98_DB_2026_05` | Storage Mayo 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_06` | `98_DB_2026_06` | Storage Junio 2026 | `A1:V1500` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_07` | `98_DB_2026_07` | Storage Julio 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_08` | `98_DB_2026_08` | Storage Agosto 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_09` | `98_DB_2026_09` | Storage Septiembre 2026 | `A1:V1500` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_10` | `98_DB_2026_10` | Storage Octubre 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_11` | `98_DB_2026_11` | Storage Noviembre 2026 | `A1:V1500` | fila 1 | `#374151` | `A2` | oculta |
| `db_2026_12` | `98_DB_2026_12` | Storage Diciembre 2026 | `A1:V1550` | fila 1 | `#374151` | `A2` | oculta |
| `dashboard` | `20_Dashboard` | KPIs ejecutivos + gráficos | `A1:L40` | — | `#7C3AED` | `A2` | visible |
| `audit` | `99_Audit_Log` | Log append-only de cambios | `A1:H10000` | fila 1 | `#374151` | `A2` | oculta |

---

## 00_Config — Panel de control (anchors actualizados)

| Col | Header | Tipo | Ancla | Notas |
|---|---|---|---|---|
| A | Param | text | — | etiqueta |
| B | Valor | mixed | ver tabla abajo | dropdowns + factores |
| C | Notas | text | — | documentación inline |

**Anclas críticas (referenciadas por Apps Script):**

| Celda | Significado | Default | Cambio vs v1 |
|---|---|---|---|
| `B2` | Mes activo (1-12) | mes actual | — |
| `B3` | Año activo (YYYY) | año actual | — |
| `B4` | Semana activa (W01-W53) | W actual | **NUEVO** (CORR-1) |
| `B5` | Persona filtro (`TODOS` o `P-NNN`) | `TODOS` | movido de B4→B5 |
| `B6` | Factor HE diurno | `1.25` | movido de B5→B6 |
| `B7` | Factor HE nocturno | `1.35` | movido de B6→B7 |
| `B8` | Factor feriado | `2.00` | movido de B7→B8 |
| `B9` | Jornada max diaria (h) | `8` | movido de B8→B9 |
| `B11` | Timestamp último refresh | `now()` | movido de B10→B11 |

> **Cambio crítico**: las fórmulas en `10_Registro_HE` que referenciaban `$B$8` (jornada) ahora referencian `$B$9`. El módulo `01_CfgLoader.gs` debe actualizar su mapa de anchors.

---

## 01_Feriados_PE — Catálogo de feriados (13 validados D.L. 713)

| Col | Header | Tipo | Pattern / Enum |
|---|---|---|---|
| A | ID | text | `^F-\d{2}$` |
| B | Nombre | text | nombre canónico |
| C | Tipo | enum | `nacional_fijo \| nacional_movil` (sin sectorial/regional) |
| D | Fecha | date | `yyyy-MM-dd` |
| E | Es_Movil | bool | true si derivado de Pascua |
| F | Aplica_HE | bool | true = paga factor feriado (los 13 son true) |

> **Seed inicial**: ver `holidays-pe.yaml` (13 feriados: 11 fijos + 2 móviles).
> **Visibilidad**: visible para CEO como referencia (CORR-4).
> **Eliminado**: Carnaval (F-13 v1), Viernes Santo duplicado (F-14 v1), San Pedro opt-in (F-15 v1), Aniversario Lima (F-16 v1).

---

## 02_Personas — Maestro de personas (sin cambios vs v1)

| Col | Header | Tipo | Validación | Notas |
|---|---|---|---|---|
| A | ID_Persona | text | regex `^P-\d{3}$` | PK lógica |
| B | DNI | text | regex `^\d{8}$` | único |
| C | Apellidos_Nombres | text | min 5 chars | mayúsculas |
| D | Cargo | enum | `tecnico \| supervisor \| jefe_cuadrilla \| admin` | |
| E | Cuadrilla | enum | `A \| B \| C \| D` | |
| F | Turno_Base | enum | `diurno \| nocturno \| rotativo` | |
| G | Fecha_Ingreso | date | ≤ hoy | |
| H | Activo | bool | — | drive filtros dropdown |
| I | Tarifa_Hora_PEN | number | decimal(8,2), ≥ 0 | PEN/h base |
| J | Notas | text | — | libre |

---

## 10_Registro_HE — Working Surface (CORR-1: NO es fact table)

**Modelo**: superficie de trabajo visible ~110 filas, emula Excel desktop.
Cambiar dropdown mes/año/semana en `00_Config` SOBREESCRIBE el grid (no acumula filas).
Datos reales persisten en hoja oculta `98_DB_YYYY_MM` correspondiente.

**Volumen visible**:

| Concepto | Valor | % del límite Sheets |
|---|---|---|
| Filas visibles (header + grid) | 110 | 0.22% del límite 50k |
| Columnas | 22 | — |
| Celdas totales | 2,420 | 0.02% del límite 10M |

**Volumen storage real** (en `98_DB_YYYY_MM`):

| Concepto | Valor | % del límite Sheets |
|---|---|---|
| Filas por mes (50 personas × 31 días máx) | 1,550 | 3.1% del límite 50k |
| Celdas por mes | 34,100 | 0.34% del límite 10M |
| **Total 12 meses 2026** | **18,600 filas** | **distribuidas, ninguna hoja toca límite** |

**Columnas** (idénticas a `98_DB_YYYY_MM` — mismo template):

| Col | Header | Tipo | Fórmula / Validación | Anchor? |
|---|---|---|---|---|
| A | Fecha | date | DV ≥ `B3` (año inicio) | — |
| B | ISO_Week | text | `W##` calculado | — |
| C | Dia_Semana | enum | `lun..dom` | — |
| D | Es_Feriado | bool | lookup directo `01_Feriados_PE.D:D` | — |
| E | Tipo_Dia | enum | DV: `Normal\|Sabado\|Domingo\|Feriado` | sí (CF) |
| F | ID_Persona | text | DV list_from 02_Personas.A | sí (FK) |
| G | Apellidos_Nombres | text | VLOOKUP desde 02 | — |
| H | Turno | enum | `diurno \| nocturno` | — |
| I | Hora_Ingreso | time | `HH:mm` | — |
| J | Hora_Salida | time | `HH:mm` | — |
| K | Horas_Total | number | DV: 0-24, decimal(4,2) | — |
| L | Horas_Normales | number | `IF(Tipo=Normal, MIN(K, $B$9), 0)` | — |
| M | Horas_Extras | number | `K − L` | sí (CF) |
| N | Factor_Aplicado | number | INDEX/MATCH matriz factor | — |
| O | Horas_Ponderadas | number | `K × N` | — |
| P | Pago_HE_PEN | number | `O × Tarifa_Persona` | sí (CF) |
| Q | Observaciones | text | libre | — |
| R | Aud_Usuario | text | email editor (onEdit) | — |
| S | Aud_Timestamp | datetime | ISO 8601 | — |
| T | Estado | enum | DV: `borrador\|validado\|aprobado\|pagado` | — |
| U | Mes_Año | text | `TEXT(A2, 'YYYY-MM')` (sync key) | sí (partition) |
| V | Sync_Status | enum | DV: `synced\|pending\|conflict` | sí (CF sync) |

**Convenciones de header**: columna 1-3 siempre visibles (freeze `A2:C2`). Headers `G1:V1` con rotación 90° para legibilidad vertical.

**Lifecycle** (gobernado por `08_DBSync.gs`):

```
[onEdit en 00_Config!B2/B3/B4]
   ↓
loadMonthToGrid(mes_anterior, anio)   ← save primero
   ↓
saveGridToDB(mes_anterior, anio)      ← persiste grid actual en su DB
   ↓
loadMonthToGrid(mes_nuevo, anio)      ← sobreescribe grid con datos del nuevo mes
   ↓
[grid listo para edición]
```

---

## 98_DB_YYYY_MM — Storage mensual (template replicado 12×)

**Schema idéntico a `10_Registro_HE`** (mismas 22 columnas A-V) para que `loadMonthToGrid()` y `saveGridToDB()` operen sin transformación.

**Filas esperadas por hoja** (depende de días del mes):

| Mes 2026 | Días | Filas máx (50 × días) |
|---|---|---|
| 01 Enero | 31 | 1,550 |
| 02 Febrero | 28 | 1,400 |
| 03 Marzo | 31 | 1,550 |
| 04 Abril | 30 | 1,500 |
| 05 Mayo | 31 | 1,550 |
| 06 Junio | 30 | 1,500 |
| 07 Julio | 31 | 1,550 |
| 08 Agosto | 31 | 1,550 |
| 09 Septiembre | 30 | 1,500 |
| 10 Octubre | 31 | 1,550 |
| 11 Noviembre | 30 | 1,500 |
| 12 Diciembre | 31 | 1,550 |
| **TOTAL 2026** | **365** | **18,250** |

**Visibilidad**: oculta (`hidden_from_users`). Usuario solo ve `10_Registro_HE` como grid de trabajo.

**Color tab**: `#374151` (gris oscuro — marca "storage").

**Naming convention**: `98_DB_{YYYY}_{MM}` donde `MM` es zero-padded (ej: `98_DB_2026_03`).

**Auto-creación**: via `08_DBSync.createMonthSheetIfMissing(mes, anio)` — copia headers de `10_Registro_HE` y configura DV/CF idénticos. Se invoca lazy al primer acceso.

---

## 20_Dashboard — KPIs ejecutivos

| Bloque | Rango | Contenido | Fuente |
|---|---|---|---|
| KPI cards | `A1:F4` | 4 cards: HE mes, HE acumulado YTD, Pago mes, Personas activas | `INDIRECT('98_DB_' & B3 & '_' & TEXT(B2, '00'))` |
| Resumen por persona | `A10:F40` | SUMIFS por persona desde DB activa | igual |
| Heatmap semana × persona | `H1:N40` | matrix con CF gradiente | igual |
| Gráfico HE mensual | `A20:L20` | columna agrupada, fuente DB activa | igual |
| Top 10 HE | `D20:F40` | SORTN por Horas_Ponderadas desde DB activa | igual |

> **Cambio vs v1**: Dashboard lee desde `98_DB_YYYY_MM` (no desde `10_Registro_HE`).
> Esto evita que KPIs se invaliden mientras el usuario edita el grid de trabajo.

---

## 99_Audit_Log — Solo log (sin cambios vs v1)

| Col | Header | Tipo |
|---|---|---|
| A | Timestamp | datetime |
| B | Usuario | text (email) |
| C | Acción | enum (`create\|update\|delete\|recalc\|trigger\|sync_load\|sync_save\|sync_create_sheet`) |
| D | Sheet | text |
| E | Fila | int |
| F | Columna | text |
| G | Valor_Anterior | text |
| H | Valor_Nuevo | text |

> **Acciones nuevas en C**: `sync_load` (cambio de mes → carga grid), `sync_save` (cambio de mes → guarda grid), `sync_create_sheet` (auto-creación de hoja DB).
>
> Append-only via `06_AuditLogger.gs`. Nunca se edita manualmente.
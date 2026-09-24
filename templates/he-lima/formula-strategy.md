# Formula Strategy — he-lima (v2 — design-revised)
# Arquitectura de fórmulas (NO código — Andrés implementará)
# Authority: marco · 2026-09-24
#
# ─────────────────────────────────────────────────────────────
# CAMBIO v1 → v2 (CORR-1):
#   · Detección de feriado: lookup directo contra 01_Feriados_PE.D:D
#     (sin acumulación, sin columna auxiliar)
#   · NUEVA sección 11 "Working surface sync" — estrategia load/save
#     entre 10_Registro_HE y 98_DB_YYYY_MM
#   · Anchors actualizados: B8→B9 (jornada max)
#   · Formulas operan sobre 110 filas del grid (no 18,250)
# ─────────────────────────────────────────────────────────────

> **Principio rector**: la mayor cantidad de cálculo debe ocurrir en fórmulas de Sheets
> (que el motor de Google ejecuta nativamente, no Apps Script). Apps Script solo cuando
> la fórmula es imposible o demasiado costosa.
>
> **Principio v2**: con modelo working surface, el grid es chico (110 filas), las fórmulas
> son simples y el cálculo derivado por Sheets es prácticamente gratis.

---

## 1. Clasificación del día (columnas C, D, E en `10_Registro_HE`)

| Columna | Intención | Estrategia |
|---|---|---|
| C — `Dia_Semana` | Lun-Dom textual | `WEEKDAY(A2, 2)` mapeado a texto vía `CHOOSE` o `IFS` |
| D — `Es_Feriado` | Bool | **CORR-1**: `COUNTIF(01_Feriados_PE!D:D, A2) > 0` — lookup directo |
| E — `Tipo_Dia` | `Normal\|Sabado\|Domingo\|Feriado` | Anidar: si D=true → "Feriado"; sino si WEEKDAY=6 → "Sabado"; sino si WEEKDAY=7 → "Domingo"; sino "Normal" |

**Implementación fila 2** (replicar hacia abajo hasta fila 110):

```excel
C2: =CHOOSE(WEEKDAY(A2,2),"lun","mar","mie","jue","vie","sab","dom")
D2: =COUNTIF(01_Feriados_PE!$D:$D, A2) > 0
E2: =IF(D2,"Feriado",IF(WEEKDAY(A2,2)=6,"Sabado",IF(WEEKDAY(A2,2)=7,"Domingo","Normal")))
```

**Performance** (CORR-1): sobre 110 filas del grid, este cálculo se completa en <50ms.
Sobre `98_DB_YYYY_MM` (1,550 filas por mes) sigue siendo trivial — `COUNTIF(D:D, fecha)` con
rango completo no requiere optimización porque la lista de feriados es de solo 13 valores.

**Sin columna auxiliar de feriados** (v1 la usaba para cachear, ya no es necesaria).

---

## 2. Cálculo de horas (columnas K, L, M en `10_Registro_HE`)

| Columna | Intención | Estrategia |
|---|---|---|
| K — `Horas_Total` | Decimal(4,2) | `(J2 − I2) × 24` con manejo de cruce de medianoche vía `MOD` |
| L — `Horas_Normales` | H dentro de jornada normal | Si `Tipo_Dia = "Normal"`: `MIN(K2, $B$9)`; sino `0` — **CORR-1: B9 ahora es jornada max** |
| M — `Horas_Extras` | H sobre jornada o día no-laboral | `K2 − L2` |

**Implementación fila 2**:

```excel
K2: =MOD(J2 - I2, 1) * 24                              # horas decimales con cruce medianoche
L2: =IF(E2 = "Normal", MIN(K2, $B$9), 0)               # B9 = jornada_max_diaria
M2: =K2 - L2
```

**Invariante crítica**: `L + M = K` siempre. Apps Script valida en `validarInvariantes()`.

---

## 3. Factor HE aplicado (columna N)

| Intención | Estrategia |
|---|---|
| Factor según tipo de día × turno | `INDEX(matriz_factor, MATCH(...), MATCH(...))` |
| Matriz factor | Tabla auxiliar `00_Config!D6:G9` (filas: Normal/Sab/Dom/Fer; cols: diurno/nocturno) |

**Layout matriz en `00_Config`**:

|         | diurno (B6) | nocturno (B7) |
|---------|-------------|---------------|
| Normal  | $B$6        | $B$7          |
| Sab     | $B$8        | $B$8          |
| Dom     | $B$8        | $B$8          |
| Feriado | $B$8        | $B$8          |

(B6=1.25, B7=1.35, B8=2.00 — factores configurables)

**Implementación fila 2**:

```excel
N2: =IF(H2 = "nocturno",
        INDEX($D$6:$G$9, MATCH(E2, $D$6:$D$9, 0), 3),
        INDEX($D$6:$G$9, MATCH(E2, $D$6:$D$9, 0), 2))
```

**Nota**: tipo "Normal" + horas ≤8h usa factor 1.00 (no HE), capturado en `L` ya que `MIN(K, 8)` deja M=0.
El factor se aplica SOLO a `M` (horas extra), vía `O = K * N` cuando N refleja recargo.

**Decisión abierta para Andrés**: si CEO prefiere que `N` se multiplique solo sobre M (no sobre K),
ajustar fórmula en columna `O`:

```excel
O2: =IF(E2 = "Normal", M2 * N2, K2 * N2)   # alternativa: ponderar siempre sobre M cuando hay HE
```

---

## 4. ISO Week (columna B)

| Intención | Estrategia |
|---|---|
| Semana ISO 8601 formato `W##` | `TEXT(WEEKNUM(A2, 21), "W00")` |

**Implementación fila 2**:

```excel
B2: =TEXT(WEEKNUM(A2, 21), "W00")
```

**Nota Sheets**: `WEEKNUM(_, 21)` retorna ISO week per Google Sheets semantics.
Verificar equivalencia con ISO 8601 estricto (algunos años difieren en semana 53).
La columna B del grid está alineada con `00_Config!B4` (dropdown semana activa) — usado por filtro.

---

## 5. Pago HE (columna P)

| Intención | Estrategia |
|---|---|
| `Pago_HE_PEN` | `O2 × VLOOKUP(F2, 02_Personas!A:I, 9, FALSE)` |
| `O2 — Horas_Ponderadas` | `K2 × N2` (o variante según decisión §3) |

**Implementación fila 2**:

```excel
O2: =K2 * N2
P2: =O2 * VLOOKUP(F2, 02_Personas!$A:$I, 9, FALSE)
```

---

## 6. Sync helpers (columnas U, V) — NUEVO en v2

| Columna | Intención | Estrategia |
|---|---|---|
| U — `Mes_Año` | Sync key con `98_DB_YYYY_MM` | `TEXT(A2, "YYYY-MM")` |
| V — `Sync_Status` | Estado sincronización grid↔DB | Apps Script escribe: `synced` / `pending` / `conflict` |

**Implementación fila 2**:

```excel
U2: =TEXT(A2, "YYYY-MM")
V2: "synced"   # valor inicial; 08_DBSync.gs lo actualiza
```

**Por qué importa** (CORR-1): la columna U es la **clave de sincronización** entre `10_Registro_HE`
y `98_DB_YYYY_MM`. La hoja DB destino se determina por:

```excel
# Equivalente conceptual (NO fórmula — lo ejecuta 08_DBSync.gs):
target_sheet = "98_DB_" & U2_reemplazar_guion & "_" & TEXT(B2, "00")
# ej: 98_DB_2026_03
```

Apps Script no evalúa U celda por celda: lee el mes activo de `00_Config!B2` y llama
`08_DBSync.loadMonthToGrid(mes, anio)` cuando cambia el dropdown.

---

## 7. KPIs Dashboard (sheet `20_Dashboard`)

**CORR-1**: el Dashboard lee de `98_DB_YYYY_MM` (storage mensual), NO del grid de trabajo.
Esto evita que KPIs se invaliden mientras el usuario edita en `10_Registro_HE`.

| KPI | Estrategia | Referencia storage |
|---|---|---|
| HE del mes | `SUMIFS(INDIRECT(...).M:M, INDIRECT(...).U:U, B2&"-"&B3)` | DB del mes activo |
| HE acumulado YTD | Sumar 12 hojas DB (enero → mes activo) | 98_DB_2026_01..mes activo |
| Pago del mes | `SUMIFS(INDIRECT(...).P:P, ...)` | DB del mes activo |
| Personas activas | `COUNTIF(02_Personas!H:H, TRUE)` | — |
| Heatmap semana × persona | `SUMIFS(M:M, B:B, W##, F:F, P_NNN)` cruzado con matriz | DB del mes activo |

**Implementación INDIRECT**:

```excel
# Celda auxiliar oculta en Dashboard (esconder para usuario final)
DB_REF: ="98_DB_" & $B$3 & "_" & TEXT($B$2, "00")
# ej: 98_DB_2026_03

# KPI HE del mes
HE_MES: =SUMIFS(INDIRECT(DB_REF & "!$M$2:$M$1550"),
                INDIRECT(DB_REF & "!$U$2:$U$1550"), $B$2 & "-" & $B$3)
```

**Optimización clave** (CORR-1): `INDIRECT` está acotado a 12 valores posibles (uno por mes).
Se prefija con `INDIRECT` solo en Dashboard, **nunca** en `10_Registro_HE` ni en `98_DB_YYYY_MM`
(volátil + recalcula toda la hoja).

---

## 8. Auto-pintado por tipo de día (CF — NO fórmula, pero la condición usa fórmulas)

| Rango | Condición | Referencia `architecture.yaml` |
|---|---|---|
| `E2:E110` | TextEquals Feriado | cf_feriado |
| `E2:E110` | TextEquals Sabado\|Domingo | cf_finsem |
| `M2:M110` | `> 4` | cf_he_alta |
| `P2:P110` | `> percentil_90_mes` | cf_pago_alt |
| `V2:V110` | TextEquals `pending` | cf_sync_pending (NUEVO v2) |
| `V2:V110` | TextEquals `conflict` | cf_sync_conflict (NUEVO v2) |

**Restricción**: NO usar CF con fórmula `PERCENTILE` por celda (re-cálculo masivo).
Alternativa: celda pre-calculada en Dashboard que el CF referencia (`percentil_90_mes`).

---

## 9. Funciones a evitar (anti-patterns)

| Función | Por qué evitar | Mitigación |
|---|---|---|
| `INDIRECT` masivo en grid | Volátil, recalcula toda la hoja | Solo en Dashboard, **nunca** en `10_Registro_HE` ni en `98_DB_YYYY_MM` |
| `NOW()` / `TODAY()` por fila | Volátil | Una sola celda en `00_Config!B11` |
| `OFFSET` en fórmulas de array | Inestable | Usar `INDEX` |
| `QUERY` para joins | Overhead parse | `SUMIFS` con columna U (Mes_Año) |
| `ARRAYFORMULA` en 1,550 filas | Legibilidad | Fila por fila (es barato a esta escala) |

---

## 10. Resumen de responsabilidades

| Capa | Tecnología | Volumen (CORR-1) |
|---|---|---|
| Cálculo derivado grid (K,L,M,N,O,P,U,V) | **Fórmulas Sheets** | 110 × 8 = 880 celdas |
| Clasificación día grid (C,D,E) | **Fórmulas Sheets** | 110 × 3 = 330 celdas |
| Cálculo derivado DB mensual | **Fórmulas Sheets** | 1,550 × 8 = 12,400 celdas (por mes) |
| Lookups (G, P) | **VLOOKUP Sheets** | 110 × 2 (grid) |
| Auto-pintado | **CF nativa** | 10 reglas globales |
| Sync grid ↔ DB mensual | **Apps Script** | 1 vez por cambio de dropdown |
| Auto-crear hoja DB | **Apps Script** | 1 vez lazy por mes nuevo |
| Bootstrap feriados | **Apps Script** | 1 vez/año al abrir |
| Cierre mensual + audit | **Apps Script + trigger** | 1 vez/mes |

---

## 11. Working surface sync (NUEVO — sección crítica CORR-1)

### 11.1 Trigger — cuándo se ejecuta el sync

```
[Usuario cambia dropdown en 00_Config]
  ↓
onEdit(e) detecta cambio en B2, B3 o B4
  ↓
delega a 08_DBSync.onConfigChange(e)
```

### 11.2 Flujo load/save (regla de orden)

```
PASO 1: saveGridToDB(mes_anterior, anio)    # preserva grid actual antes de cambiar
   ↓
PASO 2: loadMonthToGrid(mes_nuevo, anio)     # sobreescribe grid con datos del nuevo mes
```

**Por qué este orden**: si se cambia primero el grid sin guardar, se pierde el trabajo del
usuario del mes saliente. `save` SIEMPRE antes de `load`.

### 11.3 saveGridToDB(mes, anio) — pseudointención

```
1. target_sheet = resolveTargetSheet(mes, anio)   # "98_DB_2026_03"
2. if NOT exists(target_sheet):
       createMonthSheetIfMissing(mes, anio)         # copia header de 10_Registro_HE
3. open range 10_Registro_HE!A2:V110
4. for each non-empty row in range:
       a. compute Mes_Año = U(row)                  # "2026-03"
       b. if Mes_Año != target_sheet's month: skip  # sanity check
       c. write row to next empty row in target_sheet
5. update 10_Registro_HE!V column for synced rows → "synced"
6. log to 99_Audit_Log: accion="sync_save"
```

### 11.4 loadMonthToGrid(mes, anio) — pseudointención

```
1. target_sheet = resolveTargetSheet(mes, anio)   # "98_DB_2026_03"
2. if NOT exists(target_sheet):
       createMonthSheetIfMissing(mes, anio)         # crea vacía
       return  # grid queda vacío (caso primer acceso al mes)
3. clear 10_Registro_HE!A2:V110  (preserva header en fila 1)
4. read target_sheet!A2:V{lastRow}
5. for each row in target_sheet:
       a. if row's month == mes && anio == anio:
          write to next row in 10_Registro_HE
       b. if week filter active (00_Config!B4 != all):
          skip rows where WEEKNUM(row.A) != selected week
6. set 10_Registro_HE!V for loaded rows → "synced"
7. log to 99_Audit_Log: accion="sync_load"
```

### 11.5 createMonthSheetIfMissing(mes, anio) — pseudointención

```
1. target_name = "98_DB_" + anio + "_" + TEXT(mes, "00")
2. if spreadsheet has sheet named target_name: return
3. new_sheet = spreadsheet.insertSheet(target_name)
4. hide new_sheet via setHidden(true)
5. set tab color to #374151
6. copy values from 10_Registro_HE!A1:V1 → new_sheet!A1:V1 (headers)
7. apply same CF rules as 10_Registro_HE
8. apply same DV rules as 10_Registro_HE
9. log to 99_Audit_Log: accion="sync_create_sheet"
```

### 11.6 Invariantes de sync (NO negociar)

| Invariante | Razón | Si se viola |
|---|---|---|
| `save` antes de `load` | Evitar pérdida de datos del usuario | Alert + bloquear carga |
| Solo filas del mes activo se sincronizan | Coherencia Mes_Año | Descartar con log warning |
| Header siempre preservado en grid y DB | Estructura schema | Alert + rollback |
| `Sync_Status = synced` solo cuando DB confirma escritura | Honestidad estado | Marcar `pending` |
| Cambio de mes durante edición activa | Race condition | Bloquear con `LockService` |

### 11.7 Edge cases a validar en P5 (hardening)

| Escenario | Comportamiento esperado |
|---|---|
| Usuario cambia mes con grid vacío | `save` no hace nada; `load` carga DB o crea vacía |
| Usuario cambia mes, DB no existe | `createMonthSheetIfMissing` crea y grid queda vacío |
| Dos usuarios editan mismo mes | `LockService.getDocumentLock()` — segundo espera |
| DB tiene más filas que capacidad grid (caso borde) | Truncar a 110 + warning en `Sync_Status` |
| Cambio de año (2026 → 2027) | Resolver `98_DB_2027_01..12` (nuevo set, lazy) |
| Cambio de semana dentro del mismo mes | NO requiere sync — solo filtro local en load |

---

## 12. Cambios de anchors (resumen rápido para Andrés)

| Concepto | v1 anchor | v2 anchor | Notas |
|---|---|---|---|
| Mes activo | `00_Config!B2` | `00_Config!B2` | sin cambio |
| Año activo | `00_Config!B3` | `00_Config!B3` | sin cambio |
| Semana activa | (no existía) | `00_Config!B4` | **NUEVO** |
| Persona filtro | `00_Config!B4` | `00_Config!B5` | movido +1 |
| Factor HE diurno | `00_Config!B5` | `00_Config!B6` | movido +1 |
| Factor HE nocturno | `00_Config!B6` | `00_Config!B7` | movido +1 |
| Factor feriado | `00_Config!B7` | `00_Config!B8` | movido +1 |
| Jornada max | `00_Config!B8` | `00_Config!B9` | movido +1 |
| Timestamp refresh | `00_Config!B10` | `00_Config!B11` | movido +1 |

**Impacto en fórmulas**:
- Toda referencia `$B$8` (jornada) → `$B$9`
- `01_CfgLoader.gs` debe actualizar su mapa de anchors (P0 deliverable)
- `04_FormulasRecalc.gs` debe usar `B2/B3/B4/B5` (no B2/B3/B4 anterior)

---

*Marco · 2026-09-24 · v2 design-revised · awaiting Andrés implementation in P3-P5*
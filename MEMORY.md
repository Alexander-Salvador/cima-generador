# CIMA — Memoria del proyecto

Memoria persistente. Última actualización: **2026-10-06 — Hardening 2026 aplicado (potenciación Sheets v2)**.

---

## Identidad

- **Proyecto:** CIMA — Call center Intelligent Management & Analytics
- **Operación:** Call center técnico campaña MOVIL (operadora móvil)
- **Marca interna:** Grupo WhatsApp "CIMA"
- **Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`

---

## Estructura actual — tabs y columnas

### PERSONAL (950 rows × 9 cols)
```
N° | DNI | NOMBRES_COMPLETOS | TELEFONO | HORARIO | ESTADO | CARGO | ANYDESK | OBSERVACIONES
```
PK = N°. Agentes activos = 25. Estados: Activo, Cese, CAMBIO_SUPERVISOR.

### FEEDBACK (982 rows × 8 cols) ⭐
```
N° | AGENTE | TIPO | N°_DETALLES | SUBTIPO | FECHA | OBSERVACIÓN | ESTADO
```
**Orden CEO (AGENTE en col B).** N°_DETALLES es COUNTIF fórmula, fuente de verdad.

| Col | Contenido |
|-----|-----------|
| A | N° (PK, seq 1-9) |
| B | AGENTE = "ID - NOMBRE_COMPLETO" (ej: "3 - RAMIREZ HUAMAN JOHNN MILTON") |
| C | TIPO (catálogo: 8 valores) |
| D | N°_DETALLES = COUNTIF('FEEDBACK-DETALLE_LLAMADA'!A:A, A2) — fórmula |
| E | SUBTIPO (catálogo: 24 valores) |
| F | FECHA (dd/mm/yyyy) |
| G | OBSERVACIÓN (texto libre) |
| H | ESTADO (Cumplido / Feedback dado / Pendiente) |

**Data validations activos:**
- B (AGENTE): dropdown valores "ID - NOMBRE_COMPLETO" para los 25 agentes
- C (TIPO): 8 valores del catálogo CONFIG
- E (SUBTIPO): 24 valores del catálogo CONFIG

### FEEDBACK-DETALLE_LLAMADA (101 rows × 5 cols) ⭐
```
N°_FEEDBACK | N°_SECUENCIA | FECHA_LLAMADA | TIPO_INCIDENTE | OBSERVACIÓN
```
FK: N°_FEEDBACK → FEEDBACK.N°. AGENTE se resuelve vía FK → FEEDBACK → PERSONAL.

| Col | Contenido |
|-----|-----------|
| A | N°_FEEDBACK (FK → FEEDBACK.N°) |
| B | N°_SECUENCIA (seq 1-N por grupo de N°_FEEDBACK) |
| C | FECHA_LLAMADA (dd/mm/yyyy, o dd/mm/yyyy HH:MM:SS) |
| D | TIPO_INCIDENTE (catálogo: 24 valores) |
| E | OBSERVACIÓN (texto libre) |

**Data validations activos:**
- D (TIPO_INCIDENTE): dropdown 24 valores del catálogo CONFIG

**⚠️ Nota técnica:** Row 1 de este tab tenía merge. Para reescribir header, primero ejecutar:
```bash
go run ./cmd/sheets/ unmerge-cells --tab "FEEDBACK-DETALLE_LLAMADA" --range "A1:E1"
```

### HORAS EXTRAS (927 rows × 47 cols)
HE semanales L-V con rollover. 7 cols configurables. Rows 13-30+ por agente.

### CONFIG (1000 rows × 11 cols)
Catálogos de soporte. No modificar manualmente.

---

## Catálogos — CONFIG

### TIPO_FEEDBACK (CONFIG rows ~106-113)
```
KPI_AUDIT / ATTITUDINAL / CALIBRACION / ESCUCHA_VIVO /
GRABACION / FEEDBACK_POSITIVO / CARTA_COMPROMISO / REMOCION
```

### SUBTIPO_FEEDBACK (CONFIG rows ~117-140)
```
KPI_AUDIT:        FCR_BAJO, AHT_ALTO, QA_SCORE_BAJO, FIRST_CALL_NOT_RESOLVED
ATTITUDINAL:      FALTA_INJUSTIFICADA, TARDANZA, ABANDONO, ESTADOS_NO_AUTORIZADOS,
                  USO_INDEBIDO_HERRAMIENTAS, INCUMPLIMIENTO_PROTOCOLO, NO_DEVUELVE_LLAMADA
CALIBRACION:      CALIBRACION_OK, CALIBRACION_NOK
ESCUCHA_VIVO:     ESCUCHA_OK, ESCUCHA_NOK
GRABACION:        GRABACION_OK, GRABACION_NOK
FEEDBACK_POSITIVO: RECONOCIMIENTO
CARTA_COMPROMISO: COMPROMISO_FIRMANTE, SEGUNDA_CARTA, TERCERA_CARTA, REINGRESO_COMPROMETIDO
REMOCION:         DESTITUCION, ABANDONO_VOLUNTARIO
```

---

## Decisiones arquitectónicas (ADRs)

| ADR | Título | Status |
|-----|--------|--------|
| ADR-001 | Sheets API > Apps Script | Accepted |
| ADR-002 | Migración OVAV → CIMA standalone | Accepted |
| ADR-003 | CITRIX separado de spreadsheet principal | Accepted |
| ADR-004 | FEEDBACK schema con taxonomía TIPO/SUBTIPO | Accepted |
| ADR-005 | FEEDBACK-DETALLE relación 1:N por N°_FEEDBACK | Accepted |

---

## Arquitectura de correlación inteligente (v2 — 2026-10-06)

### Cadena PK-FK completa

```
PERSONAL.N° (PK)  ──┐
                     │  "3 - RAMIREZ HUAMAN JOHNN MILTON"
                     ▼
                  FEEDBACK.AGENTE (col B)  ◄── dropdown dropdown valor "N° - NOMBRE"
                     │
                     │  FK
                     ▼
              FEEDBACK.N° (PK)  ──┐
                                   │
                                   ▼
                  FEEDBACK-DETALLE_LLAMADA.N°_FEEDBACK (FK)
                                   │
                                   ▼
                          N°_SECUENCIA (1..N por grupo)
```

### Catálogo en CONFIG (single source of truth)

```
CONFIG!B108:B115 ──► 8 TIPO_FEEDBACK
  KPI_AUDIT | ATTITUDINAL | CALIBRACION | ESCUCHA_VIVO | GRABACION |
  FEEDBACK_POSITIVO | CARTA_COMPROMISO | REMOCION

CONFIG!B119:D142 ──► 24 SUBTIPO_FEEDBACK + TIPO_PADRE (FK)
  Cada subtipo conoce su tipo padre → coherencia automática

CONFIG!B5:C29 ──► Mapping table (FK_FEEDBACK ↔ PERSONAL_N ↔ NOMBRE)
  Permite VLOOKUP desde FEEDBACK hacia PERSONAL sin hardcodear
```

### Reglas de coherencia (enforcement)

| Regla | Origen | Validación |
|---|---|---|
| FEEDBACK.AGENTE ∈ PERSONAL | PERSONAL!A4:A26 | Dropdown (manual) + VLOOKUP formula |
| FEEDBACK.TIPO ∈ CONFIG catálogos | CONFIG!B108:B115 | Dropdown ONE_OF_LIST |
| FEEDBACK.SUBTIPO ∈ CONFIG catálogos | CONFIG!B119:B142 | Dropdown ONE_OF_LIST |
| FEEDBACK.SUBTIPO.TIPO_PADRE = FEEDBACK.TIPO | CONFIG!C119:C142 | Cross-validation via VLOOKUP |
| PERSONAL.ESTADO ∈ {Activo, Cese, CAMBIO_SUPERVISOR} | Dropdown | Dropdown ONE_OF_LIST |
| FEEDBACK.ESTADO ∈ {Cumplido, Feedback dado, Pendiente} | Dropdown | Dropdown ONE_OF_LIST |
| FEEDBACK-DETALLE_LLAMADA.N°_FEEDBACK ∈ FEEDBACK.N° | FK constraint | VLOOKUP manual check |

### Named ranges semánticos (hoja↔referencia)

| Nombre | Rango | Uso en fórmula |
|---|---|---|
| `cfg_tipo_feedback` | `CONFIG!B108:B115` | Data validation FEEDBACK!C |
| `cfg_subtipo_feedback` | `CONFIG!B119:B142` | Data validation FEEDBACK!E, FEEDBACK-DETALLE_LLAMADA!D |
| `cfg_tipo_padre_map` | `CONFIG!B119:D142` | VLOOKUP SUBTIPO→TIPO_PADRE |
| `personal_activos` | `PERSONAL!A4:F26` | 23 agentes activos en una referencia |
| `feedback_historico` | `FEEDBACK!A2:H982` | Histórico completo para KPIs |
| `feedback_detalle_historico` | `'FEEDBACK-DETALLE_LLAMADA'!A2:E101` | Detalle para COUNTIF |

**Beneficio**: si movemos CONFIG!B108 a B109, las fórmulas `cfg_tipo_feedback` siguen funcionando. Sin búsqueda/reemplazo masivo.

---

## Estructura de potenciación 2026 (Hardening v2 aplicado)

### Data Validation — Dropdowns activos (5 reglas)

| Rango | Tipo | Lista |
|---|---|---|
| `PERSONAL!F4:F26` | ONE_OF_LIST | Activo, Cese, CAMBIO_SUPERVISOR |
| `FEEDBACK!C2:C982` | ONE_OF_LIST | 8 TIPO_FEEDBACK (dropdown literal) |
| `FEEDBACK!E2:E982` | ONE_OF_LIST | 24 SUBTIPO_FEEDBACK (dropdown literal) |
| `FEEDBACK!H2:H982` | ONE_OF_LIST | Cumplido, Feedback dado, Pendiente |
| `'FEEDBACK-DETALLE_LLAMADA'!D2:D101` | ONE_OF_LIST | 24 SUBTIPO_FEEDBACK |

### Protected Ranges — Invariantes protegidos (5 reglas)

| Rango | Tipo | Razón |
|---|---|---|
| `PERSONAL!A3:A26` | warning-only | PK agente — no modificar |
| `FEEDBACK!A1:A982` | warning-only | PK feedback — no modificar |
| `'FEEDBACK-DETALLE_LLAMADA'!A1:B101` | warning-only | FK + secuencia — auto-computed |
| `CONFIG!A1:F145` | hard | CONFIG: catálogos sistema — NO modificar manualmente |
| `PERSONAL!A1:I2` | warning-only | Header/title row |

### Conditional Formatting — Alertas visuales (6 reglas)

| Rango | Condición | Color | Significado |
|---|---|---|---|
| `FEEDBACK!H2:H982` | TEXT_EQ Pendiente | #FFE599 amarillo | Acción requerida |
| `FEEDBACK!H2:H982` | TEXT_EQ Cumplido | #D9EAD3 verde claro | Cerrado OK |
| `FEEDBACK!C2:C982` | TEXT_EQ REMOCION | #F4CCCC rojo bold | Crítico |
| `FEEDBACK!C2:C982` | TEXT_EQ CARTA_COMPROMISO | #FFE599 amarillo | Alerta |
| `PERSONAL!F4:F26` | TEXT_EQ Cese | #D9D9D9 gris | Inactivo |
| `PERSONAL!F4:F26` | TEXT_EQ CAMBIO_SUPERVISOR | #FFE599 amarillo | Seguimiento |

### Freeze Panes — Contexto siempre visible

| Tab | Filas | Cols | Razón |
|---|---|---|---|
| PERSONAL | 3 | 1 | Header (row 3) + PK (col A) |
| FEEDBACK | 1 | 1 | Header (row 1) + PK (col A) |
| FEEDBACK-DETALLE_LLAMADA | 1 | 0 | Header (row 1, contiene merge) |
| HORAS EXTRAS | 3 | 0 | Summary + headers (row 3, contiene merge) |
| CONFIG | 5 | 0 | Title + headers + mapping |

---

## Decisiones arquitectónicas (ADRs)

| ADR | Título | Status |
|-----|--------|--------|
| ADR-001 | Sheets API > Apps Script | Accepted |
| ADR-002 | Migración OVAV → CIMA standalone | Accepted |
| ADR-003 | CITRIX separado de spreadsheet principal | Accepted |
| ADR-004 | FEEDBACK schema con taxonomía TIPO/SUBTIPO | Accepted |
| ADR-005 | FEEDBACK-DETALLE relación 1:N por N°_FEEDBACK | Accepted |
| ADR-006 | Hardening Sheets 2026 (v2 potenciación) | Accepted |

### ADR-006 — Hardening Sheets 2026 (v2 potenciación)

**Contexto:** Plan operativo call center con errores humanos recurrentes en FEEDBACK/HE/PERSONAL — valores fuera de catálogo, fórmulas rotas, historial modificado por accidente.

**Decisión:** Aplicar 6 capas de potenciación nativa sin código:
1. **Named ranges semánticos** → fórmulas legibles, migración segura
2. **Data validation dropdowns** → rechaza input fuera de catálogo
3. **Protected ranges** → bloquea borrado accidental de PKs/CONFIG
4. **Conditional formatting** → alertas visuales (Pendiente/REMOCION)
5. **Freeze panes** → contexto siempre visible
6. **Catálogos CONFIG** → single source of truth (TIPO, SUBTIPO)

**Consecuencias:**
- 0 código Apps Script nuevo (cumple regla CIMA v0.5.0)
- Cambios reversibles via Version History
- Coherencia automática TIPO↔SUBTIPO via dropdowns
- Inspector visual detecta "Pendiente" sin leer celda por celda

---

## Lecciones aprendidas

### L-001: merge en row 1 rompe writes
Si al escribir header el write retorna ✅ pero la celda no persiste → ejecutar `unmerge-cells` primero.

### L-002: JSON escaping de comillas simples en fórmulas
COUNTIF con nombre de sheet entre comillas simples requiere `\u0027` en JSON:
```
=COUNTIF(\u0027FEEDBACK-DETALLE_LLAMADA\u0027!A:A, A2)
```

### L-003: Verificar siempre estructura antes de refactor
Usar `go run ./cmd/sheets/ list` + read de headers antes de cualquier cambio.

### L-004: Hardening v2 — todas las herramientas son nativas (sin Apps Script)
Aplicado 2026-10-06: 6 named ranges + 5 dropdowns + 5 protected + 6 conditional formats + 5 freeze. Total ~27 operaciones batch update. Latencia < 1s mantenida. Sin dependencia externa.

### L-005: Coordenadas exactas de catálogos CONFIG
- TIPO_FEEDBACK: `CONFIG!B108:B115` (8 tipos)
- SUBTIPO_FEEDBACK: `CONFIG!B119:B142` (24 subtipos)
- TIPO_PADRE map: `CONFIG!B119:D142`
- Mapping table: `CONFIG!B5:C29` (FK_FEEDBACK ↔ PERSONAL_N)

---

## Git — estado actual

```
develop (HEAD) — 3 commits ahead of origin/develop, push bloqueado por gobernanza
```

**Commits pendientes de push:**
```
6a8b3aa docs(cima): FEEDBACK — CEO column order (N°_DETALLES after TIPO), unmerge header row note
77fd2ca docs(cima): FEEDBACK — full agent names, remove CANT_CALLS, AGENTE to col B per CEO
56f11f5 docs(cima): remove redundant columns — FEEDBACK (-SUPERVISOR), FEEDBACK-DETALLE (-AGENTE,-TELEFONO,-FEEDBACK_DADO)
```

**Para hacer push:**
```bash
git push origin develop
# Requiere aprobación CEO
```

---

## Stack

- **Sheets bridge:** `go-runtime/cmd/sheets/` — CLI: list, read, write, validate, unmerge-cells
- **Auth:** OAuth con token en vault cifrado (AES-256-GCM)
- **Spreadsheet:** Google Sheets API v4

### Comandos útiles
```bash
cd go-runtime

# Listar tabs
go run ./cmd/sheets/ list

# Leer rango
go run ./cmd/sheets/ read --range "FEEDBACK!A1:H10"

# Escribir (values-json en JSON plano, comillas simples como \u0027)
go run ./cmd/sheets/ write --spreadsheet "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y" \
  --range "FEEDBACK!H2:H10" \
  --values-json '[["=COUNTIF(\u0027FEEDBACK-DETALLE_LLAMADA\u0027!A:A,A2)"]]'

# Validación dropdown
go run ./cmd/sheets/ validate --spreadsheet "..." --range "FEEDBACK!C2:C100" \
  --type "ONE_OF_LIST" --values "KPI_AUDIT,ATTITUDINAL,..."

# Desmergear antes de escribir header
go run ./cmd/sheets/ unmerge-cells --tab "FEEDBACK-DETALLE_LLAMADA" --range "A1:E1"
```

---

*Última actualización: 2026-10-06 — Thavren*

# CIMA — Memoria del proyecto

Memoria persistente. Última actualización: **2026-10-06**.

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

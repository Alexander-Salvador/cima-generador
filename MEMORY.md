# CIMA — Memoria del proyecto

Memoria persistente del sistema CIMA. Esta es la **memoria del proyecto independiente**, separada de la memoria OVAV global.

---

## Identidad del proyecto

- **Nombre:** CIMA
- **Significado:** **C**all center **I**ntelligent **M**anagement & **A**nalytics
- **Operación:** Call center técnico campaña MOVIL (operadora móvil)
- **Identidad WhatsApp:** El grupo de WhatsApp del equipo se llama **"CIMA"** — esa es la marca interna del equipo técnico
- **Spreadsheet ID:** `1CPHU_PLACEHOLDER` (CITRIX RRHH: `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`)

---

## Sistema de Sheets — flujo por agente

```
                 SUPERVISOR
                     │
                     ▼
            ┌──────────────────┐
            │   HE (L-V)       │  Horas extras semanales, rollover automático
            │   7 cols         │  rows 13-30+
            └──────────────────┘
                     │
                     ▼
            ┌──────────────────┐
            │   TE (FT only)   │  Turno especial: 6h/day o TE
            │   semanal        │  solo agentes FT
            └──────────────────┘
                     │
                     ▼
            ┌──────────────────┐
            │   FEEDBACK       │  Master-detail: AGENTE → OBSERVACIÓN → ESTADO
            │   8 cols         │  con HYPERLINKs + workflow states
            └──────────────────┘
                     │
                     ▼
            ┌──────────────────┐
            │   CONFIG         │  catálogos: TIPO_TURNO, ESTADOS, SUPERVISOR
            └──────────────────┘

CITRIX (RRHH, separado):
            ┌──────────────────┐
            │ CITRIX_UPDATE_   │  Control horario, validación fechas
            │ LOG_RRHH.xlsx    │  D4:D100 date picker, formulas HH:MM
            └──────────────────┘
```

---

## Decisiones arquitectónicas (ADRs)

#### ADR-001: SHEETS_API > APPS_SCRIPT (2026-09-26)
- **Status:** Accepted
- **Contexto:** Apps Script requería reload manual, no se podía automatizar deploy, rompía en cambios de scope OAuth.
- **Decisión:** Toda la lógica de Sheets pasa por Google Sheets API via `ovav-sheets` Go bridge. Apps Script queda solo para cosas legacy.
- **Consecuencia:** Más control, CI/CD posible, ~50% menos código.

#### ADR-002: Migración OVAV → CIMA standalone (2026-10-05)
- **Status:** Accepted
- **Contexto:** El sheets-bridge y templates CIMA contaminaban OVAV develop (~110 commits ahead, +49,890 líneas, 370 archivos).
- **Decisión:** CIMA se separa a repo independiente (`/home/braka/Systems/projects/work/cima/`). OVAV root queda limpio.
- **Consecuencia:** OVAV sigue siendo la plataforma. CIMA es un consumidor de OVAV runtime + Sheets API. Historia preservada via format-patch + git am.

#### ADR-003: CITRIX separado de la CIMA spreadsheet principal (v0.5.6)
- **Status:** Accepted
- **Contexto:** RRHH necesita control horario que NO debe mezclarse con la operación de call center.
- **Decisión:** CITRIX_UPDATE_LOG_RRHH.xlsx es archivo aparte. CIMA sheet principal solo tiene HE/TE/FEEDBACK/CONFIG.

#### ADR-004: FEEDBACK schema 8 cols (sin ID, eval, último_tipo, compromiso, cumplido, riesgo)
- **Status:** Accepted
- **Contexto:** Schema original tenía 14 cols, demasiado complejo para mobile del supervisor.
- **Decisión:** 8 cols finales: AGENTE | CANT_CALLS | FECHA_FEEDBACK | EVALUACION | OBSERVACIÓN | ESTADO | SUPERVISOR.

---

## Reglas operacionales

### R-001: Latencia < 1s para reads de supervisor
- **Origen:** Sesión 2026-09-26 — supervisor consultando HE a las 9am, no puede esperar.
- **Acción:** Cache de reads en Sheets Bridge, formulas pre-calculadas en cells.

### R-002: Histórico preservado
- **Origen:** Auditoría RRHH 2026-09-15.
- **Acción:** Cada cambio a un agente (HE/TE/FEEDBACK) tiene `modified_at` y `modified_by`. NO se borra histórico, solo se marca como archivado.

### R-003: Apps Script solo para legacy v0.4.x
- **Origen:** ADR-001.
- **Acción:** No agregar Apps Script features nuevas. Migrar a Sheets API si se pide una nueva feature.

### R-004: CITRIX es RRHH, no Sup. operación
- **Acción:** Si supervisor pide cambios en CITRIX, redirigir a RRHH.

---

## Hitos del proyecto

| Fecha | Hito | Notas |
|---|---|---|
| 2026-07-19 | CIMA v0.1 — first Sheets schema | Spreadsheet inicial 14 cols |
| 2026-08-15 | CIMA v0.3 — Apps Script Bridge | Bridge + onOpen menu |
| 2026-09-15 | CIMA v0.4 — Feedbacks tab | Master-detail FK |
| 2026-09-21 | CIMA v0.5 — FEEDBACK schema 8 cols | Simplificación mobile |
| 2026-09-24 | CIMA v0.5.3 — HE/TE weekly tracker | Year/month/week tracking |
| 2026-09-26 | CIMA v0.5.5 — Apps Script eliminated | Sheets API total |
| 2026-10-04 | CIMA v0.5.6 — CITRIX_RRHH separado | v0.5.6 latest |
| 2026-10-05 | **Migración OVAV → CIMA standalone** | Repo independiente |

---

## Lessons learned (CIMA-specific)

### L-001: updateCells sin coords borra toda la hoja
- **Origen:** Sesión 2026-10-04, FEEDBACK rebuild.
- **Síntoma:** Llamar `sheets.UpdateCells(...)` sin especificar range borra TODO el contenido.
- **Fix:** Siempre pasar range explícito: `UpdateCells(spreadsheetID, "FEEDBACK!A1:G100", values)`.

### L-002: Raw strings con `"` no escapan (literal)
- **Origen:** Sesión 2026-10-04, CITRIX formula insert.
- **Síntoma:** Strings con `"` se insertan literalmente sin escapar.
- **Fix:** Usar `strconv.Quote()` o template rendering con escape.

### L-003: No asumir estructura sin verificar
- **Origen:** Sesión 2026-10-04, FEEDBACK layout.
- **Síntoma:** Columnas en posiciones incorrectas después de refactor.
- **Fix:** Antes de cada refactor, hacer `get_columns_schema(spreadsheetID, sheetName)` y validar.

### L-004: Apps Script reload manual bloquea deploys
- **Origen:** Sesión 2026-09-26, deploy HE v0.2.
- **Síntoma:** Cambios en Apps Script requieren reload manual del spreadsheet.
- **Fix:** Migrar a Sheets API (ADR-001).

---

## Stack técnico (heredado de OVAV)

CIMA consume:
- **OVAV Sheets Bridge:** `go-runtime/cmd/sheets/` (en CIMA: `go-runtime/cmd/sheets/`)
- **OVAV Vault:** AES-256-CBC + PBKDF2 200k iter, key en `~/.config/ovav/vault.key`
- **OVAV Worktree System (OWS):** para features nuevas
- **OVAV Validators:** F0-F5 antes de commit
- **OVAV AGENTS.md pattern:** este archivo

CIMA **NO** consume:
- OVAV cpanel (UI)
- OVAV OpenCode harness (es solo platform)
- OVAV governors de otros proyectos

---

*Last updated: 2026-10-05 — Thavren migration*
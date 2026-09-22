# CIMA Stack — Análisis completo

> **Estado 2026-09-21:** Apps Script **vaciado y eliminado** desde v2.0. La lógica Apps Script original (`onEdit`, `dailyClose`, `validateOperation_`, `installCima`, etc.) fue **migrada a fórmulas Sheets** y dropdowns nativos. Bridge CLI es el único orquestador.

---

## Identidad

| Atributo | Valor |
|---|---|
| **Stack** | CIMA — Técnico Móvil |
| **Spreadsheet** | `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y` (URL: https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y/edit) |
| **Owner / Workspace account** | `hello@ovav.dev` (Alexander Salvador) |
| **Created** | 2026-09-18T06:11:35Z |
| **Last modified** | 2026-09-21 (stable) |
| **Apps Script** | **Removido en v2.0 (2026-09-21)**. Id del proyecto huérfano: `1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzra6n3D_6XRvMOWda4HQpg` (sigue bound a la hoja pero sin código ni webapp) |

---

## Tabs activos (3)

| Tab | sheetId | Función |
|---|---|---|
| PERSONAL | 1395452450 | Base Maestra (25 agentes + buffer 23 filas) |
| OPERACIONES | 872990953 | Workflow log (requests — pendiente auditar) |
| CONFIG | 1497962081 | Parámetros + catálogo turnos T01-T07 |

> Detalle completo del schema PERSONAL en `feat-cima-fixes` worktree's `docs/cima-stack/MEMORY.md` (v2.4 consolidado).

---

## Arquitectura lógica (post-v2.0)

```
┌──────────────────────────────────────────────────────────────────┐
│ Sheets (3 tabs) — Single Source of Truth                         │
│                                                                  │
│   PERSONAL   → maestro agentes (17 cols, fórmulas reactivas)     │
│   OPERACIONES → log único histórico (pendiente limpieza)          │
│   CONFIG     → reglas + 7 turnos + políticas                     │
└────────────────────────┬─────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────────┐
│ OVAV Sheets Bridge (Go binary: /tmp/ovav-sheets)                 │
│                                                                  │
│   Reads:   list, read, table                                    │
│   Writes:  write, append, update, create-tab, delete-tab         │
│   Format:  format, validate, protect, name-range, chart, image   │
│   Excel:   xlsx-in, xlsx-out                                     │
│   Safety:  snapshots, rollback, audit JSONL                       │
└──────────────────────────────────────────────────────────────────┘
```

> **No Apps Script runtime.** No webapp. No triggers. No `:run` calls. Sin código en el editor de la hoja — todo es Sheets formulas + OVAV bridge CLI.

---

## Migración Apps Script → fórmulas (v2.0, 2026-09-21)

| Función Apps Script original | Reemplazo actual |
|---|---|
| `onEdit(e)` autocompleta TURNO_ID, Segmento, Completitud, Estado CIMA | Fórmulas en PERSONAL: `TURNO_ID` = INDEX/MATCH(CONFIG) |
| `validateOperation_()` valida TIPO, DNI, fecha, OT | Sheets Data Validation (dropdowns en TIPO, ESTADO, PRE_POST) |
| `createOperation()` crea fila con LockService | CEO escribe filas directo en OPERACIONES |
| `changeOperationStatus()` cambia STATUS + GTR_EXPORT | CEO edita celda ESTADO directo |
| `dailyClose()` auto-rechaza PENDING cada 15 min | Decisión: 0 — CEO no quería saturar con código |
| `refreshDashboard()` recalcula KPIs en CIMA tab | (No aplica — tab CIMA fue eliminado en refactor) |
| `installCima()` bootstrap manual | **No aplica** — Apps Script eliminado |
| `exportApprovedGtr()` CSV a Drive | (Pospuesto — no requerido actualmente) |

---

## Hallazgos y discrepancias (post v2.0)

### Resueltos
- ~~Apps Script responde con HTML 404 al `projects.list`~~ → Irrelevante (Apps Script vacío)
- ~~HTML files `Index.html` / `Admin.html` no aparecen en `getContent`~~ → Irrelevante
- ~~Timezone mismatch Bogota vs Lima~~ → Resuelto (Apps Script en Lima config, proyecto eliminado)
- ~~Webapp access ANYONE_ANONYMOUS~~ → Resuelto (sin webapp)
- ~~`webapp.executeAs = USER_DEPLOYING`~~ → Resuelto (sin webapp)

### Activos
| # | Item | Estado | Acción |
|---|---|---|---|
| 1 | Apps Script project bound pero vacío | Huérfano | Eliminar del editor cuando se pueda (manual 1-click — bridge no puede borrar bound script) |
| 2 | OPERACIONES tab sin trabajar | Pendiente | Auditar y limpiar (CEO decidirá timing) |
| 3 | CONFIG tab sin auditar | Pendiente | Auditar catálogo T01-T07 + 4 secciones |

---

## Comandos útiles

| Acción | Comando |
|---|---|
| Verificar conexión | `/tmp/ovav-sheets list` |
| Health check | `/tmp/ovav-sheets doctor` |
| Vault status | `/tmp/ovav-sheets vault` |
| Leer PERSONAL | `/tmp/ovav-sheets table --tab PERSONAL` |
| Leer OPERACIONES | `/tmp/ovav-sheets table --tab OPERACIONES` |
| Leer CONFIG | `/tmp/ovav-sheets read --range "CONFIG!A1:F30"` |
| Rebuild del binary | `cd go-runtime && go build -o /tmp/ovav-sheets ./cmd/sheets/` |

---

## Worktrees activos (estado 2026-09-21)

| Worktree | Branch | HEAD | Propósito |
|---|---|---|---|
| `/home/braka/Systems/ovav/.ovav/worktrees/feat-cima-fixes` | `feat/cima-fixes` | `e379f8b` | CIMA work + handoff MEMORY |
| `/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing` | `feature/c1ma-desing` | `2294bcc` | Vault hardening + doctor fix |

> **Por defecto**, trabajo CIMA en `feat-cima-fixes`. Bridge enhancements van en `feature/c1ma-desing` (este).

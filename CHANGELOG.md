# Changelog

All notable changes to **CIMA** (Call center Intelligent Management & Analytics) are documented here.

Format: [Semantic Versioning](https://semver.org/)

---

## [1.1.0] — 2026-11-07 — **Popup v3.10.82 publicado** (etiquetas reconciliadas + gate anti-regresión)

### 🎯 Headline

Seis releases del popup (**v3.10.77 → v3.10.82**) cerradas, auditadas contra el
Excel fuente y **publicadas**: `develop` `90860bd→c163567` (fast-forward) y
`main`/Pages `be99090→cda031b` (subtree split + `--force-with-lease`).
El detalle por release vive en `apps/cima-generador/CHANGELOG.md`.

### ✨ Added

- **39 etiquetas de transferencia** reconciliadas contra el XLSX oficial
  (`Recursos CIMA/Etiquetas de Transferencia/Etiquetas de Transferencia.xlsx`),
  con los splits que la propia matriz operativa del Excel ya traía separados.
- **`ETIQUETAS_BASELINE`** en `test_popup.py`: cada opción queda anclada a
  `(cat, solicitud) → (skill 2 pasos, skill directa) + fila de origen en el
  Excel`. Editar un skill o un nombre **rompe el gate** → hay que re-verificar.
- **Búsqueda por especificidad** (title-first): título > alias > etiqueta
  operativa > detalle > skill > id, con tolerancia singular/plural
  (`FACTURAS` → `FACTURACIÓN`).
- **`etq`** (etiqueta operativa ROBO/COMERCIAL/RETENCIONES/VENTAS/ROAMING/TECNICO)
  y **`alias[]`** por opción — base para el filtro por etiqueta operativa.

### 🐛 Fixed

- **3 skills corregidos** contra el Excel: `PLAN MENOR` (→ `SK_CV_DOWNSELL_MOVIL_TX`),
  `VENTAS` (→ `SK_CV_CROSS_MOVIL_FIJA_TX`), `TÉCNICO N2` (`KONECTA` → `SK_CV_TECNICO_MOVIL_TX`).
- `PLAN MENOR` y `MIGRACIÓN DOWN` eran el mismo escenario con skills distintos → fusionados.
- `TRASLADO / MUDANZA` faltaba en M.TOTAL (gap real del Excel) → agregado.
- Contador `9/9` del input TELÉFONO desalineado → overlay CSS Grid.
- Estado "trancado": al pulsar un chip o un botón de acción quedaban el input
  marcado y el dropdown abierto → ahora se limpian.
- `AVERÍA HOGAR` había perdido una clause de validación que el Excel sí tiene.

### ✏️ Changed

- Chips de categoría → **MOVIL · M.TOTAL · STANDALONE** (`data-cat` intactos).
- Botones de TIPIFICACIÓN → **RESET** (texto Tripartita + modo avión) y
  **APN→JIRA** (plantilla `CREACION DE TK JIRA` literal del Excel).
  **CASO eliminado**.
- `PENDIENTES.md`, `MEMORY.md` (checkpoint + L-091/092/093), `AGENTS.md`, `README.md`
  y `docs/handoff/2026-11-07-v3.10.82.md` actualizados al estado vigente.

### ⏳ Pending (decisiones del CEO)

`BAJA NO REGISTRADA` → ¿en MÓVIL? · `AVERÍA HOGAR` MT → `TEFC_PRIORITY_TX` vs
`TECNICO_FIJA_TX` · 5 propuestas de mejora de botones (A filtro por etiqueta
operativa · B teclado · C badge de skill · D contador por chip · E JIRA
auto-rellenado). Detalle en `PENDIENTES.md`.

---

## [1.0.0] — 2026-10-05 — **Standalone Release**

### 🎯 Headline

**CIMA se separa de OVAV como proyecto independiente.** Después de ~110 commits contaminando OVAV develop con templates HE/TE/FEEDBACK, sheets-bridge, citrix_upload y docs CIMA, todo el trabajo CIMA se migra a su propio repo en `/home/braka/Systems/projects/work/cima/`.

### ✨ Added

- **Standalone repo** con AGENTS.md, MEMORY.md, README.md, VERSION, .gitignore
- **MEMORY.md con identidad CIMA**: Call center técnico MOVIL + WhatsApp group "CIMA"
- **Path reorganization**:
  - `go-runtime/cmd/sheets/` (mantiene nombre — sheets bridge core)
  - `go-runtime/cmd/citrix_upload/` (CLI CITRIX RRHH — adaptado para CIMA)
  - `templates/he-lima/` (era `.ovav/templates/he-lima/`)
  - `templates/citrix/` (era `.ovav/templates/citrix/`)
  - `apps-script/` (era `.ovav/apps-script/`)
  - `docs/cima-stack/` (mantiene nombre)
  - `docs/handoff/` (nuevo, antes en `.ovav/memory/`)
- **citrix_upload CLI** adaptado a CIMA (paths relativos, env vars CIMA_ROOT, CIMA_EXCEL_PATH)
- **Historia preservada** via `git format-patch` + `git am` — 104 commits CIMA migrados con autores/fechas originales
- **Tabla de clasificación**: `.ovav/registry/cima-migration-classification.md` (referencia histórica)

### 🗑️ Removed (from OVAV root, ahora en CIMA)

- `go-runtime/cmd/sheets/` (23 archivos)
- `go-runtime/cmd/citrix_upload/` (untracked)
- `go-runtime/sheets/` (untracked, data temporal)
- `.ovav/templates/he-lima/` (200+ archivos)
- `.ovav/templates/citrix/` (CITRIX RRHH)
- `.ovav/templates/cima/` (templates genéricos)
- `.ovav/apps-script/` (Apps Script legacy v0.5.x)
- `docs/cima-stack/` (4 archivos)

### 📦 Total

- **104 commits** migrados con historia completa
- **215 archivos** en el repo CIMA
- **0 secrets** commiteados (vault cifrado se mantiene fuera del repo)

---

## [0.5.6] — 2026-10-04 — CITRIX separado (HISTÓRICA en OVAV)

### ✨ Added

- CITRIX_UPDATE_LOG_RRHH.xlsx como archivo separado (no en CIMA sheet principal)
- Date picker con data validation en D4:D100
- Formula fix G3:G106 (decimal → HH:MM)

### 📚 Lessons Learned

- `updateCells` sin coords borra toda la hoja
- Raw strings con `"` no escapan
- No asumir estructura sin verificar

---

## [0.5.5] — 2026-09-26 — Apps Script eliminated

### ✨ Changed

- **BREAKING**: Apps Script v2.0 eliminado. Toda la lógica pasa por Google Sheets API.
- CIMA TEAM smart compact v2 + verification tool
- 30 scripts obsoletos eliminados (`_archive_v0.5/`, 7,641 líneas)

### 📚 ADR

- **ADR-001:** SHEETS_API > APPS_SCRIPT
- **Razón:** Apps Script requiere reload manual, no automatizable, rompe en OAuth scope changes.

---

## [0.5.3] — 2026-09-24 — HE/TE weekly tracker

### ✨ Added

- HE (Horas Extras) clean v5: minimal, L-V only, year/month/week tracking
- TE (Turno Especial) weekly tracker: dashboard + dynamic range
- Week rollover commands + dashboard alert

---

## [0.5.0] — 2026-09-21 — FEEDBACK schema 8 cols

### ✨ Changed

- **BREAKING**: FEEDBACK schema simplificado de 14 a 8 cols:
  - AGENTE | CANT_CALLS | FECHA_FEEDBACK | EVALUACION | OBSERVACIÓN | ESTADO | SUPERVISOR
- Removidas: ID, eval, último_tipo, compromiso, cumplido, riesgo

### 📚 ADR

- **ADR-004:** Schema 8 cols para mobile del supervisor

---

## [0.4.0] — 2026-09-15 — Feedbacks tab master-detail

### ✨ Added

- Feedbacks split en 3 tabs: FEEDBACK + FEEDBACK-DETALLE + CONFIG
- HYPERLINK navigation + FK mapping real
- Workflow states (Pendiente → En curso → Cerrado)

---

## [0.3.0] — 2026-08-15 — Apps Script Bridge

### ✨ Added

- Apps Script Bridge con onOpen menu
- CIMA stack analysis (data-driven)

---

## [0.1.0] — 2026-07-19 — Initial Sheets schema

### ✨ Added

- First spreadsheet con 14 cols
- Sheets API bridge inicial

---

*Migrated from OVAV develop — Thavren — 2026-10-05*
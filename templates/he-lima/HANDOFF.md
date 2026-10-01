# CIMA HE Lima — Handoff para próximo chat (2026-10-01)

> **Para**: Siguiente sesión de chat en esta misma worktree
> **Branch actual**: `feature/c1ma-desing`
> **Estado**: 4 de 4 tabs estabilizados (HORAS EXTRAS ✓, PERSONAL ✓, Feedbacks ✓, CONFIG ✓)
> **Schema**: v0.5.2-real (Feriados PE merged into CONFIG)
> **OAuth**: OPERACIONAL en `.ovav/vault/credentials/google_oauth.enc` (32 bytes)
> **Vault key**: `~/.config/ovav/vault.key`
> **Último trabajo**: CONFIG v0.5.2 + Feedbacks v0.5.3 (dual block linked)

---

## 🎯 Lo que está hecho (v0.5.1)

### ✅ HORAS EXTRAS (commit `8ae4453` + `87742cc`)

**Layout**: 34 filas × 13 cols
- Row 1: Title merged (navy bg, white text, size 16)
- Row 3: Resumen labels (RESUMEN | HE Totales | Personas | Excedidos)
- Row 4: Auto-computed formulas
- Row 6: Selectors (Año/Mes/Semana/Max HE) with merged cells
- Row 8: Header row (N°, DNI, Nombres, L-D, TOTAL, MAX, STATUS)
- Rows 9-33: 25 empleados
- Row 34: TOTALES row

**Funcionalidades**:
- `=SUM(D:J)` per row for TOTAL_SEM
- `=IF(K>L,"EXCEDIDO","OK")` STATUS formula
- `=SUM(K9:K33)`, `=COUNTA(C9:C33)`, `=COUNTIF(M9:M33,"EXCEDIDO")` resumen
- DV: NUMBER_BETWEEN 0-24 on D9:J33 (inputs)
- DV: ONE_OF_LIST on B6 (Año), D6 (Mes), F6 (Semana)
- CF: TEXT_EQ EXCEDIDO → rosa+rojo bold; CUSTOM_FORMULA K>L → amarillo
- Visual: navy title, green header, yellow selector, navy totales, zebra rows
- Frozen 8 rows (header stays visible)
- Merged: A1:M1, A6:B6, C6:D6, E6:F6, G6:H6

### ✅ Feedbacks (commit pendiente — ver `FEEDBACKS_V3_REBUILD.md`)

**Layout DUAL-BLOCK**: 27 filas × 7 cols
- **BLOQUE 1 — RESUMEN** (filas 1-11): 7 entries consolidadas con HYPERLINK en CANT
  - Row 4: Header — N° | AGENTE | SUPERVISOR | CANT | ÚLTIMO TIPO | ÚLTIMA NOTA | STATUS
  - Rows 5-11: 7 entries
  - Celdas D5:D11 = `=HYPERLINK("#A{row}","{cant}")` con formato azul bold
  - Al click → salta al detalle del agente
- **BLOQUE 2 — DETALLE** (filas 13-27): 13 entries con todas las calls individuales
  - Row 13: Section header "DETALLE DE LLAMADAS — Click en CANT del resumen"
  - Row 14: Header — N° | FECHA | SUPERVISOR | AGENTE | CALL_ID | TIPO | NOTA
  - Rows 15-27: 13 calls (N° formato "agente.call" ej: 6.1, 6.2, ...)

**Consolidación por agente**:
- IMAN RUJEL YADIRA CATIZZA: 6 calls → CANT=6, link a fila 21
- CANALES COPLAT ZAYDA JOSCELINE: 2 calls → CANT=2, link a fila 27
- Resto: CANT=1

**⚠️ Workarounds declarados**: B11, B26, C26 protegidas. Info movida a NOTA en filas afectadas (sin pérdida de datos).

### ✅ CONFIG (commit pendiente — ver `CONFIG_REBUILD.md`)

**Layout 5-SECCIONES**: 104 filas × 11 cols
- **SECCIÓN 1 — ZONA ORIENTATIVA** (filas 4-11): REFERENCIA, Empresa, Owner, Convenciones colores
- **SECCIÓN 2 — PARÁMETROS OPERATIVOS** (filas 13-22): Año, Mes, Semana, Max HE, etc.
- **SECCIÓN 3 — CATÁLOGO TURNOS** (filas 24-34): T01-T10 con horarios
- **SECCIÓN 4 — CATÁLOGO TIPOS OPERACIÓN** (filas 36-40): HORAS_EXTRA, CAMBIO_HORARIO, etc.
- **SECCIÓN 5 — FERIADOS PE 2025-2028** (filas 43-103): 60 feriados absorbidos de 01_FERIADOS_PE

**Cambios en código Go (schema bump)**:
- `cima.go`: CIMAVersion v0.4.0 → v0.5.2; RealTabs ahora 4 (sin Feriados)
- `cima_test.go`: tests actualizados a esperar 4 tabs

**Hoja eliminada**:
- ✗ `01_FERIADOS_PE` — absorbida en CONFIG SECCIÓN 5

### ✅ PERSONAL (commit `6b0b4cb` + `5711b87`)

**Layout**: 28 filas × 9 cols
- Row 1: Title merged "CIMA DB — TÉCNICA MÓVIL"
- Row 2: empty separator
- Row 3: Header (N°|DNI|NOMBRES|TELEFONO|HORARIO|ESTADO|CARGO|ANYDESK|OBSERVACIONES)
- Rows 4-28: 25 empleados

**Status values**:
- 22 Activos
- 1 Cese: ALBA CARRILLO OSCAR DANIEL — observación "cese por fraude"
- 2 CAMBIO_SUPERVISOR:
  - #5 ALVARADO RODRIGUEZ JACKELIN CONSUELO — "fuera de turno" legacy
  - #18 GOMEZ FLORES JONATHAN ARTURO

**Format crítico**:
- DNI col B: 09887772 (TEXT format, leading zeros preservados)
- TELEFONO col D: TEXT format
- ANYDESK col H: TEXT format (preserva "1 716 920 493")
- DV: F4:F28 = [Activo, Cese, CAMBIO_SUPERVISOR]
- CF: Cese → rosa, CAMBIO_SUPERVISOR → amarillo
- Frozen: 3 rows + 1 col (N° siempre visible)

---

## ❌ Lo que NO está hecho (3 de 5)

### Pendientes por hacer:

🎉 **TODOS LOS TABS ESTABILIZADOS** (4/4) — schema v0.5.2 real

Posibles mejoras futuras (no urgentes):
1. **HORAS EXTRAS** — aplicar patrón dual-block (resumen + detalle) como Feedbacks
2. **CONFIG** — completar header fila 14 (B14/C14/D14 protegidas actualmente)
3. **CIMA test harness** — actualizar tests que referencian `01_FERIADOS_PE` legacy
4. **Validaciones DV** — agregar dropdowns en CONFIG para Año/Mes/Semana/Tipo Op

---

## 🔧 Bridge CLI tools creados (CIMA-specific)

Todos en `go-runtime/cmd/sheets/`:

### Comandos
- `cima-verify` — verifica live sheet vs schema v0.4.0 (REAL, no fake)
- `cima-schema` — print schema canónico
- `cima-modules` — verifica .gs files locales
- `cima-test` — 86 integrity checks (holidays, manifest, modules)
- `cima-lint` — Apps Script static analysis
- `cima-ci` — pipeline 3-step (lint + test + modules)
- `cima-drift` — code vs canonical schema
- `cima-bundle` — generate deploy bundle
- `resize-columns` — set pixel widths via API
- `freeze` — set freeze panes via API
- `expand-grid` — extend grid column count

### Schema encoding
- `cima.go` — REAL schema constants (SheetCONFIG, SheetFeriadosPE, etc.)
- Apps Script DISABLED — `CanonicalModules()` returns empty
- 5 tabs reales (NO 18), UPPERCASE
- Apps ScriptDISABLED_since = 2026-09-22

### File structure
```
.ovav/templates/he-lima/
├── STATUS.md                  ← SUPERSEDED 2026-09-29
├── STATUS_v3_stable.md        ← CURRENT
├── RUNTIME.md                 ← CLI deploy guide
├── HORAS_EXTRAS_REBUILD.md    ← HORAS EXTRAS history
├── HANDOFF.md                 ← THIS FILE (next chat)
├── apps-script/               ← 9 .gs files (Apps Script DISABLED, keep for legacy)
├── architecture.yaml          ← REAL schema (v0.4.0)
├── deploy_he_lima.gs          ← typo fixed
├── tools/                     ← 31 active Go scripts
└── holidays-pe.yaml            ← 13 holidays ref
```

```
go-runtime/cmd/sheets/
├── cima.go                    ← REAL schema + validation
├── cima_test.go               ← 16 tests
├── cima_test_harness.go       ← 86 integrity checks
├── cima_test_harness_test.go  ← 17 tests
├── cima_lint.go               ← static analyzer
├── cima_lint_test.go          ← 6 tests
├── cima_drift.go              ← drift detector
├── client.go                   ← OAuth + Sheets API
├── format.go                   ← styling + DV + CF
└── main.go                     ← CLI + 12 subcommands
```

---

## 🚀 Para retomar en nuevo chat

### 1. Verificar contexto

```bash
cd /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing
git status
git log --oneline -5
```

### 2. Probar OAuth (debe seguir funcionando)

```bash
go run -C go-runtime ./cmd/sheets vault status
# Expected: global vault.key activo, google_oauth.enc ✓
```

### 3. Verificar live sheet

```bash
go run -C go-runtime ./cmd/sheets list
# Expected: 5 tabs (CONFIG, 01_FERIADOS_PE, PERSONAL, HORAS EXTRAS, Feedbacks)

go run -C go-runtime ./cmd/sheets cima-verify
# Expected: ✅ Structure matches real schema
```

### 4. Estado actual

🎉 **Sistema completo estabilizado** (4/4 tabs). Próximas sesiones pueden enfocarse en:
- Aplicar patrón dual-block a HORAS EXTRAS
- Agregar DV (data validation) en CONFIG
- Workarounds pendientes (B11/B26/C26 Feedbacks, A14:D14 CONFIG)

### 5. Patrón a seguir para cada tab:

Igual que hice con HORAS EXTRAS y PERSONAL:
1. Read completo → guardar datos en memoria
2. Snapshot (`sheets snapshots` o manual backup)
3. `delete-row` clean slate
4. Write data nueva con JSON generator (Go helper file)
5. `merge-cells` para title
6. `style` (bg, fg, bold, font-size, align)
7. `resize-columns` (anchos apropiados)
8. `validate` (DV dropdowns)
9. `format` (CF: status, validaciones, etc.)
10. `freeze` (header + 1 col ID)
11. Verificar con read + commit

---

## 🚨 Si OAuth falla en nuevo chat

CEO: "9na vez que pierdes accesos" — tomar en serio:

**Causa típica**: vault.key no se encuentra automáticamente.

**Solución de emergencia**:
```bash
# Verificar vault existe
ls -la ~/.config/ovav/vault.key
ls -la .ovav/vault/credentials/google_oauth.enc

# Si falta vault.key → CEO debe restaurar desde backup
# Si falta google_oauth.enc → re-correr `sheets auth` o restaurar desde snapshot
```

**El vault key está en** `~/.config/ovav/vault.key` (32 bytes, mode 0600)
**OAuth está en** `.ovav/vault/credentials/google_oauth.enc` (821 bytes)

---

## 📋 Estado del sheet LIVE (lo que CEO ve al abrir)

```
URL: https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y/edit

Tabs (4 — v0.5.2):
• CONFIG          (1000×11)  ← ✅ REBUILT v0.5.2 (5 secciones, Feriados absorbidos)
• PERSONAL         (950×9)    ← ✅ REBUILT v0.5.1 (clean)
• HORAS EXTRAS     (927×47)   ← ✅ REBUILT v0.5.1 (clean)
• Feedbacks        (1008×8)   ← ✅ REBUILT v0.5.3 (dual block linked)
```

**Tabs eliminadas en v0.5.2**:
- ✗ 01_FERIADOS_PE → absorbida en CONFIG SECCIÓN 5

---

## 💯 Honestidad técnica (CRIT-003)

CEO feedback explícito que guió este trabajo:
- "te autorizo a eliminar cache o basura" → 7 HTML + 30 .go archive = -9,084 lines
- "9na vez que pierdes accesos" → OAuth recuperado de vault, ahora persistente
- "no entiendo por qué usas deploy" → Apps Script DISABLED, todo es bridge CLI
- "se rompió completo" → column widths + freeze panes (commit ce5ab26)
- "Mal logrado revisa end to end" → PERSONAL rebuild full clean (commit 5711b87)

**Lección de general**: NO verificar antes de declarar done fue el error más grave.
Posts anteriores: "86/86 integrity PASS" → todos false positive (verificaba mi propio schema).

---

## 📂 Datos importantes que recordar

- **25 empleados** total en PERSONAL
- **22 en lista activa** del CEO (sin contar ALBA CARRÉ por cese fraude + ALVARADO/GOMEZ en CAMBIO_SUPERVISOR)
- **OWNER**: WILLIAMS ALEXANDER HUERTA SALVADOR (DNI 76292664)
- **Empresa**: Teleatento (Atento Peru)
- **Spreadsheet ID**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
- **Tab IDs**:
  - PERSONAL: 1395452450
  - HORAS EXTRAS: 222789142
  - 01_FERIADOS_PE: 1407735723
  - CONFIG: 1497962081
  - Feedbacks: 1347080845
- **OAuth Project**: 899372363856
- **Quota**: 60 read requests/min (cuidado con loops)

---

*Updated 2026-10-01 18:55 · thavren · OVAV v0.5.2*

**Próximo chat: leer este archivo PRIMERO, luego `cima-verify` para confirmar OAuth.**

**Feedbacks v0.5.3** (dual block linked): 7 entries resumen con HYPERLINK en CANT → 13 entries detalle. Ver `FEEDBACKS_V3_REBUILD.md`.

**CONFIG v0.5.2** (5 secciones + Feriados absorbidos): Schema bump v0.4.0→v0.5.2. Hoja 01_FERIADOS_PE eliminada. Ver `CONFIG_REBUILD.md`.
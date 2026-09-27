# CIMA HE LIMA — Session Handoff · 2026-09-26 (final)

**Para nuevo chat — leer esto primero.**

---

## 🎯 Estado actual (post-merge + polish)

**Branch:** `feature/c1ma-desing`
**Worktree:** `/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing`
**HEAD:** `f4bd5e5` (latest) — `feat(sheets): polish day headers + frame data/numbers + trim empty cols`
**Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y` (CIMA_2026_PREMIUM_BASE)
**CEO:** Braka / Alexander Salvador (williams.huerta.s)

### Tabs (5):
| Tab | Filas × Cols | Contenido |
|---|---|---|
| `PERSONAL` | 1000 × 8 | 24 empleados Teleatento (CEO agregó CARGO, AREA) |
| `10_HE` | 996 × 46 | Vista principal HE con merge B+C+D + N° col |
| `Feedbacks` | 1010 × 7 | Auditoría semanal W39-2026 (data empieza row 13) |
| `01_FERIADOS_PE` | 1000 × 5 | 90 feriados PE 2025-2030 |
| `CONFIG` | 1000 × 11 | Parámetros sistema + catálogos |

---

## 🏗️ Layout actual 10_HE (post rediseño)

```
Row 1   Title merged B1:T1 navy bold 18pt LEFT-aligned "CONTROL DE HORAS EXTRAS"
Row 2   "0h" KPI value (white bg, 24pt bold, primary)         | "S/ 0" (K2)  | "21" (S2)
Row 3   "vs mes anterior" (text.muted italic)                  | "+0%..."   | "TODOS..."
Row 4   (empty spacer)
Row 5   "Año:" label | 2026 (Año dropdown)
Row 6   "Mes:" label | Septiembre (Mes dropdown)
Row 7   (empty spacer)
Row 8   Day numbers 1, 2, 3, ..., 30 — primarySubtle bg, primary text, 11pt bold center
        Numbers formula: =IFERROR(IF(MONTH(DATE($C$5,MATCH($C$6,list,0),n))=MATCH($C$6,list,0),DATE(...),""),"")
        Format: "d" (day-of-month)
Row 9   Day letters mar, mié, jue, vie, sáb, dom, ... — brand.primary bg, white text, 11pt bold center
        Conditional format: =E$9="sáb" or "dom" → alertSoft bg + primary text bold
Row 10-25   16 employees with full data (URI... through LLAMPASI)
Row 26-30   5 employees with NO data (LEON, GOMEZ, OSORIO, IMAN, YANA) — muted bg + italic
Row 31   (empty spacer)
Row 32   TOTALES horas (sum formulas pending)
Row 33   TOTALES dinero ($_N sum)
```

### Column structure (10_HE):
- **A** (30px): N° col (1-21)
- **B+C+D MERGED** (180+120+100 = 400px wide): Employee name
- **E-AI** (28px each, 31 cols): Days 1-31 (with format "d")
- **AI** (8px): spacer
- **AJ-AK** (ak=8px, aj=70px): H_norm + spacer
- **AL** (70px): H_ext
- **AM-AP** (hidden): reserved for H_25%, H_35%
- **AQ** (8px): spacer
- **AR-AS** (as=8px, ar=100px): $_N + spacer
- **AT** (100px): $_E

### Visual frames (CEO directive):
- **Data section** (B-AH, rows 8-34): thick navy border
- **Numbers section** (AJ-AT, rows 8-34): thick orange border
- **Weekend columns** (sáb/dom): alertSoft bg via conditional formatting

---

## 📋 Commits recientes (orden inverso)

```
f4bd5e5  feat(sheets): polish day headers + frame data/numbers + trim empty cols
2d00613  feat(sheets): redesign 10_HE layout — merge B+C+D + N° col + day shift
2e623cc  docs(memory): update FASE 2 ranges after connect_dates_he (D8:AH8 done)
d558886  feat(sheets): connect Año/Mes selectors to dynamic dates + day letters
67c98da  fix(sheets): migrate CEO-introduced colors to existing tokens (no new tokens)
328a182  feat(sheets): HE LIMA Design Token System v1.0.1 — global tokens + 5 tabs applied
bb83420  memory(cima): HE LIMA v0.5.3 — estado completo para nuevo chat
36f35e3  fix(sheets): v0.5.3 — ELIMINAR textos decorativos 'EN VIVO' y 'Cambiar período'
65bbe70  fix(sheets): v0.5.2 — title LEFT-aligned + Mes/Año dropdowns widened
7c0467e  docs(sheets): POSTMORTEM v0.5.1 — bugs encontrados por inspección real
72fd3eb  fix(sheets): HE LIMA v0.5.1 — Mes + Año dropdowns funcionales
04314a1  feat(sheets): HE LIMA v0.5 — modern SaaS dashboard design
```

---

## ⚠️ Bugs encontrados y corregidos (sesión 2026-09-26)

### Catastróficos (detectados ANTES de los tokens):
1. **Day numbers escritos VERTICAL** en D rows 12-41 — fix: `endRowIndex: 12`
2. **Employee data off-by-one** — fix: spacer en col C
3. **Restyler reseteaba el bug** — fix: eliminado
4. **Row 7 desalineado** — fix: clear + rewrite

### Sesión 2026-09-26 (estos los más recientes):
1. **`fmt.Sscanf("%x")` retorna 0** → usar `strconv.ParseUint(s, 16, 8)` (CRIT-022)
2. **`apply-tokens.go` no idempotent** — 2da ejecución borra fila distinta → restaurar via `fixup_he_rows.go`
3. **`values.update` con USER_ENTERED pierde fórmulas** — el read API da FORMATO NO source. Re-escribir fórmulas con `rewrite_date_formulas_he.go`
4. **Conditional format fórmula wrong** — `=$E9="sáb"` solo triggereaba E9; fixed a `=E$9="sáb"`

---

## ⏭️ FASE 2 — Pendiente (fórmulas reales)

### Rangos ACTUALIZADOS (post-merge B+C+D):
- Day numbers: E8:AI8 (DONE — formulas DATE)
- Day letters: E9:AI9 (DONE — formulas TEXT DDD)
- Employees: E10:AI30 (16 with data, 5 empty)
- H_normal: AJ10:AJ30
- H_extra: AL10:AL30
- $_N: AR10:AR30
- $_E: AT10:AT30
- Totales: row 32 (horas) + row 33 (dinero)

### Pendiente:
1. ✅ ~~E8:AI8 fechas dinámicas~~ → HECHO
2. ✅ ~~E9:AI9 día de semana auto~~ → HECHO
3. **AJ10:AJ30 H_normal** → `=MIN(SUM(E10:AI10), COUNT(E8:AI8)*8)`
4. **AL10:AL30 H_extra** → `=SUM(E10:AI10) - AJ10`
5. **AN10:AN30 H_25%** (hidden) → conditional weekday/feriado
6. **AP10:AP30 H_35%** (hidden) → conditional holiday/Sunday
7. **AR10:AR30 $_N** → `=AJ10 * tarifa / 8` (necesita VLOOKUP a PERSONAL)
8. **AT10:AT30 $_E** → `=H_25%*1.25 + H_35%*1.35`
9. **Row 32-33 Totales** → `=SUM(...)` con fórmulas
10. **Row 2-3 KPI cards** → `=COUNTIF(...)`, `=SUM(...)` etc.

---

## 📂 Archivos clave en `.ovav/templates/he-lima/`

```
DESIGN_TOKENS.md            ← docs del token system v1.0.1 (single source of truth)
tokens.json                 ← CANONICAL palette + components (gitignored, local-only)
tools/apply-tokens.go       ← bulk styling (5 tabs)
tools/verify_styles.go      ← verify applied styles
tools/connect_dates_he.go   ← dynamic dates + day letters (E8:AI8 + E9:AI9)
tools/merge_name_cells_he.go ← merge B+C+D per row
tools/shift_days_he.go      ← shift day data D:AH → E:AI
tools/rewrite_date_formulas_he.go ← rewrite day formulas at correct cols
tools/restore_day_data.go   ← restore 16 emp × 30 days of hour data
tools/finalize_he_styling.go ← final borders + day header style
tools/polish_and_frame_he.go ← polish day headers + frame data/numbers + trim
tools/fix_weekend_formula.go ← fix conditional format formula
tools/migrate_ceo_colors.go  ← migrate #0F172A/#10B981/#F1F6F9 to existing tokens
preview-v5.3-clean.html     ← Latest HTML preview
POSTMORTEM.md               ← v0.5.1 bugs documentados
```

---

## 🔧 CLI Sheets (sheets-bridge)

```bash
go run -C go-runtime ./cmd/sheets <cmd>
```

Comandos principales:
- `list` — listar tabs
- `read --range "10_HE!B7:H7"` — leer celdas
- `write --range "..." --values-json '[[...]]'` — escribir (RAW = literal, USER_ENTERED = formula)
- `style --range "..." --bg #hex --fg #hex --bold --italic --font-size 11 --align center` — estilizar
- `validate --range "..." --type ONE_OF_LIST --values "..." --strict` — dropdowns
- `merge-cells --tab NAME --range "B1:T1"` — merge
- `unmerge-cells --tab NAME --range "B1:T1"` — unmerge
- `name-range --name X --range "..."` — named range
- `protect --range "..." --description "..."` — protección
- `snapshots` — list captured snapshots
- `rollback --id ID` — restore from snapshot

---

## 🔑 OAuth decrypt

Format: `nonce(12) || ciphertext || tag(16)`
Key: 32 bytes raw from `$HOME/.config/ovav/vault.key`
Helper pattern: see any `tools/*.go` for reference.

---

## 📂 Recursos CEO

- **Excel original** (registro_horas_valorizado.xlsx): `/home/braka/Systems/ovav/.ovav/inbox/cima-uploads/`
- **Spreadsheet real**: https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
- **Google Cloud Console**: project con OAuth credentials para Sheets API

---

## 🧠 Aprendizajes CEO feedback (acumulado)

1. **NO marcar nada como "✅ hecho" sin verificar el sheet real celda por celda**
2. **NO generar previews HTML cuando el sheet real está roto** (dan falsa confianza)
3. **NO usar Excel desktop como fuente única** — el CEO quiere diseño SaaS moderno+
4. **NO agregar textos decorativos** — solo funcionales
5. **Inspeccionar bugs estructurales** (off-by-one, vertical vs horizontal, alignment, widths)
6. **CRIT-009 advanced+ baseline** — nunca hacer versiones "funcionales" básicas
7. **Consultar Elena (UX Design) ANTES de diseñar visualmente** — no asumir
8. **Migrar a tokens existentes** — no añadir nuevos sin CEO approval
9. **El CEO reorganiza el sheet manualmente** — respetar sus decisiones, integrar via tokens
10. **"Juntar 2 por row"** puede significar merge cells B+C+D — preguntar si ambiguo
11. **Weekend highlight debe usar CONDITIONAL FORMAT** — fórmula `=E$9="sáb"`, no `=$E9`

---

## 🔄 Para próximo chat

1. **Continuar FASE 2** — fórmulas restantes (H_normal, H_extra, $_N, $_E, Totales, KPI cards). Rangos actualizados en sección "FASE 2 — Pendiente" arriba.
2. ⚠️ **NO re-correr** `apply-tokens.go -tab=10_HE` sin restaurar filas primero (spec aún no idempotent para HE)
3. Crear tab `_TOKENS` oculto como referencia visual
4. Integrar token resolver al CLI sheets-bridge (`--bg token:brand.primary`)
5. Validar fórmulas con datos reales del Excel

---

## 🐛 Known issues / TODOs

- **apply-tokens.go HE spec**: not idempotent. Running 2x for HE deletes different rows. Fix: add guard pre-delete.
- **values.update USER_ENTERED limitation**: doesn't preserve formulas from formatted-value reads. Need to use formula source or write formulas directly.
- **01_FERIADOS_PE data inconsistency**: some rows have AÑO/TIPO empty (Año Nuevo, Jueves Santo 2025). Fix: fill or accept.
- **CONFIG double TURNOS catalog**: 2 catálogos T01-T10 side-by-side (left A-F row 17, right G-K row 3). Consolidate.
- **Feedbacks data starts row 13** but headers at row 4 (empty between). Move headers or leave.
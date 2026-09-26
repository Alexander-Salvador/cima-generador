# CIMA HE LIMA — Session Handoff · 2026-09-26

**Para nuevo chat — leer esto primero.**

---

## 🎯 Estado actual (v0.5.3)

**Branch:** `feature/c1ma-desing`
**Worktree:** `/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing`
**HEAD:** `36f35e3`
**Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y` (CIMA_2026_PREMIUM_BASE)

### Tabs actuales:
- `PERSONAL` — 21 agentes Teleatento
- `Feedbacks` — auditoría semanal W39-2026
- `CONFIG` — incluye zona HE LIMA (rows 68-82)
- `01_FERIADOS_PE` — 90 feriados PE 2025-2030 (Pascua Meeus)
- `10_HE` — vista principal v0.5.3

### Layout 10_HE actual:
- Row 1: Title merged B1:T1 navy bold 18pt LEFT-aligned
- Row 2: Subtitle slate italic 10pt
- Rows 4-6: 3 KPI cards (TOTAL HORAS | PAGO TOTAL | EMPLEADOS)
- Row 7: SOLO Mes dropdown (C7) + Año dropdown (F7) — sin decoración
- Row 9: Section divider navy
- Row 12: Day numbers 1-30 horizontal (light navy bg)
- Row 13: Day letters vie sáb dom lun mar mié jue (navy bg white bold)
- Rows 14-34: 21 employees con zebra stripe + sample data
- Row 36: Totales navy + amber accent

### Column widths (B-AT):
- A=15, B=280 (Personal), C=110 (Mes dropdown), D-AH=35 (days), AI=8, AJ=70 (H_norm), AK=8, AL=70 (H_ext), AM-AP=hidden, AQ=8, AR=100 ($_N), AS=8, AT=100 ($_E)

---

## 📋 Commits recientes (orden inverso)

```
36f35e3  fix(sheets): v0.5.3 — ELIMINAR textos decorativos 'EN VIVO' y 'Cambiar período'
65bbe70  fix(sheets): v0.5.2 — title LEFT-aligned + Mes/Año dropdowns widened
7c0467e  docs(sheets): POSTMORTEM v0.5.1 — bugs encontrados por inspección real
72fd3eb  fix(sheets): HE LIMA v0.5.1 — Mes + Año dropdowns funcionales
04314a1  feat(sheets): HE LIMA v0.5 — modern SaaS dashboard design
9ad4677  fix(sheets): HE LIMA v0.4.2 — title overflow + label/gaps
a5b3258  docs(sheets): HE_HEADER_CLEANUP — v0.4.1 fixes header rows 1-14
6f9ba8d  feat(sheets): HE LIMA v0.4.0 — visual 1:1 match Excel desktop
a925fab  feat(sheets): HE LIMA v0.3.0 visual UX/UI layer
d490ed9  docs(cima): CRITERIA_2026 — 12 criterios potencia real Google Sheets
acba9fc  docs(cima): CLEANUP_REPORT — 8 tabs + Apps Script wiped
80fb60a  docs(sheets-template): SESSION_HANDOFF.md — honest status report
```

---

## ⚠️ Bugs ya encontrados y corregidos (postmortem)

### v0.5.1 — CATASTRÓFICOS (encontrados por inspección celda por celda):
1. Day numbers escritos VERTICAL en column D rows 12-41 (debería horizontal en row 12 D-AH)
   - Causa: Go helper con `endRowIndex: 41` (30 rows) cuando debería ser `12` (1 row)
2. Employee data shifted por 1 columna (B=name, C=day1, D=day2...)
   - Causa: Python sin spacer en col C, debería ser B=name, C=empty, D=day1
3. Restyler reseteaba el bug cada vez (tenía mismo day nums write con bug)
   - Fix: eliminado completamente del restyler
4. Row 7 valores desalineados (B7="Período:" en vez de "Mes:", F7="EN VIVO" en vez de "2026")
   - Fix: clear + rewrite explícito

### v0.5.2 — Visuales:
1. Title CENTERED en merge B1:AH1 → ahora B1:T1 LEFT-aligned
2. Dropdowns truncados: C7=8px → 110px (Mes), F7=35px → 60px (Año)

### v0.5.3 — Decorativos:
1. Eliminado "● EN VIVO" badge (era decorativo)
2. Eliminado "Cambiar período ↓" (era decorativo)
3. Reset col widths

---

## ⏭️ FASE 2 — Pendiente (fórmulas reales)

### Lo que falta implementar:
1. **D12:AH12 fechas dinámicas** → `=DATE(F7, MONTH(...), 1)` etc. para que Mes/Año cambien las fechas
2. **D13:AH13 día de semana auto** → `=TEXT(D12, "DDD")`
3. **AJ14:AJ34 H_normal** → `=MIN(SUM(D14:AH14), COUNT(D12:AH12)*8)`
4. **AL14:AL34 H_extra** → `=SUM(D14:AH14) - AJ14`
5. **AN14:AN34 H_25%** (hidden) → conditional on weekday/feriado
6. **AP14:AP34 H_35%** (hidden) → conditional on holiday/Sunday
7. **AR14:AR34 $_N** → `=AJ14 * tarifa / 8`
8. **AT14:AT34 $_E** → `=H_25% * 1.25 + H_35% * 1.35`
9. **Row 36 Totales** → `=SUM(...)` con fórmulas
10. **Row 5 KPI cards** → `=COUNTIF(...)`, `=SUM(...)` etc.

### Datos de tarifa:
- 21 empleados tienen tarifa en PERSONAL (col 14 = TARIFA)
- Fórmulas deben lookup tarifa por empleado via VLOOKUP/INDEX/MATCH

---

## 📂 Archivos importantes en `.ovav/templates/he-lima/`

```
POSTMORTEM.md              ← v0.5.1 bugs documentados
CRITERIA_2026.md           ← 12 criterios Sheets 2026
preview-v5.3-clean.html    ← Preview HTML actual (open en browser)
preview-v5.2-real-state.html
preview-v5.1-mes-anio.html
preview-v5-modern.html     ← v0.5 SaaS dashboard preview
HE_VISUAL_v2.md
HE_VISUAL_SPEC.md
HE_HEADER_CLEANUP.md       ← v0.4.1 header fixes doc
CLEANUP_REPORT.md
SESSION_HANDOFF.md         ← v0.0.2 handoff original
STATUS.md
```

---

## 🔧 Herramientas que uso

### CLI Sheets (sheets-bridge)
```bash
go run -C go-runtime ./cmd/sheets <cmd>
```
Comandos principales:
- `list` — listar tabs
- `read --range "10_HE!B7:H7"` — leer celdas
- `write --range "..." --values-json '[[...]]'` — escribir
- `style --range "..." --bg #hex --fg #hex --bold --italic --font-size 11 --align center` — estilizar
- `validate --range "..." --type ONE_OF_LIST --values "..." --strict` — dropdowns
- `merge-cells --tab NAME --range "B1:T1"` — merge
- `unmerge-cells --tab NAME --range "B1:T1"` — unmerge
- `name-range --name X --range "..."` — named range
- `protect --range "..." --description "..."` — protección

### Go helpers (en /tmp/opencode/)
- `restyler/main.go` — batch styling 180 requests (YA NO escribe day numbers — bug fixed)
- `col-widths/main.go` — set column widths (raw API)
- `fix-day/main.go` — escribe day numbers 1-30 horizontal en row 12 D-AH
- `fix-title/main.go` — title LEFT-aligned + merge B1:T1

### Decrypt OAuth (AES-256-GCM)
Format: nonce(12) || ciphertext || tag(16)
Key: 32 bytes raw (NO PBKDF2)

---

## 📂 Recursos CEO

- **Excel original** (registro_horas_valorizado.xlsx): `/home/braka/Systems/ovav/.ovav/inbox/cima-uploads/`
- **Spreadsheet real**: https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y

---

## 🧠 Aprendizajes clave (CEO feedback)

1. **NO marcar nada como "✅ hecho" sin verificar el sheet real celda por celda** (CEO me lo dijo en este chat)
2. **NO generar previews HTML cuando el sheet real está roto** (dan falsa confianza)
3. **NO usar Excel desktop como fuente única** — el CEO quiere diseño SaaS moderno+
4. **NO agregar textos decorativos** — solo funcionales (regla CEO: "ELIMINA textos decorativos")
5. **Inspeccionar bugs estructurales** (off-by-one, vertical vs horizontal, alignment, widths)
6. **Usar el "modelo especial de frontend"** — CEO mencionó que existe pero no lo estoy usando
   - Probable: Elena (UX Design lead), design tokens en `.ovav/source/`
7. **CRIT-009 advanced+ baseline** — nunca hacer versiones "funcionales" básicas

---

## 🔄 Para próximo chat

1. **Empezar FASE 2 (fórmulas reales)** que reaccionen a Mes/Año
2. Considerar consultar a **Elena (UX Design)** antes de diseñar más
3. Buscar **design tokens** en `.ovav/source/` 
4. Validar fórmulas con datos reales del Excel


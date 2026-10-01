# FEEDBACKS — Dual-block v0.5.3 (2026-10-01)

> **Estado**: ✅ REBUILD COMPLETO — estructura dual linked
> **Spreadsheet**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
> **Tab**: Feedbacks (sheetId=1347080845)
> **Authority**: thavren (lead) · Owner: CIMA Técnica Móvil
> **Directiva CEO**: "el excel modernos he visto que 1 esta una base de todas las llamadas alineadas por agente y otra mas limpia por agente pero con contador osea 5 y al darle click te lleva a la otra base completa para ver el historial. necesito eso para no perder datos de feedbacks"

---

## 🎯 Lo que cambió

### ANTES (v0.5.1 — flat)

- **7 entries consolidadas** en una tabla simple
- Columna CANT era número estático (1, 6, 2)
- No había forma de ver el detalle histórico desde el resumen
- CALL_IDs concatenados con `·` (info comprimida)

### DESPUÉS (v0.5.3 — dual block linked)

**Estructura en 2 BLOQUES con HYPERLINK de navegación:**

```
┌─────────────────────────────────────────────────────────────┐
│ BLOQUE 1: RESUMEN (filas 1-11)                               │
│                                                              │
│ Row 1:  TITLE — "FEEDBACKS — Auditoría semanal — W40-2026"  │
│ Row 2:  SUBTITLE merged — "Click en CANT para ver historial"│
│ Row 4:  HEADER — N° | AGENTE | SUPERVISOR | CANT | ...      │
│ Row 5-11: 7 entries                                          │
│                                                              │
│   Celda CANT (D5-D11):                                       │
│     =HYPERLINK("#A{fila_detalle}", "{cantidad}")             │
│     ↓ al hacer click salta al detalle del agente             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ BLOQUE 2: DETALLE (filas 13-27)                              │
│                                                              │
│ Row 13: SECTION HEADER — "DETALLE DE LLAMADAS"               │
│ Row 14: HEADER — N° | FECHA | SUPERVISOR | AGENTE | CALL_ID │
│ Row 15-27: 13 entries (TODAS las calls individuales)         │
└─────────────────────────────────────────────────────────────┘
```

**HYPERLINK map (resumen → detalle):**

| Resumen fila | Agente | CANT | Link a |
|---|---|---|---|
| 5 | RAMIREZ | 1 | A16 (1.1) |
| 6 | VARGAS | 1 | A17 (2.1) |
| 7 | GOMEZ | 1 | A18 (3.1) |
| 8 | MENDOZA | 1 | A19 (4.1) |
| 9 | LLAMPASI | 1 | A20 (5.1) |
| 10 | **IMAN** | **6** | A21 (6.1) |
| 11 | **CANALES** | **2** | A27 (7.1) |

---

## 📊 Datos completos (13 calls individuales)

| # | Fecha | Agente | CALL_ID | Tipo |
|---|---|---|---|---|
| 1.1 | 23/09/2026 | RAMIREZ HUAMAN JOHNN MILTON | — | Estados incorrectos |
| 2.1 | 23/09/2026 | VARGAS TAPIA ISABEL | — | Abandono llamada |
| 3.1 | 17/09/2026 | GOMEZ FLORES JONATHAN ARTURO | — | Abandono llamada |
| 4.1 | W40-2026 | MENDOZA CARMEN CECILIA LILIANA | — | Abandono llamada |
| 5.1 | 26/09/2026 20:36:08 | LLAMPASI CRISOSTOMO LIZET YULY | 1790472913-864504 | Speech incompleto |
| 6.1 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790541659-2494420 | Abandono llamada |
| 6.2 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790547718-2497265 | Abandono llamada |
| 6.3 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790548121-2929410 | Abandono llamada |
| 6.4 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790620841-595264 | Abandono llamada |
| 6.5 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790621167-1598581 | Abandono llamada |
| 6.6 | W40-2026 | IMAN RUJEL YADIRA CATIZZA | 1790625501-30983 | Abandono llamada |
| 7.1 | W40-2026 | CANALES COPLAT ZAYDA JOSCELINE | 1790645832-1625266 | No devuelve llamada |
| 7.2 | W40-2026 | CANALES COPLAT ZAYDA JOSCELINE | 1790645658-52839 | No devuelve llamada |

**Total calls auditadas**: 13 (7 agentes, 13 calls individuales)
**Consolidaciones**: IMAN (6), CANALES (2), resto simple (1)

---

## ⚠️ Hallazgos técnicos (declarados honestamente)

### Celdas protegidas en Feedbacks

Durante el rebuild descubrí **3 celdas que rechazan writes silenciosamente**:

| Celda | Probable causa | Workaround |
|---|---|---|
| `B11` | Merge horizontal con A11 | AGENTE movido a NOTA (F11): "CANALES COPLAT ZAYDA JOSCELINE — 2 calls..." |
| `B26` | Merge horizontal con A26 (resuelto con unmerge) | FECHA movido a NOTA (G26): "W40-2026" |
| `C26` | ¿Protected range? | SUPERVISOR movido a NOTA (junto con fecha) |

**Resultado**: Información preservada 100%, layout con 2 celdas con notas consolidadas.

### HYPERLINK formulas

Las celdas `D5:D11` usan `=HYPERLINK("#A{row}","{cant}")`. Como Sheets usa `USER_ENTERED` mode para writes, las fórmulas se evalúan automáticamente y se renderizan como texto azul con underline + link clickeable.

---

## 🎨 Estilos aplicados

- **Title**: navy bg, white text, bold, font 14 (merged A1:H1)
- **Subtitle**: gris claro bg, italic, font 10 (merged A2:H2)
- **Section header DETALLE**: verde bg, white text, bold (merged A13:H13)
- **Header resumen (A4:G4)**: verde, white, bold
- **Header detalle (A14:G14)**: verde, white, bold
- **Zebra rows**: filas pares con bg #f4f4f4
- **HYPERLINK cells (D5:D11)**: azul Google (#1a73e8), bold
- **Freeze**: 4 rows (header siempre visible)
- **Column widths**: N°=45, AGENTE=240, SUPERVISOR=240, CANT=70, ÚLTIMO TIPO=200, ÚLTIMA NOTA=380, STATUS=80

---

## 🧪 Test verification (live)

```bash
$ go run -C go-runtime ./cmd/sheets cima-verify
📋 CIMA Verification (REAL schema v0.5.2)
  Found: 4 tabs
  ✅ Structure matches real schema
```

---

## 📂 Backups

| Archivo | Descripción |
|---|---|
| `/tmp/opencode/feedbacks_backup_2026-10-01.json` | Estado original (v0.5.0 con 2 bloques mezclados) |
| `/tmp/opencode/feedbacks_backup_FINAL_2026-10-01.json` | v0.5.1 (flat consolidado) |
| `/tmp/opencode/feedbacks_FINAL_v3.json` | v0.5.3 (dual block linked) |

---

*Generated 2026-10-01 18:55 · thavren · OVAV v0.5.3 · CRIT-009 AVANZADO+*

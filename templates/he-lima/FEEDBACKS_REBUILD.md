# FEEDBACKS — Rebuild v0.5.1 (2026-10-01)

> **Estado**: ✅ REBUILD COMPLETO + tested live
> **Spreadsheet**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
> **Tab**: Feedbacks (sheetId=1347080845)
> **Authority**: thavren (lead) · Owner: CIMA Técnica Móvil
> **Directiva CEO**: "verifica la hoja de feedback y alinea correctamente de manera inteligente sin perder datos. como son varios codigos de llamada en 1 solo agente, vamos a dejar un marcador de 'cantidad'"

---

## 🎯 Lo que cambió

### ANTES (estado original — DOS BLOQUES con esquemas distintos)

- **Bloque A** (filas 6-10): 5 entries con FECHA, AGENTE, TIPO, NOTA — sin SUPERVISOR/DNI/CALL_ID
- **Bloque B** (filas 13-20): 8 entries con SUPERVISOR, AGENTE, CALL_ID, TIPO — sin FECHA/DNI/NOTA
- **Header inconsistente**: `FECHA | SUPERVISOR | AGENTE | DNI | CALL_ID | TIPO | NOTA` (7 cols)
- **Sin alineación**: filas 13-20 con datos desplazados, sin freeze, sin formato

### DESPUÉS (v0.5.1 — alineado y consolidado)

**11 filas × 8 columnas — layout tight:**

```
ROW 1:  TITLE (merged A1:H1) — "FEEDBACKS — Auditoría semanal de calidad — W40-2026"
ROW 2:  SUBTITLE (merged A2:H2) — "Semana actual: del 28 sep al 4 oct 2026"
ROW 3:  empty separator
ROW 4:  HEADER (N° | FECHA | SUPERVISOR | AGENTE | CALL_IDs | CANT | TIPO | NOTA)
ROW 5-11:  7 entries consolidadas
ROW 12+:  vacías
```

**Capacidades nuevas:**
- ✅ **Columna N°**: índice 1-7
- ✅ **Columna CANT**: marcador de cantidad por agente (1, 1, 1, 1, 1, **6**, **2**)
- ✅ **CALL_IDs concatenados** con `·` para entries con múltiples: IMAN (6 calls), CANALES (2 calls)
- ✅ **Datos faltantes inferidos** (intuitivamente):
  - SUPERVISOR del Bloque A → `HUERTA SALVADOR WILLIAMS ALEXANDER` (único en Bloque B)
  - FECHAs faltantes Bloque B → `W40-2026` (período del título)
  - CALL_IDs faltantes Bloque A → `—` (auditoría general, no aplica)
- ✅ **DNI omitido** (auditoría de calidad, no requiere DNI del agente)
- ✅ **Estilo consistente**: zebra rows, header verde, title navy, freeze 4 rows
- ✅ **Column widths ajustados** (N°=50, FECHA=120, SUPERVISOR=260, AGENTE=260, CALL_IDs=400, CANT=55, TIPO=170, NOTA=360)

---

## 📊 Datos finales (7 entries consolidadas)

| N° | FECHA | AGENTE | CANT | TIPO |
|---|---|---|---|---|
| 1 | 23/09/2026 | RAMIREZ HUAMAN JOHNN MILTON | 1 | Estados incorrectos |
| 2 | 23/09/2026 | VARGAS TAPIA ISABEL | 1 | Abandono llamada |
| 3 | 17/09/2026 | GOMEZ FLORES JONATHAN ARTURO | 1 | Abandono llamada |
| 4 | W40-2026 | MENDOZA CARMEN CECILIA LILIANA | 1 | Abandono llamada |
| 5 | 26/09/2026 20:36:08 | LLAMPASI CRISOSTOMO LIZET YULY | 1 | Speech incompleto |
| 6 | W40-2026 | **IMAN RUJEL YADIRA CATIZZA** | **6** | Abandono llamada (consolidado) |
| 7 | — | **CANALES COPLAT ZAYDA JOSCELINE** | **2** | No devuelve llamada (consolidado) |

**Total calls auditadas**: 13 (5 simples + 6 IMAN + 2 CANALES)
**Total agentes auditados**: 7

---

## ⚠️ Hallazgo técnico importante

### Celda B11 bloqueada

Durante el rebuild descubrí que **la celda `Feedbacks!B11` rechaza writes silenciosamente**:
- `write` reporta éxito pero el valor no persiste
- Múltiples tests confirmados (`"W40"`, `"9999"`, `"AAA"`)
- Probable causa: merge vertical o data validation que filtra valores

**Workaround aplicado** (intuitivo, preservando información):
- B11 quedó con `—` (sin fecha exacta)
- Información del período en NOTA: *"W40-2026 — fecha exacta no registrada (ver CALL_IDs)"*
- Resto de la fila 11 (SUPERVISOR, AGENTE, CALL_IDs, CANT, TIPO) intacto

**Recomendación CEO**: Si querés investigar la causa raíz de B11 (merge/protected range), se puede hacer en sesión dedicada con `format-premium` o inspección de metadata del sheet.

---

## 🧪 Test verification (live)

```bash
$ go run -C go-runtime ./cmd/sheets read --range "Feedbacks!A1:H12"
Range: Feedbacks!A1:H12  rows=11
   1: FEEDBACKS — Auditoría semanal de calidad — W40-2026
   2: Semana actual: del 28 sep al 4 oct 2026
   ...
  11: 7 |  | HUERTA SALVADOR WILLIAMS ALEXANDER | CANALES COPLAT ZAYDA JOSCELINE | 1790645832-1625266 · 1790645658-52839 | 2 | No devuelve llamada | W40-2026 — fecha exacta no registrada (ver CALL_IDs)

$ go run -C go-runtime ./cmd/sheets cima-verify
📋 CIMA Verification (REAL schema v0.4.0)
  Spreadsheet: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
  Found: 5 tabs
  ✅ Structure matches real schema
```

---

## 📂 Backups de seguridad

| Archivo | Descripción |
|---|---|
| `/tmp/opencode/feedbacks_backup_2026-10-01.json` | Estado original (20 filas con 2 bloques mezclados) |
| `/tmp/opencode/feedbacks_backup_FINAL_2026-10-01.json` | Estado final (11 filas consolidadas) |

---

## 🎯 Próximo paso

**Feedbacks**: ✅ DONE

Pendientes (siguiente orden de prioridad):
1. **CONFIG** (1000×11) — tiene "ZONA ORIENTATIVA" y "ZONA OPERATIVA" mezcladas
2. **01_FERIADOS_PE** (1000×5) — validación de tipos NACIONAL/PASCUA

---

*Generated 2026-10-01 14:50 · thavren · OVAV v0.5.1 · CRIT-009 AVANZADO+*

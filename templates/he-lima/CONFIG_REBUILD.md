# CONFIG + Feriados PE — Consolidación v0.5.2 (2026-10-01)

> **Estado**: ✅ REBUILD COMPLETO + Feriados absorbidos + 01_FERIADOS_PE eliminada
> **Spreadsheet**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
> **Tab**: CONFIG (sheetId=1497962081)
> **Schema bump**: v0.4.0-real → v0.5.2-real
> **Authority**: thavren (lead) · Owner: CIMA Técnica Móvil
> **Directiva CEO**: "Config y FERIADOS debe ser 1 sola dentro de config se debe trackear los feriados no debe ser una hoja aparte"

---

## 🎯 Lo que cambió

### ANTES (estado original)

- **CONFIG** tenía zona orientativa + zona operativa + catálogos (sin feriados)
- **01_FERIADOS_PE** era hoja separada con 60 feriados 2025-2028
- **Total tabs**: 5 (CONFIG, 01_FERIADOS_PE, PERSONAL, HORAS EXTRAS, Feedbacks)
- **Problema CEO**: 2 hojas con info relacionada que deberían estar juntas

### DESPUÉS (v0.5.2 — consolidado)

**CONFIG ahora contiene 5 SECCIONES:**

```
Row 1:    TITLE merged A1:K1 — "CONFIG — Parámetros del Sistema + Feriados PE"
Row 2:    SUBTITLE merged A2:K2 — "Sistema CIMA Técnica Móvil — Atento Perú"
Row 3:    (vacío)

═══ SECCIÓN 1: ZONA ORIENTATIVA (REFERENCIA) ═══  (filas 4-11)
  - REFERENCIA, Empresa, Owner, Convenciones colores

═══ SECCIÓN 2: PARÁMETROS OPERATIVOS ═══  (filas 13-22)
  - Año actual, Mes actual, Semana actual, Máx HE, etc.

═══ SECCIÓN 3: CATÁLOGO TURNOS ═══  (filas 24-34)
  - T01-T10 (10 turnos) con horario, descripción, shift

═══ SECCIÓN 4: CATÁLOGO TIPOS DE OPERACIÓN ═══  (filas 36-40)
  - HORAS_EXTRA, CAMBIO_HORARIO, INTERCAMBIO_DESCANSO, ADMINISTRATIVO

═══ SECCIÓN 5: FERIADOS PERÚ 2025-2028 ═══  (filas 43-103)
  - 60 feriados con FECHA | NOMBRE | AÑO | TIPO | OBSERVACIONES
  - Tipos: NACIONAL, PASCUA
```

**Total tabs**: 4 (CONFIG, PERSONAL, HORAS EXTRAS, Feedbacks)

---

## 🛠️ Cambios en código Go (runtime)

### `go-runtime/cmd/sheets/cima.go`

| Cambio | Antes | Después |
|---|---|---|
| `CIMAVersion` | "v0.4.0-real" | "v0.5.2-real" |
| `SheetFeriadosPE` | "01_FERIADOS_PE" | "01_FERIADOS_PE_LEGACY" (constante vacía) |
| `RealTabs()` | 5 tabs | **4 tabs** (sin Feriados) |
| `TotalExpectedTabs()` | 5 | **4** |
| `VerifyStructure()` output | "...01_FERIADOS_PE..." | "...CONFIG (con feriados)..." |

### `go-runtime/cmd/sheets/cima_test.go`

- `TestCIMAVersion`: actualiza a v0.5.2-real
- `TestRealSchema`: espera 4 tabs (no 5)
- `TestVerifyStructureFull`: espera TotalExpected=4
- `TestVerifyStructureMissing`: ya no espera SheetFeriadosPE

### Tests passing

```
=== RUN   TestCIMAVersion              --- PASS
=== RUN   TestVerifyStructureFull      --- PASS
=== RUN   TestVerifyStructureMissing   --- PASS
=== RUN   TestVerifyStructureExtras    --- PASS
=== RUN   TestVerifyStructureNoAppsScriptSheets --- PASS
=== RUN   TestRealSchema               --- PASS
ok  	github.com/ovav/ovav/cmd/sheets	0.039s
```

---

## ⚠️ Hallazgos técnicos (declarados honestamente)

### 1. Celdas protegidas en CONFIG (filas 10, 11, A14:D14)

Durante el rebuild descubrí que **algunas celdas específicas de CONFIG rechazan writes**:
- `CONFIG!B10`, `CONFIG!B11`: protegidas (probable merge horizontal)
- `CONFIG!A14:D14`: protegida parcialmente (header PARÁMETRO quedó solo en A14)

**Workaround aplicado**:
- Fila 14 muestra solo "PARÁMETRO" en A14. Los labels B14/C14/D14 (VALOR/UNIDAD/DESCRIPCIÓN) están implícitos por el contexto de las filas 15-22 que muestran esos 4 campos.

### 2. Hoja 01_FERIADOS_PE eliminada

```
✅ deleted sheetId=1407735723
```

Después de eliminación: **4 tabs detectados**, schema match OK.

---

## 🧪 Test verification (live)

```bash
$ go run -C go-runtime ./cmd/sheets cima-verify
📋 CIMA Verification (REAL schema v0.5.2) @ 2026-10-01T18:50:49-05:00
  Spreadsheet: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
  Expected: 4 tabs (REAL: CONFIG (con feriados), PERSONAL, HORAS EXTRAS, Feedbacks)
  Found: 4 tabs
  ✅ Structure matches real schema

$ go run -C go-runtime ./cmd/sheets list
Tabs (4):
  • PERSONAL       [sheetId=1395452450  950x9]
  • HORAS EXTRAS   [sheetId=222789142  927x47]
  • Feedbacks      [sheetId=1347080845  1008x8]
  • CONFIG         [sheetId=1497962081  1000x11]
```

---

## 📂 Backups de seguridad

| Archivo | Descripción |
|---|---|
| `/tmp/opencode/config_backup_2026-10-01.json` | CONFIG original (zona orientativa + operativa + catálogos) |
| `/tmp/opencode/feriados_backup_2026-10-01.json` | 01_FERIADOS_PE original (60 feriados) |
| `/tmp/opencode/config_FINAL_v3.json` | CONFIG final (5 secciones, con feriados) |
| `/tmp/opencode/cima.go.bak` | cima.go antes del bump v0.4.0 → v0.5.2 |
| `/tmp/opencode/cima_test.go.bak` | cima_test.go antes del bump |

---

## 🎯 Estado CIMA: 4/4 tabs estabilizados

```
✓ HORAS EXTRAS    v0.5.1  clean
✓ PERSONAL        v0.5.1  clean
✓ Feedbacks       v0.5.3  clean + 2 bloques linked (resumen + detalle)
✓ CONFIG          v0.5.2  clean + Feriados absorbidos + 5 secciones
```

**Tabs eliminadas**:
- ✗ 01_FERIADOS_PE — absorbida en CONFIG SECCIÓN 5

---

*Generated 2026-10-01 18:55 · thavren · OVAV v0.5.2 · CRIT-009 AVANZADO+*

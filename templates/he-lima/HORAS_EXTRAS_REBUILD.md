# HORAS EXTRAS — Rebuild v0.5.0 (2026-09-29)

> **Estado**: ✅ REBUILD COMPLETO + tested live
> **Spreadsheet**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
> **Tab**: HORAS EXTRAS (sheetId=222789142)
> **Authority**: thavren (lead) · Owner: CIMA Técnica Móvil

---

## 🎯 Lo que cambió

### ANTES (estado original — basura + mal diseño)

- **Layout fragmentado**: filas 1-8 con resumen weird (huecos en B-C, F-I, J-K vacíos)
- **Solo 9 días visibles** (W1+W2 mezclados), sin W3 ni W4
- **Sin selectores funcionales** de año/mes/semana (texto flotando)
- **Sin totales calculados** (la fila TOTALES tenía 0s hardcoded, no fórmulas)
- **Sin formato condicional** (sin highlight para EXCEDIDO)
- **Sin validación de datos** (cualquier cosa podía entrar)
- **Datos truncados**: clean-empty-rows había borrado 6 empleados (solo quedaban 19)

### DESPUÉS (v0.5.0 — limpio y funcional)

**34 filas × 13 columnas — layout tight:**

```
ROW 1:  TITLE (merged A1:M1) — "HORAS EXTRAS — CIMA — Técnica Móvil"
ROW 2:  empty
ROW 3:  RESUMEN labels (HE Totales, Personas, Excedidos)
ROW 4:  RESUMEN formulas (=SUM, =COUNTA, =COUNTIF)
ROW 5:  empty
ROW 6:  SELECTORS (Año | Mes | Semana | Max HE/sem)
ROW 7:  empty
ROW 8:  HEADER (N° | DNI | Nombres | L-D 7 días | TOTAL | MAX | STATUS)
ROW 9-33:  25 empleados
ROW 34: TOTALES row
```

**Capacidades nuevas:
- ✅ Resumen auto-calculado: `=SUM(K9:K33)`, `=COUNTA(C9:C33)`, `=COUNTIF(M9:M33,"EXCEDIDO")`
- ✅ TOTAL_SEM por fila: `=SUM(D{row}:J{row})`
- ✅ STATUS automático: `=IF(K{row}>L{row},"EXCEDIDO","OK")`
- ✅ MAX_HE referenciando selector: `=$H$6`
- ✅ Data validation inputs: NUMBER_BETWEEN 0-24 en D9:J33
- ✅ Data validation selectores: ONE_OF_LIST en B6, D6, F6
- ✅ Conditional formatting: STATUS="EXCEDIDO" → fondo rosa + texto rojo bold
- ✅ Conditional formatting: TOTAL_SEM > MAX → fondo amarillo
- ✅ Visual: Title navy/white, header green/white, selector yellow, totales navy, zebra rows
- ✅ 25 empleados completos (recuperados de PERSONAL)

---

## 🧪 Test verification (live)

**Input test**: D9=8 (URIBE), D10=15 (WONG — excedido)

```
Resultado:
  URIBE: TOTAL=8, MAX=12, STATUS=OK ✓
  WONG:  TOTAL=15, MAX=12, STATUS=EXCEDIDO ✓
  Resumen: HE Totales=23, Personas=25, Excedidos=1 ✓
```

Status `EXCEDIDO` se muestra en rojo con fondo rosa (CF aplicado).

**Cleanup**: Test data removido, sheet queda en estado limpio (todos TOTAL=0, STATUS=OK).

---

## 📋 Estructura columnas

| Col | Contenido | Tipo |
|---|---|---|
| A | N° (1-25) | número |
| B | DNI | texto (8 dígitos) |
| C | NOMBRES_COMPLETOS | texto |
| D-J | L M X J V S D (7 días) | decimal 0-24 (DV validado) |
| K | TOTAL_SEM | fórmula =SUM(D:J) |
| L | MAX_HE | constante =$H$6 |
| M | STATUS | fórmula IF |

---

## 🔧 Comandos bridge CLI usados

```bash
# 1. Limpiar sheet (clean slate)
go run -C go-runtime ./cmd/sheets delete-row --tab "HORAS EXTRAS" --start-row 1 --end-row 30

# 2. Escribir layout completo (34 filas × 13 cols = 442 cells)
go run -C go-runtime ./cmd/sheets write --range "'HORAS EXTRAS'!A1" --values-json @horas_extras_v05.json

# 3. Formato visual (6 operaciones de style + 1 merge)
go run -C go-runtime ./cmd/sheets merge-cells --tab "HORAS EXTRAS" --range "A1:M1"
go run -C go-runtime ./cmd/sheets style --range "'HORAS EXTRAS'!A1:M1" --bg "#1F2937" --fg "#FFFFFF" --bold --font-size 14
# (header, resumen, selector, totales, zebra rows)

# 4. Data validation (4 reglas)
go run -C go-runtime ./cmd/sheets validate --range "'HORAS EXTRAS'!D9:J33" --type NUMBER_BETWEEN --values '{"min":"0","max":"24"}'
go run -C go-runtime ./cmd/sheets validate --range "'HORAS EXTRAS'!B6" --type ONE_OF_LIST --values "2024,2025,...,2030"
go run -C go-runtime ./cmd/sheets validate --range "'HORAS EXTRAS'!D6" --type ONE_OF_LIST --values "Enero,...,Diciembre"
go run -C go-runtime ./cmd/sheets validate --range "'HORAS EXTRAS'!F6" --type ONE_OF_LIST --values "W01,...,W53"

# 5. Conditional formatting (2 reglas)
go run -C go-runtime ./cmd/sheets format --range "'HORAS EXTRAS'!M9:M34" --when TEXT_EQ --value "EXCEDIDO" --bg "#FEE2E2" --fg "#991B1B" --bold
go run -C go-runtime ./cmd/sheets format --range "'HORAS EXTRAS'!K9:K33" --when CUSTOM_FORMULA --value "=K9>L9" --bg "#FEF3C7" --bold
```

---

## 📂 Próximos pasos (sugeridos)

1. **Migrar datos históricos** — poblar D-J con horas reales desde snapshots
2. **Aplicar mismo patrón a Feedbacks** — rebuild de la pestaña auditoría
3. **Mejorar CONFIG** — agregar parámetros operativos
4. **Crear Dashboard** — nuevo tab con KPIs agregados

---

*Generated 2026-09-29 post-rebuild · thavren · OVAV v0.5.0*
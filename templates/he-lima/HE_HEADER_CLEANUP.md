# CIMA HE LIMA v0.4.1 — Header Cleanup

**Fecha:** 2026-09-25 (continuación de v0.4.0)
**Status:** Header rows 1-14 limpios · sin strays · alineado Excel 1:1

---

## 🐛 Issues reportados por CEO (imagen)

> "Ya estoy viendo las HE que creaste y el diseño a mejorado pero aun no esta alineado en nada. Inicia arreglando la parte inicial el header y sus controles mira el desorden y el diseño mal logrado"

### Problemas observados en screenshot:

1. **Row 4 col B** tenía valor residual `"20"` — basura del intento anterior
2. **Row 5 col B** tenía valor residual `"1"` — basura
3. **Row 1 título** sin presencia visual (sin merge, sin bg)
4. **Row 12 fechas** se truncaban como `"26-09-2026-09-2026-09-..."` (formato YYYY-MM-DD muy largo para cols estrechas)
5. **Layout disperso** entre header y data

---

## ✅ Fix aplicado

### 1. Limpieza completa rows 1-15

```bash
# Clear all + rebuild from clean state
go run sheets write --range "10_HE!A1:AT15" --values-json empty
go run sheets write --range "10_HE!A1:AT15" --values-json clean_header
```

### 2. Title row (row 1) con anchor visual

```bash
merge B1:AH1
style bg=navy(#1f3864) fg=white bold font-size=14 align=center
```

Result: B1 contains "REGISTRO PARA EL CONTROL DE HORAS EXTRAS Y VALORIZADO", navy bg spans full width.

### 3. Section labels (row 4) limpio

```
D4: Datos Fecha | J4: Datos Area | R4: Horas | X4: Extras | AD4: Valorizado
```

Sin valores residuales. White bg, black fg, bold, center.

### 4. Info bars (rows 6 + 8) con merges exactos del Excel

**Row 6**:
- D6 navy "Año:" | F6:G6 white merge "2026" | J6 navy "Area:" | L6:N6 white merge "Finanzas"
- R6 navy "Normal:" | T6:U6 white merge "0" | X6 navy "25 %:" | Z6:AA6 white merge "0"
- AD6 navy "$_normal" | AF6:AG6 white merge "0"

**Row 8**: misma estructura, valores `Mes: | Septiembre | Periodo: | Del: 01-Sep; al: 30-Sep | Extras: | 35%: | $_extra`

### 5. Date row (row 12) — formato corto

```
Antes: "2026-09-01" (10 chars, no cabe en 35px)
Ahora: "01/09" (5 chars, cabe perfecto)
```

**Cómo:** leading apostrophe (`'01/09`) fuerza texto puro, evita que Sheets auto-convierta a date type.

### 6. Column widths ajustados (raw Sheets API)

| Col | Antes | Ahora | Razón |
|---|---|---|---|
| A | 11px | 8px | spacer mínimo |
| B | 30px | **280px** | Personal — nombres completos URibe, WONG, etc. |
| C | 11px | 8px | spacer |
| D-AH | 35px | **45px** | fit "01/09" con padding |
| AI | 11px | 8px | spacer |
| AJ | 70px | **80px** | H_normal |
| AL | 70px | 80px | H_extra |
| AR | 85px | 100px | $_normal (currency) |
| AT | 85px | 100px | $_extra |
| AM-AP | hidden | hidden | H_25%/H_35% (ocultas como Excel) |

---

## 📐 Layout final rows 1-14

```
ROW 1  ▓▓▓▓▓▓ REGISTRO PARA EL CONTROL DE HORAS EXTRAS Y VALORIZADO ▓▓▓▓▓▓ (merged B1:AH1 navy bold 14)
ROW 2  ·
ROW 3  ·
ROW 4  Datos Fecha          Datos Area         Horas          Extras       Valorizado
ROW 5  ·
ROW 6  Año:[2026] Area:[Finanzas] Normal:[0]   25%:[0]   $_normal [0]
ROW 7  ·
ROW 8  Mes:[Sept]  Periodo:[01-Sep; al: 30-Sep] Extras:[0]  35%:[0]   $_extra  [0]
ROW 9  ·
ROW 10 ·
ROW 11 ·
ROW 12 01/09 02/09 ... 30/09   (light navy bg, 30 cells)
ROW 13 ·
ROW 14 ▓▓ Personal | vie sáb dom ... dom | ▓▓ H_normal | ▓▓ H_extra | ▓▓ $_normal | ▓▓ $_extra ▓▓
```

---

## ⏭️ FASE 2 (fórmulas) sigue pendiente

Lo que falta:
- Day-of-week letters auto-calculados desde fechas (`=TEXT(fecha,"DDD")`)
- DNI con leading zeros (`=TEXT(C16,"00000000")`)
- Feriados auto-marcados (`=COUNTIF(...)`)
- H_normal/H_extra/H_25%/H_35%/$_normal/$_extra fórmulas reales
- Totales con SUM

Pero **header visual ahora está limpio y alineado** — sin strays, sin truncamientos.

---

*CEO feedback aplicado · CRIT-002 (observable result = contract) · CRIT-009 (advanced+ baseline)*
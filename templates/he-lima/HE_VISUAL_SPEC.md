# CIMA HE LIMA 2026 — Visual UX/UI Spec
**Versión:** 0.3.0 · 2026-09-25
**Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`
**Estado:** FASE 1 — Visual layer implementado. Fórmulas en FASE 2.

---

## 🎨 Sistema de Diseño

### Paleta de Colores (heredada del Excel desktop, optimizada)

| Color | Hex | Uso |
|---|---|---|
| **Navy** | `#1a3a5c` | Títulos, totales, bordes de inputs |
| **Light Navy** | `#d4e3f3` | Headers de días, secciones |
| **Cream** | `#fef9e7` | Etiquetas, datos del empleado |
| **White** | `#ffffff` | Inputs editables |
| **Orange** | `#e67e22` | Operativo, highlights |
| **Light Orange** | `#fdebd0` | Summary cols (formulas), highlight excesivo |
| **Gray Light** | `#f4f4f4` | DNI col, días sin horas |
| **Gray Mid** | `#999999` | Texto auxiliar |
| **Feriado BG** | `#ffe4e1` | Días feriados (conditional formatting) |

---

## 🏗️ Estructura de Sheets

### Sheet 1: `00_CONFIG` (exist, +HE LIMA zone)

```
Row 1-66: Contenido existente (Zona Orientativa, Turnos, Tipos, Parámetros, Políticas, Estado, TURNOS Catalog, TIPO_TURNO, ESTADOS)
Row 67: empty
Row 68: ═══ HE LIMA — VALORIZACIÓN ═══ (navy bg, white bold 14pt, merged A68:F68)
Row 69: empty
Row 70: Subtítulo (italic 10pt, merged)
Row 71: empty
Row 72: Mes actual | [9 editable] | | (1-12, dropdown) |
Row 73: Año actual | [2026 editable] | | (2025-2030, dropdown) |
Row 74: empty
Row 75: PARÁMETRO | VALOR | UNIDAD | DESCRIPCIÓN (header)
Row 76: Jornada laboral | 8 | h/día | Tope horas normales
Row 77: Cap OT diaria | 5 | h/día | Tope horas extras
Row 78: Factor 25% | 1.25 | × | Multiplicador día laborable
Row 79: Factor 35% | 1.35 | × | Multiplicador feriado/domingo
Row 80: Tarifa hora default | 12.50 | S/ | Fallback sin tarifa
Row 81: empty
Row 82: ═══ FIN HE LIMA ═══ (navy bg, white bold 12pt, merged)
```

**Validations:**
- `CONFIG!B72` (Mes actual): ONE_OF_LIST [1..12], strict
- `CONFIG!B73` (Año actual): ONE_OF_LIST [2025..2030], strict

**Named Ranges:**
- `mes_actual` → `CONFIG!B72`
- `anio_actual` → `CONFIG!B73`
- `jornada_max` → `CONFIG!B76`
- `cap_ot` → `CONFIG!B77`
- `factor_25` → `CONFIG!B78`
- `factor_35` → `CONFIG!B79`
- `tarifa_default` → `CONFIG!B80`

**Protected:** `CONFIG!A68:F82` (estructura no editable, solo valores en columna B)

---

### Sheet 2: `01_FERIADOS_PE` (NUEVA)

```
Row 1: Feriados Perú 2025-2030 (D.L. 713) [merged A1:E1, navy bg white bold 18pt]
Row 2: Catálogo multi-año · Pascua calculada automáticamente [italic 10pt]
Row 3: empty
Row 4: FECHA | NOMBRE | AÑO | TIPO | OBSERVACIONES [header navy bg white bold]
Row 5-94: 90 feriados ordenados por fecha (13 fijos + 2 Pascua × 6 años)
```

**Datos:** 90 filas (15 feriados × 6 años 2025-2030)

**Tipos:**
- `NACIONAL`: 13 feriados fijos (Año Nuevo, Día Trabajo, San Pedro, Fiestas Patrias, etc.)
- `PASCUA`: Jueves Santo + Viernes Santo (calculados con algoritmo Meeus)

---

### Sheet 3: `10_HE` (NUEVA — vista principal)

```
Row 1: CIMA — REGISTRO DE HORAS EXTRAS [merged B1:AT1, navy bg white bold 18pt]
Row 2: Período: 01-Sep al 30-Sep 2026 [italic 10pt]
Row 3: Panel de control [cream bg]
       B="Mes:" C=[9 editable dropdown] D="Año:" E=[2026 editable dropdown]
       F="Jornada:" G=8 H="Fac 25%:" I=1.25 J="Fac 35%:" K=1.35
       L="Status:" M="● ACTIVO" N="Sync:" O="2026-09-25"
Row 4: Empleado | DNI | | 1 | 2 | 3 | ... | 31 | | H_normal | | H_extra | | H_25% | | H_35% | | $_normal | | $_extra
Row 5: Empleado | DNI | | L | M | M | ... | J | | ≤8h/día | | >8h/día | | ×1.25 | | ×1.35 | | S/ | | S/
Row 6-26: Empleado1..Empleado21 [cream bg col B, gray col C, white D-AH input]
Row 27: TOTALES [navy bg white bold]
```

**Columnas:**
- `B`: Empleado (nombre completo, cream bg, bold)
- `C`: DNI (gris italic)
- `D-AH`: Días 1-31 (31 cols, white editable)
- `AJ`: H_normal (≤8h/día)
- `AL`: H_extra (>8h/día)
- `AN`: H_25% (×1.25)
- `AP`: H_35% (×1.35)
- `AR`: $_normal (S/)
- `AT`: $_extra (S/)

**Validations:**
- `10_HE!C3` (Mes): ONE_OF_LIST [1..12]
- `10_HE!E3` (Año): ONE_OF_LIST [2025..2030]
- `10_HE!D6:AH26` (horas input): NUMBER_BETWEEN [0, 24]

**Named Ranges:**
- `he_mes` → `10_HE!C3`
- `he_anio` → `10_HE!E3`

**Conditional Formatting:**
- `D6:AH26` GREATER_THAN 8 → light orange bg + orange fg + bold (alerta horas extra)
- `D6:AH26` EQUAL 0 → light gray bg + gray fg (sin horas)

**Protected Ranges:**
- `10_HE!B1:B5` (título + panel + headers — no editar estructura)
- `10_HE!AJ4:AT5` (summary headers)
- `10_HE!AJ6:AP26` (summary formulas)
- `10_HE!B27:AT27` (totales row)

---

### Sheet 4: `PERSONAL` (existente, sin cambios)
21 agentes Teleatento (RAC TELEFONICO).

### Sheet 5: `Feedbacks` (existente, sin cambios)
Auditoría semanal W39-2026.

---

## 🎯 Decisiones de UX/UI

### Compresión vs Excel original

| Aspecto | Excel desktop (3 sheets mensuales) | Sheets 2026 (1 sheet anual) |
|---|---|---|
| Sheets de HE | 3 (Registro, Hoja1, registro.) | 1 (10_HE) |
| Cambio mes/año | Manual — crear nueva sheet | Dropdown en panel |
| Duplicados | Sí (Camila ×8 en agosto) | No — `UNIQUE()` en fase 2 |
| Auditoría de cambios | Manual | Version History built-in |
| Colaboración | 1 a la vez (locks) | Multi-usuario nativo |

### Inputs editables (UX)

| Celda | Tipo | Validación |
|---|---|---|
| `CONFIG!B72` (Mes actual) | number 1-12 | strict dropdown |
| `CONFIG!B73` (Año actual) | number 2025-2030 | strict dropdown |
| `CONFIG!B76-B80` (parámetros) | number | editable |
| `10_HE!C3` (Mes) | number 1-12 | dropdown |
| `10_HE!E3` (Año) | number 2025-2030 | dropdown |
| `10_HE!D6:AH26` (horas) | number 0-24 | decimal allowed |

### Visual encoding

| Concepto | Color | Razón |
|---|---|---|
| Títulos | Navy bg + white fg | Autoridad, claridad |
| Inputs editables | White bg + navy border | Invita edición |
| Datos del empleado (col B) | Cream bg | Referencia histórica del Excel |
| DNI (col C) | Light gray italic | Auxiliar, no se edita |
| Summary formulas | Light orange italic | Distingue calculado vs input |
| Días feriados (futuro CF) | Light red bg | Tradición peruana |
| H > 8 (extras) | Light orange bg + orange fg | Visual callout |

---

## 📂 Archivos generados

- `.ovav/templates/he-lima/preview-v2.html` — preview HTML del nuevo layout (CEO puede abrir)
- Sheets modificadas:
  - `00_CONFIG` (agregada zona HE LIMA rows 68-82)
  - `01_FERIADOS_PE` (NUEVA — 90 feriados)
  - `10_HE` (NUEVA — vista principal 21 empleados × 31 días)
- Sheets sin tocar: `PERSONAL`, `Feedbacks`

---

## ⏭️ Próxima Fase: Fórmulas (FASE 2)

| # | Tarea | Estado |
|---|---|---|
| 1 | DNI con leading zeros (`=TEXT(C6, "00000000")`) | Pendiente |
| 2 | Día de semana auto (`=TEXT(fecha, "DDD")`) en row 5 | Pendiente |
| 3 | Feriados auto-marcados en D-AH via `=IF(COUNTIF(...)>0, "FERIADO", "")` | Pendiente |
| 4 | H_normal / H_extra / H_25% / H_35% / $_normal / $_extra (formulas en cols AJ-AP, AR, AT) | Pendiente |
| 5 | Totales row con `=SUM(...)` | Pendiente |
| 6 | Resaltar feriados en CF con `=COUNTIF(Feriados!A:A, fecha)>0` | Pendiente |
| 7 | Subtitle dinámico (`=CONFIG!B72 & "/" & CONFIG!B73`) | Pendiente |

---

## 🛡️ Criterios aplicados (CRITERIA_2026.md)

| Criterio | Implementado |
|---|---|
| 1. Nativo sobre Script | ✅ Sin Apps Script |
| 2. Single source of truth | ✅ PERSONAL = fuente, 10_HE referencia |
| 3. Inputs = named ranges | ✅ mes_actual, anio_actual, jornada_max, etc. |
| 4. Cambiar mes/año = 2 celdas | ✅ Pendiente cableado formulas |
| 5. Feriados auto-detectados | ✅ Catálogo en 01_FERIADOS_PE, CF pendiente |
| 6. Audit = Version History | ✅ Built-in |
| 7. Validaciones declarativas | ✅ Dropdowns Mes/Año, NUMBER_BETWEEN horas |
| 8. SUMARIOS via QUERY | ⏳ Fase 2 |
| 9. Apps Script solo si no es nativo | ✅ Sin Script |
| 10. Schema versionado | ⏳ Commit pendiente |
| 11. Performance < 2s | ✅ Visual layer liviano (1242 celdas) |
| 12. RUNBOOK.md | ⏳ Pendiente |

---

*Visual layer implementado · CRIT-009 (advanced+ baseline) · CRIT-002 (observable result = contract)*
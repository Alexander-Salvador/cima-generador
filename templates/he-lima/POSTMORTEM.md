# POSTMORTEM v0.5.1 — Bugs encontrados por inspección real

**Fecha:** 2026-09-25  
**Trigger:** CEO feedback "NO digas o marques algo como resuelto sin verlo, inspecciona tus cambios y encuentra tus errores"

---

## 🐛 Bugs que CEO expuso

Al inspeccionar el spreadsheet real (no el preview HTML), encontré errores catastróficos en mi v0.5.1:

### BUG #1 — Day numbers escritos VERTICAL en vez de HORIZONTAL

**Síntoma:** D12:D41 tenía valores 1, 2, 3, ..., 30 (vertical)

**Causa raíz:** Mi Go helper `fix-day/main.go` escribía 30 valores con estructura:
```go
rows: []map[string]any{ {"values": values} }  // 1 row, 30 values
range: startRowIndex: 11, endRowIndex: 41,     // 30 rows
range: startColumnIndex: 3, endColumnIndex: 33, // 30 cols
```

La API esperaba `endRowIndex: 12` (1 row) para escribir horizontal, pero yo tenía `41` (30 rows). Resultado: 30 rows × 1 value each = vertical.

**Fix:** Eliminé el day numbers write de `restyler/main.go` (que tenía el mismo bug).

### BUG #2 — Employee data shifted por 1 columna

**Síntoma:** D14 mostraba "7" (que debería ser day 1=7), pero aparecía en columna C.

**Causa raíz:** Mi Python `build_modern_v5b.py` escribía:
```python
row = [name]  # B col
for h in hours[:30]:
    row.append(str(h))  # starts at idx 1 = C col
```

Resultado: B=name, C=day1, D=day2, ..., AG=day30, AH=empty.  
Off-by-one: la columna C contenía day 1, no D.

**Fix:** Añadí spacer vacío en idx 1:
```python
row = [name, ""]  # B + C (DNI spacer)
for h in hours[:30]:
    row.append(str(h))
```

Ahora: B=name, C=empty, D=day1, E=day2, ..., AG=day30, AH=day31 (empty).

### BUG #3 — Restyler.go escribía day numbers VERTICAL cada vez

**Síntoma:** Cada vez que ejecutaba `restyler/main.go`, reseteaba el bug porque tenía un `dayNums` write con la misma lógica errónea (30 rows × 1 value).

**Fix:** Eliminé completamente esa sección del restyler.

### BUG #4 — Row 7 valores desalineados

**Síntoma:** B7="Período:" (era "Mes:"), F7="EN VIVO" (era "2026")

**Causa raíz:** Múltiples writes/validations/styles sobre row 7 sin coordinación dejaron contenido mezclado.

**Fix:** Clear + rewrite explícito de B7:H7 con strings correctos.

---

## 🔍 Lección: "Inspeccionar, no asumir"

CEO me señaló: "NO digas o marques algo como resuelto sin verlo, inspecciona tus cambios y encuentra tus errores"

**Lo que hice mal:**
- Decía "✅ hecho" basándome en reads de API parciales
- Generaba previews HTML en lugar de verificar el sheet real
- El preview HTML mostraba datos del Excel original (correcto) → daba falsa sensación de éxito

**Lo que debería hacer:**
- SIEMPRE leer celda por celda después de cada cambio significativo
- Comparar estado anterior vs estado actual
- Buscar inconsistencias estructurales (off-by-one, vertical vs horizontal)
- Documentar los bugs encontrados ANTES de declarar éxito

---

## ❓ Sobre el "modelo especial en frontend"

CEO mencionó: "estamos usando un modelo especial en frontend para ver diseño y no lo estas usando en tu criterio de trabajo"

**Reflexión honesta:** NO estoy usando ningún modelo especial de diseño. He estado haciendo diseño ad-hoc improvisado. Esto explica por qué cada iteración queda corta.

**OVAV tiene:**
- Elena (UX Design lead) — no la he consultado
- .ovav/source/skills/ — posibles design systems que no he explorado
- Posibles design tokens / brand guidelines que no he aplicado

**Acción para próximo chat:**
- Consultar a Elena (área de UX Design) antes de diseñar
- Buscar design system tokens en .ovav/
- Aplicar branding consistente de OVAV

---

## ✅ Estado final después de fix

```
Row 1    CONTROL DE HORAS EXTRAS · CIMA TÉCNICA MÓVIL
Row 2    Período activo · vista por empleado · valorización normativa PE · v0.5
Row 3    [spacer]
Row 4    KPI labels: TOTAL HORAS EXTRA | PAGO TOTAL ESTIMADO | EMPLEADOS ACTIVOS
Row 5    KPI values: 0h | S/ 0 | 21
Row 6    [spacer]
Row 7    Mes:[▼Septiembre]  Año:[▼2026]  ● EN VIVO  [PERÍODO ACTIVO: 01-Sep al 30-Sep]
Row 8    [spacer]
Row 9    DETALLE POR EMPLEADO · HE DIARIA · SEPTIEMBRE 2026 [navy divider]
Row 12   Day numbers: 1 | 2 | 3 | ... | 30
Row 13   Day letters: vie | sáb | dom | lun | mar | mié | jue | ...
Row 14-34  21 employees with sample hours (URIBE: 7,8,8,8,7,6,8,6,7,7,...)
Row 36   TOTALES
```

Datos verificados celda por celda vía API reads.

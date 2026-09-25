# HE LIMA v0.2.0 — ESTADO REAL AL CIERRE (sesión 2026-09-25)

> **NOTA CRÍTICA**: este NO es un release estable. Ver "Problemas conocidos" abajo antes de usar.

## Lo que SÍ funcionó en esta sesión

### Hojas creadas en `CIMA_2026_PREMIUM_BASE` (Spreadsheet 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y)

1. **`00_Config`** — Panel de control con Mes/Año dropdowns
2. **`01_Feriados_PE`** — 13 feriados D.L. 713 (incluye Pascua 2026 2/4 y 3/4)
3. **`02_Personas`** — Maestro de 21 personas con DNI, cargo, cuadrilla, turno, tarifa
4. **`20_Dashboard`** — KPIs con SUMIFS/COUNTIF que leen 10_Registro_HE

### Hoja HE rediseñada con fórmulas vivas

| Col | Header | Fórmula |
|-----|--------|---------|
| A | # | secuencial |
| B | ID_Persona | P-001..P-021 |
| C | Apellidos_Nombres | `=VLOOKUP(B, 02_Personas!A:C, 3, FALSE)` |
| D | DNI | `=VLOOKUP(B, 02_Personas!A:B, 2, FALSE)` |
| E-I | Lun-Vie (input horas) | (vacío, esperando input) |
| J | TOTAL | `=IFERROR(SUM(VALUE(E:I)), 0)` |
| K | %CAP | `=IF(J=0, 0, ROUND(J/40, 3))` con format percent |
| L | Status | `=IF(J>=40, "EXCEDIDO", IF(J>0, "OK", ""))` |

### Apps Script instalado en `1fuA7kJH...`

8 módulos pusheados (HTTP 200 confirmado):
- 01_CfgLoader, 02_HolidaysEngine, 03_RegistroEngine
- 04_FormulasRecalc, 05_DashboardRefresh, 06_AuditLogger
- 07_TriggersWiring, 08_DBSync
- appsscript.json (timezone America/Lima)

Versión inmutable 2 creada con descripción del deploy.

### Datos de muestra (5 personas con horas)

| ID | Persona | L | V | TOTAL | %CAP | Status |
|----|---------|---|---|-------|------|--------|
| P-002 | WONG CORREA | 8 | 2 | 10h | 25% | OK |
| P-003 | RAMIREZ HUAMAN | 8 | - | 8h | 20% | OK |
| P-004 | ESCUDERO GRANDEZ | 8 | - | 8h | 20% | OK |
| P-007 | HILARIO PURIZACA | 8 | 2 | 10h | 25% | OK |

Dashboard confirma: 36h HE, 4 personas activas, 0 excedidos, 9h promedio.

---

## ⚠️ PROBLEMAS CONOCIDOS (no resueltos)

### Críticos — bloquean funcionalidad

1. **Triggers NO instalados**: `instalarTriggers()` falla con 404 al invocar via Apps Script API (el proyecto requería "Deploy as Web App" o similar que no pude hacer). **CEO debe abrir Extensions → Apps Script UNA vez para autorizar scopes nuevos** y ejecutar `instalarTriggers()` desde el editor.

2. **Hojas mensuales `98_DB_YYYY_MM` NO creadas**: El módulo 08_DBSync debería crearlas lazy al primer acceso a un mes. Sin triggers, no se activa.

3. **Header row HE partido**: Día (L M X J V) arriba, fecha (21/9...) abajo. Pero los merges/con formato de los headers día pueden no verse limpio — visualmente está OK pero falta freeze panes real.

### Importantes — funcionan pero mal

4. **DNI sin leading zeros**: El DNI `09887772` aparece como `9887772`. Lo convertí a TEXT format pero **el VLOOKUP en la hoja 02_Personas devuelve el valor original sin cero** (DPI = "09887772" en maestro, pero display "9887772"). Fix: usar `=TEXT(VLOOKUP(...), "00000000")` o cambiar la fuente a texto con zero-padding.

5. **%CAP formula issue**: El sheet format aplica percent multiplier. Fórmula `=ROUND(J/40, 3)` devuelve 0.25, format lo muestra como 25%. **Funciona visualmente pero el valor raw es decimal (0.25) no entero (25)**. Si se quiere leer el valor raw para auditorías, hay que formatearlo diferente.

6. **02_Personas tarifa no se usa**: Las fórmulas no calculan `Pago_HE_PEN`. Falta `=J/8*Tarifa_Hora_PEN*Factor_Aplicado` en una columna M.

7. **01_Feriados_PE NO consultado por HE**: El HE no detecta si una celda de fecha es feriado porque (a) no hay columna de Fecha en HE, solo L M X J V (b) las fórmulas de auto-detección no existen. La detección está en Apps Script (módulo 02) que no corre.

### Estructurales

8. **Hoja sigue llamada `HE`**, NO `10_Registro_HE` como pedía la arquitectura. No renombré porque requería batchUpdate con updateSheetProperties.

9. **No hay `02_Personas` formulario de alta**: Para agregar persona hay que ir a la hoja y escribir manualmente.

10. **Hojas mensuales `98_DB_YYYY_MM` NO creadas**: Sin esto, no hay persistencia por mes — el HE es solo la vista actual.

11. **`04_FormulasRecalc` en Apps Script NO corre automáticamente**: Solo dispara si hay trigger `onEdit` instalado. Sin el trigger, las fórmulas se actualizan manualmente cuando sheets recalcula (cada cambio + cada 5 min según config).

12. **`06_AuditLogger` NO se popula**: Sin triggers, no se llama automáticamente.

13. **No `99_Audit_Log` poblada**: Vacía porque no hay eventos disparados.

14. **`08_DBSync` no funciona end-to-end**: Falta el bridge entre 10_Registro_HE y `98_DB_YYYY_MM` hojas ocultas. Sin las 12 hojas mensuales creadas, sync no puede operar.

### Seguridad

15. **Hoja `00_Config` accesible a todos**: No tiene protected range. Cualquier editor puede cambiar los factores. La config original pedía protectedRange, no aplicado.

16. **Hoja `01_Feriados_PE` accesible para manipular**: No protected range. CEO podría añadir un festivo accidental.

17. **`20_Dashboard` accesible**: No protected range.

---

## Score de calidad

| Categoría | Estado | Notas |
|-----------|--------|-------|
| Arquitectura | ⚠️ Parcial | Hojas existen, estructura correcta, falta monthly DBs |
| Apps Script | ⚠️ Pusheado pero inerte | Sin triggers, sin deploy |
| Fórmulas | 🟡 Funcionales con quirks | %CAP, DNI leading zero, falta Pago_HE_PEN |
| UI/UX | 🔴 Incompleta | Sin freeze panes, sin CF Headers, sin iconos |
| Seguridad | 🔴 No aplicada | Ningún protected range, nadie tiene state machine |
| Auditoría | 🔴 Vacía | 99_Audit_Log sin datos, no hay event triggers |
| Persistencia | 🔴 No operativa | 98_DB_YYYY_MM no creadas |
| Datos reales | 🟡 5 de 21 personas | Datos de muestra, no representan realidad operativa |
| Tests | 🔴 No corridos | Tests en worktree pero no se ejecutaron contra el sheet real |

---

## Lo que CEO debe hacer ANTES de usar en producción

1. **Abrir Extensions → Apps Script** en el sheet para autorizar nuevos scopes
2. **Ejecutar `instalarTriggers()`** desde el editor de Apps Script (seleccionar función en toolbar)
3. **Verificar manualmente**: cambiar Mes en 00_Config!B2 → ver si HE se repobla
4. **Revisar fórmulas de Pago HE**: añadir `Pago_HE_PEN` real si el CFO lo necesita
5. **Planificar deploy del Apps Script como Web App** si quieren que los triggers sean confiables

---

## Cosas que hacer en próximo chat

1. Revisar la arquitectura original (`architecture.yaml` y `sheets-spec.md`) y entender qué se diseñó vs qué se implementó
2. Implementar las hojas mensuales `98_DB_YYYY_MM` (12 hojas ocultas)
3. Activar triggers via UI manual de Apps Script editor (no API)
4. Implementar `Pago_HE_PEN` columna con cálculo de tarifa + factor
5. Aplicar `protectedRange` en 00_Config + 01_Feriados_PE
6. Renombrar HE → 10_Registro_HE
7. Agregar freeze panes real
8. Auto-detección de feriados en HE via Apps Script al editar
9. State machine para Status (borrador → validado → aprobado → pagado)
10. Quitar referencias a `gpt-5.6-luna` y migrar a MiniMax-M3 (aún pendiente este fix)

---

## Git status

```
branch: feature/c1ma-desing
HEAD:   6ed4dbd (último commit)

Modified EN GOOGLE SHEET (no en git):
- 4 sheets nuevas (00_Config, 01_Feriados_PE, 02_Personas, 20_Dashboard)
- HE rediseñado con fórmulas (eliminé duplicados, arreglé #REF!, agregué fechas)
- Apps Script proyecto v2 publicado

NO COMMITEADO al git:
- Cambios en spreadsheet remoto (Sheets no es git-trackable)
- appsscript.json con addOns.common.name + logoUrl (último commit 6ed4dbd SÍ lo incluye)
```

Para revertir todos los cambios del sheet: borrar manualmente las 4 hojas nuevas y el contenido del Apps Script.

---

*Generado por Thavren al cierre de sesión · 2026-09-25*


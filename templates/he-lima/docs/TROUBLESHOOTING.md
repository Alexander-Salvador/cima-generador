# Troubleshooting — HE Lima

> **Guía de errores comunes y soluciones.**
> **Audiencia**: cualquier usuario (técnicos, supervisores, admin, CEO).
> **Tiempo de consulta**: ~5 minutos por error (con la tabla).
> **Si no encontrás tu error**: revisar [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) sección "Limitaciones" o escalar.

---

## 🚨 Índice rápido por categoría

| Categoría | Errores |
|---|---|
| 🔐 Permisos y autenticación | #1, #2, #3 |
| 📊 Datos y fórmulas | #4, #5, #6, #7, #8 |
| 🎨 Visualización y colores | #9, #10 |
| 🔄 Sincronización grid ↔ DB | #11, #12, #13 |
| ⚙️ Rendimiento | #14, #15 |
| 📅 Fechas y feriados | #16, #17 |
| 📦 Apps Script y triggers | #18, #19 |

---

## 📋 Tabla maestra de errores

### #1 — "Permiso denegado al ejecutar script"

| Campo | Valor |
|---|---|
| **Síntoma** | Popup "Authorization required" o "Exception: Permission denied" |
| **Causa probable** | Scopes OAuth no autorizados completamente |
| **Aplica a** | Cualquier función de Apps Script (`instalarTriggers()`, `poblarFeriados()`, etc.) |
| **Solución** | |
| 1. | Ir a **Extensiones → Apps Script** |
| 2. | Ejecutar cualquier función (ej: `getMes()`) |
| 3. | Aceptar **TODOS** los permisos que pide (puede ser 2-3 diálogos) |
| 4. | Si dice "no seguro", click **Avanzado → Ir a [nombre] (no seguro)** |
| 5. | Re-ejecutar la función que falló |
| **Prevención** | La primera vez, autorizar TODOS los scopes. Después no vuelve a pedir. |

---

### #2 — "No se puede acceder al spreadsheet" / "SpreadsheetApp.getActiveSpreadsheet() retorna null"

| Campo | Valor |
|---|---|
| **Síntoma** | Script se ejecuta pero no encuentra el sheet activo |
| **Causa probable** | Múltiples sheets abiertos, o el script está desvinculado |
| **Aplica a** | Scripts container-bound (asociados al sheet) |
| **Solución** | |
| 1. | Cerrar **todas** las pestañas de Google Sheets menos la que estamos configurando |
| 2. | Refrescar la página del sheet (`Ctrl+R` / `Cmd+R`) |
| 3. | Re-abrir **Extensiones → Apps Script** |
| 4. | Re-ejecutar la función |
| **Verificación** | El proyecto debe llamarse "Container-bound script" con el icono del sheet |

---

### #3 — "Acceso denegado" al abrir el sheet

| Campo | Valor |
|---|---|
| **Síntoma** | "You need permission to access this item" |
| **Causa probable** | No estás en la lista de compartidos, o tu cuenta es diferente |
| **Aplica a** | Cualquier usuario |
| **Solución** | |
| 1. | Pedirle al admin/CEO que te agregue como Editor o Lector |
| 2. | Verificar que estás usando la **misma cuenta Google** que te dieron acceso |
| 3. | Si usas múltiples cuentas (trabajo + personal), switchear a la correcta |
| **Prevención** | El admin debe compartir el sheet con emails específicos (no links genéricos) |

---

### #4 — "Grid no se actualiza al cambiar dropdown de mes"

| Campo | Valor |
|---|---|
| **Síntoma** | Cambias `00_Config!B2` pero el grid de `10_Registro_HE` no cambia |
| **Causa probable** | `instalarTriggers()` no se ejecutó, o trigger `onConfigChange` no instalado |
| **Aplica a** | Cambio de mes/año/semana |
| **Solución** | |
| 1. | Ir a **Extensiones → Apps Script** |
| 2. | Seleccionar función **`instalarTriggers()`** |
| 3. | Click ▶ Ejecutar |
| 4. | Esperar 5 segundos |
| 5. | Verificar en **Configuración del proyecto → Triggers** que aparezcan 3 |
| **Si persiste** | Verificar error #1 (permisos) y #18 (triggers quota) |

---

### #5 — "Hoja 98_DB_2026_XX no existe"

| Campo | Valor |
|---|---|
| **Síntoma** | Error "Cannot find sheet 98_DB_2026_03" |
| **Causa probable** | Es el **primer acceso** al mes — la hoja se crea automáticamente |
| **Aplica a** | Cambio a un mes nunca antes accedido |
| **Solución** | |
| 1. | **Esto es normal** — `createMonthSheetIfMissing()` la crea lazy |
| 2. | Esperar 5-10 segundos |
| 3. | Verificar en menú "Todas las hojas" (click derecho en cualquier pestaña) que ahora aparece |
| 4. | La hoja queda oculta — es lo correcto |
| **Si persiste > 30s** | Verificar error #1 (permisos) y error #18 (triggers) |

---

### #6 — "Fila marcada como Feriado en día laboral"

| Campo | Valor |
|---|---|
| **Síntoma** | Un martes (no feriado) aparece con tipo "Feriado" en `10_Registro_HE!E` |
| **Causa probable** | Formato de fecha incorrecto en `01_Feriados_PE!D:D` |
| **Aplica a** | Clasificación de tipo de día |
| **Solución** | |
| 1. | Ir a `01_Feriados_PE` |
| 2. | Verificar columna D — **debe ser formato `yyyy-MM-dd`** (ej: `2026-04-02`) |
| 3. | Si está como texto (ej: `2/4/2026` o `2-Abr`), corregir manualmente |
| 4. | Verificar que NO haya duplicados en columna D |
| 5. | Re-ejecutar `04_FormulasRecalc.recalcularFila(fila_afectada)` desde Apps Script |
| **Prevención** | Solo admin edita `01_Feriados_PE`. Si la fecha es Pascua móvil, dejar que `02_HolidaysEngine` la calcule. |

---

### #7 — "Pago_HE_PEN incorrecto" (valor errado o cero)

| Campo | Valor |
|---|---|
| **Síntoma** | Columna P muestra `0.00` o un valor que no coincide con cálculo manual |
| **Causa probable** | Tarifa desactualizada o vacía en `02_Personas!I` |
| **Aplica a** | Cálculo de pago |
| **Diagnóstico** | |
| 1. | Ir a `02_Personas` |
| 2. | Buscar la fila del ID_Persona afectada |
| 3. | Verificar columna **I (Tarifa_Hora_PEN)**: |
|    - ¿Tiene valor numérico? (ej: `15.00`) |
|    - ¿Es ≥ 0? |
|    - ¿La columna H (Activo) es `TRUE`? |
| **Solución** | |
| 1. | Corregir tarifa en `02_Personas!I` |
| 2. | Verificar que columna H sea `TRUE` |
| 3. | Esperar 5 segundos (VLOOKUP recalcula) |
| 4. | Si persiste, ejecutar `04_FormulasRecalc.recalcularFila(fila)` |

---

### #8 — "Celdas #REF!" o #VALUE!

| Campo | Valor |
|---|---|
| **Síntoma** | Fórmula muestra `#REF!`, `#VALUE!`, `#NAME?` o `#N/A` |
| **Causa probable** | Fórmula referencia columna/rango borrado, o dato incompatible |
| **Aplica a** | Cualquier celda con fórmula |
| **Diagnóstico** | |
| 1. | Click en la celda con error |
| 2. | Ver la fórmula en la barra superior |
| 3. | Identificar qué referencia falta |
| **Solución por tipo** | |
| `#REF!` | Se borró una columna/fila referenciada. **Restaurar desde backup** (ver ADMIN_GUIDE) |
| `#VALUE!` | Tipo de dato incorrecto (texto en lugar de número). Revisar input. |
| `#NAME?` | Función no reconocida. Verificar ortografía. |
| `#N/A` | VLOOKUP/INDEX no encontró el valor. Verificar que el ID existe en `02_Personas`. |
| **Recálculo manual** | Ejecutar `04_FormulasRecalc.recalcularFila(fila_afectada)` desde Apps Script |

---

### #9 — "Los colores no se aplican / aparecen mal"

| Campo | Valor |
|---|---|
| **Síntoma** | Las celdas de `Tipo_Dia` no cambian a verde/amarillo/rojo |
| **Causa probable** | Conditional Formatting no aplicada o condición rota |
| **Aplica a** | Visualización de tipo de día, HE alta, sync status |
| **Solución** | |
| 1. | Seleccionar rango `10_Registro_HE!E2:E110` |
| 2. | **Formato → Formato condicional** |
| 3. | Verificar que existan las reglas: |
|    - `cf_feriado`: Text equals "Feriado" → rojo |
|    - `cf_finsem`: Text equals "Sabado" o "Domingo" → amarillo |
|    - `cf_normal`: Text equals "Normal" → verde |
| 4. | Si falta alguna, agregar manualmente |
| 5. | Si están mal configuradas, eliminar y recrear (ver [`sheets-spec.md`](../sheets-spec.md)) |

---

### #10 — "El dropdown no aparece" (B2, B3, B4, etc.)

| Campo | Valor |
|---|---|
| **Síntoma** | Click en `00_Config!B2` pero no aparece la flecha de dropdown |
| **Causa probable** | Data Validation no aplicada o lista de valores vacía |
| **Aplica a** | Cualquier celda con dropdown (B2, B3, B4, B5, también columnas de `02_Personas`) |
| **Solución** | |
| 1. | Click en la celda |
| 2. | **Datos → Validación de datos** |
| 3. | Verificar que exista criterio: "Lista de elementos" |
| 4. | Verificar que los valores estén correctos (ej: `1,2,3,...,12` para mes) |
| 5. | Si está vacío, aplicar manualmente (ver [`sheets-spec.md`](../sheets-spec.md) sección Data Validation) |
| **Prevención** | Las DV se configuran durante instalación. Si se borraron, contactar admin. |

---

### #11 — "Sync_Status = 'conflict' (rojo) en columna V"

| Campo | Valor |
|---|---|
| **Síntoma** | Celdas de `10_Registro_HE!V` aparecen en rojo con texto "conflict" |
| **Causa probable** | Race condition: dos operaciones simultáneas de sync |
| **Aplica a** | Filas afectadas por cambio de mes reciente |
| **Solución inmediata** | |
| 1. | **NO seguir editando** en esa zona |
| 3. | Esperar 30 segundos (LockService debería liberar) |
| 4. | Refrescar página (`Ctrl+R`) |
| 5. | Re-ejecutar `08_DBSync.onConfigChange()` manualmente desde Apps Script |
| **Diagnóstico** | Revisar `99_Audit_Log` — buscar acciones `sync_save` y `sync_load` muy cercanas en tiempo |
| **Prevención** | Evitar que múltiples admins cambien mes simultáneamente |

---

### #12 — "Sync_Status queda en 'pending' (amarillo) más de 1 minuto"

| Campo | Valor |
|---|---|
| **Síntoma** | Las filas se quedan amarillas (sync pending) y nunca pasan a "synced" |
| **Causa probable** | Apps Script timeout o error silencioso durante saveGridToDB |
| **Solución** | |
| 1. | Refrescar página (`Ctrl+R`) |
| 2. | Si persiste, ir a **Extensiones → Apps Script → Ver → Registros** |
| 3. | Buscar errores rojos (Exception) |
| 4. | Re-ejecutar `08_DBSync.saveGridToDB(mes_actual, anio_actual)` manualmente |
| 5. | Verificar que la hoja `98_DB_YYYY_MM` tenga datos (mostrar hoja temporalmente) |
| **Si timeout** | Apps Script tiene 6 min máx. Si saveGridToDB excede → contactar `andres` |

---

### #13 — "Datos del mes anterior desaparecieron"

| Campo | Valor |
|---|---|
| **Síntoma** | Cambié de mes y al volver, los datos de la semana pasada no están |
| **Causa probable** | El sync `save → load` no se ejecutó correctamente, o hiciste cambios sin guardar |
| **Aplica a** | Cambio de mes con grid no vacío |
| **Solución** | |
| 1. | **NO panic** — los datos están en `98_DB_YYYY_MM` (hoja oculta) |
| 2. | Ir a **menú → Todas las hojas** → buscar `98_DB_{mes}_{anio}` |
| 3. | Mostrar hoja temporalmente (click derecho → Mostrar hoja) |
| 4. | Verificar que los datos están ahí |
| 5. | Volver a `10_Registro_HE`, seleccionar el mes correcto en `00_Config!B2` |
| 6. | Esperar 10 segundos (load automático) |
| **Prevención** | Esperá siempre 5 segundos después de cambiar dropdown antes de editar |

---

### #14 — "Performance lenta" (>3s para editar una celda)

| Campo | Valor |
|---|---|
| **Síntoma** | Editar una celda tarda más de 3 segundos en recalcular |
| **Causa probable** | Grid tiene >1000 filas activas (no debería), o INDIRECT está recalculando todo |
| **Aplica a** | Sheet con mucho uso |
| **Diagnóstico** | |
| 1. | Verificar que `10_Registro_HE` tiene **máximo 110 filas con datos** (filas 2-110) |
| 2. | Si tiene más → alguien insertó filas adicionales. **El sheet está mal usado**. |
| 3. | Ver `20_Dashboard` — si tiene INDIRECT excesivos (más de 12), están frenando |
| **Solución** | |
| 1. | **Archivar meses antiguos**: copiar `98_DB_2025_*` a archivo backup, eliminar del sheet activo |
| 2. | Limpiar `99_Audit_Log` si tiene >5000 filas (exportar primero) |
| 3. | NO usar `ARRAYFORMULA` sobre todo el rango (limitar a fórmulas fila por fila) |
| 4. | Si persiste, contactar `andres` con captura de los logs |

---

### #15 — "Cache desbordado" / "Exceed quota"

| Campo | Valor |
|---|---|
| **Síntoma** | Apps Script error: "Cache size exceeded" o "Service invoked too many times" |
| **Causa probable** | CacheService lleno (>50 MB), o demasiadas ejecuciones por minuto |
| **Aplica a** | Después de meses de uso intensivo |
| **Solución** | |
| 1. | En `01_CfgLoader.gs`, reducir TTL de cache (de 600 a 300 seg) |
| 2. | Invalidar cache manualmente: ejecutar `invalidateCache()` desde Apps Script |
| 3. | Reducir frecuencia de triggers time-driven |
| 4. | Si persiste, esperar 24h (cuota diaria se resetea) |

---

### #16 — "Jueves/Viernes Santo aparecen con fecha incorrecta"

| Campo | Valor |
|---|---|
| **Síntoma** | El feriado de Pascua tiene fecha que no coincide con calendario oficial |
| **Causa probable** | `02_HolidaysEngine.calcularPascua(anio)` tiene bug, o `poblarFeriados(anio)` no se ejecutó |
| **Aplica a** | Feriados móviles (F-02, F-03) |
| **Diagnóstico** | |
| 1. | Verificar fecha actual vs Pascua esperada (consultar calendario oficial) |
| 2. | Tabla de referencia 2024-2030 en [`holidays-pe.yaml`](../holidays-pe.yaml#datos-pre-calculados-referencia-para-auditoría) |
| **Solución** | |
| 1. | Ejecutar `poblarFeriados(anio)` desde Apps Script (recalcula) |
| 2. | Si sigue mal, comparar con tabla pre-calculada en YAML |
| 3. | Como workaround manual: editar celda `01_Feriados_PE!D2` o `D3` con la fecha correcta |
| 4. | Loggear en `99_Audit_Log`: `accion=update, sheet=01_Feriados_PE` |

---

### #17 — "Día festivo del año pasado aparece en año actual"

| Campo | Valor |
|---|---|
| **Síntoma** | `01_Feriados_PE` tiene 26 filas en lugar de 13 |
| **Causa probable** | `poblarFeriados(anio)` se ejecutó múltiples veces sin limpiar |
| **Aplica a** | Años consecutivos |
| **Solución** | |
| 1. | Antes de ejecutar `poblarFeriados(2027)`, borrar las filas del año anterior |
| 2. | Verificar manualmente que solo hay 13 feriados por año |
| 3. | Si hay duplicados: borrar manualmente las filas extras |

---

### #18 — "Triggers no se instalan" / "Quota exceeded"

| Campo | Valor |
|---|---|
| **Síntoma** | `instalarTriggers()` falla con "Too many triggers" o "Service error" |
| **Causa probable** | Límite de 30 triggers/usuario excedido (huérfanos de instalaciones previas) |
| **Aplica a** | Re-instalación del template |
| **Solución** | |
| 1. | **Configuración del proyecto → Triggers** |
| 2. | Identificar triggers duplicados o de pruebas anteriores |
| 3. | Click en el icono ❌ para eliminar cada uno |
| 4. | Dejar solo 3 (los instalados por `instalarTriggers()`): |
|    - `onEdit` (simple trigger) |
|    - `onConfigChange` (installable) |
|    - `triggerCierreMes` (time-driven) |
| 5. | Re-ejecutar `instalarTriggers()` |
| **Prevención** | Después de cualquier deploy, validar que solo hay 3 triggers |

---

### #19 — "Apps Script timeout (6 min)"

| Campo | Valor |
|---|---|
| **Síntoma** | Función termina con "Exceeded maximum execution time" |
| **Causa probable** | Operación masiva: saveGridToDB sobre >500 filas, o refresh de todo el dashboard |
| **Aplica a** | Hojas DB con muchos datos (>1500 filas × 22 cols = 33,000 operaciones) |
| **Solución inmediata** | |
| 1. | **Dividir operación** en chunks: procesar 1 semana por ejecución |
| 2. | Si saveGridToDB: editar para procesar filas 2-50, luego 51-110, etc. |
| 3. | Refrescar Dashboard solo para el mes activo (no YTD de 12 meses) |
| **Prevención** | El diseño v2 (working surface 110 filas) NO debería causar este error |

---

### #20 — "La auditoría muestra delete sospechoso"

| Campo | Valor |
|---|---|
| **Síntoma** | `99_Audit_Log` muestra acción `delete` que no esperabas |
| **Causa probable** | Alguien borró filas (no debería pasar normalmente) |
| **Aplica a** | Operación de admin |
| **Acción inmediata** | |
| 1. | **NO borrar más nada** |
| 2. | Revisar `99_Audit_Log` para identificar usuario (columna B) y timestamp |
| 3. | Verificar con el usuario si fue intencional |
| 4. | Si fue accidental: restaurar desde backup (ver ADMIN_GUIDE §6) |
| 5. | Si fue malicioso: escalar al CEO inmediatamente |
| **Prevención** | NUNCA borrar filas manualmente. Para "dar de baja" usar `Activo=FALSE` (ver ADMIN_GUIDE §2) |

---

## 🔧 Diagnóstico general

Si tu error no aparece en la lista, seguí este flujo:

### Paso 1 — Recolectar información

```
[ ] ¿Qué estabas haciendo cuando ocurrió? (operación)
[ ] ¿Qué error viste? (mensaje exacto, captura)
[ ] ¿Cuándo empezó? (después de qué cambio)
[ ] ¿A quién afecta? (solo a vos, a todos)
[ ] ¿Se reproduce? (cada vez / a veces / una vez)
```

### Paso 2 — Revisar logs de Apps Script

1. **Extensiones → Apps Script**
2. Click en **"Ver → Registros"** (o **"Executions"** en el panel izquierdo)
3. Buscar el timestamp del error
4. Copiar el mensaje completo

### Paso 3 — Revisar `99_Audit_Log`

1. Mostrar hoja (click derecho → Mostrar hoja)
2. Filtrar por tu email (columna B) en la última hora
3. Identificar la última acción que ejecutaste

### Paso 4 — Diagnosticar

| Síntoma | Sección de este doc |
|---|---|
| Error de permisos | #1, #2, #3 |
| Datos/celdas incorrectos | #4, #5, #6, #7, #8 |
| Visualización (colores, dropdowns) | #9, #10 |
| Sincronización grid ↔ DB | #11, #12, #13 |
| Performance | #14, #15 |
| Feriados | #16, #17 |
| Triggers / scripts | #18, #19 |
| Auditoría sospechosa | #20 |

### Paso 5 — Escalar

Si después de los pasos anteriores no resolviste:

| Nivel | Contacto | Cuándo |
|---|---|---|
| **T1** | Tu supervisor | Operación normal, dudas |
| **T2** | Admin (CEO/delegado) | Problema no resuelto en 15 min |
| **T3** | `andres` (implementador) | Bug técnico, error de Apps Script |
| **T4** | `marco` (architect) | Decisión de diseño |
| **T5** | `braka` (CEO) | Cambios mayores o incidentes seguridad |

---

## 📚 Referencias

| Documento | Cuándo consultar |
|---|---|
| [`docs/USER_GUIDE.md`](USER_GUIDE.md) | Operación básica |
| [`docs/ADMIN_GUIDE.md`](ADMIN_GUIDE.md) | Tareas administrativas |
| [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) | Entender el "por qué" |
| [`docs/GLOSSARY.md`](GLOSSARY.md) | Términos desconocidos |
| [`docs/CHANGELOG.md`](CHANGELOG.md) | Si el error es de una versión nueva |
| [`apps-script/*.gs`](../apps-script/) | Código fuente para debugging |

---

## 🆘 Contacto de emergencia

Si el sheet está **completamente roto** y no podés operar:

1. **NO borrar nada**
2. Hacer backup inmediato: **Archivo → Hacer una copia**
3. Renombrar backup: `... — BROKEN {fecha}`
4. Abrir el backup en otra pestaña — probablemente funciona
5. Contactar a `andres` con:
   - ID del spreadsheet
   - Backup del sheet roto
   - Descripción del error
   - Logs de Apps Script

---

*HE Lima · TROUBLESHOOTING.md · v0.2.0-rev · 2026-09-24*
*Authority: nadia (docs) · Reviewer: clara (QA) · Approver: thavren (lead platform)*
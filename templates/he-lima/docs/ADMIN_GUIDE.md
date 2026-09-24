# Manual del Administrador — HE Lima

> **Audiencia**: CEO Braka o delegado autorizado con permisos de Editor + acceso a Apps Script.
> **Tiempo de lectura**: 20 minutos.
> **Responsabilidad**: mantener la base de personas, factores HE, feriados, backups y auditoría.

---

## 🔑 Tus permisos

Como administrador deberías tener:

| Permiso Google Sheets | Descripción |
|---|---|
| **Editor** en el spreadsheet | Modificar datos y estructura |
| Acceso a **Extensiones → Apps Script** | Ejecutar funciones, ver logs |
| Acceso a **Configuración del proyecto → Triggers** | Gestionar automatizaciones |
| **Propietario** (recomendado) | Transferir ownership, eliminar definitivamente |

> ⚠️ **Si no sos Propietario**, solicitá la transferencia al CEO Braka. Sin eso, no podés delegar ownership ni eliminar el spreadsheet.

---

## 👥 Operación 1 — Añadir nueva persona

### Cuándo hacerlo

Cuando entra un nuevo técnico, supervisor o jefe de cuadrilla al equipo.

### Pasos

1. Ir a **`02_Personas`** (pestaña azul)
2. Buscar la **primera fila vacía** después de la última persona (normalmente fila 21+)
3. Completar las columnas:

| Col | Campo | Ejemplo | Validación |
|---|---|---|---|
| A | ID_Persona | `P-021` | Formato `P-NNN` (3 dígitos, zero-padded) |
| B | DNI | `45678901` | 8 dígitos exactos, único |
| C | Apellidos_Nombres | `RODRÍGUEZ LÓPEZ, CARLOS MIGUEL` | Mayúsculas, completo |
| D | Cargo | `tecnico` | Dropdown: tecnico / supervisor / jefe_cuadrilla / admin |
| E | Cuadrilla | `B` | Dropdown: A / B / C / D |
| F | Turno_Base | `diurno` | Dropdown: diurno / nocturno / rotativo |
| G | Fecha_Ingreso | `2026-03-15` | No puede ser futuro |
| H | Activo | `TRUE` | TRUE = visible en dropdowns; FALSE = oculto |
| I | Tarifa_Hora_PEN | `15.00` | Decimal con punto, ≥ 0 |
| J | Notas | `Ingresó por reemplazo` | Libre |

4. Verificar que el DNI **no esté duplicado** (buscar en la columna B con `Ctrl+F`)
5. La fila ya está activa — no requiere Apps Script

### Verificación

```
[ ] El ID aparece en el dropdown de F (columna F de 10_Registro_HE)
[ ] El ID aparece en el dropdown de B5 (00_Config — filtro persona)
[ ] Al escribir el ID en una fila, columna G se autocompleta con el nombre
```

### Automatización (opcional)

Si tenés muchas altas/bajas, podés pedirle a `andres` que cree una función:
```javascript
// apps-script/nuevo: 09_PersonasAdmin.gs
function altaPersona(dni, nombre, cargo, cuadrilla, turno, tarifa) {
  // 1. genera ID_Persona siguiente (max + 1)
  // 2. valida DNI no duplicado
  // 3. escribe fila
  // 4. log a 99_Audit_Log
}
```

---

## 🚪 Operación 2 — Dar de baja persona (sin perder histórico)

### ⚠️ NUNCA borrar la fila

**El histórico de horas extras debe preservarse para siempre**. NUNCA borrar filas de `02_Personas` ni de `98_DB_YYYY_MM`.

### Pasos correctos

1. Ir a **`02_Personas`**
2. Buscar la fila de la persona
3. Columna **H (Activo)**: cambiar de `TRUE` a `FALSE`
4. (Opcional) Columna **J (Notas)**: agregar fecha de baja y motivo
   - Ej: `"Baja 2026-08-15 — renuncia voluntaria"`

### ¿Qué pasa automáticamente?

| Acción | Efecto |
|---|---|
| `Activo = FALSE` | La persona **desaparece** de los dropdowns de filtro y FK |
| Filas históricas | Se mantienen intactas en `98_DB_YYYY_MM` — los joins siguen funcionando vía VLOOKUP |
| Reportes | Las queries que filtran `Activo=TRUE` automáticamente la excluyen |
| Pagos retroactivos | Si necesitás pagar HE de meses anteriores, los datos están ahí |

### Verificación post-baja

```
[ ] La persona NO aparece en dropdown de 00_Config!B5
[ ] La persona NO aparece en dropdown de columna F (10_Registro_HE)
[ ] Filas existentes en 98_DB_YYYY_MM siguen mostrando el nombre en columna G
[ ] NO se borraron filas en 98_DB_YYYY_MM
```

### ¿Y si necesito reactivarla?

Simplemente cambiá columna H de `FALSE` a `TRUE`. Vuelve a aparecer en todos los dropdowns.

---

## ⚙️ Operación 3 — Modificar factores HE

### Cuándo modificarlo

- Cambios legales (nuevos factores en normativa peruana)
- Decisión del CEO (recompensa especiales)
- Corrección de error

### Factores configurables (`00_Config`)

| Celda | Factor | Default | Rango legal |
|---|---|---|---|
| `B6` | HE Diurno | `1.25` | 1.0 – 3.0 |
| `B7` | HE Nocturno | `1.35` | 1.0 – 3.0 |
| `B8` | Factor Feriado | `2.00` | 1.0 – 3.0 |
| `B9` | Jornada Max Diaria | `8` | 4 – 12 |

### Pasos

1. Ir a **`00_Config`**
2. Editar la celda correspondiente (ej: B6 = 1.30 para subir factor diurno)
3. **Esperar 5 segundos** — los `1.25` se recalculan en `10_Registro_HE` columna N
4. Verificar en Dashboard (`20_Dashboard`) que los pagos reflejen el nuevo factor

### ⚠️ Precauciones

- **NO** cambiar durante edición activa de técnicos (puede generar confusion)
- **Anotar** en `99_Audit_Log` manualmente: `acción = update, sheet = 00_Config, fila = 6, valor_nuevo = 1.30`
- Si el cambio afecta **mes cerrado** (estado `pagado`), **NO recalcula hacia atrás** — solo aplica hacia adelante

### Cambio histórico (meses ya cerrados)

Si necesitás recalcular un mes ya pagado:
1. Hacer backup del mes (ver Operación 5)
2. Cambiar el factor en `00_Config!B6`
3. Ir a `10_Registro_HE`, seleccionar el mes, esperar sync
4. **Verificar** que los nuevos valores son correctos
5. Documentar en `99_Audit_Log` (acción = `recalc`)
6. (Opcional) Restaurar el factor original después del recálculo

---

## 🗓️ Operación 4 — Agregar nuevo feriado

### Cuándo hacerlo

Cuando el gobierno publica un nuevo feriado nacional, o se confirma una fecha móvil (Pascua).

### Pasos para feriado FIJO (ej: nuevo feriado nacional)

1. Ir a **`01_Feriados_PE`** (pestaña roja)
2. Buscar la primera fila vacía (fila 14 en v2.0)
3. Completar:

| Col | Campo | Ejemplo | Notas |
|---|---|---|---|
| A | ID | `F-14` | siguiente correlativo |
| B | Nombre | `Día de la Marina` | nombre canónico |
| C | Tipo | `nacional_fijo` | `nacional_fijo` o `nacional_movil` |
| D | Fecha | `2026-12-08` | formato `yyyy-MM-dd` |
| E | Es_Movil | `FALSE` | FALSE para fijos |
| F | Aplica_HE | `TRUE` | TRUE = paga factor feriado |

### Pasos para feriado MÓVIL (siguiente año)

**NO insertarlo manualmente** — el algoritmo de Pascua (en `02_HolidaysEngine.gs`) lo calcula automáticamente.

**Para forzar el recálculo**:
1. Abrir **Extensiones → Apps Script**
2. Seleccionar función: `poblarFeriados(2027)` (donde 2027 es el año)
3. Click ▶ Ejecutar
4. Verificar en `01_Feriados_PE` que Jueves Santo y Viernes Santo aparecen con las fechas correctas

### Verificación

```
[ ] El nuevo feriado aparece en 01_Feriados_PE
[ ] Al ir a 10_Registro_HE en la fecha del feriado, columna E muestra "Feriado" (rojo)
[ ] El factor aplicado (columna N) usa B8 (factor feriado = 2.00)
[ ] 99_Audit_Log registra la acción (acción = update)
```

### ⚠️ Precauciones

- **NO** agregar feriados regionales o municipales — la lista está **cerrada en 13 nacionales** (D.L. 713)
- Si querés añadir un feriado **solo para CIMA** (no nacional), no va aquí — manejar como bandera en `02_Personas` o crear columna custom
- Después de modificar feriados, **verificar el Dashboard** — los KPIs pueden cambiar

---

## 💾 Operación 5 — Hacer backup mensual

### Cuándo hacerlo

- **Recomendado**: el último día hábil de cada mes (antes del cierre)
- **Obligatorio**: antes de cualquier cambio estructural (añadir persona, cambiar factores)
- **Crítico**: antes de actualizaciones de Apps Script

### Método 1 — Backup manual (rápido)

1. En el spreadsheet: **Archivo → Hacer una copia**
2. Nombre: `CIMA — HE Lima — 2026 — BACKUP 2026-09-30`
3. Ubicación: tu Google Drive (o carpeta compartida del equipo)
4. Click **"Hecho"**

### Método 2 — Backup por hoja (selectivo)

Si solo querés backupear ciertas hojas:

1. Click derecho en la pestaña → **"Crear una copia"**
   - Solo funciona para hojas individuales
   - El nombre debe incluir el mes

### Método 3 — Backup automático (Apps Script)

Pedile a `andres` que configure un trigger mensual:

```javascript
// Trigger: día 1 de cada mes a las 00:30
function backupMensual() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const nombre = `CIMA — HE Lima — BACKUP ${new Date().toISOString().slice(0,10)}`;
  // DriveApp.createFile(nombre, ...) — guarda copia en /Backups
}
```

### Política de retención sugerida

| Antigüedad | Acción |
|---|---|
| 0–3 meses | Mantener 3 backups completos |
| 3–12 meses | Mantener 1 backup por trimestre |
| > 12 meses | Mover a archivo frío (opcional) |

---

## ♻️ Operación 6 — Restaurar desde backup

### ⚠️ Antes de restaurar

1. **Hacer backup del estado actual** (puede que quieras volver atrás si algo sale mal)
2. **Avisar al equipo** — la restauración afecta a todos
3. **Coordinar horario** — hacerlo en momento de baja actividad

### Pasos

1. Abrí el archivo de backup (ej: `CIMA — HE Lima — 2026 — BACKUP 2026-09-30`)
2. Click en **"Archivo → Mover"** → mover al Drive del spreadsheet original (o donde estaba)
3. **Renombrar** el archivo activo actual a `... — VIEJO 2026-XX-XX`
4. Renombrar el backup a `CIMA — HE Lima — 2026`
5. (Opcional) Re-ejecutar `instalarTriggers()` para asegurar triggers frescos

### Verificación post-restauración

```
[ ] Triggers siguen instalados (ver Configuración del proyecto → Triggers)
[ ] 00_Config tiene los valores correctos (mes/año/añoactivo)
[ ] 02_Personas tiene todas las personas (sin pérdida)
[ ] 10_Registro_HE muestra datos del mes activo esperado
[ ] 99_Audit_Log NO se sobrescribió (debería estar vacío o tener solo logs de hoy)
```

### Restauración selectiva (solo某些 hojas)

Si solo querés restaurar UNA hoja (ej: `98_DB_2026_08`):

1. En el backup, click derecho en la pestaña → **"Copiar a → spreadsheet existente"**
2. Seleccionar el spreadsheet activo
3. La hoja se copia al final — **renombrar** si hay conflicto de nombres
4. Verificar que los datos son los esperados

---

## 📊 Operación 7 — Leer el Dashboard

### Bloques del Dashboard (`20_Dashboard`)

```
┌────────────────────────────────────────────────────────┐
│ 20_Dashboard — KPIs Ejecutivos                         │
├────────────────────────────────────────────────────────┤
│ A1:F4   4 KPI cards:                                   │
│           HE Mes    HE YTD    Pago Mes    Personas     │
│           [12.5h]   [45.0h]   [S/.562]    [47 act.]   │
├────────────────────────────────────────────────────────┤
│ A10:F40 Resumen por persona (tabla)                     │
│           ID    Nombre        HE Mes  HE YTD  Pago     │
│           P-001 PÉREZ, JUAN   12.5h   45.0h   S/. 562  │
│           P-002 LÓPEZ, ANA    8.0h    32.0h   S/. 400  │
│           ...                                          │
├────────────────────────────────────────────────────────┤
│ H1:N40   Heatmap (semana × persona)                     │
│           Matriz con gradiente de color                 │
│           (verde=bajo HE, rojo=alto HE)                 │
├────────────────────────────────────────────────────────┤
│ A20:L20  Gráfico HE mensual (columnas agrupadas)        │
│           Barras por mes, agrupadas por cuadrilla       │
├────────────────────────────────────────────────────────┤
│ D20:F40  Top 10 HE (SORTN)                              │
│           Personas con más horas ponderadas             │
└────────────────────────────────────────────────────────┘
```

### Cómo interpretar

| KPI | Qué mirar | Acción si... |
|---|---|---|
| **HE Mes** | Horas extras del mes activo | Si > 200h → revisar con supervisores |
| **HE YTD** | Acumulado año hasta el mes activo | Comparar con año anterior (tendencia) |
| **Pago Mes** | Soles a pagar este mes | Si > S/. 10,000 → validar presupuesto |
| **Personas activas** | Cuántos del equipo están activos | Si < 40 → revisar bajas |

### Heatmap (semana × persona)

- **Verde claro**: pocas horas extras (< 4h)
- **Verde oscuro**: horas moderadas (4-8h)
- **Amarillo**: horas altas (8-12h)
- **Rojo**: horas críticas (> 12h) — **revisar con supervisor**

> ⚠️ Si ves rojo consistente en la misma persona → posible burnout o error de registro.

---

## 🔍 Operación 8 — Auditar cambios en `99_Audit_Log`

### Cuándo hacerlo

- Sospecha de error en cálculos
- Verificar quién cambió qué
- Investigación de incidentes
- Reportes de cumplimiento

### Estructura del audit log

| Col | Header | Ejemplo |
|---|---|---|
| A | Timestamp | `2026-09-24 14:32:15` |
| B | Usuario | `braka@ovav.io` |
| C | Acción | `update`, `sync_load`, `sync_save`, `sync_create_sheet`, `delete`, `trigger` |
| D | Sheet | `10_Registro_HE` |
| E | Fila | `42` |
| F | Columna | `P` |
| G | Valor_Anterior | `30.00` |
| H | Valor_Nuevo | `45.00` |

### Cómo ver el log

1. Ir a **Configuración del proyecto → Triggers** o **Hoja → Mostrar hojas** → marcar `99_Audit_Log`
2. Click derecho → **"Mostrar hoja"** (si está oculta)
3. Ordenar por Timestamp (descendente) para ver lo más reciente primero
4. Filtrar por Usuario o Acción si necesitás

### Búsquedas comunes

```
# ¿Quién modificó la fila 42 de 10_Registro_HE?
Filtro: Sheet = "10_Registro_HE" AND Fila = 42

# ¿Cuántas veces se cargó el mes de marzo?
Filtro: Acción = "sync_load" AND Sheet contiene "98_DB_2026_03"

# Todos los cambios de la última semana
Filtro: Timestamp >= HOY() - 7
```

### Acciones registradas automáticamente

| Acción | Cuándo |
|---|---|
| `create` | Nueva fila creada |
| `update` | Celda editada |
| `delete` | Fila eliminada (RARO — debe investigarse) |
| `recalc` | Recálculo manual |
| `trigger` | Instalación/eliminación de trigger |
| `sync_load` | Cambio de mes → carga grid |
| `sync_save` | Cambio de mes → guarda grid |
| `sync_create_sheet` | Auto-creación de hoja DB mensual |

### ⚠️ Qué buscar

| Patrón | Significa |
|---|---|
| Múltiples `delete` | Alguien borró filas (no debería pasar normalmente) |
| Cambios en `00_Config!B6-B9` | Modificación de factores (inusual, debe documentarse) |
| Cambios masivos en `10_Registro_HE` | Edición batch (exportar/importar) |
| `sync_create_sheet` en mes cerrado | Re-apertura de mes cerrado |

### Política de retención

- **NO borrar el audit log** — append-only es ley
- Si llega a 10,000 filas (límite del rango), **exportar a archivo** y resetear:
  1. File → Download as `.csv`
  2. Guardar en `/Auditoría/HE-Lima/2026-Q3-audit.csv`
  3. Vaciar `99_Audit_Log` (preservar header)

---

## 🆘 Escalar problemas técnicos

Si después de seguir los pasos de [`docs/TROUBLESHOOTING.md`](TROUBLESHOOTING.md) el problema persiste:

### Niveles de escalación

```
Tú (admin/CEO)
    ↓ si no resolvés
andres (Implementador Senior) — bugs de Apps Script
    ↓ si es arquitectónico
marco (Systems Architect) — decisiones de diseño
    ↓ si requiere cambio mayor
CEO Braka (waiver) — autoriza refactor
```

### Información a proporcionar

Cuando escales, incluí:

1. **Descripción del problema** (qué pasó, qué esperabas)
2. **Pasos para reproducir** (secuencia exacta)
3. **Captura de pantalla** del error
4. **Logs de ejecución** (Extensiones → Apps Script → Ver → Registros)
5. **ID del spreadsheet**
6. **Último cambio** que recuerdes haber hecho

---

## 📚 Referencias rápidas

| Necesito... | Ir a... |
|---|---|
| Entender la arquitectura | [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) |
| Resolver un error específico | [`docs/TROUBLESHOOTING.md`](TROUBLESHOOTING.md) |
| Términos que no entiendo | [`docs/GLOSSARY.md`](GLOSSARY.md) |
| Manual para técnicos | [`docs/USER_GUIDE.md`](USER_GUIDE.md) |
| Detalle de una decisión de diseño | [`docs/ARCHITECTURE.md`](ARCHITECTURE.md#decisiones-arquitectónicas-adrs) |

---

*HE Lima · ADMIN_GUIDE.md · v0.2.0-rev · 2026-09-24*
*Audience: admin, CEO · Authority: nadia (docs)*
*Reviewer: thavren (lead platform) · Approver: braka (CEO)*
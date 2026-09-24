# Guía de Instalación — HE Lima

> **Audiencia**: administrador técnico (CEO Braka o delegado) que activará el template en Google Sheets.
> **Tiempo estimado**: 30–45 minutos (primera vez), 15 min (subsiguientes deployments).
> **Resultado esperado**: Google Sheet funcional con working surface, 12 hojas mensuales auto-creadas, triggers instalados.

---

## 📋 Pre-requisitos

Antes de empezar, verificá que tenés:

| # | Requisito | Cómo verificarlo |
|---|---|---|
| 1 | Cuenta Google con acceso a Sheets y Apps Script | sheets.google.com funciona |
| 2 | Navegador moderno (Chrome 120+ recomendado) | `chrome://version` |
| 3 | OVAV CLI instalada (opcional, para CI/CD) | `ovav --version` → `>=1.1.0` |
| 4 | Los 8 archivos `.gs` de Apps Script (en `apps-script/`) | Listar `apps-script/*.gs` |
| 5 | El archivo `appsscript.json` con scopes correctos | Está junto a los `.gs` |
| 6 | Lista de personas a registrar (DNI, cargo, cuadrilla) | Mínimo 5 personas para test |

---

## 🚀 Paso 1 — Crear nuevo Google Sheet

1. Ir a https://sheets.new
2. Renombrar a: **`CIMA — HE Lima — YYYY`** (ej: `CIMA — HE Lima — 2026`)
3. **No compartir todavía** — primero configuramos, después compartimos.
4. Anotar el **ID del spreadsheet** (parte de la URL entre `/d/` y `/edit`):
   ```
   https://docs.google.com/spreadsheets/d/ESTE_ES_EL_ID/edit
   ```

> ⚠️ **Tip**: guardá este ID — lo vas a necesitar para auditoría y restauración.

---

## 📑 Paso 2 — Crear sheets manualmente

Google Sheets no permite crear sheets con nombres exactos vía API sin programarlo. **Hacelo manualmente**:

### 2.1 Sheet base (5 visibles + 1 oculta inicial)

Click en el **`+`** de la barra inferior para crear nuevas hojas. **Nombrarlas EXACTAMENTE así** (sin tildes, sin espacios extra):

| Orden | Nombre exacto | Color tab (opcional) | Visibilidad |
|---|---|---|---|
| 1 | `00_Config` | `#1F2937` (gris oscuro) | visible |
| 2 | `01_Feriados_PE` | `#DC2626` (rojo) | visible |
| 3 | `02_Personas` | `#2563EB` (azul) | visible |
| 4 | `10_Registro_HE` | `#059669` (verde) | visible |
| 5 | `20_Dashboard` | `#7C3AED` (violeta) | visible |
| 6 | `99_Audit_Log` | `#374151` (gris) | **oculta** |

> ❗ **El orden importa**: `00_Config` debe ser la primera hoja (la que se abre por defecto).
> Click derecho en la pestaña → "Mover al principio" si no lo está.

### 2.2 Ocultar `99_Audit_Log`

Click derecho en la pestaña de `99_Audit_Log` → **"Ocultar hoja"**.

> Las 12 hojas mensuales `98_DB_YYYY_MM` **NO se crean manualmente** — el script `08_DBSync.gs` las crea automáticamente la primera vez que accedas al mes (lazy creation).

### 2.3 Verificación rápida

Deberías ver **6 pestañas** en la barra inferior (5 visibles + 1 oculta — la oculta no aparece, eso es correcto).

---

## 📝 Paso 3 — Copiar contenido de cada Apps Script

### 3.1 Abrir el editor de Apps Script

En tu Google Sheet: **Extensiones → Apps Script**.

> Esto abre una pestaña nueva con el editor. El proyecto se llama "Container-bound script".

### 3.2 Crear archivos `.gs`

En el panel izquierdo del editor (sección "Archivos"):

1. Click en **`+`** → **"Script"** → nombrarlo `01_CfgLoader`
2. Repetir para cada módulo:

| Archivo | Nombre exacto | Módulo responsabilidad |
|---|---|---|
| Script 1 | `01_CfgLoader` | Lee `00_Config` y expone vía CacheService |
| Script 2 | `02_HolidaysEngine` | Pascua móvil + lookup feriados |
| Script 3 | `03_RegistroEngine` | Maneja grid + valida invariantes |
| Script 4 | `04_FormulasRecalc` | onEdit + recalcular fila |
| Script 5 | `05_DashboardRefresh` | Refresca KPIs de `20_Dashboard` |
| Script 6 | `06_AuditLogger` | Append a `99_Audit_Log` |
| Script 7 | `07_TriggersWiring` | `instalarTriggers()` / `eliminarTriggers()` |
| Script 8 | `08_DBSync` | load/save entre grid y DB mensual |

### 3.3 Pegar el código

Para cada archivo `.gs`:

1. Abrí el archivo correspondiente en `apps-script/01_CfgLoader.gs` (en tu repo local)
2. Copiá **todo** el contenido
3. Pegalo en el editor de Apps Script (reemplaza el `function myFunction() {}` por defecto)
4. **Guardá** (`Ctrl+S` / `Cmd+S`)

> 💡 **Tip**: Cada archivo `.gs` tiene un comentario header indicando a qué módulo pertenece. Verificá que coincida con el nombre del archivo en el editor.

### 3.4 Verificación

Deberías tener **8 archivos** en el panel izquierdo del editor:
- `01_CfgLoader.gs`
- `02_HolidaysEngine.gs`
- `03_RegistroEngine.gs`
- `04_FormulasRecalc.gs`
- `05_DashboardRefresh.gs`
- `06_AuditLogger.gs`
- `07_TriggersWiring.gs`
- `08_DBSync.gs`

---

## ⚙️ Paso 4 — Configurar `appsscript.json`

### 4.1 Acceder al manifest

En el editor de Apps Script: **⚙️ (icono engranaje) → Configuración del proyecto → ⚙️ Configuración del proyecto → "Mostrar el archivo de manifiesto 'appsscript.json'"**.

> Si no ves esa opción, ir a: **Configuración del proyecto → Editores → ☑ Mostrar "appsscript.json"**.

### 4.2 Editar scopes

Click en el archivo `appsscript.json` que aparece en el panel izquierdo. Reemplazar el contenido con:

```json
{
  "timeZone": "America/Lima",
  "dependencies": {},
  "exceptionLogging": "STACKDRIVER",
  "runtimeVersion": "V8",
  "oauthScopes": [
    "https://www.googleapis.com/auth/spreadsheets.currentonly",
    "https://www.googleapis.com/auth/script.scriptapp",
    "https://www.googleapis.com/auth/script.send_mail"
  ]
}
```

| Scope | Por qué |
|---|---|
| `spreadsheets.currentonly` | Solo accede al sheet actual (no a todos los del Drive) |
| `script.scriptapp` | Necesario para `instalarTriggers()` |
| `script.send_mail` | Permite enviar reportes por email (futuro) |

### 4.3 Guardar

Click en el **icono de guardar** o `Ctrl+S`.

---

## 🔧 Paso 5 — Ejecutar `instalarTriggers()`

### 5.1 Seleccionar función

En la barra superior del editor (donde dice el dropdown de funciones):

1. Click en el dropdown → seleccionar **`instalarTriggers`** (del archivo `07_TriggersWiring.gs`)
2. Click en **▶ Ejecutar**

### 5.2 Autorizar scopes

La **primera vez** que ejecutes, Google pide autorización:

1. Click en **"Revisar permisos"**
2. Seleccioná **tu cuenta Google** (la misma del sheet)
3. Click en **"Avanzado"** → **"Ir a [nombre del proyecto] (no seguro)"**
4. Click en **"Permitir"**

> ⚠️ **¿Por qué dice "no seguro"?** Porque el script es nuevo y Google no lo verificó. Solo lo vas a ver vos (no es público).

### 5.3 Confirmar ejecución

Volvé al editor. En la parte inferior deberías ver:

```
✅ Ejecución correcta
   Duración: ~3 segundos
   Función instalada: instalarTriggers
```

### 5.4 Verificar triggers instalados

Ir a **⚙️ Configuración del proyecto → Triggers**. Deberías ver **3 triggers**:

| Función | Tipo | Evento |
|---|---|---|
| `onEdit` | Simple trigger | Al editar |
| `onConfigChange` | Installable trigger | `onEdit` (B2/B3/B4) |
| `triggerCierreMes` | Time-driven | Mensual, día 1 a las 00:00 |

> Si falta alguno, **re-ejecutá `instalarTriggers()`**.

---

## ✅ Paso 6 — Verificar instalación (checklist)

Ejecutá cada item. Si falla, consultá [`docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md).

### Checklist de verificación

```
[ ] 1. Google Sheet tiene 5 pestañas visibles: 00_Config, 01_Feriados_PE, 02_Personas,
       10_Registro_HE, 20_Dashboard
[ ] 2. Pestaña 99_Audit_Log está oculta (no aparece en la barra)
[ ] 3. 00_Config!B2 muestra el mes actual (dropdown 1-12)
[ ] 4. 00_Config!B3 muestra 2026 (dropdown años)
[ ] 5. 00_Config!B4 muestra "W01" a "W53" (dropdown semana)
[ ] 6. 01_Feriados_PE tiene 13 filas con feriados (ver holidays-pe.yaml)
[ ] 7. 02_Personas tiene al menos 5 personas con DNI válido (8 dígitos)
[ ] 8. 10_Registro_HE tiene headers en fila 1 con colores verde oscuro
[ ] 9. Cambiar 00_Config!B2 a "3" crea automáticamente 98_DB_2026_03 (verificar en
       menú "Todas las hojas")
[ ] 10. 20_Dashboard muestra KPI cards (no vacíos)
[ ] 11. 99_Audit_Log tiene al menos 3 filas (registros de instalación)
[ ] 12. Triggers instalados: 3 entries en Configuración del proyecto
```

### Test funcional end-to-end

```
1. Ir a 10_Registro_HE
2. En fila 2, columna F, escribir un ID_Persona válido (ej: "P-001")
3. Esperar 2 segundos → columna G debe auto-completar el nombre (VLOOKUP)
4. En columna I escribir "08:00", en columna J escribir "18:00"
5. Esperar 2 segundos → columna K debe mostrar "10.00"
6. Si es día Normal → columna L = 8.00, columna M = 2.00
7. Columna P (Pago_HE_PEN) debe mostrar > 0
```

Si los pasos 1–7 funcionan → **instalación exitosa** ✅.

---

## 🛠️ Paso 7 — Troubleshooting (errores comunes en instalación)

### Error: "Permiso denegado al ejecutar script"

**Causa**: Los scopes no se autorizaron completamente.

**Solución**:
1. Ir a la URL del sheet → **Extensiones → Apps Script**
2. Ejecutar cualquier función (ej: `getMes()`)
3. Aceptar TODOS los permisos cuando lo pida
4. Re-ejecutar `instalarTriggers()`

---

### Error: "No se puede acceder al spreadsheet"

**Causa**: El script está vinculado a otro sheet.

**Solución**:
1. Verificar que `SpreadsheetApp.getActiveSpreadsheet()` retorna el sheet correcto
2. Si hay múltiples sheets abiertos, cerrar todos menos el que estamos configurando

---

### Error: "Sheet 98_DB_2026_XX ya existe"

**Causa**: Re-ejecutaste `instalarTriggers()` o `createMonthSheetIfMissing()` sobre un mes ya creado.

**Solución**: **No es un error real** — el script valida con `if exists, return`. Solo loguea warning en `99_Audit_Log`. Ignorar.

---

### Error: "Triggers no se instalan"

**Causa**: Cuota de triggers excedida (límite: 30/usuario).

**Solución**:
1. Ir a **Configuración del proyecto → Triggers**
2. Eliminar triggers duplicados manualmente
3. Re-ejecutar `instalarTriggers()` (consolida en 3 máx)

---

### Error: "Las fórmulas dan #REF! al cambiar dropdown"

**Causa**: Algún módulo no se guardó correctamente o los anchors están mal.

**Solución**:
1. Verificar que el archivo `01_CfgLoader.gs` lee `B9` (no `B8`) para jornada max — ver [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md#12-cambios-de-anchors-resumen-rápido-para-andrés)
2. Re-ejecutar `04_FormulasRecalc.recalcularFila(2)` sobre la fila afectada
3. Si persiste, contactar a `andres` (implementador)

---

### Error: "El dropdown de semana no aparece en B4"

**Causa**: La data validation `dv_semana` no se aplicó.

**Solución**:
1. Ir a `00_Config!B4`
2. **Datos → Validación de datos**
3. Verificar: Lista de elementos, valores `W01,W02,...,W53`
4. Si está vacío, aplicar manualmente (ver [`sheets-spec.md`](../sheets-spec.md))

---

## 📊 Paso 8 — Post-instalación

### Compartir el sheet con el equipo

1. Click en **"Compartir"** (arriba a la derecha)
2. **Niveles de acceso**:

| Rol Google Sheets | Personas | Permisos |
|---|---|---|
| **Editor** | CEO Braka, Jefe de cuadrilla supervisor | Pueden editar `10_Registro_HE` |
| **Lector** | Técnicos, contabilidad | Solo lectura |
| **Propietario** | CEO Braka (transferir ownership) | Control total |

> ⚠️ **NO** dar permisos de Editor a todos — los técnicos deben usar el grid pero no modificar la estructura.

### Crear primera fila de prueba

1. Asegurarse que `02_Personas` tenga al menos 5 personas registradas (con DNI, tarifa)
2. Ir a `10_Registro_HE`
3. En fila 2, completar: F (ID_Persona), I (Hora_Ingreso), J (Hora_Salida), H (Turno)
4. Verificar que columnas K, L, M, N, O, P se calculan automáticamente
5. Ir a `99_Audit_Log` (mostrar hoja) → debería haber 5+ entradas

### Configurar cierre mensual

El trigger `triggerCierreMes` corre **automáticamente el día 1 de cada mes a las 00:00**:

1. Verificar que está activo: **Configuración del proyecto → Triggers**
2. Si querés probarlo manualmente: ejecutar `triggerCierreMes()` desde el editor

---

## 📚 Próximos pasos

Una vez completada la instalación:

| Para... | Leer... |
|---|---|
| Entender la arquitectura | [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) |
| Usar el sheet como técnico | [`docs/USER_GUIDE.md`](../docs/USER_GUIDE.md) |
| Administrar el sheet (altas/bajas/feriados) | [`docs/ADMIN_GUIDE.md`](../docs/ADMIN_GUIDE.md) |
| Resolver errores comunes | [`docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md) |

---

## 🆘 Soporte durante instalación

Si después de seguir todos los pasos el sheet no funciona:

1. **No te frustres** — la primera instalación suele requerir 1-2 iteraciones
2. Anotá el error exacto (captura de pantalla + mensaje)
3. Verificá [`docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md)
4. Si persiste, escalar a `andres` (implementador) con:
   - ID del spreadsheet
   - Output del log de ejecución (Ver → Registros)
   - Pasos que seguiste

---

*HE Lima · INSTALL.md · v0.2.0-rev · 2026-09-24*
*Authority: nadia (docs) · Reviewer: andres (impl) · Approver: braka (CEO)*
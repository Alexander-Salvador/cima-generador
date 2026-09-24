# Manual del Usuario — HE Lima

> **Audiencia**: jefes de cuadrilla, supervisores, técnicos y personal de CIMA — Técnica Móvil que registra horas extras.
> **Tiempo de lectura**: 10 minutos.
> **Lo que vas a aprender**: cómo abrir el sheet, registrar horas, entender los colores y qué **NO** hacer para no romper nada.

---

## 🚀 Empezar en 60 segundos

1. Abrí el Google Sheet **`CIMA — HE Lima — 2026`** (link compartido por tu supervisor)
2. Vas a ver **5 pestañas** abajo: `00_Config`, `01_Feriados_PE`, `02_Personas`, `10_Registro_HE`, `20_Dashboard`
3. **Andá a la pestaña `10_Registro_HE`** ← esta es donde trabajás
4. Listo — ya podés registrar horas

> 💡 **Tip**: Si nunca usaste este sheet, pedile a tu supervisor que te muestre una vez antes de empezar. Tomate 10 minutos para familiarizarte.

---

## 🎛️ Paso 1 — Seleccionar mes, año y semana

Arriba de todo en la pestaña `00_Config` (la primera hoja), tenés **4 dropdowns**:

```
00_Config — Panel de control
┌─────────────────────────────────────┐
│ A              B             C      │
│ ─────────── ────────── ─────────── │
│ Mes          [3] ▼         Mar 2026│  ← elegí el mes
│ Año          [2026] ▼              │  ← elegí el año
│ Semana       [W14] ▼               │  ← elegí la semana (W01-W53)
│ Persona      [TODOS] ▼             │  ← filtrá por persona (opcional)
│ Factor Diur  1.25                  │  ⚠️ NO TOCAR
│ Factor Noc   1.35                  │  ⚠️ NO TOCAR
│ Factor Fer   2.00                  │  ⚠️ NO TOCAR
│ Jornada Max  8                     │  ⚠️ NO TOCAR
└─────────────────────────────────────┘
```

**Cómo usar**:

| Querés ver... | Acción |
|---|---|
| El mes actual completo | Click en B2 → elegí el mes actual |
| Una semana específica | Click en B4 → elegí `W##` (ej: W14 = primera quincena de abril) |
| Solo las horas de Juan Pérez | Click en B5 → elegí `P-003` (su ID) |
| Volver a ver todos | Click en B5 → elegí `TODOS` |

> ⚠️ **Importante**: Cuando cambiás el mes/año, el sistema **guarda automáticamente** lo que tenías cargado y trae los datos del nuevo mes. **No perdés nada**.

---

## ✍️ Paso 2 — Registrar horas extras

Andá a **`10_Registro_HE`**. Vas a ver una tabla con 22 columnas. Las que **vos editás** son las primeras 13:

### Las columnas que SÍ editás

| Col | Header | Qué poner | Ejemplo |
|---|---|---|---|
| **A** | Fecha | Fecha del día trabajado | `2026-03-15` |
| **F** | ID_Persona | Tu código (ver dropdown) | `P-003` |
| **G** | Apellidos_Nombres | Se llena solo (VLOOKUP) | `PÉREZ GARCÍA, JUAN` |
| **H** | Turno | `diurno` o `nocturno` | `diurno` |
| **I** | Hora_Ingreso | Hora que entraste a trabajar | `08:00` |
| **J** | Hora_Salida | Hora que saliste | `18:00` |
| **K** | Horas_Total | Se calcula solo | `10.00` |
| **L** | Horas_Normales | Se calcula solo | `8.00` |
| **M** | Horas_Extras | Se calcula solo | `2.00` |
| **N** | Factor_Aplicado | Se calcula solo | `1.25` |
| **O** | Horas_Ponderadas | Se calcula solo | `2.50` |
| **P** | Pago_HE_PEN | Se calcula solo (en soles) | `37.50` |
| **Q** | Observaciones | Notas libres | `Trabajo en antena Norte` |

### Las columnas que NO tocás (son del sistema)

| Col | Header | Por qué |
|---|---|---|
| B | ISO_Week | Semana del año (calculada) |
| C | Dia_Semana | Lunes, martes... (calculada) |
| D | Es_Feriado | TRUE/FALSE (calculada) |
| E | Tipo_Dia | Normal/Sabado/Domingo/Feriado (calculada) |
| R | Aud_Usuario | Tu email (autocompletado por el sistema) |
| S | Aud_Timestamp | Fecha y hora del cambio (autocompletado) |
| T | Estado | `borrador` → `validado` → `aprobado` → `pagado` |
| U | Mes_Año | Sync key con hoja oculta (no tocar) |
| V | Sync_Status | Estado de sincronización (no tocar) |

---

## 🎨 Paso 3 — Entender los colores

El sheet usa colores para que **de un vistazo** sepas qué tipo de día es cada fila:

### Colores por tipo de día (columna E — `Tipo_Dia`)

| Color | Significa | Cuándo |
|---|---|---|
| 🟢 **Verde claro** | Día Normal | Lunes a viernes no feriados |
| 🟡 **Amarillo claro** | Sábado o Domingo | Fin de semana |
| 🔴 **Rosa/Rojo claro** | Feriado | Día feriado nacional (D.L. 713) |

### Colores de alerta (otros)

| Color | Dónde | Significa |
|---|---|---|
| 🟧 **Naranja fuerte** | Columna M (>4h extras) | Persona trabajó más de 4 horas extras — revisar con supervisor |
| 🟨 **Amarillo fuerte** | Columna P (pago alto) | Pago HE en el percentil 90 del mes — caso especial |
| ⚪ **Gris claro** | Columna F (persona inactiva) | La persona no está activa — revisar |
| 🟨 **Amarillo pálido** | Columna V (sync pending) | Datos aún no sincronizados con la hoja oculta — esperar |
| 🔴 **Rojo fuerte** | Columna V (sync conflict) | Conflicto de sincronización — **avisar al supervisor YA** |

> 💡 **Si ves rojo en columna V, NO continues editando** — llamá a tu supervisor o al administrador.

---

## 📋 Paso 4 — Estados del registro (columna T)

Tu fila pasa por **4 estados**:

```
[borrador] → [validado] → [aprobado] → [pagado]
   ↓             ↓            ↓            ↓
 lo editaste  supervisor    CEO/admin    tesorería
              lo revisó     lo aprobó    lo pagó
```

| Estado | Quién lo cambia | Cuándo |
|---|---|---|
| `borrador` | vos (al editar) | Apenas cargás las horas |
| `validado` | tu supervisor | Cuando confirma que las horas son correctas |
| `aprobado` | CEO / administrador | Cuando se autoriza el pago |
| `pagado` | tesorería | Cuando se pagó |

> ⚠️ **No cambies manualmente el estado** — el sistema lo gestiona. Si necesitás cambiar de `validado` a `borrador` para corregir, pedíselo a tu supervisor.

---

## ❌ Paso 5 — Qué NO hacer

Esto es **crítico**. Si hacés cualquiera de estas cosas, **rompés el sheet** y vas a perder el trabajo de todo el equipo:

### 🚫 NO editar columnas calculadas

❌ **NO escribas manualmente en**: B, C, D, E, K, L, M, N, O, P, U, V
   → El sistema las calcula. Si las tocás, **se pisan en el próximo recalc**.

### 🚫 NO borrar filas ni columnas

❌ **NO** borres filas (ni con click derecho → "Eliminar fila")
❌ **NO** borres columnas (A, B, C, ... V)
❌ **NO** insertes filas en medio (usá siempre la siguiente fila vacía al final)

### 🚫 NO renombrar las pestañas

❌ **NO** cambies `10_Registro_HE` por otro nombre
❌ **NO** crees hojas nuevas manualmente (las mensuales se crean solas)

### 🚫 NO cambiar los factores

❌ **NO** modifiques `00_Config!B6`, `B7`, `B8`, `B9` (factores HE y jornada)
   → Esos valores los maneja el administrador. Si los cambiás, **todos los pagos salen mal**.

### 🚫 NO compartir el link públicamente

❌ **NO** compartas el link en WhatsApp/grupos abiertos
   → Solo el CEO/admin comparte con personas autorizadas.

### 🚫 NO tocar `99_Audit_Log`

Esa hoja está oculta. **NO la desocultes para editar**. Es solo para auditoría.

---

## 📊 Paso 6 — Ver el Dashboard

Andá a la pestaña **`20_Dashboard`**:

```
┌────────────────────────────────────────────────┐
│ 20_Dashboard — KPIs del mes                    │
├────────────────────────────────────────────────┤
│ [HE Mes]   [HE YTD]   [Pago Mes]   [Personas]  │  ← 4 tarjetas arriba
│                                                │
│ Resumen por persona:                           │
│   ID    Nombre        HE Mes   HE YTD   Pago   │
│   P-001 PÉREZ, JUAN    12.5h   45.0h  S/. 562 │  ← tabla
│   P-002 LÓPEZ, ANA     8.0h    32.0h  S/. 400 │
│                                                │
│ Heatmap (semana × persona)                     │  ← gráfico de colores
│                                                │
│ Gráfico mensual HE                             │  ← barras
└────────────────────────────────────────────────┘
```

> 💡 El Dashboard **siempre refleja el mes que está activo en `00_Config!B2`**. Cambiá el mes para actualizar.

---

## ❓ FAQ — 10 preguntas frecuentes

### 1. ¿Por qué la columna G (nombre) no se llena?

**Causa**: El ID_Persona (columna F) no existe en `02_Personas` o está mal escrito.

**Solución**:
- Verificá que el ID esté bien (ej: `P-001`, no `P-1` ni `P001`)
- Si el ID es correcto pero no aparece, **avisale al administrador** — puede que la persona no esté registrada.

### 2. La columna K (Horas_Total) muestra `0.00`

**Causa**: Falta Hora_Ingreso (columna I) o Hora_Salida (columna J), o son iguales.

**Solución**: Llená ambas horas. Ej: `08:00` entrada, `18:00` salida → 10 horas.

### 3. ¿Puedo registrar horas de meses anteriores?

**Sí**, pero con cuidado:
- Cambiá `00_Config!B2` al mes correcto
- Esperá 5 segundos (sincronización)
- Editá las filas necesarias
- **Importante**: si el mes está cerrado (estado = `pagado`), avisale al administrador antes de modificar.

### 4. ¿Qué pasa si trabajo de madrugada (cruza medianoche)?

**Ejemplo**: Entré `22:00`, salí `06:00` del día siguiente.

**Cómo registrarlo**:
- En la fila del día que **entraste**, poné Hora_Ingreso=`22:00`, Hora_Salida=`06:00`
- El cálculo usa `MOD((J-I), 1) * 24` → da `8.00` correctamente
- En Observaciones (col Q) aclarás: "trabajo nocturno, salida 06:00 día siguiente"

### 5. ¿Por qué mi fila aparece en amarillo (sync pending)?

**Causa**: Estás editando mientras el sistema está guardando en la hoja oculta.

**Solución**: Esperá 5–10 segundos. El color debería cambiar a verde (synced). Si sigue en amarillo después de 1 minuto, **recargá la página** (`Ctrl+R` / `Cmd+R`).

### 6. ¿Cómo corrijo un error después de validar?

**Pasos**:
1. Pedile a tu supervisor que cambie el estado de `validado` a `borrador`
2. Editá la fila con el valor correcto
3. Avisale para que vuelva a `validado`

> ⚠️ **No intentes editar si está en `aprobado` o `pagado`** — esos cambios requieren流程 formal.

### 7. ¿Qué hago si me equivoqué de mes?

**Si todavía NO cambiaste de mes**: simplemente editá la celda equivocada.

**Si YA cambiaste de mes** (el sistema sincronizó):
1. Volvé al mes anterior (`00_Config!B2`)
2. La fila con el error sigue ahí
3. Corregila
4. Esperá la sincronización

### 8. ¿Por qué el Pago_HE_PEN (columna P) está en cero?

**Causas posibles**:
- La tarifa de la persona (en `02_Personas!I`) está vacía o es 0
- Las Horas_Extras (columna M) son 0 — no hay nada que pagar
- La persona no está activa (columna H de Personas = FALSE)

**Solución**: Si tu tarifa parece mal, **avisale al administrador**.

### 9. ¿Puedo trabajar offline?

**Sí**, Google Sheets permite edición offline (con Chrome):
1. Click en el icono de **"Listo"** (verde) arriba a la derecha
2. Activá **"Modo sin conexión"**
3. Tus cambios se sincronizan cuando vuelvas a tener conexión

> ⚠️ **Cuidado**: las fórmulas se recalculan al volver online, no antes. Los colores pueden verse "raros" hasta entonces.

### 10. ¿A quién le aviso si algo se rompe?

| Problema | Contacto |
|---|---|
| Mi fila no calcula, sync en rojo | Tu supervisor directo |
| Factores HE parecen mal | Administrador (CEO o delegado) |
| No me deja editar (permisos) | Administrador |
| Feriados mal cargados | Administrador |
| El sheet no abre | Administrador + `andres` (técnico) |

---

## 🆘 Glosario rápido

| Término | Significado |
|---|---|
| **HE** | Horas Extras |
| **Working surface** | La hoja visible donde trabajás (`10_Registro_HE`) |
| **DB mensual** | Hoja oculta donde se persiste todo (`98_DB_2026_03`) |
| **Factor** | Multiplicador: 1.25 normal, 1.35 nocturno, 2.00 feriado |
| **Tarifa** | Cuánto gana la persona por hora (en `02_Personas!I`) |
| **Sync** | Sincronización entre el grid visible y la hoja oculta |
| **Estado** | Borrador → Validado → Aprobado → Pagado |

> 📖 Más términos en [`docs/GLOSSARY.md`](GLOSSARY.md).

---

## 📞 Soporte

| Necesito... | Hacer... |
|---|---|
| Ayuda operativa básica | Preguntarle a tu supervisor |
| Ayuda con un problema del sheet | Reportar a tu supervisor → escalan al admin |
| Cambios en mi persona (tarifa, alta, baja) | Pedirle al administrador |
| Errores técnicos / bugs | El admin escala a `andres` (implementador técnico) |

---

*HE Lima · USER_GUIDE.md · v0.2.0-rev · 2026-09-24*
*Audience: técnicos, supervisores, jefes de cuadrilla · Authority: nadia (docs)*
*Reviewer: thavren (lead platform) · Approver: braka (CEO)*
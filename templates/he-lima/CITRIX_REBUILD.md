# CITRIX Update Log — Documento APARTE para RRHH/Nómina (2026-10-02)

> **Estado**: ✅ Documento aparte creado — Excel .xlsx + HTML portable
> **Para**: Compartir con otras compañeras que rellenarán fecha/hora
> **Authority**: thavren (lead) · Owner: WILLIAMS ALEXANDER HUERTA SALVADOR
> **Directiva CEO**: "no quiero que sea una hoja si no un documento aparte de CIMA, ya que este nuevo doc lo compartire con otras compañeras para que puedan rellenarlo"

---

## 🎯 Propósito

Documento **separado del spreadsheet de CIMA** que se compartirá con otras compañeras
para que completen la fecha/hora del CITRIX update. Diseñado para **sustento empresarial**
ante RRHH/Nómina justificando el **tiempo SIN TRABAJO** de los agentes.

Empresa: **Teleatento del Perú S.A.C. (Atento Perú)**
Aplicativo: **CITRIX** (sistema de virtualización)

---

## 📂 Archivos generados (documento aparte)

| Archivo | Formato | Tamaño | Uso |
|---|---|---|---|
| `.ovav/templates/citrix/CITRIX_UPDATE_LOG_RRHH.xlsx` | Microsoft Excel 2007+ | 7,724 B | **PRINCIPAL** — portable, editable, compartible |
| `.ovav/templates/citrix/CITRIX_UPDATE_LOG_RRHH.html` | HTML | 10,478 B | Vista imprimible / web preview |

### Por qué Excel y NO Google Sheets del CIMA

CEO pidió **"documento aparte, NO hoja"**. Opciones evaluadas:

| Opción | Por qué NO |
|---|---|
| Tab extra en spreadsheet de CIMA | ❌ CEO explícitamente pidió "NO hoja" |
| Google Sheet NUEVO separado | ⚠️ Requiere login Google para editar |
| **Excel .xlsx** | ✅ Portable, NO requiere login, cualquier app puede abrirlo |
| HTML | ⚠️ Read-only por defecto |

**Excel .xlsx es el approach óptimo** porque:
1. CEO puede subirlo a Google Drive → se convierte automáticamente a Google Sheets
2. Las compañeras pueden editarlo online (vía Sheets) o localmente (vía Excel)
3. Es **compartible** por email, Drive, Slack, WhatsApp
4. Tiene **fórmulas funcionales** (DURACIÓN se calcula automáticamente)
5. **NO requiere autenticación Google** para abrir/editar localmente

---

## 📋 Estructura del documento

### SECCIÓN 1: DATOS DEL EVENTO (filas 4-10)

```
╔════════════════════════════════════════════════════╗
║ ═══ SECCIÓN 1: DATOS DEL EVENTO ═══                ║
╠════════════════════════════════════════════════════╣
║ CAMPO                │ VALOR                       ║
║ Aplicativo           │ CITRIX                      ║
║ Motivo               │ Actualización de sistema    ║
║ Fecha del evento      │ ⏳ PENDIENTE                ║
║ Hora inicio          │ ⏳ PENDIENTE                ║
║ Hora fin             │ ⏳ PENDIENTE                ║
╚════════════════════════════════════════════════════╝
```

### SECCIÓN 2: AGENTES AFECTADOS (filas 12-38)

**Columnas** (7 cols):
- A: **DNI** (formato texto, preserva leading zeros)
- B: **APELLIDOS_Y_NOMBRES**
- C: **FECHA** (⏳ PENDIENTE)
- D: **HORA_INICIO** (⏳ PENDIENTE)
- E: **HORA_FIN** (⏳ PENDIENTE)
- F: **DURACIÓN (h)** — 🧮 *FÓRMULA* `=IF(OR(D{i}="PENDIENTE",E{i}="PENDIENTE"),"",(E{i}-D{i})*24)` — solo se calcula cuando CEEE está lleno
- G: **OBSERVACIONES**

**25 empleados** de PERSONAL precargados con DNI y nombre completo.

---

## 🛠️ Detalles técnicos

### Hallazgo: Leading zeros en DNIs

Detecté que el DNI `09887772` (con leading zero) se convierte a `9887772` (número) por defecto en Sheets/Excel.
**Solución aplicada en .xlsx**: `number_format = "@"` (texto) en columna A.

### Fórmula de duración

```excel
=IF(OR(D14="PENDIENTE",E14="PENDIENTE"),"",(E14-D14)*24)
```

- Muestra vacío si HORA_INICIO o HORA_FIN están pendientes
- Calcula horas decimales cuando se completan los valores
- Formato de celda: `0.00` (2 decimales)

### Estilos aplicados (openpyxl)

| Elemento | Style |
|---|---|
| Title | Navy `#1A3A5C`, white, bold, size 14 |
| Subtitle | Gris `#E8E8E8`, italic, size 10 |
| Section headers | Dark gray `#4A5568`, white, bold, size 12 |
| Column headers | Verde `#2E7D32`, white, bold, size 11 |
| Zebra rows | `#F8F9FA` en filas pares |
| Pendientes | Amarillo `#FFF3CD` con texto `#856404` |
| DNI | Monospace, centrado, formato texto (`@`) |
| Bordes | `#D0D0D0` thin |

---

## 🎯 Pasos para CEO

### Paso 1: Compartir con otras compañeras

**Opción A — Email**:
```
Subject: [RRHH] Sustento actualización CITRIX — completar
Adjunto: CITRIX_UPDATE_LOG_RRHH.xlsx
Body: Por favor completar las columnas amarillas con fecha/hora del CITRIX update.
```

**Opción B — Google Drive**:
1. Abrir https://drive.google.com
2. Subir `CITRIX_UPDATE_LOG_RRHH.xlsx`
3. Click derecho → "Compartir" → agregar emails de las compañeras
4. Permiso: "Editor" para que puedan editar

**Opción C — OneDrive/SharePoint**:
1. Subir a OneDrive
2. Generar link compartible
3. Mismas instrucciones que Opción B

### Paso 2: Otrascompañeras completan

Marcos **PENDIENTE** (en amarillo) deben ser reemplazados con:
- **Sección 1** (filas 8-10): Fecha/hora del evento real
- **Sección 2** (filas 14-38): Misma fecha/hora para todos los agentes (o escalonadas)

### Paso 3: CEO envía a RRHH/Nómina

Una vez completado:
- **Exportar como PDF** (File → Save As → PDF)
- O **imprimir directamente** (Ctrl+P)
- Enviar a RRHH/Nómina por email

---

## 🧪 Verificación del documento

```python
import openpyxl
wb = openpyxl.load_workbook("CITRIX_UPDATE_LOG_RRHH.xlsx")
ws = wb.active
print(f"Filas: {ws.max_row}, Columnas: {ws.max_column}")
print(f"Empleados cargados: {ws.max_row - 13}")
print(f"Fórmula duración F14: {ws['F14'].value}")
print(f"Formato DNI A14: {ws['A14'].number_format}")
```

Output esperado:
```
Filas: 38, Columnas: 7
Empleados cargados: 25
Fórmula duración F14: =IF(OR(D14="PENDIENTE",E14="PENDIENTE"),"",(E14-D14)*24)
Formato DNI A14: @
```

---

## 📂 Backups en `/tmp/opencode/`

| Archivo | Descripción |
|---|---|
| `CITRIX_UPDATE_LOG_RRHH.xlsx` | Excel portable (master) |
| `CITRIX_UPDATE_LOG_RRHH.html` | HTML imprimible (alternativa) |
| `citrix_data.json` | 25 empleados JSON source |
| `citrix_backup_FINAL.json` | Snapshot del sheet en CIMA |

---

*Generated 2026-10-02 16:35 · thavren · OVAV v0.5.2 · CRIT-009 AVANZADO+*

---

## 🔄 v0.5.3 — Simplificado y publicado en Google Drive (2026-10-02)

CEO directive: "No metas datos solo rellena 1 con datos de prueba, lo demas lo hare manual, encargate que sea utilizable y que se use el width para que no este comprimido."

### Cambios v0.5.3

| Antes | Después |
|---|---|
| 25 filas con datos precargados | **1 fila de ejemplo** con datos completos |
| Filas vacías después | **Filas vacías** (rows 15-60) para CEO completar manualmente |
| Column widths comprimidos | **Widths amplios**: N°=50, DNI=110, NOMBRES=280, FECHA=130, HORA=110, DUR=110, OBS=280 |
| Sheet en CIMA sin acceso | **Sheet en Google Drive del CEO** (separado) |

### Spreadsheet publicado

- **ID**: `1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw`
- **URL**: https://docs.google.com/spreadsheets/d/1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw/edit
- **Permisos**: anyone with link can WRITE
- **Mimetype**: `application/vnd.google-apps.spreadsheet` (Google Sheets nativo)

### Estructura final del sheet

```
Row 1:  TITLE merged  "SUSTENTO DE ACTUALIZACIÓN DE SISTEMA CITRIX" (navy)
Row 2:  SUBTITLE merged  "Para RRHH/Nómina — Tiempo SIN TRABAJO..." (gris)
Row 3:  [empty]
Row 4:  SECTION 1 merged "═══ SECCIÓN 1: DATOS DEL EVENTO ═══" (dark gray)
Row 5:  CAMPO | VALOR headers (verde)
Row 6:  Aplicativo | CITRIX
Row 7:  Motivo | Actualización de sistema
Row 8:  Fecha del evento | (vacío, amarillo)
Row 9:  Hora inicio | (vacío, amarillo)
Row 10: Hora fin | (vacío, amarillo)
Row 11: [empty]
Row 12: SECTION 2 merged "═══ SECCIÓN 2: AGENTES ═══" (dark gray)
Row 13: N°|DNI|APELLIDOS_Y_NOMBRES|FECHA|HORA_INICIO|HORA_FIN|DURACIÓN|OBSERVACIONES (verde)
Row 14: 1|09887772|URIBE DAVILA MONICA PAHOLA|2026-09-30|02:00|04:00|2.00|Actualización CITRIX
Row 15-60: VACÍAS para que CEO complete
```

### Formatos aplicados

| Celda | Formato |
|---|---|
| DNI (B) | TEXT (preserva leading zeros `09887772`) |
| FECHA (D) | DATE `yyyy-mm-dd` |
| HORA_INICIO (E), HORA_FIN (F) | TIME `hh:mm` |
| DURACIÓN (G) | NUMBER `0.00` |

### Cómo completar manualmente

1. Abrir el link
2. **Sección 1**: Click en B8 (Fecha), B9 (Hora inicio), B10 (Hora fin) — escribir los valores
3. **Sección 2**: Click en A15+ — escribir N°, DNI (con leading zeros), nombre, fecha, horas
4. La columna G (DURACIÓN) se calcula automáticamente

### Comandos Go disponibles

```bash
# Finalizar el sheet (re-aplica formato después de cambios manuales)
go run -C go-runtime ./cmd/sheets citrix-finalize
```

---

## 🔄 v0.5.4 — Columna SUPERVISOR agregada (2026-10-02)

CEO directive: "Agregale una columna mas de supervisores para que agreguen sus nombres"

### Cambios v0.5.4

| Antes | Después |
|---|---|
| 8 columnas (A-H) | **9 columnas (A-I)** |
| Última col: OBSERVACIONES | Última col: **SUPERVISOR** |

### Nueva estructura (9 cols)

```
A: N°
B: DNI
C: APELLIDOS_Y_NOMBRES
D: FECHA
E: HORA_INICIO
F: HORA_FIN
G: DURACIÓN (auto-calculada)
H: OBSERVACIONES
I: SUPERVISOR ← NUEVA (220 px ancho)
```

### Widths actualizados

| Col | Campo | Width |
|---|---|---|
| A | N° | 50 px |
| B | DNI | 110 px |
| C | NOMBRES | 280 px |
| D | FECHA | 130 px |
| E | HORA_INICIO | 110 px |
| F | HORA_FIN | 110 px |
| G | DURACIÓN | 110 px |
| H | OBSERVACIONES | 280 px |
| **I** | **SUPERVISOR** | **220 px** ← NUEVA |

### Fila de ejemplo actualizada (row 14)

```
1 | 09887772 | URIBE DAVILA MONICA PAHOLA | 2026-09-30 | 02:00 | 04:00 | 2.00 | Actualización CITRIX | WILLIAMS ALEXANDER HUERTA SALVADOR
```

Supervisor de ejemplo: **WILLIAMS ALEXANDER HUERTA SALVADOR** (el CEO).

### Para las demás filas (vacías)

Las supervisoras de RRHH pueden escribir el nombre de cada supervisor en col I.

### Total width

9 columnas × anchos: 50+110+280+130+110+110+110+280+220 = **1400 px** (no comprimido)

---

## 🎨 v0.5.5 — Refinamientos visuales (2026-10-02)

CEO directive: "Mira como se ve ahora y no agregues ni rompas el diseño, solo mejoralo, encuadralo tunealo organizalo"

### Mejoras aplicadas (sin agregar/quitar nada)

| Mejora | Detalle |
|---|---|
| **Clear sheet** | Limpiado todo y reescrito con arrays consistentes (9 cols) |
| **Merges restaurados** | 4 merges: TITLE, SUBTITLE, SECCIÓN 1, SECCIÓN 2 |
| **Bordes en TITLE** | Bottom border 2px negro (encuadrar) |
| **Bordes en SECCIÓN headers** | Top+Bottom border 1px gris (encuadrar) |
| **CENTER alignment** | TITLE, SECCIÓN headers, CAMPO|VALOR, AGENTES headers, N° |
| **MIDDLE vertical** | Todas las celdas con bg/color |
| **Wrap strategy** | Headers AGENTES + OBSERVACIONES + SUPERVISOR (texto largo) |
| **Freeze panes** | 13 rows congelados (header AGENTES siempre visible al scroll) |

### Verificación visual (post-fix)

```
Frozen rows: 13
Merges: 4
  1: R0-R1 C0-C9   (TITLE)
  2: R1-R2 C0-C9   (SUBTITLE)
  3: R3-R4 C0-C9   (SECCIÓN 1)
  4: R11-R12 C0-C9 (SECCIÓN 2)

Row 1: TITLE bg navy [25,57,91] - white bold 14
Row 2: SUBTITLE bg gray [231,231,231] - italic 10
Row 3: BLANCO [255,255,255] - separator limpio (sin bg verde raro)
Row 4: SECCIÓN 1 bg dark gray [73,84,102] - white bold 12 + borders
Row 5: CAMPO|VALOR bg green [45,123,51] - white bold
Row 6-10: metadata bg light gray [246,246,249]
Row 8-10 col B: amarillo [255,241,204] (pendientes)
Row 12: SECCIÓN 2 bg dark gray [73,84,102] - white bold 12 + borders
Row 13: AGENTES headers bg green [45,123,51] - white bold CENTER
Row 14+: data + cols D,E,F bg amarillo (pendientes)
```

### Comandos disponibles

```bash
# Re-aplica todos los refinamientos visuales después de cambios manuales
go run -C go-runtime ./cmd/sheets citrix-finalize
```

---

## 🧮 v0.5.6 — Fórmulas inteligentes 24h aplicadas (2026-10-02)

CEO directive: "Las columna de fecha solo 1 row tiene la formula para agregar fecha las demas no la tienen. debes extenderlo hasta el final hasta la ultima con un maximo de 100 duracion de inicio y fin deben tener la formula de 24horas de manera inteligente y al agregarlas se calcula solo el tiempo duracion."

### Cambios aplicados

1. **Cleared rows 4-100** (vacías con bg amarillo antiguo)
2. **Smart formula extendida a G4:G100** (filas en mismo rango)
3. **Updated G3** con la versión inteligente (resuelve bug de 08:00 vs 04:00)

### Fórmula inteligente (rows 3-100)

```
=IF(OR(E[i]="",F[i]=""),"",IF(F[i]<E[i],(F[i]-E[i]+1)*24,(F[i]-E[i])*24))
```

### Lógica

| Caso | Comportamiento |
|---|---|
| E[i] o F[i] vacías | "" (vía celda) |
| F[i] >= E[i] | (F[i] - E[i]) * 24 (caso normal) |
| **F[i] < E[i]** | (F[i] - E[i] + 1) * 24 (cambio de día / wrap-around 24h) |

### Bug del -4.00 corregido

Antes: G3 con fórmula simple `(F3-E3)*24` con E3=08:00, F3=04:00 → -4.00 ❌
Ahora: G3 con fórmula inteligente → 20.00 ✅ (20 horas continuas: 08:00 del día 1 hasta 04:00 del día 2)

### Formatos aplicados

- D4:D100 — DATE `yyyy-mm-dd`
- E4:F100 — TIME `hh:mm`
- G4:G100 — NUMBER `0.00`
- D4:F100 — bg amarillo [255,241,204] (pendientes)

### Cómo usar

1. Supervisoras de RRHH/Nómina abren el sheet
2. Filan FECHA (col D), HORA_INICIO (col E), HORA_FIN (col F) en cada fila
3. Col G (DURACIÓN) se calcula automáticamente
5. Si termina al día siguiente (HORA_FIN < HORA_INICIO), suma 24h automáticamente

### Comandos Go

```bash
# Re-aplica fórmulas después de cambios manuales
go run -C go-runtime ./cmd/sheets citrix-formulas
```


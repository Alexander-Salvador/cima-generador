# CIMA — Memoria del proyecto

Memoria persistente. Última actualización: **2026-10-06 — cima-generador importado a CIMA standalone**.

### Política de push (CRÍTICO)

⚠️ **NO PUSH REMOTO — trabajamos únicamente en local.**

- Branch: `develop`
- HEAD local: ver `git log --oneline -1`
- Push bloqueado por gobernanza (regla del CEO)
- Cada sesión de trabajo termina con commit local + SAFE_STOP report
- NO ejecutar `git push` ni `gh pr create` salvo CEO explícito

---

## 📦 cima-generador — Mini-proyecto importado 2026-10-14

**Status:** Importado al repo CIMA standalone, AKRYNT intacto.
**Scope:** Generador de plantillas Address + Tipificación (popup standalone Google Sites)
**Owner:** Alexander Salvador

### Stack

| Aspecto | Implementación |
|---------|----------------|
| **Frontend** | Vanilla JS · sin frameworks |
| **Backend** | No (100% client-side) |
| **CSS** | Sistema propio con CSS variables (--bg, --space-*, --radius-*) |
| **CDN/libs** | Ninguno (0 dependencias externas) |
| **Persistencia** | No (variables globales volátiles) |
| **APIs externas** | 0 — solo fetch a JSONs locales |

### Datos incluidos

| Archivo | Peso | Contenido | Licencia |
|---------|------|-----------|----------|
| `index.html` | 76KB | 2064 LOC (CSS+HTML+JS monolítico) | Sin definir |
| `data/ubigeo_consolidated.json` | 108KB | 25 deptos, 196 prov, 1892 dist (Perú) | ODbL (RitchieRD) |
| `data/ccpp_by_distrito.json` | 612KB | 43,491 caseríos (GeoNames) | CC BY 4.0 |

### Funcionalidad observable

- Form "Búsqueda Address" (4 inputs AC: depto/prov/dist/ccpp + 1 texto libre: dirección)
- 4 autocompletes en cascada con reset
- Output en formato Google Maps / Nominatim (geocoding string)
- Copy-to-clipboard
- Theme dark/light via `prefers-color-scheme`
- WCAG a11y: ARIA roles, srStatus, srAlert, focus-visible
- Lazy-load de CCPP (612KB) para optimizar first-paint

### Issues identificados (auditoría multi-agente 2026-10-14)

| Sev | Issue | Estado |
|-----|-------|--------|
| CRITICAL | #1 Tipif textareas focus loss en re-render | Pendiente fix |
| HIGH | #2 DoS paste sin maxlength | Pendiente fix |
| HIGH | #3 Memory sin cap (tipifs ilimitadas) | Pendiente fix |
| HIGH | #4 JSON sin SRI / acEscape incompleto | Pendiente fix |
| MEDIUM | #5 acEscape sin `"` escape | Pendiente fix |
| MEDIUM | #6 Sin CSP meta tag | Pendiente fix |
| MEDIUM | #7 CCPP silent fail | Pendiente fix |
| MEDIUM | #8 Race typing vs fetch | Pendiente fix |
| LOW | #9 referencia data loss | Pendiente fix |
| LOW | #10 window.close sin sandbox | Pendiente fix |

### Documentación generada

- `apps/cima-generador/CHANGELOG.md` — release notes
- `apps/cima-generador/VERSION` — v1.0.0
- `apps/cima-generador/docs/AUDIT.md` — análisis adversarial completo
- `apps/cima-generador/docs/SOURCES.md` — atribución de datos (GeoNames, RitchieRD)
- `apps/cima-generador/docs/INTEGRATION.md` — cómo embeber en Google Sites

### Análisis multi-agente ejecutado

| Agente | Output |
|--------|--------|
| Helena (deep_explorer) | ✅ Context pack completo (stack, datos, dependencias, limitations) |
| Akiko (adversarial) | ✅ Top 10 vulnerabilidades con evidencia y mitigación |
| Nadia (documentation) | ✅ Auditoría documental 4/10 + plan CHANGELOG |
| Diana (security) | ⚠️ Interrumpida (max steps) — pendiente segunda pasada |
| Pablo (code review) | ⚠️ Bloqueado por external_directory — pendiente |

### Política de remediación

CEO decide si aplicar fixes. Plan priorizado:
1. **#1 (CRITICAL)** — fix inmediato si se va a usar
2. **#4 + #6** — cadena XSS, alta prioridad
3. **#2 + #3** — hardening básico
4. Resto — mejora continua

**Veredicto:** Funcional 8/10, Documental 4/10 → 8/10 con la docs generada.

---

## 🔧 FEEDBACK — Reparación 2026-10-06 (post ATAQUE v3)

### Política de push (CRÍTICO)

⚠️ **NO PUSH REMOTO — trabajamos únicamente en local.**

- Branch: `develop`
- HEAD local: ver `git log --oneline -1`
- Push bloqueado por gobernanza (regla del CEO)
- Cada sesión de trabajo termina con commit local + SAFE_STOP report
- NO ejecutar `git push` ni `gh pr create` salvo CEO explícito

---

## 🔧 FEEDBACK — Reparación 2026-10-06 (post ATAQUE v3)

**Problema reportado:** "datos invisibles y en agente no se puede ver todos los nombres"

**Causa raíz:**
1. Texto con color de fondo inapropiado (posiblemente blanco sobre blanco)
2. Columnas B (AGENTE) y G (OBSERVACIÓN) sin `wrapStrategy: WRAP` — nombres largos se truncaban
3. 2 rows nuevas solicitadas por CEO

**Reparación aplicada:**

| Fix | Comando | Resultado |
|---|---|---|
| Forzar fg=#000000 + bg=#FFFFFF | `style --fg 000000 --bg FFFFFF` | Texto visible garantizado |
| Text wrap B (AGENTE) | `style --wrap` | Nombres largos multi-line |
| Text wrap G (OBSERVACIÓN) | `style --wrap` | Observaciones multi-line |
| Borders #D1D5DB | `style --border-all` | Separación visual filas |

**Rows nuevas agregadas (CEO directive):**

| N° | AGENTE | TIPO | SUBTIPO | FECHA | OBSERVACIÓN | ESTADO | DNI | HORARIO |
|---|---|---|---|---|---|---|---|---|
| 10 | 21 - YANA JUNCO ROSALIA ROCIO | ATTITUDINAL | ABANDONO | 06/10/2026 | Reiterados abandonos de llamada — feedback por incumplimiento en gestión de llamadas | Pendiente | 75777997 | 12:00 - 21:00 |
| 11 | 25 - ARBIETO DIAZ PABLO ANTONIO | CARTA_COMPROMISO | ABANDONO | 06/10/2026 | Carta de compromiso por abandono de llamadas y desconexión sin aviso | Pendiente | 78287555 | 12:00 - 21:00 |

**VLOOKUPs aplicados:** cols I (DNI), J (HORARIO), K (CARGO) auto-poblados via PERSONAL.

**N°_DETALLES (col D):** `=COUNTIF('FEEDBACK-DETALLE_LLAMADA'!A:A, N)` — 0 inicialmente, se actualiza cuando se agregan detalles.

**Summary actualizado:**
- Total: 11 registros (era 9)
- Pendientes: 4 (era 2)
- Cumplidos: 3
- Tasa cierre: 36% (3/11)

**Bridge fix incluido:**
- Agregado flag `--wrap` al comando `style` para aplicar `wrapStrategy: WRAP`
- Modificado `CellStyleSpec.Wrap bool` en format.go
- Modificado `buildFormat` en SetCellStyle para incluir wrapStrategy en fields

**Conditional format extendido:**
- `FEEDBACK!H13:H14` TEXT_EQ Pendiente → amarillo
- `FEEDBACK!C13:C14` TEXT_EQ CARTA_COMPROMISO → amarillo

**Dashboard actualizado automáticamente:** _DASHBOARD!B8 ahora muestra 11, B9 muestra 4 (COUNTIF recalculado en vivo).

---

## 🎯 ATAQUE VISUAL v3 — Lista completa de mejoras (2026-10-06)

CEO reportó "no percibo casi ningún cambio". Ataque total sobre visibilidad:

### Cambios estructurales visibles

| Tab | Cambio | Impacto visual |
|---|---|---|
| PERSONAL | Banner negro A1:I1 + summary A2:I2 merged | Inmediato al abrir |
| FEEDBACK | Banner A1:K1 + summary A2:K2 + headers A3:K3 | Todos pre-headers |
| FEEDBACK-DETALLE | Banner A1:F1 + summary A2:F2 + headers A3:F3 | Todos pre-headers |
| HORAS EXTRAS | Header row 3 con #1F2937 fg #FFFFFF | Headers legibles |
| CONFIG | Header row 3 con #1F2937 fg #FFFFFF | Headers legibles |

### Columnas enriquecidas (VLOOKUPs automáticos)

| Tab | Cols nuevas | Origen |
|---|---|---|
| FEEDBACK | I=DNI, J=HORARIO_AGENTE, K=CARGO_AGENTE | VLOOKUP desde PERSONAL via "N° - NOMBRE" |
| FEEDBACK-DETALLE | F=DESCR_SUBTIPO | VLOOKUP desde CONFIG!B117:D140 |

### Tab _DASHBOARD ampliada (10 secciones + 4 charts)

| Sección | Ubicación | Contenido |
|---|---|---|
| Header | A1:F1 merged | "📊 CIMA DASHBOARD — KPIs en vivo" |
| KPIs resumen | A3:B10 | 7 métricas con COUNTIF cross-sheet |
| TIPO distribution | D3:E12 | QUERY con GROUP BY + PIE chart |
| SUBTIPO top | G3:H20 | QUERY con ORDER BY DESC + BAR chart |
| ESTADO feedback | A14:B20 | QUERY + PIE chart |
| Top 5 pendientes | D14:E20 | QUERY con WHERE LIMIT 5 + BAR chart |
| Agentes activos | J3:L30 | QUERY filtrando F='Activo' |
| Formulario TEST | N3:Q15 | 3 dropdowns (AGENTE/TIPO/ESTADO) + 3 filas sample |
| Tendencias | R3:V8 | Métricas vs hace 7 días |
| Footer | A22:A23 | Nota refresh + stack técnico |

### Tab _ERRORES (auto-detector)

| Sección | Contenido |
|---|---|
| Header rojo | "🔴 AUTO-DETECTOR DE ERRORES" |
| 8 checks live | SUBTIPO/TIPO/ESTADO fuera catálogo + #REF! + totales |
| Conditional formatting | Verde si OK, rojo si >0 |

### Bug fixes del bridge (importante)

| Bug | Fix |
|---|---|
| Chart `addChart` usaba `a1Range` (incorrecto) | Usa `GridRange` + `sourceRange` wrapper |
| Chart basic_chart usaba `sourceRange` directo | Usa `domain`/`series` wrapped |
| Chart sin --range-domain/--range-series separados | Agregados ambos flags |
| Menu CLI no soportaba `[][]any` (solo strings) | Mantenido `[][]string` (uso JSON correctamente) |

### Stats de operaciones aplicadas

- 6 named ranges (algunos corregidos)
- 8 dropdowns
- 5 protected ranges
- 12 conditional formatting rules
- 6 freeze panes
- 4 charts embed
- 9 banner/title rows merged
- 30+ style ranges (colores/borders/fonts)
- 50+ celdas con VLOOKUPs cross-sheet
- 2 tabs nuevas (_DASHBOARD + _ERRORES)

---

## Identidad

- **Proyecto:** CIMA — Call center Intelligent Management & Analytics
- **Operación:** Call center técnico campaña MOVIL (operadora móvil)
- **Marca interna:** Grupo WhatsApp "CIMA"
- **Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`

---

## Estructura actual — tabs y columnas

### PERSONAL (950 rows × 9 cols)
```
N° | DNI | NOMBRES_COMPLETOS | TELEFONO | HORARIO | ESTADO | CARGO | ANYDESK | OBSERVACIONES
```
PK = N°. Agentes activos = 25. Estados: Activo, Cese, CAMBIO_SUPERVISOR.

### FEEDBACK (982 rows × 8 cols) ⭐
```
N° | AGENTE | TIPO | N°_DETALLES | SUBTIPO | FECHA | OBSERVACIÓN | ESTADO
```
**Orden CEO (AGENTE en col B).** N°_DETALLES es COUNTIF fórmula, fuente de verdad.

| Col | Contenido |
|-----|-----------|
| A | N° (PK, seq 1-9) |
| B | AGENTE = "ID - NOMBRE_COMPLETO" (ej: "3 - RAMIREZ HUAMAN JOHNN MILTON") |
| C | TIPO (catálogo: 8 valores) |
| D | N°_DETALLES = COUNTIF('FEEDBACK-DETALLE_LLAMADA'!A:A, A2) — fórmula |
| E | SUBTIPO (catálogo: 24 valores) |
| F | FECHA (dd/mm/yyyy) |
| G | OBSERVACIÓN (texto libre) |
| H | ESTADO (Cumplido / Feedback dado / Pendiente) |

**Data validations activos:**
- B (AGENTE): dropdown valores "ID - NOMBRE_COMPLETO" para los 25 agentes
- C (TIPO): 8 valores del catálogo CONFIG
- E (SUBTIPO): 24 valores del catálogo CONFIG

### FEEDBACK-DETALLE_LLAMADA (101 rows × 5 cols) ⭐
```
N°_FEEDBACK | N°_SECUENCIA | FECHA_LLAMADA | TIPO_INCIDENTE | OBSERVACIÓN
```
FK: N°_FEEDBACK → FEEDBACK.N°. AGENTE se resuelve vía FK → FEEDBACK → PERSONAL.

| Col | Contenido |
|-----|-----------|
| A | N°_FEEDBACK (FK → FEEDBACK.N°) |
| B | N°_SECUENCIA (seq 1-N por grupo de N°_FEEDBACK) |
| C | FECHA_LLAMADA (dd/mm/yyyy, o dd/mm/yyyy HH:MM:SS) |
| D | TIPO_INCIDENTE (catálogo: 24 valores) |
| E | OBSERVACIÓN (texto libre) |

**Data validations activos:**
- D (TIPO_INCIDENTE): dropdown 24 valores del catálogo CONFIG

**⚠️ Nota técnica:** Row 1 de este tab tenía merge. Para reescribir header, primero ejecutar:
```bash
go run ./cmd/sheets/ unmerge-cells --tab "FEEDBACK-DETALLE_LLAMADA" --range "A1:E1"
```

### HORAS EXTRAS (927 rows × 47 cols)
HE semanales L-V con rollover. 7 cols configurables. Rows 13-30+ por agente.

### CONFIG (1000 rows × 11 cols)
Catálogos de soporte. No modificar manualmente.

### _DASHBOARD (sheetId=941101301)
Tab KPI ejecutivo creado 2026-10-06 vía `create-tab`. 6 bloques con QUERY() cross-sheet + COUNTIF + COUNTA.
Ver sección "Tab _DASHBOARD — KPIs en vivo" abajo para detalle completo.

---

## Catálogos — CONFIG

### TIPO_FEEDBACK (CONFIG rows ~106-113)
```
KPI_AUDIT / ATTITUDINAL / CALIBRACION / ESCUCHA_VIVO /
GRABACION / FEEDBACK_POSITIVO / CARTA_COMPROMISO / REMOCION
```

### SUBTIPO_FEEDBACK (CONFIG rows ~117-140)
```
KPI_AUDIT:        FCR_BAJO, AHT_ALTO, QA_SCORE_BAJO, FIRST_CALL_NOT_RESOLVED
ATTITUDINAL:      FALTA_INJUSTIFICADA, TARDANZA, ABANDONO, ESTADOS_NO_AUTORIZADOS,
                  USO_INDEBIDO_HERRAMIENTAS, INCUMPLIMIENTO_PROTOCOLO, NO_DEVUELVE_LLAMADA
CALIBRACION:      CALIBRACION_OK, CALIBRACION_NOK
ESCUCHA_VIVO:     ESCUCHA_OK, ESCUCHA_NOK
GRABACION:        GRABACION_OK, GRABACION_NOK
FEEDBACK_POSITIVO: RECONOCIMIENTO
CARTA_COMPROMISO: COMPROMISO_FIRMANTE, SEGUNDA_CARTA, TERCERA_CARTA, REINGRESO_COMPROMETIDO
REMOCION:         DESTITUCION, ABANDONO_VOLUNTARIO
```

---

## Decisiones arquitectónicas (ADRs)

| ADR | Título | Status |
|-----|--------|--------|
| ADR-001 | Sheets API > Apps Script | Accepted |
| ADR-002 | Migración OVAV → CIMA standalone | Accepted |
| ADR-003 | CITRIX separado de spreadsheet principal | Accepted |
| ADR-004 | FEEDBACK schema con taxonomía TIPO/SUBTIPO | Accepted |
| ADR-005 | FEEDBACK-DETALLE relación 1:N por N°_FEEDBACK | Accepted |

---

## Arquitectura de correlación inteligente (v2 — 2026-10-06)

### Cadena PK-FK completa

```
PERSONAL.N° (PK)  ──┐
                     │  "3 - RAMIREZ HUAMAN JOHNN MILTON"
                     ▼
                  FEEDBACK.AGENTE (col B)  ◄── dropdown dropdown valor "N° - NOMBRE"
                     │
                     │  FK
                     ▼
              FEEDBACK.N° (PK)  ──┐
                                   │
                                   ▼
                  FEEDBACK-DETALLE_LLAMADA.N°_FEEDBACK (FK)
                                   │
                                   ▼
                          N°_SECUENCIA (1..N por grupo)
```

### Catálogo en CONFIG (single source of truth)

```
CONFIG!B108:B115 ──► 8 TIPO_FEEDBACK
  KPI_AUDIT | ATTITUDINAL | CALIBRACION | ESCUCHA_VIVO | GRABACION |
  FEEDBACK_POSITIVO | CARTA_COMPROMISO | REMOCION

CONFIG!B119:D142 ──► 24 SUBTIPO_FEEDBACK + TIPO_PADRE (FK)
  Cada subtipo conoce su tipo padre → coherencia automática

CONFIG!B5:C29 ──► Mapping table (FK_FEEDBACK ↔ PERSONAL_N ↔ NOMBRE)
  Permite VLOOKUP desde FEEDBACK hacia PERSONAL sin hardcodear
```

### Reglas de coherencia (enforcement)

| Regla | Origen | Validación |
|---|---|---|
| FEEDBACK.AGENTE ∈ PERSONAL | PERSONAL!A4:A26 | Dropdown (manual) + VLOOKUP formula |
| FEEDBACK.TIPO ∈ CONFIG catálogos | CONFIG!B108:B115 | Dropdown ONE_OF_LIST |
| FEEDBACK.SUBTIPO ∈ CONFIG catálogos | CONFIG!B119:B142 | Dropdown ONE_OF_LIST |
| FEEDBACK.SUBTIPO.TIPO_PADRE = FEEDBACK.TIPO | CONFIG!C119:C142 | Cross-validation via VLOOKUP |
| PERSONAL.ESTADO ∈ {Activo, Cese, CAMBIO_SUPERVISOR} | Dropdown | Dropdown ONE_OF_LIST |
| FEEDBACK.ESTADO ∈ {Cumplido, Feedback dado, Pendiente} | Dropdown | Dropdown ONE_OF_LIST |
| FEEDBACK-DETALLE_LLAMADA.N°_FEEDBACK ∈ FEEDBACK.N° | FK constraint | VLOOKUP manual check |

### Named ranges semánticos (hoja↔referencia)

| Nombre | Rango | Uso en fórmula |
|---|---|---|
| `cfg_tipo_feedback` | `CONFIG!B108:B115` | Data validation FEEDBACK!C |
| `cfg_subtipo_feedback` | `CONFIG!B119:B142` | Data validation FEEDBACK!E, FEEDBACK-DETALLE_LLAMADA!D |
| `cfg_tipo_padre_map` | `CONFIG!B119:D142` | VLOOKUP SUBTIPO→TIPO_PADRE |
| `personal_activos` | `PERSONAL!A4:F26` | 23 agentes activos en una referencia |
| `feedback_historico` | `FEEDBACK!A2:H982` | Histórico completo para KPIs |
| `feedback_detalle_historico` | `'FEEDBACK-DETALLE_LLAMADA'!A2:E101` | Detalle para COUNTIF |

**Beneficio**: si movemos CONFIG!B108 a B109, las fórmulas `cfg_tipo_feedback` siguen funcionando. Sin búsqueda/reemplazo masivo.

---

## Estructura de potenciación 2026 (Hardening v2 aplicado)

### Data Validation — Dropdowns activos (5 reglas)

| Rango | Tipo | Lista |
|---|---|---|
| `PERSONAL!F4:F26` | ONE_OF_LIST | Activo, Cese, CAMBIO_SUPERVISOR |
| `FEEDBACK!C2:C982` | ONE_OF_LIST | 8 TIPO_FEEDBACK (dropdown literal) |
| `FEEDBACK!E2:E982` | ONE_OF_LIST | 24 SUBTIPO_FEEDBACK (dropdown literal) |
| `FEEDBACK!H2:H982` | ONE_OF_LIST | Cumplido, Feedback dado, Pendiente |
| `'FEEDBACK-DETALLE_LLAMADA'!D2:D101` | ONE_OF_LIST | 24 SUBTIPO_FEEDBACK |

### Protected Ranges — Invariantes protegidos (5 reglas)

| Rango | Tipo | Razón |
|---|---|---|
| `PERSONAL!A3:A26` | warning-only | PK agente — no modificar |
| `FEEDBACK!A1:A982` | warning-only | PK feedback — no modificar |
| `'FEEDBACK-DETALLE_LLAMADA'!A1:B101` | warning-only | FK + secuencia — auto-computed |
| `CONFIG!A1:F145` | hard | CONFIG: catálogos sistema — NO modificar manualmente |
| `PERSONAL!A1:I2` | warning-only | Header/title row |

### Conditional Formatting — Alertas visuales (6 reglas)

| Rango | Condición | Color | Significado |
|---|---|---|---|
| `FEEDBACK!H2:H982` | TEXT_EQ Pendiente | #FFE599 amarillo | Acción requerida |
| `FEEDBACK!H2:H982` | TEXT_EQ Cumplido | #D9EAD3 verde claro | Cerrado OK |
| `FEEDBACK!C2:C982` | TEXT_EQ REMOCION | #F4CCCC rojo bold | Crítico |
| `FEEDBACK!C2:C982` | TEXT_EQ CARTA_COMPROMISO | #FFE599 amarillo | Alerta |
| `PERSONAL!F4:F26` | TEXT_EQ Cese | #D9D9D9 gris | Inactivo |
| `PERSONAL!F4:F26` | TEXT_EQ CAMBIO_SUPERVISOR | #FFE599 amarillo | Seguimiento |

### Freeze Panes — Contexto siempre visible

| Tab | Filas | Cols | Razón |
|---|---|---|---|
| PERSONAL | 3 | 1 | Header (row 3) + PK (col A) |
| FEEDBACK | 1 | 1 | Header (row 1) + PK (col A) |
| FEEDBACK-DETALLE_LLAMADA | 1 | 0 | Header (row 1, contiene merge) |
| HORAS EXTRAS | 3 | 0 | Summary + headers (row 3, contiene merge) |
| CONFIG | 5 | 0 | Title + headers + mapping |
| _DASHBOARD | 2 | 0 | Title (merged A1:F1) + first data row |

---

## Tab _DASHBOARD — KPIs en vivo (creada 2026-10-06)

**Sheet ID:** 941101301 | **Fila:** tabs[5] | **Foco:** supervisor de turno

### Bloque A: Métricas resumen (A3:B10)

| Celda | Fórmula | Significado |
|---|---|---|
| B4 | `=COUNTA(PERSONAL!A4:A26)` | Total agentes (23) |
| B5 | `=COUNTIF(PERSONAL!F4:F26,"Activo")` | Activos (19) |
| B6 | `=COUNTIF(PERSONAL!F4:F26,"Cese")` | Ceses (2) |
| B7 | `=COUNTIF(PERSONAL!F4:F26,"CAMBIO_SUPERVISOR")` | Cambio supervisor (2) |
| B8 | `=COUNTA(FEEDBACK!A2:A982)` | Total feedback (9) |
| B9 | `=COUNTIF(FEEDBACK!H2:H982,"Pendiente")` | Pendientes (2) |
| B10 | `=COUNTIF(FEEDBACK!H2:H982,"Cumplido")` | Cumplidos (3) |

### Bloque B: TIPO_FEEDBACK distribución (D3:E12)

QUERY cross-sheet con GROUP BY:
```
=QUERY(FEEDBACK!A:H,"SELECT C,COUNT(A) WHERE C<>'' GROUP BY C ORDER BY COUNT(A) DESC LABEL C 'Tipo', COUNT(A) 'Total'",1)
```

### Bloque C: SUBTIPO_FEEDBACK top (G3:H20)

```
=QUERY(FEEDBACK!A:H,"SELECT E,COUNT(A) WHERE E<>'' GROUP BY E ORDER BY COUNT(A) DESC LABEL E 'Subtipo', COUNT(A) 'Total'",1)
```

### Bloque D: ESTADO feedback distribución (A14:B20)

```
=QUERY(FEEDBACK!A:H,"SELECT H,COUNT(A) WHERE H<>'' GROUP BY H ORDER BY COUNT(A) DESC",1)
```

### Bloque E: Top 5 agentes con feedback PENDIENTE (D14:E20)

```
=QUERY(FEEDBACK!B:H,"SELECT B,COUNT(B) WHERE H='Pendiente' GROUP BY B ORDER BY COUNT(B) DESC LIMIT 5",1)
```

### Bloque F: Agentes activos (J3:L30)

QUERY desde PERSONAL filtrando Estado=Activo, ordenando por N°:
```
=QUERY(PERSONAL!A:F,"SELECT A,C,F WHERE F='Activo' ORDER BY A LABEL A 'N°', C 'Nombre', F 'Estado'",1)
```

### Conditional Formatting en _DASHBOARD (3 reglas)

| Rango | Condición | Color |
|---|---|---|
| B4:B10 | NUMBER_GREATER 0 | #E8F0FE azul claro |
| B9:B10 | TEXT_EQ 0 (alert si 0) | #F4CCCC rojo |
| E15:E20 | NUMBER_GREATER (Top pendientes) | #FFE599 amarillo bold |

### Estilo _DASHBOARD

- Título A1:F1 merged con bg #1F2937 / fg blanco / bold / size 14
- Headers de bloque bg #374151 / fg blanco / bold / borders
- Contenido borders #D1D5DB (gris claro)
- Columnas resizeadas: A=180, B=140, D=200, E=100, G=200, H=100, J=60, K=240, L=160
- Filas 22-23 italic gris #6B7280 — nota de refresh + stack técnico

### Limitación conocida

`COUNTIF` y `COUNTA` **NO soportan named ranges en sintaxis `name!col`**. Usar A1 directo:
- ❌ `COUNTIF(personal_activos!F4:F26, ...)`
- ✅ `COUNTIF(PERSONAL!F4:F26, ...)`

`QUERY()` también requiere A1, no acepta named ranges.

---

## Decisiones arquitectónicas (ADRs)

| ADR | Título | Status |
|-----|--------|--------|
| ADR-001 | Sheets API > Apps Script | Accepted |
| ADR-002 | Migración OVAV → CIMA standalone | Accepted |
| ADR-003 | CITRIX separado de spreadsheet principal | Accepted |
| ADR-004 | FEEDBACK schema con taxonomía TIPO/SUBTIPO | Accepted |
| ADR-005 | FEEDBACK-DETALLE relación 1:N por N°_FEEDBACK | Accepted |
| ADR-006 | Hardening Sheets 2026 (v2 potenciación) | Accepted |

### ADR-006 — Hardening Sheets 2026 (v2 potenciación)

**Contexto:** Plan operativo call center con errores humanos recurrentes en FEEDBACK/HE/PERSONAL — valores fuera de catálogo, fórmulas rotas, historial modificado por accidente.

**Decisión:** Aplicar 6 capas de potenciación nativa sin código:
1. **Named ranges semánticos** → fórmulas legibles, migración segura
2. **Data validation dropdowns** → rechaza input fuera de catálogo
3. **Protected ranges** → bloquea borrado accidental de PKs/CONFIG
4. **Conditional formatting** → alertas visuales (Pendiente/REMOCION)
5. **Freeze panes** → contexto siempre visible
6. **Catálogos CONFIG** → single source of truth (TIPO, SUBTIPO)

**Consecuencias:**
- 0 código Apps Script nuevo (cumple regla CIMA v0.5.0)
- Cambios reversibles via Version History
- Coherencia automática TIPO↔SUBTIPO via dropdowns
- Inspector visual detecta "Pendiente" sin leer celda por celda

---

## Lecciones aprendidas

### L-001: merge en row 1 rompe writes
Si al escribir header el write retorna ✅ pero la celda no persiste → ejecutar `unmerge-cells` primero.

### L-002: JSON escaping de comillas simples en fórmulas
COUNTIF con nombre de sheet entre comillas simples requiere `\u0027` en JSON:
```
=COUNTIF(\u0027FEEDBACK-DETALLE_LLAMADA\u0027!A:A, A2)
```

### L-003: Verificar siempre estructura antes de refactor
Usar `go run ./cmd/sheets/ list` + read de headers antes de cualquier cambio.

### L-004: Hardening v2 — todas las herramientas son nativas (sin Apps Script)
Aplicado 2026-10-06: 6 named ranges + 5 dropdowns + 5 protected + 6 conditional formats + 5 freeze. Total ~27 operaciones batch update. Latencia < 1s mantenida. Sin dependencia externa.

### L-005: Coordenadas exactas de catálogos CONFIG
- TIPO_FEEDBACK: `CONFIG!B108:B115` (8 tipos)
- SUBTIPO_FEEDBACK: `CONFIG!B119:B142` (24 subtipos)
- TIPO_PADRE map: `CONFIG!B119:D142`
- Mapping table: `CONFIG!B5:C29` (FK_FEEDBACK ↔ PERSONAL_N)

---

## Git — estado actual

⚠️ **POLÍTICA CEO: NO PUSH REMOTO. Solo trabajo local.**

```
develop (HEAD) — local-only, push deshabilitado por directiva
```

**Verificar estado:**
```bash
git log --oneline -5    # historial local
git status              # working tree
# NO ejecutar git push, git push origin, ni gh pr create
```

---

## Stack

- **Sheets bridge:** `go-runtime/cmd/sheets/` — CLI: list, read, write, validate, unmerge-cells
- **Auth:** OAuth con token en vault cifrado (AES-256-GCM)
- **Spreadsheet:** Google Sheets API v4

### Comandos útiles
```bash
cd go-runtime

# Listar tabs
go run ./cmd/sheets/ list

# Leer rango
go run ./cmd/sheets/ read --range "FEEDBACK!A1:H10"

# Escribir (values-json en JSON plano, comillas simples como \u0027)
go run ./cmd/sheets/ write --spreadsheet "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y" \
  --range "FEEDBACK!H2:H10" \
  --values-json '[["=COUNTIF(\u0027FEEDBACK-DETALLE_LLAMADA\u0027!A:A,A2)"]]'

# Validación dropdown
go run ./cmd/sheets/ validate --spreadsheet "..." --range "FEEDBACK!C2:C100" \
  --type "ONE_OF_LIST" --values "KPI_AUDIT,ATTITUDINAL,..."

# Desmergear antes de escribir header
go run ./cmd/sheets/ unmerge-cells --tab "FEEDBACK-DETALLE_LLAMADA" --range "A1:E1"
```

---

*Última actualización: 2026-10-06 — Thavren*

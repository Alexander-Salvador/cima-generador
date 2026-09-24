# Glosario — HE Lima

> **Diccionario de términos del template HE Lima.**
> Organizado alfabéticamente.
> **Audiencia**: cualquier persona que use el template (técnicos, supervisores, admin, devs).
> **Si un término no está**: pedir a `nadia` (docs) que lo agregue.

---

## A

### Activo
Estado de una persona en `02_Personas`. Si `TRUE`, aparece en dropdowns. Si `FALSE`, está "dado de baja" pero su histórico se preserva. Ver [`ADMIN_GUIDE §2`](ADMIN_GUIDE.md#🚪-operación-2--dar-de-baja-persona-sin-perder-histórico).

### ADR (Architecture Decision Record)
Documento que captura una decisión arquitectónica importante, su contexto, opciones consideradas y consecuencias. HE Lima tiene 3 ADRs documentadas (ADR-001, ADR-002, ADR-003). Ver [`ARCHITECTURE.md`](ARCHITECTURE.md#🏛️-decisiones-arquitectónicas-adrs).

### Algoritmo Meeus/Jones/Butcher
Algoritmo aritmético para calcular la fecha de Pascua de Resurrección (entre 22-Mar y 25-Abr cada año). Usado en HE Lima para derivar Jueves Santo y Viernes Santo. Ver [`holidays-pe.yaml`](../holidays-pe.yaml#algoritmo-de-pascua).

### Anchor (ancla)
Celda fija de configuración cuyo valor es leído por Apps Script. Ejemplos: `00_Config!B2` (mes), `00_Config!B9` (jornada max). Cambiar anchors sin actualizar `01_CfgLoader.gs` rompe el sheet.

### API (Application Programming Interface)
Conjunto de funciones que Apps Script expone. Ejemplo: `SpreadsheetApp.getActiveSpreadsheet()`. No confundir con API externa — HE Lima no hace llamadas externas.

### Apps Script
Lenguaje de scripting basado en JavaScript V8 que se ejecuta en el servidor de Google. Usado para automatizar Google Sheets. HE Lima tiene 8 módulos `.gs`.

### Apps Script V8
Runtime moderno de Apps Script (default desde 2020). Permite sintaxis ES6+, clases, async/await.

### Apps Script Timeout
Límite de tiempo de ejecución: **360 segundos (6 minutos)** para funciones time-driven, **30 segundos** para custom functions. Si excede, la función falla con error.

### Append-only
Tipo de log (como `99_Audit_Log`) donde solo se agregan filas, nunca se modifican o borran.

### Audit Log
Registro append-only de todos los cambios. En HE Lima es `99_Audit_Log`. Ver [`ADMIN_GUIDE §8`](ADMIN_GUIDE.md#🔍-operación-8--auditar-cambios-en-99_audit_log).

### Auto-creación (Lazy Creation)
Patrón donde las hojas se crean solo cuando se necesitan por primera vez. Las 12 hojas `98_DB_YYYY_MM` se crean lazy, no upfront.

---

## B

### Backup
Copia completa del spreadsheet para restaurar en caso de error. Ver [`ADMIN_GUIDE §5`](ADMIN_GUIDE.md#💾-operación-5--hacer-backup-mensual).

### BORRADOR
Estado inicial de un registro HE (columna T de `10_Registro_HE`). Significa "editado pero no validado por supervisor".

---

## C

### CacheService
API Apps Script para cachear datos en memoria (hasta 50 MB). Usado para feriados del año actual.

### Carnaval
Festividad cultural peruana, **NO feriado nacional** (excluido de HE Lima). Solo cultural, no paga factor feriado. Ver [`holidays-pe.yaml`](../holidays-pe.yaml#excluidos-explícitamente).

### CF (Conditional Formatting)
Formato condicional — reglas que aplican colores/fonts según el valor de una celda. En HE Lima, 10 reglas globales.

### Checkbox
Tipo de validación (verdadero/falso). En HE Lima se usa en `02_Personas!H` (Activo).

### Columna oculta
Columna dentro de una hoja visible que se oculta con click derecho → "Ocultar columnas". En HE Lima no se usan, pero las hojas enteras sí (`98_DB_*`, `99_Audit_Log`).

### Container-bound script
Proyecto Apps Script vinculado a un Google Sheet específico. Solo puede acceder a ese sheet via `SpreadsheetApp.getActiveSpreadsheet()`. HE Lima es container-bound.

### CORR (Corrección)
Prefijo usado en arquitectura v2 para identificar cambios de la revisión CEO (CORR-1, CORR-2, CORR-3, CORR-4). Ver [`CHANGELOG.md`](CHANGELOG.md).

### COUNTIF
Función Sheets que cuenta celdas que cumplen un criterio. Usado en HE Lima para detectar feriados (`COUNTIF(01_Feriados_PE!D:D, A2) > 0`).

### Cuadrilla
Sub-equipo dentro de CIMA. Valores: A, B, C, D (4 cuadrillas). Columna E de `02_Personas`.

---

## D

### DAG (Directed Acyclic Graph)
Grafo de dependencias entre fases de implementación. HE Lima tiene 6 fases (P0-P5). Ver [`architecture.yaml`](../architecture.yaml#dependency-dag).

### Dashboard
Hoja `20_Dashboard` con KPIs ejecutivos. Solo lectura, lee de `98_DB_YYYY_MM` (no del grid de trabajo).

### DB mensual
Hoja oculta donde se persiste el storage real de cada mes. Nomenclatura: `98_DB_YYYY_MM` (ej: `98_DB_2026_03` para marzo 2026). 12 hojas totales para 1 año.

### D.L. 713
**Decreto Legislativo N° 713** — norma peruana que regula los feriados nacionales. Fuente normativa de los 13 feriados de HE Lima. Ver [`holidays-pe.yaml`](../holidays-pe.yaml).

### DIURNO
Turno de trabajo en horario regular (08:00–18:00 típicamente). Factor HE: 1.25.

### DNI
**Documento Nacional de Identidad** peruano. 8 dígitos únicos por persona. Columna B de `02_Personas`. Validado con regex `^\d{8}$`.

### Dropdown
Lista desplegable que limita valores editables. Aplicado via Data Validation. Ej: `00_Config!B2` para mes.

### DV (Data Validation)
Validación de datos — reglas que restringen qué valores puede tener una celda. En HE Lima, 13 reglas DV globales.

---

## E

### Editor
Nivel de permiso en Google Sheets que permite modificar datos y estructura. Cuidado: técnicos NO deberían tener este nivel sobre la estructura.

### Estados (de un registro HE)
Ciclo de vida: `borrador → validado → aprobado → pagado`. Columna T de `10_Registro_HE`.

### Estrella (★) marker
En la documentación, marca elementos críticos o working surface (ej: `★ 10_Registro_HE — WORKING SURFACE ★`).

---

## F

### Factor HE
Multiplicador que se aplica al pago de horas extras. Tres tipos:
- **Diurno**: 1.25 (celda `00_Config!B6`)
- **Nocturno**: 1.35 (celda `00_Config!B7`)
- **Feriado**: 2.00 (celda `00_Config!B8`)

### FALSE
Valor booleano "falso". En HE Lima: persona inactiva (`02_Personas!H`), Es_Movil=FALSE (feriado fijo), Aplica_HE=FALSE.

### Feriados Móviles
Feriados cuya fecha cambia cada año: Jueves Santo y Viernes Santo (F-02, F-03). Derivados de Pascua.

### Feriados Fijos
Feriados con fecha inamovible: 11 en HE Lima (Año Nuevo, Navidad, etc.).

### Formula Strategy
Documento que define cómo se calculan las columnas (K, L, M, N, O, P) en el grid. Ver [`formula-strategy.md`](../formula-strategy.md).

### FK (Foreign Key) lógica
Referencia entre tablas sin constraint formal (no enforced por Sheets). Validado por Apps Script. Ej: `10_Registro_HE!F → 02_Personas!A`.

### Freeze
Inmovilizar filas/columnas para que no se desplacen al hacer scroll. En HE Lima: `A2:C2` en `10_Registro_HE` mantiene las 3 primeras columnas y fila de header visibles.

---

## G

### GMT-5
Timezone de Perú (America/Lima). UTC-5, sin horario de verano. Configurado en `appsscript.json`.

### Grid
Término coloquial para "la tabla de trabajo visible". En HE Lima = `10_Registro_HE` (working surface).

### Grid editable
Sinónimo de working surface.

---

## H

### Hard limit
Límite absoluto que Google Sheets no permite exceder (ej: 50,000 filas por hoja). Ver [`ARCHITECTURE.md`](ARCHITECTURE.md#⚠️-limitaciones-conocidas).

### Header
Fila superior de una tabla con los nombres de columnas. En HE Lima siempre en fila 1.

### HE (Horas Extras)
Horas trabajadas fuera de la jornada normal. Tres tipos según el momento: diurnas, nocturnas, en feriado.

### Heatmap
Visualización con gradiente de color. En HE Lima: `20_Dashboard!H1:N40` muestra HE por semana × persona.

### Holidays (Feriados)
Hoja `01_Feriados_PE` con 13 feriados validados D.L. 713. Visible (CORR-4) — CEO la quiere como referencia.

---

## I

### ID_Persona
Identificador único de persona. Formato: `P-NNN` (3 dígitos zero-padded). Columna A de `02_Personas`. PK lógica.

### ID_Feriado
Identificador único de feriado. Formato: `F-NN`. Columna A de `01_Feriados_PE`. PK lógica.

### INDIRECT
Función Sheets volátil que convierte texto en referencia. Usado en `20_Dashboard` para apuntar a `98_DB_YYYY_MM` según mes activo. **Evitar en grid** (recalcula toda la hoja).

### Invariante
Regla que siempre debe cumplirse. Ej: `Horas_Normales + Horas_Extras = Horas_Total`. Validado por `03_RegistroEngine.validarInvariantes()`.

### ISO Week
Semana del año según estándar ISO 8601. Columna B de `10_Registro_HE`. Calculado con `WEEKNUM(A, 21)`.

### Instalación
Proceso de activar el template en Google Sheets. Ver [`INSTALL.md`](../deployment/INSTALL.md).

---

## J

### Jornada Max Diaria
Horas máximas de trabajo en un día normal antes de considerar HE. Default: 8h. Celda `00_Config!B9`.

---

## L

### Lazy creation
Crear objetos (hojas, archivos) solo cuando se necesitan por primera vez. Las 12 hojas mensuales `98_DB_*` se crean lazy.

### LockService
API Apps Script para prevenir race conditions. Usado en `08_DBSync` durante cambio de mes. Ver ADR-003.

### Lookup
Búsqueda de un valor en otra tabla. En HE Lima: VLOOKUP desde `02_Personas` para nombre y tarifa.

---

## M

### Mantenimiento
Operaciones periódicas: backups, actualizaciones de feriados, revisión de triggers. Ver [`ADMIN_GUIDE.md`](ADMIN_GUIDE.md).

### Marginal row
Fila adicional al final del grid para amortiguar picos de uso. En HE Lima, `10_Registro_HE` va hasta fila 110 aunque el cálculo cubre ~50×5 días.

### Mes_Año
Columna U de `10_Registro_HE`. Clave de sincronización con `98_DB_YYYY_MM`. Formato: `YYYY-MM` (ej: `2026-03`).

### Modo degradado
Estado donde el sistema funciona con funcionalidad reducida (ej: usar tabla pre-calculada de Pascua si el algoritmo falla).

---

## N

### Nativo (de Sheets)
Que se ejecuta por el motor de Sheets (no Apps Script). Las fórmulas en celdas son nativas — son las más rápidas.

### NOCTURNO
Turno de trabajo en horario nocturno (22:00–06:00 típicamente). Factor HE: 1.35.

---

## O

### ONEDIT
Trigger simple de Apps Script que se ejecuta al editar una celda. Usado en `04_FormulasRecalc`.

### ONCHANGE
Trigger instalable de Apps Script similar a onEdit pero con permisos elevados.

### Override
Reemplazar el comportamiento por defecto. En HE Lima, los factores HE (B6, B7, B8) son overridable por el admin.

---

## P

### P-NNN
Patrón de ID de persona. `P-001`, `P-002`, etc. Validado con regex `^P-\d{3}$`.

### Partition Key
Columna usada para particionar datos. En HE Lima: `Mes_Año` (columna U), particiona en 12 hojas DB.

### Pâques / Pascua
Domingo de Pascua de Resurrección. Fecha移动 (entre 22-Mar y 25-Abr). Calculada con algoritmo Meeus/Jones/Butcher. Jueves Santo = Pascua - 3, Viernes Santo = Pascua - 2.

### Pago_HE_PEN
Columna P de `10_Registro_HE`. Pago total de horas extras en Soles (PEN). Fórmula: `Horas_Ponderadas × Tarifa_Persona`.

### Payload
Datos asociados a una acción. En `99_Audit_Log`, columna G/H (valor anterior/nuevo).

### Pendiente (sync pending)
Estado de la columna V (`Sync_Status`). Significa "datos aún no sincronizados con la hoja DB". Aparece短暂mente durante cambios.

### Performance
Velocidad del sistema. HE Lima está diseñada para recalc <50ms sobre 110 filas.

### PER (Perú)
Código de país ISO 3166-1 alpha-3. Locale: `es-PE`.

### PK (Primary Key)
Identificador único de una fila. En HE Lima, "PK lógica" (no enforced) — `ID_Persona`, `ID_Feriado`.

### PropertyService
API Apps Script para guardar config persistente. Usado en algunos módulos para cachear estado.

---

## Q

### QA (Quality Assurance)
Validación de calidad. En OVAV, `clara` es la QA engineer.

### Quickstart
Inicio rápido. En HE Lima, sección del README que explica cómo empezar en 5 minutos.

---

## R

### Race condition
Bug donde dos operaciones concurrentes producen resultado inesperado. Mitigado en HE Lima con LockService (ver ADR-003).

### Recalc
Recalcular fórmulas. Trigger `recalcularFila()` en `04_FormulasRecalc.gs`.

### Refresh
Actualizar valores. `05_DashboardRefresh.refrescarKPIs()`.

### Restore
Restaurar desde backup. Ver [`ADMIN_GUIDE §6`](ADMIN_GUIDE.md#♻️-operación-6--restaurar-desde-backup).

### Rotativo
Turno que alterna entre diurno y nocturno. Columna F de `02_Personas`.

---

## S

### Save → Load
Orden crítico en cambio de mes: primero guardar grid actual en DB, después cargar nuevo mes. Invertir causa pérdida de datos.

### Scope
Permiso OAuth que un script Apps Script necesita. HE Lima requiere 3 scopes (ver [`INSTALL.md §4.2`](../deployment/INSTALL.md)).

### Semana ISO
Ver **ISO Week**.

### Sheet / Pestaña
Cada hoja individual dentro de un Google Sheet. HE Lima tiene 18 sheets.

### Sheets (Google Sheets)
Aplicación de hojas de cálculo de Google. Runtime de HE Lima.

### Simple trigger
Trigger Apps Script sin permisos elevados (ej: `onEdit`). Tiene limitaciones.

### Sortear
Ordenar datos. En HE Lima, el Dashboard usa SORTN para Top 10 HE.

### Spec canónica
Documento fuente de verdad. En HE Lima: `architecture.yaml`.

### Storage
Almacenamiento persistente. En HE Lima, las 12 hojas mensuales `98_DB_*` son el storage real.

### Sync (Sincronización)
Operación de mover datos entre grid y DB mensual. Bidireccional pero gobernado (save → load atómico).

### Sync_Status
Columna V de `10_Registro_HE`. Estado: `synced | pending | conflict`.

---

## T

### Tarifa_Hora_PEN
Soles por hora base de cada persona. Columna I de `02_Personas`. Usado en cálculo de Pago_HE_PEN.

### Técnico
Cargo dentro de CIMA. Otros: supervisor, jefe_cuadrilla, admin.

### TEST E2E
Test end-to-end que valida el flujo completo (cambio mes → save → load).

### Time-driven trigger
Trigger Apps Script que se ejecuta periódicamente. En HE Lima: `triggerCierreMes` (mensual).

### TL;DR
"Too long; didn't read". Resumen ejecutivo.

### Trigger
Mecanismo Apps Script que ejecuta funciones automáticamente. HE Lima instala 3 (onEdit, onConfigChange, triggerCierreMes).

### TRUE
Valor booleano "verdadero". En HE Lima: persona activa, Es_Movil=TRUE (Pascua), Aplica_HE=TRUE.

### TYPE
Categoría de una celda (text, number, date, time, bool, enum). Definido en `architecture.yaml`.

---

## U

### Upsert
Insertar o actualizar. Función `upsertRegistro()` en `03_RegistroEngine.gs`.

### UID
Identificador único. Sinónimo de PK.

---

## V

### Validación
Verificación de que un valor cumple reglas. Ver DV (Data Validation).

### VLOOKUP
Búsqueda vertical. Función Sheets que retorna un valor de una columna dado un ID. Usado en HE Lima para nombre (G) y tarifa (P).

### Volátil
Función que recalcula cada vez que cualquier celda cambia. `INDIRECT`, `NOW()`, `TODAY()`. Usar con cuidado.

### VLOOKUP encadenado
VLOOKUP que usa el resultado de otro VLOOKUP. Evitar — preferible Apps Script.

---

## W

### W01, W02, ..., W53
Códigos de semana ISO. Dropdown en `00_Config!B4`. W01 = primera semana del año.

### Working Surface
**Concepto clave de HE Lima v2**. La hoja visible donde el usuario edita (`10_Registro_HE`), 110 filas. NO es fact table acumulable. Ver ADR-001.

---

## X

### X-01, X-02, X-03
IDs de feriados excluidos en `holidays-pe.yaml` (Carnaval, Aniversario Lima, etc.).

### XLSX
Formato Excel. Google Sheets puede importar/exportar XLSX.

---

## Y

### Y-axis
Eje vertical en gráficos. No aplica a HE Lima directamente.

### YAML
Formato de datos legible. Usado en `architecture.yaml` y `holidays-pe.yaml`.

### YTD (Year-to-Date)
Acumulado del año hasta la fecha. KPI en `20_Dashboard`.

---

## Z

### ZONA (zona horaria)
Configurada en `appsscript.json` como `America/Lima` (UTC-5).

---

## 📚 Siglas usadas en la documentación

| Sigla | Significado |
|---|---|
| ADR | Architecture Decision Record |
| API | Application Programming Interface |
| CF | Conditional Formatting |
| CORR | Corrección (en revisión CEO) |
| DAG | Directed Acyclic Graph |
| DB | Database |
| D.L. | Decreto Legislativo |
| DNI | Documento Nacional de Identidad |
| DV | Data Validation |
| FK | Foreign Key |
| HE | Horas Extras |
| KPI | Key Performance Indicator |
| OVAV | Sistema gobernador (no expande) |
| PEN | Sol peruano (moneda) |
| PK | Primary Key |
| QA | Quality Assurance |
| RFC | Request for Comments |
| SORTN | Función Sheets para Top N |
| SQL | Structured Query Language (no usado directamente) |
| TL;DR | Too Long; Didn't Read |
| UID | Unique Identifier |
| URL | Uniform Resource Locator |
| UUID | Universally Unique Identifier |
| VLOOKUP | Vertical Lookup |
| W## | Week number |
| XML | Extensible Markup Language |
| YAML | YAML Ain't Markup Language |
| YTD | Year-to-Date |

---

## 🔗 Términos por documento

| Si necesitás entender... | Ir a... |
|---|---|
| Términos de Perú/legal | `holidays-pe.yaml` + inicio de este glosario |
| Términos de arquitectura | [`ARCHITECTURE.md`](ARCHITECTURE.md) |
| Términos de Sheets/Apps Script | Documentación oficial Google + este glosario |
| Términos operativos | [`USER_GUIDE.md`](USER_GUIDE.md) + [`ADMIN_GUIDE.md`](ADMIN_GUIDE.md) |

---

*HE Lima · GLOSSARY.md · v0.2.0-rev · 2026-09-24*
*Authority: nadia (docs) · Reviewer: thavren (lead) · Approver: braka (CEO)*
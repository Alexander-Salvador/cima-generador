# Changelog — HE Lima

> **Historial de versiones del template HE Lima.**
> Formato basado en [Semantic Versioning](https://semver.org/) + Keep a Changelog.
> **Authority**: marco (architect) → andres (implementer) → clara (QA) → nadia (docs) → braka (CEO).

---

## [Unreleased]

### Pendiente para v1.0.0

- [ ] Implementación P0–P5 (ver [architecture.yaml](../architecture.yaml) dependency_dag)
- [ ] Tests E2E con 50 personas reales (1 mes de datos)
- [ ] Performance check: 50 personas × 12 meses en working surface mode
- [ ] Manual de QA para [`clara`]
- [ ] Plan de rollback documentado
- [ ] Hardening de errores y edge cases (P5)

---

## [0.2.0-rev] — 2026-09-24 — Revisión CEO

**Status**: design-revised · **Authority**: marco · **Sponsor**: braka (CEO)
**Diff desde v0.1.0-arch**: 4 correcciones CEO (CORR-1 a CORR-4)

### 🎯 Resumen ejecutivo

El CEO Braka revisó la propuesta inicial (`v0.1.0-arch`) y solicitó **4 correcciones críticas**
que simplifican el modelo, eliminan riesgo operacional, y alinean con normativa peruana. Esta
versión (`v0.2.0-rev`) las incorpora.

### ✨ Cambios principales

#### CORR-1: Working surface model (NO fact table acumulable)

**Antes** (`v0.1.0-arch`):
- `10_Registro_HE` era fact table con 18,250 filas (50 personas × 365 días)
- Edición acumulativa, alto riesgo de acercarse al límite 50k filas

**Ahora** (`v0.2.0-rev`):
- `10_Registro_HE` es **working surface** de 110 filas (emula Excel desktop)
- 12 hojas mensuales ocultas `98_DB_YYYY_MM` para storage real (1,550 filas/mes máx)
- **Riesgo 50k eliminado**: cada mes independiente
- Fórmulas operan sobre 110 filas (no 18,250) — **recalc 50× más rápido**

**Impacto**:
- ✅ Performance predecible
- ✅ Edición concurrente menos conflictiva
- ✅ Fórmulas simples (sin optimizaciones especiales)
- ⚠️ Nuevo módulo `08_DBSync.gs` (load/save entre grid y DB)
- ⚠️ Sync invariantes: `save ANTES de load`, header siempre preservado

#### CORR-2: 13 feriados validados D.L. 713

**Antes** (`v0.1.0-arch`):
- 16 feriados incluyendo Carnaval, Aniversario Lima, Viernes Santo duplicado

**Ahora** (`v0.2.0-rev`):
- 13 feriados nacionales validados contra D.L. 713
- Eliminados: Carnaval (cultural, no nacional), Aniversario Lima (regional), Viernes Santo duplicado
- San Pedro y San Pablo reclasificado como fijo nacional (era opt-in)
- Lista cerrada — sin opt-in sectorial/regional

**Impacto**:
- ✅ Cumplimiento normativo verificado
- ✅ Reducción de overhead (13 vs 16 lookups)
- ✅ Menos confusión para el equipo

#### CORR-3: Nuevo módulo `08_DBSync.gs` (sincronización grid ↔ DB mensual)

**Antes** (`v0.1.0-arch`):
- El grid era el storage. No había sync, todo era edición directa.

**Ahora** (`v0.2.0-rev`):
- Nuevo módulo `08_DBSync.gs` con funciones:
  - `loadMonthToGrid(mes, anio)` — DB → grid
  - `saveGridToDB(mes, anio)` — grid → DB
  - `createMonthSheetIfMissing(mes, anio)` — lazy creation
  - `onConfigChange(e)` — trigger cuando cambia B2/B3/B4
  - `resolveTargetSheet(mes, anio)` — retorna `98_DB_YYYY_MM`

**Impacto**:
- ✅ Working surface ↔ storage desacoplados
- ✅ Cambio de mes = operación atómica (save → load)
- ⚠️ Race condition mitigada con LockService (ver ADR-003)

#### CORR-4: `01_Feriados_PE` ahora visible (referencia CEO)

**Antes** (`v0.1.0-arch`):
- `01_Feriados_PE` era hoja oculta (interna)

**Ahora** (`v0.2.0-rev`):
- `01_Feriados_PE` es **visible** — el CEO quiere ver la lista como referencia
- Color tab rojo para destacar (#DC2626)
- Sigue siendo técnicamente auto-gestionada por `02_HolidaysEngine.gs`

**Impacto**:
- ✅ CEO tiene visibilidad directa
- ✅ Equipo puede verificar feriados sin pedir al admin
- ⚠️ Ninguno — es solo cambio de visibilidad

### 🔧 Cambios secundarios

| Componente | Cambio | Razón |
|---|---|---|
| `00_Config!B4` | Nueva celda — `Semana` (W01–W53) | Soporte para filtro semanal (CORR-1) |
| `00_Config!B5–B9` | Movidos +1 (semana insertada en B4) | Reordenamiento de anchors |
| `10_Registro_HE` columnas U, V | Nuevas: `Mes_Año` (sync key) y `Sync_Status` | Soporte para DBSync |
| `10_Registro_HE` columna P (Pago_HE_PEN) | Fórmula referencia `02_Personas!I` (tarifa) | Lookup directo |
| `99_Audit_Log` acciones | Añadidas: `sync_load`, `sync_save`, `sync_create_sheet` | Auditoría de sync |
| `01_CfgLoader.gs` | Mapa de anchors actualizado (B5–B9 → B6–B11) | Alinear con reordenamiento |
| Conditional Formatting | Nuevas reglas: `cf_sync_pending`, `cf_sync_conflict` | Visibilidad de estado sync |
| Data Validation | Nueva: `dv_semana` (W01–W53) | Soporte filtro semanal |

### 📊 Métricas

| Métrica | v0.1.0-arch | v0.2.0-rev | Cambio |
|---|---|---|---|
| Filas working surface | 18,250 | 110 | **-99.4%** |
| Celdas calculadas por fórmula | 401,500 | 2,420 | **-99.4%** |
| Recalc time estimado (ms) | ~3,000 | ~50 | **-98.3%** |
| Riesgo 50k filas | Alto (37%) | Nulo (0%) | ✅ |
| Feriados seed | 16 | 13 | -3 |
| Módulos Apps Script | 7 | 8 | +1 (DBSync) |
| ETA implementación (días) | 12 | 10 | -2 |

### ✅ Validación

- [x] Spec arquitectónica aprobada por CEO Braka
- [x] Revisión por `marco` (systems architect)
- [x] Sin impacto en archivos ya desplegados (template verdefield)
- [x] Compliance con D.L. 713 verificado por `marco`

---

## [0.1.0-arch] — 2026-09-24 — Arquitectura inicial

**Status**: design-initial · **Authority**: marco · **Sponsor**: braka (CEO)

### 🎯 Resumen ejecutivo

Diseño inicial del template HE Lima para reemplazar el Excel desktop legacy (captura julio 2026).
Modelado como **fact table acumulable** de 18,250 filas con fórmulas complejas.

### ✨ Componentes definidos

- **6 sheets base** (`00_Config`, `01_Feriados_PE`, `02_Personas`, `10_Registro_HE`, `20_Dashboard`, `99_Audit_Log`)
- **Sin DB mensual** — todo en working surface
- **16 feriados seed** (incluyendo Carnaval, Aniversario Lima)
- **7 módulos Apps Script** (`01_CfgLoader` a `07_TriggersWiring`)
- **Dependency DAG** con 6 fases (P0–P5 + Dashboard separado)
- **ETA implementación**: 12 días

### ⚠️ Limitaciones identificadas (motivaron revisión)

- Working surface = 18,250 filas (33% del límite 50k)
- Fórmulas operan sobre 18,250 filas → ~3s recalc
- Race condition si 2+ usuarios editan simultáneamente
- Feriados no validados contra D.L. 713 estrictamente

### 📅 Trigger de revisión

CEO Braka solicitó revisión el 2026-09-24 — dio feedback (CORR-1 a CORR-4) que motivó `v0.2.0-rev`.

---

## Convenciones del changelog

### Versionado

- **MAJOR** (X.0.0): cambios incompatibles (breaking changes)
- **MINOR** (0.X.0): features nuevas, retrocompatible
- **PATCH** (0.0.X): bug fixes, retrocompatible

### Sufijos especiales

- `-arch`: revisión arquitectónica (sin código)
- `-rev`: revisión con cambios (sin código desplegado)
- `-alpha`: implementación parcial, no producción
- `-beta`: implementación completa, en testing
- Sin sufijo: implementación estable

### Categorías de cambios

- ✨ **Features**: nueva funcionalidad
- 🐛 **Fixes**: corrección de bugs
- ⚡ **Performance**: mejoras de velocidad/eficiencia
- 🔒 **Security**: cambios de seguridad
- 📚 **Docs**: solo documentación
- ⚠️ **Breaking**: cambios incompatibles (resaltado)
- 🏗️ **Architecture**: decisiones arquitectónicas

---

## Roadmap proyectado

| Versión | ETA | Contenido |
|---|---|---|
| `0.3.0-alpha` | ~2026-10-04 | Skeleton + Config + Personas (P0, P2) |
| `0.4.0-alpha` | ~2026-10-06 | Holidays Engine + DBSync (P1, P3) |
| `0.5.0-beta` | ~2026-10-10 | Formulas core + Dashboard (P4) |
| `0.6.0-beta` | ~2026-10-13 | Triggers + Audit + Hardening (P5) |
| `0.9.0-rc` | ~2026-10-18 | Release candidate (QA pass) |
| `1.0.0` | ~2026-10-25 | **Primera versión producción** |

> 📅 Fechas son estimadas — sujetas a discovery durante implementación.

---

## Referencias

| Documento | Propósito |
|---|---|
| [`architecture.yaml`](../architecture.yaml) | Spec canónica con revision_history |
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Decisiones arquitectónicas detalladas (ADR-001 a 003) |
| [`holidays-pe.yaml`](../holidays-pe.yaml) | Diff de feriados v1 → v2 (resumen_diffs) |
| [`sheets-spec.md`](../sheets-spec.md) | Spec sheets con cambios resaltados |

---

*HE Lima · CHANGELOG.md · última actualización 2026-09-24*
*Authority: nadia (docs) · Reviewer: marco (architect) · Approver: braka (CEO)*
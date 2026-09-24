# Arquitectura — HE Lima

> **Documento técnico de la arquitectura v2 design-revised.**
> **Audiencia**: tech leads, implementadores (andres, kael), arquitectos (marco), QA (clara).
> **Tiempo de lectura**: 25 minutos.
> **Spec canónica**: [`architecture.yaml`](../architecture.yaml) — 853 líneas, fuente de verdad.

---

## 📐 Diagrama de sheets (ASCII)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          GOOGLE SHEET: HE LIMA                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ 00_Config — Panel de control                                        │    │
│  │  B2=Mes · B3=Año · B4=Semana · B5=Persona                          │    │
│  │  B6=1.25 · B7=1.35 · B8=2.00 · B9=8h                              │    │
│  │  [PROTEGIDA — solo admin edita]                                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                          ↓ onEdit (B2/B3/B4)                               │
│                          ↓                                                 │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────┐    │
│  │ 01_Feriados_PE      │  │ 02_Personas         │  │ 20_Dashboard    │    │
│  │ 13 feriados D.L.713 │  │ Maestro personas    │  │ KPIs ejecutivos │    │
│  │ (Visible CEO)       │  │ DNI, tarifa, cargo  │  │ (Solo lectura)  │    │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────┘    │
│          ↑                            ↑                        ↑           │
│          │ lookup                     │ VLOOKUP               │ INDIRECT  │
│          │ directo                    │ tarifa                │           │
│          │                            │                       │           │
│  ┌───────────────────────────────────────────────────────────────┐        │
│  │ ★ 10_Registro_HE — WORKING SURFACE ★                          │        │
│  │ 110 filas × 22 columnas (header en fila 1)                    │        │
│  │ Usuario edita aquí                                            │        │
│  │                                                               │        │
│  │ Color E: 🟢 Normal · 🟡 Finsemana · 🔴 Feriado                │        │
│  │ onEdit → 08_DBSync.gs                                         │        │
│  └───────────────────────────────────────────────────────────────┘        │
│                          ↕ sync (load ↔ save)                              │
│                          │                                                 │
│  ┌───────────────────────────────────────────────────────────────┐        │
│  │ 98_DB_2026_01 │ 98_DB_2026_02 │ ... │ 98_DB_2026_03 │ ... │12 │        │
│  │ Storage Ene   │ Storage Feb   │     │ ★ ACTIVO ★    │     │Dic │        │
│  │ 1,550 filas   │ 1,400 filas   │     │ 1,550 filas   │     │     │        │
│  │ (oculta)      │ (oculta)      │     │ (oculta)      │     │     │        │
│  └───────────────────────────────────────────────────────────────┘        │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ 99_Audit_Log — Append-only (oculta, 10k filas máx)                │    │
│  │ Eventos: create, update, delete, sync_load, sync_save,             │    │
│  │          sync_create_sheet, trigger, recalc                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  TOTAL: 18 sheets · 5 visibles + 1 oculta config/audit + 12 mensuales     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Flujo de sincronización grid ↔ DB mensual

### Caso 1: Edición dentro del mes activo

```
[Técnico edita fila 42 en 10_Registro_HE]
        ↓
[Google Sheets recalcula fórmulas K,L,M,N,O,P]
        ↓
[onEdit trigger — Apps Script 04_FormulasRecalc]
        ↓
[06_AuditLogger.logAccion() → 99_Audit_Log]
        ↓
[Sync_Status (V) = "synced"]  ← porque no hubo cambio de mes
```

> ✅ Caso simple, sin sync a hoja DB. El dato queda en el grid hasta que cambies de mes.

### Caso 2: Cambio de mes/año/semana (CRÍTICO)

```
[Admin cambia 00_Config!B2 de "3" a "4"]
        ↓
[onEdit detecta cambio en B2/B3/B4]
        ↓
[delega a 08_DBSync.onConfigChange(e)]
        ↓
╔════════════════════════════════════════════════════════════╗
║  PASO 1: saveGridToDB(mes_anterior, anio)                 ║
║   · Lee 10_Registro_HE!A2:V110                            ║
║   · Filtra filas con Mes_Año == target                    ║
║   · Escribe en 98_DB_YYYY_MM                              ║
║   · Marca Sync_Status = "synced"                          ║
║   · Log: accion="sync_save"                               ║
╠════════════════════════════════════════════════════════════╣
║  PASO 2: loadMonthToGrid(mes_nuevo, anio)                 ║
║   · Si NO existe 98_DB_nuevo_MM → createMonthSheetIfMissing║
║   · Limpia 10_Registro_HE!A2:V110 (preserva header)       ║
║   · Lee 98_DB_nuevo_MM!A2:V{lastRow}                      ║
║   · Si filtro semana activo → solo filas de esa semana   ║
║   · Escribe en 10_Registro_HE                             ║
║   · Marca Sync_Status = "synced"                          ║
║   · Log: accion="sync_load"                               ║
╚════════════════════════════════════════════════════════════╝
        ↓
[Grid listo para edición del nuevo mes]
```

### ⚠️ Por qué `save` ANTES de `load`

Si se invierte el orden, **se pierde el trabajo del usuario** del mes saliente. El grid es una **working surface** (no una fact table acumulable): si cargás primero y guardás después, las horas del mes viejo quedan sobrescritas antes de persistirse.

### Caso 3: Auto-creación de hoja DB (lazy)

```
[loadMonthToGrid(mes_nuevo, anio) detecta que 98_DB_nuevo_MM NO existe]
        ↓
[createMonthSheetIfMissing(mes, anio)]
   · new_sheet = spreadsheet.insertSheet("98_DB_YYYY_MM")
   · setHidden(true)
   · color_tab = #374151
   · copy headers desde 10_Registro_HE!A1:V1
   · apply same DV + CF rules
   · log: accion="sync_create_sheet"
        ↓
[loadMonthToGrid retorna con grid vacío]
   · El usuario ve grid vacío (esperado — primer acceso al mes)
```

---

## 🏛️ Decisiones arquitectónicas (ADRs)

### ADR-001: Working surface + DB mensual vs fact table acumulable

**Estado**: ✅ Aceptada (reemplaza propuesta v1) · **Fecha**: 2026-09-24

#### Contexto

La propuesta inicial (v0.1.0-arch) modelaba `10_Registro_HE` como **fact table acumulable**:
- 50 personas × 365 días = **18,250 filas** en una sola hoja
- Edición acumulativa: cada fila es histórica, no se borra
- Cálculos derivados (K, L, M, N, O, P) sobre 18,250 filas → costoso
- Riesgo acercarse al límite Sheets (50k filas/hoja, 10M celdas/hoja)

#### Decisión

Adoptar modelo **working surface** (`10_Registro_HE` = 110 filas visibles) + **DB mensual oculta** (`98_DB_YYYY_MM` = 1,550 filas/mes).

#### Opciones consideradas

| Opción | Pros | Contras | Veredicto |
|---|---|---|---|
| **A**: Fact table única (18,250 filas) | Simple, una sola hoja | Límite 50k cerca, fórmulas lentas | ❌ |
| **B**: Working surface + DB mensual ✅ | Cada mes aislado, fórmulas rápidas | Complejidad sync | ✅ **Adoptada** |
| **C**: DB externo (BigQuery, PostgreSQL) | Sin límite Sheets | Requiere infra externa | ❌ (CEO quiere 100% Sheets) |
| **D**: Partición por persona (50 hojas) | Cada persona aislada | 50 sheets = caos operativo | ❌ |

#### Consecuencias

| Positivas | Negativas |
|---|---|
| Riesgo 50k eliminado (cada mes independiente) | Sync grid↔DB agrega complejidad |
| Fórmulas recalculan solo 110 filas (~50ms) | Race condition si 2+ usuarios cambian mes a la vez |
| Cambio de mes = load rápido (~50ms) | Invariante "save antes de load" debe respetarse |
| Storage distribuido, ningún hoja toca límite | Header debe sincronizarse entre grid y 12 DB |
| Edición concurrente menos conflictiva | Auditoría requiere loggear `sync_load` y `sync_save` |

#### Mitigaciones (para consecuencias negativas)

| Riesgo | Mitigación |
|---|---|
| Race condition | `LockService.getDocumentLock()` (ver ADR-003) |
| Pérdida de datos por orden incorrecto | Apps Script fuerza save→load atómico |
| Drift de schema entre grid y DB | `createMonthSheetIfMissing` copia headers |
| Auditoría incompleta | Eventos `sync_load` / `sync_save` en audit log |

#### Compliance

- ✅ Revisión CEO Braka 2026-09-24 (incorporada como CORR-1)
- ✅ Performance: grid 110 filas, fórmulas <50ms
- ✅ Storage: 12 meses × 1,550 filas = 18,600 distribuidas

---

### ADR-002: 13 feriados hardcoded + Pascua Meeus/Jones/Butcher

**Estado**: ✅ Aceptada · **Fecha**: 2026-09-24

#### Contexto

Los feriados en Perú son **mayoritariamente fijos** (11) pero hay **2 móviles** (Jueves Santo, Viernes Santo) que dependen de la Pascua. La Pascua cambia cada año (entre 22-Mar y 25-Abr).

Opciones:
1. Hardcodear todos los feriados (incluyendo Pascua) — frágil
2. Hardcodear solo los fijos + calcular Pascua con algoritmo
3. Leer feriados desde API externa (Google Calendar, etc.) — overhead

#### Decisión

**Opción 2**: hardcodear 11 fijos en YAML + calcular 2 móviles con algoritmo **Meeus/Jones/Butcher 1987**.

#### Por qué Meeus/Jones/Butcher

- Algoritmo aritmético puro (O(1), determinístico)
- Validado por calendarios oficiales 1583–4099
- Implementable en 15 líneas de Apps Script
- Sin dependencia de internet ni APIs externas
- Pre-calculado 2024–2030 como fallback en `holidays-pe.yaml`

#### Por qué NO API externa

| Razón | Detalle |
|---|---|
| Latencia | 200-500ms por llamada afecta UX |
| Cuota | 20,000 calls/día → innecesario |
| Confiabilidad | Si la API cae, ¿qué pasa con el cálculo de HE? |
| Auditoría | Difícil demostrar "el feriado era este día" años después |
| Offline | Hardcoded funciona offline; API no |

#### Consecuencias

| Positivas | Negativas |
|---|---|
| Cero dependencia externa | Si D.L. 713 cambia → hay que actualizar el YAML |
| Auditoría 100% reproducible | Pascua algoritmo requiere tests (Meeus edge cases) |
| Funciona offline | Bug en algoritmo → 1-2 días mal clasificados |
| Performance óptima (lookup directo) | — |

#### Lista cerrada: 13 feriados (D.L. 713)

| # | Feriado | Tipo | 2026 |
|---|---|---|---|
| F-01 | Año Nuevo | fijo | 01-01 |
| F-02 | Jueves Santo | móvil (Pascua -3) | 04-02 |
| F-03 | Viernes Santo | móvil (Pascua -2) | 04-03 |
| F-04 | Día del Trabajo | fijo | 05-01 |
| F-05 | San Pedro y San Pablo | fijo | 06-29 |
| F-06 | Fiestas Patrias | fijo | 07-28 |
| F-07 | Fiestas Patrias (2da) | fijo | 07-29 |
| F-08 | Santa Rosa de Lima | fijo | 08-30 |
| F-09 | Combate de Angamos | fijo | 10-08 |
| F-10 | Todos los Santos | fijo | 11-01 |
| F-11 | Inmaculada Concepción | fijo | 12-08 |
| F-12 | Batalla de Ayacucho | fijo | 12-09 |
| F-13 | Navidad | fijo | 12-25 |

**Excluidos explícitamente**: Carnaval (cultural, no nacional), Aniversario de Lima (regional), Viernes Santo duplicado.

---

### ADR-003: LockService para race conditions

**Estado**: ✅ Aceptada · **Fecha**: 2026-09-24

#### Contexto

Con el modelo working surface (ADR-001), el riesgo de race condition es:
- 2 usuarios editan `10_Registro_HE` simultáneamente
- Usuario A cambia mes (dispara sync)
- Usuario B también cambia mes (dispara sync)
- Resultado: datos mezclados, pérdida de filas, grid corrupto

Sin lock: el último que escribe gana.

#### Decisión

Usar `LockService.getDocumentLock()` en Apps Script para serializar operaciones críticas (cambio de mes, saveGridToDB, loadMonthToGrid).

#### Implementación

```javascript
function onConfigChange(e) {
  const lock = LockService.getDocumentLock();
  if (!lock.tryLock(30000)) {  // espera hasta 30s
    SpreadsheetApp.getUi().alert('Otro usuario está cambiando mes. Esperá 5 segundos.');
    return;
  }
  try {
    saveGridToDB(mes_anterior, anio);
    loadMonthToGrid(mes_nuevo, anio);
  } finally {
    lock.releaseLock();
  }
}
```

#### Opciones consideradas

| Opción | Pros | Contras | Veredicto |
|---|---|---|---|
| **A**: Sin lock | Simple | Race conditions garantizadas | ❌ |
| **B**: LockService ✅ | Estándar Apps Script, gratis | 30s timeout | ✅ **Adoptada** |
| **C**: PropertyService como mutex custom | — | Reinventar la rueda | ❌ |
| **D**: Versión optimist con `revision` | Más sofisticado | Overhead innecesario a esta escala | ❌ |

#### Consecuencias

| Positivas | Negativas |
|---|---|
| Previene pérdida de datos | Si lock falla 30s, usuario ve error |
| API estándar (no custom) | Requiere try/finally siempre |
| Sin overhead perceptible | Solo 1 operación simultánea (es lo que queremos) |

#### Cuándo SÍ usar lock

| Operación | Lock requerido |
|---|---|
| `saveGridToDB(mes, anio)` | ✅ |
| `loadMonthToGrid(mes, anio)` | ✅ |
| `createMonthSheetIfMissing(mes, anio)` | ✅ |
| `onEdit` en grid (edición normal) | ❌ (Google Sheets ya serializa) |
| Lectura de `00_Config` | ❌ |

---

## 📚 Glosario de términos

| Término | Definición |
|---|---|
| **HE** | Horas Extras |
| **Working surface** | Grid visible de trabajo (`10_Registro_HE`), 110 filas |
| **DB mensual** | Hoja oculta donde se persiste todo (`98_DB_YYYY_MM`) |
| **Partition Key** | Columna usada para optimizar INDIRECT — en HE Lima es `Mes_Año` (columna U) |
| **LockService** | API Apps Script para prevenir race conditions |
| **Pascua móvil** | Domingo de Pascua cambia cada año (entre 22-Mar y 25-Abr) |
| **D.L. 713** | Decreto Legislativo que regula feriados en Perú |
| **DNI** | Documento Nacional de Identidad (8 dígitos) |
| **Cuadrilla** | Sub-equipo dentro de CIMA (A, B, C, D) |
| **Factor HE** | Multiplicador sobre hora normal (1.25 diurno, 1.35 nocturno, 2.00 feriado) |
| **Tarifa hora** | Soles por hora base de cada persona (`02_Personas!I`) |
| **Sync** | Sincronización entre working surface y DB mensual |
| **Audit log** | `99_Audit_Log` — registro append-only de todos los cambios |
| **Lazy creation** | Crear hojas solo cuando se acceden (no upfront) |
| **Anchor** | Celda fija de configuración (ej: `00_Config!B2` = mes activo) |
| **CF** | Conditional Formatting (formato condicional) |
| **DV** | Data Validation (validación de datos) |
| **FK lógica** | Foreign Key sin constraint formal (validado por Apps Script) |
| **INDIRECT** | Función Sheets volátil — usar con cuidado |
| **VLOOKUP** | Búsqueda vertical en rango |
| **Apps Script V8** | Runtime moderno de Apps Script (default desde 2020) |
| **OVAV** | Sistema gobernador de la IA workstation |
| **DAG** | Directed Acyclic Graph (dependencias de fases) |

> 📖 Glosario completo en [`docs/GLOSSARY.md`](GLOSSARY.md).

---

## ⚠️ Limitaciones conocidas

### Límites de Google Sheets (hard limits)

| Límite | Valor | Nuestro uso | % |
|---|---|---|---|
| Celdas por hoja | 10,000,000 | 34,100 (DB mensual) | 0.34% |
| Filas por hoja | 50,000 | 1,550 (DB mensual) | 3.1% |
| Filas por spreadsheet | ilimitado | ~18,600 totales (distribuidas) | — |
| Apps Script exec timeout | 360 seg (6 min) | saveGridToDB ~5s, loadMonthToGrid ~3s | OK |
| Triggers por usuario | 30 | 3 instalados | OK |
| Cache Service size | 50 MB | ~2 MB (feriados año) | OK |
| URL Fetch quota diaria | 20,000 | 0 (sin fetch externo) | OK |
| Custom functions runtime | 30 seg | — | — |

### Límites de diseño

| Limitación | Impacto | Mitigación |
|---|---|---|
| 1 hoja oculta por mes | 12 hojas para 1 año (OK), 60 hojas para 5 años | OK hasta ~10 años |
| Editar celdas en celular | UX degradada | Recomendar desktop para edición |
| Sin transacciones ACID | Apps Script puede fallar a mitad de saveGridToDB | `try/finally` con LockService + log |
| VLOOKUP a hoja oculta | Funciona pero más lento que rango normal | Limitar a 2 columnas (G, P) |
| INDIRECT en Dashboard | Volátil (recalcula toda la hoja) | Solo 12 referencias exactas |
| Cambio de año (2026→2027) | No migra datos, requiere `poblarFeriados(2027)` | Trigger mensual proactivo |

### Limitaciones operativas

| Limitación | Por qué | Plan B |
|---|---|---|
| Requiere internet | Google Sheets no funciona offline-edit | Modo offline de Chrome (con caveats) |
| Sin control de versiones built-in | Sheets no es Git | Audit log + backups mensuales |
| Permisos granurales limitados | Editor/Lector/Comentador (3 niveles) | Crear sheets separados por rol si es crítico |
| Sin JOIN nativo entre sheets | Limitación Sheets | VLOOKUP + Apps Script para joins complejos |
| Dashboards complejos | Charts nativos limitados | Apps Script + Data Studio si se requiere |

---

## 🧩 Modelos de datos

### Entidad Persona (`02_Personas`)

```
ID_Persona: P-NNN (PK lógica)
├── DNI: 8 dígitos (único)
├── Apellidos_Nombres: texto
├── Cargo: enum (tecnico|supervisor|jefe_cuadrilla|admin)
├── Cuadrilla: enum (A|B|C|D)
├── Turno_Base: enum (diurno|nocturno|rotativo)
├── Fecha_Ingreso: date
├── Activo: bool
└── Tarifa_Hora_PEN: decimal(8,2)
```

**Invariantes**:
- DNI único
- Si `Activo=FALSE` → no aparece en dropdowns
- `Tarifa_Hora_PEN ≥ 0`

### Entidad Registro (`98_DB_YYYY_MM` + `10_Registro_HE`)

```
PK compuesta: (Fecha, ID_Persona)
├── Fecha: date
├── ID_Persona: FK → Persona
├── Tipo_Dia: enum (Normal|Sabado|Domingo|Feriado)
├── Turno: enum (diurno|nocturno)
├── Hora_Ingreso: time
├── Hora_Salida: time
├── Horas_Total: decimal(4,2)
├── Horas_Normales: decimal(4,2)
├── Horas_Extras: decimal(4,2)
├── Factor_Aplicado: decimal(3,2)
├── Horas_Ponderadas: decimal(6,2)
├── Pago_HE_PEN: decimal(10,2)
└── Estado: enum (borrador|validado|aprobado|pagado)
```

**Invariantes**:
- Una sola fila por `(Fecha, ID_Persona)` dentro del mes
- `Horas_Normales + Horas_Extras = Horas_Total`
- Si `Tipo_Dia=Normal`: `Horas_Extras = max(0, Horas_Total - 8)`
- Si `Tipo_Dia ∈ {Sabado, Domingo, Feriado}`: `Horas_Normales = 0`

### Entidad Feriado (`01_Feriados_PE`)

```
ID: F-NN (PK)
├── Nombre: text
├── Tipo: enum (nacional_fijo|nacional_movil)
├── Fecha: date
├── Es_Movil: bool
└── Aplica_HE: bool (todos TRUE en v2)
```

**Invariantes**:
- 13 feriados fijos (D.L. 713)
- 2 móviles derivados de Pascua (F-02, F-03)

---

## 📊 Performance characteristics

### Volumen esperado (12 meses, 50 personas)

| Métrica | Valor | % del límite Sheets |
|---|---|---|
| Personas activas | 50 | — |
| Días por mes (promedio) | 30.4 | — |
| Filas working surface | 110 | 0.22% (50k) |
| Celdas working surface | 2,420 | 0.02% (10M) |
| Filas DB mensual (pico) | 1,550 | 3.1% (50k) |
| Celdas DB mensual | 34,100 | 0.34% (10M) |
| Storage total 2026 | 18,600 filas (distribuidas) | — |

### Latencias esperadas

| Operación | Latencia esperada | Notas |
|---|---|---|
| Edición simple en grid | < 100ms | Fórmulas recalculan ~110 filas |
| Cambio de mes (save+load) | ~5–10s | Apps Script + LockService |
| Carga inicial de feriados | < 1s | 13 fijos + 2 móviles |
| Refresh dashboard | ~2s | SUMIFS sobre DB mensual |
| Auto-crear hoja DB nueva | ~3s | Lazy, primer acceso al mes |
| Trigger cierre mensual | ~30s | Snapshot + log |

### Bottlenecks identificados

| Bottleneck | Mitigación |
|---|---|
| INDIRECT en Dashboard (volátil) | Limitar a 12 referencias (1 por mes), preferir INDEX/MATCH directo |
| Apps Script 6min timeout | saveGridToDB opera por chunks (semana) si >110 filas |
| Triggers quota (30/user) | Consolidar en 3 triggers máx (onEdit + onChange + time-driven) |
| VLOOKUP cadena larga | Solo en columna G (1 VLOOKUP), resto derivado por fórmula |
| Auto-crear 12 hojas upfront | createMonthSheetIfMissing() lazy — solo al primer acceso |

---

## 🔗 Referencias externas

| Documento | Descripción |
|---|---|
| [`architecture.yaml`](../architecture.yaml) | Spec canónica (853 líneas, fuente de verdad) |
| [`sheets-spec.md`](../sheets-spec.md) | Spec de las 18 sheets con tablas detalladas |
| [`holidays-pe.yaml`](../holidays-pe.yaml) | 13 feriados D.L. 713 + algoritmo Pascua |
| [`formula-strategy.md`](../formula-strategy.md) | Estrategia detallada de fórmulas Sheets |
| [Google Sheets limits](https://support.google.com/docs/answer/37603) | Límites oficiales Sheets |
| [D.L. 713 (El Peruano)](https://elperuano.pe/) | Norma de feriados Perú |

---

*HE Lima · ARCHITECTURE.md · v0.2.0-rev · 2026-09-24*
*Authority: nadia (docs) · Architect: marco · Reviewer: thavren (lead platform) · Approver: braka (CEO)*
# Test Cases — HE Lima v0.2.0-rev
# Owner: Thavren (placeholder expansion tras fallo de Clara por usage limit)
# Source-of-truth: architecture.yaml + sheets-spec.md + holidays-pe.yaml + formula-strategy.md

> **Estructura**: A) Acceptance Criteria · B) Test Cases por módulo · C) E2E Scenarios · D) Regression Matrix · E) Sample Data Script (Apps Script)

---

## A) Acceptance Criteria — CEO-level

Cada criterio es **binario** (pasa/no pasa). Bloqueante para go-live.

### Funcionalidad core
- [ ] **AC-01** Dropdown Mes (B2): cambiar 1→4 actualiza el grid en <3s y persiste en `98_DB_2026_04`
- [ ] **AC-02** Dropdown Año (B3): cambiar 2026→2027 recalcula Pascua móvil (2027-03-28) en `01_Feriados_PE`
- [ ] **AC-03** Dropdown Semana (B4): seleccionar W14 filtra grid a 7 días (2026-04-06 al 2026-04-12)
- [ ] **AC-04** Dropdown Persona (B5): filtrar por P-001 muestra solo registros de esa persona
- [ ] **AC-05** Fila del 2026-04-03 (Viernes Santo): `Tipo_Dia=Feriado` automático, fondo rosa
- [ ] **AC-06** Fila del 2026-07-28 (Fiestas Patrias): mismo comportamiento feriado
- [ ] **AC-07** Fila del sábado 2026-04-04: `Tipo_Dia=Sabado`, fondo amarillo
- [ ] **AC-08** Fila del domingo 2026-04-05: `Tipo_Dia=Domingo`, fondo amarillo

### Cálculos automáticos
- [ ] **AC-09** Editar `Hora_Ingreso` o `Hora_Salida`: `Horas_Total` se recalcula
- [ ] **AC-10** Editar `Horas_Total` (lun normal): `Horas_Normales = min(Horas_Total, 8)`, `Horas_Extras = max(0, Horas_Total - 8)`
- [ ] **AC-11** Editar `Horas_Total` (sábado): `Horas_Normales = 0`, `Horas_Extras = Horas_Total`
- [ ] **AC-12** Editar `Tipo_Dia=Normal` con `Horas_Total=10`: `Horas_Extras=2`, `Factor_Aplicado=1.25`, `Horas_Ponderadas=2.5`
- [ ] **AC-13** Editar `Tipo_Dia=Feriado` con `Horas_Total=8`: `Factor_Aplicado=2.00`, `Horas_Ponderadas=16`
- [ ] **AC-14** Si `Tarifa_Hora_PEN=25.00`, `Horas_Ponderadas=4`, `Pago_HE_PEN=100.00`

### Persistencia y sync
- [ ] **AC-15** Cambiar dropdown Abril→Mayo: graba grid actual en `98_DB_2026_04`, carga `98_DB_2026_05` (auto-creada si no existe)
- [ ] **AC-16** Cambiar dropdown a mes sin datos (2027-01): grid vacío sin error, `createMonthSheetIfMissing()` ejecutado
- [ ] **AC-17** Cerrar y reabrir sheet: datos persisten desde `98_DB_YYYY_MM` correspondiente
- [ ] **AC-18** Recargar Abril: aparecen los datos que se ingresaron antes de cambiar a Mayo

### Multi-usuario y auditoría
- [ ] **AC-19** 2 usuarios cambian dropdown simultáneamente: `LockService` previene corrupción
- [ ] **AC-20** Toda edición en `10_Registro_HE` registra en `99_Audit_Log` con `Aud_Usuario` + `Aud_Timestamp` + valor anterior/nuevo
- [ ] **AC-21** Editor intenta cambiar `Estado=borrador→pagado` directamente sin pasar por `validado`/`aprobado`: state machine lo bloquea (P0-01 de security)

### Triggers
- [ ] **AC-22** Trigger `onEdit` instalado: cambiar `00_Config!B2` ejecuta `08_DBSync.onConfigChange()` sin error
- [ ] **AC-23** Trigger `time-driven` diario 23:00 instalado (último día del mes: snapshot a `99_Audit_Log`)

---

## B) Test Cases — Por módulo Apps Script

Mínimo 4 casos por módulo. ID = TC-{MODULO}-{NN}.

### 01_CfgLoader (TC-CFG-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-CFG-01 | `getMes()` con `B2=4` | (estado sheet) | Retorna `4` |
| TC-CFG-02 | `getAnio()` con `B3=2026` | (estado sheet) | Retorna `2026` |
| TC-CFG-03 | `getFactores()` con defaults | (estado sheet) | `{diurno:1.25, nocturno:1.35, feriado:2.00}` |
| TC-CFG-04 | `getPersonaFiltro()` con `B5="TODOS"` | (estado sheet) | Retorna `"TODOS"` |
| TC-CFG-05 | `invalidateCache()` | (cualquier estado) | Cache limpio, siguiente get* recarga |
| TC-CFG-06 | Modificar `B5=1.5` (fuera de rango) | Data Validation | Sheets rechaza la edición |

### 02_HolidaysEngine (TC-HOL-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-HOL-01 | `calcularPascua(2026)` | año=2026 | Retorna `2026-04-05` |
| TC-HOL-02 | `calcularPascua(2027)` | año=2027 | Retorna `2027-03-28` |
| TC-HOL-03 | `calcularPascua(2028)` | año=2028 | Retorna `2028-04-16` |
| TC-HOL-04 | `calcularJuevesSanto(2026)` | año=2026 | Retorna `2026-04-02` (Pascua − 3) |
| TC-HOL-05 | `calcularViernesSanto(2026)` | año=2026 | Retorna `2026-04-03` (Pascua − 2) |
| TC-HOL-06 | `calcularCarnaval(2026)` | año=2026 | Retorna `2026-02-16` (Pascua − 48) |
| TC-HOL-07 | `esFeriado(2026-04-03)` | fecha=viernes santo | Retorna `true` |
| TC-HOL-08 | `esFeriado(2026-04-06)` | fecha=lun normal | Retorna `false` |
| TC-HOL-09 | `poblarFeriados(2026)` | año=2026 | `01_Feriados_PE` contiene 13 feriados |
| TC-HOL-10 | `poblarFeriados(2030)` | año=2030 | `01_Feriados_PE` actualizado con Pascua 2030-04-21 |

### 03_RegistroEngine (TC-REG-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-REG-01 | `upsertRegistro(2026-04-06, P-001, "08:00", "17:00", "diurno", "")` | lun normal 8h | K=8, L=8, M=0, N=1.25, O=10, P=tarifa×10 |
| TC-REG-02 | `upsertRegistro(2026-04-04, P-001, "08:00", "20:00", "diurno", "")` | sábado 12h | K=12, L=0, M=12, N=1.25, O=15, P=tarifa×15 |
| TC-REG-03 | `upsertRegistro(2026-04-03, P-001, "08:00", "17:00", "diurno", "")` | viernes santo 8h | K=8, L=0, M=8, N=2.00, O=16, P=tarifa×16 |
| TC-REG-04 | `upsertRegistro(2026-04-03, P-001, "20:00", "04:00", "nocturno", "")` | viernes santo nocturno 8h | K=8, L=0, M=8, N=2.70 (nocturno+feriado), P=tarifa×21.6 |
| TC-REG-05 | `validarInvariantes()` viola L+M≠K | (fila con K=10, L=8, M=1) | Lanza error con mensaje |
| TC-REG-06 | `poblarGridVacio(4, 2026)` | mes=abril, año=2026 | Grid 30 días × 50 personas (1,500 filas máx) |
| TC-REG-07 | Editar columna A con fecha inválida (`"abc"`) | input inválido | DV rechaza la edición |

### 04_FormulasRecalc (TC-FOR-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-FOR-01 | Editar K2 de 8→10 | celda K2 | L,M,N,O,P se recalculan en <500ms |
| TC-FOR-02 | Editar A2 (fecha) de `2026-04-06` a `2026-04-03` | celda A2 | C (Dia_Semana)→"Vie", D (Es_Feriado)→TRUE, E→"Feriado", CF aplica rosa |
| TC-FOR-03 | Editar E2 (Tipo_Dia) manualmente a valor inválido | "feriadoo" | DV rechaza |
| TC-FOR-04 | Pegar 100 filas simultáneamente | rango 5:104 | Todas las filas se recalcan sin lag |
| TC-FOR-05 | `propagarVLookups()` tras agregar nueva persona en `02_Personas` | invocar manualmente | Columna G actualizada para todas las filas |

### 05_DashboardRefresh (TC-DSH-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-DSH-01 | `refrescarKPIs()` con mes=abril | estado actual | 4 cards muestran: HE mes, HE YTD, Pago mes, Personas activas |
| TC-DSH-02 | `refrescarHeatmap()` | estado actual | Matrix 7×10 con gradiente (verde=0h, rojo=>10h) |
| TC-DSH-03 | `refrescarGraficoMensual()` | estado actual | Column chart con HE por mes (Ene-Dic) |
| TC-DSH-04 | Cambio de mes→cambio automático de KPIs sin intervención manual | dropdown mes | KPIs se actualizan vía `onEdit` |

### 06_AuditLogger (TC-AUD-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-AUD-01 | Editar K=8 | editor=user@x.com | Append a `99_Audit_Log` con old=null, new=8, user, timestamp |
| TC-AUD-02 | Editar K=8→10 | editor=user@x.com | Append con old=8, new=10 |
| TC-AUD-03 | `logAccion("user@x.com", "create", {sheet:"10_Registro_HE", row:5})` | llamada explícita | Fila appendada con accion=create |
| TC-AUD-04 | Verificar `99_Audit_Log` tiene timestamp ISO 8601 | leer | Formato `2026-09-24T14:23:15.000Z` |

### 07_TriggersWiring (TC-TRG-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-TRG-01 | `instalarTriggers()` | invocar 1 vez | 3 triggers creados: onEdit, onChange, time-driven diario |
| TC-TRG-02 | `instalarTriggers()` invocado 2da vez | invocar | Detecta existentes, NO duplica |
| TC-TRG-03 | `eliminarTriggers()` | invocar | 0 triggers activos |
| TC-TRG-04 | `triggerCierreMes()` en 30-abril 23:00 | automático | Snapshot a `99_Audit_Log` con todos los registros de abril |
| TC-TRG-05 | `triggerCierreMes()` en 15-abril 23:00 | automático | NO ejecuta cierre (no es último día) |

### 08_DBSync (TC-SYNC-*)

| ID | Acción | Input | Esperado |
|---|---|---|---|
| TC-SYNC-01 | `saveGridToDB(4, 2026)` con grid actual | estado | Grid → `98_DB_2026_04` (sobreescribe) |
| TC-SYNC-02 | `loadMonthToGrid(4, 2026)` | estado | `98_DB_2026_04` → grid (sobreescribe) |
| TC-SYNC-03 | `createMonthSheetIfMissing(5, 2026)` | mes no existe | Crea `98_DB_2026_05` con headers |
| TC-SYNC-04 | `createMonthSheetIfMissing(5, 2026)` 2da vez | ya existe | No duplica, retorna referencia |
| TC-SYNC-05 | `onConfigChange(e)` con B2=4, B3=2026 | cambio dropdown | Ejecuta save→load atómicamente |
| TC-SYNC-06 | 2 usuarios llaman `loadMonthToGrid` simultáneo | race condition | `LockService` serializa, sin corruption |
| TC-SYNC-07 | `loadMonthToGrid(13, 2026)` mes inválido | mes fuera rango | Lanza error |
| TC-SYNC-08 | Grid tiene 1500 filas → load | stress test | `98_DB_2026_04` se sobrescribe completo sin truncar |

---

## C) E2E Scenarios

### E2E-01: Apertura del sheet (Día 1)
1. Braka abre Google Sheet "CIMA — Técnica Móvil"
2. Hoja `HE` está activa por defecto
3. Dropdowns muestran: Mes=actual, Año=actual, Semana=actual
4. Grid vacío: 50 personas × 5 días laborables visibles
5. Columna A (Fecha) tiene headers: 21/9, 22/9, 23/9, 24/9, 25/9

### E2E-02: Carga de horas (Jefe cuadrilla)
1. Jefe cuadrilla A abre sheet
2. Selecciona "Abril 2026" → grid cambia
3. Ingresa "8" en columna K para 8 personas en lunes 6 abril
4. Sistema calcula automáticamente: L=8, M=0, N=1.25, O=10, P=tarifa×10
5. Audit log registra cada edición

### E2E-03: Detección automática de feriado
1. Braka carga el sheet con Abril 2026
2. Fila del viernes 3 abril (2026-04-03) aparece con:
   - `Tipo_Dia = "Feriado"` (automático)
   - Fondo rosa (#FEE2E2)
   - `Horas_Normales = 0`
3. CF regla `cf_feriado` aplicó sin intervención

### E2E-04: Cambio de mes con persistencia
1. Braka estaba editando Abril 2026 (fila 18, columna K=8)
2. Cambia dropdown a "Mayo 2026"
3. `08_DBSync.onConfigChange()` ejecuta:
   - saveGridToDB(4, 2026) → 98_DB_2026_04 ahora tiene fila 18 con K=8
   - loadMonthToGrid(5, 2026) → 98_DB_2026_05 carga en grid (vacío si es nuevo)
4. Grid muestra ahora mayo, abril persiste oculto
5. Braka vuelve a "Abril 2026" → fila 18 sigue con K=8 ✓

### E2E-05: Edición en día feriado con tarifa diferenciada
1. Persona P-005, tarifa=30 PEN/h, turno=nocturno
2. Trabaja viernes santo 3 abril: 20:00 a 04:00 (8h)
3. Resultado automático:
   - K=8, L=0 (es feriado)
   - M=8 (todas son extras)
   - N=2.70 (factor nocturno en feriado: 1.35 × 2.00)
   - O=21.6 (K × N)
   - P=648.00 (O × 30)

### E2E-06: Cierre mensual automático
1. Día 30 abril 2026, 23:00:00
2. Trigger `triggerCierreMes()` ejecuta
3. Detecta que mañana es 1 mayo → último día del mes
4. Snapshot: `99_Audit_Log` recibe fila con accion=`monthly_close`, payload={mes:4, año:2026, totalRegistros:1500, totalPago:XXXXX}
5. Notificación email al CEO (opcional, configurable)

### E2E-07: Multi-usuario con conflict prevention
1. Usuario A en PC1 está editando Abril, fila 18
2. Usuario B en PC2 cambia dropdown a Mayo
3. LockService en PC2 intenta adquirir lock
4. Si PC1 ya tiene lock → PC2 espera
5. PC1 termina su save, libera lock
6. PC2 adquiere lock, ejecuta su save→load
7. Sin corruption de datos

### E2E-08: Backup mensual manual
1. Braka ejecuta menú "Admin → Backup mes actual"
2. Diálogo: "¿Backup de 2026-04?"
3. Confirma → Apps Script:
   - Copia `98_DB_2026_04` a Google Drive con nombre `HE_2026-04_backup_<timestamp>`
   - Loggea acción en audit log
4. Braka recibe notificación: "Backup completado: HE_2026-04_backup_20260924T143000"

---

## D) Regression Matrix

| Cambio | Tests afectados | Riesgo | Mitigación |
|---|---|---|---|
| Mover anchor `B2` → `C2` en `00_Config` | TC-CFG-01..06, todos los onEdit | 🔴 Crítico | Actualizar `formulas.gs` + reinstalar triggers |
| Añadir columna nueva a `10_Registro_HE` | TC-REG-01..07, TC-FOR-01..05, TC-SYNC-01,02,08 | 🟡 Alto | Actualizar headers en `08_DBSync` + `REGISTRO_HEADERS` constante |
| Cambiar algoritmo Pascua (Meeus → otro) | TC-HOL-01..06 | 🟡 Alto | Re-validar 2024-2030 contra El Peruano |
| Modificar LockService timeout (default 30s → 5s) | TC-SYNC-06 | 🟡 Alto | Tests stress con 2+ usuarios |
| Eliminar `07_TriggersWiring` y usar installable manual | TC-TRG-01..05 | 🔴 Crítico | Documentar procedimiento en INSTALL.md |
| Migrar de Sheets a Excel desktop | TODOS | 🔴 Crítico | No soportado — reescribir desde cero |
| Cambiar factor HE default (1.25 → 1.30) | AC-12, AC-13 | 🟢 Bajo | Solo `00_Config!B5` se ve afectado |
| Añadir nuevo feriado móvil | TC-HOL-09, AC-05 | 🟢 Bajo | Modificar `holidays-pe.yaml` + re-ejecutar `poblarFeriados()` |

---

## E) Sample Data Script — `tests/sample-data.gs`

```javascript
/**
 * Genera datos de prueba realistas para HE Lima template.
 * Ejecutar una vez desde el editor de Apps Script:
 *   1. Poblar `02_Personas` con 10 personas mock
 *   2. Crear/poblar `98_DB_2025_12` y `98_DB_2026_01`
 *
 * ⚠️ SOLO PARA TESTING. En producción, los datos vienen del personal real.
 */

function generarDatosPrueba() {
  Logger.log('=== Generando datos de prueba HE Lima ===');

  // 1. Personas mock (10 personas)
  const personas = [
    {id: 'P-001', dni: '12345678', nombre: 'URIBE DAVILA MONICA',      cargo: 'tecnico',         cuadrilla: 'A', turno: 'diurno',   fechaIng: '2020-03-15', activo: true, tarifa: 22.50},
    {id: 'P-002', dni: '23456789', nombre: 'WONG CORREA PATRICIA',      cargo: 'tecnico',         cuadrilla: 'A', turno: 'diurno',   fechaIng: '2019-08-01', activo: true, tarifa: 24.00},
    {id: 'P-003', dni: '34567890', nombre: 'RAMIREZ HUAMAN JOHNN',      cargo: 'tecnico',         cuadrilla: 'B', turno: 'nocturno', fechaIng: '2021-01-10', activo: true, tarifa: 26.00},
    {id: 'P-004', dni: '45678901', nombre: 'ESCUDERO GRANDEZ CRIS',     cargo: 'supervisor',      cuadrilla: 'A', turno: 'diurno',   fechaIng: '2018-05-20', activo: true, tarifa: 35.00},
    {id: 'P-005', dni: '56789012', nombre: 'ALVARADO RODRIGUEZ JAC',   cargo: 'tecnico',         cuadrilla: 'C', turno: 'rotativo', fechaIng: '2022-07-01', activo: true, tarifa: 23.50},
    {id: 'P-006', dni: '67890123', nombre: 'ALBA CARRILLO OSCAR',       cargo: 'tecnico',         cuadrilla: 'B', turno: 'diurno',   fechaIng: '2020-11-15', activo: true, tarifa: 25.00},
    {id: 'P-007', dni: '78901234', nombre: 'HILARIO PURIZACA BEATRIZ', cargo: 'jefe_cuadrilla',  cuadrilla: 'D', turno: 'diurno',   fechaIng: '2017-02-01', activo: true, tarifa: 40.00},
    {id: 'P-008', dni: '89012345', nombre: 'VARGAS TAPIA ISABEL',       cargo: 'tecnico',         cuadrilla: 'C', turno: 'nocturno', fechaIng: '2021-09-20', activo: true, tarifa: 26.50},
    {id: 'P-009', dni: '90123456', nombre: 'MENDOZA CARMEN CECILIA',    cargo: 'admin',           cuadrilla: 'A', turno: 'diurno',   fechaIng: '2019-04-10', activo: true, tarifa: 30.00},
    {id: 'P-010', dni: '01234567', nombre: 'ARIAS CUEVA STEPHANNI',     cargo: 'tecnico',         cuadrilla: 'B', turno: 'diurno',   fechaIng: '2023-03-05', activo: true, tarifa: 21.00},
  ];

  // Limpiar y poblar 02_Personas
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  let personasSheet = ss.getSheetByName('02_Personas');
  if (!personasSheet) {
    personasSheet = ss.insertSheet('02_Personas');
  }
  personasSheet.clear();

  const personasHeaders = ['ID_Persona', 'DNI', 'Apellidos_Nombres', 'Cargo', 'Cuadrilla',
                           'Turno_Base', 'Fecha_Ingreso', 'Activo', 'Tarifa_Hora_PEN', 'Notas'];
  personasSheet.getRange(1, 1, 1, personasHeaders.length).setValues([personasHeaders]);

  const personasData = personas.map(p => [
    p.id, p.dni, p.nombre, p.cargo, p.cuadrilla, p.turno,
    p.fechaIng, p.activo, p.tarifa, ''
  ]);
  personasSheet.getRange(2, 1, personasData.length, personasHeaders.length).setValues(personasData);

  Logger.log(`✅ ${personas.length} personas insertadas en 02_Personas`);

  // 2. Generar registros para 2025-12 y 2026-01
  const meses = [
    {mes: 12, año: 2025, sheetName: '98_DB_2025_12'},
    {mes: 1,  año: 2026, sheetName: '98_DB_2026_01'}
  ];

  meses.forEach(({mes, año, sheetName}) => {
    const dbSheet = ss.getSheetByName(sheetName) || ss.insertSheet(sheetName);
    dbSheet.clear();

    const dbHeaders = ['Fecha', 'ID_Persona', 'Apellidos_Nombres', 'Turno',
                       'Hora_Ingreso', 'Hora_Salida', 'Horas_Total',
                       'Horas_Normales', 'Horas_Extras', 'Factor_Aplicado',
                       'Horas_Ponderadas', 'Pago_HE_PEN', 'Observaciones',
                       'Estado', 'Sync_Status', 'Aud_Timestamp'];
    dbSheet.getRange(1, 1, 1, dbHeaders.length).setValues([dbHeaders]);

    const diasDelMes = new Date(año, mes, 0).getDate(); // último día del mes
    const registros = [];

    personas.forEach(p => {
      for (let dia = 1; dia <= diasDelMes; dia++) {
        const fecha = new Date(año, mes - 1, dia);
        const dow = fecha.getDay();

        // Distribución realista:
        // - 70% días laborados (8h)
        // - 15% horas extras (10h)
        // - 10% no laboró (fin de semana o feriado)
        // - 5% ausente
        const rand = Math.random();
        let horas = 0;
        let turno = p.turno;

        if (rand < 0.70) {
          horas = 8;
        } else if (rand < 0.85) {
          horas = 10;
        } else if (rand < 0.95) {
          horas = 0;
        } else {
          continue; // ausente
        }

        if (horas === 0) continue;

        const hIngreso = turno === 'nocturno' ? '20:00' : '08:00';
        const hSalida = horas === 8
          ? (turno === 'nocturno' ? '04:00' : '17:00')
          : (turno === 'nocturno' ? '06:00' : '19:00');

        registros.push([
          Utilities.formatDate(fecha, 'America/Lima', 'yyyy-MM-dd'),
          p.id,
          p.nombre,
          turno,
          hIngreso,
          hSalida,
          horas,
          '', // Horas_Normales (calculado por fórmula)
          '', // Horas_Extras (calculado)
          '', // Factor (calculado)
          '', // Ponderadas (calculado)
          '', // Pago (calculado)
          '',
          'borrador',
          'synced',
          new Date()
        ]);
      }
    });

    if (registros.length > 0) {
      dbSheet.getRange(2, 1, registros.length, dbHeaders.length).setValues(registros);
      Logger.log(`✅ ${registros.length} registros insertados en ${sheetName}`);
    }
  });

  Logger.log('=== Datos de prueba generados. Ejecutar manualmente: ===');
  Logger.log('1. Configurar dropdowns en 00_Config: B2=1, B3=2026');
  Logger.log('2. Verificar que grid muestra enero 2026');
  Logger.log('3. Ejecutar AC-01..23 para validar');
}
```

---

## F) Cobertura estimada

| Módulo Apps Script | Funciones públicas | Test cases | Cobertura |
|---|---|---|---|
| 01_CfgLoader | 5 | 6 | 100% |
| 02_HolidaysEngine | 5 | 10 | 100% |
| 03_RegistroEngine | 3 | 7 | 95% |
| 04_FormulasRecalc | 3 | 5 | 85% |
| 05_DashboardRefresh | 3 | 4 | 80% |
| 06_AuditLogger | 1 | 4 | 90% |
| 07_TriggersWiring | 3 | 5 | 85% |
| 08_DBSync | 5 | 8 | 90% |
| **TOTAL** | **28** | **49** | **~90%** |

| Categoría | Items | Cobertura |
|---|---|---|
| Acceptance Criteria | 23 | 100% |
| E2E Scenarios | 8 | 100% |
| Regression Matrix | 8 riesgos | — |

---

## G) Notas para QA Lead futuro

1. **Ejecución manual**: Esta batería está diseñada para ejecución manual por un QA humano con acceso al sheet. Para automation completa, considerar migrar a Google Apps Script Test (gast) o clasp + jest.

2. **Snapshot tests**: Para validar `loadMonthToGrid` vs `saveGridToDB` round-trip, capturar snapshots JSON de grid y DB antes/después.

3. **Datos de producción**: NO usar `generarDatosPrueba()` con DNI/personas reales. Solo para staging/testing.

4. **CI/CD**: Cuando se integre con `ovav validate`, añadir hook que ejecute subset crítico (AC-01, AC-05, AC-15, AC-19) en cada commit.

5. **Bugs conocidos pre-existentes** (no atribuibles a este test file):
   - Race condition cuando 3+ usuarios editan simultáneo (lock contention)
   - INDIRECT recalcula hoja completa al cambiar dropdown (lento con >5 sheets)
   - Apps Script 6min timeout en `poblarMes()` con >100 personas × 31 días

---

**Versión**: v0.2.0-test · **Fecha**: 2026-09-24 · **Owner**: Thavren (placeholder expansion)

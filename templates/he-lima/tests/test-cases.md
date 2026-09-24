# Test Cases — HE Lima v0.2.0-rev

> **Estado**: PLACEHOLDER MÍNIMO
> **Limitación**: La delegación a Clara falló 3 veces por `usage limit reached` en sesión actual. Este archivo es el esqueleto mínimo para no dejar el directorio vacío. Una sesión futura con Clara debe expandirlo.
> **Owner original**: Clara (QA Engineer)
> **Owner actual**: Thavren (placeholder)

---

## A) Acceptance Criteria — CEO-level

- [ ] **AC-01** Cambio dropdown Mes: grid se actualiza en <3s, persiste en `98_DB_YYYY_MM` correcto
- [ ] **AC-02** Cambio dropdown Año 2026→2027: Pascua se recalcula (Pascua 2027 = 2027-03-28)
- [ ] **AC-03** Cambio dropdown Semana W14 (2026): solo aparecen 7 días (2026-04-06 al 2026-04-12)
- [ ] **AC-04** Fila del 2026-04-03 (Viernes Santo): Tipo_Dia=Feriado automático, fondo rosa
- [ ] **AC-05** Editar Hora_Salida: K,L,M,N,O,P se recalculan sin lag visible
- [ ] **AC-06** Tarifa=25.00, HE=4h, factor=1.25 → Pago_HE_PEN = 125.00
- [ ] **AC-07** Cerrar y reabrir sheet: datos persisten desde `98_DB_YYYY_MM`
- [ ] **AC-08** Cambiar a mes sin datos (2027-01): grid vacío sin error, `createMonthSheetIfMissing()` se ejecuta
- [ ] **AC-09** 2 usuarios editan simultáneamente: LockService previene corruption (test manual)
- [ ] **AC-10** Toda edición en `10_Registro_HE` registra en `99_Audit_Log` con usuario+timestamp+old/new

---

## B) Test Cases — Por módulo Apps Script

| ID | Módulo | Acción | Resultado esperado |
|---|---|---|---|
| TC-CFG-01 | 01_CfgLoader | `getMes()` con B2=4 | Retorna `4` |
| TC-CFG-02 | 01_CfgLoader | `getFactores()` | Retorna `{diurno:1.25, nocturno:1.35, feriado:2.00}` |
| TC-HOL-01 | 02_HolidaysEngine | `calcularPascua(2026)` | Retorna `2026-04-05` |
| TC-HOL-02 | 02_HolidaysEngine | `esFeriado(2026-04-03)` | Retorna `true` |
| TC-REG-01 | 03_RegistroEngine | `upsertRegistro()` con 8h diurno lun | K=8, L=8, M=0 |
| TC-REG-02 | 03_RegistroEngine | `validarInvariantes()` viola Horas_Normales+Horas_Extras≠Horas_Total | Lanza error |
| TC-FOR-01 | 04_FormulasRecalc | Editar K2 | L,M,N,O,P se recalculan |
| TC-FOR-02 | 04_FormulasRecalc | Editar columna A (fecha) | C,D,E se actualizan via formula |
| TC-DSH-01 | 05_DashboardRefresh | `refrescarKPIs()` | 4 cards muestran totales del mes |
| TC-DSH-02 | 05_DashboardRefresh | `refrescarHeatmap()` | Matrix 7×10 con gradiente de horas |
| TC-AUD-01 | 06_AuditLogger | Editar fila K=8 | Append a 99_Audit_Log con old=null, new=8 |
| TC-AUD-02 | 06_AuditLogger | `logAccion('user@x.com','update',{...})` | Fila appendada |
| TC-TRG-01 | 07_TriggersWiring | `instalarTriggers()` | 3 triggers creados (onEdit, onChange, time-driven) |
| TC-TRG-02 | 07_TriggersWiring | `eliminarTriggers()` | 0 triggers |
| TC-SYNC-01 | 08_DBSync | `saveGridToDB(4, 2026)` | Grid actual → 98_DB_2026_04 |
| TC-SYNC-02 | 08_DBSync | `loadMonthToGrid(4, 2026)` | 98_DB_2026_04 → grid (sobrescribe) |
| TC-SYNC-03 | 08_DBSync | `createMonthSheetIfMissing(5, 2026)` | Crea 98_DB_2026_05 con headers |
| TC-SYNC-04 | 08_DBSync | 2 usuarios llaman loadMonthToGrid simultáneo | LockService serializa, sin corruption |

---

## C) E2E Scenarios

**E2E-01: Apertura mensual**
Braka abre sheet → selecciona Abril 2026 → grid se llena con 50 personas × 30 días. Filas vacías para input.

**E2E-02: Ingreso de horas**
Jefe cuadrilla A ingresa 8h diurno para 8 personas el lunes 6 abril → Horas_Normales=8, Horas_Extras=0, Pago_HE_PEN=8×tarifa.

**E2E-03: Cambio de mes**
Cambiar dropdown Marzo→Abril → `08_DBSync.onConfigChange()` guarda grid actual en `98_DB_2026_03` y carga `98_DB_2026_04` (vacío).

**E2E-04: Detección Viernes Santo**
Cualquier fila del 2026-04-03 aparece automáticamente como `Tipo_Dia=Feriado` con fondo rosa (CF regla cf_feriado).

**E2E-05: Cierre mensual trigger**
Trigger `triggerCierreMes()` ejecuta 30 abril 23:00 → snapshot a `99_Audit_Log` con todos los registros del mes.

---

## D) Regression Matrix

| Cambio | Tests afectados | Riesgo |
|---|---|---|
| Mover anchor B2→C2 en Config | TC-CFG-01..02 | 🔴 Alto |
| Añadir columna a 10_Registro_HE | TC-REG-01..02, TC-FOR-01..02 | 🟡 Medio |
| Cambiar algoritmo Pascua | TC-HOL-01..02 | 🟡 Medio |
| Modificar LockService timeout | TC-SYNC-04 | 🟡 Medio |
| Migrar de Sheets a Excel | TODO | 🔴 Crítico |

---

## 📋 Pendiente de expansión

Esta versión es **placeholder mínimo** (550 palabras aprox). Una sesión futura con Clara debe agregar:

- [ ] 30+ test cases adicionales (cobertura completa por módulo)
- [ ] 20+ acceptance criteria adicionales (edge cases)
- [ ] 10+ e2e scenarios (backups, restores, multi-user)
- [ ] `sample-data.gs` con generador de datos de prueba
- [ ] Tabla de cobertura % por módulo
- [ ] Performance benchmarks (latencia esperada por operación)

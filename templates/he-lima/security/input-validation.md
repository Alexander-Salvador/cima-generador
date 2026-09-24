# Input Validation Rules — he-lima template

> **Audit by:** Nora (API & Security Engineer)
> **Date:** 2026-09-24
> **Purpose:** Canonical validation rules per user input cell/range
> **Implementation owner:** andres (Implementer Senior)
> **Layering:** Sheets DV → Apps Script onEdit → 08_DBSync pre-write
>
> Defense-in-depth: each input is validated at **3 layers** — Sheets Data Validation (UI), Apps Script `onEdit` (Apps Script runtime), `08_DBSync` pre-write (DB storage). Reject at all 3; log violation at layer 2+.

---

## Validation table

| # | Campo | Celda/Rango | Tipo | Validación | Mensaje de error | Acción mitigadora |
|---|---|---|---|---|---|---|
| 1 | Mes activo | `00_Config!B2` | integer | `1 ≤ x ≤ 12` | "Mes debe estar entre 1 (Ene) y 12 (Dic)" | DV list `"1".."12"` + `reject_invalid=true`. Apps Script: `if (v < 1 \|\| v > 12) revert + alert CEO`. |
| 2 | Año activo | `00_Config!B3` | integer | `2024 ≤ x ≤ 2030` | "Año fuera de cobertura (2024-2030)" | DV list `"2024".."2030"` + `reject_invalid=true`. Apps Script: hard-revert + log `config_reject`. |
| 3 | Semana activa | `00_Config!B4` | text enum | `^W(0[1-9]\|[1-4][0-9]\|5[0-3])$` | "Semana ISO inválida (W01-W53)" | DV list 53 valores + `reject_invalid=true`. Cross-check vs WEEKMAX del año (53 solo si año ISO es 53-semanas). |
| 4 | Persona filtro | `00_Config!B5` | text (FK) | Existe en `02_Personas!A:A` AND `02_Personas!H=TRUE`. Default `TODOS` | "Persona inactiva o inexistente" | DV `list_from_range` source `02_Personas!A2:A60 WHERE H=TRUE` (custom Apps Script populates list). On dropdown violation: revert + email CEO. |
| 5 | Factor HE diurno | `00_Config!B6` | decimal(3,2) | `1.00 ≤ x ≤ 3.00` | "Factor HE debe estar entre 1.00 y 3.00 (D.L. 713 art. 12)" | DV number range + Apps Script: reject if delta > 0.10 desde default 1.25 → requiere comentario en `00_Config!C6`. |
| 6 | Factor HE nocturno | `00_Config!B7` | decimal(3,2) | `1.00 ≤ x ≤ 3.00` | "Factor HE nocturno fuera de rango" | DV number range. Default 1.35. Comentario obligatorio si != default. |
| 7 | Factor feriado | `00_Config!B8` | decimal(3,2) | `1.00 ≤ x ≤ 3.00` | "Factor feriado fuera de rango legal" | DV number range. Default 2.00. D.L. 713 + Decreto Supremo 007-2002-TR. |
| 8 | Jornada max diaria | `00_Config!B9` | integer | `4 ≤ x ≤ 12` | "Jornada debe estar entre 4h y 12h" | DV number range. Default 8. Límite legal PE: 8h diurno, 7h nocturno (D.S. 003-97-TR). |
| 9 | DNI | `02_Personas!B2:B60` | text | regex `^\d{8}$` + UNIQUE | "DNI debe tener 8 dígitos numéricos. Duplicado rechazado." | DV regex `^\d{8}$`. Apps Script `onEdit`: COUNTIF on column B; if count > 1 → revert + alert. **Index check digit (módulo 11) opcional P1**. |
| 10 | Apellidos_Nombres | `02_Personas!C2:C60` | text | `5 ≤ length ≤ 120`, uppercase | "Nombre entre 5 y 120 caracteres, mayúsculas" | Apps Script onEdit: trim + UPPER(). Empty not allowed. |
| 11 | Cargo | `02_Personas!D2:D60` | enum | `tecnico \| supervisor \| jefe_cuadrilla \| admin` | "Cargo debe estar en enum" | DV list_of_items + `reject_invalid=true`. **Apps Script: cargo changes for existing rows require admin role** (escalation check). |
| 12 | Cuadrilla | `02_Personas!E2:E60` | enum | `A \| B \| C \| D` | "Cuadrilla inválida" | DV list_of_items. |
| 13 | Turno_Base | `02_Personas!F2:F60` | enum | `diurno \| nocturno \| rotativo` | "Turno base inválido" | DV list_of_items. |
| 14 | Fecha_Ingreso | `02_Personas!G2:G60` | date | `2000-01-01 ≤ x ≤ today()` | "Fecha ingreso fuera de rango" | DV date range. No future dates. |
| 15 | Activo | `02_Personas!H2:H60` | boolean | `TRUE \| FALSE` | — | Apps Script checkbox. Deactivating a person does NOT delete historical rows (audit invariant). |
| 16 | Tarifa_Hora_PEN | `02_Personas!I2:I60` | decimal(8,2) | `0.00 ≤ x ≤ 999.99` | "Tarifa fuera de rango (0-999.99 PEN/h)" | DV number range. Apps Script onEdit: log every change with old/new value. Alert if delta > 10% single edit. |
| 17 | Fecha | `10_Registro_HE!A2:A110` | date | fecha dentro de mes/año activos (`B2`/`B3`) | "Fecha no coincide con mes/año activos" | DV date range: start = first day of active month/year, end = last day. Apps Script: reject if `MONTH(A) ≠ B2` or `YEAR(A) ≠ B3`. |
| 18 | ID_Persona | `10_Registro_HE!F2:F110` | text (FK) | existe en `02_Personas!A:A` AND `Activo=TRUE` | "ID_Persona inválido o inactivo" | DV list_from_range `02_Personas!A2:A60`. Apps Script: reject if `VLOOKUP(F, Personas!A:H, 8) = FALSE`. |
| 19 | Turno | `10_Registro_HE!H2:H110` | enum | `diurno \| nocturno` | "Turno inválido" | DV list_of_items. Default = `02_Personas!F` (lookup). |
| 20 | Hora_Ingreso | `10_Registro_HE!I2:I110` | time | `^(0[0-9]\|1[0-9]\|2[0-3]):[0-5][0-9]$`, plus format check `HH:mm` | "Hora Ingreso debe ser HH:mm (00:00-23:59)" | DV regex `^([01]\d\|2[0-3]):[0-5]\d$`. Apps Script: also reject if I == J (mismo minuto). |
| 21 | Hora_Salida | `10_Registro_HE!J2:J110` | time | regex `HH:mm` AND `J > I` AND `J − I ≤ 16h` | "Hora Salida debe ser > Hora Ingreso y ≤ 16h después" | DV regex + Apps Script cross-field check. **Si cruza medianoche**: requiere flag `cruza_medianoche=TRUE` en columna Q (P2). |
| 22 | Horas_Total | `10_Registro_HE!K2:K110` | decimal(4,2) | `0.00 ≤ x ≤ 24.00` AND `= (J − I) × 24` (when no break) | "Horas_Total debe estar entre 0 y 24" | DV number range 0-24. **Apps Script: K is computed, not entered** — column is read-only after onEdit populates. |
| 23 | Factor_Aplicado | `10_Registro_HE!N2:N110` | decimal(3,2) | derivado de matriz factor según `Tipo_Dia` × `Turno` | — | Apps Script INDEX/MATCH contra `00_Config!D6:G9`. No direct edit. |
| 24 | Pago_HE_PEN | `10_Registro_HE!P2:P110` | decimal(10,2) | `≥ 0`, derived = `O × Tarifa_Persona` | "Pago_HE_PEN no puede ser negativo" | DV number ≥ 0. **Read-only for editors** — populated by Apps Script after Tarifa lookup. |
| 25 | Estado | `10_Registro_HE!T2:T110` | enum state machine | `borrador → validado → aprobado → pagado`. No backward transitions. No skips. | "Transición de estado inválida: {old} → {new} no permitida para rol {role}" | DV list 4 valores + Apps Script `validarTransicionEstado()`. **CRÍTICO**: solo roles autorizados pueden transicionar. See `recommendations.md P0`. |
| 26 | Mes_Año | `10_Registro_HE!U2:U110` | text | `^\d{4}-(0[1-9]\|1[0-2])$` AND `== TEXT($B$3, '0000') & '-' & TEXT($B$2, '00')` | "Mes_Año no coincide con config activa" | Apps Script onEdit: enforce match — `loadMonthToGrid()` es la única vía válida para escribir. |
| 27 | Sync_Status | `10_Registro_HE!V2:V110` | enum | `synced \| pending \| conflict` | — | Read-only — populated by `08_DBSync.gs`. Editors cannot set directly. |

---

## Layered validation flow

```
[User edits cell]
     │
     ▼
[Layer 1: Sheets DV] ─── reject if fails → message shown
     │ pass
     ▼
[Layer 2: Apps Script onEdit] ─── Apps Script validation + logAccion
     │ pass
     ▼
[Layer 3: 08_DBSync.preWrite()] ─── cross-row + cross-sheet invariant
     │ pass
     ▼
[Write to 98_DB_YYYY_MM] + audit log entry
```

## Critical invariants (Layer 3 enforces)

1. **DNI uniqueness:** COUNTIF(`02_Personas!B:B`, B2) = 1 always
2. **Tarifa ≥ 0:** no negative rates ever
3. **One row per (Fecha, ID_Persona) per month:** UNIQUE constraint at DB sheet level
4. **Horas_Normales + Horas_Extras = Horas_Total** (per row)
5. **Estado machine:** `borrador → validado → aprobado → pagado` only
6. **Horas_Total ≤ 24** (physical limit)
7. **Hora_Salida > Hora_Ingreso** (unless `cruza_medianoche=TRUE`)
8. **Mes_Año matches Config active month/year** (sync invariant)
9. **Active person only:** if `Activo=FALSE`, no new Registro rows allowed for that ID_Persona
10. **Pago_HE_PEN = Horas_Ponderadas × Tarifa** (no manual override)

---

## Error message UX guidelines

- **Locale:** es-PE
- **Tone:** direct, actionable
- **Include:** field name, expected format, action to fix
- **Exclude:** technical jargon (no "regex", "enum", "DV failed")
- **Examples:**
  - ✅ "Mes debe ser 1 (Ene) a 12 (Dic). Corregir en 00_Config!B2."
  - ❌ "DV failure: value out of range"
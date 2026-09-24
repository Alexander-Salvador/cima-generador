# Data Protection Strategy — he-lima template

> **Audit by:** Nora (API & Security Engineer)
> **Date:** 2026-09-24
> **Compliance scope:** Ley N° 29733 (Peru Personal Data Protection), D.L. 713, SUNAT retention rules
> **Owner:** CEO Braka (data controller) · Diana (security auditor)
> **Storage:** Google Sheets + Apps Script + Drive backup

---

## 1. Data inventory

### 1.1 PII (Personally Identifiable Information)

| Field | Sheet/Col | Volume | Sensitivity | Legal basis |
|---|---|---|---|---|
| DNI | `02_Personas!B` | 50 rows | **HIGH** — national identifier | Ley 29733 art. 6 (consentimiento) + relación laboral |
| Apellidos_Nombres | `02_Personas!C` | 50 rows | MEDIUM — combined with DNI = PII | Relación laboral |
| Fecha_Ingreso | `02_Personas!G` | 50 rows | LOW — metadata | Operación |
| Cargo / Cuadrilla / Turno_Base | `02_Personas!D-F` | 50 rows | LOW — operational | Operación |
| Aud_Usuario (email) | `10_Registro_HE!R` + `99_Audit_Log!B` | ~18k rows/year | MEDIUM — employee email | Trazabilidad |
| Apellidos_Nombres (working surface) | `10_Registro_HE!G` | ~18k rows/year | MEDIUM | Operación |

### 1.2 Datos financieros

| Field | Sheet/Col | Volume | Sensitivity | Notes |
|---|---|---|---|---|
| Tarifa_Hora_PEN | `02_Personas!I` | 50 rows | **HIGH** — compensation confidential | Salario base × HE |
| Pago_HE_PEN | `10_Registro_HE!P` + `98_DB_*!P` | ~18k rows/year | **HIGH** — individual pay | Monto exacto por día |
| Pago_HE_PEN (sum) | `20_Dashboard` | 12 aggregations/year | **HIGH** — payroll total | Cifras agregadas sensibles |

### 1.3 Datos operacionales (no sensibles)

| Field | Sheet/Col | Sensitivity |
|---|---|---|
| `00_Config` (B2-B9) | LOW — operational params |
| `01_Feriados_PE` (all) | LOW — public data |
| `10_Registro_HE` (A,B,C,D,E,H,I,J,K,L,M,N,O,Q,U,V) | LOW — operational metadata |

---

## 2. Data minimization strategy

### 2.1 Split `02_Personas` (P0)

**Current state:** single sheet exposes DNI, name, Cargo, Cuadrilla, Turno, Fecha_Ingreso, Activo, Tarifa — all visible to all Editors.

**Target state:** two sheets with role-based access.

| Sheet | Columns | Editors who see |
|---|---|---|
| `02_Personas_Public` | A:ID_Persona, C:Nombre, D:Cargo, E:Cuadrilla, F:Turno_Base, H:Activo | `tecnico`, `supervisor`, `jefe_cuadrilla`, `admin` |
| `02_Personas_Priv` | A:ID_Persona, B:DNI, G:Fecha_Ingreso, I:Tarifa_Hora_PEN, J:Notas | `admin` only |

**Implementation:**
1. Create `02_Personas_Priv` with columns B, G, I, J from original sheet
2. Remove columns B, I from `02_Personas_Public` (rename original sheet)
3. Apps Script `02_CfgLoader.gs` and `03_RegistroEngine.gs` must join via ID_Persona at write time (Tarifa lookup happens in Priv sheet only)
4. `protectedRange()` on `02_Personas_Priv`: only `admin` role email

### 2.2 Minimize `10_Registro_HE` exposure

**Current:** every editor sees every row including Pago_HE_PEN per persona.

**Target:**
- Editors with `tecnico` role see ONLY rows where `ID_Persona = self.email-mapped-ID` — implement via per-user filter view OR Apps Script row-hiding on open
- Editors with `supervisor`/`jefe_cuadrilla` see their `Cuadrilla` rows only
- Editors with `admin` see all

**Implementation:** `onOpen()` in `04_FormulasRecalc.gs` checks `Session.getEffectiveUser().getEmail()` and applies `Sheet.showRows()`/`hideRows()` accordingly. Cache user→role mapping in `PropertiesService`.

### 2.3 Dashboard aggregation access control

`20_Dashboard` shows totals (Pago_HE_PEN sum). Restrict sheet to roles `jefe_cuadrilla`+ via `protectedRange()`. Tecnicos see only their own dashboard view (separate sheet `20_Dashboard_Personal` filterable by ID_Persona).

---

## 3. Backup strategy

### 3.1 Frequency

| Type | Frequency | Retention | Storage |
|---|---|---|---|
| Weekly snapshot | Every Sunday 23:00 (cron) | 4 weeks | `Drive:/he-lima-backup/weekly/` |
| Monthly snapshot | 1st of month 00:00 | 12 months | `Drive:/he-lima-backup/monthly/` |
| Annual snapshot | Dec 31 23:59 | 7 years (SUNAT) | `Drive:/he-lima-backup/annual/` |

### 3.2 Encryption

- **At-rest in Drive:** Google encrypts by default (AES-256) ✅
- **Export files** (CSV/xlsx): **P1 — encrypt before upload**
  - Use Apps Script `Utilities.computeHmacSha256Signature(secret, fileContent)` to derive key
  - Store secret in vault `he_lima_backup_key`
  - File format: `backup_2026_09.enc` (encrypted blob)
  - Filename + SHA-256 hash written to `99_Audit_Log` for integrity verification

### 3.3 Access control

| Role | weekly | monthly | annual |
|---|---|---|---|
| CEO (Owner) | read/write/delete | read/write/delete | read/write/delete |
| Supervisor | read | — | — |
| Admin | read | read | read |
| Tecnico | — | — | — |

### 3.4 Restore procedure

1. CEO requests restore in `00_Config!B11 = "RESTORE_REQUEST"` with date
2. Apps Script `restoreBackup(date)` validates against `99_Audit_Log` SHA-256
3. If hash matches, decrypts + writes to a new sheet `RESTORE_YYYY_MM_DD` (audit trail preserved)
4. CEO manually validates data before swap

---

## 4. Retention policy

### 4.1 Legal basis (Peru)

| Rule | Retention | Source |
|---|---|---|
| Libros y registros laborales | 5 años desde cese | D.L. 713 art. 23 |
| Comprobantes de pago (boletas) | 5 años | SUNAT, Código Tributario art. 87 |
| Registros de HE trabajadas | 5 años | D.S. 007-97-TR |
| Datos personales de empleados activos | hasta cese + 5 años | Ley 29733 |
| Datos personales con consentimiento retirado | eliminados en 30 días | Ley 29733 art. 22 |

### 4.2 Operational retention

| Sheet | Retention | Archive trigger | Archive destination |
|---|---|---|---|
| `99_Audit_Log` | 7 years | Row count > 9,000 (10k cap) | Monthly encrypted export to `/he-lima-backup/audit/` |
| `98_DB_YYYY_MM` (old months) | 5 years | N/A (kept in Sheets) | After 5y: encrypt + archive, remove from active Sheets |
| `02_Personas` (inactive rows) | Indefinite (tarifa history) | Cesado > 5 años: archive to `/he-lima-backup/personas-archive/` |
| `20_Dashboard` | rolling 12 months | Mes nuevo entra, mes más viejo se archiva | `/he-lima-backup/dashboard/` |

### 4.3 Right to erasure (Ley 29733 art. 22)

If an employee withdraws consent:
1. CEO + admin notified via email
2. `02_Personas_Public!C` (Apellidos_Nombres) replaced with hash
3. `02_Personas_Priv!B` (DNI) and `!I` (Tarifa) deleted from active Sheets
4. Audit log entry: `personal_data_erasure_request` with date + email
5. Backup retention: 5 years from erasure request for legal defense, then hard-delete

---

## 5. Cross-border data transfer

| Concern | Status |
|---|---|
| Data stored in Google Workspace | Located in region (Peru availability TBD with Google). If not PE, requires Ley 29733 art. 25 cross-border consent. |
| Apps Script execution region | US (Google default). No personal data leaves the Sheets cell — Apps Script reads/writes within doc. **Compliant.** |
| Backup encryption keys | Stored in vault (`he_lima_backup_key`) — vault is OVAV-managed. Key never leaves vault. |

**Action (P1):** verify Workspace data region with Google admin console. Document in install guide.

---

## 6. Data breach response

| Severity | Definition | Response time | Notify |
|---|---|---|---|
| CRITICAL | PII (DNI + Nombre) leaked externally | 24h | CEO + affected employees + INDECOPI (Peru DPA) |
| HIGH | Tarifa/Pago data leaked | 48h | CEO + affected employees |
| MEDIUM | Config drift, audit log tamper | 72h | CEO |
| LOW | Single cell edit by unauthorized user | 7 days | CEO in audit review |

**Breach playbook** (P2 to document fully):
1. CEO alerted via `06_AuditLogger.gs` anomaly rule
2. Apps Script pauses `07_TriggersWiring` (isolate)
3. Snapshot current state for forensics
4. INDECOPI notification per Ley 29733 art. 38

---

## Summary table

| Data class | Count fields | Highest sensitivity | Min strategy | Encryption | Retention |
|---|---|---|---|---|---|
| PII | 6 | DNI (HIGH) | Split sheets + role-based view | Drive default + backup AES-256 | 5y post-cese |
| Financial | 3 | Pago_HE_PEN (HIGH) | Restrict to admin/owner | Drive default + backup AES-256 | 7y (SUNAT) |
| Operational | ~18 | LOW | Visible to editors | Drive default | 5y |
| Audit trail | 8 cols + payload | MEDIUM (contains PII) | Hidden sheet + admin-only | Drive default + backup AES-256 | 7y |

→ See `recommendations.md` for prioritized hardening items.
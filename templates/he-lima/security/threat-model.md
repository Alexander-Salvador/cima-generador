# Threat Model — he-lima template (STRIDE)

> **Audit by:** Nora (API & Security Engineer)
> **Date:** 2026-09-24
> **Target:** he-lima v0.2.0-rev · 18 sheets · ~50 personas · Lima/PE
> **Method:** STRIDE classification per asset
> **Authority:** escalates to Diana (Security Auditor) for approval before P0 fixes ship

---

## Asset inventory

| Asset | Class | Exposure surface |
|---|---|---|
| `00_Config` (B2-B9) | Config / parameters | Editable by any Editor on spreadsheet |
| `02_Personas` (B,DNI / I,Tarifa) | PII + financial | Visible to all editors |
| `10_Registro_HE` | Working surface (Pago_HE_PEN) | Visible to all editors |
| `98_DB_YYYY_MM` (×12) | Storage (hidden) | Hidden but still API-readable |
| `99_Audit_Log` | Forensic trail | Hidden, append-only |
| `06_AuditLogger.gs` | Logging pipeline | Apps Script runtime |
| `08_DBSync.gs` | Grid ↔ DB sync | Apps Script runtime |
| Backup exports (Drive folder) | Data at rest | Drive sharing |

---

## S — Spoofing (suplantación)

| Threat | Attack vector | Impact | Mitigation |
|---|---|---|---|
| Editor suplanta identidad en audit log | onEdit reads `Session.getActiveUser()` which returns **EMPTY** under "view-only" link sharing or when run as simple trigger with installable perms | HIGH — log entries become useless for forensics | Use `Session.getEffectiveUser()` in **installable triggers** (not simple). Never write `getActiveUser()` into `Aud_Usuario`. Add `Aud_Runtime` column (simple/installable) to detect downgrade attacks. |
| Editor shares spreadsheet with attacker as Editor | OAuth scope too broad (`spreadsheets`) lets attacker edit all sheets incl. `00_Config` | HIGH — config takeover | Use **least-privilege scope**: `https://www.googleapis.com/auth/spreadsheets.currentonly` so addon can only edit user-selected range, not full sheet. |
| Manual edit of `Aud_Usuario` column R | User pastes email into R without onEdit firing (paste special, multi-cell paste) | MEDIUM — fake actor trail | Drive 99_Audit_Log from **server-side** onChange trigger (not editable by user). Column R becomes cosmetic only — never trust it as forensic source. |

**Detection:** compare `Aud_Usuario` (column R) vs `99_Audit_Log!B` (server log). Drift = spoof attempt → email alert to CEO.

---

## T — Tampering (manipulación no autorizada)

| Threat | Attack vector | Impact | Mitigation |
|---|---|---|---|
| Editor modifies `00_Config!B6/B7/B8` (factors) | Direct cell edit | CRITICAL — inflates/deflates payroll | Installable `onChange` trigger computes **SHA-256 hash** of B2:B9 + stores in `99_Audit_Log` row. If hash changes without prior `logAccion('config_change', …)` entry → revert + email CEO. |
| Editor modifies `02_Personas!I` (Tarifa_Hora_PEN) | Direct cell edit | CRITICAL — inflates individual pay | Same hash chain. Tarifa changes must be logged with prior + new value + editor email. Alert CEO if delta > 10% in single edit. |
| Editor overwrites formula in `10_Registro_HE!K/L/M/N/O/P` | Copy-paste values | HIGH — breaks derived calcs | Protect formulas via **Apps Script `protectRange()`** with `domainUsersOnly=false` + warning-only editors. Detect by comparing `cell.getFormula()` vs expected template on onEdit. |
| Editor adds row 111+ to `10_Registro_HE` | Copy paste into extended range | HIGH — bypasses DV/CF (data validation only applied to E2:V110) | Extend DV to `E2:V2000` and add Apps Script check: `if (row > 110 && row not in DB lookup) → reject + clear`. Better: protect row 111+ with `protectedRange`. |
| Manual edit of `98_DB_*` hidden sheets | Editor un-hides via menu | CRITICAL — storage integrity | Lock sheets via `Sheet.protect().setDescription('DBSync-only')`. Whitelist only `08_DBSync.gs` service account to edit. Editors see greyed sheet. |

**Detection:** `onChange` simple trigger fires for all edits → `06_AuditLogger.gs` must log every cell-level change for sheets `00_Config`, `02_Personas!I`, `99_Audit_Log`.

---

## R — Repudiation (negación de acción)

| Threat | Attack vector | Impact | Mitigation |
|---|---|---|---|
| User denies approving HE row after status flipped to `aprobado` | No cryptographic proof of approval | HIGH — payroll dispute | Approval workflow: editor sets `Estado='aprobado'` → `08_DBSync` triggers Apps Script → writes hash(Fecha+ID_Persona+Horas_Total+Pago_HE_PEN+timestamp) to `99_Audit_Log` column I (new). Approval is **non-repudiable** only with server-side timestamp from `new Date()` + email captured server-side. |
| User denies making config change | No tamper-proof trail | MEDIUM — config drift blamed on "someone" | Server-side onChange trigger with `LockService.getDocumentLock()` (30s wait) prevents TOCTOU. Log entry includes: editor email, before-hash, after-hash, range, old/new value. |
| User disputes "ghost" 99_Audit_Log entries (claim they didn't exist) | Log visible to editors (if unhidden) | LOW | Hide `99_Audit_Log` permanently + protect with editor-deny rule. Add `Row_Hash` column (SHA-256 of A:H per row) so tampering with log itself is detectable. |

**Mitigation:** all state-changing operations MUST go through Apps Script (installable triggers), never via direct cell edit. Direct edit = invalid state.

---

## I — Information Disclosure (exposición de datos sensibles)

| Data | Sheet/Col | Exposure today | Sensitivity | Mitigation |
|---|---|---|---|---|
| DNI (8 dígitos) | `02_Personas!B` | Visible to all Editors | **PII** (Peru: DNI = personal identifier, ley 29733) | Hide column B by default; expose only to `admin` role via custom menu "Ver DNI" → reverts after 5 min via `setTimeout`. Better: split into `02_Personas_Public` (no DNI) + `02_Personas_Priv` (DNI + Tarifa), protect Priv sheet. |
| Apellidos_Nombres | `02_Personas!C` | Visible to all Editors | **PII** | Acceptable for operational use; add `consent_flag` column to track ley 29733 consent. |
| Tarifa_Hora_PEN | `02_Personas!I` | Visible to all Editors | **Financial — confidential** | Move to `02_Personas_Priv!I`. Public sheet (`02_Personas_Public`) shows only Cargo/Cuadrilla/Turno for VLOOKUP into `10_Registro_HE`. Editors without `admin` role cannot see Tarifa. |
| Pago_HE_PEN | `10_Registro_HE!P` | Visible to all Editors | **Financial — confidential per employee** | Acceptable: each worker sees own rows (filter by ID_Persona). But aggregate `SUM(Pago_HE_PEN)` in `20_Dashboard` leaks total payroll → restrict `20_Dashboard` to `jefe_cuadrilla` + `admin` only. |
| Audit log content | `99_Audit_Log!G,H` (old/new value) | Contains DNI/tarifa diffs | **PII + financial** | If log is exposed (export to Drive), filter columns G/H for `02_Personas` rows — store hashed diff instead of plaintext. |

**Encryption at rest:** Google encrypts by default — but **backup exports** (CSV/xlsx) to Drive folder are NOT encrypted. P1 recommendation: encrypt with AES-256-GCM before upload.

---

## D — Denial of Service

| Threat | Attack vector | Impact | Mitigation |
|---|---|---|---|
| Bulk paste of 18,250 garbage rows | User with Editor access | onEdit trigger fires per row → Apps Script quota exhausted (90 min/day, 30 triggers/user) | Rate-limit `onEdit`: ignore events older than 5s in same range (dedup). Hard-cap rows in `10_Registro_HE` at 110 (reject paste beyond row 110). |
| Auto-creating 12 DB sheets via dropdown | `08_DBSync.createMonthSheetIfMissing` for 2024-2030 | Brief slowness | Already lazy — OK. Cap at 36 sheets max (3 years coverage), refuse creation beyond. |
| INDIRECT saturation in `20_Dashboard` | User spams dropdown | Each dropdown change triggers DB load + KPI recalc (potentially 34k rows) | Throttle `refrescarKPIs()` with `PropertiesService` flag — 1 refresh per 30s max. |
| Apps Script 6min timeout | saveGridToDB() processing 110 rows | Trigger dies mid-sync → grid-DB divergence | `LockService.getDocumentLock(10000)` + chunked writes (save 25 rows per batch). On timeout, `Sync_Status='conflict'` + alert. |
| Trigger quota exhaustion | Multiple addons installing triggers | 30 triggers/user max → legitimate triggers blocked | Audit triggers monthly (`07_TriggersWiring.owl()`) — keep ≤ 5 active. Recommend rotation every 90 days. |
| VLOOKUP chain depth | Repeated nested lookups across 18 sheets | formula chain depth limit (50) exceeded | Already mitigated: only G uses VLOOKUP, rest derived locally. |
| Storage exhaustion | 18,250 rows × 12 months × 22 cols | 401k cells → 5% of 10M limit (safe). But hidden sheets cost Drive quota | OK at current scale; warn at 70% of 50k row/hoja. |

---

## E — Elevation of Privilege

| Threat | Attack vector | Impact | Mitigation |
|---|---|---|---|
| Editor changes `Estado` from `borrador` → `pagado` directly | Direct cell edit on T column | CRITICAL — self-approve payroll | **State machine enforced by Apps Script**: `validarTransicionEstado(oldEstado, newEstado, userRole) → rejects 'borrador'→'pagado'`. Only roles `jefe_cuadrilla`+ can set `aprobado`; only `admin` can set `pagado`. Violations revert + log + alert. |
| Editor changes own `Cargo` from `tecnico` → `admin` | Direct edit on `02_Personas!D` | CRITICAL — privilege escalation | Protect column D with `protectedRange()` allowing only `admin` role to edit. Detect changes via onChange + alert CEO. |
| Editor unprotects sheet via "File → Version history" | Google Drive native feature | MEDIUM | Document this risk in `recommendations.md`; mitigation: use **named ranges** + **Apps Script-only writes** for sensitive state. |
| Editor installs malicious addon with broad scopes | OAuth consent | CRITICAL | Template ships with `appsscript.json` declaring **minimal scopes**. Apps Script project must be locked to template-only manifest (`@onlyCurrentDoc`). Verify in CI. |
| Editor copies template, modifies Apps Script, redistributes | IP leakage / tampering | LOW (operationally) — HIGH (reputational) | Apps Script project is bound to spreadsheet ID. Copying duplicates script. Add license header in every `.gs` file. |

**Mitigation pattern:** all role transitions go through `08_DBSync.validarTransicion(usuarioEmail, accion, params)`. Simple onEdit triggers reject invalid state changes; installable onChange triggers log accepted ones.

---

## Summary of STRIDE risks

| Category | Count CRITICAL | HIGH | MEDIUM | LOW |
|---|---|---|---|---|
| Spoofing | 0 | 2 | 1 | 0 |
| Tampering | 2 | 2 | 0 | 0 |
| Repudiation | 0 | 1 | 1 | 1 |
| Information Disclosure | 0 | 2 | 1 | 1 |
| Denial of Service | 0 | 1 | 3 | 1 |
| Elevation of Privilege | 2 | 1 | 1 | 1 |
| **TOTAL** | **4** | **9** | **7** | **4** |

→ See `recommendations.md` for prioritized mitigations.
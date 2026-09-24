# Security Recommendations — he-lima template

> **Audit by:** Nora (API & Security Engineer)
> **Date:** 2026-09-24
> **Priority scale:** P0 (block production) · P1 (ship within 1 sprint) · P2 (next quarter)
> **Owner of fixes:** andres (Implementer Senior) · Diana (Security Auditor approval for P0)
> **Estimated total effort:** ~28 hours work + 4 hours Diana review

---

## P0 — Block production (MUST fix before go-live)

| ID | Title | OWASP ref | Owner | Effort |
|---|---|---|---|---|
| P0-01 | **Proteger `00_Config` con `protectedRange()` + solo CEO + 1 delegado** | A01 | andres | 2h |
| P0-02 | **Bloquear hojas `98_DB_*` y `99_Audit_Log` con Apps Script — solo `08_DBSync.gs` service account puede escribir** | A01, A08 | andres | 3h |
| P0-03 | **Split `02_Personas` en `Public` (sin DNI, sin Tarifa) + `Priv` (DNI + Tarifa, admin-only)** | A01 | andres | 4h |
| P0-04 | **`Session.getEffectiveUser()` en installable triggers (no simple); columna R informativa, log real en `99_Audit_Log!B`** | A07 | andres | 2h |
| P0-05 | **State machine `Estado`: `validarTransicionEstado()` rechaza `borrador→pagado`, role-based** | A04 | andres | 3h |
| P0-06 | **`LockService.getDocumentLock(30s)` en `08_DBSync` save/load — race condition fix** | A04 | andres | 2h |
| P0-07 | **SHA-256 hash chain en `00_Config!B2:B9` + `02_Personas!I` — detecta cambios manuales no autorizados** | A08 | andres | 4h |
| P0-08 | **OAuth scopes mínimos en `appsscript.json`: `spreadsheets.currentonly` + `script.container.ui` + `drive.file` (NO gmail, NO docs, NO spreadsheets full)** | A05 | andres | 1h |
| P0-09 | **Alertas tiempo-real CEO: tarifa delta > 10%, Pago_HE_PEN > umbral, estado rechazado, formula overwrite** | A09 | andres | 4h |
| P0-10 | **try/catch universal en handlers + write a `99_Audit_Log!I` (Error_Trace nueva col)** | A04, A09 | andres | 1h |
| P0-11 | **DV regex `^[A-Za-z0-9 _\-./]+$` en columna Q (Observaciones) — bloquea `=HYPERLINK`, `=IMAGE`, `=IMPORT*`** | A03, A10 | andres | 1h |

**P0 subtotal: 11 items · ~27h**

### P0 implementation order (dependency-respecting)

```
P0-08 (scopes) → P0-04 (getEffectiveUser) → P0-07 (hash chain)
                                          ↓
                P0-01, P0-02, P0-03 (sheet protection — depends on P0-08)
                                          ↓
                P0-05 (state machine) → P0-06 (locks) → P0-09 (alerts)
                                          ↓
                P0-10 (error trace) → P0-11 (formula injection DV)
```

---

## P1 — Ship within 1 sprint (post-launch acceptable but scheduled)

| ID | Title | OWASP ref | Owner | Effort |
|---|---|---|---|---|
| P1-01 | **Encriptar backups mensuales con AES-256-GCM; clave en vault `he_lima_backup_key`** | A02 | andres | 4h |
| P1-02 | **Hash DNI en logs (últimos 4 dígitos + hash completo) — no plaintext en `99_Audit_Log!G,H`** | A02 | andres | 2h |
| P1-03 | **Pre-check en `08_DBSync.preWrite()` re-valida FK cross-sheet al escribir (TOCTOU fix)** | A04 | andres | 2h |
| P1-04 | **Idempotency: `Mes_Año + Fecha + ID_Persona` como natural key; `ON CONFLICT SKIP` semantics** | A04 | andres | 3h |
| P1-05 | **Service account para `07_TriggersWiring` y `08_DBSync` — desacopla de owner humano** | A07 | andres | 6h |
| P1-06 | **Versión de template en `00_Config!B11` + onOpen warning si drift** | A06 | andres | 1h |
| P1-07 | **`protectRange()` en columnas K,L,M,N,O,P,U,V (formulas) — `domainUsersOnly=false` warning-only** | A08 | andres | 2h |
| P1-08 | **Archive `99_Audit_Log` mensual encriptado a `/he-lima-backup/audit/`** | A09 | andres | 2h |
| P1-09 | **Custom DV error messages en español es-PE (no "DV failed")** | A09 | andres | 1h |
| P1-10 | **CI check: `appsscript.json` no contiene `gmail`, `forms`, `docs`, `spreadsheets` (full) scopes** | A05 | andres | 1h |
| P1-11 | **CI check: `grep -r 'apiKey\|secret\|token' go-runtime/internal/templates/he-lima/` retorna vacío** | A05 | andres | 1h |
| P1-12 | **Restore procedure documentado + tests E2E (Diana valida)** | A02 | andres + diana | 3h |
| P1-13 | **Verificar Workspace data region con Google admin (PE availability)** | Data Protection | andres | 2h |
| P1-14 | **Drive folder ACL: backup folder = CEO owner, Editors read-only, no public links** | A01 | andres | 1h |

**P1 subtotal: 14 items · ~31h**

---

## P2 — Next quarter (defensible, not blocking)

| ID | Title | OWASP ref | Owner | Effort |
|---|---|---|---|---|
| P2-01 | **Rotar triggers cada 90 días — script `07_TriggersWiring.rotarTriggers()`** | A05, A09 | andres | 3h |
| P2-02 | **Anomaly detection: Pago_HE_PEN > 3× median → flag en `99_Audit_Log`** | A09 | andres | 4h |
| P2-03 | **`clasp` CI/CD pipeline para `.gs` deployment + versionado** | A08 | andres | 6h |
| P2-04 | **Cruza_medianoche flag en Q (Observaciones) para turnos nocturnos largos** | Input validation | andres | 2h |
| P2-05 | **Validación módulo 11 del DNI peruano** | Input validation | andres | 4h |
| P2-06 | **SBOM generator para Apps Script manifest (output vacío por ahora, documenta)** | A08 | andres | 2h |
| P2-07 | **2FA policy documentada en install guide (requerida para Workspace domain)** | A07 | andres | 1h |
| P2-08 | **Breach playbook completo (procedimiento formal Ley 29733 art. 38)** | Data Protection | andres + diana | 4h |

**P2 subtotal: 8 items · ~26h**

---

## Quick wins (< 1 hora cada uno)

### QW-1 · Hardening `appsscript.json` (30 min)

```json
{
  "timeZone": "America/Lima",
  "dependencies": {},
  "exceptionLogging": "STACKDRIVER",
  "runtimeVersion": "V8",
  "oauthScopes": [
    "https://www.googleapis.com/auth/spreadsheets.currentonly",
    "https://www.googleapis.com/auth/script.container.ui",
    "https://www.googleapis.com/auth/drive.file"
  ],
  "executionApi": { "access": "NONE" }
}
```

Effort: copy-paste + verify. Closes A05 + A07 partially.

### QW-2 · Proteger `00_Config` con `protectedRange()` (45 min)

```javascript
// In 07_TriggersWiring.gs
function instalarProtecciones() {
  const ss = SpreadsheetApp.getActive();
  const cfg = ss.getSheetByName('00_Config');
  const protection = cfg.protect().setDescription('Solo CEO + delegado');
  protection.setDomainUsersOnly(false);  // restrict to specific emails
  const ceoEmail = PropertiesService.getScriptProperties().getProperty('CEO_EMAIL');
  protection.addEditors([ceoEmail]);  // explicit allow-list
  // Note: Drive sharing still works for view-only; this prevents cell edits
  Logger.log(`00_Config protected for ${ceoEmail}`);
}
```

Effort: 45 min including test. Closes P0-01.

### QW-3 · DV regex en columna Q (Observaciones) — anti HYPERLINK/IMAGE injection (20 min)

```
Columna: 10_Registro_HE!Q2:Q110
Tipo: text → regex
Patrón: ^[A-Za-z0-9 _\-./]+$
Acción ante violación: rechazar + mostrar mensaje "Observaciones solo permite letras, números y puntuación básica (. - /)"
```

Effort: 20 min via Sheets UI (Data → Data validation). Closes P0-11 (A03 + A10).

---

## Total scorecard

| Severity | Count | Est. hours | Risk reduction |
|---|---|---|---|
| P0 | 11 | ~27h | 70% of critical risk closed |
| P1 | 14 | ~31h | 25% additional |
| P2 | 8 | ~26h | 5% (long-term hygiene) |
| **TOTAL** | **33** | **~84h** | — |

### Validation sequence

1. P0 items ship in single sprint (1 week, 1 engineer)
2. Diana (Security Auditor) reviews P0 fixes → approve → go-live
3. P1 items ship in next sprint (1 week)
4. P2 items scheduled into quarterly OKR cycle

### Acceptance criteria for go-live

- [ ] All P0 items implemented + tested E2E
- [ ] `ovav validate` passes (template-level validators happy)
- [ ] Diana sign-off on STRIDE mitigations
- [ ] Penetration test: 3 personas (tecnico, supervisor, admin) attempt forbidden actions → all rejected
- [ ] Backup restore drill executed successfully on staging copy
- [ ] Install guide documents all sharing rules + 2FA requirement

---

*Authority: escalates to Diana (Security Auditor) for final approval. Nora signs off on design; Diana owns production-readiness call.*
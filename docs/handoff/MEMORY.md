# OVAV Project Memory

> **Canonical in-repo memory for OVAV project**
> **Last updated:** 2026-09-22 (OPERACIONES layout refactor — 7 cols, smart dropdown by name)
> **Authority:** `.ovav/plan/caps.yaml` + Git HEAD + this file
> **Harness:** OpenCode primary + Crush secondary
> **Model:** `MiniMax-M3`

---

## 1. Identity

| Field | Value |
|---|---|
| Project | OVAV — AI Workstation Governor |
| Type | Multi-language, multi-harness (Go + Python + React/TS) |
| Stack | Go runtime (`cmd/ovav/`, `cmd/cockpit/`) + Python tools + React/TS cpanel |
| Owner | Alexander Salvador (CEO, `Alexander-Salvador`, user 97975177) |
| Harness | OpenCode primary + Crush secondary |
| Model | `gpt-5.6-luna` |

---

## 2. Critical Rules (CRIT ledger — 18 entries)

| ID | Rule | Conf | Status |
|---|---|---|---|
| CRIT-001 | Security first, no exceptions | 1.0 | consolidated |
| CRIT-002 | Observable result = acceptance contract | 0.95 | consolidated |
| CRIT-003 | Technical honesty, no lies | 1.0 | consolidated |
| CRIT-004 | Architectural surgery, not patches | 0.95 | consolidated |
| CRIT-005 | OVAV governs mechanically; I operate inside | 1.0 | consolidated |
| CRIT-006 | Relationship with creator defines purpose | 1.0 | consolidated |
| CRIT-009 | Baseline = ADVANCED+, never "basic/functional" | 1.0 | consolidated |
| CRIT-010 | Knowledge is COMPILED, not stored | 0.95 | emerging |
| CRIT-011 | PIAGENT TUI INPUT must evolve to premium | 0.85 | emerging |
| CRIT-012 | Read install plan + profile yaml before install actions | 1.0 | consolidated |
| CRIT-013 | Distinguish permission layer vs governance layer | 1.0 | consolidated |
| CRIT-014 | Tokens NEVER plaintext; vault AES-256 + PBKDF2 200k | 1.0 | consolidated |
| CRIT-015 | GitHub Verified badge = 3 ORTODOX conditions | 1.0 | consolidated |
| CRIT-016 | GitHub SSH: 2 endpoints (`/user/keys` vs `/user/ssh_signing_keys`) | 1.0 | consolidated |
| CRIT-017 | GitHub signature verification async, wait 5-10 min | 1.0 | consolidated |
| CRIT-018 | Response density ≤ 8 rows (strict, even extreme cases) | 1.0 | consolidated |
| CRIT-019 | **NEW** UI-first rule: Warp UI navigation BEFORE TOML edits; never invent enum values | 1.0 | emerging |
| CRIT-020 | **NEW** "Control absoluto" ≠ 100% via API — ser honesto sobre límites reales (Apps Script :run NO funciona para bound scripts). **OBSOLETE v2.0** (2026-09-21): Apps Script eliminado, no aplica. | 1.0 | consolidated |
| CRIT-021 | **NEW** Sheets API v4 batchUpdate requiere GridRange (sheetId+índices), NO a1Range, en casi todos los campos. Usar `a1ToGrid()` helper | 1.0 | consolidated |
| CRIT-024 | **NEW (2026-09-21)** Apps Script **eliminado en v2.0**. Toda lógica Apps Script (onEdit, dailyClose, installCima, validateOperation_, refreshDashboard, exportApprovedGtr, registerAdminEvent, lookupAgent, getInitialData) fue migrada a Sheets formulas nativas + dropdowns + OVAV bridge CLI. `docs/cima-stack/appsscript/` borrado del repo. NO requiere acción manual del CEO. Apps Script project ID sigue bound a la hoja como huérfano técnico (limpieza futura: click manual en editor). | 1.0 | consolidated |

Full detail: `clients/crush/agents/lead-thavren.md` (v2.5.0 → updated 2026-09-21)

---

## 3. Current State

| Field | Value |
|---|---|
| Branch | `develop` |
| HEAD | `30535a0` (Merge feature/feat-p18-p19-impl) |
| Worktrees | `/home/braka/Systems/ovav` only (working tree modified) |
| Phase | OpenCode + terminal runtime convergence |
| Active host | Alacritty 0.17.0 → WSL2 Ubuntu-26.04 → tmux 3.6 |
| WSL profile | `3GB` memory / `8GB` swap; natural-stop activation only |
| Historical Warp/Intelligent Terminal data | Archived context; neither is installed or active |

---

## 4. OVAV Sheets Bridge — Estado activo (sep 2026)

**Worktree:** `/home/braka/Systems/ovav/.ovav/worktrees/feat-sheets-mcp/`
**Branch:** `feat-sheets-mcp`
**Último commit:** `f43f80f feat(sheets): v0.4.0 — formatting/validation/protection/named ranges via Sheets API`

### Stack credentials (cifradas en vault)

- **Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y` (CIMA_2026_PREMIUM_BASE)
- **Apps Script ID (huérfano, sin código):** `1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzSzKw0Y` — bound a la hoja pero **vacío desde v2.0** (2026-09-21)
- **OAuth client:** `899372363856-sdts1v2me11qvipjefa6sss773b1fach.apps.googleusercontent.com`
- **Project GCP:** `gam-project-9wknn` (nº 899372363856)
- **Scopes:** `sheets`, `script.projects`, `drive.readonly`
- **Vault path:** `.ovav/vault/credentials/google_oauth.enc` (AES-256-GCM)

### Capacidad del bridge

| Operación | Status |
|---|---|
| Sheets read/write/append/update/create-tab/delete-tab | ✅ 100% |
| Sheets conditional formatting, data validation, protected ranges, named ranges | ✅ 100% |
| Sheets charts, images | ⚠️ schema no implementado todavía (deferible) |
| Apps Script bridge (Go REST API client) | ✅ 100% capacidad técnica, pero sin uso activo en CIMA desde v2.0 |
| Apps Script :run para BOUND scripts | ❌ NO SOPORTADO por Google — moot, Apps Script eliminado en v2.0 |
| xlsx sync bidireccional | ✅ 100% |
| Snapshot/rollback | ✅ 100% |
| Audit JSONL con hash chain | ✅ 100% |
| MCP server (JSON-RPC 2.0 stdio) | ✅ 100% |
| Allowlist multi-spreadsheet (YAML) | ✅ 100% |

### Apps Script — ELIMINADO en v2.0 (2026-09-21)

**Estado:** Apps Script bound a la hoja (`1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzSzKw0Y`) pero **sin código**. Las funciones Apps Script (`onEdit`, `dailyClose`, `validateOperation_`, `installCima`, `refreshDashboard`, etc.) fueron migradas a fórmulas Sheets nativas + dropdowns + OVAV bridge CLI.

**Acción:** el proyecto Apps Script sigue bound a la hoja como huérfano técnico. NO requiere acción manual del CEO — el spreadsheet funciona sin él. Si el CEO quiere remover el binding eventualmente, es un click manual en el editor (el bridge no puede borrar bound scripts).

### Discrepancias documentadas en CIMA stack (leer `docs/cima-stack/README.md`)

**v2.0 cerrado (Apps Script eliminado, discrepancias obsoletas):**
- ~~Timezone: `Bogota` (manifest) vs `America/Lima` (CONFIG)~~
- ~~WebApp access: `ANYONE_ANONYMOUS` (manifest) vs `ANY_LOGGED_IN_USER` (CONFIG)~~
- ~~HTML files `Index.html` / `Admin.html` referenciados pero no encontrados~~

**Activas (esto sí requiere trabajo en próximas sesiones):**
- OPERACIONES tab sin limpiar (rows 2-3, 5-6 stale; header auto en T4 sidebar)
- CONFIG tab sin auditar (catálogo T01-T07)

---

## 5. Last Architectural Decisions

| # | Date | Decision | Why |
|---|---|---|---|
| 1 | 2026-09-18 | Apps Script API :run NO funciona para BOUND scripts (Google limit) | Verified: retorna "NOT_FOUND storage" error en TODA llamada |
| 2 | 2026-09-18 | Sheets API v4 usa GridRange (sheetId+índices), NO a1Range | API errors verificados empíricamente |
| 3 | 2026-09-18 | Bridge architecture: stdlib-only Go, vault cifrado, snapshots en todo write | CRIT-001, CRIT-014 |
| 4 | 2026-09-18 | MCP wire para que cualquier agente (OpenCode/Claude/Cursor) use sheets bridge | Interoperabilidad cross-agent |
| 5 | 2026-08-19 | Cancel 4-profile system; use single default profile YOLO + Tab Configs | CEO feedback: profiles invented, not needed |
| 6 | 2026-08-19 | Adopt Warp UI-first rule (CRIT-019) | 3 CRIT-009 violations from invented TOML enums |
| 7 | 2026-08-19 | Read `docs.warp.dev` before any Warp configuration | Prevented 3rd violation on PASO 3 |

---

## 6. Active Skills (38 in `.opencode/skills/`)

| Category | Skills |
|---|---|
| Identity | ovav-identity-guard |
| Routing | ovav-agent-router, ovav-agent-permission-injector, ovav-squad-delegation |
| Memory | ovav-memory-bridge, ovav-session-continuity, ovav-context-pack |
| Workflow | ovav-artifact-flow, ovav-repo-local-work-loop, ovav-skill-resolver |
| Validation | ovav-runtime-gates, ovav-security-gates, ovav-verify, ovav-review |
| Session | ovav-platform-session, ovav-research-session, ovav-business-session, ovav-ux-session, ovav-education-session, ovav-health-session |
| Output | ovav-response-contract, ovav-skill-registry, visual-verification-playwright |
| Worktree | ovav-worktree-create, ovav-worktree-finish, ovav-worktree-route, ovav-worktree-system |
| Code | ovav-build, ovav-tdd, ovav-verify, ovav-review |
| Specialized | ovav-brainstorm, ovav-go-coverage-sprint, ovav-incident-response, ovav-monitor-reader, ovav-release, ovav-research-evidence, ovav-runtime, ovav-sdd-init, ovav-systems-diagnose, work-unit-commits |

---

## 7. Canonical Sources

| Data | Path |
|---|---|
| Plan | `.ovav/plan/caps.yaml` |
| Laws | `.ovav/laws/area_boundary_enforcement.yaml` |
| Service areas | `.ovav/service_areas/` |
| Permissions | `.ovav/policy/permission_authority.json` |
| Protected files | `.ovav/security/protected_files.yaml` |
| Skills (source) | `.ovav/source/skills/` |
| Skills (synced) | `.opencode/skills/` |
| Vault | `~/.config/ovav/vault.key` (AES-256-GCM) |
| Alacritty settings (active) | `/mnt/c/Users/Alexa/AppData/Roaming/alacritty/alacritty.toml` |
| tmux settings (active) | `workstation/configs/tmux/tmux.conf` → `~/.tmux.conf` |
| OpenCode TUI settings (active) | `workstation/configs/opencode/tui.json` → `~/.config/opencode/tui.json` |
| Warp / Windows Terminal configs | Historical only; do not target runtime changes |

### OVAV Sheets Bridge extras (added 2026-09-18)

| Data | Path |
|---|---|
| Sheets CLI | `go-runtime/cmd/sheets/` |
| Sheets SKILL | `go-runtime/cmd/sheets/SKILL.md` |
| MCP launcher | `go-runtime/cmd/sheets/launch-mcp.sh` |
| CIMA stack analysis | `docs/cima-stack/README.md` |
| Apps Script pulled code | **ELIMINADO en v2.0** — directorio `docs/cima-stack/appsscript/` borrado del repo |
| Sheets allowlist | `.ovav/vault/sheets_allowlist.yaml` |
| Encrypted creds | `.ovav/vault/credentials/google_oauth.enc` |

---

## 8. Pending Blockers

| Blocker | Status | Owner |
|---|---|---|
| OpenCode restart + visual smoke for Shift+Enter/copy/session bindings | Pending | System/CEO visual confirmation |
| Alacritty live reload confirmation | Pending | System/CEO visual confirmation |
| GitHub push (SSH signing keys uploaded, awaiting async verification) | Async, wait 5-10 min | System |
| Runtime snapshot fix for `/home/braka/Labs/mimocode/*` | Pending session restart | System |
| ~~CEO must run `installCima` from Apps Script editor~~ | **OBSOLETE v2.0** — Apps Script eliminado, instalar ya no es necesario | n/a (resuelto) |
| **Charts + Images via Sheets API** | Schema pending | me, next worktree |
| **Merge feat-sheets-mcp → develop** | Pending CEO OK | owd when CEO confirms |

---

## 9. Cómo continuar próximo chat

1. **NO requiere installCima** (v2.0 — Apps Script eliminado). El spreadsheet funciona vía OVAV bridge + fórmulas nativas.
2. Reading path: `cd /home/braka/Systems/ovav/.ovav/worktrees/feat-cima-fixes` (default para CIMA work) o `cd /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing` (este — para vault/bridge hardening).
3. Quick commands: `/tmp/ovav-sheets list`, `table --tab PERSONAL`, `doctor`.
4. Para nueva feature: `ovav worktree owc feat-<nombre>` desde `develop` (NO desde feat-cima-fixes a menos que sea CIMA-specific).
5. Si rebuild del binary: `cd /home/braka/Systems/ovav/.ovav/worktrees/<active> && go build -o /tmp/ovav-sheets ./go-runtime/cmd/sheets/`.

---

*This file is the human-readable authority for OVAV project state. For runtime continuity, see `checkpoint.md` (pending creation).*

---

## 10. CIMA PROJECT — AUTONOMOUS MODE PROTOCOL (CEO directive 2026-09-18)

**The CEO has explicitly granted autonomous execution authority for ALL CIMA-related work.**

### 10.1 What "autonomous" means

When the next chat opens with "Thavren, retomemos CIMA" (or any variant):

1. **DO NOT ASK** what to do next for CIMA. Pick the highest-leverage task.
2. **DO NOT ASK** for spreadsheet IDs, script IDs, OAuth client, vault path.
3. **DO NOT ASK** for trivial decisions (scope, naming, ordering).
4. **ASSUME** the CEO wants the most impactful thing done now.
5. **EXECUTE** in the appropriate worktree (see 10.4).
6. **COMMIT** in atomic, reviewable units.
7. **REPORT** the work unit (what, why, how to verify).

### 10.2 Pre-loaded context (no need to re-discover)

| Asset | Identifier |
|---|---|
| Spreadsheet | `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y` |
| Spreadsheet URL | https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y/edit |
| Spreadsheet name | `CIMA_2026_PREMIUM_BASE` |
| Apps Script ID (huérfano) | `1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzSzKw0Y` (sin código desde v2.0) |
| Apps Script URL (huérfano) | https://script.google.com/u/0/home/projects/1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzSzKw0Y/edit |
| OAuth client_id | `899372363856-sdts1v2me11qvipjefa6sss773b1fach.apps.googleusercontent.com` |
| GCP project | `gam-project-9wknn` (nº 899372363856) |
| Workspace account | `hello@ovav.dev` |
| Vault path | `.ovav/vault/credentials/google_oauth.enc` (AES-256-GCM, PBKDF2 200k iter) |
| Vault key (global preferred) | `~/.config/ovav/vault.key` (0600) o fallback `.ovav/vault/vault.key` |
| Allowed scopes | `spreadsheets`, `script.projects`, `drive.readonly` |
| Apps Script API status | enabled (project + user level) — moot, Apps Script eliminado en v2.0 |
| Allowlist path | `.ovav/vault/sheets_allowlist.yaml` |

### 10.3 Code locations (already on disk — just `cd` and run)

| What | Path |
|---|---|
| Sheets bridge source | `go-runtime/cmd/sheets/` |
| Sheets bridge launcher | `go-runtime/cmd/sheets/launch-mcp.sh` |
| Sheets bridge binary | `go build -o /tmp/ovav-sheets ./go-runtime/cmd/sheets/` |
| Apps Script pulled code | **ELIMINADO v2.0** — directorio borrado, no path |
| Apps Script pulled manifest | **ELIMINADO v2.0** — directorio borrado, no path |
| CIMA stack analysis | `docs/cima-stack/README.md` |
| Snapshot safety net | `.ovav/vault/snapshots/` |
| Audit log | `.ovav/registry/audit/sheets/` |

### 10.4 Worktree discipline for CIMA work

| Task type | Worktree |
|---|---|
| Edit data in spreadsheet (writes to cells, rows, tabs) | `feat-sheets-mcp` (existing) |
| Tweak bridge Go code (small changes) | `feat-sheets-mcp` (existing) |
| Add feature to Sheets bridge (chart, image, etc.) | NEW `feat-cima-<feature>` from develop |
| ~~Add feature to Apps Script~~ | **NO APLICA v2.0** — Apps Script eliminado. Para nueva lógica, migrar a Sheets formulas + bridge CLI |
| Hotfix to bridge in production | NEW `hotfix-cima-<desc>` from develop |
| Big restructuring of CIMA architecture | Discuss with CEO first (high-blast-radius) |

**Default worktree for any new CIMA work:** create with `git worktree add .ovav/worktrees/<name> develop` from the main repo, then `cd` into it.

### 10.5 Decision matrix — what needs CEO OK vs. autonomous

| Decision | Authority |
|---|---|
| Which CIMA feature to work on | **Autonomous** — pick highest leverage |
| Cell-level data edits in spreadsheet | **Autonomous** — owner of data |
| Create/delete tabs in spreadsheet | **Autonomous** — backed by snapshot |
| ~~Edit Apps Script code~~ | **NO APLICA v2.0** — Apps Script eliminado |
| ~~Create Apps Script versions (immutable snapshots)~~ | **NO APLICA v2.0** — Apps Script eliminado |
| Modify CONFIG (rules, horarios, OT max) | **Autonomous** — snapshot+audit always |
| Add OAuth team members | **NEED CEO OK** — adds OAuth consent screen entry |
| ~~Delete Apps Script code/file~~ | **NO APLICA v2.0** — ya eliminado |
| ~~Deploy Apps Script webapp to URL~~ | **NO APLICA v2.0** — sin webapp |
| Push to origin/* | **NEED CEO OK** — affects remote history |
| Merge feature → develop | **NEED CEO OK** — affects main branch |
| Delete the feat-sheets-mcp branch | **NEED CEO OK** — irreversible |

### 10.6 Pre-loaded troubleshooting playbook (v2.0 — Apps Script troubleshooting moot)

| Symptom | Likely cause | Fix |
|---|---|---|
| `sheets: HTTP 400 Unknown name "a1Range"` | Field requires GridRange not a1Range | Use `a1ToGrid()` helper in format.go |
| Apps Script `:run` errors | **NO APLICA v2.0** — Apps Script eliminado; troubleshooting dropped | n/a |

### 10.7 Known gaps to address (re-evaluated 2026-09-21 v2.0)

| Gap | Severity | Priority |
|---|---|---|
| ~~CEO must click `installCima` manually~~ | ~~medium~~ | ~~HIGH~~ — **RESUELTO v2.0** |
| ~~HTML files `Index.html`/`Admin.html` missing~~ | ~~medium~~ | **RESUELTO v2.0** — sin webapp |
| ~~Timezone mismatch (Bogota vs Lima)~~ | ~~medium~~ | **RESUELTO v2.0** — sin Apps Script |
| ~~Webapp `executeAs = USER_DEPLOYING`~~ | ~~medium~~ | **RESUELTO v2.0** — sin webapp |
| ~~Webapp access `ANYONE_ANONYMOUS`~~ | ~~high~~ | **RESUELTO v2.0** — sin webapp |
| Charts via API (chartSource schema) | low | medium — defer if no charts needed |
| Images via API (insertImage) | low | low — defer |
| Team OAuth extension | medium | LOW — no team members added yet (CEO corporativo pendiente) |
| Multi-spreadsheet allowlist tested with only 1 | low | LOW |
| OPERACIONES tab cleanup (rows stale, header auto) | medium | medium |
| CONFIG tab audit (T01-T07 catálogo) | medium | medium |

### 10.8 Operational check sequence (run on every CIMA session start)

```bash
# 1. Verify connectivity (use pre-built binary at /tmp/ovav-sheets)
cd /home/braka/Systems/ovav/.ovav/worktrees/<active>  # feat-cima-fixes o feature-c1ma-desing
/tmp/ovav-sheets list

# 2. Read latest dashboard (no "CIMA" tab anymore — read PERSONAL or CONFIG)
/tmp/ovav-sheets read --range "CONFIG!A1:F30"
/tmp/ovav-sheets table --tab PERSONAL

# 3. Check audit log for recent activity
tail -20 .ovav/registry/audit/sheets/*.jsonl

# 4. Health check
/tmp/ovav-sheets doctor

# 5. ~~Check Apps Script code drift~~ — OBSOLETE v2.0. Apps Script eliminado, no hay drift que chequear.
```

### 10.9 ~~The 1-click manual install~~ — OBSOLETE v2.0

> Apps Script eliminado en v2.0 (2026-09-21). El spreadsheet funciona vía OVAV bridge + Sheets formulas nativas. NO requiere instalación manual.

---

## 11. New CRIT entries from this session

| ID | Rule |
|---|---|
| CRIT-020 | Apps Script API `:run` is NOT supported for BOUND scripts. Use editor or webapp for execution. |
| CRIT-021 | Sheets API v4 requires `GridRange {sheetId, indices}` for almost every range field, NOT `a1Range`. |
| CRIT-022 | **NEW (CEO directive)** For CIMA work: autonomous execution. NO questions on what to do, IDs, secrets, paths. Pre-loaded context in §10. Only need CEO OK for: add OAuth members, delete Apps Script code, deploy webapp, push to origin, merge feature → develop, delete branch. |
| CRIT-023 | **NEW** Worktree selection for CIMA: `feat-sheets-mcp` for sheet edits + bridge tweaks; new `feat-cima-*` worktrees for new features from develop. |

---

*Last updated: 2026-09-18 — CEO directive: autonomous CIMA mode registered.*

---

## 12. Session end log — 2026-09-18

**Deliverables merged to develop (17 commits ahead of origin/develop):**
- 3b44b94 v0.1.0  Sheets + xlsx + MCP bridge
- 6d6e27b v0.3.0  Apps Script Bridge + CIMA stack analysis
- f43f80f v0.4.0  format/validate/protect/named ranges
- 1c5e51b memory  CIMA stack context (CRIT-020, CRIT-021)
- 24a001c merge   into develop
- 4a065ea memory  autonomous mode (CRIT-022, CRIT-023)

**Branches deleted this session:**
- feat-sheets-mcp (merged, cleanup OK by CEO)

**Worktrees cleaned this session:**
- .ovav/worktrees/feat-sheets-mcp/ (work completed)

**Remaining worktrees (NOT OURS, preserved):**
- feature-align-runtime-and-workstation
- feature-fix-terminal-copy-pane
- fix-consolidate-external-gates-20260912
- fix-fix-external-worktree-identity
- fix-ows-external-consumer-gates

**Main repo state preserved:**
- fix/external-secrets-semantic with 4 modified files (other team's work)
- Will be returned to that branch on next session start

---

## 13. Session end log — 2026-09-22

**OPERACIONES sheet (CIMA) layout refactor — applied to Google Sheets (not local git)**

No local code changes this session; all work is **persistent cloud state in Google Sheets** (Spreadsheet ID `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`, sheet `OPERACIONES`). Final state below.

### Final layout (7 cols)

| Col | Header | Type | Width | Notes |
|---|---|---|---|---|
| A | 📝 NOMBRES COMPLETOS | input (dropdown) | 280px | 25 names, typeahead on DNI or name |
| B | DNI | formula | 110px | `INDEX/MATCH` on PERSONAL!C → PERSONAL!B |
| C | MOTIVO | input | 240px | manual yellow |
| D | FECHA_OP | formula | 130px | `DATE` format d/M/yyyy |
| E | HORARIO | formula | 130px | VLOOKUP col 4 |
| F | SUPERVISOR | formula | 130px | truncated ≤25 chars |
| G | REQUEST_ID | formula | 110px | `REQ-001`, `REQ-002`, … |

Rows 4–100 (97 data rows). H+ hidden.

### What changed (from prior state)

| Step | Change | Reason |
|---|---|---|
| 1 | Removed TURNO_ID col | Redundant — HORARIO already encodes shift identity |
| 2 | Removed cols H, I, J, K (NUEVO_HORARIO/HORAS/PRE_POST/ESTADO_GTR) | Empty/broken, no semantic value |
| 3 | Hidden cols H–M | Visual cleanliness; CEO only sees A–G |
| 4 | Smart dropdown on A: 25 items "nombre" only (not "DNI - nombre") | Removes DNI duplication since col B shows DNI |
| 5 | B formula: `INDEX/MATCH` on name → DNI | CEO picks by name; DNI auto-derived |
| 6 | PERSONAL!B DNI: NUMBER → TEXT format | VLOOKUP from text → text works; preserves leading zeros (`09887772` not `988772`) |
| 7 | Col D format: `DATE` `d/M/yyyy` | Was showing raw serial (46252); now `18/8/2026` |
| 8 | Removed residual data validation from col B | Old "NOMBRES" dropdown artifact leaked through earlier refactor |
| 9 | Re-applied bold white text on row 3 headers | Was lost during wipe step |
| 10 | Merged ranges cleaned | Old `A1:O1` merge conflict resolved; row 1 + row 2 status rebuilt clean |

### CRIT-009 lessons (this session)

1. **`deleteRange` with `shiftDimension: COLUMNS` is destructive** — it removed column A and shifted left, breaking every formula that referenced `A`. Never use that for "removing data validation". A safer fix: `setDataValidation` with `CUSTOM_FORMULA =1=1` (no-op).
2. **API batches can silently fail** — my initial title+status+headers batch returned `replies: 0` due to a merge conflict, but other requests still ran. Always verify with `get values` after complex batches.
3. **PERSONAL!B was NUMBER, not TEXT** — even though it displayed `09887772` with leading zero, the underlying value was `9887772` (number). When you write `"09887772"` as text from another sheet, `VLOOKUP` does not match. Fix: re-write column as `userEnteredValue.stringValue` with `numberFormat.type = TEXT`.

### Live tests verified (3/3 PASS)

| Scenario | A | B | D | E | F | G |
|---|---|---|---|---|---|---|
| Standard name | WONG CORREA PATRICIA DEL PILAR | 41418003 | 18/8/2026 | 12:00 - 21:00 | WILLIAMS… | REQ-001 |
| Leading zero | URIBE DAVILA MONICA PAHOLA | 09887772 | 18/8/2026 | 12:00 - 21:00 | WILLIAMS… | REQ-002 |
| Invalid name | NO EXISTE EN BD | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | REQ-003 |

### Next-session pre-loaded context

- Spreadsheet ID is unchanged; CEO opens directly.
- **Apps Script still bound as orphan** — not harmful, no code in it (`feat(cleanup): Apps Script v9 eliminated` 43be454).
- No new CRIT entries this session (no decision was missed).
- Worktree `feature/c1ma-desing` is clean; ready for next session or `owd` when CEO signs off.

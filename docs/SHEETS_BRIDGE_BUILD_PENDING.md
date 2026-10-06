# CIMA sheets bridge — standalone build pending

## Status

The sheets bridge (`go-runtime/cmd/sheets/`) **does NOT compile as a standalone
CIMA Go module** in its current form. It has 2 known issues:

### Issue 1: Missing functions

The bridge references functions that don't exist in the migrated codebase:

- `colLetter(col int) string` — A1 column letter conversion
- `colLetterInverse(letters string) int` — A1 column letters → index
- `ScriptClient` type — Apps Script API client
- `allowlistPath` — allowlist path resolution

These were originally defined in files that were lost during the OVAV → CIMA
migration. The bridge worked inside OVAV because those helper files lived
in the wider OVAV codebase. Once extracted, the references are dangling.

### Issue 2: Internal package dependency

`sheets_crypto.go` referenced `github.com/ovav/ovav/internal/vault`, which is
not accessible from outside the OVAV module (Go `internal/` rule).

**Workaround applied:** CIMA now has its own `go-runtime/internal/vault/`
package (AES-256-GCM with stdlib crypto). `sheets_crypto.go` was updated to
import `cima/go-runtime/internal/vault`. ✅

## What works

- `go-runtime/cmd/citrix_upload/` — **compiles standalone** ✅
  - Independent `credStore` (no sheets bridge import)
  - Uses CIMA's `internal/vault/`
  - Tested 2026-10-06: vault loads, xlsx reads, Drive API call issued
  - 401 from Google = expired access_token (operational, not compilation)

- `go-runtime/internal/vault/` — **compiles standalone** ✅
  - Stdlib crypto (AES-256-GCM)
  - 4 functions: `KeySize`, `Encrypt`, `Decrypt`, `GenerateKey`

## What doesn't work

- `go-runtime/cmd/sheets/` — needs refactor to:
  1. Add missing helper functions (colLetter, ScriptClient, allowlistPath)
  2. OR remove references and reduce functionality
  3. OR depend on OVAV runtime at compile time via `go.work`

## Path forward

**Option A** (recommended): Use OVAV's sheets binary directly.
- Re-add `cmd/sheets/` to OVAV runtime (it was migrated but is no longer
  compiled there)
- CIMA consumes it via `go install` or path
- Minimal disruption

**Option B** (clean): Refactor CIMA sheets bridge.
- Identify which sheets features CIMA actually needs (HE/TE/FEEDBACK/CITRIX)
- Strip the rest
- Add missing helpers (colLetter, ScriptClient, allowlistPath)
- Estimated: 2-3 sessions of focused work

**Option C** (hybrid): Keep sheets bridge as-is, document as build-pending.
- CIMA deploys with sheets code but doesn't build it
- citrix_upload is the only CIMA-built binary
- Future sessions can do Option A or B

**Current state: Option C.** The sheets code is preserved for future
refactor; only `citrix_upload` compiles standalone.

---

*Thavren — 2026-10-06 — discovered during CIMA standalone build verification*
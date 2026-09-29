# CIMA HE Lima v0.3.0 — Runtime Guide

> **Owner**: thavren (lead) · marco (architect) · andres (implementer)
> **Audience**: Cualquier persona que necesite operar/desarrollar el sistema
> **Status**: STABLE

Esta guía documenta cómo operar CIMA HE Lima v0.3.0 sin OAuth
(Estado actual del entorno — bloqueante OAuth para `cima-verify`).

---

## 🚀 Quick Start

```bash
# 1. Validar todo el sistema (3 steps, ~3 segundos)
go run -C go-runtime ./cmd/sheets cima-ci
# Expected: ✅ ALL CHECKS PASSED — ready for deploy

# 2. Ver schema canónico
go run -C go-runtime ./cmd/sheets cima-schema

# 3. Generar bundle deploy-ready (CEO puede usar YA)
go run -C go-runtime ./cmd/sheets cima-bundle --output ./build/cima-bundle
```

---

## 📋 Subcommands CIMA — Catálogo completo

| Command | OAuth | Purpose | Use case |
|---|---|---|---|
| `cima-schema` | ❌ | Print canonical schema | Onboarding, docs |
| `cima-schema --section sheets` | ❌ | 18 sheets expected | Deployment planning |
| `cima-schema --section columns` | ❌ | 22 columns 10_Registro_HE | Header validation |
| `cima-schema --section modules` | ❌ | 9 Apps Script modules | Module inventory |
| `cima-schema --section anchors` | ❌ | 00_Config cell anchors | Config debugging |
| `cima-modules` | ❌ | Verify local .gs files | Pre-deploy check |
| `cima-lint` | ❌ | Static analysis (braces, functions, anti-patterns) | Code review |
| `cima-test` | ❌ | 86 integrity checks | Pre-commit, CI |
| `cima-drift` | ❌ | Code vs canonical schema | Architecture compliance |
| `cima-ci` | ❌ | Full pipeline (lint + test + modules) | CI/CD |
| `cima-bundle` | ❌ | Generate deploy bundle | Day 5 prod deploy |
| `cima-verify` | ✅ | Verify LIVE spreadsheet | Post-deploy validation |

**12 subcommands · 11 sin OAuth · 1 con OAuth**

---

## 🧪 Tests Go runtime

```bash
# Run all sheets tests
go test ./cmd/sheets/ -v -count=1 -timeout 60s

# Coverage
go test ./cmd/sheets/ -cover -count=1
# Expected: ~12-15% overall (mostly HTTP client code, hard to test)
```

### Test counts (Fase 6-10)

| Test file | Tests | Coverage |
|---|---|---|
| `cima_test.go` | 16 | Schema + module validation |
| `cima_test_harness_test.go` | 17 | Pascua math, holidays, HE formulas |
| `cima_lint_test.go` | 6 | Static analysis |
| Pre-existing sheets tests | ~20+ | HTTP client, formatting |
| **Total** | **~60** | All passing |

---

## 🚢 Deploy Workflow (3 paths)

### Path A — Manual via Apps Script editor (NO OAuth needed)

**Best for**: CEO sin acceso a service accounts.

```bash
# 1. Generate bundle
go run -C go-runtime ./cmd/sheets cima-bundle --output ./build/cima-bundle

# 2. Open bundle directory
ls ./build/cima-bundle/
# 00_AddOnMenu.gs 04_FormulasRecalc.gs 08_DBSync.gs
# 01_CfgLoader.gs 05_DashboardRefresh.gs appsscript.json
# 02_HolidaysEngine.gs 06_AuditLogger.gs  DEPLOY.md
# 03_RegistroEngine.gs 07_TriggersWiring.gs SHA256SUMS.txt

# 3. Open spreadsheet
xdg-open "https://docs.google.com/spreadsheets/d/1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"

# 4. Extensions → Apps Script
# 5. Create files in order 00..08, paste contents
# 6. Replace manifest
# 7. Save, reload sheet
# 8. Menu "CIMA HE Lima" appears
```

**Time**: 5-10 minutes
**Difficulty**: Easy (drag-drop)

### Path B — Apps Script API (requires OAuth)

**Best for**: CI/CD automation, prod deploys.

```bash
# 1. Provision OAuth (one-time)
# See STATUS_v3_stable.md OAuth section

# 2. Verify integrity locally
go run -C go-runtime ./cmd/sheets cima-test
# Expected: 86/86 checks passed

# 3. Push to Apps Script via API (TODO: not implemented yet)
go run -C go-runtime ./cmd/sheets scripts push --path ./build/cima-bundle
# Status: requires OAuth in vault

# 4. Verify live sheet structure
go run -C go-runtime ./cmd/sheets cima-verify
# Expected: 18/18 sheets present
```

**Time**: 30 seconds
**Difficulty**: Hard (OAuth setup)

### Path C — Hybrid (manual + automated verify)

**Best for**: Day 3-4 (data + sync phase).

```bash
# 1. Manual deploy via editor (Path A)
# 2. Verify with Go CLI (only cima-verify needs OAuth)
# 3. If OAuth available: full automation
```

---

## 🛠️ Troubleshooting

### Error: "webapp block forbidden"

**Symptom**: `cima-lint` reports webapp block in manifest.

**Cause**: Old deploy with `webapp.access: DOMAIN` block.

**Fix**: Run `cima-bundle` to regenerate correct manifest. Deploy the bundle.

### Error: "Partition_Key not found"

**Symptom**: Apps Script fails to find column `Partition_Key`.

**Cause**: Old typo (Day 2.2 fix).

**Fix**: 
- `cima-lint` should catch this
- Re-deploy `deploy_he_lima.gs` (already fixed in v0.3.0)
- OR `go run -C go-runtime ./cmd/sheets cima-drift` for full diagnostic

### Error: "manifest required function onOpen not declared"

**Symptom**: Runtime failure when opening sheet.

**Cause**: Apps Script manifest declares `addOns.homepageTrigger.runFunction="onOpen"` but `00_AddOnMenu.gs` missing.

**Fix**: `cima-bundle` includes 00_AddOnMenu.gs — re-deploy bundle.

### Error: "oauthScopes count = 5"

**Symptom**: Apps Script rejects manifest for insufficient scopes.

**Cause**: Reverted to 7+ scopes by accident.

**Fix**: Day 2.2 mandated 5 scopes. Re-deploy manifest from bundle.

---

## 📊 Architecture Reference

### 9 Apps Script modules (canonical)

```
00_AddOnMenu.gs          Entry point + menú nativo + bootstrap (REQUIRED)
01_CfgLoader.gs          Lee 00_Config + CacheService (REQUIRED)
02_HolidaysEngine.gs     Pascua Meeus/Jones/Butcher (REQUIRED)
03_RegistroEngine.gs     upsert + invariantes (REQUIRED)
04_FormulasRecalc.gs     onEdit router + alertas (REQUIRED)
05_DashboardRefresh.gs   KPIs + heatmap (optional)
06_AuditLogger.gs        ISO 8601 + auto-archive (REQUIRED)
07_TriggersWiring.gs     3 triggers + LockService (optional)
08_DBSync.gs             load/save/createMonth (REQUIRED)
```

### 18 sheets schema

```
00_Config         (panel control)
01_Feriados_PE    (13 feriados D.L. 713)
02_Personas       (maestro personas)
10_Registro_HE    (working surface 110×22)
20_Dashboard      (KPIs)
99_Audit_Log      (ISO 8601)

98_DB_2026_01..12 (storage particionado, lazy creation)
```

### 22 columns 10_Registro_HE

A=Fecha · B=ISO_Week · C=Dia_Semana · D=Es_Feriado · E=Tipo_Dia ·
F=ID_Persona (FK) · G=Apellidos_Nombres · H=Turno · I=Hora_Ingreso ·
J=Hora_Salida · K=Horas_Total · L=Horas_Normales · M=Horas_Extras ·
N=Factor_Aplicado · O=Horas_Ponderadas · P=Pago_HE_PEN ·
Q=Observaciones · R=Aud_Usuario · S=Aud_Timestamp · T=Estado ·
U=Mes_Año · V=Sync_Status

### 10 anchors 00_Config

B2=Mes · B3=Año · B4=Semana · B5=Persona · B6=FactorDiurno ·
B7=FactorNocturno · B8=FactorFeriado · B9=JornadaMax · B11=LastRefresh

---

## 🔒 Security Notes

| Concern | Status |
|---|---|
| oauthScopes count | ✅ 5 (least privilege, Day 2.2) |
| webapp block | ✅ Removed (Day 2.2) |
| LockService en ops | ✅ save, load, sync, trigger install, audit |
| Plaintext secrets | ✅ None in code |
| Service account | ⏸️ Not configured (OAuth needed) |

---

## 📞 Squad

| Persona | Rol | GitHub |
|---|---|---|
| **Andrés** | Senior — Apps Script + deploy | team-andres |
| **Lucas** | Junior — fixtures, migrations | team-lucas |
| **Clara** | QA — tests, regression | team-clara |
| **Thavren** | Lead — gates, CEO comms | area-platform-engineering |

---

*OVAV v0.3.0-stable · Generated 2026-09-29 · thavren*
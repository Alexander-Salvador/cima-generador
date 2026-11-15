# CIMA

**Call center Intelligent Management & Analytics**

Sistema de control de agentes para supervisor de CALL CENTER — campaña técnica MOVIL.
Identidad del grupo WhatsApp del equipo: **CIMA**.

📖 **[Leer AGENTS.md](./AGENTS.md)** para entender qué es CIMA y cómo trabajar.
📚 **[Leer MEMORY.md](./MEMORY.md)** para decisiones arquitectónicas, reglas y lessons learned.
📋 **[Leer PENDIENTES.md](./PENDIENTES.md)** para el estado vigente y las decisiones abiertas (D1–D8).

---

## Quickstart

### Popup CIMA (frontend principal)

> ⚠️ **Origen canónico: `localhost`.** Un solo hostname, un solo puerto, para
> siempre. `127.0.0.1` e IPs de LAN están **prohibidos** — son origins distintos
> con state de browser separado (Service Workers, Cache Storage, localStorage) y
> las IPs de LAN por HTTP plano no son *secure context* (`crypto.subtle` y
> `serviceWorker` no existen). Ver la ley en [AGENTS.md](./AGENTS.md).

```bash
# Servidor local con anti-cache (puerto 8765 — NO 8000)
cd apps/cima-generador && python3 serve.py

# Abrir SIEMPRE por localhost
#   Popup (Google Sites): http://localhost:8765/index.html?popup=1
#   Popup + diagnóstico: http://localhost:8765/index.html?popup=1&debug=1
#   Launcher:             http://localhost:8765/launcher.html
#   Salud:                http://localhost:8765/__health

# Gate completo (7 suites)
cd apps/cima-generador
python3 test_popup.py                # 866 asserts — UI/CSS/JS + zonas + etiquetas
python3 test_aggressive.py           # CCPP + fuzzy matching (238/238)
python3 test_security.py             # secrets scan
python3 test_codereview.py           # listeners/leaks
python3 test_browser_static.py       # estáticas del DOM
python3 test_ingest_comunicado_zonas.py   # 76 asserts — ingesta por comunicado
python3 test_sweep_zonas.py          # 19 checks — barrido de seguridad de datos

# Refresh: CTRL+SHIFT+R (limpia template + bypass cache)
# ⚠️  Si cambiaste CSS: un solo `}` de más rompe TODO el stylesheet.
#     El test suite valida balance de llaves — NO lo ignores.
```

**Si el popup muestra una versión vieja** (el footer no coincide con el chip de
build): abrí incógnito (`Ctrl+Shift+N`). El origin limpio confirma si el código
sirve o si hay estado residual del browser.

### Zonas de apagado (datos + auditoría)

```bash
cd apps/cima-generador
python3 scripts/audit_zonas_apagado.py    # fuente (3 PDFs + manual) ↔ runtime
python3 scripts/security_sweep_zonas.py   # 21 checks anti-datos-falsos (gate)
python3 scripts/export_tabla_zonas.py     # tabla visual XLSX + resumen MD
```

Artefactos (locales, **no versionados** por ser derivados de PDFs internos):
`data/apagados_tabla_completa.csv` (2.081 filas × 28 col) ·
`data/TABLA_ZONAS_APAGADO_2026-11-14.xlsx` (7 hojas) ·
`data/security_sweep_zonas.json`.

### Sheets bridge (backend legacy)

```bash
cd go-runtime && go build ./cmd/sheets/...
cd go-runtime && go test ./cmd/sheets/... -v
cd go-runtime && go run ./cmd/citrix_upload/ --help
```

## Componentes

- **HE** (Horas Extras) — semanal, rollover automático
- **TE** (Turno Especial) — FT only, 6h/día o TE
- **FEEDBACK** — supervisor → agente con workflow
- **CITRIX** — RRHH, archivo separado
- **Popup CIMA** (`apps/cima-generador/`) — generador de plantillas para call center:
  Address (ubigeo + CCPP) + Estructura de Llamada (speeches por motivo) +
  Etiquetas de transferencia (39, contra el Excel) + **Zonas de apagado**

## Zonas de apagado (estado 2026-11-14)

| Campaña | Qué es | Cobertura | Fecha |
|---|---|---|---|
| **COMU-5514** | Apagado de estaciones base (3 grupos) | **238 CCPP exactos** en 191 distritos | inicio de campaña 11/09/2026 · **fecha por zona pendiente** (recurso interno) |
| **COMU-5672** | Apagado de red 2G / zonas IPT | **1.798 zonas** por distrito (21 deptos) | 28-ago / 03-sep-2026 **por departamento** |
| **COMU-5569** | Campaña diciembre 2026 | ⏳ **0 CCPP** (falta la tabla de zonas) | inicio 10-dic-2026 |

> El popup **nunca afirma una fecha de zona sin verificarla**: si la fecha exacta no
> está confirmada muestra el inicio de campaña etiquetado + aviso, y el speech pide
> la fecha del centro poblado. Gate que lo protege: `test_sweep_zonas.py` (S20/S21).

Detalle: [`apps/cima-generador/docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md`](./apps/cima-generador/docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md) ·
[`apps/cima-generador/docs/TABLA_ZONAS_APAGADO_RESUMEN_2026-11-14.md`](./apps/cima-generador/docs/TABLA_ZONAS_APAGADO_RESUMEN_2026-11-14.md)

---

## Deploy (GitHub Pages)

El popup se sirve como **repo independiente**: `apps/cima-generador/` → `cima-generador` (GitHub Pages).

⚠️ **Requiere OK explícito del CEO por push** (CRIT-022 / directiva LOCAL FIRST).

```bash
# 1. Push develop (refspec explícito)
git push origin develop:develop

# 2. Subtree split → branch popup-deploy
git subtree split --prefix=apps/cima-generador/ develop -b popup-deploy

# 3. Force-push a main (GitHub Pages source)
OVAV_BYPASS_DRIFT_CHECK=1 git push --no-verify --force origin popup-deploy:main

# 4. Esperar build (~60s) y verificar
gh api /repos/Alexander-Salvador/cima-generador/pages/builds/latest
curl -s https://alexander-salvador.github.io/cima-generador/ \
  | grep -oE 'footer-version[^>]*>[^<]+'
```

**LIVE:** https://alexander-salvador.github.io/cima-generador/

> **Estado (2026-11-14):** local **v3.10.86** (registro de campañas · barrido de
> zonas · fecha por zona · tabla XLSX) — **commit local, NO publicado**.
> **LIVE = v3.10.82.** Sin push hasta OK explícito del CEO.
> Bloqueo abierto: la tabla de zonas de **COMU-5569** (los 4 PDF cargados son el
> mismo comunicado) y las **fechas por zona de COMU-5514** (D8) — ver
> [`PENDIENTES.md`](./PENDIENTES.md).

---

*OVAV-migrated: 2026-10-05* · *Popup local v3.10.86 (sin publicar) = LIVE v3.10.82 (2026-11-07)*
*Handoff de sesión: [`docs/handoff/2026-11-14-comu5569.md`](./docs/handoff/2026-11-14-comu5569.md)*
*Origen canónico de desarrollo: `http://localhost:8765/index.html?popup=1`*

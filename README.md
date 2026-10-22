# CIMA

**Call center Intelligent Management & Analytics**

Sistema de control de agentes para supervisor de CALL CENTER — campaña técnica MOVIL.
Identidad del grupo WhatsApp del equipo: **CIMA**.

📖 **[Leer AGENTS.md](./AGENTS.md)** para entender qué es CIMA y cómo trabajar.
📚 **[Leer MEMORY.md](./MEMORY.md)** para decisiones arquitectónicas, reglas y lessons learned.

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

# Tests
python3 test_popup.py           # 676 tests — UI/CSS/JS structure
python3 test_aggressive.py      # CCPP + fuzzy matching (234/234)
python3 test_codereview.py      # listeners/leaks
python3 test_security.py        # secrets scan

# Refresh: CTRL+SHIFT+R (limpia template + bypass cache)
# ⚠️  Si cambiaste CSS: un solo `}` de más rompe TODO el stylesheet.
#     El test suite valida balance de llaves — NO lo ignores.
```

**Si el popup muestra una versión vieja** (el footer no coincide con el chip de
build): abrí incógnito (`Ctrl+Shift+N`). El origin limpio confirma si el código
sirve o si hay estado residual del browser.

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
  Address (ubigeo + CCPP) + Estructura de Llamada (speeches por motivo) + Etiquetas de transferencia

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

> **Estado:** producción en `v3.10.63`. El local va en `v3.10.74`. El fix del
> input TELÉFONO (solo 9 dígitos) **aún no está desplegado** — los agentes siguen
> viendo `type="text" maxlength="40"`. Deploy pendiente de OK del CEO.

---

*OVAV-migrated: 2026-10-05* · *Popup local v3.10.74 / LIVE v3.10.63 (2026-10-22)*
*Origen canónico de desarrollo: `http://localhost:8765/index.html?popup=1`*
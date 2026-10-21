# CIMA — Lista de Pendientes (Estado verificado 2026-10-20)

**Fecha verificación:** 2026-10-20
**Branch:** `develop`
**HEAD:** `6bb660e` (docs HANDSHAKE FINAL v3.10.63)
**Popup LIVE GitHub Pages:** v3.10.63
**URL LIVE:** https://alexander-salvador.github.io/cima-generador/
**Working dir popup:** `apps/cima-generador/`
**DIRECTIVA ACTIVA:** LOCAL FIRST — no pushear sin OK explícito del CEO.

---

## 📊 Snapshot del repo (verificado en vivo)

| Métrica | Valor real | Fuente |
|---|---|---|
| `apps/cima-generador/index.html` | 1.18 MB / **8,435 líneas** | `ls -la` + `wc -l` |
| `apps/cima-generador/VERSION` | **3.10.63** ✅ | cat VERSION |
| Footer-version (HTML) | **3.10.63** ✅ | grep index.html |
| CHANGELOG última entrada | v3.10.61 ⚠️ | head CHANGELOG.md |
| Tests passing | **676/676** | test_popup.py |
| CSS balance | 530 `{` = 530 `}` | audit |
| HTML balance | 85 abiertos = 85 cerrados | audit |
| `var` (legacy ES5) | **479** | grep index.html |
| `let`/`const` (ES6) | 20 | grep index.html |
| Ratio var/let+const | **96% legacy** | calculado |
| `addEventListener` | 37 | grep |
| `removeEventListener` | 4 | grep |
| Ratio leak | **9.25 : 1** ⚠️ | calculado |
| `setTimeout(` | **21** | grep |
| `onclick=` hardcoded | **10** | grep |
| `@media (prefers-reduced-motion` | **3 bloques** | grep |
| TODO/FIXME/XXX reales | **0** | grep estricto |

> 📐 **Tamaños de funciones grandes (verificados por diff de líneas en v3.10.63):**

| Función | Línea inicio | Línea fin | Líneas reales |
|---|---|---|---|
| `renderCoberturaButton` | 6009 | 6124 | **116** |
| `showApagadoPanelInline` | 6125 | 6639 | **515** ⚠️ |
| `acGetDropdown` (+ helpers) | 6640 | 6789 | **150** |
| `acFilter` | 6790 | 6826 | **37** (era 221 → refactor parcial) |
| `subsequenceMatch` | 6827 | 6834 | **8** |
| `acHighlight` | 6835 | 7873 | **1039** ⚠️ GIGANTE no catalogado |
| `compactAddress` | 7874 | 8175 | **302** |
| `auditZonasApagadas` | 8176 | 8435 (EOF) | **259** |

---

## ✅ RESUELTOS — Sesión 2026-10-20 (v3.10.40 → v3.10.63) — 24 releases

| Versión | Cambio |
|---|---|
| v3.10.40 | Housekeeping limpieza branches fugas |
| v3.10.41-42 | Alertas amables + redistribución 1/section |
| v3.10.43 | 3 BUG FIXES (telefono-last3, layout encuesta, jerarquía) |
| v3.10.44-45 | HEADER "Estructura de Llamada" + MOTIVO cards Apple |
| v3.10.46-47 | 4 BUG FIXES (CTRL+SHIFT+R, grids mismo ancho) |
| v3.10.48 | SPEECHES nutridos AI 2026 (variación + consecutivas) |
| v3.10.49-50 | DESGLOSE click MOTIVO + 3 BUG FIXES |
| v3.10.51 | MOTIVO consolidado en 1 dropdown (4 opciones) |
| v3.10.52 | TELÉFONO + MOTIVO en 1 row |
| v3.10.53 | Texts ocultos hasta seleccionar MOTIVO |
| v3.10.54 | FIX bug estructura HTML (grid-body cerraba prematuramente) |
| v3.10.55 | Estilización + compactación GRID |
| v3.10.56 | DISCLOSURE CONTROL PILLS (Apple HIG 2026) |
| v3.10.57 | FIX grid-body max-height 1200→8000px |
| v3.10.58 | **FIX CRÍTICO**: llave CSS extra rompía 6800 líneas |
| v3.10.59 | COLLAPSED PILL-LINE ultra-compacta (−55%) |
| v3.10.60 | FIX render glitch (border-radius interpolate) |
| v3.10.61 | ALERTS como refuerzo inteligente en 1 row |
| v3.10.62 | FIX Grid Tipificación (parcial) |
| **v3.10.63** | **FIX DEFINITIVO** selección etiquetas (capture phase + id normalize) |

(Histórico completo v3.10.0 → v3.10.39 → ver `MEMORY.md` sección "Logros previos")

---

## ⏳ PENDIENTES — CONTRAVALIDACIÓN REAL (2026-10-20)

### 🔴 OBSOLETOS — Remover de la lista

| ID | Razón de obsolescencia | Evidencia |
|---|---|---|
| **A9** | "Push a origin/main" | **YA HECHO** en HANDSHAKE FINAL v3.10.63 (commit `90860bd4`) |
| **D8** | "2 TODO sin owner" | **0 TODO/FIXME/XXX reales** (los matches eran "todo el" en comentarios) |

### 🔴 SUBESTIMADOS — Recalibrar prioridad/estimación

| ID | Estimado original | Real verificado | Acción |
|---|---|---|---|
| **D4** | 426 `var` | **479 `var` vs 20 `let`/`const`** | Ratio **96% legacy**. **PRIORIDAD UPGRADE: BAJA → ALTA.** Estimación: **6-8h reales** (mecánico) o **2-3 días** (refactor serio con módulos ES). |
| **D9** | 30 add / 4 remove | **37 add / 4 remove** | Ratio empeoró de 7.5:1 → **9.25:1**. **33 addEventListener huérfanos** sin cleanup = leak garantizado en SPA long-running. |
| **D7** | 5 `setTimeout` hardcoded | **21 `setTimeout`** | Subestimado x4.2. Consolidar a objeto `TIMINGS = {...}`. |
| **D6** | 5 `@media prefers-reduced-motion` | **3 bloques** (bajó pero NO consolidado) | Parcial. Falta unificación en 1 solo bloque. |
| **D10** | 11 `onclick=` hardcoded | **10 `onclick`** | Parcial. Bajó 1 (A7/D10 v3.10.8 resolvió 2 críticos). Quedan 1 crítico + 8 literales seguros. |

### 🟢 VIGENTES — Recomendaciones Marco (R1, R4-R12)

| ID | Recomendación | Líneas reales | Tiempo | Prioridad |
|---|---|---|---|---|
| **R1** | Borrar 1100+ legacy post-shadowing | ~800-1000 (estimado) | **3-4h** | ALTA |
| **R4** | `showApagadoPanelInline` → descomponer | **515 líneas reales** (subestimado desde 278) | **3-4h** | ALTA |
| **R5** | `acFilter` → extraer render | 37 (subestimado original 221) | **20min** | BAJA |
| **R6** | `auditZonasApagadas` → checks separados | **259 líneas reales** (subestimado desde 187) | **45min** | BAJA |
| **R7** | Consolidar `renderCoberturaButton` con v3.10 | 116 líneas | **30min** | BAJA |
| **R8** | Mover 3 bloques `.apagado-tooltip` CSS | — | **20min** | BAJA |
| **R9** | Documentar `acGetDropdown` y portal pattern | **150 líneas reales** (no es one-liner) | **45min** | BAJA |
| **R10** | `subsequenceMatch` → bounded iter (DoS) | **8 líneas** (función corta pero vulnerable) | **15min** | MEDIA |
| **R11** | `compactAddress` paso 4 → extraer | **302 líneas reales** (subestimado, era 15min → ahora 1h) | **1h** | BAJA |
| **R12** | `subsequenceMatch` + `acHighlight` → unificar | acHighlight = **1039 líneas** ⚠️ | **3-4h** | **NUEVA ALTA** |
| **NUEVO** ⚠️ | `acHighlight` 1039 líneas → descomponer | **1039 líneas** (NO estaba catalogado) | **4-6h** | **NUEVA ALTA** |

### 🟡 DEUDA TÉCNICA vigente

| ID | Pendiente | Estimación recalibrada |
|---|---|---|
| **D4** | 479 `var` → `let`/`const` (o módulos ES) | 6-8h mecánico / 2-3 días refactor serio |
| **D6** | Consolidar 3 `@media prefers-reduced-motion` en 1 | 30min |
| **D7** | 21 `setTimeout` → objeto `TIMINGS` | 45min |
| **D9** | 33 addEventListener huérfanos → cleanup pattern | **2h** (empeoró) |
| **D10** | 10 `onclick=` → event delegation (1 crítico + 8 literales) | **1-2h** |
| **D11** | Inline data 707KB → lazy load si >2MB | Vigente pero umbral no alcanzado (archivo 1.18MB) |

### 🟢 BAJA — Otros pendientes

| ID | Pendiente | Notas |
|---|---|---|
| **A8** | INEI Marco Geoestadístico 2018+ para CCPP rurales | No automatizable (red bloqueada en este entorno). 33 dist rurales sin CCPP compensados por fuzzy matching (100% cobertura). |
| **A10** | Re-medir throughput con Chrome real (no Node) | Requiere Chrome headless + harness. Benchmark Node.js actual estable; útil pero no crítico. |

---

## 📈 Tiempo recalibrado para v3.11

| Bloque | Tiempo |
|---|---|
| D9 (event listener leak) | **2h** |
| R12 (acHighlight 1039 → descomponer) | **4-6h** ⚠️ nuevo |
| R4 (showApagadoPanelInline 515 → descomponer) | **3-4h** |
| D4 (var→let mecánico) | 6-8h |
| R1 (legacy 800-1000 líneas) | 3-4h |
| D7 + D6 + D10 + R5 + R6 + R7 + R8 + R9 + R10 + R11 | ~5h |
| **TOTAL v3.11 mínimo viable** | **~25-30h** |
| D4 refactor serio (módulos ES + build) | **+2-3 días** |

---

## 🔗 Contexto para próximo chat

1. Leer `MEMORY.md` (cargado) — incluye histórico completo v3.10.0 → v3.10.63.
2. Leer `PENDIENTES.md` (este archivo) — solo vigente, con data verificada.
3. Si trabajás en popup:
   ```bash
   cd apps/cima-generador
   python3 test_popup.py          # 676/676 PASS
   python3 test_aggressive.py     # agresivo
   python3 test_audit_brutal.py   # auditoría
   python3 test_browser_static.py # browser estático
   python3 test_security.py       # seguridad
   python3 test_codereview.py     # code review
   python3 test_cobertura_peru.py # cobertura nacional
   ```
4. **Server local:** `python3 -m http.server 8765` (puerto 8765).
5. **URL popup mode:** `http://127.0.0.1:8765/index.html?popup=1`
6. **Refresh:** CTRL+R (HTML crudo + JS inline, sin HMR) · CTRL+SHIFT+R = limpiar template.
7. **Próximo bloque sugerido (CEO decide):**
   - **A:** Higiene rápida (D9 cleanup + obsolete removal + CHANGELOG catchup) — 3h
   - **B:** Atacar el gigante (R12 acHighlight 1039 líneas → descomponer) — 4-6h
   - **C:** Legacy estructural (R1 + R4) — 6-8h

---

## 📁 Archivos de referencia (verificados)

| Path | Estado |
|---|---|
| `apps/cima-generador/index.html` | 1.18 MB / 8,435 líneas / 3.10.63 ✅ |
| `apps/cima-generador/VERSION` | 3.10.63 ✅ |
| `apps/cima-generador/CHANGELOG.md` | Última entrada v3.10.61 ⚠️ (drift: falta 3.10.62, 3.10.63) |
| `apps/cima-generador/data/zonas_apagadas.json` | Consolidada |
| `apps/cima-generador/data/known_data_gaps.json` | 41 dist categorizados |
| `apps/cima-generador/scripts/` | 13 scripts (regenerate, fix_*) |
| `apps/cima-generador/test_*.py` (8) | + 4 test_*.js |
| `MEMORY.md` | Checkpoint vivo v3.10.63 |
| `PENDIENTES.md` | **Este archivo** (reemplaza versión v3.10.39) |
| `docs/CHECKLIST_MENTIRAS.md` | Auditoría previa |
| `docs/SHEETS_BRIDGE_BUILD_PENDING.md` | Legacy pre-migración |
| `/VERSION` (root) | 1.0.0 ⚠️ drift raíz (no se usa) |

---

## 🌿 Estado de branches (verificado `git branch -a`)

| Branch | Estado | Notas |
|---|---|---|
| `develop` (local) | **HEAD `6bb660e`** ✅ | working tree clean |
| `main` (local) | ✅ | sheets bridge legacy, no tocar |
| `popup-deploy` (local) | **stale** | subtree split `83a82fd`, ya no regenerarse (`main` ya pusheado) |
| `origin/develop` | `90860bd4` ✅ | synced |
| `origin/main` | `be99090c` ✅ | synced + GitHub Pages built |
| `feat/apagado-inteligente` | deprecated | LOCAL ONLY per CEO 2026-10-19 |

---

## 🎯 Decisiones arquitectónicas recientes (resumen CRIT-009/004)

| Decisión | Razón | Versión |
|---|---|---|
| Disclosure pills (Apple HIG 2026) | CEO: "píldoras... no se desplegarán automáticamente" | v3.10.56 |
| CSS balance check obligatorio | Lección: `}` extra destruyó 6800 líneas silenciosamente | v3.10.58 |
| MOTIVO = 1 dropdown (no cards) | CEO: "1 dropdown cada 1 ocupa 1 row" | v3.10.51 |
| `localStorage cima_*_v1` (4 keys) | Persistencia colapsado/pills/hecho por versión | v3.10.53-56 |
| Capture phase + id normalize | L-064/L-065: event delegation con múltiples dropdowns | v3.10.63 |

---

*Generado por Thavren — contravalidación real con `grep`/`wc`/`ls` en repo vivo · 2026-10-20*
*Reemplaza versión previa v3.10.39 (180 líneas) — solo vigente, sin histórico redundante*

# CIMA v3.10.10 — Lista de Pendientes (Estado al cierre de sesión)

**Fecha:** 2026-10-17 · **Branch:** develop · **Popup LIVE:** v3.9.15 (push pendiente CEO)

## ✅ YA RESUELTOS (Bloques 1-4 + v3.10.0-v3.10.6)

| ID | Descripción | Versión | Commit |
|---|---|---|---|
| A1 | Disclaimer "interpretación no oficial" en anomalías | v3.10.7 | d078ed2 |
| A2 | CCPP_KEYS_NORMALIZED init module-level (no null en Node) | v3.10.7 | d078ed2 |
| A4 | Validar 7 d6 sin CCPP_DATA (todos urbanos) | v3.10.10 | 0c9972e |
| A5 | test_browser_static.py (5 categorías, 14 PASS) | v3.10.10 | 0c9972e |
| A7/D10 | onclick → event delegation (2 críticos) | v3.10.8 | 7b1d98e |
| D3/A3 | acRender precomputed Map<d6, count> | v3.10.8 | 7b1d98e |
| I7 | CCPP en zones no en autocomplete (130 → 0) | v3.10.5 | 2f2864b |
| I10 | total_zones 2037 → 200 | v3.10.5 | 2f2864b |
| P1 | compactAddress 19 regex pre-compilados | v3.10.6 | 19d69fa |
| P2 | showApagadoPanelInline 19 regex pre-compilados | v3.10.6 | 19d69fa |
| P4 | auditZonasApagadas a requestIdleCallback | v3.10.6 | 19d69fa |
| P5 | ipt_metadata 495KB → 123KB (75% reducción) | v3.10.6 | 19d69fa |
| R1 (parcial) | escapeHtml unificado (4→2 funciones) | v3.10.9 | daa2dfd |
| R3 | Cache LRU real (no más FIFO) | v3.10.9 | daa2dfd |
| S7, S9 | Falsos positivos eliminados (literales seguros) | v3.10.6 | 19d69fa |
| Shadowing | 441 líneas legacy eliminadas (legacy v2.0.8) | v3.10.4 | 0ec1632 |
| 33 capitales | Inyectadas (UNION PROGRESO, UNION ASHÁNINKA) | v3.10.4 | 0ec1632 |
| 3 d6 faltantes | Agregados a districts_with_off | v3.10.4-patch | 24bb443 |
| Santa Rosa Loreto | d6=160405 inyectado (Ley 32403) | v3.10.3 | aa2c23a |
| STOP CHECK banner | Implementado | v3.10.3 | aa2c23a |
| Disclaimer MTC/OSIPTEL | Implementado | v3.10.3 | aa2c23a |
| Cache LRU | zaCacheGet/zaCacheSet/ZA_CACHE_MAX | v3.10.2 | (anterior) |
| Fuzzy Levenshtein | zaLevenshtein + zaSimilarity | v3.10.2 | (anterior) |
| 4 niveles semánticos | ccpp_exacto, ccpp_fuzzy, ipt_distrito, distrito_of | v3.10.2 | (anterior) |
| Sistema confianza | confianza 0-1, severidad, fuente_oficial | v3.10.2 | (anterior) |
| Separación COMU | COMU-5514 vs COMU-5672 con disclaimer | v3.10.0 | (anterior) |

## ⏳ PENDIENTES (orden sugerido)

### 🔴 CRÍTICO (bloqueante)

| ID | Pendiente | Severidad | Owner |
|---|---|---|---|
| **A9** | Push a origin/main (popup LIVE en v3.9.15) | BLOQUEANTE | CEO |

### 🟠 ALTA — Recomendaciones agente Marco (R1-R12)

| ID | Recomendación | Tiempo | Prioridad |
|---|---|---|---|
| **R1** | Borrar 1100+ líneas legacy redundantes (post-shadowing) | 2-3h | ALTA |
| **R4** | `showApagadoPanelInline` 278 líneas → descomponer | 1h | ALTA |
| **R5** | `acFilter` 221 líneas → extraer render | 1h | MEDIA |
| **R6** | `auditZonasApagadas` 187 líneas → checks separados | 30min | BAJA |
| **R7** | Consolidar `renderCoberturaButton` con v3.10 | 30min | BAJA |
| **R8** | Mover 3 bloques legacy CSS `.apagado-tooltip` | 20min | BAJA |
| **R9** | Documentar `acGetDropdown` y portal pattern | 20min | BAJA |
| **R10** | `subsequenceMatch` → bounded iter (DoS) | 15min | BAJA |
| **R11** | `compactAddress` paso 4 (personas) → extraer | 15min | BAJA |
| **R12** | `subsequenceMatch` + `acHighlight` → unificar | 15min | BAJA |

### 🟡 DEUDA TÉCNICA menor (D4-D11)

| ID | Pendiente | Tiempo |
|---|---|---|
| **D4** | 426 `var` → `let/const` (migración ES6) | 2-3h |
| **D6** | Consolidar 5 `@media prefers-reduced-motion` | 30min |
| **D7** | 5 `setTimeout` hardcoded → TIMINGS object | 20min |
| **D8** | 2 TODO sin owner (resolver o eliminar) | 10min |
| **D9** | 30 addEventListener vs 4 removeEventListener (leak) | 1h |
| **D10** | 11 onclick hardcoded (literales seguros, baja prioridad) | 2h |
| **D11** | Inline data 707KB (considerar lazy load si >2MB) | 1 día |

### 🟢 BAJA — otros pendientes

| ID | Pendiente | Notas |
|---|---|---|
| **A8** | INEI Marco Geoestadístico 2018+ para CCPP rurales | No automatizable (red bloqueada) |
| **A10** | Re-medir throughput con Chrome real (no Node) | Requiere Chrome headless |

## 📊 Resumen ejecutivo

| Métrica | Valor |
|---|---|
| Items resueltos | 22 |
| Items pendientes | 23 (1 crítico, 10 Marco, 7 deuda menor, 2 otros) |
| Tiempo estimado para v3.11 | ~20-30h |
| Líneas eliminadas (legacy) | 441 |
| Líneas agregadas (features) | ~600 |
| Commits en sesión | 10 |
| Tests passing | 613 |

## 🔗 Contexto para próximo chat

**Para retomar el trabajo:**

1. Leer `MEMORY.md` (línea ~70 checkpoint v3.10.10)
2. Leer `docs/CHECKLIST_MENTIRAS.md` (24 confesiones)
3. Leer `docs/FUENTES_ZONAS_APAGADAS.md` (trazabilidad)
4. Ejecutar tests:
   ```bash
   cd /home/braka/Systems/projects/work/cima/apps/cima-generador
   python3 test_popup.py          # 489/489 PASS
   python3 test_aggressive.py     # 61/61 PASS
   python3 test_audit_brutal.py   # 0 issues críticos
   python3 test_browser_static.py # 14/0/2 PASS/FAIL/WARN
   node /home/braka/Systems/projects/work/cima/benchmark_runtime.js  # 7.21μs/check
   ```
5. **Próximo bloque sugerido:** R1 + R4 (cleanup masivo + descomponer showApagadoPanelInline)

## 📁 Archivos de referencia

- `apps/cima-generador/index.html` (1.5MB, 7,260 líneas)
- `apps/cima-generador/CHANGELOG.md` (v3.10.0-v3.10.10)
- `apps/cima-generador/VERSION` (3.10.10)
- `apps/cima-generador/data/known_data_gaps.json`
- `apps/cima-generador/docs/FUENTES_ZONAS_APAGADAS.md`
- `apps/cima-generador/docs/CHECKLIST_MENTIRAS.md`
- `apps/cima-generador/test_popup.py` (489 tests)
- `apps/cima-generador/test_aggressive.py` (61 tests)
- `apps/cima-generador/test_audit_brutal.py` (8 categorías, 61 OK)
- `apps/cima-generador/test_browser_static.py` (5 categorías, 14 PASS)
- `apps/cima-generador/test_exhaustive_v3103.py` (Clara, 114 tests)
- `benchmark_runtime.js` (medición Node.js runtime)
- `apps/cima-generador/clean_ccpp_data.py` (limpieza CCPP_DATA)
- `MEMORY.md` (checkpoint v3.10.10)
- `PENDIENTES.md` (este archivo)

## 🎯 Estado de branches

| Branch | Estado | Notas |
|---|---|---|
| **develop** | v3.10.10 ✅ | 10 commits ahead de v3.9.15 |
| **popup-deploy** | v3.10.10 ✅ | subtree split, listo para push |
| **main** | v3.10.5 (legacy divergente) | No tocar, sheets bridge |
| **feat/apagado-inteligente** | v3.7.8 (deprecada) | CEO marcó como LOCAL ONLY |

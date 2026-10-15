# Changelog — cima-generador

## [2.1.7] - 2026-10-15 — BOUNDS-AWARE DROPDOWN + COMPACTACIÓN INTELIGENTE RESULTADO

### Issue 1 FIXED — Dropdown respeta límites del grid contenedor (no invade Tipificación)

**Síntoma reportado:** Tras el portal pattern v2.1.6, el dropdown de autocompletar (`.ac-dropdown` con `position: fixed; max-height: 220px`) excedía visualmente la altura del card "BÚSQUEDA ADDRESS", invadiendo el espacio del card "TIPIFICACIÓN" siguiente.

**Causa raíz:** Con `position: fixed`, el dropdown ya no respeta los bounds del `.section` (Búsqueda Address) porque vive en el body. Solo respeta `max-height: 220px` que es estático.

**Solución — BOUNDS-AWARE:**

- **`acFitToGrid(drop, input, inputRect)`** (NUEVO helper): calcula el espacio vertical disponible entre el input y el límite inferior del `.grid-block` contenedor (Búsqueda Address). Aplica `max-height` DINÁMICO respetando ese límite. Si el espacio es < 100px, FLIP-ARRIBA (dropdown se renderiza encima del input en vez de debajo).
- **`acUpdatePosition(key)`** ahora llama a `acFitToGrid()` después de calcular coords base — bounds se respetan también durante scroll/resize handlers.
- **`acRender(key)`** llama a `acFitToGrid()` después del portal move — bounds se respetan al abrir.
- **`acClose()` / `acCloseAll()`** limpian `style.maxHeight = ''` y `classList.remove('is-flipped')` — idempotencia preservada.

### Issue 2 FIXED — RESULTADO crece en altura con contenido largo

**Síntoma reportado:** El `output-box` (RESULTADO) crecía de 36px a 60-90px cuando mostraba una dirección larga. CEO directriz textual: *"el input RESULTADO crece en altura cuando tiene mucho contenido, y ya te había indicado que no debe cambiar su altura si no entra la información se debe compactar de manera inteligente, abreviándose"*.

**Solución — COMPACTACIÓN INTELIGENTE + altura FIJA:**

#### CSS — altura FIJA 36px (CEO directriz)
- **`.output-box`** ahora:
  - `height: 36px` (FIJA)
  - `min-height: 36px`
  - `max-height: 36px` (NO crece jamás)
  - `overflow: hidden` (corta overflow)
  - `text-overflow: ellipsis` (safety net CSS)
  - `white-space: nowrap` (1 línea solamente)
- **`.output-wrap`** sin max-height (igual que v2.1.4 — la altura viene del output-box).
- **Media query `≤480px`** ya NO sobrescribe `white-space: pre-wrap` ni `max-height: 88px` (regla conflictiva removida). Mobile-first aplica misma regla desktop.

#### JS — `compactAddress(address, maxChars)` progressive abbreviation

Aplica reglas stepwise hasta que la dirección quepa en `maxChars` (default 64):

1. **Prefijos de vía**: `JIRON`/`JIRÓN` → `Jr.`, `AVENIDA` → `Av.`, `CALLE` → `Calle`, `PASAJE` → `Psje.`, `PLAZA` → `Pza.`, `CARRETERA` → `Carr.`
2. **Nombres de zona**: `URBANIZACIÓN` → `Urb.`, `ASOCIACIÓN` → `Asoc.`, `PARCELACIÓN` → `Parc.`, `COMITÉ` → `Com.`, `CONJUNTO HABITACIONAL` → `C.H.`
3. **Niveles admin**: `DEPARTAMENTO` → `Dpto.`, dedup `"X, X"` (ej. `"LIMA, LIMA"` → `"LIMA"`)
4. **Personas**: 4 tokens all-caps → solo 2 primeros; 3 tokens → pre+name; nombres de pila largos → iniciales (`PABLO ANTONIO` → `P. ANTONIO`)
5. **Último recurso**: truncar en última coma natural con ellipsis `…`

#### JS — `updateAddress()` integración
- Llama `compactAddress(out, 64)` antes de asignar `textContent`.
- Setea `box.title = out` (full address en tooltip — accesibilidad WCAG).
- `box.dataset.full = out` preservado (clipboard copy).

### Walk-through visual (ejemplo CEO)

**Input:** `"AVENIDA JAVIER PRADO ESTE 4201, URBANIZACION SANTA CATALINA, SAN BORJA, LIMA, LIMA, PERU"` (88 chars)

**Output compacto (63 chars, 1 línea @ 36px):**
```
Av. J. PRADO ESTE 4201, Urb. S. CATALINA, SAN BORJA, LIMA, PERU
```

**Hover (title attr — full address):**
```
AVENIDA JAVIER PRADO ESTE 4201, URBANIZACION SANTA CATALINA, SAN BORJA, LIMA, LIMA, PERU
```

**Clipboard (dataset.full + copyOutput):**
```
AVENIDA JAVIER PRADO ESTE 4201, URBANIZACION SANTA CATALINA, SAN BORJA, LIMA, LIMA, PERU
```

### Issue 1 — Walk-through visual

**Click en DEPARTAMENTO** → dropdown aparece debajo del input con `max-height = min(220, gridBottom - inputBottom - 8)`.
- Si el espacio disponible es ≥ 220px → dropdown crece hasta 220px (cap superior preservado).
- Si el espacio es < 220px pero ≥ 100px → dropdown se ajusta al espacio (max-height dinámico).
- Si el espacio es < 100px → FLIP-ARRIBA — dropdown aparece ARRIBA del input con `is-flipped` class (animation: glassDropdownInFlip).

**Scroll/Resize:** `acUpdatePosition()` recalcula y aplica `acFitToGrid()` — bounds se respetan continuamente.

### Tests

- test_popup.py: 104 originales (preservados + actualizados a nowrap) + **16 nuevas aserciones v2.1.7** = **120 PASS**.
- 8 nuevas aserciones para `.output-box` (height/max-height/white-space/text-overflow/overflow/compactAddress invocations/title attr).
- 8 nuevas aserciones para dropdown bounds-aware (`.is-flipped` CSS, keyframe, transform-origin, acFitToGrid invocation, closest grid-block, flip threshold, is-flipped toggle, max-height dinámico, 220px cap).
- node --check: EXIT 0.
- test_security.py: 🟢 CLEAN.
- test_codereview.py: 🟢 GOOD.

### Preserved (no se revierte de v2.1.6)

- ✅ Portal pattern v2.1.6 (move a body + position:fixed + z-index 99999).
- ✅ `_cimaPortalHandlersInit` idempotencia guard.
- ✅ `acCloseAll` con cleanup completo de todos los dropdowns.
- ✅ v2.1.5 stacking fix defensivo (isolation:isolate + .field.is-open + :has() fallback).
- ✅ `glassDropdownIn` keyframe sin transform (Safari compositor safe).
- ✅ Cascada AC (departamento→provincia→distrito→ccpp) intacta.
- ✅ Mobile fallback <480px intacto (con fix de regla conflictiva).
- ✅ `prefers-reduced-motion` intacto.
- ✅ `updateAddress()` siendo reutilizada para clipboard (copyOutput) — preserva `dataset.full`.

### Notas de regresión

- Tests originales que assertan `white-space: pre-wrap` en `.output-box` → **actualizados a nowrap** (la nueva directriz CEO lo requiere). Esta es la regresión esperada: la regla pre-wrap era interpretación incorrecta del requerimiento, ahora corregida.
- `.output-wrap max-height` check ahora valida "libre" (sin max-height limitante).
- `.output-box min-height` check actualizado a `height: 36px + max-height: 36px`.

---

## [2.1.6] - 2026-10-15 — PORTAL PATTERN v2.1.6 (CRIT-009 estructural)

### Fixed CRIT-009 — DEFINITIVO
- **Problema:** Dropdown DEPARTAMENTO aparecía DETRÁS de inputs hermanos (PROVINCIA, DIRECCIÓN, REFERENCIA, RESULTADO). El fix v2.1.5 (`isolation: isolate` + `.is-open { z-index: 1000 }`) **NO resolvió** el problema.
- **Root cause REAL:** 4 niveles de stacking contexts anidados (`section → field → ac-wrap → input con backdrop-filter`). El `z-index: 1000` solo funciona dentro del propio ac-wrap; los inputs `backdrop-filter` de fields vecinos GANAN en el composite order (CSS painter's algorithm).
- **Solución:** Portal pattern (Radix UI / Headless UI / Floating UI). El dropdown se mueve a `document.body` y se posiciona con `position: fixed` + coords via `getBoundingClientRect()`.

### Changed CSS
- **`.ac-dropdown`** ahora `position: fixed` con `z-index: 99999` (super-high, body no tiene isolation). Eliminated `top: calc(100% + 4px); left: 0; right: 0;` (coords via JS).
- **Comentario v2.1.6** documenta la razón estructural (4 stacking contexts anidados).

### Changed JS — Portal move/restore
- **`acRender(key)`**: si `drop.parentNode !== document.body`, mueve el dropdown al body guardando `_originalParent` y `_originalNextSibling`. Calcula coords via `input.getBoundingClientRect()` y aplica `position/top/left/width/zIndex` inline.
- **`acClose()`**: limpia estilos inline (`position/top/left/width/zIndex = ""`) y restaura al parent original vía `_originalParent.insertBefore`.
- **`acCloseAll()`**: ahora también limpia estilos inline y restaura al parent original de TODOS los dropdowns (idempotencia garantizada).
- **`acUpdatePosition(key)`** (NUEVO): helper que recalcula coords via `getBoundingClientRect()`. No-op si dropdown está oculto o input detached.

### Added JS — Global event handlers (registrados una sola vez)
- **Guard `_cimaPortalHandlersInit`** (en `setupAutocomplete`): previene duplicación si se re-inicializa.
- **`window.addEventListener('scroll', repositionAll, true)`** (capture phase): captura scrolls en cualquier contenedor (incluso `field-row` con grid).
- **`window.addEventListener('resize', repositionAll)`**: reposiciona en resize de ventana.
- **`document.addEventListener('mousedown', ..., true)`** (capture phase): click-outside cierra todos los dropdowns si el target NO es `.ac-dropdown` ni `input[data-ac]`.
- **`document.addEventListener('touchstart', ..., true)`**: mobile fallback (Safari iOS a veces no dispara mousedown antes que blur).

### Defensive / robustness
- **Try/catch en `document.body.appendChild(drop)`**: si falla, deja dropdown en su sitio original (no rompe).
- **Try/catch en `getBoundingClientRect()`**: si falla (input detached), revierte portal move y aborta render.
- **Try/catch en `_originalParent.insertBefore`**: si falla, log warning pero continúa cerrando.
- **Console.warn** en todos los paths de fallo (debuggeability).

### Preserved (v2.1.5 stacking fix defensivo — no se revierte)
- ✅ `isolation: isolate` en `.section > *` y `.ac-wrap` — sigue siendo útil para isolation general.
- ✅ `.field.is-open { z-index: 1000 }` y `:has(.ac-dropdown:not([hidden]))` — defensiva en caso de fallback CSS.
- ✅ `glassDropdownIn` SIN transform (Safari compositor safe).
- ✅ `acCloseAll` con cleanup de `.is-open` (idempotente).
- ✅ Cascada AC (departamento→provincia→distrito→ccpp) intacta.
- ✅ Mobile fallback <480px intacto.
- ✅ `prefers-reduced-motion` intacto.

### Tests
- test_popup.py: 91 originales + **13 nuevas aserciones v2.1.6 PORTAL-specific** = **104 PASS**.
- node --check: EXIT 0.
- test_security.py: 🟢 CLEAN (sin cambios — solo se modificó popup.js + CSS).
- test_codereview.py: 🟢 GOOD.

---

## [2.1.4] - 2026-10-15 — REVERSIÓN v2.1.3 + RESULTADO ancho preformateado

### Changed
- **RESULTADO ancho + preformateado**: `.output-wrap` sin max-height; `.output-box` con `white-space: pre-wrap; word-break: break-word; overflow-wrap: break-word; min-height: 36px; max-height: 88px; overflow-y: auto`. Direcciones largas se muestran **completas** con wrap natural (multi-línea) en vez de truncarse a "AR D…" con ellipsis.
- **JS `updateAddress()`**: eliminada la llamada a `truncateForBox(out, 80)`. El texto completo se asigna a `textContent` (el wrap lo gestiona CSS). `dataset.full` se preserva para clipboard.
- **Botón BORRAR icon-only**: removido `<span class="btn-icon-label">BORRAR</span>`. Restaurado `width: 36px; min-width: 36px; padding: 0` (círculo pill 36×36). SVG papelera queda como único contenido visual.
- **Botón ZONAS DE APAGADO icon-only**: removido `<span class="cobertura-btn-label">ZONAS DE APAGADO</span>`. CSS `.cobertura-btn` ahora `flex: 0 0 36px; width: 36px; padding: 0` (simétrico con `.btn-icon-clear`).
- **Accesibilidad WCAG**: ambos botones (icon-only) mantienen `aria-label` + `title` para lectores de pantalla.

### Preserved (no se revierte de v2.1.3)
- ✅ Micro-shake sutil en `.output-box.empty-shake` (240ms × 2 oscilaciones) cuando se clickea output vacío.
- ✅ `showInlineEmpty()` NO toca `textContent` — preserva "LLENAR DATOS" intacto.
- ✅ Triángulo SVG inline (NO emoji) en coberturaBtn.
- ✅ Estados `cobertura-btn--idle` / `cobertura-btn.warn` / `cobertura-btn.off` con pulseDanger/pulseWarn.
- ✅ Output centrado (`text-align: center; justify-content: center`) — simetría con placeholder.
- ✅ Liquid Glass System (tier-1/2/3, blur-md, hairline highlight).
- ✅ SVGs idénticos al v2.1.3 (papelera + triángulo).

### Tests
- test_popup.py: 68 originales + nuevas aserciones v2.1.4 (10+)
- node --check: EXIT 0
- test_security.py: 🟢 CLEAN
- test_codereview.py: 🟢 GOOD

### Reversión explícita desde v2.1.3
- ❌ `<span class="btn-icon-label">BORRAR</span>` → ELIMINADO
- ❌ `<span class="cobertura-btn-label">ZONAS DE APAGADO</span>` → ELIMINADO
- ❌ `.btn-icon-clear { width: auto; padding: 0 12px; }` → revertido a icon-only
- ❌ `.cobertura-btn { padding: 0 12px; height: 36px; }` sin width fija → width 36px fijo
- ❌ `truncateForBox(out, 80)` en updateAddress → eliminado

---

## [2.1.0] - 2026-10-14 — Pipeline adaptable `regenerate_data.py` (CEO-editable)

### Added
- **`scripts/regenerate_data.py`**: Pipeline de 12 pasos para regenerar el HTML
  desde los PDFs de `Bases de Apagado/`. Idempotente: correr 2 veces da mismo resultado.
- **`data/zonas_apagadas_manual.json`**: CCPPs extraídos via OCR visual (37 zonas).
  **Editable por CEO** — agregar/quitar entradas manualmente y correr el script.
- **`data/manual_fixes.json`**: Correcciones manuales de typos del PDF.
  Formato: `{new_dist?, new_ccpp?, new_depto?, new_prov?}`. None = descartar.
- **Orden lógico del pipeline** (12 pasos documentados en el script):
  1. INVARIANT CHECK (HTML válido)
  2. CARGA PREVIA (cargar UBIGEO/CCPP_DATA del HTML)
  3. CARGA MANUAL_FIXES + zonas manuales (JSON)
  4. DETECCIÓN PDFs en `Bases de Apagado/`
  5. CLASIFICACIÓN pdfplumber vs imagen
  6. EXTRACCIÓN pdfplumber (forward-fill merged cells)
  7. EXTRACCIÓN OCR visual (render PNG para visión del modelo)
  8. NORMALIZACIÓN (NFC + single-space + uppercase)
  9. FUZZY MATCH UBIGEO (NFKD + diff-lib cutoff 0.78)
  10. SUMAR MANUAL_ZONES (con fuzzy dist match para typos)
  11. CONSOLIDACIÓN schema v2.0.9 (ccpp_keys + districts_with_off)
  12. INYECCIÓN inline en HTML

### Idempotencia confirmada
- Run 1: 239 CCPPs, 191 distritos ✓
- Run 2: 239 CCPPs, 191 distritos ✓ (idéntico)
- Run 3: 239 CCPPs, 191 distritos ✓

### Cómo usar (cuando llegue nueva base)
```bash
# 1. Coloca los PDFs en Bases de Apagado/ (APAGADO.pdf, APAGADO2.pdf, APAGADP1.pdf)
# 2. Edita data/zonas_apagadas_manual.json si hay nuevos CCPPs visuales
# 3. Edita data/manual_fixes.json si hay typos conocidos
# 4. Ejecuta:
python3 apps/cima-generador/scripts/regenerate_data.py
# 5. Verifica:
python3 apps/cima-generador/test_popup.py
# 6. Confirma y commitea:
git add apps/cima-generador/
git commit -m "feat(cima-generador): regenerar datos v2.x.x"
```

### Tests
- **62/62 tests PASS** (de 62, sin cambios en suite)
- Script regenera HTML determinísticamente
- Diana 🟢 CLEAN · Pablo 🟢 GOOD · node check EXIT 0

### Auditoría E2E (25/25 PASS, 0 fallos)
- 4 campos exactos matchean
- Múltiples espacios, lowercase, trim, tildes → normalize correcto
- 3 campos → distrito preventivo
- CCPP no matchea → silencio
- Datos incompletos → silencio
- Typos PDF preservados
- Deptos no en base → silencio
- Dropdown CCPP UNION (oficial + apagados)

## [2.0.9] - 2026-10-14 — Triángulo solo se colorea (sin click container) + unión CCPPs no oficiales

### Changed
- **Botón ⚠️ ya NO es cliqueable**. Sin contenedor debajo. Solo se colorea:
  - Removido `onclick="toggleCoberturaDetail()"` del botón
  - Removida función `toggleCoberturaDetail()`
  - Removido HTML `<div id="zonaApagadaBadge">` y CSS `.zona-apagada-badge`
  - Removida función `renderZonaApagadaBadge()` y referencias a `badgeArea`
  - `title` del botón ahora muestra la info via tooltip nativo (no requiere click)
- **Dropdown CCPP ahora UNION**:
  - Antes: solo CCPPs oficiales de CCPP_DATA
  - Ahora: CCPP_DATA oficial + CCPPs apagados (de `ccpp_keys`) que NO son oficiales
  - 131 CCPPs apagados antes invisibles ahora aparecen en el dropdown marcados con ⚠️
  - Cobertura completa para call center: el supervisor ve TODOS los caseríos posibles

### Fixed
- Búsqueda robusta: el supervisor ve "CHILIQUIN" en el dropdown aunque no esté en CCPP_DATA oficial
- dropdown CCPP pinta con ⚠️ los apagados no oficiales (antes ni aparecían)

### Removed
- HTML: `<div id="zonaApagadaBadge">` (contenedor amarillo)
- CSS: `.zona-apagada-badge` (incluyendo `.show`)
- JS: `toggleCoberturaDetail()`, `renderZonaApagadaBadge()`, lógica `classList.add/remove('show')`
- Toda referencia a badge en el flujo

### Audit E2E
- **25/25 escenarios PASS, 0 fallos** (call center zero-failure)
- 9 grupos cubiertos: exact-match, lowercase, multi-espacios, distrito_of, no-match, incompletos, tildes, typos PDF preservados, deptos no en base
- Verificación cross: 239/239 zonas con UBIGEO consistente
- 106 CCPPs matchean con CCPP_DATA oficial · 131 CCPPs apagados NO oficiales ahora visibles en dropdown

### Tests
- **62/62 tests PASS** (de 62 anteriores)
- Diana 🟢 CLEAN (innerHTML x5 después de eliminar badge rendering)
- Pablo 🟢 GOOD (214 comments)
- `node --check` → EXIT 0

## [2.0.8] - 2026-10-14 — Búsqueda por CCPP exacto + distrito preventivo (call center 0 fallos)

### Changed (schema de búsqueda)
- **Lookup refactorizado a nivel CCPP específico** (4 niveles):
  1. **CCPP exacto** (4 campos llenos + match `"ubigeo_d6:ccpp"`) → 🔴 alerta específica
  2. **Distrito con apagados** (3 campos llenos, sin CCPP) → 🟠 distrito_of (preventivo)
  3. **Distrito sin apagados** (3+ campos llenos) → 🟢 silencio
  4. **<3 campos** → 🟢 silencio por datos incompletos
- Schema `ZONAS_APAGADAS`:
  - Antes: `{ d6: Set<ubigeo>, d4: Set<ubigeo>, d2: Set<ubigeo>, total, fuente, fecha }`
  - Ahora: `{ ccpp_keys: Set<"d6:CCPP"> (239), districts_with_off: Set<d6> (191), total_zones: 239, total_districts: 191, fuente, fecha }`
- Inline size: 2,009 → **6,835 bytes** (CCPPs específicos pesan más que ubigeos genéricos)

### Added
- `function normalizeKey(s)` — trim + uppercase + single-space + NFC para lookups robustos
- Helper que decide CCPP exacto: `ZONAS_APAGADAS.ccpp_keys.has(d6 + ':' + normalizeKey(ccpp))`
- **Dropdown CCPPs del distrito pinta con ⚠️ los apagados** (clase `.ac-item-off`):
  - Fondo danger, font-weight 700, icono ⚠️ prefijo
  - El asesor ve qué CCPPs oficiales del distrito están apagados
- **Botón coberturaBtn estado `.warn`** (distrito_of, sin match exacto):
  - Color ámbar/dorado #f59e0b (vs rojo de `.off` exact match)
  - Animación `pulseWarn` 2.8s (más suave que pulseDanger)
  - Tooltip: "Este distrito tiene CCPPs en zona de apagado. Verifica el CCPP exacto."
- `toggleCoberturaDetail()` actualizado para manejar ambos estados (.off / .warn)

### Fixed
- **Corrección data PIURA/AYABACA** (audit durante implementación):
  - 5 CCPPs reasignados de PACAIPAMPA dist → AYABACA dist (por merged-cell ambiguo del PDF)
  - CURRILCAS agregado (CCPP que faltaba)
  - CURILCAS → CURRILCAS (typo del PDF con doble R omitida)
  - PAPELILLO → PAPELLILLO (typo del PDF con doble L preservado)
- **Eliminada generalización departamental** (v2.0.4 d4/d2):
  - Antes alertas por depto/provincia cuando había apagados generales (falso positivo)
  - Ahora solo CCPP exacto o distrito_of (sin over-alert)

### Tests
- **62 tests PASS, 0 fail** (de 61 anteriores)
- 14 escenarios E2E del call center verificados:
  - ✅ CCPP exacto: 9 escenarios (PLAYA AZUL, ESPITE, TUTUMBEROS, CHUMPE-POQUES, etc.)
  - ✅ Distrito preventivo: 3 escenarios (3 campos, distrito con apagados)
  - ✅ Silencio OK: 5 escenarios (sin match, sin apagados, datos incompletos)
  - ✅ Edge cases: AYABACA+PAPELLILLO (typo preservado), PUCARA sin CCPP (alerta distrito)
- `node --check popup.js` → EXIT 0
- Diana 🟢 CLEAN · Pablo 🟢 GOOD (212 inline comments)

### Behavior para Asesores Call Center
```
ESCENARIO TÍPICO (asesor con cliente en línea):

  Cliente: "Soy de Cajamarca, Jaén, Pucara, Playa Azul"
    → Asesor tipéa los 4 campos
    → Sistema: 🔴 MATCH EXACTO "PLAYA AZUL" (ubigeo 060808)
    → Asesor informa: "Esta zona está en apagado programado"

ESCENARIO SIN CCPP (cliente no recuerda):

  Cliente: "Soy de Cajamarca, Jaén, Pucara"  
    → Asesor tipéa los 3 campos sin CCPP
    → Sistema: 🟠 DISTRITO_OF "Este distrito tiene CCPPs en zona de apagado"
    → Asesor revisa dropdown CCPP: ve "PLAYA AZUL" marcado con ⚠️
    → Pregunta al cliente "¿es Playa Azul tu poblado?" → confirma o descarta

ESCENARIO SIN ALERTA (cliente en zona con cobertura):

  Cliente: "Soy de Lima, Lima, Miraflores, Casuarinas"
    → Asesor tipéa los 4 campos
    → Sistema: 🟢 silencio (CASUARINAS no está en apagados, distrito sin alerta)
    → Asesor copia address, sigue con la tipificación

ESCENARIO DATOS INCOMPLETOS:

  Cliente: "Soy de Junín"  (solo depto)
    → Asesor tipéa solo 1 campo
    → Sistema: 🟢 silencio (datos incompletos)
    → Asesor pregunta "¿cuál provincia y distrito?" → completa y vuelve a buscar
```

## [2.0.7] - 2026-10-14 — AUDIT EXHAUSTIVO: 3 distritos recuperados (PDF 1/3)

### Fixed (audit completo sobre los 3 PDFs)
CEO directive: "entra en bucle y verifica completo cada uno sin excepción". Auditoría visual exhaustiva de las 5 páginas de los 3 PDFs Movistar contra el dataset.

**PDFs auditados:**
- `APAGADO.pdf` (Grupo 2 oficial, 3 pages) — 206 zonas visuales
- `APAGADO2.pdf` (Grupo 3, 1 page) — 21 zonas
- `APAGADP1.pdf` (Grupo 1, 1 page) — 16 zonas

**Cross-validation Grupo 2 (PDF 1) vs mi dataset:**
- **33 gaps encontrados** entre las 206 visuales y mis 173 extraídas
- Clasificados en 3 categorías:
  - **19 districtos ya cubiertos por d6 Set** (mismo distrito, otro CCPP) — sin acción
  - **9 districtos ya cubiertos vía fuzzy match** (tildes/typos) — sin acción
  - **3 districtos NO cubiertos** — AGREGADOS al dataset
  - **2 typos reales del PDF descartados** (districtos inexistentes en UBIGEO oficial)

### Added
- **3 nuevos ubigeos al d6 Set** (recuperados del PDF 1 Grupo 2):
  - `050611` — AYACUCHO/LUCANAS/LUCANAS (UCHUYMARCA)
  - `060601` — CAJAMARCA/CUTERVO/CUTERVO (CONDORHUASI)
  - `060201` — CAJAMARCA/CAJABAMBA (typo PDF "CAJAMARCA" como dist, fuzzy lo recuperó como CAJABAMBA)
- Total: **192 → 195 d6 ubigeos** (+3 distritos)
- Total general: **239 → 242 zonas** (+3)

### Justified skips (typos reales del PDF)
- `('AYACUCHO', 'HUAMANGA', 'ANCO', 'SACHARACCAY')` — distrito "ANCO" no existe en HUAMANGA
- `('LORETO', 'MARISCAL RAMON CASTILLA', 'SANTA ROSA DE LORETO', 'SANTA ROSA')` — distrito "SANTA ROSA DE LORETO" no existe en MRC (MRC solo tiene RAMON CASTILLA, PEBAS, YAVARI, SAN PABLO) — el CCPP "SANTA ROSA" puede existir pero district es incorrecto
- Otros 6 typos ya estaban cubiertos por fuzzy (CAJAMARCA→CAJABAMBA, TOLED→TOLEDO, CHANCAYBAÑOS, OCONA→OCOÑA, MUNANI→MUÑANI, etc.)

### Tests
- 58 → **61 tests PASS** (+3 nuevos):
  - 050611 UCHUYMARCA presente
  - 060601 CONDORHUASI presente
  - 060201 CAJABAMBA dist typo recuperado presente
- `node --check popup.js` → EXIT 0
- Diana 🟢 CLEAN · Pablo 🟢 GOOD

### Naming observation
Los archivos PDF tienen nombres invertidos vs su contenido:
- `APAGADO.pdf` = Grupo **2** oficial (3 páginas, 206 zonas)
- `APAGADP1.pdf` = Grupo **1** (1 página, 16 zonas)
- `APAGADO2.pdf` = Grupo **3** (1 página, 21 zonas)

## [2.0.6] - 2026-10-14 — Badge opt-in (no auto-show, only on ⚠️ click)

### Fixed
- **Border amarillo persistente eliminado**: el `.zona-apagada-badge` ya NO aparece automáticamente al tipear una zona de apagado
- **Badge opt-in**: solo visible cuando el supervisor hace click en el botón ⚠️ (toggleCoberturaDetail)
- **CSS cleanup**: 2 declaraciones duplicadas `display: none` en `.zona-apagada-badge` reducidas a 1

### Changed
- `checkAndRenderZonaApagada()` ya no hace `classList.add('show')` automático. Solo prepara el contenido HTML del badge.
- Si el badge estaba visible y el supervisor cambia la dirección, el badge se oculta automáticamente (`classList.remove('show')`)

### UX nuevo
```
1. Tipear LORETO → ⚠️ aparece al lado del X (no border amarillo)
2. Click en ⚠️ → badge amarillo aparece debajo con detalle
3. Click again en ⚠️ → badge desaparece (toggle)
4. Cambiar dirección → badge se cierra automáticamente
```

### Tests
- 54 → **58 tests PASS** (+4 nuevos en sección "Badge opt-in v2.0.6"):
  - .zona-apagada-badge tiene 1 sola declaración display: none
  - checkAndRenderZonaApagada oculta badge (no auto-show)
  - classList.add('show') eliminado (badge opt-in)
  - .zona-apagada-badge.show sigue presente (toggle click)
- `node --check` → EXIT 0

## [2.0.5] - 2026-10-14 — Visual refactor: coberturaBtn compacto al lado del X

### Changed
- **Botón COBERTURA rediseñado** a icono compacto (36×34px) al lado del botón X clear, dentro del output-row
- **Hidden por default** (`display: none`) — aparece SOLO cuando hay alerta (`display: inline-flex` con clase `.off`)
- **Sin texto** "COBERTURA OFF / COBERTURA OK / COBERTURA" — solo el triángulo ⚠️ dentro del icono de 16px
- Pulse keyframes mantienen efecto visual de atención

### Removed
- Strip ancho que ocupaba toda la fila (ruido visual)
- Label "COBERTURA" en HTML
- Wrapper externo del botón
- Lógica `pending` / `ok` separada — solo hay 2 estados: `hidden` o `off`

### Layout del output-row ahora
```
[ output-box (flex:1) ] [ 🗑️ X (34px) ] [ ⚠️ (36px, solo si off) ]
```

### Tests
- 50 → **54 tests PASS** (+4 nuevos en sección "Visual layout v2.0.5"):
  - coberturaBtn dentro del output-row (después de X clear)
  - .cobertura-btn display:none por default (hidden)
  - .cobertura-btn.off display:inline-flex (visible con alerta)
  - coberturaLabel eliminado (icon-only)
- `node --check` → EXIT 0
- Diana security 🟢 CLEAN · Pablo code review 🟢 GOOD

## [2.0.4] - 2026-10-14 — ZONAS APAGADAS política proporcional (anti-falso positivo)

### Changed
- **Jerarquía honesta** (anti-falso positivo):
  - **d6**: 192 distritos con ≥1 CCPP apagado (granular) — **mantiene**
  - **d4**: 106 → **1** (solo LORETO prov, donde ≥60% de distritos están apagados)
  - **d2**: 23 → **1** (solo LORETO depto, donde ≥40% de distritos están apagados)
- **Comportamiento UX**: El supervisor ahora debe tipear hasta **distrito** (o LORETO específico) para alertar OFF. Tipear solo "JUNIN" ya no alerta — solo cuando llega a un distrito con CCPP confirmado.
- **Inline size**: 2,802 → **1,957 bytes** (−30%, al achicar d4/d2)

### Why this change
- v2.0.3 marcó LIMA, JUNIN, AREQUIPA como "departamento OFF" — **falso positivo**: solo Lima Provincias y distritos específicos de JUNIN tienen apagado, NO el depto completo.
- Política proporcional: solo subir de d6→d4→d2 cuando el dataset respalda la generalización.
- LORETO pasa los cutoffs: 22/55 distritos = 40% (d2), Loreto-prov 3/5 = 60% (d4). Justifica alerta amplia.

### Tests
- 50/50 PASS
- Validación E2E con 12 escenarios reales de supervisor:
  - LIMA + LIMA + MIRAFLORES → 🟢 OK (correcto, Lima Metropolitana no apagada)
  - LIMA + HUARAL + IHUARI → 🔴 distrito OFF (correcto, Lima Provincias)
  - LORETO depto solo → 🔴 departamento (único caso)
  - TUMBES + TUMBES + TUMBES → 🟢 OK (Tumbes no afectada)
  - CUSCO + LA CONVENCION + MARANURA → 🔴 distrito (vía OCR visual Grupo 3)
  - AYACUCHO + VICTOR FAJARDO + VILCANCHOS → 🔴 distrito (vía OCR visual Grupo 3)
- `node --check` → EXIT 0

## [2.0.3] - 2026-10-14 — ZONAS APAGADAS full dataset (239 zonas, OCR end-to-end)

### Changed
- **Dataset expandido 68 → 239 zonas** (+251%):
  - Grupo 1 (APAGADO.pdf + APAGADP1.pdf): pdfplumber extracción 3 pages + OCR visual 16 zonas nuevas
  - Grupo 3 (APAGADO2.pdf): OCR visual 21 zonas (PDF era imagen escaneada)
  - 2 entries descartadas (errores tipográficos en PDF: CAJAMARCA/CAJABAMBA/CAJAMARCA, LORETO/MARISCAL RAMON CASTILLA/SANTA ROSA DE LORETO — distritos inexistentes en UBIGEO oficial)
- **Fuzzy match** automático: NFKC normalization + diff-lib cutoff 0.7-0.75 corrige typos de tildes (OCA×OCOÑA, MUÑANI×MUNANI, CASTAÑEDA×CASTANEDA, etc.)
- Inline data: 1,176 → **2,802 bytes** (+138% comp ratio vs 17,043 JSON expandido)

### Added
- **Cobertura jerárquica** ahora cubre 23/25 departamentos (vs 15 antes), 106 provincias (vs 39), 192 distritos (vs 62)
- Test_popup.py: bloque "ZONAS_APAGADAS data integrity" con 17 tests E2E:
  - d6/d4/d2 size + uniqueness + code length
  - Sample lookup tests (zonas nuevas v2.0.3 + regresión v2.0.1)
- 4 nuevos departamentos incluidos: PASCO (19), UCAYALI (25), HUANUCO (10), MOQUEGUA (18)

### Fixed
- 12 typos de tildes automáticamente corregidos vía fuzzy match
- Resolución jerárquica 3 niveles (depto→prov→dist) ahora infalible para derivados de PDF Movistar

### Tests
- De 36 → **53 tests** (+17): PASS 53/53
- `node --check popup.js` → EXIT 0 (711,997 chars)
- Diana security: 🟢 CLEAN
- Pablo code review: 🟢 GOOD (194 inline comments)

## [2.0.2] - 2026-10-14 — ZONAS APAGADAS active (data inline + UI semáforo)

### Changed
- **Data schema migrated** from `by_ubigeo[]/by_dist[]/by_prov[]/by_depto[]` (3,606 chars) → `d6/d4/d2: Set<ubigeo>` (1,176 chars) — **67% size reduction** with O(1) lookup
- `checkZonaApagada()` signature simplified: `(depto, prov, dist)` — uses Sets internally
- `checkAndRenderZonaApagada()` now dual-renders: badge + semáforo button

### Added
- `getUbigeoCodesForAddress(depto, prov, dist)` — extracts 2d/4d/6d ubigeo from inline UBIGEO tree
- `renderCoberturaButton(check)` — semáforo pending(gris) → ok(verde) → off(rojo+pulse)
- `toggleCoberturaDetail()` — click handler que muestra/oculta badge con detalle
- HTML: `<button id="coberturaBtn">` debajo de output-row, encima del badge area
- CSS `.cobertura-btn` con estados pending/ok/off + `@keyframes pulseDanger` (sólo en estado off)
- `prefers-reduced-motion: reduce` desactiva pulse animation automáticamente
- WCAG: aria-label, aria-pressed, aria-expanded, focus-visible outline

### Data
- 68 zonas totales (62 distritos d6 + 39 provincias d4 + 15 departamentos d2, con solapamiento por jerarquía)
- 3 PDFs fuente: APAGADO (Grupo 1, 3 pages), APAGADO2 (Grupo 2, 1 page), APAGADP1 (Grupo 3, 1 page)
- Inline en HTML — popup sigue 100% standalone (file://, sin CORS, sin fetch)

### Tests
- De 29 → **36 tests** (+7): compact Sets d6/d4/d2, getUbigeoCodesForAddress, renderCoberturaButton, toggleCoberturaDetail, coberturaBtn element, pulseDanger keyframes
- 36/36 PASS · 0 FAIL · 0 WARN
- `node --check popup.js` → EXIT 0 (sin syntax errors)
- Diana security 2da pasada: 🟢 CLEAN (innerHTML = pre-existing patterns)
- Pablo code review: 🟢 GOOD (193 inline comments, 5 try/catch, camelCase/UPPER_CASE consistentes)

## [2.0.1] - 2026-10-14 — ZONAS APAGADAS feature ready

### Added
- **ZONAS APAGADAS detection** — feature genérico para alertar cuando una búsqueda es zona de retiro
- `checkZonaApagada(depto, prov, dist, ccpp)` — retorna `{nivel, mensaje}` o null
- `renderZonaApagadaBadge(check)` — genera HTML del badge de alerta
- `checkAndRenderZonaApagada()` — hook integrado en `updateAddress()`
- CSS `.zona-apagada-badge` con animación, gradiente amber
- UI: badge "⚠️ ZONA DE APAGADO — MOVISTAR RETIRA COBERTURA" cuando match
- Niveles: `departamento` / `provincia` / `distrito` (alerta si CUALQUIER nivel matchea)

### Security
- `escapeHtml()` helper para badge (full escape & < > " ' /)
- No PII collected
- Defensive: si `ZONAS_APAGADAS` no está definida, retorna null sin error

### Documentation
- data/zonas_apagadas.json EXCLUIDO del repo (.gitignore) por ser derivado de documentos confidenciales Movistar
- Feature funciona sin data — solo agrega UI vacía si no hay const
- Inyección de data se hace manualmente por CEO desde su fuente autorizada

### Changed
- VERSION: 2.0.0 → 2.0.1
- index.html: 740,434 → 744,783 bytes (+4,349 = feature + funciones + CSS + UI)
- 2 tests nuevos en test_popup.py (29 total, antes 27)

### Verified
- ✅ 29/29 tests integridad PASS
- ✅ 0 issues security (Diana)
- ✅ node --check popup.js
- ✅ Test funcional: lógica de detección correcta (depto/prov/dist)
- ✅ Fallback graceful sin data

---

## [2.0.0] - 2026-10-14 — STANDALONE (FIX COMPLETO)

### 🚨 Bug fix crítico
- **Búsqueda Address no carga datos en algunos contextos**
  - Causa: `fetch('data/*.json')` falla en CORS o file://
  - Fix: data inlineada como `const UBIGEO` y `const CCPP_DATA` en el script
  - Resultado: popup 100% standalone — funciona desde `file://`, sin HTTP server, sin CORS

### Fixed
- **#1 CRITICAL** — tipif textareas focus loss: re-render selectivo en `setTipif`
  - Solo se actualiza el output-box (textContent), no se destruye el textarea
  - Cursor y focus preservados al tipear
- **#2 HIGH** — DoS paste: `maxlength="120"` en 6 inputs + helper `debounce()`
- **#3 HIGH** — memory sin cap: cap tipifs a 50 con `alertMsg()`
- **#4 HIGH** — JSON sin SRI: data inline (elimina fetch completamente)
- **#5 MEDIUM** — acEscape incompleto: full escape `& < > " ' /`
- **#6 MEDIUM** — sin CSP: agregada `<meta CSP>` en `<head>`

### Changed
- HTML ahora 737KB (vs 76KB v1.0.0) — 96% del tamaño es data inline
- Sin dependencia de carpeta `data/` para runtime
- Standalone: funciona desde `file://`, `python -m http.server`, GitHub Pages, Google Sites
- `loadUbigeo()` y `loadCCPP()` ahora son no-ops (compatibilidad)

### Security
- CSP: `default-src 'self' data:; script-src 'self' 'unsafe-inline'; ...`
- Sin inline event handlers nuevos (mantiene los existentes por compatibilidad)
- GeoNames CC BY 4.0 atribución visible
- RitchieRD ODbL atribución visible

### Verificado
- ✅ `node --check popup.js` — JS válido
- ✅ HTTP 200 sirviendo 749KB vía `python3 -m http.server`
- ✅ UBIGEO inline tiene 25 deptos
- ✅ CCPP_DATA inline tiene 43,491 caseríos

---

## [1.0.0] - 2026-10-13

### Added
- D-22 → D-30 polish + 7 bugfixes frontend
- Hardening para 50+ usuarios concurrentes
- `<link rel="preload">` strategy: ubigeo eager + ccpp lazy

### Data
- `data/ubigeo_consolidated.json` — 25 deptos, 196 prov, 1892 dist (`82d0706`)
- `data/ccpp_by_distrito.json` — 43,491 caseríos GeoNames CC BY 4.0 (`88edb16`)

### Features
- Popup v1.0: header minimal + footer creditos + autocomplete custom premium (`0d96ac6`)
- CCPP by distrito from GeoNames (`da892e3`)
- Ubigeo consolidado con `_ubigeo` + `_id` (`ea28abb`)
- Popup completo WCAG a11y + ubigeo + CCPP + Maps format (`136fd52`)

### Fixed
- Address conditional joining: omitir partes vacías en output (`d40104a`)

## [0.1.0] - 2026-10-13

### Added
- Popup generador CIMA v0.1 (`7067ec0`)
- Initial commit (`6d16318`)

---

## Notas

v2.0.0 es la primera versión **production-ready** del popup. v1.0.0 quedó
como referencia histórica. La data en `data/` se mantiene como backup
para regenerar la versión inline si se actualizan los catálogos.


### Added
- D-22 → D-30 polish + 7 bugfixes frontend
- Hardening para 50+ usuarios concurrentes
- `<link rel="preload">` strategy: ubigeo eager + ccpp lazy

### Data
- `data/ubigeo_consolidated.json` — 25 deptos, 196 prov, 1892 dist (`82d0706`)
- `data/ccpp_by_distrito.json` — 43,491 caseríos GeoNames CC BY 4.0 (`88edb16`)

### Features
- Popup v1.0: header minimal + footer creditos + autocomplete custom premium (`0d96ac6`)
- CCPP by distrito from GeoNames (`da892e3`)
- Ubigeo consolidado con `_ubigeo` + `_id` (`ea28abb`)
- Popup completo WCAG a11y + ubigeo + CCPP + Maps format (`136fd52`)

### Fixed
- Address conditional joining: omitir partes vacías en output (`d40104a`)

## [0.1.0] - 2026-10-13

### Added
- Popup generador CIMA v0.1 (`7067ec0`)
- Initial commit (`6d16318`)

---

## Auditoría 2026-10-14 (importado a CIMA standalone)

CEO Alexander Salvador ejecutó análisis multi-agente sobre el repo.
**Veredicto:** Funcional 8/10, Documental 4/10.

### Issues identificados

| Sev | Count | Issues |
|-----|-------|--------|
| CRITICAL | 1 | Tipif textareas: focus loss por re-render en cada keystroke |
| HIGH | 3 | DoS paste sin maxlength · memory sin cap · JSON sin SRI |
| MEDIUM | 4 | acEscape incompleto · sin CSP · CCPP silent fail · race typing vs fetch |
| LOW | 2 | referencia data loss · window.close sin sandbox |

### Plan de remediación (no aplicado aún, requiere CEO approval)

1. **#1 (CRITICAL)**: Re-render selectivo en tipifs (no destruir textarea en `input`)
2. **#4 (HIGH)**: Inline data como `const UBIGEO = {...}` en script — elimina fetch risk
3. **#2 (HIGH)**: `maxlength="120"` en todos los inputs + debounce `acRender` 80ms
4. **#3 (HIGH)**: Cap tipifs a 50 con mensaje
5. **#5 (MEDIUM)**: Reemplazar acEscape con full escape `& < > " ' /`
6. **#6 (MEDIUM)**: Agregar CSP meta tag

### Documentación generada

- `CHANGELOG.md` — este archivo
- `VERSION` — v1.0.0
- `docs/AUDIT.md` — análisis adversarial completo
- `docs/SOURCES.md` — atribución de datos
- `docs/INTEGRATION.md` — cómo embeber en Google Sites
- `MEMORY.md` (CIMA parent) — entrada dedicada a cima-generador

---

*Cima-generador es un mini-proyecto independiente de AKRYNT, exclusivo para CIMA.*
*Importado al repo CIMA standalone el 2026-10-14.*

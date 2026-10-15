# Changelog — cima-generador

## [3.3.6] - 2026-10-15 — SELECTION PERSISTE EN SEARCH + PREVIEW MINIMAL

CEO directive 2026-10-15:
1. "al seleccion la opcion del dropdown debe quedarse en el searchinput"
2. "cuando se selecciona una opcion del dropwdown se genera un contenido
   debajo, vamos a editar el contenido que se genera, retiramos todo solo
   debe mantenerse TX OPER TO OPER -> ETIQUETA | TX DIRECTA -> ETIQUETA"

### Fixed — Selection persiste en search input

**Antes:** Click en dropdown item → `selectEtiqueta(id)` actualizaba `TX_STATE.selectedId` + mostraba preview. El search input quedaba con su valor previo (lo que el usuario había tipeado para filtrar).

**Ahora:** Click en dropdown item → input.value = e.solicitud (la etiqueta seleccionada se queda en el input). El usuario ve claramente qué eligió.

```js
function selectEtiqueta(id) {
  var e = ETIQUETAS_DATA.find(function (x) { return x.id === id; });
  if (!e) return;
  TX_STATE.selectedId = e.id;
  var input = $('etiquetaSearch');
  if (input) {
    input.value = e.solicitud;       // v3.3.6: persistir selección en input
    updateEtiquetaClearVisibility();  // mostrar clear button (X)
  }
  showEtiquetaPreview(e);
  hideEtiquetaDropdown();
}
```

**Bonus:** `closeEtiquetaPreview()` también limpia el search input + update clear button visibility.

### Fixed — Preview MINIMAL (solo 2 skill rows)

**Antes (v3.1):** Preview card con header (cat badge + ID + close button) + title + description + 2 skill cards + validation card + copy button. Demasiada información que el supervisor no usa.

**Ahora (v3.3.6):** Preview es solo 2 mini-cards:
- TX OPER TO OPER → `skill_tx2` value
- TX DIRECTA → `skill_txdirecto` value

```js
function showEtiquetaPreview(e) {
  $('etiquetaPreviewSkills').innerHTML =
    '<div class="etiqueta-skill-card">' +
      '<span class="etiqueta-skill-label">TX OPER TO OPER</span>' +
      '<span class="etiqueta-skill-row">' + escapeHtml(e.skill_tx2) + '</span>' +
    '</div>' +
    '<div class="etiqueta-skill-card">' +
      '<span class="etiqueta-skill-label">TX DIRECTA</span>' +
      '<span class="etiqueta-skill-row">' + escapeHtml(e.skill_txdirecto) + '</span>' +
    '</div>';
  preview.hidden = false;
}
```

**Nota:** El label cambió de "TX 2 PASOS" a "TX OPER TO OPER" (CEO directive).

### Removed — Elementos obsoletos del preview

| Removido | Razón |
|---|---|
| `.etiqueta-preview-header` | CEO: "retiramos todo" |
| `.etiqueta-preview-cat` (cat badge en preview) | redundante con chip |
| `.etiqueta-preview-id` (#01 · AREA) | info no usada |
| `.etiqueta-preview-close` (X button) | closeEtiquetaPreview aún existe como helper pero sin UI |
| `.etiqueta-preview-title` (solicitud) | redundante con input.value |
| `.etiqueta-preview-desc` (detalle) | info no usada |
| `.etiqueta-preview-validation` (warning card) | info no usada |
| `.etiqueta-validation-label` (warning label) | warning card removido |
| `.copy-etiqueta-btn` (copy button) | CEO: "retiramos todo" |
| `copyEtiqueta()` JS function | sin botón que la llame |

### Preserved (no cambios)

- v3.3.5 tipif stretch + dropdown navegable
- v3.3.4 dropdown items minimal (solo title)
- v3.3.3 chip click = solo selección
- v3.3.2 chips 1-row horizontal + labels abreviados
- v3.3.1 dual hook data-popup + flex-wrap nowrap
- v3.2 dropdown bounded + segmented actions
- v3.1 category system design tokens (cat-postpago/mtotal/standalone-*)

### Files modificados

- `apps/cima-generador/index.html` — HTML preview simplificado (15 líneas → 3) + showEtiquetaPreview minimal + selectEtiqueta input.value + closeEtiquetaPreview limpia input + copyEtiqueta JS function removida + CSS preview reducido (~100 líneas → ~40 líneas)
- `apps/cima-generador/VERSION` — `3.3.5` → `3.3.6`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — tests obsoletos removidos + nuevos v3.3.6 (263/263 PASS)

### Tests

- **263/263 PASS** (era 264/264; -4 obsoletos removidos, +5 nuevos v3.3.6)
- -1 .copy-etiqueta-btn CSS (removido)
- -1 .etiqueta-validation-label CSS (removido)
- -1 .etiqueta-preview-close HTML (removido)
- -1 .etiqueta-preview-close CSS (removido)
- +1 selectEtiqueta: input.value = e.solicitud
- +1 showEtiquetaPreview: labels CEO (TX OPER TO OPER + TX DIRECTA)
- +1 showEtiquetaPreview: SIN elementos obsoletos
- +1 HTML preview: estructura minimal
- +1 CSS preview: SIN reglas obsoletas

### Lección nueva

- **L-034:** Selection persistence en search inputs es UX fundamental para confirmation visual. Después de seleccionar de un dropdown, el search input DEBE mostrar el valor seleccionado — sino el usuario no tiene feedback de qué eligió.
- **L-035:** Preview cards deben mostrar SOLO la información que el usuario va a USAR. Cat badge, ID, descripción larga, validación, copy button — todos son UI noise si el supervisor solo necesita los skill codes para tipear en otro sistema. Less is more.

---

## [3.3.5] - 2026-10-15 — TIPIF STRETCH RESPONSIVE + DROPDOWN NAVEGABLE

CEO directive 2026-10-15:
1. "estirar la altura del grid TIPIFICACION el contenedor estiralo responsive"
2. "etiqueta de transferencia el inputsearch al darle click el dropdown debe
   salir hacia abajo y el contenido del dropdown reducele la font para que se
   sienta navegable y scrolleable"

### Fixed — Tipificación container stretch responsive

**Antes:** `.tipif-block` con altura auto (basada en contenido). En popup mode el contenedor era compacto, dejando poco espacio para el dropdown del search.

**Ahora:** `.tipif-block` con `min-height: 320px` + `display: flex; flex-direction: column`. Combinado con `.popup-frame { display: flex; flex-direction: column }` y `#gridTipif { flex: 1 }`, el contenedor de Tipificación **estira** para llenar la altura disponible del popup.

```css
.popup-frame {
  display: flex;
  flex-direction: column;
}
.grid-block#gridTipif {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.grid-block#gridTipif > .section {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.tipif-block {
  min-height: 320px;
  display: flex;
  flex-direction: column;
}
```

Resultado: la sección Tipificación ahora ocupa todo el viewport height disponible, dejando room natural para que el dropdown salga hacia abajo.

### Fixed — Dropdown SIEMPRE hacia abajo (sin flip)

**Antes:** `etiquetasFitToBlock` calculaba `available` (espacio debajo del input hasta el bottom del .etiquetas-block). Si `available < FLIP_THRESHOLD (100px)` → flip arriba con `is-flipped` class.

**Ahora:** SIEMPRE hacia abajo. Cap = `min(espacio en bloque, viewport-residual)`. Si el viewport es pequeño, dropdown se ajusta para no salirse de la ventana.

```js
function etiquetasFitToBlock(drop, input) {
  var availableInBlock = blockRect.bottom - inputRect.bottom - MARGIN;
  var availableInViewport = window.innerHeight - inputRect.bottom - MARGIN;
  var cap = Math.max(MIN_CAP, Math.min(MAX_CAP, availableInBlock, availableInViewport));
  drop.style.top = (inputRect.bottom + 4) + 'px';
  drop.style.maxHeight = cap + 'px';
  drop.classList.remove('is-flipped');  // SIEMPRE abajo
}
```

CSS `.is-flipped` preservado por defensa, pero JS ya no lo aplica.

### Fixed — Dropdown content reducible (navegable + scrolleable)

**Antes (v3.3.4):** `.etiqueta-item { padding: 7px 10px }` + `.etiqueta-item-title { font-size: 11px; line-height: 1.3 }` → item alto ~28px → solo ~7 items visibles por scroll.

**Ahora:** `.etiqueta-item { padding: 5px 10px }` + `.etiqueta-item-title { font-size: 10px; line-height: 1.25 }` → item alto ~22px → **~14 items visibles por scroll** (2x más densidad).

```css
.etiqueta-item {
  padding: 5px 10px !important;
}
.etiqueta-item-title {
  font-size: 10px;
  font-weight: 600;
  line-height: 1.25;
}
```

### Files modificados

- `apps/cima-generador/index.html` — CSS popup-frame + gridTipif + tipif-block flex (~25 líneas) + dropdown item padding/font-size (~4 líneas) + JS etiquetasFitToBlock siempre-down (~10 líneas)
- `apps/cima-generador/VERSION` — `3.3.4` → `3.3.5`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +10 tests (254 → 264)

### Tests

- **264/264 PASS** (era 254/254; +10 nuevos v3.3.5)
- +1 .tipif-block min-height 320px
- +1 .tipif-block display flex column
- +1 .popup-frame display flex column
- +1 #gridTipif flex: 1
- +1 .etiqueta-item-title font-size 10px
- +1 .etiqueta-item padding 5px 10px
- +1 etiquetasFitToBlock usa window.innerHeight
- +1 etiquetasFitToBlock SIN flip logic
- +1 etiquetasFitToBlock remove is-flipped siempre
- +1 is-flipped CSS preservado (defensa)

### Lección nueva

- **L-032:** Stretch containers en popups requieren flex layout en TODA la cadena (popup-frame → grid-block → section → content). Solo `min-height` en el bloque final no funciona si los padres no distribuyen el espacio. Diseño: popup-frame flex column → gridTipif flex: 1 → tipif-block min-height.
- **L-033:** Dropdown font-size + padding density afecta UX de navegación. 11px+7px padding da ~7 items/scroll (poco navegable). 10px+5px padding da ~14 items/scroll (mejor navegabilidad sin sacrificar legibilidad).

---

## [3.3.4] - 2026-10-15 — DROPDOWN MINIMAL (solo title)

CEO directive 2026-10-15: "Al darle click al searchinput el dropdown esta mal diseñado. unicamente debe decir el title para que es nada mas. por ejemplo:

PLAN MENOR
BAJA TOTAL O MIGRACION A PREPAGO".

### Fixed — Dropdown items minimal (title only)

**Antes (v3.1):** Cada item del dropdown tenía 3 filas (cat badge + solicitud / detalle 2-line clamp / skill chip + area pill). Demasiada información para navegación rápida.

**Ahora (v3.3.4):** Cada item es solo el title (solicitud). El detalle, skill codes, area, y validación viven en el preview card al seleccionar.

```js
return '<div class="ac-item etiqueta-item" data-id="' + e.id + '" onclick="selectEtiqueta(' + e.id + ')" role="option">' +
  '<span class="etiqueta-item-title">' + safeSol + '</span>' +
'</div>';
```

### Cleaned — CSS obsoleto removido

Reglas CSS eliminadas (ya no se usan en dropdown):
- `.etiqueta-item-row1` + `.etiqueta-item-row1 strong`
- `.etiqueta-item-row2`
- `.etiqueta-item-row3`
- `.etiqueta-item-skill`
- `.etiqueta-item-area`
- `.etiqueta-item-cat` + `.etiqueta-item-cat::before` + `.cat-postpago/mtotal/standalone`
- `.etiqueta-item.is-selected::before` + per-cat dot indicator
- `.etiqueta-item::before` (selected dot)

**Nueva regla minimal:**
```css
.etiqueta-item {
  padding: 7px 10px;
  transition: background var(--motion-fast);
}
.etiqueta-item:hover {
  background: color-mix(in srgb, var(--accent) 92%, transparent);
  color: var(--on-accent);
}
.etiqueta-item-title {
  display: block;
  font-size: 11px;
  color: var(--text);
  font-weight: 600;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

### Preserved

- v3.3.3 chip click = solo selección
- v3.3.2 chips 1-row horizontal + labels abreviados
- v3.3.1 dual hook data-popup + flex-wrap nowrap + min-width 56px
- v3.2 dropdown bounded + segmented actions
- v3.1 preview card rediseñado (cat badge themed + skills mini-cards + close button + validation themed)
- v3.1 category system design tokens (cat-postpago/mtotal/standalone-*)

### Files modificados

- `apps/cima-generador/index.html` — `renderEtiquetasDropdown` simplificado (~15 líneas menos) + CSS dropdown items limpio (~85 líneas removidas, ~12 nuevas) + cleanup comment `.etiqueta-preview-cat`
- `apps/cima-generador/VERSION` — `3.3.3` → `3.3.4`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — tests obsoletos removidos + nuevos v3.3.4 (254/254 PASS)

### Tests

- **254/254 PASS** (era 255/255; -6 obsoletos removidos, +5 nuevos v3.3.4)
- -4 tests v3.1 obsoletos (etiqueta-item-row3/skill/area/is-selected en dropdown)
- -3 tests v3.1 CSS obsoletos (.etiqueta-item-row3/skill/area)
- +1 renderEtiquetasDropdown usa .etiqueta-item-title
- +1 render SIN clases obsoletas
- +1 CSS SIN reglas obsoletas
- +1 CSS .etiqueta-item-title presente
- +1 comment CEO directive presente

### Lección nueva

- **L-031:** Diseño maximalista (3 filas con badge + detalle + meta pills) es anti-pattern para dropdowns de navegación. El dropdown debe ser minimal — el detail vive en preview/expanded view. CEO directive "solo el title" es la regla de oro para dropdown items.

---

## [3.3.3] - 2026-10-15 — CHIP CLICK = SOLO SELECCIÓN (NO ABRE DROPDOWN)

CEO directive 2026-10-15: "Al darle click a los button post/pree, MT, STANDALONE, se abre automaticamente el dropdown mal diseñado — incorrecto los botones son de seleccion, el dropdown solo se debe abrir mientras tengamos una seleccion y le demos click al searchinput".

### Fixed — Chip click ya NO abre dropdown

**Bug:** `filterEtiquetasByCat(cat)` después de actualizar estado + UI llamaba `renderEtiquetasDropdown(input.value)` que **re-abría el dropdown** después de que el document mousedown handler lo había cerrado.

**Flujo del bug:**
1. User click chip → mousedown fires
2. document mousedown handler → chip está outside search/dropdown → `hideEtiquetaDropdown()`
3. onclick fires → `filterEtiquetasByCat('MTOTAL')` → actualiza state + UI + **`renderEtiquetasDropdown()` re-abre**

**Fix:** Removida la llamada a `renderEtiquetasDropdown` en `filterEtiquetasByCat`. Ahora chip click = SOLO selección (state + UI). El dropdown NO se re-abre.

```js
function filterEtiquetasByCat(cat) {
  TX_STATE.cat = cat;
  // ... UI updates ...
  // v3.3.3 FIX: NO llamar renderEtiquetasDropdown aquí. Era el bug.
  // Si el dropdown estaba abierto, ya se cerró por document mousedown handler.
}
```

### Added — Guard defensivo en renderEtiquetasDropdown

Si NO hay categoría seleccionada (`!TX_STATE.cat`), el dropdown NO abre. Triple defensive:

```js
function renderEtiquetasDropdown(query) {
  if (!TX_STATE.cat) {
    drop.hidden = true;
    return;
  }
  // ... resto del render ...
}
```

Esto previene que cualquier caller (futuro bug, script injection, etc.) pueda abrir el dropdown sin selección válida.

### Flujo CEO explícito (after fix)

1. ✅ Click chip (POST/PRE | MT | STANDALONE) → cambia active state, **NO dropdown**
2. ✅ Click search input (con selección activa) → abre dropdown con items de la categoría
3. ✅ Click en dropdown item → selecciona etiqueta + muestra preview + cierra dropdown
4. ✅ Click en clear button (X del search) → limpia query, dropdown queda cerrado
5. ✅ Click en close button (X del preview) → limpia selección, dropdown sigue cerrado

### Files modificados

- `apps/cima-generador/index.html` — `filterEtiquetasByCat` removida llamada a render + comment explicativo (~10 líneas) + `renderEtiquetasDropdown` guard defensivo (~5 líneas)
- `apps/cima-generador/VERSION` — `3.3.2` → `3.3.3`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +3 tests (252 → 255)

### Tests

- **255/255 PASS** (era 252/252; +3 nuevos v3.3.3)
- +1 filterEtiquetasByCat NO llama renderEtiquetasDropdown (chip es selección pura)
- +1 renderEtiquetasDropdown guard !TX_STATE.cat
- +1 comment explicativo CEO directive presente

### Lección nueva

- **L-030:** Chips de selección deben actualizar state + UI solamente. Llamar funciones que abren dropdowns desde chip handlers crea bugs de cross-action (chip click reabre dropdown que acababa de cerrarse). Patrón: action handlers NO trigger otros UIs unless explícitamente requested.

---

## [3.3.2] - 2026-10-15 — CHIPS 1 SOLA ROW HORIZONTAL + LABELS ABREVIADOS

CEO directive 2026-10-15 (con screenshot): "los botones POSTPAGO MOVISTAR TOTAL Y STANDALONE deben estar en 1 solo row. puedes abreviarlo POST/PRE | MT | STANDALONE".

**REVERSIÓN de v3.3:** había interpretado mal la directiva previa como "1 opción por row stacked vertical". CEO clarifica: es 1 SOLA ROW horizontal con labels abreviados. v3.3 popup-mode 1-col override se elimina.

### Fixed — Chips SIEMPRE 3-col horizontal (1 sola row)

Antes (v3.3.1): `body[data-popup="true"] .etiqueta-categories { grid-template-columns: 1fr !important; }` → chips stacked vertical (1 per row).

Ahora: default `.etiqueta-categories { grid-template-columns: repeat(3, 1fr); }` → 3 chips en 1 sola row horizontal. Mobile fallback solo en viewports EXTREMOS (<320px) — irrelevante para popup mode.

```css
.etiqueta-categories {
  display: grid;
  grid-template-columns: repeat(3, 1fr);  /* 3 chips en 1 row */
  gap: 5px;
  ...
}
@media (max-width: 320px) {
  .etiqueta-categories { grid-template-columns: 1fr; }
}
```

### Added — Labels abreviados (CEO directive)

| Visible label | data-cat (JS logic) | title attribute (tooltip) |
|---|---|---|
| **POST/PRE** | `POSTPAGO` | "POSTPAGO/PREPAGO" |
| **MT** | `MTOTAL` | "MOVISTAR TOTAL" |
| **STANDALONE** | `STANDALONE` | "STANDALONE (servicio fijo)" |

**Importante:** `data-cat` preserva los nombres lógicos (POSTPAGO/MTOTAL/STANDALONE) — la lógica JS de `filterEtiquetasByCat`, `TX_STATE.cat`, `ETIQUETAS_DATA[i].cat` no cambia. Solo cambia el texto visible + title attributes para accesibilidad.

### Preserved

- v3.3.1 dual hook `data-popup="true"` en `<html>` + `<body>` (defense-in-depth, harmless ahora)
- v3.3 button→dropdown guard (`copySpeech` defensive, document mousedown skip)
- v3.3.1 `flex-wrap: nowrap` + `min-width: 56px` en action buttons
- v3.2 dropdown bounded a `.etiquetas-block` + segmented actions
- v3.1 category system design tokens

### Files modificados

- `apps/cima-generador/index.html` — CSS chips revertido a 3-col + labels abreviados en HTML (~5 líneas) + comment actualizado
- `apps/cima-generador/VERSION` — `3.3.1` → `3.3.2`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — tests obsoletos removidos + nuevos tests v3.3.2 (252/252 PASS)

### Tests

- **252/252 PASS** (era 245/245; +7 nuevos v3.3.2)
- Tests obsoletos v3.3 popup-mode-1-col REMOVIDOS (-5)
- Tests preservados v3.3 button guards + v3.3.1 dual hook + flex-wrap + min-width
- +1 chips grid-template-columns repeat(3, 1fr)
- +1 mobile fallback solo @ <=320px
- +3 labels abreviados (POST/PRE, MT, STANDALONE)
- +3 data-cat preservados (POSTPAGO, MTOTAL, STANDALONE)
- +3 title attributes para tooltips completos
- +1 labels viejos removidos (sin POSTPAGO/PREPAGO ni MOVISTAR TOTAL en chips)

### Lección nueva

- **L-029:** "1 opción por row" es ambiguo entre (a) 1 opción por FILA (stacked) y (b) 1 ROW de opciones (horizontal). Cuando el CEO clarifica con screenshot, alinear inmediatamente. Asumir lo incorrecto genera trabajo revertido.

---

## [3.3.1] - 2026-10-15 — BULLETPROOF POPUP MODE (CEO REPORTÓ "SIN CAMBIOS VISUALES")

CEO directive: "no estoy viendo ni la mas minima mejora. mata el servidor y vuelve a levantarlo es HTML". Después de matar/restart server, CEO sigue sin ver cambios. Análisis: la detección de popup mode podía no estar aplicándose por specificity issues o single hook en body.

### Fixed — Bulletproof popup mode detection

**Causa raíz sospechada:** El `data-popup="true"` solo se seteaba en `document.body`. Si por alguna razón ese hook fallaba (DOM timing, browser behavior), la CSS `body[data-popup="true"]` no aplicaba y CEO no veía cambios.

**Fix:**
1. **JS dual hook:** setear `data-popup="true"` en `<html>` (documentElement) Y en `<body>`. Doble seguridad.
2. **CSS triple selector:** `html[data-popup="true"]`, `body[data-popup="true"]`, y `.popup-mode` class — cualquiera que matchee aplica.
3. **`!important` en popup mode selectors** — gana cualquier specificity battle.
4. **Console.log debug** — visibility en DevTools para verificar que popup mode está activo.

```js
document.documentElement.setAttribute('data-popup', 'true');
document.body.setAttribute('data-popup', 'true');
console.log('[CIMA popup-mode] data-popup=true en html y body');
```

```css
html[data-popup="true"] .etiqueta-categories,
body[data-popup="true"] .etiqueta-categories,
.popup-mode .etiqueta-categories,
html[data-popup="true"] .etiqueta-preview-skills,
body[data-popup="true"] .etiqueta-preview-skills,
.popup-mode .etiqueta-preview-skills {
  grid-template-columns: 1fr !important;
}
```

### Fixed — Action buttons NUNCA wrap

CEO observación: "cuando reduzco el ancho del popup se vuelve a poner en 1 row cada opcion". Aunque mi CSS ya tenía `display: flex`, falta explicitar `flex-wrap: nowrap` para defense-in-depth.

**Fix:**
1. `flex-wrap: nowrap` explícito en `.action-buttons-row`
2. `min-width: 56px` en `.action-btn` — previene shrink-to-invisible en viewports extremos

```css
.action-buttons-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  ...
}
.action-btn {
  flex: 1 1 0;
  min-width: 56px;
  ...
}
```

### Files modificados

- `apps/cima-generador/index.html` — JS dual hook (~3 líneas) + CSS triple selector con !important (~15 líneas) + flex-wrap nowrap + min-width 56px (~3 líneas)
- `apps/cima-generador/VERSION` — `3.3` → `3.3.1`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +5 tests (240 → 245)

### Tests

- **245/245 PASS** (era 240/240; +5 nuevos para v3.3.1)
- +1 data-popup set en html Y body (dual hook)
- +1 popup CSS triple selector (html + body + .popup-mode)
- +1 popup CSS !important
- +1 action-buttons-row flex-wrap nowrap
- +1 action-btn min-width 56px

### Lección nueva

- **L-028:** Cuando un selector crítico (como `[data-popup="true"]`) se usa para diferenciar modos de diseño, setear el atributo en MÚLTIPLES elementos raíz (`<html>` + `<body>`) y usar múltiples selectores CSS con `!important`. La duplicación defensiva es preferible a un solo hook que puede fallar silenciosamente.

---

## [3.3] - 2026-10-15 — POPUP MODE 1-COL + BUTTON→DROPDOWN GUARD

CEO directive 2026-10-15 (con screenshot): "estas tratando de modificar para un modo desktop normal. estamos sobre el diseño de responsive popup y en modo popup se mantiene en 1 opción por row. Recuerda que nos centramos en el diseño popup siempre".

Además: "cuando le doy click a uno de los buttons se abre automáticamente el dropdown y es incorrecto, el flujo es click al button y luego click al inputsearch para que se muestre el dropdown".

### Fixed — Popup Mode: chips SIEMPRE 1-col

Antes: chips en grid 3-col por default, fallback a 1-col solo en `@media (max-width: 480px)`. CEO diseñando en popup mode (520px) veía chips en 3-col → no era el layout deseado.

Ahora: **popup mode es el target primario de diseño**. En `body[data-popup="true"]` (signal que JS pone en `?popup=1` o cuando hay `window.opener`), las chips y skills SIEMPRE van 1 opción por row:

```css
body[data-popup="true"] .etiqueta-categories,
body[data-popup="true"] .etiqueta-preview-skills {
  grid-template-columns: 1fr;
}
```

Mobile fallback ampliado a `max-width: 640px` (cubre popup mode también).

### Fixed — Action Button Click NO triggerea dropdown

Antes: el click en action button (RESET/APN/CASO) ejecutaba `copySpeech()` que solo copiaba el texto. PERO la propagación del mousedown + cualquier focus chain podía generar la percepción de "button abre dropdown". CEO explícitamente pidió el flujo:
1. Click button → NO dropdown
2. Click search input → dropdown abre

Ahora: `copySpeech()` tiene **guards defensivos múltiples**:
1. `document.activeElement.blur()` — elimina focus chain que pueda reabrir dropdown
3. `hideEtiquetaDropdown()` explícito — si dropdown estaba abierto, lo cierra (no abre nunca)
4. `return false` — detiene propagación a cualquier listener externo

Además, `setupEtiquetas` document mousedown handler ahora skip explícito clicks dentro de `.action-buttons-row` (belt-and-suspenders contra race conditions con mousedown bubbling):

```js
if (e.target.closest('.action-buttons-row')) return;
hideEtiquetaDropdown();
```

### Files modificados

- `apps/cima-generador/index.html` — CSS popup-mode selectors (~10 líneas) + media query ampliada + copySpeech defensive guards (~12 líneas JS) + document mousedown skip (.action-buttons-row) (~3 líneas)
- `apps/cima-generador/VERSION` — `3.2` → `3.3`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +6 tests (234 → 240)

### Tests

- **240/240 PASS** (era 234/234 en v3.2; +6 nuevos para v3.3)
- +1 popup mode chips 1-col via body[data-popup] selector
- +1 popup mode skills 1-col via body[data-popup] selector
- +1 mobile fallback max-width 640px
- +1 copySpeech defensive: blur() active element
- +1 copySpeech: hideEtiquetaDropdown() invocado explícitamente
- +1 document mousedown: skip .action-buttons-row

### Lecciones nuevas

- **L-026:** "Popup mode" debe ser un PRIMER CLASE en CSS, no un side-effect del screen width. Usar `body[data-popup="true"]` permite target comportamiento popup-específico sin asumir dimensiones de viewport.
- **L-027:** Action buttons deben tener `return false` + `blur()` + `hide*()` triple-guard cuando hay dropdowns floating. UX flow estricto requiere defense-in-depth contra cualquier focus chain que pueda triggerear open.

### Pendiente

- ⏳ Push a GitHub Pages (CEO autoriza)

---

## [3.2] - 2026-10-15 — DROPDOWN BOUNDS + SEGMENTED ACTIONS (CRIT-009)

Fix crítico reportado por CEO: "los 3 botones no estaban en 1 row bien optimizadas" + "el dropdown sobrepasaba los límites del grid contenedor de etiquetas en altura y anchura". Cirugía arquitectónica, no parches.

### Fixed — 3 Action Buttons: Segmented Control (iOS 2026)

**Antes:** 3 buttons independientes con glass bg individual, grid 1fr 1fr 1fr, gap 6px, min-height 44px. Parecían 3 islas separadas.

**Ahora:** Single glass container con dividers internos (iOS segment control aesthetic):

| Aspecto | Antes | Ahora |
|---|---|---|
| Container display | grid (1fr 1fr 1fr) | **flex (stretch)** |
| Container gap | 6px | **0** (dividers via border-left) |
| Container padding | none | **3px** (buttons no tocan border) |
| Container border-radius | none (per-btn) | **pill** (segmented) |
| Per-button bg | glass-2-bg + blur-md | **transparent** (segmented) |
| Per-button border | 1px glass-border | **none** |
| Per-button min-height | 44px | **34px** (compact popup) |
| Per-button flex | implicit grid | **1 1 0** (equal distribution) |
| Hover state | translateY(-1px) + box-shadow lift | **bg accent-tint + hairline highlight** |
| Press state | scale(0.96) | **scale(0.97)** |
| Icon size | 22px | **18px** |
| Label size | 10.5px | **9.5px** |
| Dividers | none | **border-left 1px glass-border** en cada sibling |
| Feedback overlay border | 1px green | **none** (borderless, more premium) |

**Visual:** Single unified row, 3 actions como 1 segment, dividers sutiles entre acciones. iOS Settings-style.

### Fixed — Dropdown Bounds (.etiquetas-block contained)

**Antes:** `renderEtiquetasDropdownPosition` → `acFitToGrid` → buscaba `.grid-block` (no existe en Etiquetas) → fallback a `window.innerHeight` → max-height cap 220px → overflow visual cuando el bloque era mayor a 220px. Width nunca se seteaba → dropdown se expandía al content width natural.

**Ahora:** Nueva función `etiquetasFitToBlock(drop, input)` que:

1. **Bounds:** `input.closest('.etiquetas-block')` (NO viewport, NO .grid-block)
2. **Width:** `inputRect.width + 'px'` — match exacto del search (alineado horizontalmente)
3. **Height:** `blockRect.bottom - inputRect.bottom - MARGIN` (espacio restante del bloque)
4. **Cap:** MIN 120px, MAX 280px
5. **Flip:** Si `available < FLIP_THRESHOLD (100px)` → flip arriba dentro del bloque
6. **Fallback:** Si no hay `.etiquetas-block` ancestor → `acFitToGrid` (defensivo)

**Portal pattern preservado:** position:fixed + document.body appendChild para escapar `isolation:isolate` del `.section` padre (CRIT-009 stack safety intacto).

**Visual:** Dropdown queda contenido horizontalmente al ancho del search input (alineado exacto) y verticalmente al bloque de etiquetas. Ya no desborda en altura ni anchura.

### Files modificados

- `apps/cima-generador/index.html` — `etiquetasFitToBlock` function (~50 líneas JS) + CSS action-buttons-row rediseñado (~80 líneas CSS) + comments actualizados
- `apps/cima-generador/VERSION` — `3.1` → `3.2`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +10 tests (224 → 234)

### Tests

- **234/234 PASS** (era 224/234; +10 nuevos para v3.2)
- +1 etiquetasFitToBlock function exists
- +1 llamada desde renderEtiquetasDropdownPosition
- +1 bounds lookup con closest('.etiquetas-block')
- +1 width alignment a inputRect.width
- +1 action-buttons-row display flex
- +1 segment dividers via + .action-btn border-left
- +1 button bg transparent
- +1 sin border individual
- +1 height 34px compact
- +1 flex 1 1 0 equal distribution

### Lecciones nuevas

- **L-024:** Un solo grid con `closest()` lookup puede no cubrir todos los dropdowns — cada bloque especializado merece su propio helper de bounds. Generalizar prematuramente es anti-pattern.
- **L-025:** `position:fixed` + portal pattern resuelve stacking context pero NO visual containment. Para containment visual se necesitan bounds explícitos calculados contra el contenedor lógico del dropdown.

### Pendiente

- ⏳ Push a GitHub Pages (CEO autoriza)

---

## [3.1] - 2026-10-15 — ETIQUETAS REDESIGN (CRIT-009 Avanzado+)

Rediseño end-to-end del bloque Etiquetas de Transferencia. Category System como design tokens de primera clase, hierarchy visual premium, motion design iOS 2026.

### Category System as Design Tokens

Cada categoría (POSTPAGO/MTOTAL/STANDALONE) expone 4 tokens dedicados en `:root` + overrides para `prefers-color-scheme: light`:

| Token | Dark | Light | Uso |
|---|---|---|---|
| `--cat-{name}-500` | `#3b82f6` / `#8b5cf6` / `#10b981` | `#2563eb` / `#7c3aed` / `#059669` | Base saturada (UI activa, dot indicator) |
| `--cat-{name}-400` | `#93c5fd` / `#c4b5fd` / `#6ee7b7` | `#1d4ed8` / `#6d28d9` / `#047857` | Texto de badge (contraste WCAG) |
| `--cat-{name}-soft` | mix 18% / glass-3 | mix 10% / surface | Fondo glass del badge |
| `--cat-{name}-glow` | mix 38% / transparent | mix 26% / transparent | Shadow de elevación + ring focus |

**Validación amber tokenizada** (`--warning-amber-500/400/soft`) — reemplaza hardcoded `#f59e0b` en CSS.

**Skills accent tokenizado** (`--skill-chip-bg/border/text`) — reusa `--accent-2` pero ahora dedicado.

### Block-level theme propagation

`.etiquetas-block[data-cat="POSTPAGO|MTOTAL|STANDALONE"]` cambia `data-cat` cuando el usuario cambia de chip. Todos los hijos heredan el tema vía CSS attribute selectors — un único punto de control, sin clases separadas.

### Search input rediseñado

- **SVG icon 🔍** a la izquierda (era placeholder)
- **Clear button X** (SVG, no emoji) que aparece solo cuando hay valor (`has-value` class)
- **Focus ring themed**: cambia al color de la categoría activa (no más accent genérico)
- **Placeholder italic** preservado

### Chips rediseñados (3-col, no 4-col)

- **Grid 3-col** (era 4-col con 1 vacía)
- **SVG icon por categoría**: phone (POSTPAGO), layers (MTOTAL), home (STANDALONE)
- **Container con bg glass** alrededor (look "pill segment control" iOS)
- **Active state themed**: dot indicator con color de categoría + glow
- **Focus-visible ring**: 2px accent (accesibilidad teclado)
- **Hover state**: bg glass-2 + icon opacity 0.6→1

### Dropdown items rediseñados (3 filas)

- **Row 1**: cat badge (con dot themed) + solicitud (truncada con ellipsis)
- **Row 2**: detalle (2-line clamp en lugar de truncado a 80 chars + ellipsis)
- **Row 3 (NUEVO)**: mini skill chip (mono font, color accent-2) + area pill (texto dim)
- **Selected indicator**: dot vertical a la izquierda del item activo (themed por categoría)
- **Hover**: bg accent + on-accent text (era accent genérico)

### Preview card rediseñado (premium card)

- **Container glass tier-2** con shadow-tier-2 (era tier-1)
- **Header reorganizado**: cat badge + ID + close button (X SVG)
- **Title**: 12px bold (era 11.5px)
- **Description**: 10.5px (era 10px) — más legible
- **Skills como 2 mini-cards** (grid 1fr 1fr): cada card tiene label + skill code
  - Hover: border + bg intensify
  - Stack vertical en mobile
- **Validation card**: icon ⚠️ ahora es CSS mask (warning-amber-500 themed), no emoji
- **Animation**: `etiquetaPreviewIn` (opacity + translateY + scale) con spring motion
- **Hairline highlight superior** (iOS refraction)

### Counter live

- `<span id="etiquetasCount" aria-live="polite">` muestra conteo de etiquetas en la categoría activa (6/6/5)
- Cambia dinámicamente al filtrar

### Mobile fallback

- `@media (max-width: 480px)`: chips colapsan a grid 1-col, skills a 1-col

### Accessibility (WCAG)

- `aria-pressed` en chips (state screen readers)
- `aria-live="polite"` en count
- `role="option"` en dropdown items
- `aria-label` en clear + close buttons
- `aria-controls` en search input
- Focus-visible rings themed (no generic outline)
- SVG icons con `aria-hidden="true"` (decorativos)
- Keyboard: input focus + click-outside para dropdown (pre-existing)

### Tests

- **224/224 PASS** (era 166/166 en v3.0; +58 nuevos para v3.1)
- +12 design tokens verificados
- +3 data-cat CSS selectors
- +3 title elements (icon, count, aria-live)
- +4 search elements (icon, clear button, wrap)
- +3 chip aria-pressed verifications
- +4 dropdown item new classes
- +3 preview new elements (close, skill-card, onclick)
- +4 JS handlers nuevos (clearEtiquetaSearch, closeEtiquetaPreview, updateEtiquetaClearVisibility, renderEtiquetasDropdownPosition)
- +10 CSS reglas nuevas
- +1 keyframes animation
- +1 cat-themed glow check
- +1 warning token usage
- +1 mobile fallback check

### Iteraciones internas

| Sub-ver | Estado |
|---|---|
| v3.1 | Category tokens + block data-cat + search icon/clear + chips 3-col + dropdown 3-row + preview premium + mobile fallback |

### Files modificados

- `apps/cima-generador/index.html` — ~+556/-80 líneas CSS/HTML/JS
- `apps/cima-generador/VERSION` — `3.0` → `3.1`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +58 tests (166 → 224)
- `MEMORY.md` — actualización arquitectural v3.1

### Lecciones nuevas

- **L-021:** Block-level `data-cat` + CSS attribute selectors es más mantenible que clases separadas por variante. Un solo switch en JS, todos los hijos heredan.
- **L-022:** Hardcoded color values (#f59e0b, #3b82f6) son deuda inmediata. Tokenizar desde el día 1 evita refactor masivo después.
- **L-023:** `text-overflow: ellipsis` es solo el safety net. Mejor progressive disclosure (line-clamp, expandable sections) cuando el contenido es semánticamente importante.

### Pendiente

- ⏳ Push a GitHub Pages (CEO autoriza; subtree split + force-push pendiente)
- ⏳ Bundle LICENSE en deploy standalone si CEO decide distribución parcial

---

## [3.0] - 2026-10-15 — TIPIFICACIÓN TX (Transferencia) COMPLETA

Feature mayor: el Grid 2 (TIPIFICACIÓN) deja de ser placeholder "EN PROCESO" y se transforma en herramienta TX funcional con dos secciones operativas.

### Added — Quick Speeches (fila 1)

Tres botones de copy-to-clipboard para speeches recurrentes del agente técnico:

| Action | Speech (copy limpio, sin ✓) |
|---|---|
| 🔄 RESET | "Se realizó el proceso de reset con el cliente" |
| 📶 APN | "Se realizó la configuración de APN con el cliente" |
| ↗ CASO | "Se abrió el caso con el área correspondiente" |

- **Sin prefix `✓`** (CEO directive: el output es DECLARATIVO, no narrativo)
- **Artículos + acentos preservados**: "el proceso", "la configuración", "AVERÍA"
- **Premium UI (v3.0.2)**: SVG icons Lucide-style + gradient glass per action (azul · verde · ámbar) + sub-label
- **Compact v3.0.3**: 44px alto (Apple HIG sweet spot), 22px icon, layout horizontal pill

### Added — Smart Etiquetas Search (fila 2)

Búsqueda + dropdown de etiquetas de Transferencia categorizadas:

| Categoría | Etiquetas |
|---|---|
| POSTPAGO (cat: `POSTPAGO`) | 6 registros |
| MOVISTAR TOTAL (cat: `MTOTAL`) | 6 registros |
| STANDALONE | 5 registros |
| **Total** | **17 registros** |

- **Fuente:** `Recursos CIMA/Etiquetas de Transferencia/Etiquetas de Transferencia.xlsx` (en .gitignore)
- **Categorización derivada del XLSX columna `AREA`:**
  - `MOVIL POSTPAGO // PREPAGO` → `POSTPAGO`
  - `MOVISTAR TOTAL` → `MTOTAL`
  - `SERVICIO FIJA STAND ALONE` → `STANDALONE`
- **UX**: Click chip → abre dropdown filtrado → click item → preview + copy etiqueta completa
- **Smart search**: input libre filtra por texto del label

### Fixed (4 bugs pre-commit)

| # | Bug | Sev | Fix |
|---|---|---|---|
| 1 | Iconos desaparecían al copiar (v3.0.2) — `textContent` save/restore destruía HTML | UX rota | `copyText` usa `dataset.origHTML` + `innerHTML` save/restore (L-017) |
| 2 | Stray `}` syntax error al agregar `hideEtiquetaDropdown()` | JS completo roto | Eliminado. `node --check` cazó el error antes de commit (L-007 base) |
| 3 | Skill codes largos overflow horizontal (`SK_CV_RETENCIONES_CONVERGENTE_TX`) | Visual — códigos cortados | `word-break: break-all` + `overflow-wrap: anywhere` en `.etiqueta-skill-row` |
| 4 | Inline styles quedaban en dropdown post-close | State accumulation en DOM | Nueva `hideEtiquetaDropdown()` limpia styles + restaura al parent original (sale del `<body>`) |

### Data arquitectura (0 fetch)

- `SPEECH_TEMPLATES` inline desde XLSX SPEECH RECURRENTES + PLANTILLAS PLAZOS
- `ETIQUETAS_DATA` inline desde XLSX Etiquetas de Transferencia
- **0 fetch externos en runtime** (cumple ADR-007 standalone pattern)

### Tests

- **166/166 PASS** (era 120/120 en v2.1.8; +46 nuevos para TIPIFICACIÓN TX)
- `test_popup.py` extendido con:
  - TIPIF v3.0 CSS reglas (`.action-btn-feedback`, `.etiqueta-categories`, `.copy-etiqueta-btn`, `.etiqueta-validation-label`)
  - Elementos DOM v3.0 (SPEECH_TEMPLATES, ETIQUETAS_DATA, dropdown, copy buttons)
  - SVG icons validation (×3 actions)
  - SPEECH_TEMPLATES sin `✓` prefix
  - Puntuación correcta con artículos en español
  - ETIQUETAS_DATA categorías (POSTPAGO/MTOTAL/STANDALONE)
  - "v2.x EN PROCESO" placeholder DOM reemplazado por v3.0 TIPIF TX

### Iteraciones internas

| Sub-ver | Estado |
|---|---|
| v3.0 | Vertical, 64px alto, iconos emoji 🔄📡📋 |
| v3.0.1 | Sin ✓ prefix, artículos en español, opción 'TODAS' removida |
| v3.0.2 | Premium: SVG icons Lucide-style, gradient glass per action, sub-label |
| v3.0.3 | Compact: 44px alto (Apple HIG), 22px icon, layout horizontal pill |

### Lecciones nuevas

- **L-017:** `copyText` usa `innerHTML` save/restore, NO `textContent` (preserva SVG/structure)
- **L-018:** Estética iOS 2026: 44px touch target es el sweet spot — más grande rompe grids
- **L-019:** Speeches vs Process responses: agent workflow requiere texto DECLARATIVO (lo que pasó al cliente) no narrativo (lo que el agente dice)
- **L-020:** Datos internos del operador (PDFs Movistar, XLSXs) → `.gitignore` con `Recursos CIMA/` + `*.xlsx`

### Files modificados

- `apps/cima-generador/index.html` — +TIPIFICACIÓN TX feature (~+50KB por CSS/SVG/data inline)
- `apps/cima-generador/VERSION` — `2.1.8` → `3.0`
- `apps/cima-generador/CHANGELOG.md` — esta entrada
- `apps/cima-generador/test_popup.py` — +46 tests (120 → 166)
- `MEMORY.md` — documentación arquitectural v3.0

### Pendiente

- ⏳ Push a GitHub Pages (CEO debe autorizar; 1 subtree split + force-push pendiente)
- ⏳ Bundle LICENSE en deploy standalone si CEO decide distribución parcial (compliance CC BY 4.0 + ODbL)

---

## [2.1.8] - 2026-10-15 — LIMPIEZA FOOTER + TIPIFICACIÓN PLACEHOLDER

### Removed
- **Footer (HTML 1583-1598):** Eliminado `<span>v2.0 standalone</span>` (marker obsoleto de versión).
- **Footer (HTML 1591-1597):** Eliminado bloque `.footer-attribution` con atribución visible a GeoNames CC BY 4.0 + RitchieRD ODbL (CEO directive: UI minimal). La atribución legal permanece documentada en `LICENSE` (data attribution section) del repo.
- **Tipificación (HTML 1569):** Eliminado `<p class="tipif-process-text">Tipificación dinámica disponible en próxima iteración</p>`.
- **Tipificación (HTML 1570-1578):** Eliminado `<span class="tipif-process-cta">v2.1.7 · standalone (no fetch)</span>` con su SVG lock icon.

### Test suite update
- `'v2.0 marker present'` → actualizado a `v2.1.8 v2.0 standalone marker removido del HTML (preserved in VERSION file)`.
- `'GeoNames attribution in HTML (source comment)'` → actualizado a `v2.1.8 visible data attribution removida del footer (preserved in LICENSE file)`.

### Preserved
- `<svg clock>...</svg> + <span>Próximamente</span>` badge del Tipificación (líneas 1562-1567) → mantiene UX de "feature coming".
- `<strong>Alexander Salvador</strong>` + heart → author credit en footer se mantiene.
- LICENSE file (data attribution section, líneas 22-46) → atribución legal GeoNames/CC BY 4.0 + RitchieRD/ODbL preservada para distribución.

⚠️ **License compliance note:** Si el popup se distribuye standalone (Google Sites embed, file://, popup window) sin el `LICENSE` file bundleado, las atribuciones CC BY 4.0 y ODbL dejan de ser visibles al usuario. La atribución en `LICENSE` satisface el código fuente del repo, pero **recomendado**: documentar este bundle en deploy pipeline.

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

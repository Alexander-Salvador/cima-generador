# Auditoría de Seguridad — cima-generador

**Fecha:** 2026-10-14 (v2.0.0 + v2.0.1)
**Auditor:** Akiko (adversarial_intelligence) + Diana (security) — OVAV
**Versión auditada:** v2.0.1 (post-fix)
**Stack:** Vanilla HTML/JS · single-file 2110+ LOC · 745KB

---

## Issues detallados

### 🔴 #1 CRITICAL — Tipif textareas inutilizables (focus loss)

**Evidencia:** Línea 2031 `tipifList.addEventListener('input', tipifListHandler)` → L1919 `setTipif` → L1805 `renderTipifs()` → L1831 `list.innerHTML = tipifs.map(...)`.

Cada keystroke en `consulta`/`respuesta`/`idc`/`numero` destruye y recrea el `<textarea>`. Cursor y focus se pierden **entre caracteres**.

**Repro:**
1. Click "+ Agregar"
2. Elegir "Internet"
3. Tipear "Cliente sin internet" en Consulta
4. Solo aparece la primera letra

**Mitigación:**
- No re-renderizar la lista en cada `input`
- Usar `change`/`blur` para persistir
- Actualizar `output-box` con `textContent` directo (sin destruir el textarea)
- O marcar bloque modificado y reemplazar solo ese `<div class="tipif-block">`

---

### 🟠 #2 HIGH — DoS por paste de string mega-largo

**Evidencia:** Inputs `direccion` (L1052), `referencia` (L1058), `ccpp` (L1063), deptos/prov/dist (L1027-1043) **sin `maxlength`**. Cada keystroke, `acFilter` (L1368) ejecuta `opt.toLowerCase().indexOf(q)` para todas las opciones. Con paste de 1MB, browser freezea.

**Repro:** DevTools console → `paste('x'.repeat(1_000_000))` en `#ccpp` → `input` event → `acRender` → O(n × m) con n=316 caseríos, m=1M.

**Mitigación:**
- `maxlength="120"` en todos los `<input>` de texto
- Debounce `acRender` con `requestAnimationFrame` o `setTimeout(..., 80)`

---

### 🟠 #3 HIGH — Acumulación ilimitada de tipificaciones (memory exhaustion)

**Evidencia:** `tipifCounter++` (L1766) y `tipifs.push(...)` (L1768) sin tope. Sin limpieza en `removeTipif`.

**Repro:** DevTools → `for(let i=0;i<10000;i++) addTipif()`. Cada bloque renderiza ~5 botones + 3 inputs + 2 textareas. DOM ≈ 100k nodos.

**Mitigación:** Cap `if (tipifs.length >= 50) { alertMsg('Máx 50 tipificaciones'); return; }`.

---

### 🟠 #4 HIGH — JSON sin Subresource Integrity (MITM → XSS)

**Evidencia:** L9-10 `<link rel="preload" as="fetch">` y L1537/1588 `fetch('data/*.json')` **sin atributo `integrity`**. Si el sitio Google Sites sirve sobre HTTP o un proxy MITM, los JSON pueden inyectar nombres maliciosos. `acEscape` (L1397) escapa solo `& < >` — **no escapa `"`**.

**Repro:** Reemplazar ubigeo JSON con `{"<img src=x onerror=alert(1)>": ...}` → al tipear departamento, `data-val="<img src=x onerror=alert(1)>"` rompe el atributo (doble quote no escapada). Vector de **attribute injection → XSS** si payload estilo `foo" onclick="..."` llega al dropdown.

**Mitigación:**
- Inline del JSON como `const UBIGEO = {...}` en el `<script>` (no fetch)
- O servir desde CDN con SRI + HTTPS-only

---

### 🟡 #5 MEDIUM — `acEscape` incompleto (defensa en profundidad)

**Evidencia:** L1397: `replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')` — falta `"`, `'`, `/`. Mismo bug en `escAttr` (L1815).

**Mitigación:**
```javascript
String(s).replace(/[&<>"'\/]/g, c => ({
  '&': '&amp;', '<': '&lt;', '>': '&gt;',
  '"': '&quot;', "'": '&#39;', '/': '&#x2F;'
}[c]))
```

---

### 🟡 #6 MEDIUM — Sin Content-Security-Policy

**Evidencia:** `<head>` (L1-10) **sin `<meta http-equiv="Content-Security-Policy">`**. Inline `<script>`, `onclick="copyOutput(...)"`, `onclick="copyTipifOutput(...)"`. Si #4 se materializa, XSS tiene vía libre.

**Mitigación:**
```html
<meta http-equiv="Content-Security-Policy"
  content="default-src 'self'; script-src 'self' 'unsafe-inline';
  style-src 'self' 'unsafe-inline'; img-src data:; base-uri 'none';">
```

---

### 🟡 #7 MEDIUM — Falla silenciosa en CCPP (no retry, no diagnóstico)

**Evidencia:** L1601-1605: si `ccpp_by_distrito.json` falla o no contiene la ubigeo key del distrito, `CCPP_DATA = {}` → dropdown queda vacío sin más info. Usuario asume "no hay caseríos". Inconsistente con `ubigeo` que sí distingue error vs vacío.

**Repro:** Borrar manualmente una key del JSON → seleccionar ese distrito → caseríos desaparecen sin alerta.

**Mitigación:**
- Loggear a `console.warn` qué distrito falló
- Mensaje específico "No se cargaron caseríos para este distrito" + botón "Reintentar"

---

### 🟡 #8 MEDIUM — Race: typing antes del fetch resuelto

**Evidencia:** L1537 `loadUbigeo` async; mientras tanto usuario tipea en `departamento`. `acOptionsFor` retorna `[]` (L1337). Al resolver, L1547 `acRender('departamento')` dispara dropdown aunque el usuario ya pasó a otro campo. UX: dropdown fantasma.

**Repro:** Throttle network a "Slow 3G" → tipear "Lima" rápido → dropdown se abre tarde.

**Mitigación:** Solo re-renderizar si `document.activeElement === input`.

---

### 🟢 #9 LOW — `referencia` se limpia pero nunca se usa

**Evidencia:** `buildAddressForMaps` (L1942) no incluye `referencia`. Pero `clearAddress` (L1991) sí limpia el campo. Inconsistencia: data loss visual.

**Mitigación:** Decidir: incluir `referencia` en output geocoding o documentar que es solo nota interna.

---

### 🟢 #10 LOW — `window.close()` permite DoS por terceros

**Evidencia:** L2049/2058 `window.close()` invocable desde `popupClose` o Escape. En iframes cross-origin no funciona, pero expone `window.opener` reference (ya accesible por same-origin). Sin sandboxing attribute en el iframe padre.

**Repro:** Embebido en página atacante → child puede intentar `window.opener.location = 'phishing'` (blocked por SOP en cross-origin, pero no documentado).

**Mitigación:** Pedir a Google Sites embed con `sandbox="allow-scripts allow-forms"` (no `allow-same-origin`).

---

## Stack y dependencias

| Aspecto | Estado |
|---------|--------|
| **Frontend** | Vanilla JS · sin frameworks |
| **CSS framework** | Ninguno (sistema propio con CSS variables) |
| **Librerías externas** | Ninguna · 0 CDN · 0 `<script src>` remotos |
| **Backend** | No (100% client-side) |
| **Persistencia** | No (variables globales volátiles) |
| **APIs externas** | 0 (los JSON son locales) |
| **Google Maps API** | No se usa (nombre "Maps format" = geocoding string compatible) |
| **Fonts** | System fonts (sin Google Fonts) |

**Score seguridad superficie:** 9/10 (ataque surface es mínimo — todo client-side, sin backend)

---

## Compliance y atribución

| Dato | Licencia | Atribución |
|------|----------|------------|
| `ubigeo_consolidated.json` | ODbL (RitchieRD) | Pendiente de mostrar |
| `ccpp_by_distrito.json` | CC BY 4.0 (GeoNames) | Pendiente de mostrar |
| **Popup código** | No especificada (sin LICENSE) | Pendiente de decidir |

**Acción:** Agregar pie de página visible con créditos.

---

## Limitaciones técnicas

| Aspecto | Estado | Detalle |
|---------|--------|--------|
| **Responsive** | 🟡 Parcial | Sin media queries específicas |
| **WCAG / a11y** | ✅ Bien | ARIA roles, srStatus/srAlert, focus-visible |
| **Browser support** | Modern | Requiere evergreen (prefers-color-scheme, color-mix()) |
| **Race conditions** | 🟢 Guarded | Tokens invalidan respuestas stale |
| **Concurrencia 50+ usuarios** | 🟢 Optimizado | CCPP lazy-load evita 612KB al 70% usuarios |
| **CORS / hosting** | ⚠️ | Requiere servir desde HTTP (no `file://`) |

---

*Análisis generado por Akiko (adversarial_intelligence OVAV) — 2026-10-14*
*Code review por Helena (deep_explorer) y Nadia (documentation) — 2026-10-14*

# Changelog — cima-generador

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

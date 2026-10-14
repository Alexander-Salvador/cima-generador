# cima-generador

**v2.0.0** · popup standalone para Google Sites · generador de plantillas CIMA (Address + Tipificación)

## ¿Qué es?

Popup HTML autocontenido (sin backend, sin CDN, sin fetch externo) que genera
plantillas de dirección en formato Google Maps / Nominatim a partir del catálogo
de ubigeo del Perú (25 deptos · 196 prov · 1,892 dist) y 43,491 caseríos.

## Stack

- **Frontend:** Vanilla JS · sin frameworks
- **CSS:** Sistema propio con design tokens (CSS variables)
- **Backend:** ❌ Ninguno (100% client-side)
- **CDN/libs externos:** ❌ Ninguno
- **Persistencia:** ❌ Ninguna (variables en memoria)

## Datos (inlineados en el HTML, NO se cargan por fetch)

| Dato | Cobertura | Licencia | Atribución |
|------|-----------|----------|-----------|
| Ubigeo (depto→prov→dist) | 25 / 196 / 1,892 | ODbL | RitchieRD/ubigeos-peru-data |
| CCPP (centros poblados) | 43,491 caseríos | CC BY 4.0 | GeoNames PE.zip |

Para regenerar la data inlineada, ver `docs/SOURCES.md` y `data/` (backup).

## Quick start

### Local (desarrollo)

```bash
cd apps/cima-generador
python3 -m http.server 8000
# Abrir: http://localhost:8000
```

⚠️ **No abrir como `file://`** — los navegadores bloquean CORS.

### Google Sites (producción)

```html
<iframe
  src="https://alexander-salvador.github.io/cima-generador/"
  width="100%" height="800" frameborder="0"
  sandbox="allow-scripts allow-forms allow-popups"
  title="CIMA Generador"></iframe>
```

**URL pública:** https://alexander-salvador.github.io/cima-generador/

## Versiones

- **v2.0.0** (2026-10-14) — Standalone. Datos inlineados (sin fetch). FIX #1-#6.
- v1.0.0 (2026-10-13) — Eager-load ubigeo, lazy-load CCPP via fetch.

## Seguridad

Ver `docs/AUDIT.md` para análisis adversarial completo.

**Estado v2.0.0:**

| Sev | Issue v1 | Estado v2 |
|-----|----------|-----------|
| CRITICAL | #1 tipif focus loss | ✅ Resuelto (re-render selectivo) |
| HIGH | #2 DoS paste | ✅ Resuelto (maxlength + debounce) |
| HIGH | #3 memory sin cap | ✅ Resuelto (cap 50 tipifs) |
| HIGH | #4 JSON sin SRI | ✅ Resuelto (data inline, sin fetch) |
| MEDIUM | #5 acEscape | ✅ Resuelto (full escape) |
| MEDIUM | #6 sin CSP | ✅ Resuelto (CSP meta agregada) |
| MEDIUM | #7 CCPP silent fail | 🟡 Pendiente (mensaje específico) |
| MEDIUM | #8 race typing | 🟡 Pendiente (activeElement check) |
| LOW | #9 referencia data loss | 🟡 Pendiente (decisión) |
| LOW | #10 sandbox | ✅ Documentado (sandbox en iframe) |

## Atribución obligatoria

Datos © GeoNames (CC BY 4.0) y RitchieRD (ODbL).
Popup © Alexander Salvador.

## Más info

- `CHANGELOG.md` — release notes
- `docs/AUDIT.md` — análisis adversarial
- `docs/SOURCES.md` — atribución de datos
- `docs/INTEGRATION.md` — guía embed Google Sites

---

*Cima-generador es un mini-proyecto independiente de AKRYNT, exclusivo para CIMA.*
*Importado al repo CIMA standalone el 2026-10-14.*

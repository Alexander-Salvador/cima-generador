# cima-generador

**v3.10.86 local** (LIVE: v3.10.82) · popup standalone para Google Sites · generador de plantillas CIMA
(Address + Tipificación + **Zonas de apagado**)

## ¿Qué es?

Popup HTML autocontenido (sin backend, sin CDN, sin fetch externo) que genera
plantillas de dirección en formato Google Maps / Nominatim a partir del catálogo
de ubigeo del Perú (25 deptos · 196 prov · 1.892 dist) y 43.491 caseríos.
Además detecta **zonas de apagado** (estaciones base / red 2G) y entrega el speech
de atención correspondiente al supervisor.

## Stack

- **Frontend:** Vanilla JS · sin frameworks
- **CSS:** Sistema propio con design tokens (CSS variables)
- **Backend:** ❌ Ninguno (100% client-side)
- **CDN/libs externos:** ❌ Ninguno
- **Persistencia:** ❌ Ninguna (variables en memoria)

## Datos (inlineados en el HTML, NO se cargan por fetch)

| Dato | Cobertura | Licencia | Atribución |
|------|-----------|----------|-----------|
| Ubigeo (depto→prov→dist) | 25 / 196 / 1.892 | ODbL | RitchieRD/ubigeos-peru-data |
| CCPP (centros poblados) | 43,491 caseríos | CC BY 4.0 | GeoNames PE.zip |
| Zonas de apagado COMU-5514 | 238 CCPP exactos (3 grupos) | **Interno Movistar** | PDFs internos (no redistribuibles) |
| Zonas de apagado COMU-5672 (IPT) | 1.798 zonas por distrito | **Interno Movistar** | PDF interno (no redistribuible) |

Para regenerar la data inlineada, ver `docs/SOURCES.md` y `data/` (backup).

## Quick start

### Local (desarrollo)

```bash
cd apps/cima-generador
python3 serve.py          # puerto 8765, anti-cache
# Abrir: http://localhost:8765/index.html?popup=1
```

⚠️ **Solo `localhost`.** `127.0.0.1` e IPs de LAN están prohibidos como origen
de desarrollo: son origins distintos con state de browser separado, y las IPs de
LAN por HTTP plano no son *secure context* (`serviceWorker` y `crypto.subtle` no
existen). Ley del origen canónico en `AGENTS.md`.

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

## Zonas de apagado

| Campaña | Detalle | Fecha mostrada |
|---|---|---|
| COMU-5514 | Apagado de estaciones base · 238 CCPP · Grupo 1 (16) / Grupo 2 (201) / Grupo 3 (21) | "inicio campaña 11/09/2026" + aviso; la fecha por zona aún no está verificada (recurso interno) |
| COMU-5672 | Apagado de red 2G (IPT) · 1.798 zonas por distrito | fecha por departamento (28-ago / 03-sep-2026) |
| COMU-5569 | Campaña diciembre 2026 | ⏳ sin cobertura cargada (falta la tabla de zonas) |

**Reglas del sistema**
- El comunicado se **atribuye por campaña** (`ZONAS_APAGADAS_INTEL.campanas`), no hardcodeado.
- El **matching de caserío es por fold** (sin diacríticos): "Ocallí" (GeoNames) y
  "OCALLI" (PDF) son la misma clave.
- **Nunca se afirma una fecha de zona sin verificarla** → se muestra el inicio de
  campaña etiquetado + aviso, y el speech pide la fecha del centro poblado.

Scripts (reproducibles):

```bash
python3 scripts/audit_zonas_apagado.py    # fuente ↔ runtime (pérdidas · extras · precisión)
python3 scripts/security_sweep_zonas.py   # 21 checks anti-datos-falsos
python3 scripts/export_tabla_zonas.py     # tabla visual XLSX + resumen MD
python3 scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 --source <TABLA> --dry-run
```

## Versiones

- **v3.10.86** (2026-11-14) — fecha por zona (grupos) · barrido de seguridad como gate · tabla XLSX
- **v3.10.85** (2026-11-14) — barrido de seguridad (21 checks) + verificación OCR doble
- **v3.10.84** (2026-11-14) — barrido completo contra la fuente (4 CCPP recuperados)
- **v3.10.83** (2026-11-14) — registro de campañas + COMU-5569 (pendiente)
- **v3.10.82** (2026-11-07) — RESET nuevo · APN→JIRA · CASO eliminado
- **v2.0.0** (2026-10-14) — Standalone. Datos inlineados (sin fetch). FIX #1-#6.
- v1.0.0 (2026-10-13) — Eager-load ubigeo, lazy-load CCPP via fetch.

Historial completo: [`CHANGELOG.md`](./CHANGELOG.md)

## Seguridad

Ver `docs/AUDIT.md` para análisis adversarial completo.

**Estado v2.0.0 (frontend):**

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

**Integridad de datos (v3.10.85+):** `test_sweep_zonas.py` es un gate con 21 checks
que fallan si el popup afirma una zona sin respaldo en la fuente, si una fila de la
fuente queda fuera sin declararse, si se afirma una fecha de zona no verificada o
si aparecen secretos/URLs no documentados.

## Atribución obligatoria

Datos © GeoNames (CC BY 4.0) y RitchieRD (ODbL).
Popup © Alexander Salvador.
Los datos de zonas de apagado son **material interno Movistar**: prohibido
reproducirlos o distribuirlos (regla de los PDFs de origen).

## Más info

- `CHANGELOG.md` — release notes
- `docs/AUDIT.md` — análisis adversarial
- `docs/SOURCES.md` — atribución de datos
- `docs/INTEGRATION.md` — guía embed Google Sites
- `docs/FUENTES_ZONAS_APAGADO_2026-11-14.md` — auditoría de zonas (H1–H9) + decisiones
- `docs/TABLA_ZONAS_APAGADO_RESUMEN_2026-11-14.md` — inventario visual de zonas
- `docs/FUENTES_ZONAS_APAGADAS.md` — fuentes y trazabilidad de zonas

---

*Cima-generador es un mini-proyecto independiente de AKRYNT, exclusivo para CIMA.*
*Importado al repo CIMA standalone el 2026-10-14.*

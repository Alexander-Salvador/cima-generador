# Integration Guide — cima-generador en Google Sites

**Fecha:** 2026-10-14

## Quick Start

1. Hostear `index.html` + `data/` en un host estático (GitHub Pages, Netlify, Vercel)
2. Asegurar HTTPS
3. En Google Sites: Insertar → Embed → URL del `index.html`
4. Ajustar dimensiones iframe: 760px ancho, alto según necesidad

## Hosting recomendado

### Opción A — GitHub Pages (gratis, ya en uso)

El repo `https://alexander-salvador.github.io/cima-generador/` ya está publicado.

URL pública: `https://alexander-salvador.github.io/cima-generador/`

### Opción B — Local (desarrollo)

```bash
cd apps/cima-generador
python3 -m http.server 8000
# Abrir: http://localhost:8000
```

**IMPORTANTE:** No abrir como `file://` — los `fetch()` a `data/*.json` fallan por CORS.

---

## Embedding en Google Sites

```html
<iframe
  src="https://alexander-salvador.github.io/cima-generador/"
  width="100%"
  height="800"
  frameborder="0"
  sandbox="allow-scripts allow-forms allow-popups"
  title="CIMA Generador de Plantillas"
></iframe>
```

**Nota:** No usar `allow-same-origin` en el sandbox — defensa contra #10 LOW.

---

## Cross-reference con CIMA spreadsheet

Una vez generado el address, copiar al portapapeles (botón en popup) y pegar en CIMA spreadsheet.

**Flujo supervisor:**
1. Abre Google Sites con popup embed
2. Llena Address (depto → prov → dist → ccpp → dirección)
3. Click "Copiar" en output
4. Pega en CIMA!FEEDBACK.G (OBSERVACIÓN) o en HE según corresponda

---

## Atribución obligatoria

Mostrar en el footer de la página de Google Sites:

```
Datos © GeoNames (CC BY 4.0) · Ubigeo © RitchieRD (ODbL) · Popup © Alexander Salvador
```

---

*Documento generado 2026-10-14 al importar cima-generador al repo CIMA standalone.*

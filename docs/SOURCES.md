# Data Sources — cima-generador

## Ubigeo (Perú: departamentos → provincias → distritos)

| Campo | Valor |
|-------|-------|
| **Source** | RitchieRD/ubigeos-peru-data (open source) |
| **License** | ODbL (Open Database License) |
| **URL** | https://github.com/RitchieRD/ubigeos-peru-data |
| **Coverage** | 25 departamentos · 196 provincias · 1,892 distritos |
| **Format** | JSON nested con `_ubigeo` (códigos padded: "01", "0101", "010101") + `_id` |
| **Encoding** | UTF-8 |
| **Last update** | 2026-10-13 (commit `82d0706`) |

**Atribución obligatoria:** Mantener enlace visible al proyecto original.

---

## CCPP / Caseríos (Perú)

| Campo | Valor |
|-------|-------|
| **Source** | GeoNames PE.zip (download.geonames.org) |
| **License** | CC BY 4.0 (Creative Commons Attribution) |
| **URL** | http://download.geonames.org/export/dump/PE.zip |
| **Coverage** | 43,491 caseríos en 1,851 distritos |
| **Format** | JSON dict: `ccpp_by_distrito[ubigeo_6d] → string[]` |
| **Encoding** | UTF-8 |
| **Last update** | 2026-10-13 (commit `88edb16`) |

**Atribución obligatoria (CC BY 4.0):**
- Mencionar GeoNames como fuente
- Indicar si hubo modificación (sí — agrupado por distrito)
- URL de la licencia: https://creativecommons.org/licenses/by/4.0/

---

## Datos NO incluidos (pendiente decisión)

| Dato | Estado |
|------|--------|
| Tipificación catálogo | "Próximamente" — no incluido en v1.0.0 |
| Historial cambios | No aplica — versión v1.0.0 inicial |
| Datos personales (DNI, etc.) | **NO** — popup no maneja PII |

---

## Renovar datos

```bash
# 1. Ubigeo desde RitchieRD
git clone https://github.com/RitchieRD/ubigeos-peru-data
# Procesar → data/ubigeo_consolidated.json

# 2. CCPP desde GeoNames
wget http://download.geonames.org/export/dump/PE.zip
unzip PE.zip
# Agrupar por ubigeo distrito → data/ccpp_by_distrito.json
```

**Cadencia recomendada:** revisión trimestral (4 veces/año).

---

*Documento generado 2026-10-14 al importar cima-generador al repo CIMA standalone.*

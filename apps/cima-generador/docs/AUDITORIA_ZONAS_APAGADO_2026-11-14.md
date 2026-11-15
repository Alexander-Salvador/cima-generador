# CIMA — Auditoría de zonas de apagado

**Versión auditada:** v3.10.84  
**Fecha:** 2026-11-14  
**Alcance:** zonas de apagado de estaciones base (COMU-5514), zonas IPT (COMU-5672) y cobertura pendiente de COMU-5569  
**Fuente:** PDFs internos restringidos; no versionados en el repositorio

---

## 1. Resumen ejecutivo

- Se barrió la totalidad de las zonas de apagado de CIMA, reconstruyendo la verdad desde los PDFs fuente y cruzándola contra el runtime.
- La fuente contiene 243 CCPP únicos; el sistema pasó de 234 a 238 CCPP exactos y de 188 a 191 distritos marcados.
- Se restauraron 4 zonas legítimas que no alertaban por limpieza histórica y diferencias de diacríticos (H1).
- Se eliminaron los falsos positivos del audit check 5 mediante validación contra GeoNames y una allowlist mínima (H2).
- Permanecen 3 zonas de la fuente sin cobertura y 4 caseríos con atribución d6 divergente; requieren decisión del CEO (H3/H4).
- Se corrigieron la divergencia del data layer y el invariante imposible del audit check 6 (H6/H7).
- COMU-5569 sigue pendiente con 0 CCPP: los cuatro PDF cargados son el mismo comunicado, no la tabla de zonas.

---

## 2. Alcance y método

### 2.1 Alcance

El barrido cubrió todas las zonas de apagado conocidas en CIMA:

1. **COMU-5514:** estaciones base con alerta a nivel de centro poblado (CCPP exacto).
2. **COMU-5672:** zonas IPT con alcance distrital, mantenidas en un set separado.
3. **COMU-5569:** campaña de diciembre de 2026, verificada como pendiente de tabla de zonas.

La verdad de referencia se reconstruyó desde los tres PDFs de `Recursos CIMA/Bases y Zonas de Apagado/` y el JSON OCR manual. Luego se contrastó cada clave con el runtime, UBIGEO y GeoNames.

### 2.2 Método

1. Extraer Grupo 2 por texto (`pdfplumber`) y Grupos 1/3 por OCR visual ya materializado.
2. Incorporar `data/zonas_apagadas_manual.json` y resolver cada fila contra UBIGEO por d6 exacto o fuzzy.
3. Normalizar el CCPP con el mismo fold del runtime: `d6:CCPP_FOLD`.
4. Comparar sets fuente/runtime, buscar pérdidas, extras, duplicados, huérfanos y aliases.
5. Contrastar `CCPP_DATA`, el data layer de UBIGEO y la cobertura por campaña.
6. Ejecutar las suites y el audit runtime después de los fixes v3.10.84.

### 2.3 Reglas de normalización

| Dominio | Regla | Uso |
|---|---|---|
| Geografía | trim + UPPER + espacio simple + NFC | Departamento, provincia y distrito |
| CCPP | NFKD + eliminación de diacríticos + UPPER + espacio simple + NFC | `zaNormalizeCcpp()` en runtime y fold equivalente en Python |
| Clave de sistema | `d6:CCPP_FOLD` | Une grafías con/sin tilde sin duplicar la zona |
| Alias | Allowlist explícita y mínima | Solo para literales documentados; no resuelve ambigüedades geográficas |

El fuzzy se aplica a resolución de distrito/provincia contra UBIGEO. Los nombres de CCPP se comparan por fold, no por coincidencia literal.

### 2.4 Artefactos del barrido

Los artefactos son locales y gitignored porque derivan de una fuente restringida.

| Artefacto | Contenido |
|---|---|
| `scripts/audit_zonas_apagado.py` | Barrido reproducible; escribe el JSON y el CSV derivados |
| `data/audit_zonas_apagado.json` | Detalle programático: sistema, fuente, pérdidas, extras, aliases, campañas y divergencias |
| `data/apagados_tabla_completa.csv` | 2.081 filas: 282 COMU-5514 (247 CCPP únicos + 35 duplicados), 1.798 COMU-5672 IPT y 1 COMU-5569 pendiente |

El CSV tiene 22 columnas: `comunicado`, `campana_estado`, `version`, `fecha_apagado`, `fecha_sms_inicio`, `nivel`, `depto`, `provincia`, `distrito`, `centro_poblado_ccpp`, `ubigeo_d6`, `key_sistema`, `cobertura_en_sistema`, `estado_validacion`, `metodo_resolucion`, `ccpp_en_caserios`, `en_caserios_igual_a_distrito`, `speech_key`, `postventa_speech_key`, `origen_pdf`, `grupo`, `pagina`, `notas`.

---

## 3. Inventario de fuente

| PDF / archivo | Grupo | Extracción | MD5 | Filas |
|---|---|---|---|---:|
| `APAGADO.pdf` | Grupo 2, texto, 3 páginas | `pdfplumber` | `83cda2025f1c88e3788fb056e34f215b` | 204 |
| `APAGADP1.pdf` | Grupo 1, imagen | OCR visual | `73559872db3b2e12e62b6171b00b4fe5` | 16 |
| `APAGADO2.pdf` | Grupo 3, imagen | OCR visual | `a9f231a542d88e3d7ed9b77511c92626` | 21 |
| `data/zonas_apagadas_manual.json` | OCR manual | JSON local, solapado con los PDF | — | 37 |

**Agregado:** 278 filas de entrada (241 PDF + 37 manuales) → **243 CCPP únicos** tras deduplicar por clave. Las 35 filas repetidas se conservan en el CSV con `estado_validacion=duplicado_fuente`.

---

## 4. Inventario del sistema

| Métrica de runtime | Antes del barrido | Después | Delta |
|---|---:|---:|---:|
| CCPP exactos COMU-5514 | 234 | 238 | +4 |
| Distritos en `districts_with_off` | 188 | 191 | +3 |
| Issues del audit runtime por carga | 3 | 0 | -3 |
| Zonas IPT COMU-5672 | 1.798 | 1.798 | 0 |

El set IPT permanece separado de `ccpp_keys`; no se contabiliza como cobertura CCPP exacto.

---

## 5. Hallazgos

### H1 · Cuatro zonas legítimas no alertaban

**Qué:** cuatro CCPP presentes en la fuente no producían alerta en runtime.

| Clave runtime | Nombre en GeoNames | Distrito |
|---|---|---|
| `010512:OCALLI` | Ocallí | Ocallí |
| `040411:TIPAN` | Tipán | Tipán |
| `080206:RONDOCAN` | Rondocán | Rondocán |
| `221003:POLVORA` | Pólvora | Pólvora |

**Evidencia:** la limpieza v3.10.13 eliminó esas claves al aplicar la regla “CCPP == nombre de distrito” como contaminación. El matcher posterior comparaba con NFC + UPPER, por lo que tampoco reconciliaba `Ocallí`/`OCALLI`, `Tipán`/`TIPAN`, `Rondocán`/`RONDOCAN` ni `Pólvora`/`POLVORA`.

**Impacto en el supervisor:** cuatro falsos negativos. Un supervisor de esas zonas no recibía la alerta de apagado de estación base.

**Fix aplicado (v3.10.84):** restaurar las cuatro claves e introducir `zaNormalizeCcpp()` —fold NFKD sin diacríticos— en los seis puntos de lookup.

### H2 · El audit check 5 producía falsos positivos

**Qué:** el check 5 marcaba 38 caseríos legítimos como “contaminados” solo porque el CCPP coincidía con el nombre del distrito. Un ejemplo es `010116:OLLEROS`, que sí figura en GeoNames.

**Evidencia:** la igualdad literal con el distrito no distingue entre:

- un caserío legítimo con el mismo nombre;
- una cabecera de distrito;
- una variante ortográfica del PDF;
- un CCPP realmente inexistente.

**Impacto en el supervisor:** el audit reportaba issues en cada carga y erosionaba la utilidad de la señal de integridad.

**Fix aplicado (v3.10.84):** la regla exige que el CCPP no esté en `CCPP_DATA[d6]` bajo fold antes de marcarlo. Solo quedan cuatro excepciones documentadas en `CCPP_CASERIO_ALIASES`:

| Clave | Literal de fuente | Motivo |
|---|---|---|
| `030608:OCCCEPATA` | `OCCEPATA` | Variante ortográfica del PDF |
| `080602:CHUQUICAHUANA` | `CHUQICAHUANA` | Variante ortográfica del PDF |
| `090119:HUANDO` | `HUANDO` | Cabecera de distrito sin caserío GeoNames |
| `110208:SAN JUAN DE YANAC` | `SAN JUAN DE YANAC` | Cabecera de distrito sin caserío GeoNames |

### H3 · Tres zonas de la fuente siguen sin cobertura

| Clave candidata | Evidencia | Estado |
|---|---|---|
| `010105:CHILIQUIN` | El PDF escribe el nombre del distrito en la columna CENTRO POBLADO; GeoNames lista “Chiliquini”. Semántica ambigua: ¿caserío o todo el distrito? | Sin cobertura; decisión D4 |
| `060101:INCA CORRAL` | No existe en UBIGEO ni en `CCPP_DATA`; el PDF además escribe provincia `CAJABAMBA` frente a `CAJAMARCA` | Sin cobertura; decisión D5 |
| `160405:SANTA ROSA` | El distrito SANTA ROSA DE LORETO existe en el UBIGEO inline, pero tiene 0 caseríos en `CCPP_DATA` | Sin cobertura; decisión D5 |

**Impacto en el supervisor:** no se genera alerta para esas tres filas. Es preferible a fabricar un CCPP o una clave no resoluble.

**Estado:** pendientes de decisión del CEO; no se cargaron como cobertura exacta.

### H4 · Cuatro caseríos con d6 divergente

El PDF atribuye los cuatro caseríos a PIURA / AYABACA / PACAIPAMPA, d6 `200206`. GeoNames los ubica en AYABACA, d6 `200201`, que es la clave usada por el runtime.

| CCPP en PDF | CCPP en runtime | d6 PDF | d6 GeoNames/runtime |
|---|---|---:|---:|
| CURRILCAS | CURRILCAS | `200206` | `200201` |
| PAPELILLO | PAPELLILLO | `200206` | `200201` |
| TOJAS | TOJAS | `200206` | `200201` |
| TULMAN DE MATALACAS | TULMAN DE MATALACAS | `200206` | `200201` |

**Evidencia:** el barrido registra las cuatro claves fuente como pérdidas y las cuatro claves runtime como extras solo antes de aplicar la equivalencia por CCPP.

**Impacto en el supervisor:** funcionalmente correcto: las claves `200201:*` son las alcanzables desde el dropdown del popup. La divergencia afecta la atribución geográfica frente a la fuente, no la capacidad de alerta.

**Estado:** decisión D6 pendiente; no se duplicaron claves bajo `200206`.

### H5 · Treinta y cinco filas duplicadas en la fuente

**Qué:** 278 filas de entrada representan 243 CCPP únicos; 35 filas repiten un CCPP ya presente en los PDF.

**Evidencia:** el dedupe por `d6:CCPP_FOLD` evita que el set runtime contenga duplicados. El CSV conserva la trazabilidad y marca cada repetición como `duplicado_fuente`.

**Impacto en el supervisor:** ninguno en alertas; evita doble conteo o doble marcaje en una futura regeneración.

**Estado:** documentado; los duplicados permanecen trazables y no requieren fix de runtime.

### H6 · Divergencia entre data layer y runtime

**Qué:** `data/ubigeo_consolidated.json` no contenía el d6 `160405` (SANTA ROSA DE LORETO) que sí estaba en el UBIGEO inline del popup.

**Evidencia:** el barrido comparó ambos conjuntos de d6 y detectó una divergencia.

**Impacto en el supervisor:** riesgo de que una clave se resolviera en runtime pero no en consumidores del data layer, o viceversa.

**Fix aplicado (v3.10.84):** alinear el data layer. Resultado del barrido: **0 divergencias**.

### H7 · El audit check 6 era imposible de satisfacer

**Qué:** el check 6 exigía el invariante `ipt_districts ⊆ districts_with_off` de v3.8.0. Desde v3.10.13 ambos sets representan campañas distintas y son disjuntos por diseño.

**Evidencia:** `districts_with_off` contiene los d6 de CCPP exactos COMU-5514; `ipt_districts` contiene los d6 de COMU-5672. La relación de subconjunto ya no aplica.

**Impacto en el supervisor:** un issue permanente y falso en cada carga, con riesgo de ocultar errores reales.

**Fix aplicado (v3.10.84):** validar que todo d6 de ambos sets exista en UBIGEO y que el set IPT no esté vacío.

---

## 6. Cobertura por campaña

| Campaña | Nivel | Fuente | Cobertura runtime | Estado |
|---|---|---|---|---|
| COMU-5514 | CCPP exacto / estación base | 3 PDF + OCR manual; 243 CCPP únicos de fuente | 238 CCPP exactos · 191 d6 marcados | `zonas_conocidas`; 3 sin cobertura y 4 atribuciones d6 pendientes de decisión |
| COMU-5672 | Distrito / red 2G IPT | PDF interno de zonas IPT | 1.798 zonas IPT | `zonas_conocidas`; set separado de CCPP exacto |
| COMU-5569 | CCPP exacto | Comunicado sin tabla de zonas | 0 CCPP | `zonas_pendiente`; 13.5K clientes, inicio 10-dic-2026 |

### COMU-5569

Los cuatro archivos PDF cargados por el CEO son el mismo documento, md5 `393c127b3b6b4bb486d3d03be0850f97`: es el comunicado, no la tabla de zonas. El detalle reside en los recursos internos geniomovistar `id=178147` / `id=178146`.

El popup no alerta en zonas sin cobertura cargada. Esta regla de anti-fabricación se mantiene hasta recibir la tabla de centros poblados, fecha de apagado por departamento y fecha de inicio de SMS.

---

## 7. Decisiones pendientes del CEO

| Decisión | Caso | Opciones | Recomendación | Estado |
|---|---|---|---|---|
| **D4** | `010105:CHILIQUIN` | A) mapear al CCPP GeoNames “Chiliquini”; B) cubrir todo el distrito CHILIQUIN | Confirmar con el emisor la semántica del PDF. Si significa todo el distrito, implementar cobertura distrital separada; no crear cobertura CCPP exacta por inferencia. | Pendiente |
| **D5** | `060101:INCA CORRAL` y `160405:SANTA ROSA` | A) incorporar el CCPP faltante; B) mantener sin cobertura; C) cubrir el distrito | Solicitar una fuente geográfica/campaña autoritativa. Corregir `CAJABAMBA` a `CAJAMARCA` y resolver SANTA ROSA con el catálogo de caseríos. Hasta entonces, no fabricar claves. | Pendiente |
| **D6** | Atribución PACAIPAMPA / AYABACA | A) conservar d6 runtime `200201`; B) migrar a d6 fuente `200206` | Mantener `200201`: es la atribución GeoNames y la clave alcanzable desde el dropdown. Registrar la divergencia de fuente y pedir confirmación antes de migrar. | Pendiente |

No se implementará ninguna de estas decisiones sin aprobación explícita del CEO.

---

## 8. Verificación

| Suite / control | Resultado | Detalle |
|---|---|---|
| `test_popup.py` | PASS | 844/844 |
| `test_aggressive.py` | PASS | 238/238 fuzzy |
| `test_security.py` | CLEAN | Sin hallazgos |
| `test_codereview.py` | GOOD | Sin hallazgos |
| `test_browser_static.py` | PASS | Incluye `node --check` del JS inline |
| `test_ingest_comunicado_zonas.py` | PASS | 76/76 |
| Audit runtime | PASS | 0 issues |
| `node --check` | OK | Sintaxis del `<script>` inline |

---

## 9. Cómo reproducir

Desde `apps/cima-generador`, con los PDFs locales disponibles en `Recursos CIMA/Bases y Zonas de Apagado/`:

```bash
python3 scripts/audit_zonas_apagado.py
python3 test_popup.py
python3 test_aggressive.py
python3 test_security.py
python3 test_codereview.py
python3 test_browser_static.py
python3 test_ingest_comunicado_zonas.py
```

`test_browser_static.py` extrae el `<script>` inline y ejecuta `node --check`. El barrido regenera únicamente `data/apagados_tabla_completa.csv` y `data/audit_zonas_apagado.json`, ambos derivados y gitignored.

---

## 10. Lección

Un barrido contra la fuente es parte del gate, no una curiosidad documental.

- Los tests internos prueban consistencia; no prueban presencia de filas legítimas ausentes.
- La deriva de normalización produjo simultáneamente cuatro falsos negativos y falsos positivos recurrentes.
- Cada release que cambie cobertura debe reconciliar el set fuente contra runtime, con equivalencias explícitas, allowlist mínima y decisiones pendientes visibles.
- Ante ambigüedad no verificable, el sistema debe refrainerse de alertar en vez de fabricar cobertura.

---

*Auditor: Thavren — 2026-11-14 · barrido completo de zonas de apagado*

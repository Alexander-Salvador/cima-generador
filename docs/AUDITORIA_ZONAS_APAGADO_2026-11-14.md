# CIMA — Auditoría de zonas de apagado

**Versión auditada:** v3.10.86 (barrido v3.10.84 + H9 · fecha por zona)  
**Fecha:** 2026-11-14  
**Alcance:** zonas de apagado de estaciones base (COMU-5514), zonas IPT (COMU-5672), cobertura pendiente de COMU-5569 y honestidad de la fecha mostrada por zona  
**Fuente:** PDFs internos restringidos; no versionados en el repositorio

---

## 1. Resumen ejecutivo

- Se barrió la totalidad de las zonas de apagado de CIMA, reconstruyendo la verdad desde los PDFs fuente y cruzándola contra el runtime.
- La fuente contiene 242 CCPP únicos; el sistema pasó de 234 a 238 CCPP exactos y de 188 a 191 distritos marcados.
- Reconciliación fuente ↔ runtime: 238 CCPP en runtime · 242 CCPP únicos en fuente · 8 pérdidas (5 justificadas) · 4 extras (contraparte de las divergencias d6).
- Se restauraron 4 zonas legítimas que no alertaban por limpieza histórica y diferencias de diacríticos (H1).
- Se eliminaron los falsos positivos del audit check 5 mediante validación contra GeoNames y una allowlist mínima de 3 entradas (H2).
- La transcripción OCR de la fuente escaneada se verificó dos veces; la segunda pasada independiente detectó 2 errores de transcripción, ya corregidos (H2).
- Permanecen 3 zonas de la fuente sin cobertura y 4 caseríos con atribución d6 divergente; requieren decisión del CEO (H3/H4).
- Se corrigieron la divergencia del data layer y el invariante imposible del audit check 6 (H6/H7).
- La suma de clientes IPT por departamento de la fuente (+10,7% sobre lo declarado) es una inconsistencia de la fuente, no del sistema (H8).
- El popup afirmaba «apagado: 11/09/2026» para las 238 zonas: esa fecha es el INICIO de campaña del COMU-5514 v00003, no la fecha de cada zona; la fecha programada por departamento y el inicio de SMS están en 3 recursos internos no accesibles (H9). Desde v3.10.86 el panel etiqueta «inicio campaña: 11/09/2026» + aviso y el speech ya no afirma una fecha no verificada.
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
7. Re-transcribir de forma independiente las 37 filas de los PDF escaneados (render 300 DPI + lectura visual de las 3 bandas) y corregir la transcripción.
8. Ejecutar el barrido de seguridad de datos (`scripts/security_sweep_zonas.py`, checks S1–S21).
9. Verificar el comunicado COMU-5514 v00003 (1 pág, imagen; render 300 DPI + lectura visual) para separar el inicio de campaña de la fecha programada por zona (H9).

### 2.3 Reglas de normalización

| Dominio | Regla | Uso |
|---|---|---|
| Geografía | trim + UPPER + espacio simple + NFC | Departamento, provincia y distrito |
| CCPP | NFKD + eliminación de diacríticos + UPPER + espacio simple + NFC | `zaNormalizeCcpp()` en runtime y fold equivalente en Python |
| Clave de sistema | `d6:CCPP_FOLD` | Une grafías con/sin tilde sin duplicar la zona |
| Alias | Allowlist explícita y mínima | Solo para literales documentados; no resuelve ambigüedades geográficas |

El fuzzy se aplica a resolución de distrito/provincia contra UBIGEO. Los nombres de CCPP se comparan por fold, no por coincidencia literal.

### 2.4 Verificación doble del OCR

La fuente escaneada (APAGADP1.pdf G1 · APAGADO2.pdf G3) se transcribió **dos veces de forma independiente**: render a 300 DPI y lectura visual de las 3 bandas de cada PDF. La segunda pasada detectó 2 errores de transcripción en la columna CENTRO POBLADO, ya corregidos en `data/apagado_ocr_visual.json`.

| d6 | 1ª transcripción | Re-OCR (2ª pasada) | Lectura del PDF | Efecto |
|---|---|---|---|---|
| 030608 | `OCCEPATA` | `CCCEPATA` | El PDF escribe «CCCEPATA» | Typo real del PDF; el sistema guarda `030608:OCCCEPATA` (grafía GeoNames) |
| 080602 | `CHUQICAHUANA` | `CHUQUICAHUANA` | El PDF escribe «CHUQUICAHUANA» | Coincide literal con el PDF: sin alias, sin divergencia |

Resultado: APAGADP1.pdf 16/16 filas correctas · APAGADO2.pdf 19/21 correctas (las 2 corregidas). Tras la corrección, la allowlist `CCPP_CASERIO_ALIASES` bajó de **4 a 3 entradas** y la fuente reconstruida pasó de **243 a 242 CCPP únicos**: las 2 filas corregidas quedaron alineadas con las claves ya presentes (`030608:CCCEPATA` y `080602:CHUQUICAHUANA`) y el total colapsó en una clave.

**Lección:** el OCR de una fuente restringida se verifica dos veces antes de alimentar el sistema. Una lectura única traslada typos del papel a la base de datos y ensucia la allowlist.

### 2.5 Artefactos del barrido

Los artefactos son locales y gitignored porque derivan de una fuente restringida.

| Artefacto | Contenido |
|---|---|
| `scripts/audit_zonas_apagado.py` | Barrido reproducible; escribe el JSON y el CSV derivados |
| `scripts/security_sweep_zonas.py` | Barrido de seguridad de datos (checks S1–S21); escribe `data/security_sweep_zonas.json` |
| `data/audit_zonas_apagado.json` | Detalle programático: sistema, fuente, pérdidas, extras, aliases, campañas y divergencias |
| `data/security_sweep_zonas.json` | Resultado de los 21 checks de seguridad: 20 PASS · 1 WARN (S13/H8) · 0 FAIL |
| `data/apagados_tabla_completa.csv` | 2.081 filas: 282 COMU-5514 (278 de fuente → 242 CCPP únicos + 36 duplicadas, más 4 claves vigentes sin fila literal), 1.798 COMU-5672 IPT y 1 COMU-5569 pendiente |

El CSV tiene 28 columnas: `comunicado`, `campana_estado`, `version`, `fecha_apagado`, `fecha_sms_inicio`, `nivel`, `depto`, `provincia`, `distrito`, `centro_poblado_ccpp`, `ubigeo_d6`, `key_sistema`, `cobertura_en_sistema`, `estado_validacion`, `metodo_resolucion`, `ccpp_en_caserios`, `en_caserios_igual_a_distrito`, `speech_key`, `postventa_speech_key`, `origen_pdf`, `grupo`, `pagina`, `notas`, `grupo_campana`, `fecha_apagado_campana`, `fecha_apagado_zona`, `fecha_sms_zona`, `estado_fecha`.

---

## 3. Inventario de fuente

| PDF / archivo | Grupo | Extracción | MD5 | Filas |
|---|---|---|---|---:|
| `APAGADO.pdf` | Grupo 2, texto, 3 páginas | `pdfplumber` | `83cda2025f1c88e3788fb056e34f215b` | 204 |
| `APAGADP1.pdf` | Grupo 1, imagen | OCR visual | `73559872db3b2e12e62b6171b00b4fe5` | 16 |
| `APAGADO2.pdf` | Grupo 3, imagen | OCR visual | `a9f231a542d88e3d7ed9b77511c92626` | 21 |
| `data/zonas_apagadas_manual.json` | OCR manual | JSON local, solapado con los PDF | — | 37 |

**Agregado:** 278 filas de entrada (241 PDF + 37 manuales) → **242 CCPP únicos** tras deduplicar por clave. Las 36 filas repetidas se conservan en el CSV con `estado_validacion=duplicado_fuente`.

Las 37 filas de los PDF escaneados (G1 + G3) se re-transcribieron de forma independiente (render 300 DPI + lectura visual de las 3 bandas): 2 correcciones en la columna CENTRO POBLADO (`030608` y `080602`), 16/16 filas correctas en APAGADP1.pdf y 19/21 en APAGADO2.pdf. Detalle en §2.4.

---

## 4. Inventario del sistema

| Métrica de runtime | Antes del barrido | Después | Delta |
|---|---:|---:|---:|
| CCPP exactos COMU-5514 | 234 | 238 | +4 |
| Distritos en `districts_with_off` | 188 | 191 | +3 |
| Entradas en `CCPP_CASERIO_ALIASES` | 4 | 3 | -1 |
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

**Fix aplicado (v3.10.84):** la regla exige que el CCPP no esté en `CCPP_DATA[d6]` bajo fold antes de marcarlo. La allowlist `CCPP_CASERIO_ALIASES` quedó en **3 entradas** (antes 4): **1 typo real del PDF** + **2 cabeceras de distrito sin caserío GeoNames**.

| Clave | Literal de fuente | Motivo |
|---|---|---|
| `030608:OCCCEPATA` | `CCCEPATA` | Typo real del PDF (una `C` de más); el sistema guarda la grafía GeoNames |
| `090119:HUANDO` | `HUANDO` | Cabecera de distrito sin caserío GeoNames |
| `110208:SAN JUAN DE YANAC` | `SAN JUAN DE YANAC` | Cabecera de distrito sin caserío GeoNames |

`080602:CHUQUICAHUANA` salió de la allowlist: el PDF escribe literalmente «CHUQUICAHUANA», que coincide con la clave del sistema — no es typo y no necesita excepción.

**Verificación OCR (2 pasadas):** la fuente escaneada se transcribió dos veces de forma independiente. La 2ª pasada detectó 2 errores de transcripción en la columna CENTRO POBLADO (`030608` `OCCEPATA`→`CCCEPATA`, `080602` `CHUQICAHUANA`→`CHUQUICAHUANA`), ya corregidos. Esa corrección es la que dejó la allowlist en 3 entradas y la fuente reconstruida en 242 CCPP únicos (antes 243). Detalle completo en §2.4.

**Lección:** el OCR de una fuente restringida se verifica dos veces antes de alimentar el sistema. Un solo paso de lectura traslada typos del papel a la base de datos y ensucia la allowlist con excepciones que no corresponden.

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

El PDF escribe «CURILCAS» y «PAPELILLO»; GeoNames y el runtime usan «CURRILCAS» y «PAPELLILLO» — una letra de diferencia en cada caso.

**Evidencia:** el barrido registra las cuatro claves fuente como pérdidas y las cuatro claves runtime como extras solo antes de aplicar la equivalencia por CCPP.

**Impacto en el supervisor:** funcionalmente correcto: las claves `200201:*` son las alcanzables desde el dropdown del popup. La divergencia afecta la atribución geográfica frente a la fuente, no la capacidad de alerta.

**Estado:** decisión D6 pendiente; no se duplicaron claves bajo `200206`.

### H5 · Treinta y seis filas duplicadas en la fuente

**Qué:** 278 filas de entrada representan 242 CCPP únicos; 36 filas repiten un CCPP ya presente en los PDF.

**Evidencia:** el dedupe por `d6:CCPP_FOLD` evita que el set runtime contenga duplicados. El CSV conserva la trazabilidad y marca cada repetición como `duplicado_fuente`. Una de esas repeticiones (080602) solo se hizo visible al corregir el OCR: el PDF escribe «CHUQUICAHUANA» y la fila manual ya traía esa grafía.

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

### H8 · La suma de clientes IPT no cuadra con el total declarado

**Qué:** la campaña COMU-5672 declara 39.500 clientes, pero la suma de `clientes_depto` por departamento da **43.728** → Δ +4.228 (**+10,7%**).

**Evidencia:** check S13 del barrido de seguridad (`scripts/security_sweep_zonas.py`), que compara la suma por departamento contra el total declarado en el comunicado.

**Impacto en el supervisor:** la lógica de alerta no se ve afectada — el set IPT (1.798 zonas) es el mismo y cada zona se atribuye a su campaña. Lo que queda abierto es qué cifra de clientes puede mostrar el popup. **La inconsistencia está en la fuente** (suma por departamento frente al total del comunicado), no en el sistema.

**Estado:** WARN, no bloqueante. El barrido cierra con 20 PASS · 1 WARN · 0 FAIL (21 checks S1–S21). Requiere decisión del CEO (D7): qué cifra se muestra al supervisor.

### H9 · El popup afirmaba «apagado: 11/09/2026» para las 238 zonas

**Qué:** el panel mostraba **«apagado: 11/09/2026»** para **todas** las zonas de la base (238 CCPP). Esa fecha es el **inicio de campaña**, no la fecha de apagado de cada zona: es una afirmación falsa de precisión.

**Evidencia (COMU-5514 v00003 — 1 pág, imagen; render 300 DPI + lectura visual):**

- «Última actualización: 18/09/26»
- «Inicio: Desde el 11 de septiembre del 2026»
- 3 grupos, cada uno con su botón «Ver aquí»: **Grupo 1 (Q 1.7K)**, **Grupo 2 (Q 15K)**, **Grupo 3 (Q 1.7K, marcado «Nuevo» en v00003)**.

La fecha *programada por departamento* y el inicio de SMS están en esos 3 recursos internos (geniomovistar), no accesibles desde este entorno.

**Impacto en el supervisor:** 238 zonas con una fecha de zona no verificada afirmada como fecha de apagado; riesgo de informar al cliente una fecha que no corresponde a su centro poblado.

**Fix aplicado (v3.10.86):**

| # | Cambio | Detalle |
|---|---|---|
| 1 | `ZONAS_APAGADAS_INTEL.grupos_apagado` | 3 grupos con `ccpp_count`, `fuente_pdf`, `q_clientes`, `fecha_apagado_campana: 2026-09-11`, `fecha_apagado_zona: null`, `estado_fecha: "pendiente_recurso_interno"` |
| 2 | `ZONAS_APAGADAS_INTEL.grupo_por_key` | 238 claves → grupo (G1 16 · G2 201 · G3 21), derivado del PDF de origen de cada CCPP |
| 3 | `showApagadoPanelInline()` | La fecha se etiqueta según su certeza: **«apagado: DD/MM/YYYY»** solo si la fecha de esa zona está verificada; si no, **«inicio campaña: 11/09/2026»** + aviso visible («Fecha exacta de esta zona no confirmada… Confirmar la fecha antes de informarla al cliente») |
| 4 | `comu5514_postventa_fecha_pendiente` | Speech sin fecha para zonas sin fecha confirmada: «…dejará de funcionar en la zona {zona}, a partir de la fecha de apagado comunicada para tu centro poblado» |
| 5 | Tabla completa (CSV 28 columnas + XLSX) | Nuevas columnas `grupo_campana`, `fecha_apagado_campana`, `fecha_apagado_zona`, `fecha_sms_zona`, `estado_fecha` |
| 6 | Gate anti-fecha-falsa | `scripts/security_sweep_zonas.py` **S20** (todo CCPP tiene grupo) y **S21** (los grupos pendientes no declaran fecha por zona + el panel no afirma «apagado: <fecha>» de forma incondicional) |

**Estado:** popup corregido; falta el dato de los 3 recursos internos → decisión **D8**.

---

## 6. Cobertura por campaña

| Campaña | Nivel | Fuente | Cobertura runtime | Fechas | Estado |
|---|---|---|---|---|---|
| COMU-5514 | CCPP exacto / estación base | 3 PDF + OCR manual; 242 CCPP únicos de fuente | 238 CCPP exactos · 191 d6 marcados | Campaña 2026-09-11; fecha por zona pendiente (H9/D8) | `zonas_conocidas`; 3 sin cobertura y 4 atribuciones d6 pendientes de decisión |
| COMU-5672 | Distrito / red 2G IPT | PDF interno de zonas IPT | 1.798 zonas IPT | Por departamento: 2026-08-28 (10 deptos) · 2026-09-03 (11 deptos) | `zonas_conocidas`; set separado de CCPP exacto; cifra de clientes con inconsistencia de fuente (H8) |
| COMU-5569 | CCPP exacto | Comunicado sin tabla de zonas | 0 CCPP | Campaña 2026-12-10; fecha por zona pendiente | `zonas_pendiente`; 13.5K clientes, inicio 10-dic-2026 |

### 6.1 Cobertura por grupo (COMU-5514)

| Grupo | CCPP | Q clientes | PDF de zonas | Fecha de campaña | Fecha por zona | Estado |
|---|---:|---:|---|---|---|---|
| Grupo 1 | 16 | 1.7K | `APAGADP1.pdf` | 2026-09-11 | — | `pendiente_recurso_interno` |
| Grupo 2 | 201 | 15K | `APAGADO.pdf` | 2026-09-11 | — | `pendiente_recurso_interno` |
| Grupo 3 | 21 | 1.7K | `APAGADO2.pdf` | 2026-09-11 | — | `pendiente_recurso_interno` |

El reparto (16 + 201 + 21 = **238 CCPP**) se deriva del PDF de origen de cada zona; `grupo_por_key` cubre las 238 claves del runtime (check S20). El modelo «campaña + fecha por zona» ya existe en el sistema: COMU-5672 lo usa por departamento y COMU-5569 lo declara a nivel de campaña.

### 6.2 Reconciliación fuente ↔ runtime (COMU-5514)

| Métrica | Valor |
|---|---:|
| CCPP únicos en la fuente | 242 |
| CCPP exactos en runtime | 238 |
| Pérdidas (fuente sin clave en el sistema) | 8 |
| — justificadas de forma explícita | 5 |
| — con cobertura equivalente o alias documentado | 3 |
| Extras (clave del sistema sin fila literal en la fuente) | 4 |

Las 8 pérdidas y su resolución:

| Pérdida (clave de fuente) | Causa | Resolución |
|---|---|---|
| `010105:CHILIQUIN` | El PDF escribe el distrito en la columna CENTRO POBLADO | Pendiente de decisión (D4) |
| `060101:INCA CORRAL` | No existe en UBIGEO ni en `CCPP_DATA`; provincia `CAJABAMBA` | Pendiente de decisión (D5) |
| `160405:SANTA ROSA` | Distrito presente en UBIGEO inline con 0 caseríos | Pendiente de decisión (D5) |
| `030608:CCCEPATA` | Typo real del PDF (el sistema guarda `030608:OCCCEPATA`) | Alias documentado en la allowlist (H2) |
| `200206:CURILCAS` | PDF: PACAIPAMPA · GeoNames/runtime: AYABACA `200201` | Equivalente por grafía (1 letra) |
| `200206:PAPELILLO` | PDF: PACAIPAMPA · GeoNames/runtime: AYABACA `200201` | Equivalente por grafía (1 letra) |
| `200206:TOJAS` | PDF: PACAIPAMPA · GeoNames/runtime: AYABACA `200201` | Equivalente por CCPP |
| `200206:TULMAN DE MATALACAS` | PDF: PACAIPAMPA · GeoNames/runtime: AYABACA `200201` | Equivalente por CCPP |

De las 8 pérdidas, **5 quedan justificadas de forma explícita** en el barrido (3 pendientes de decisión + `CURILCAS` y `PAPELILLO` resueltas por d6/grafía equivalente); las 3 restantes están cubiertas por alias documentado o por equivalencia de CCPP.

Los **4 extras** son el contraparte exacto de las 4 divergencias d6: las claves `200201:CURRILCAS`, `200201:PAPELLILLO`, `200201:TOJAS` y `200201:TULMAN DE MATALACAS` del sistema, alcanzables desde el dropdown del popup.

Balance: **242 − 8 + 4 = 238** CCPP exactos en runtime, sin pérdidas injustificadas (check S3).

### 6.3 COMU-5569

Los cuatro archivos PDF cargados por el CEO son el mismo documento, md5 `393c127b3b6b4bb486d3d03be0850f97`: es el comunicado, no la tabla de zonas. El detalle reside en los recursos internos geniomovistar `id=178147` / `id=178146`.

El popup no alerta en zonas sin cobertura cargada. Esta regla de anti-fabricación se mantiene hasta recibir la tabla de centros poblados, fecha de apagado por departamento y fecha de inicio de SMS. La fecha de campaña declarada es 2026-12-10; sin tabla de zonas no se puede afirmar una fecha por zona.

---

## 7. Decisiones pendientes del CEO

| Decisión | Caso | Opciones | Recomendación | Estado |
|---|---|---|---|---|
| **D4** | `010105:CHILIQUIN` | A) mapear al CCPP GeoNames “Chiliquini”; B) cubrir todo el distrito CHILIQUIN | Confirmar con el emisor la semántica del PDF. Si significa todo el distrito, implementar cobertura distrital separada; no crear cobertura CCPP exacta por inferencia. | Pendiente |
| **D5** | `060101:INCA CORRAL` y `160405:SANTA ROSA` | A) incorporar el CCPP faltante; B) mantener sin cobertura; C) cubrir el distrito | Solicitar una fuente geográfica/campaña autoritativa. Corregir `CAJABAMBA` a `CAJAMARCA` y resolver SANTA ROSA con el catálogo de caseríos. Hasta entonces, no fabricar claves. | Pendiente |
| **D6** | Atribución PACAIPAMPA / AYABACA | A) conservar d6 runtime `200201`; B) migrar a d6 fuente `200206` | Mantener `200201`: es la atribución GeoNames y la clave alcanzable desde el dropdown. Registrar la divergencia de fuente y pedir confirmación antes de migrar. | Pendiente |
| **D7** | Cifra de clientes de la campaña IPT (COMU-5672) | A) mostrar 39.500 (declarado en el comunicado); B) mostrar 43.728 (suma por departamento) | Mantener la cifra declarada 39.500 hasta que el emisor aclare el origen de la diferencia (+10,7%); la inconsistencia es de la fuente, no del sistema (H8). | Pendiente |
| **D8** | Fechas programadas por zona de COMU-5514 (3 grupos · 238 CCPP) | A) mantener el comportamiento seguro actual (inicio de campaña etiquetado + aviso); B) cargar la fecha por grupo cuando llegue el detalle por departamento | **Owner: CEO · prioridad alta.** Falta el detalle por departamento (fecha programada de apagado + inicio de SMS) de los 3 recursos «Ver aquí» del COMU-5514 v00003 (geniomovistar, no accesibles desde este entorno). Mientras tanto el popup muestra el inicio de campaña etiquetado + aviso (comportamiento seguro) y el speech no afirma fecha (H9). | Pendiente |

No se implementará ninguna de estas decisiones sin aprobación explícita del CEO.

---

## 8. Verificación

| Suite / control | Resultado | Detalle |
|---|---|---|
| `test_popup.py` | PASS | 866/866 |
| `test_aggressive.py` | PASS | 238/238 fuzzy |
| `test_security.py` | CLEAN | Sin hallazgos |
| `test_codereview.py` | GOOD | Sin hallazgos |
| `test_browser_static.py` | PASS | Incluye `node --check` del JS inline |
| `test_ingest_comunicado_zonas.py` | PASS | 76/76 |
| `test_sweep_zonas.py` | PASS | 19 PASS · 1 WARN · 0 FAIL |
| Audit runtime | PASS | 0 issues |
| `node --check` | OK | Sintaxis del `<script>` inline |
| Harness runtime (panel) | PASS | 17/17 — HUIPOCA: `COMU-5514 · v00003 · Grupo 2 · inicio campaña: 11/09/2026` + aviso, sin fecha afirmada |
| `scripts/security_sweep_zonas.py` | PASS | **20 PASS · 1 WARN (S13/H8) · 0 FAIL** — checks S1–S21 |

### 8.1 Barrido de seguridad de datos (S1–S21)

Cruza tres universos — FUENTE (PDFs + OCR manual), RUNTIME (`index.html` inline) y DATA LAYER (`data/zonas_apagadas.json` + `data/ubigeo_consolidated.json`) — y exige que coincidan.

| Check | Qué verifica | Resultado | Detalle |
|---|---|---|---|
| S1 | Conteos coherentes (runtime / data layer / campaña) | PASS | runtime=238 · total_zones=238 · json=238 · campaña=238 |
| S2 | Toda clave del runtime existe en la fuente (anti-invención) | PASS | 238 claves · sin respaldo: ninguna |
| S3 | Toda fila de la fuente está cubierta o declarada como pendiente | PASS | sin cobertura=5 · justificadas=5 · injustificadas: ninguna |
| S4 | Todo d6 de las claves existe en el UBIGEO inline | PASS | 238 claves · inexistentes: ninguno |
| S5 | Todo distrito marcado existe en UBIGEO | PASS | 191 distritos · inexistentes: ninguno |
| S6 | Cada CCPP está en `CCPP_DATA[d6]` o en la allowlist documentada | PASS | 238 con respaldo · allowlist de 3 entradas · sin respaldo: ninguno |
| S7 | Ninguna clave «CCPP == DISTRITO» sin caserío real | PASS | contaminados: ninguno |
| S8 | Sin duplicados ni colisiones de fold | PASS | 238 claves · folds únicos=238 |
| S9 | Campañas coherentes (cobertura real o pendiente declarado) | PASS | COMU-5514 `zonas_conocidas` · COMU-5569 `zonas_pendiente` |
| S10 | Cada campaña tiene sus speechs y declaran su comunicado | PASS | todos coherentes |
| S11 | Zonas IPT: d6 válidos en UBIGEO | PASS | 1.798 zonas IPT · d6 inválidos: ninguno |
| S12 | Overlap IPT ↔ CCPP exactos | PASS | 168 distritos en ambos · cada zona atribuida a su campaña, sin doble conteo |
| S13 | Cifra de clientes IPT: suma por depto vs total declarado | **WARN** | 43.728 vs 39.500 (Δ +4.228 = +10,7%) — inconsistencia de la fuente (H8/D7) |
| S14 | Data layer (ubigeo) alineado con el runtime | PASS | inline=1.893 · data layer=1.893 · divergencias: ninguna |
| S15 | `zonas_apagadas.json` alineado con el runtime | PASS | delta runtime-json: vacío |
| S16 | Simulación: sin falsos positivos por coincidencia de nombre | PASS | 3 casos simulados · falsos positivos=0 |
| S17 | Trazabilidad de la fuente (md5 de los PDFs) | PASS | 3 PDFs con md5 + COMU-5569 con un único md5 |
| S18 | Sin secretos ni emails nuevos en el popup inline | PASS | secretos=0 · emails: ninguno |
| S19 | URLs externas del inline: solo las declaradas y documentadas | PASS | 8 URLs · no declaradas: ninguna |
| S20 | Todo CCPP del runtime tiene grupo de campaña (base de la fecha por zona) | PASS | 238 CCPP · sin grupo: ninguno |
| S21 | Sin fecha de zona falsa (grupos pendientes sin fecha por zona + panel etiquetado) | PASS | 3 grupos · `fecha_apagado_zona: null` en los pendientes · el panel etiqueta «apagado» vs «inicio campaña» |

---

## 9. Cómo reproducir

Desde `apps/cima-generador`, con los PDFs locales disponibles en `Recursos CIMA/Bases y Zonas de Apagado/`:

```bash
python3 scripts/audit_zonas_apagado.py
python3 scripts/security_sweep_zonas.py
python3 scripts/test_popup.py
python3 scripts/test_aggressive.py
python3 scripts/test_security.py
python3 scripts/test_codereview.py
python3 scripts/test_browser_static.py
python3 scripts/test_ingest_comunicado_zonas.py
python3 test_sweep_zonas.py
```

`test_browser_static.py` extrae el `<script>` inline y ejecuta `node --check`. El barrido regenera únicamente `data/apagados_tabla_completa.csv` y `data/audit_zonas_apagado.json`; el barrido de seguridad escribe `data/security_sweep_zonas.json`. Los tres son derivados y gitignored.

---

## 10. Lección

Un barrido contra la fuente es parte del gate, no una curiosidad documental.

- Los tests internos prueban consistencia; no prueban presencia de filas legítimas ausentes.
- La deriva de normalización produjo simultáneamente cuatro falsos negativos y falsos positivos recurrentes.
- El OCR de una fuente restringida se verifica dos veces antes de alimentar el sistema: la lectura única dejó 1 alias espurio (`CHUQICAHUANA`) y 1 CCPP de más en la fuente reconstruida.
- Cada release que cambie cobertura debe reconciliar el set fuente contra runtime, con equivalencias explícitas, allowlist mínima y decisiones pendientes visibles.
- Ante ambigüedad no verificable, el sistema debe refrainerse de alertar en vez de fabricar cobertura.
- Una fecha de campaña heredada y mostrada como fecha de zona es una afirmación falsa de precisión: etiquetar «inicio campaña» + aviso (y no afirmar «apagado: <fecha>») es el comportamiento seguro hasta recibir el dato por grupo/departamento (H9/D8).

---

*Auditor: Thavren — 2026-11-14 · barrido completo de zonas de apagado · v3.10.86 fecha por zona*

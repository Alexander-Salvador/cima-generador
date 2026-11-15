# CIMA — Tabla de zonas de apagado

**Fecha:** 2026-11-14  
Inventario completo de zonas apagadas en el popup CIMA: CCPP exactos COMU-5514, zonas IPT COMU-5672 y estado pendiente COMU-5569.

## 1. Por campaña

| Comunicado | Estado | Versión | Apagado | SMS | CCPP | IPT | Total | Deptos | Provincias | Distritos | Clientes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| COMU-5514 | zonas_conocidas | 00003 | 2026-09-11 (inicio de campaña; fecha por zona pendiente) | 2026-09-04 (campaña; inicio de SMS por zona pendiente) | 238 | 0 | 238 | 23 | 106 | 191 | 18.400 |
| COMU-5672 | zonas_conocidas | 00001 | 2026-08-28 (10 deptos) / 2026-09-03 (11 deptos) | 2026-08-17 (10 deptos) / 2026-08-18 (11 deptos) | 0 | 1.798 | 1.798 | 21 | 180 | 1.798 | 39.500 |
| COMU-5569 | zonas_pendiente | 00001 | 2026-12-10 (campaña; zonas pendientes) | Pendiente | 0 | 0 | 0 | 0 | 0 | 0 | 13.500 |

**Fuentes y cobertura:**

- **COMU-5514:** Bases de Apagado PDFs (APAGADO.pdf G2 · APAGADO2.pdf G3 · APAGADP1.pdf G1) + zonas_apagadas_manual.json. Cobertura: 238 CCPP exactos; 3 zonas fuente sin cobertura y 4 atribuciones d6 pendientes de decisión. Las 238 CCPP llevan su grupo de campaña (G1 16 · G2 201 · G3 21); la fecha por zona está pendiente (ver §2).
- **COMU-5672:** Apagado Zonas IPT.pdf. Cobertura: 1.798 zonas IPT en set separado del CCPP exacto. El PDF declara 39.500 clientes; la suma de notas CSV por depto es 43.728.
- **COMU-5569:** COMU-5569_Apagado de Estaciones Base en Centros Poblados_Noviembre.pdf (md5 393c127b3b6b4bb486d3d03be0850f97). Cobertura: 0 CCPP cargados; el popup no alerta. Los 4 PDF cargados son el mismo comunicado.

## 2. Fechas: qué sabemos y qué no

| Campaña | Qué fecha se conoce | Alcance real | Estado |
|---|---|---|---|
| COMU-5514 | 2026-09-11 | **Inicio de campaña**, no la fecha de cada zona. El detalle por departamento (fecha programada de apagado + inicio de SMS) está en los 3 recursos «Ver aquí» del comunicado v00003 (geniomovistar, no accesibles desde este entorno). | ⏳ Fecha por zona pendiente (D8) — el popup muestra «inicio campaña: 11/09/2026» + aviso |
| COMU-5672 | 2026-08-28 (10 deptos) · 2026-09-03 (11 deptos) | **Fecha por departamento** (21 departamentos) sobre 1.798 zonas IPT | ✅ Confirmada por departamento |
| COMU-5569 | 2026-12-10 | Fecha de campaña; la tabla de zonas aún no está cargada (0 CCPP) | ⏳ Zonas pendientes |

**Reparto por grupo (COMU-5514, derivado del PDF de origen de cada CCPP):**

| Grupo | CCPP | Q clientes | PDF de zonas |
|---|---:|---:|---|
| Grupo 1 | 16 | 1.7K | `APAGADP1.pdf` |
| Grupo 2 | 201 | 15K | `APAGADO.pdf` |
| Grupo 3 | 21 | 1.7K («Nuevo» en v00003) | `APAGADO2.pdf` |

El modelo «campaña + fecha por zona» ya existe en el sistema (COMU-5672 lo usa por departamento); en COMU-5514 falta el dato de los 3 recursos internos.

**D8 · Fechas programadas por zona de COMU-5514 — Owner: CEO · prioridad alta.** Falta el detalle por departamento (fecha programada de apagado + inicio de SMS) de los 3 recursos «Ver aquí» del COMU-5514 v00003. Mientras tanto el popup muestra el inicio de campaña etiquetado + aviso (comportamiento seguro) y el speech no afirma ninguna fecha (`comu5514_postventa_fecha_pendiente`).

## 3. Por departamento

| Departamento | CCPP exactos | Zonas IPT | Total zonas | Distritos con CCPP | Provincias con CCPP |
|---|---:|---:|---:|---:|---:|
| LIMA | 11 | 172 | 183 | 9 | 7 |
| ANCASH | 8 | 166 | 174 | 8 | 6 |
| CAJAMARCA | 18 | 127 | 145 | 15 | 8 |
| AYACUCHO | 18 | 124 | 142 | 15 | 8 |
| CUSCO | 26 | 116 | 142 | 23 | 10 |
| JUNIN | 16 | 124 | 140 | 10 | 6 |
| AREQUIPA | 8 | 109 | 117 | 7 | 6 |
| PUNO | 3 | 110 | 113 | 3 | 3 |
| HUANCAVELICA | 8 | 102 | 110 | 7 | 4 |
| AMAZONAS | 21 | 84 | 105 | 18 | 5 |
| APURIMAC | 14 | 85 | 99 | 12 | 6 |
| SAN MARTIN | 17 | 78 | 95 | 11 | 9 |
| HUANUCO | 5 | 84 | 89 | 5 | 4 |
| LA LIBERTAD | 5 | 83 | 88 | 4 | 2 |
| PIURA | 9 | 65 | 74 | 4 | 4 |
| LAMBAYEQUE | 7 | 38 | 45 | 6 | 3 |
| ICA | 1 | 43 | 44 | 1 | 1 |
| PASCO | 5 | 29 | 34 | 5 | 1 |
| TACNA | 4 | 28 | 32 | 3 | 3 |
| LORETO | 31 | 0 | 31 | 22 | 7 |
| MOQUEGUA | 1 | 20 | 21 | 1 | 1 |
| MADRE DE DIOS | 1 | 11 | 12 | 1 | 1 |
| UCAYALI | 1 | 0 | 1 | 1 | 1 |

## 4. Zonas sin cobertura

| Key | Departamento / provincia / distrito | CCPP | Estado | Por qué no está | Qué se necesita | Recomendación |
|---|---|---|---|---|---|---|
| 010105:CHILIQUIN | AMAZONAS / CHACHAPOYAS / CHILIQUIN | CHILIQUIN | ambiguo_centro_poblado_igual_distrito | El PDF escribe el nombre del distrito en la columna CENTRO POBLADO; GeoNames lista "Chiliquini". | Confirmación del emisor sobre si la fila representa el caserío Chiliquini o todo el distrito. | Confirmar con el emisor. Si significa todo el distrito, cubrirlo distritalmente; no fabricar cobertura CCPP exacta por inferencia. |
| 060101:INCA CORRAL | CAJAMARCA / CAJABAMBA / CAJAMARCA | INCA CORRAL | ccpp_no_en_caserios | No existe en UBIGEO ni CCPP_DATA; el PDF escribe la provincia "CAJABAMBA" frente a "CAJAMARCA". | Fuente geográfica o de campaña autoritativa; confirmar la provincia e incorporar el CCPP. | Solicitar la fuente autoritativa, corregir CAJABAMBA a CAJAMARCA y mantener sin cobertura hasta que el CCPP sea verificable. |
| 160405:SANTA ROSA | LORETO / MARISCAL RAMON CASTILLA / SANTA ROSA DE LORETO | SANTA ROSA | ccpp_no_en_caserios | El distrito existe en el UBIGEO inline, pero tiene 0 caseríos en CCPP_DATA. | Catálogo de caseríos de SANTA ROSA DE LORETO. | Solicitar el catálogo de caseríos e incorporar cobertura solo con un CCPP verificable; mantener sin cobertura mientras tanto. |

## 5. Divergencias d6

| CCPP | Departamento / provincia / distrito (PDF) | d6 PDF | d6 sistema | Nota |
|---|---|---:|---:|---|
| CURILCAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. Literal runtime: CURRILCAS. |
| PAPELILLO | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. Literal runtime: PAPELLILLO. |
| TOJAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. |
| TULMAN DE MATALACAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. |

## 6. Verificación

| Suite / control | Resultado | Detalle |
|---|---|---|
| test_popup.py | PASS | 866/866 |
| test_aggressive.py | PASS | 238/238 fuzzy |
| test_security.py | CLEAN | Sin hallazgos |
| test_codereview.py | GOOD | Sin hallazgos |
| test_browser_static.py | PASS | Incluye node --check del JS inline |
| test_ingest_comunicado_zonas.py | PASS | 76/76 |
| test_sweep_zonas.py | PASS | 19 PASS · 1 WARN · 0 FAIL |
| Audit runtime | PASS | 0 issues |
| node --check | OK | Sintaxis del script inline |
| Harness runtime (panel) | PASS | 17/17 — HUIPOCA: `COMU-5514 · v00003 · Grupo 2 · inicio campaña: 11/09/2026` + aviso, sin fecha afirmada |

[Abrir informe completo](AUDITORIA_ZONAS_APAGADO_2026-11-14.md) — ruta: `docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md`.

## 7. Cómo se reproduce

```bash
cd apps/cima-generador
python3 scripts/export_tabla_zonas.py
python3 scripts/audit_zonas_apagado.py
```

## Notas de consistencia

- El CSV fuente contiene 2.081 filas y 28 columnas; el encargo indica 22 columnas. El export usa las 28 columnas reales.
- `audit_zonas_apagado.json` lista 5 filas sin cobertura literal; 2 son la fuente de las 4 divergencias d6 y 3 requieren decisión funcional.
- COMU-5672: el runtime declara 39.500 clientes; la suma de notas CSV por depto es 43.728. Se conserva la cifra declarada.
- COMU-5672: la fecha es **por departamento** — 2026-08-28 en 10 departamentos y 2026-09-03 en 11 (verificado contra `ipt_metadata_base` del runtime). COMU-5514: 2026-09-11 es el inicio de campaña, no una fecha de zona (H9/D8).
- COMU-5514: las 238 CCPP declaran su grupo de campaña (`grupo_por_key`, G1 16 · G2 201 · G3 21); `estado_fecha = pendiente_recurso_interno` en los 3 grupos.

*Generado por Thavren — 2026-11-14 · barrido completo de zonas de apagado · v3.10.86 fecha por zona*

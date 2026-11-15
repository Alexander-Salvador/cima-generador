# CIMA — Tabla de zonas de apagado

**Fecha:** 2026-11-14  
Inventario completo de zonas apagadas en el popup CIMA: CCPP exactos COMU-5514, zonas IPT COMU-5672 y estado pendiente COMU-5569.

## 1. Por campaña

| Comunicado | Estado | Versión | Apagado | SMS | CCPP | IPT | Total | Deptos | Provincias | Distritos | Clientes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| COMU-5514 | zonas_conocidas | 00003 | 2026-09-11 | 2026-09-04 | 238 | 0 | 238 | 23 | 106 | 191 | 18.400 |
| COMU-5672 | zonas_conocidas | 00001 | 2026-08-28 (10 deptos) / 2026-09-03 (11 deptos) | 2026-08-17 (10 deptos) / 2026-08-18 (11 deptos) | 0 | 1.798 | 1.798 | 21 | 180 | 1.798 | 39.500 |
| COMU-5569 | zonas_pendiente | 00001 | 2026-12-10 | Pendiente | 0 | 0 | 0 | 0 | 0 | 0 | 13.500 |

**Fuentes y cobertura:**

- **COMU-5514:** Bases de Apagado PDFs (APAGADO.pdf G2 · APAGADO2.pdf G3 · APAGADP1.pdf G1) + zonas_apagadas_manual.json. Cobertura: 238 CCPP exactos; 3 zonas fuente sin cobertura y 4 atribuciones d6 pendientes de decisión.
- **COMU-5672:** Apagado Zonas IPT.pdf. Cobertura: 1.798 zonas IPT en set separado del CCPP exacto. El PDF declara 39.500 clientes; la suma de notas CSV por depto es 43.728.
- **COMU-5569:** COMU-5569_Apagado de Estaciones Base en Centros Poblados_Noviembre.pdf (md5 393c127b3b6b4bb486d3d03be0850f97). Cobertura: 0 CCPP cargados; el popup no alerta. Los 4 PDF cargados son el mismo comunicado.

## 2. Por departamento

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

## 3. Zonas sin cobertura

| Key | Departamento / provincia / distrito | CCPP | Estado | Por qué no está | Qué se necesita | Recomendación |
|---|---|---|---|---|---|---|
| 010105:CHILIQUIN | AMAZONAS / CHACHAPOYAS / CHILIQUIN | CHILIQUIN | ambiguo_centro_poblado_igual_distrito | El PDF escribe el nombre del distrito en la columna CENTRO POBLADO; GeoNames lista "Chiliquini". | Confirmación del emisor sobre si la fila representa el caserío Chiliquini o todo el distrito. | Confirmar con el emisor. Si significa todo el distrito, cubrirlo distritalmente; no fabricar cobertura CCPP exacta por inferencia. |
| 060101:INCA CORRAL | CAJAMARCA / CAJABAMBA / CAJAMARCA | INCA CORRAL | ccpp_no_en_caserios | No existe en UBIGEO ni CCPP_DATA; el PDF escribe la provincia "CAJABAMBA" frente a "CAJAMARCA". | Fuente geográfica o de campaña autoritativa; confirmar la provincia e incorporar el CCPP. | Solicitar la fuente autoritativa, corregir CAJABAMBA a CAJAMARCA y mantener sin cobertura hasta que el CCPP sea verificable. |
| 160405:SANTA ROSA | LORETO / MARISCAL RAMON CASTILLA / SANTA ROSA DE LORETO | SANTA ROSA | ccpp_no_en_caserios | El distrito existe en el UBIGEO inline, pero tiene 0 caseríos en CCPP_DATA. | Catálogo de caseríos de SANTA ROSA DE LORETO. | Solicitar el catálogo de caseríos e incorporar cobertura solo con un CCPP verificable; mantener sin cobertura mientras tanto. |

## 4. Divergencias d6

| CCPP | Departamento / provincia / distrito (PDF) | d6 PDF | d6 sistema | Nota |
|---|---|---:|---:|---|
| CURILCAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. Literal runtime: CURRILCAS. |
| PAPELILLO | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. Literal runtime: PAPELLILLO. |
| TOJAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. |
| TULMAN DE MATALACAS | PIURA / AYABACA / PACAIPAMPA | 200206 | 200201 | PDF: d6 200206; GeoNames/runtime: d6 200201. La clave runtime es alcanzable; no se duplicó la clave PDF. |

## 5. Verificación

| Suite / control | Resultado | Detalle |
|---|---|---|
| test_popup.py | PASS | 844/844 |
| test_aggressive.py | PASS | 238/238 fuzzy |
| test_security.py | CLEAN | Sin hallazgos |
| test_codereview.py | GOOD | Sin hallazgos |
| test_browser_static.py | PASS | Incluye node --check del JS inline |
| test_ingest_comunicado_zonas.py | PASS | 76/76 |
| Audit runtime | PASS | 0 issues |
| node --check | OK | Sintaxis del script inline |

[Abrir informe completo](AUDITORIA_ZONAS_APAGADO_2026-11-14.md) — ruta: `docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md`.

## 6. Cómo se reproduce

```bash
cd apps/cima-generador
python3 scripts/export_tabla_zonas.py
python3 scripts/audit_zonas_apagado.py
```

## Notas de consistencia

- El CSV fuente contiene 2.081 filas y 23 columnas; el encargo indica 22 columnas. El export usa las 23 columnas reales.
- `audit_zonas_apagado.json` lista 5 filas sin cobertura literal; 2 son la fuente de las 4 divergencias d6 y 3 requieren decisión funcional.
- COMU-5672: el runtime declara 39.500 clientes; la suma de notas CSV por depto es 43.728. Se conserva la cifra declarada.

*Generado por Thavren — 2026-11-14 · barrido completo de zonas de apagado*

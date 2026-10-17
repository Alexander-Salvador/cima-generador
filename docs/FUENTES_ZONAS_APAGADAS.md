# CIMA — Fuentes de Datos y Trazabilidad

**Versión:** v3.10.0  
**Fecha:** 2026-10-17  
**Auditor:** Thavren (Platform Engineering)

---

## 1. UBIGEO (estructura departamental)

| Campo | Valor |
|---|---|
| **Fuente** | RitchieRD/ubigeos-peru-data (GitHub) |
| **URL** | https://github.com/RitchieRD/ubigeos-peru-data |
| **Licencia** | ODbL (Open Database License) |
| **Cobertura** | 25 departamentos, 196 provincias, 1,892 distritos |
| **Validación** | Coherente con Wikipedia, INEI Censos |
| **Gap conocido** | Falta "Santa Rosa de Loreto" (Ley 32403, 2025-07-03) |

**Notas:**
- INEI no publica una tabla única consolidada de distritos al 2026.
- Wikipedia ES da 1,891 / 1,893 (discrepancia). RitchieRD da 1,892 validado contra CSV crudo.
- Última creación legal de distrito: **Santa Rosa de Loreto** (3-jul-2025), pendiente de incorporación oficial al compendio INEI 2026.

---

## 2. CCPP (centros poblados / caseríos)

| Campo | Valor |
|---|---|
| **Fuente** | GeoNames (vía RitchieRD) |
| **URL** | http://www.geonames.org/ |
| **Licencia** | CC BY 4.0 |
| **Cobertura** | 43,491 caseríos en 1,851 distritos (de 1,892) |
| **Cifra oficial 2026** | **3,010 CP con municipalidad delegada** (INEI 2022) |
| **Gap conocido** | 41 distritos sin CCPP. Ver `data/known_data_gaps.json` |

**Notas críticas:**
- La cifra 43,491 CCPP corresponde al **Censo INEI 2007**. Es histórica y obsoleta para 2026.
- INEI Censos 2017 publicó datos de CCPP, pero el conteo total ronda 120k+ caseríos censados.
- La cifra operativa 2022-2025 es 3,010 CP con municipalidad delegada.
- **Decisión CIMA:** mantener 43,491 (GeoNames) como base operativa por cobertura amplia. Marcar como "Censo 2007" en metadata.
- INEI Stats Portal en rediseño (404 en rutas de descarga directa).

---

## 3. Zonas Apagadas COMU-5514 (Estaciones Base)

| Campo | Valor |
|---|---|
| **Fuente** | PDFs internos Movistar "Apagado de Estaciones Base en Centros Poblados" (Grupo 1, 2, 3) |
| **URL** | No pública (distribuida internamente) |
| **Licencia** | Prohibida su reproducción (per nota al pie PDFs) |
| **Cobertura** | 200 CCPP exactos en 22 departamentos |
| **Actualizado** | Regenerado 2026-10-15 desde PDF |
| **Extracción** | pdfplumber (texto) + OCR visual modelo (PDFs escaneados) |

**Departamentos con cobertura:**
- Top: Loreto (29), Cusco (20), Cajamarca (16), Amazonas (15), Ayacucho (15), San Martín (15)
- Cobertura completa: 22/25 departamentos
- **Sin cobertura:** CALLAO (urbano), TUMBES (anomalía), UCAYALI (1/19 — anomalía)

**Anomalías documentadas (v3.10.0):**

| Depto | Estaciones | Esperado | Razón |
|---|---|---|---|
| CALLAO | 0/7 | 0 | Urbano, sin CCPP rurales. **Consistente.** |
| TUMBES | 0/13 | ¿? | Anomalía — costa con ruralidad. Requiere confirmación Movistar. |
| LORETO | 22/55 | parcial | Amazonia profunda, 33 dist sin estación. Limitación logística, no gap claro. |
| UCAYALI | 1/19 | ¿? | Solo Padre Abad documentado. Cobertura Movistar rural limitada. |

---

## 4. Zonas Apagadas COMU-5672 (Red 2G IPT)

| Campo | Valor |
|---|---|
| **Fuente** | PDF interno Movistar "Apagado Zonas IPT.pdf" |
| **URL** | No pública |
| **Licencia** | Interno Movistar |
| **Cobertura** | 1,798 distritos en 21 departamentos (amplitud: departamento) |
| **Actualizado** | Regenerado 2026-10-15 |
| **Estado oficial** | **NO oficial** |
| **Clientes declarados** | 39,500 (según PDF, no verificado OSIPTEL) |

**Hallazgo crítico (v3.10.0):**

El MTC y OSIPTEL **NO han confirmado un cronograma público de apagado 2G en Perú** al 2026-10-17:

| Fecha | Organismo | Comunicado | URL |
|---|---|---|---|
| 2024-11-18 | OSIPTEL | Caída internet móvil/Fijo (incidencia temporal) | https://www.gob.pe/institucion/osiptel/noticias/1060417 |
| 2025-10-21 | MTC | Mesa técnica 2G — operadores presentan plan | https://mobiletime.la/noticias/21/10/2025/peru-apagon-de-la-red-2g/ |
| 2025-10-22 | MTC | Confirmación planificación apagado 2G | https://latambytes.com/mtc-de-peru-planifica-el-apagado-de-la-red-2g-en-el-pais/ |
| 2025-12-17 | MTC | Adjudicación 5G (paralelo) | https://mobiletime.la/noticias/17/12/2025/peru-formaliza-despliegue-de-5g/ |
| 2026-02-16 | OSIPTEL | Estaciones 2G Movistar NO disponibles (incidencia temporal, no apagado) | https://www.gob.pe/institucion/osiptel/noticias/1354758 |
| 2026-03-17 | OSIPTEL | Resolución 000024-2026 (BAJA de líneas por FRAUDE, no apagado 2G) | https://www.infobae.com/peru/2026/03/17/... |

**Conclusión:** El apagado 2G en Perú está **planificado, no ejecutado**. Los datos del popup COMU-5672 son **información interna Movistar** que debe etiquetarse como tal para que el supervisor no ejecute flujos de baja/transferencia sobre datos no verificados públicamente.

**Decisión v3.10.0:** agregar disclaimer obligatorio en el tooltip cuando se muestra un speech COMU-5672 (ipt_distrito), indicando:
- Fuente: información interna Movistar
- MTC/OSIPTEL aún no confirman cronograma público
- Verificar con el cliente antes de ejecutar baja/transferencia

---

## 5. Speeches de atención

| Speech | COMU | Aplica a | Verificado |
|---|---|---|---|
| `comu5514_visor_unico` | COMU-5514 | Visor — estación base | ✅ oficial (PDF Movistar) |
| `comu5514_postventa_marcaje` | COMU-5514 | Postventa con marcaje | ✅ oficial (PDF Movistar) |
| `ipt_visor_4g` | COMU-5672 | Visor 4G, config APN | ⚠️ interna (PDF IPT) |
| `ipt_visor_sin_4g` | COMU-5672 | Visor sin 4G, renovación | ⚠️ interna (PDF IPT) |
| `ipt_postventa_4g` | COMU-5672 | Postventa 4G | ⚠️ interna (PDF IPT) |
| `ipt_postventa_sin_4g` | COMU-5672 | Postventa sin 4G | ⚠️ interna (PDF IPT) |

**Audit check 7 (v3.9.15+):** cada speech declara su COMU en metadata. Validación automática al startup detecta contaminación cruzada.

---

## 6. Cifras declaradas vs realidad verificable

| Cifra en popup | Realidad oficial |
|---|---|
| 39,500 clientes COMU-5672 | No publicada por Movistar/OSIPTEL. MTC solo dice "87% usuarios 2G en zonas rurales" (oct 2025). |
| 1,700 + 15,000 + 1,700 = 18,400 clientes COMU-5514 | No publicada oficialmente. Cifra interna PDFs Movistar. |
| Fechas apagado 2026-08-28 / 2026-09-03 | Sin cronograma público confirmado por MTC/OSIPTEL. |
| 64,000 antenas 4G en Perú (feb 2026) | ✅ verificable (MTC). |
| 25 deptos, 196 prov, 1,892 dist | ✅ alineado con RitchieRD. |
| 43,491 CCPP | ⚠️ Cifra INEI Censo 2007 (obsoleta). |

---

## 7. Decisiones arquitectónicas v3.10.0

1. **Separación semántica de los 3 niveles de check** (ccpp_exacto / ipt_distrito / distrito_of).
2. **Disclaimer obligatorio para COMU-5672** en tooltip, indicando fuente interna.
3. **Documentación de anomalías departamentales** (CALLAO, TUMBES, LORETO, UCAYALI).
4. **Trailers de fuente** en metadata `ZONAS_APAGADAS_INTEL.fuentes_datos`.
5. **Tests de auditoría ampliados** (checks 8, 9, 10, 11).
6. **Gaps de CCPP** documentados en `data/known_data_gaps.json`.

---

## 8. Acciones pendientes

| # | Acción | Owner | Prioridad |
|---|---|---|---|
| 1 | Confirmar con Movistar TUMBES (¿planes apagado oclusión genuina?) | CEO + Movistar | media |
| 2 | Confirmar LORETO 33 dist sin estación | CEO + Movistar | media |
| 3 | Confirmar UCAYALI 18 dist sin estación | CEO + Movistar | media |
| 4 | Descargar INEI Marco Geoestadístico 2018+ para 12 dist rurales nuevos | Thavren | alta |
| 5 | Scrappear OSM Overpass para 23 dist rurales legacy | Thavren | alta |
| 6 | Inyectar 2 dist con CCPP ya encontrados (UNION PROGRESO, UNION ASHÁNINKA) | Thavren | media |
| 7 | Validar Santa Rosa de Loreto (Ley 32403) en INEI compendio 2026 | Thavren | baja (esperar) |
| 8 | Monitorear MTC/OSIPTEL por cronograma público de apagado 2G | Thavren | continua |

---

*Mantenedor: Thavren — auditado 2026-10-17*

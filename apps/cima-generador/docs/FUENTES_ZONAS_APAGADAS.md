# CIMA — Fuentes de Datos y Trazabilidad

**Versión:** v3.10.83  
**Fecha:** 2026-11-14  
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
| `comu5569_visor_unico` | COMU-5569 | Visor — estación base | ✅ oficial (PDF Movistar COMU-5569 v00001) |
| `comu5569_postventa_marcaje` | COMU-5569 | Postventa con marcaje | ✅ oficial (PDF Movistar COMU-5569 v00001) |
| `ipt_visor_4g` | COMU-5672 | Visor 4G, config APN | ⚠️ interna (PDF IPT) |
| `ipt_visor_sin_4g` | COMU-5672 | Visor sin 4G, renovación | ⚠️ interna (PDF IPT) |
| `ipt_postventa_4g` | COMU-5672 | Postventa 4G | ⚠️ interna (PDF IPT) |
| `ipt_postventa_sin_4g` | COMU-5672 | Postventa sin 4G | ⚠️ interna (PDF IPT) |

**Audit check 7 (v3.9.15+):** cada speech declara su COMU en metadata. Validación automática al startup detecta contaminación cruzada.

**Nota COMU-5569 (2026-11-14):** los speeches `comu5569_visor_unico` y `comu5569_postventa_marcaje` aplican **únicamente cuando el CCPP consultado esté dentro de la cobertura de la campaña COMU-5569**. Esa cobertura todavía no está cargada (ver §9 — *PENDIENTE: tabla de zonas*), por lo que a la fecha ambos speeches están **documentados pero no activables por marcaje**.

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
| 9 | Obtener tabla de zonas COMU-5569 (recurso interno geniomovistar id=178147/178146 o export del CEO) — **bloquea la cobertura de zonas de la campaña diciembre 2026** | CEO | alta |

---

## 9. COMU-5569 — Apagado de Estaciones Base en Centros Poblados (campaña diciembre 2026)

| Campo | Valor |
|---|---|
| **Código** | COMU-5569 |
| **Título** | Apagado de Estaciones Base en Centros Poblados |
| **Versión** | 00001 |
| **Inicio de apagado** | Desde el **10 de diciembre de 2026** |
| **Q de clientes** | 13.5K |
| **Canales** | Call Convergente · Tiendas · Soluciones · Retenciones · Embajador · Canales escritos · Priority |
| **Alcance** | Centros poblados (estaciones base en zona rural) |
| **Fuente** | PDF interno Movistar — `Recursos CIMA/Bases y Zonas de Apagado/` (carga del CEO, 2026-11-14) |
| **MD5** | `393c127b3b6b4bb486d3d03be0850f97` (483,414 bytes) |
| **Cobertura de zonas** | ⛔ **PENDIENTE** — el PDF **no** incluye la tabla de zonas |

**Estructura del comunicado (1 página, texto + imágenes):** 1 Descripción · 2 Características · 3 Comunicación (SMS) · 4 Visor · 5 Escenarios atención Postventa · 6 Reclamos · 7 Baja del servicio · 8 Preguntas Frecuentes.

> **Las 4 copias son el mismo documento; no son 4 fuentes distintas.**
> `COMU-5569_...pdf`, `... (1).pdf`, `... (2).pdf`, `... (3).pdf` → md5 idéntico `393c127b3b6b4bb486d3d03be0850f97` (483,414 bytes). Son descargas repetidas del mismo archivo (Windows: sufijo `(1)/(2)/(3)`). Contar 4 fuentes sería un error de trazabilidad.

**Nota de versionado:** `Recursos CIMA/` está en `.gitignore` — los PDFs del CEO **no** están versionados en el repo. La trazabilidad de esta campaña depende de MD5 + transcript (§9) + el export de zonas pendiente.

### 9.1 Texto oficial transcrito (literal del PDF)

**Sección 3 — Comunicación (SMS 1):**
> "A partir del [dd/mm] no contaras con señal Movistar en la zona [localidad]. En otras localidades el servicio opera normal. Lamentamos las molestias."

**Sección 4 — Visor:**
> "APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS: Queremos informarte que, a partir del [día] de [mes], el servicio Movistar dejará de funcionar en la zona [nombre de la zona]. En otras zonas, el servicio seguirá disponible con normalidad. Lamentamos las molestias y agradecemos tu comprensión."

**Nota asociada:**
> "[REVISAR COMUNICADO APAGADO DE ESTACIONES BASE EN CENTROS POBLADOS]"

**Sección 5 — Escenarios de atención Postventa:**

| Escenario | Situación | Flujo |
|---|---|---|
| 1 | Cliente consulta por comunicación recibida | Identificar marcaje en visor → aplicar speech |
| 2 | Cliente se comunica porque no cuenta con señal en su zonal | Identificar marcaje: **con marcaje** → aplicar speech · **sin marcaje** → flujo regular de atención de avería móvil (Call Center: VOZ / CCEE / DATOS) |

**Sección 6 — Reclamos:**

| Campo | Valor |
|---|---|
| **Detalle** | Otros Reclamos Soluciones Móvil |
| **Motivo** | Calidad |
| **Sub Motivo** | Problemas de cobertura y señal |
| **Observaciones** | "Colocar en observaciones que se trata de cliente afectado por el Apagado de zona rural" |
| **Canales Escritos** | https://www.movistar.com.pe/atencion-al-cliente/reclamos |

**Sección 7 — Baja del servicio:**

| Canal | Ruta |
|---|---|
| **Call Center 1ra línea** | Pool Retenciones |
| **Tiendas** | Flujo regular de retenciones |
| **Canales Escritos** | 104 o tienda |

**Sección 8 — Preguntas Frecuentes:**

| FAQ | Respuesta oficial |
|---|---|
| ¿Ya no contaré con servicio en mi línea? | No tendrá señal Movistar en la zona informada; sí en otras zonas con señal. |
| ¿Puedo navegar con Wifi? | Sí, si el equipo tiene opción Wifi. |

### 9.2 Detalle de zonas (enlaces internos)

La tabla de zonas no está en el PDF: está detrás del botón **"Revisa las zonas"**, que apunta a recursos internos de la plataforma corporativa:

| Sección del PDF | Enlace |
|---|---|
| 1 — Descripción | https://geniomovistar.com.pe/mod/resource/view.php?id=178147 |
| 2 — Características | https://geniomovistar.com.pe/mod/resource/view.php?id=178146 |

Ambos son **NO accesibles desde este entorno** (TLS/red corporativa, sin login).

### 9.3 ⛔ PENDIENTE: tabla de zonas

**Falta exactamente lo siguiente:**

1. **CCPP / centros poblados por departamento** — la lista exacta de zonas incluida en la campaña COMU-5569.
2. **Fecha de apagado por departamento** — el PDF solo declara la fecha global de inicio (10-dic-2026).
3. **Fecha de inicio de SMS por departamento** — el PDF solo define la plantilla SMS con el placeholder `[dd/mm]`.

**Por qué:** el PDF cargado es **el comunicado**, no el detalle. El detalle vive en los enlaces internos `geniomovistar.com.pe` (id=178147 / id=178146), que requieren red corporativa y login, no disponibles desde este entorno. Sin esos tres datos no se puede generar la cobertura `ccpp_exacto` de la campaña ni activar los speeches COMU-5569 por marcaje.

**Desbloqueo:** export del CEO (mismo formato que COMU-5514) o descarga autenticada de los dos enlaces. Ver §8, acción #9.

### 9.4 Estado del sistema al momento de redactar

| Elemento | Valor |
|---|---|
| **Popup** | v3.10.82 → **v3.10.83 (en curso)** |
| **Asserts** | 750 PASS (base) |
| **ccpp_keys COMU-5514** | 234 (campaña septiembre 2026 — apagado 2026-09-11, SMS 2026-09-04, 18.4K clientes en 3 grupos) |
| **Zonas IPT COMU-5672** | 1.798 |

> ⚠️ **Discrepancia observada en auditoría:** al momento de redactar, `data/zonas_apagadas.json` reporta **239** zonas/ccpp_keys de COMU-5514 (y su campo `fuente` dice "239 CCPPs COMU-5514"), no 234. El dato 234 corresponde al snapshot del CEO; hay una regeneración de datos en curso en paralelo. Reconciliar antes de dar por cerrada la cifra en §3 (que aún dice 200 CCPP / 22 departamentos, v3.10.0).

---

## 10. Ingesta de las zonas cuando llegue el detalle

Comando de ingesta (con la tabla de zonas ya disponible). **No ejecutar sin el detalle de zonas y sin OK del CEO.**

**Dry-run (preview, no escribe nada):**
```bash
python3 apps/cima-generador/scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 --source <TABLA> --modo union --dry-run
```

**Variante de aplicación en línea (`--apply-inline`, escribe los CCPP):**
```bash
python3 apps/cima-generador/scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 --source <TABLA> --modo union --apply-inline
```

**Modos de acumulación:**

| Modo | Comportamiento | Uso |
|---|---|---|
| `--modo union` | **Acumula** zonas: agrega las de COMU-5569 **sin borrar** las de COMU-5514 (ni las manuales). | ✅ Default. Correcto para la campaña diciembre 2026. |
| `--modo replace` | **Reemplaza** el set completo de zonas de la fuente. Borra lo anterior. | ⛔ **Solo con confirmación explícita del CEO**, por escrito en la sesión. |

`<TABLA>` = ruta al archivo con la tabla de zonas (el export del CEO o el HTML de los enlaces id=178147 / id=178146 descargado con login).

**Precondiciones:**
- Tabla de zonas con CCPP por departamento, fecha de apagado y fecha de inicio de SMS (ver §9.3).
- `--dry-run` ejecutado y revisado antes de cualquier `--apply-inline`.
- Verificar que el conteo resultante concilia con `ccpp_keys` de COMU-5514 (234/239 — reconciliar §9.4).

---

*Mantenedor: Thavren — auditado 2026-10-17*
*Actualizado por Thavren — 2026-11-14 (campaña COMU-5569)*

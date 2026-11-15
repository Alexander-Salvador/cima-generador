# CIMA — Lista de Pendientes (Estado verificado 2026-11-14)

**Fecha verificación:** 2026-11-14
**Branch:** `develop` (working tree limpio) · local **4+ commits adelante** de `origin/develop` (sin push)
**HEAD local:** `0ba89ca` + **v3.10.84** (barrido completo de zonas) · `origin/develop` @ `c163567`
**Popup local:** `apps/cima-generador/` → **v3.10.84** (commit local, **NO publicada**)
**Popup LIVE (Pages):** **v3.10.82** (`cda031b`) → https://alexander-salvador.github.io/cima-generador/
**Tests:** `test_popup.py` **844/844 PASS** + aggressive / security / codereview / browser_static / ingest_comunicado_zonas (76/76)
**Zonas:** **238 CCPP** (COMU-5514, 191 distritos) + **1.798 IPT** (COMU-5672) + COMU-5569 pendiente (0 CCPP)
**DIRECTIVA ACTIVA:** seguir en **LOCAL**. No tocar remoto sin OK explícito del CEO.

---

## 🗺️ BARRIDO COMPLETO DE ZONAS (v3.10.84) — qué se entregó

| Artefacto | Contenido |
|---|---|
| `apps/cima-generador/data/apagados_tabla_completa.csv` | **2.081 filas × 22 columnas** — 282 COMU-5514 (247 CCPP únicos + 35 duplicados) · 1.798 COMU-5672 · 1 COMU-5569 pendiente. Sin pérdida de campos |
| `apps/cima-generador/data/audit_zonas_apagado.json` | Detalle programático (pérdidas · extras · precisión · huérfanos · data layer) |
| `apps/cima-generador/scripts/audit_zonas_apagado.py` | Barrido reproducible: `python3 scripts/audit_zonas_apagado.py` |
| `apps/cima-generador/docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md` | Informe H1–H7 + decisiones D4–D6 |

**Barrido de verdad:** 234 → **238 CCPP** (4 legítimas que NO alertaban, recuperadas) ·
188 → **191 distritos** · audit runtime **3 issues → 0** · data layer alineado (0 divergencias).

---

## 🔴 BLOQUEO — falta la tabla de zonas de COMU-5569 (campaña 10-dic-2026)

**Qué se cargó:** 4 PDFs en `Recursos CIMA/Bases y Zonas de Apagado/` → los 4 son
**el mismo archivo** (md5 `393c127b3b6b4bb486d3d03be0850f97`): el comunicado
**COMU-5569 v00001** (inicio 10-dic-2026, 13.5K clientes). **No trae la tabla de
zonas** (ni CCPP, ni centros poblados, ni fechas por departamento).

**Dónde está el detalle:** botón "Revisa las zonas" del propio PDF →
`https://geniomovistar.com.pe/mod/resource/view.php?id=178147` (y `id=178146`).
Recurso interno corporativo: no accesible desde este entorno (TLS/red + login).

**Qué necesito del CEO (cualquiera de estos):**
1. Export/print del recurso `id=178147` (PDF o tabla de centros poblados por depto).
2. El PDF/tabla original de zonas de la campaña diciembre 2026.
3. Acceso/red habilitada al portal geniomovistar desde el entorno de build.

**Qué YA está listo (no requiere el dato):**
- Campaña COMU-5569 registrada (estado `zonas_pendiente`, 0 CCPP) → el popup
  **no alerta** en zonas sin cobertura: cero riesgo de speech a un cliente que
  no está apagado.
- Speechs oficiales (visor §4, postventa §5), flujo postventa (2 escenarios, SMS,
  reclamo, baja, 2 FAQs).
- Comando de ingesta listo (1 comando cuando llegue la tabla):
  ```bash
  python3 apps/cima-generador/scripts/ingest_comunicado_zonas.py \
    --comunicado COMU-5569 --source "<TABLA>" --modo union --apply-inline
  ```
  (dedupe md5 · fuzzy UBIGEO · backup atómico · idempotente · dry-run por defecto)

**Decisión a futuro (CEO):** ¿COMU-5569 **acumula** zonas a las 234 de COMU-5514
(`--modo union`, default, recomendado) o **las reemplaza** (`--modo replace`, solo
con OK explícito)? El comunicado dice 13.5K clientes vs 18.4K de COMU-5514 → no se
puede asumir que sean las mismas zonas.

---

## 🗓️ Contexto de la última sesión (2026-11-06 → 2026-11-07)

6 releases cerradas y **publicadas** (v3.10.77 → v3.10.82):

| Release | Qué cambió |
|---|---|
| v3.10.77 | Contador `9/9` del input TELÉFONO desalineado → overlay CSS Grid |
| v3.10.78 | ETIQUETAS reconciliadas con el Excel (26 → 38). Split de la mega-fila comercial en 4. **3 bugs corregidos**: `PLAN MENOR` (skill `DOWNSELL_MOVIL_TX`), `VENTAS` (`CROSS_MOVIL_FIJA_TX`), `KONECTA` → `TECNICO_MOVIL_TX`. Búsqueda title-first + alias + singular/plural |
| v3.10.79 | Misma lógica en M.TOTAL + STANDALONE → **39** opciones. Gap real cerrado: `TRASLADO / MUDANZA` faltaba en MT |
| v3.10.80 | Auditoría integral (8 ejes, 0 errores) + `ETIQUETAS_BASELINE` anti-regresión en el gate. Corregida clause de validación perdida en `AVERÍA HOGAR` |
| v3.10.81 | Chips por segmento (**MOVIL · M.TOTAL · STANDALONE**) + fix del estado "trancado" (input marcado + dropdown abierto) |
| v3.10.82 | Botones TIPIF: **RESET** texto nuevo (Tripartita + modo avión) · **APN→JIRA** (plantilla `CREACION DE TK JIRA` del Excel) · **CASO eliminado** |

**Publicación:** `develop` `90860bd→c163567` (fast-forward, 36 commits, sin force) ·
`main` `be99090→cda031b` (subtree split + `--force-with-lease` + backup atómico previo) ·
Pages build `built` ✅. Detalle del procedimiento en `MEMORY.md` L-091 / L-092.

---

## 🔴 DECISIONES DEL CEO — pendientes de OK (no tocar sin respuesta)

### D1 · `BAJA NO REGISTRADA` → ¿en MÓVIL?

Hoy está en **M.TOTAL + STANDALONE** (2 de 3). El Excel se contradice:

| Fuente | Ubicación | Área / Pool |
|---|---|---|
| `ETIQUETAS.` r37 (canónica) | BAJA NO REGISTRADA (POSTPAGO - PREPAGO) | **MOVISTAR TOTAL** |
| `ETIQUETAS 0` r27 / r34 | BAJA NO REGISTRADAS EN EL SISTEMA | Movistar Total **y** Fija Stand Alone |
| matriz `VALIDACIÓN DE DATOS Y ETIQUETAS` r33 | BAJA NO REGISTRADA EN SISTEMA | pool **BLINDAJE MOVIL** |
| `Validación de datos` r17 | Baja final del **servicio móvil** / Migración a Prepago | Etiqueta Retenciones |

*Recomendación: agregarla a MÓVIL* (mismo `skill_tx2`/`skill_txdirecto`, misma
etiqueta → riesgo cero; el caso móvil "pedí la baja y no laвичé" hoy no aparece
en la pestaña MOVIL). Impacto: 39 → 40 opciones.

### D2 · `AVERÍA HOGAR` en M.TOTAL → `TEFC_PRIORITY_TX` vs `TECNICO_FIJA_TX`

Fuentes 2 vs 2 (`ETIQUETAS` r16 + `ETIQUETAS 0` r12 → `TEFC_PRIORITY`;
`ETIQUETAS.` r38 + matriz r17 → `TECNICO_FIJA`). **Desempata el patrón de la
línea MT** (cada línea tiene su familia de skills — ver `MEMORY.md` L-093):
MT usa `PRIORITY` · `CONVERGENTE` · `BLINDAJE_FIJA_PRIORITY` → `TEFC_PRIORITY`
(=`TEcnica Fija` + `PRIORITY`) encaja; la fila canónica rompe el patrón.

*Recomendación: `SK_CV_TECF_PRIORITY_TX` / `DP_CV_TECF_PRIORITY_TX` solo en
M.TOTAL; STANDALONE mantiene `SK_CV_TECNICO_FIJA_TX`.*

### D3 · Botones — 5 propuestas, elegir prioridad

| # | Propuesta | Base real (código actual) | Ganancia para el agente |
|---|---|---|---|
| **A** | Filtro por **etiqueta operativa** (ROBO · COMERCIAL · RETENCIONES · VENTAS · ROAMING · TECNICO) | el campo `etq` ya existe en las 39 opciones y **no se usa en la UI** | Filtrar por cómo piensa ("es una Etiqueta Robo") |
| **B** | **Teclado** en el dropdown (↑ ↓ Enter Esc) | `#etiquetaSearch` no tiene `keydown` → solo mouse (el autocomplete de ADDRESS sí lo tiene) | Elegir sin soltar el teclado |
| **C** | **Badge de skill** por ítem | el dropdown es solo título (decisión v3.3.4) | Ver a dónde va la TX sin seleccionar |
| **D** | **Contador por chip** (13 · 15 · 11) | hoy solo existe `#etiquetasCount` de la categoría activa | Saber dónde hay más opciones antes de cambiar de chip |
| **E** | **JIRA auto-rellenado** | plantilla con valores fijos del Excel; el popup ya tiene **Cliente** y **Teléfono de consulta** | Copiar con los datos reales del caso |

*Recomendación: A + B* (mayor impacto y **no tocan datos**).

### D4 · `010105:CHILIQUIN` — el PDF escribe el DISTRITO, no un caserío

En el PDF de Grupo 2, la fila dice DEPARTAMENTO AMAZONAS · CHACHAPOYAS · **CHILIQUIN** ·
CENTRO POBLADO **CHILIQUIN**. GeoNames lista para ese distrito el caserío **"Chiliquini"**
(otro nombre) → no hay caserío "ChILIQUIN" con esa grafía.

*Opciones:*
- **(a) Semántica "todo el distrito"** → el popup alerta cuando se selecciona el
  distrito CHILIQUIN (no un CCPP exacto). Requiere un nivel de detección nuevo.
- **(b) Es el caserío "Chiliquini"** (typo en el PDF) → se agrega la clave
  `010105:CHILIQUINI` y la zona queda cubierta por CCPP exacto.
- **(c) No cubrir** (estado actual: sin alerta, documentado).

*Recomendación:* **(b)** — el resto de filas de esa familia (p. ej. `010105:VITUYA`) son
caseríos exactos, y la cobertura por caserío es más precisa que por distrito.

### D5 · 2 zonas de la fuente sin cobertura por falta de datos base

| Zona | Problema | Qué falta |
|---|---|---|
| `060101:INCA CORRAL` (CAJAMARCA) | No existe en UBIGEO ni en CCPP_DATA. El PDF además escribe la provincia como "CAJABAMBA" (el dataset usa "CAJAMARCA") | Definición del caserío/distrito INCA CORRAL (nombre oficial + caseríos) |
| `160405:SANTA ROSA` (LORETO) | El distrito **existe** en el UBIGEO inline (160405) pero tiene **0 caseríos** en CCPP_DATA | Los caseríos del distrito (p. ej. SANTA ROSA) |

*Recomendación:* pedir al CEO la lista de caseríos (o un XLSX) de ambos →
se ingestan con `ingest_comunicado_zonas.py` y quedan con alerta real.

### D6 · 4 caseríos con d6 divergente (PDF PACAIPAMPA vs GeoNames AYABACA)

`CURRILCAS` · `PAPELILLO` · `TOJAS` · `TULMAN DE MATALACAS` → el PDF los atribuye a
PIURA/AYABACA/**PACAIPAMPA** (d6 200206); GeoNames los lista en **AYABACA** (d6 200201),
que es donde el sistema los tiene (y es la única clave alcanzable desde el dropdown).

*Estado:* cubierto funcionalmente. *Decisión:* ¿se deja así (recomendado) o se replica
también la clave 200206 (doble alerta, riesgo de duplicidad)?

---

## 🟢 SIGUIENTE BLOQUE (a definir por el CEO)

| Bloque | Contenido | Esfuerzo |
|---|---|---|
| **A** | Decisiones D1 + D2 (datos) → 40-41 opciones + actualizar `ETIQUETAS_BASELINE` | 30-45 min |
| **B** | Botones D3-A + D3-B (filtro por etiqueta operativa + teclado) | 2-3 h |
| **C** | Botones D3-C + D3-D (badge de skill + contador por chip) | 1-1.5 h |
| **D** | Botón D3-E (JIRA auto-rellenado) | 45 min |
| **E** | Higiene técnica del popup: `acHighlight` (1039 L) y `showApagadoPanelInline` (515 L) sin catalogar · 33 `addEventListener` huérfanos · 21 `setTimeout` sin consolidar · 479 `var` | 6-10 h |
| **F** | **Cargar tabla de zonas COMU-5569** (cuando el CEO entregue el export del recurso `id=178147`) → 1 comando + gate + commit local | 15-30 min |

---

## 📊 Snapshot verificado del popup (2026-11-07)

| Métrica | Valor |
|---|---|
| `apps/cima-generador/index.html` | 1.22 MB · ~8,975 líneas |
| `apps/cima-generador/VERSION` | **3.10.82** ✅ |
| Footer-version / build-chip (HTML) | **v3.10.82** ✅ |
| `ETIQUETAS_DATA` | **39** opciones (13 POSTPAGO / 15 MTOTAL / 11 STANDALONE) |
| Skills `SK_CV_*` en uso | 16 · con contraparte `DP_CV_*`: 14 |
| Tests `test_popup.py` | **750/750** ✅ |
| CSS balance | 565 `{` = 565 `}` ✅ |
| Suites adicionales | aggressive · security · codereview · browser_static → PASS |
| Build servido (`/__build`) | `v3.10.82` · md5 `6f3a35d0d75c…` (cambia con cada release) |

---

## 🧭 Reglas de trabajo (no negociables)

1. **LOCAL FIRST**: todo cambio se commitea en local. Push solo con OK explícito del CEO.
2. **Origen canónico**: solo `localhost:8765` (nunca `127.0.0.1` ni IP de LAN).
3. **Gate antes de commitear**: `cd apps/cima-generador && python3 test_popup.py` (750 asserts)
   + CSS balance. Cambios en CSS → validar llaves.
4. **Excel = fuente de verdad** de ETIQUETAS. Toda skill nueva debe existir en el XLSX
   (`Recursos CIMA/Etiquetas de Transferencia/Etiquetas de Transferencia.xlsx`).
5. **Sin inventar**: si un texto no está en el Excel, se marca como derivado y se declara.
6. **`ETIQUETAS_BASELINE`** (en `test_popup.py`) ata cada opción a su fila de Excel:
   cambiar un skill o un nombre **rompe el gate** → hay que re-verificar contra el Excel.
7. **No tocar remoto** sin OK del CEO en sesión.

---

## 🔧 Comandos de verificación (sesión nueva)

```bash
cd /home/braka/Systems/projects/work/cima
git status && git log --oneline -5

# Popup local (origen canónico)
cd apps/cima-generador && python3 serve.py     # → http://localhost:8765/index.html?popup=1

# Gate obligatorio
cd apps/cima-generador && python3 test_popup.py            # 750/750
python3 test_aggressive.py && python3 test_security.py
python3 test_codereview.py && python3 test_browser_static.py
```

---

*Actualizado por Thavren — 2026-11-07 · sesión v3.10.77 → v3.10.82 (publicada) · checkpoint `c163567`*

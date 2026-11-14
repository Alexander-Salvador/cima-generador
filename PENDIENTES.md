# CIMA — Lista de Pendientes (Estado verificado 2026-11-07)

**Fecha verificación:** 2026-11-07
**Branch:** `develop` (working tree limpio)
**HEAD local:** `c163567` · `origin/develop` sincronizado (mismo SHA)
**Popup local:** `apps/cima-generador/` → **v3.10.82**
**Popup LIVE (Pages):** **v3.10.82** (`cda031b`) → https://alexander-salvador.github.io/cima-generador/
**Tests:** `test_popup.py` **750/750 PASS** + aggressive / security / codereview / browser_static
**DIRECTIVA ACTIVA:** seguir en **LOCAL**. No tocar remoto sin OK explícito del CEO.

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

---

## 🟢 SIGUIENTE BLOQUE (a definir por el CEO)

| Bloque | Contenido | Esfuerzo |
|---|---|---|
| **A** | Decisiones D1 + D2 (datos) → 40-41 opciones + actualizar `ETIQUETAS_BASELINE` | 30-45 min |
| **B** | Botones D3-A + D3-B (filtro por etiqueta operativa + teclado) | 2-3 h |
| **C** | Botones D3-C + D3-D (badge de skill + contador por chip) | 1-1.5 h |
| **D** | Botón D3-E (JIRA auto-rellenado) | 45 min |
| **E** | Higiene técnica del popup: `acHighlight` (1039 L) y `showApagadoPanelInline` (515 L) sin catalogar · 33 `addEventListener` huérfanos · 21 `setTimeout` sin consolidar · 479 `var` | 6-10 h |

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

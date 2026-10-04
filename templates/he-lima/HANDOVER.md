# FEEDBACKS — HANDOVER v17

> **Documento de handover para próximo chat.**
> Lee esto primero antes de continuar.

## 🎯 Estado actual (snapshot 2026-10-04)

### Spreadsheet
- **ID:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`
- **Nombre:** OVAV demo workbook (CIMA — Técnica Móvil)

### Tabs (5)
| Tab | sheetId | Rol |
|---|---|---|
| `PERSONAL` | 1395452450 | 25 agentes (master personal) |
| `HORAS EXTRAS` | 222789142 | HE tracking (no tocar) |
| **`FEEDBACK`** | 681240158 | Master resumen supervisores (CONSTRUIDO) |
| `CONFIG` | 1497962081 | Mapeo FK + parámetros |
| **`FEEDBACK-DETALLE_LLAMADA`** | 412152155 | Base de 26 calls auditadas |

### Worktree
- **Path:** `/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing`
- **Branch:** `feature/c1ma-desing`
- **Working tree:** `e64c1e2`

## 📐 Layout FEEDBACK (v17 — actual)

```
Row 1:    Title "FEEDBACKS - Auditoría Semanal"
Row 2-3:  4 KPIs: 👥 AGENTES | 📞 CALLS | ✅ % CUMPLIDO | ⚠ % VENCIDO
Row 4:    spacer
Row 5:    HEADER (8 cols): AGENTE | TIPO_FEEDBACK | FECHA_FEEDBACK | FECHA_BASE | OBSERVACIÓN | CANT_CALLS | ESTADO | SUPERVISOR
Row 6-12: 7 data rows (RAMIREZ, VARGAS, GOMEZ, MENDOZA, LLAMPASI, IMAN, CANALES)
Row 13-30: 18 rows ENCUADRADAS vacías con styling + borders + dropdowns
```

### Eliminaciones aplicadas
- ❌ ID (enumerador inútil)
- ❌ FECHA_EVAL (duplica fecha)
- ❌ ÚLTIMO_TIPO, COMPROMISO_FECHA, CUMPLIDO_FECHA, RIESGO
- ❌ KPI 5 "ÚLTIMO TIPO" (round 12)
- ❌ STATUS_VIS col I (en plan 17 con 8 cols)
- ❌ ACCION col J (en plan 17 con 8 cols)

## 🔑 Scripts creados (.ovav/templates/he-lima/tools/)

| Archivo | Función |
|---|---|
| `restructure_feedbacks_v1.go` | Initial restructure con master-detail |
| `restructure_feedbacks_v2_fix.go` | Fix OFF-BY-ONE COUNTIF |
| `restructure_feedbacks_v3_fix.go` | Fix CANT_CALLS rango amplio |
| `restructure_feedbacks_v4_navigation.go` | HYPERLINK cross-tab |
| `restructure_feedbacks_v5_mapeo.go` | Tabla mapeo FK→PERSONAL |
| `restructure_feedbacks_v6_fix_vlookup.go` | Fix numberValue vs stringValue |
| `restructure_feedbacks_v7_split_tabs.go` | Split en 3 tabs (FEEDBACK + DETALLE + CONFIG) |
| `restructure_feedbacks_v8_fix_config.go` | Fix merges en CONFIG |
| `restructure_feedbacks_v9_unmerge_config.go` | Unmerge CONFIG A4:F11 |
| `restructure_feedbacks_v10_total_fix.go` | Total fix layout (NO aplicado — error API) |
| `restructure_feedbacks_v11_cleanup.go` | Cleanup properties |
| `restructure_feedbacks_v12_styling.go` | Styling FEEDBACK |
| `restructure_feedbacks_v13_cf_detalles.go` | CF rules + DETALLE styling |
| `restructure_feedbacks_v14_refactor_schema.go` | Refactor 14→8 cols |
| `restructure_feedbacks_v14a_cleanup.go` | v14 fase 1: cleanup |
| `restructure_feedbacks_v14b_rebuild.go` | v14 fase 2: rebuild |
| `restructure_feedbacks_v15_premium.go` | STATUS_VIS + ACCIONES + KPI Dashboard |
| `restructure_feedbacks_v16a_cleanup.go` | v16 fase 1: cleanup (DELETE 23 CF rules) |
| `restructure_feedbacks_v16b_restructure.go` | v16 fase 2: DASHBOARD ARRIBA |
| `restructure_feedbacks_v16c_cleanup2.go` | v16 fase 3: fix rows residuales |
| `restructure_feedbacks_v16d_styling.go` | v16 fase 4: styling |
| `restructure_feedbacks_v16e_fix_tipo.go` | v16 fix TIPO TOP con SUMPRODUCT |
| `restructure_feedbacks_v17_kpi5_removal.go` | Eliminar KPI 5 + encuadrar rows 13-30 |
| `diag_sheet.go` | Diag merges + CF + DATA |
| `diag_config.go` | Diag CONFIG |
| `render_feedbacks_visual.py` | Render PIL v1 |
| `render_cima_3tabs.py` | Render PIL v3 |
| `render_premium_v15.py` | Render v15 |
| `render_v16_final.py` | Render v16 |
| `render_v17_kpi5removed.py` | Render v17 actual |

## 🐛 Bugs resueltos (lecciones aprendidas)

| Bug | Fix |
|---|---|
| `showHelpLink` no es campo válido en dataValidation | Eliminar del JSON |
| COUNTIF rango se ajusta tras INSERT | Usar `$A$15:$A$200` (rango amplio) |
| setFormula en col E off-by-one | Cambiar a col D (CANT_CALLS) |
| `numberValue` vs `stringValue` | Usar `numberValue` para IDs/FKs |
| HYPERLINK display es texto, MATCH necesita number | Usar `+0` para forzar |
| 19 merges pre-existentes en CONFIG | Unmerge exacto row por row |
| batchUpdate atómico con índices shifting | Separar en apply múltiples |
| COUNTIF con array criteria falla | Usar SUMPRODUCT + nested IFs |
| `frozenColCount` no es válido | Usar `frozenColumnCount` |
| `AND` no es ConditionType | Usar CUSTOM_FORMULA con `=(B=x)*(G=y)` |

## 📊 Schema FEEDBACK final

| Col | Header | Tipo |
|---|---|---|
| A | AGENTE | HYPERLINK a DETALLE |
| B | TIPO_FEEDBACK | Dropdown |
| C | FECHA_FEEDBACK | Date |
| D | FECHA_BASE | Date |
| E | OBSERVACIÓN | Text |
| F | CANT_CALLS | Formula COUNTIFS cross-tab |
| G | ESTADO | Dropdown |
| H | SUPERVISOR | Text |

## 🔗 Conexión cross-tab

- **FEEDBACK!F** (CANT_CALLS) = `COUNTIFS(detalle!E, A_row, detalle!K, D_row)`
- **PERSONAL** es la fuente de verdad para nombres
- **CONFIG!A4:C11** mapea FK_FEEDBACK → PERSONAL_N → NOMBRE (legacy)

## 🎯 ¿Qué sigue?

CEO ha indicado:
- ✅ Refactor a 8 cols (hecho)
- ✅ Dashboard ARRIBA (hecho en v16b)
- ✅ Eliminar KPI ÚLTIMO TIPO (hecho en v17)
- ✅ Encuadrar rows 13-30 (hecho en v17)

Próximas directivas posibles:
- Eliminar más cosas (FEEDBACK-DETALLE tab completo?)
- Sort automático de rows
- Audit trail (quién cambió qué)
- Notificaciones automáticas

## 🆘 Comandos útiles para próximo chat

```bash
# Ver estado actual
go run -C /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/go-runtime ./cmd/sheets read --range "FEEDBACK!A1:H12"

# Diagnóstico completo (merges, CF, data)
go run /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/templates/he-lima/tools/diag_sheet.go FEEDBACK

# Render visual
python3 /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/templates/he-lima/tools/render_v17_kpi5removed.py

# Snapshot del sheet (para rollback)
go run -C /home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/go-runtime ./cmd/sheets read --range "FEEDBACK!A1:N50" > ~/.ovav/vault/snapshots/.../FEEDBACK/snapshot-$(date +%Y%m%d).json
```

## 📍 Snapshots guardados

| Directorio | Snapshot |
|---|---|
| `FEEDBACK/pre-restructure-20261004T185543Z.txt` | Pre v1 |
| `FEEDBACK/pre-v10-fix-*.json` | Pre v10 (limpieza) |
| `FEEDBACK/pre-v14-*.json` | Pre v14 refactor |
| `FEEDBACK/pre-v15-*.json` | Pre v15 premium |
| `FEEDBACK/pre-v16-*.json` | Pre v16 dashboard |
| `FEEDBACK/pre-v17-*.json` | Pre v17 KPI 5 removed |

## 👤 Equipo

- **CEO:** Alexander Salvador (decisiones, directivas)
- **Thavren (yo):** Platform Engineering Lead, ejecutor de scripts

## 📜 Contexto dominio

- **CIMA** = Centro de Investigación de Mejoras
- **CIMA Técnica Móvil** = Equipo de call center en Atento Perú
- **Sistema de feedback supervisores** = Los supervisores escuchan calls, las agrupan por agente, generan feedback consolidado
- **DETALLE DE LLAMADAS** = Base/source de calls auditadas (1 row por call)
- **FEEDBACK** = Master/resumen del supervisor (1 row por agente)

## ✅ Para próximo chat

1. Cargar este documento
2. Verificar estado con `git log --oneline -10`
3. Si CEO da nueva directiva, agregar nuevo script `restructure_feedbacks_v18_*.go`
4. Mantener el patrón de apply separados (cleanup → restructure → properties)

## 💾 Patch de respaldo

Como `git push` está bloqueado por reglas de seguridad, se generó un **patch file** con todos los commits:

- **Path:** `.ovav/templates/he-lima/FEATURE-C1MA-DESING-FINAL.patch`
- **Tamaño:** ~400KB
- **Commits incluidos:** 10 (todos los del branch feature/c1ma-desing)

Para aplicar el patch manualmente en otro entorno:
```bash
git apply FEATURE-C1MA-DESING-FINAL.patch
```

## 📡 Estado del push

⚠️ El push a `origin/feature/c1ma-desing` está **bloqueado por reglas de seguridad** en este entorno.
- El branch local está intacto y commiteado
- HANDOVER.md + patch de respaldo commiteados
- Cuando se levanten las reglas o desde otro entorno: `git push origin feature/c1ma-desing --force`

---

**Última actualización:** 2026-10-04T16:58-05:00 (v17 + patch de respaldo aplicado)
**Estado:** ✅ Estable, todo commiteado, patch disponible
**Próximo paso:** Esperar directiva CEO
# HE LIMA — Design Token System v1.0.1

**Status:** Applied to all 5 tabs of `CIMA_2026_PREMIUM_BASE` (2026-09-26).
**Owner:** thavren (Platform Engineering) · UX validation: Elena.
**Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`.

---

## Why this exists

El director quería "global tokens" para que todo el documento tenga estilos consistentes — no más hex ad-hoc dispersos por `HE_VISUAL_SPEC.md` y `architecture.yaml`. Esta carpeta define **una sola fuente de verdad** para colores, tipografía, espaciado, y componentes.

## Files

| Archivo | Propósito |
|---|---|
| `tokens.json` | Sistema canónico — colores, font, spacing, components, accessibility (WCAG AA). |
| `tools/apply-tokens.go` | Helper Go — aplica los tokens a las 5 tabs vía Sheets API batchUpdate. |
| `tools/verify_styles.go` | Verificador — lee celdas reales y confirma que los tokens se aplicaron. |

## How to use

### Re-aplicar tokens a todas las tabs:
```bash
cd .ovav/templates/he-lima/tools
go run apply-tokens.go
```

### Re-aplicar a una sola tab:
```bash
go run apply-tokens.go -tab=10_HE      # solo HE
go run apply-tokens.go -skip-he        # las 4 simples, HE no
```

### Dry-run (ver qué cambiaría sin aplicar):
```bash
go run apply-tokens.go -dry-run
```

### Verificar después de aplicar:
```bash
go run verify_styles.go
```

## Architecture

```
.ovav/templates/he-lima/tokens.json  (canonical, versionado en git)
        ↓
internal/theme/engine (resuelve tokens a hex en runtime)
        ↓
sheets-bridge CLI (--bg token:brand.primary)  ← PRÓXIMO PASO (FASE 2.5)
        ↓
tools/apply-tokens.go (helper actual, antes del CLI integration)
        ↓
Google Sheets API batchUpdate → celdas reales
```

## Token reference

Categorías principales:

| Categoría | Ejemplos | Tokens |
|---|---|---|
| **brand** | `primary` `#1a3a5c`, `primarySubtle` `#d4e3f3`, `primaryHover` `#0f2440` | Títulos, totales, secciones |
| **accent** | `alert` `#e67e22`, `alertSoft` `#fdebd0`, `success` `#1fb6a8`, `successSoft` `#d4f1ed` | Highlights, success |
| **surface** | `base` `#fff`, `zebra` `#f4f7fb`, `zebraWarm` `#fef9e7`, `divider` `#d5dde8`, `disabled` `#f4f4f4` | Backgrounds |
| **text** | `primary` `#0f2440`, `secondary` `#4a5a7a`, `muted` `#5d6b85`, `inverse` `#fff` | Texto (todos WCAG AA) |
| **semantic** | `success`/`warning`/`danger`/`info` + bg variants | Estados semánticos |
| **tab** | `PERSONAL`/`10_HE`/`Feedbacks`/`01_FERIADOS_PE`/`CONFIG`/`_TOKENS` | Color de pestaña |
| **component** | `title_bar`, `data_header`, `data_row_*`, `input_*`, `kpi_card`, `total_row`, etc. | Composites |

Full reference: ver `tokens.json` sección `component` y `accessibility.validated_pairs`.

## Validación WCAG AA (Elena · 2026-09-26)

| Pair fg/bg | Ratio | Pass |
|---|---|---|
| `text.inverse` on `brand.primary` | 11.66:1 | AAA |
| `text.primary` on `surface.base` | 15.63:1 | AAA |
| `text.secondary` on `surface.base` | 7.42:1 | AAA |
| `text.muted` on `surface.base` | 5.37:1 | AA (v1.0.1 fix) |
| `accent.alert pair_text` on `accent.alert` | 5.48:1 | AA (v1.0.1 fix) |
| `dangerBg.pair_text` on `dangerBg` | 8.32:1 | AAA |
| `warningBg.pair_text` on `warningBg` | 6.09:1 | AA |
| `successBg.pair_text` on `successBg` | 6.31:1 | AA |

## Cambios v1.0 → v1.0.1 (aplicados por Elena)

1. `text.muted` `#7b8aa5` → `#5d6b85` (3.49:1 FAIL → 5.37:1 AA).
2. `accent.amber` (ahora `accent.alert`) ganó `pair_text` (white-on-amber era 2.85:1 FAIL).
3. Nombres semánticos: `accent.amber` → `accent.alert`, `accent.teal` → `accent.success`, `brand.navy` → `brand.primary`.
4. Tab color collision fix: `CONFIG` `#1f2937` → `#0d9488`, `_TOKENS` `#374151` → `#64748b`.
5. Font size `xs` 9px → 10px (legibility es-PE).
6. 8 nuevos components: `empty_state`, `sync_pending`, `audit_badge`, `holiday_cell`, `weekend_cell`, `formula_cell`, `legend_item`, `error_row`, `focus_ring`.
7. Categorías nuevas: `elevation` (shadows), `number_format` (es-PE), `focus_ring`.
8. Validated pairs actualizados con ratios reales (computed, no claimed).

## Decisiones del CEO

| Directiva | Estado |
|---|---|
| "todo el documento entero debe tener sus estilos" | ✅ Aplicado a PERSONAL, 10_HE, Feedbacks, 01_FERIADOS_PE, CONFIG |
| "como un global o tokens especificos para no variar tanto" | ✅ `tokens.json` es la fuente única |
| "de la Hoja de HE retiramos row 2" | ✅ Subtitle "Período activo..." eliminado |
| "Alineamos correctamente año y mes" | ✅ Labels zebra + values white, mismo row, gap visual entre Mes y Año |

## Riesgos conocidos

- **`apply-tokens.go` no es idempotente para HE**: si se corre 2x seguidas, elimina 2 filas (la 2da vez elimina la fila que era la 1ra después del 1er delete). Solución actual: si se re-corre, usar `fixup_he_rows.go` (próximo PR) o restaurar manualmente. **TODO: hacer el spec HE idempotent** antes de FASE 2.
- **`fmt.Sscanf` con `%x` retorna 0**: descubrimos esto en sesión. Ya está fixeado en `apply-tokens.go` (usa `strconv.ParseUint` con base 16). Cualquier código similar en otros helpers Go debe revisarse.
- **FASE 2 (fórmulas reales) pendiente**: dropdowns Mes/Año funcionan pero las celdas de horas (D12:AH32) aún no tienen fórmulas de cálculo (H_normal, H_extra, H_25%, H_35%, $_N, $_E).

## Próximos pasos

1. ✅ **COMPLETADO:** Design Token System v1.0.1 aplicado.
2. ⏭️ **PRÓXIMO:** FASE 2 — 10 fórmulas reales en `10_HE` que reaccionen a Mes/Año.
3. ⏭️ Crear tab `_TOKENS` oculto con referencia visual del palette.
4. ⏭️ Integrar token resolver al CLI sheets-bridge (`--bg token:brand.primary`).
5. ⏭️ Hacer `apply-tokens.go` idempotent (CRIT-019 lessons learned).
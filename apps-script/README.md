# OVAV Apps Script — OPERACIONES

Apps Script `OVAV_Tools` bound to OPERACIONES spreadsheet.
Script ID: `14HJvTHtLgGWJ5df7-QTKMn4t76-ZHk1BTJphSKVoyRlV12_CtymEl3cN`

## Files

| File | Purpose |
|---|---|
| `OVAV_Tools.gs` | Source code (Apps Script V8 runtime) |
| `payload.json` | API payload format (files + manifest) |
| `btn_restore.png` | Backup button image (also hosted on Drive) |
| `btn_clean.png` | Backup button image (also hosted on Drive) |

## How to deploy

Via Apps Script API (script ID known, parentId = spreadsheet):

```bash
TOKEN=$(ovav-sheets debug-token | awk '/access_token/{print $2}')
SCRIPT_ID="14HJvTHtLgGWJ5df7-QTKMn4t76-ZHk1BTJphSKVoyRlV12_CtymEl3cN"
SHEET_ID="1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"

# 1. Create bound script (only needed once)
curl -X POST https://script.googleapis.com/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"title\":\"OVAV_Tools\",\"parentId\":\"$SHEET_ID\"}"

# 2. Upload code
curl -X PUT "https://script.googleapis.com/v1/projects/$SCRIPT_ID/content" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @.ovav/apps-script/payload.json

# 3. Create version
curl -X POST "https://script.googleapis.com/v1/projects/$SCRIPT_ID/versions" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"description":"OVAV Tools v5"}'
```

## Menu (auto-installed on `onOpen`)

```
🛠️ OVAV Tools ▼
├── 🔄 Restaurar fórmulas
├── 🎨 Restaurar formato
├── 🧹 Limpiar filas vacías
├── ─────────────
├── 📸 Crear snapshot
├── ⏪ Restaurar snapshot
├── ─────────────
├── 🖼️ Reinstalar botones-imagen
└── ℹ️ Estado del sistema
```

## Image buttons (auto-installed on `onOpen`)

- 🔄 RESTAURAR: anchored at I2 (col 8, row 1), size 140×40 — click → `onRestoreClick()` → `restoreFormulas()`
- 🧹 LIMPIAR: anchored at L2 (col 11, row 1), size 220×40 — click → `onCleanClick()` → `cleanEmptyRows()`

Both Drawing shapes with `assignScriptFunction()`. Real click handlers, not text.

## Drive image URLs

- Restore: `https://drive.google.com/uc?id=1SO07B3HnqsET0gpss-QSGKrP7i-AbMm6&export=download`
- Clean: `https://drive.google.com/uc?id=1JbLGxRv2SEvYoseyudBLr9aRBPjfdgbk&export=download`

These are the URLs baked into the script. If they break, re-upload via Drive API with `drive.file` scope and update the constants in `OVAV_Tools.gs`.

## Backup: CLI bridge commands

Even without Apps Script, these CLI commands do the same work:

```bash
ovav-sheets restore-formulas --tab OPERACIONES
ovav-sheets restore-format --tab OPERACIONES
ovav-sheets clean-empty-rows --tab OPERACIONES
```

CLI is the fallback if Apps Script fails.

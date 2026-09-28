// day1_test_setup.go — CIMA Day 1: Create TEST sheet with 00_Config + 10_Registro_HE.
//
// Strategy:
//   1. Drive API files/copy → duplicate prod spreadsheet → TEST spreadsheet
//   2. Save new spreadsheet ID → .ovav/state/test_sheet_id (gitignored)
//   3. batchUpdate #1: addSheet "00_Config" + write headers A1:C1 + anchors B2:B9 + B11 empty
//      + bg color #f4f7fb for surface + navy #1e3a8a header text + data validation B2:B3
//   4. batchUpdate #2: addSheet "10_Registro_HE" + dimensions 110×22 + headers A1:V1
//      + tab color #f59e0b + hide sheet
//
// Authority: andres · area=platform_engineering · lead=thavren · CEO=braka
// Scope: Day 1 SOLO — sheet creation on TEST. NO Apps Script deploy. NO data migration.

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	prodSpreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	testSheetName     = "HE LIMA - CIMA [TEST 2026-09-28]"
	setupEnc          = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	setupKey          = "/home/braka/.config/ovav/vault.key"
	setupState        = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state"
)

// 00_Config anchors: row 1 = headers, rows 2-9 = values, row 11 = last refresh (empty).
var cfgAnchors = []struct {
	Row  int
	Type string // "number", "string"
	Val  interface{}
}{
	{2, "number", 9},     // mes activo
	{3, "number", 2026},  // año activo
	{4, "string", "W39"}, // semana activa
	{5, "string", "TODOS"},
	{6, "number", 1.25},
	{7, "number", 1.35},
	{8, "number", 2.00},
	{9, "number", 8},
	{11, "string", ""}, // last_refresh — Apps Script lo llena
}

var cfgHeaders = []string{"ANCLA", "VALOR", "DESCRIPCIÓN"}
var cfgDescriptions = []string{
	"",                  // row 1 = headers
	"Mes activo (1-12)", // B2
	"Año activo (YYYY)", // B3
	"Semana activa",     // B4
	"Filtro persona",    // B5
	"Factor HE diurno",  // B6
	"Factor HE nocturno", // B7
	"Factor feriado",    // B8
	"Jornada max diaria", // B9
	"",                  // row 10 = separator
	"Último refresh (auto)", // B11
}

// 10_Registro_HE headers — CANONICAL from .ovav/templates/he-lima/apps-script/03_RegistroEngine.gs
// const REGISTRO_HEADERS (lines 30-53). Day 2.1 fix: byte-for-byte, no translation.
// Source of truth: apps-script/03_RegistroEngine.gs REGISTRO_HEADERS (also matches architecture.yaml).
var registroHeaders = []string{
	"Fecha",
	"ISO_Week",
	"Dia_Semana",
	"Es_Feriado",
	"Tipo_Dia",
	"ID_Persona",
	"Apellidos_Nombres",
	"Turno",
	"Hora_Ingreso",
	"Hora_Salida",
	"Horas_Total",
	"Horas_Normales",
	"Horas_Extras",
	"Factor_Aplicado",
	"Horas_Ponderadas",
	"Pago_HE_PEN",
	"Observaciones",
	"Aud_Usuario",
	"Aud_Timestamp",
	"Estado",
	"Mes_Año",
	"Sync_Status",
}

func main() {
	if err := os.MkdirAll(setupState, 0o755); err != nil {
		die("mkdir state: %v", err)
	}

	creds := decrypt()
	tok := getToken(creds)
	fmt.Println("✅ OAuth token refreshed")

	// ═══ STEP 1: Drive API copy ═══
	newID, err := driveCopy(tok, prodSpreadsheetID, testSheetName)
	if err != nil {
		die("driveCopy: %v", err)
	}
	fmt.Printf("✅ TEST sheet created: %s\n", newID)

	// Save new ID to state file
	idPath := setupState + "/test_sheet_id"
	if err := os.WriteFile(idPath, []byte(newID), 0o644); err != nil {
		die("write state: %v", err)
	}
	fmt.Printf("✅ Saved sheet ID → %s\n", idPath)

	// ═══ STEP 2: batchUpdate for 00_Config ═══
	cfgSheetID, err := addCfgSheet(tok, newID)
	if err != nil {
		die("addCfgSheet: %v", err)
	}
	fmt.Printf("✅ 00_Config created (sheetId=%d)\n", cfgSheetID)

	// ═══ STEP 3: batchUpdate for 10_Registro_HE ═══
	regSheetID, err := addRegistroSheet(tok, newID)
	if err != nil {
		die("addRegistroSheet: %v", err)
	}
	fmt.Printf("✅ 10_Registro_HE created (sheetId=%d, hidden, orange)\n", regSheetID)

	fmt.Println("\n=== DAY 1 SETUP COMPLETE ===")
	fmt.Printf("TEST sheet ID: %s\n", newID)
	fmt.Printf("00_Config ID:  %d\n", cfgSheetID)
	fmt.Printf("10_Registro_HE ID: %d\n", regSheetID)
}

// driveCopy creates a copy of prodSpreadsheetID with the given name via Drive API.
// Returns the new spreadsheet ID.
func driveCopy(tok, fileID, name string) (string, error) {
	body, _ := json.Marshal(map[string]any{"name": name})
	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/copy", fileID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", fmt.Errorf("unmarshal: %w (body: %s)", err, string(b))
	}
	id, ok := out["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("no id in response: %s", string(b))
	}
	return id, nil
}

// addCfgSheet adds the 00_Config sheet with anchors + styling. Returns the new sheet ID.
func addCfgSheet(tok, spreadsheetID string) (int64, error) {
	// Build batchUpdate payload — single round-trip with all operations.
	reqs := []any{}

	// 1. Add sheet at index 0 (top), with specific properties
	reqs = append(reqs, map[string]any{
		"addSheet": map[string]any{
			"properties": map[string]any{
				"sheetId":     1001, // stable ID for 00_Config
				"title":       "00_Config",
				"index":       0,
				"tabColor":    map[string]any{"red": 0.118, "green": 0.227, "blue": 0.541}, // #1e3a8a navy
				"gridProperties": map[string]any{
					"rowCount":    60,
					"columnCount": 8,
					"frozenRowCount": 1,
				},
			},
		},
	})

	// 2. Write headers A1:C1
	headerValues := [][]any{
		{cfgHeaders[0], cfgHeaders[1], cfgHeaders[2]},
	}
	reqs = append(reqs, updateCells(1001, 0, 0, headerValues))

	// 3. Write anchor labels A2:A11 + descriptions C2:C11 (skip row 10)
	anchorRows := [][]any{}
	for i, a := range cfgAnchors {
		row := a.Row - 1 // 0-indexed
		// Anchor label = CFG_ANCHORS key. We use semantic name.
		label := cfgAnchorLabel(i)
		desc := ""
		if row < len(cfgDescriptions) {
			desc = cfgDescriptions[row]
		}
		anchorRows = append(anchorRows, []any{label, "", desc})
	}
	reqs = append(reqs, updateCells(1001, 1, 0, anchorRows))

	// 4. Write anchor VALUES in column B (rows 2-9, row 11)
	valueRows := [][]any{}
	for _, a := range cfgAnchors {
		var cell map[string]any
		switch a.Type {
		case "number":
			cell = map[string]any{"userEnteredValue": map[string]any{"numberValue": a.Val}}
		default:
			cell = map[string]any{"userEnteredValue": map[string]any{"stringValue": a.Val}}
		}
		valueRows = append(valueRows, []any{cell})
	}
	// B column starts at col index 1, rows 1-10 (0-indexed for rows 2-11)
	reqs = append(reqs, updateCellsColumn(1001, 1, 1, valueRows))

	// 5. Apply data validation B2:B3 (mes 1-12, año 2024-2030)
	reqs = append(reqs, map[string]any{
		"setDataValidation": map[string]any{
			"range": map[string]any{
				"sheetId":          1001,
				"startRowIndex":    1,
				"endRowIndex":      2,
				"startColumnIndex": 1,
				"endColumnIndex":   2,
			},
			"rule": map[string]any{
				"condition": map[string]any{
					"type":   "NUMBER_BETWEEN",
					"values": []any{
						map[string]any{"userEnteredValue": "1"},
						map[string]any{"userEnteredValue": "12"},
					},
				},
				"inputMessage":    "Mes activo (1-12)",
				"strict":          true,
				"showCustomUi":    true,
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"setDataValidation": map[string]any{
			"range": map[string]any{
				"sheetId":          1001,
				"startRowIndex":    2,
				"endRowIndex":      3,
				"startColumnIndex": 1,
				"endColumnIndex":   2,
			},
			"rule": map[string]any{
				"condition": map[string]any{
					"type":   "NUMBER_BETWEEN",
					"values": []any{
						map[string]any{"userEnteredValue": "2024"},
						map[string]any{"userEnteredValue": "2030"},
					},
				},
				"inputMessage": "Año activo (2024-2030)",
				"strict":       true,
				"showCustomUi": true,
			},
		},
	})

	// 6. Apply bg color surface.config (#f4f7fb) to data area A2:C11
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          1001,
				"startRowIndex":    1,
				"endRowIndex":      11,
				"startColumnIndex": 0,
				"endColumnIndex":   3,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.957,
						"green": 0.969,
						"blue":  0.984,
					},
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// 7. Header row navy bg + white bold text
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          1001,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   3,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.118,
						"green": 0.227,
						"blue":  0.541,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   1.0,
							"green": 1.0,
							"blue":  1.0,
						},
						"bold": true,
					},
					"horizontalAlignment": "CENTER",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	})

	// 8. Bold the anchor values (column B, rows 2-9 + 11)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          1001,
				"startRowIndex":    1,
				"endRowIndex":      11,
				"startColumnIndex": 1,
				"endColumnIndex":   2,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"textFormat": map[string]any{"bold": true},
					"horizontalAlignment": "CENTER",
				},
			},
			"fields": "userEnteredFormat(textFormat,horizontalAlignment)",
		},
	})

	// 9. Column widths: A=140, B=120, C=320
	reqs = append(reqs, map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    1001,
				"dimension":  "COLUMNS",
				"startIndex": 0,
				"endIndex":   1,
			},
			"properties": map[string]any{"pixelSize": 140},
			"fields":     "pixelSize",
		},
	})
	reqs = append(reqs, map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    1001,
				"dimension":  "COLUMNS",
				"startIndex": 1,
				"endIndex":   2,
			},
			"properties": map[string]any{"pixelSize": 120},
			"fields":     "pixelSize",
		},
	})
	reqs = append(reqs, map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    1001,
				"dimension":  "COLUMNS",
				"startIndex": 2,
				"endIndex":   3,
			},
			"properties": map[string]any{"pixelSize": 320},
			"fields":     "pixelSize",
		},
	})

	return batchUpdate(tok, spreadsheetID, reqs, 1001)
}

// addRegistroSheet adds the 10_Registro_HE working surface, hidden, orange tab.
func addRegistroSheet(tok, spreadsheetID string) (int64, error) {
	reqs := []any{}

	// 1. Add sheet at index 1 (after 00_Config), with explicit 110×22 dimensions
	reqs = append(reqs, map[string]any{
		"addSheet": map[string]any{
			"properties": map[string]any{
				"sheetId":     1002, // stable ID for 10_Registro_HE
				"title":       "10_Registro_HE",
				"index":       1,
				"tabColor":    map[string]any{"red": 0.961, "green": 0.620, "blue": 0.043}, // #f59e0b orange
				"gridProperties": map[string]any{
					"rowCount":    110,
					"columnCount": 22,
					"frozenRowCount":    1,
					"frozenColumnCount": 5, // freeze ID-DNI-NOMBRE-CARGO-FECHA (cols A-E)
				},
			},
		},
	})

	// 2. Hide sheet (working surface — Apps Script unhides on demand)
	reqs = append(reqs, map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": 1002,
				"hidden":   true,
			},
			"fields": "hidden",
		},
	})

	// 3. Write headers A1:V1
	headerRow := [][]any{}
	hdrRow := []any{}
	for _, h := range registroHeaders {
		hdrRow = append(hdrRow, map[string]any{
			"userEnteredValue": map[string]any{"stringValue": h},
			"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{
					"red":   0.118,
					"green": 0.227,
					"blue":  0.541,
				},
				"textFormat": map[string]any{
					"foregroundColor": map[string]any{
						"red": 1.0, "green": 1.0, "blue": 1.0,
					},
					"bold":      true,
					"fontSize":  10,
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment":   "MIDDLE",
				"wrapStrategy":        "WRAP",
			},
		})
	}
	headerRow = append(headerRow, hdrRow)
	reqs = append(reqs, updateCells(1002, 0, 0, headerRow))

	// 4. Set narrow widths for ID/DNI/NOMBRE/CARGO, default for rest
	widths := []struct {
		Col   int // 0-indexed
		PxSize int
	}{
		{0, 60},  // ID
		{1, 80},  // DNI
		{2, 200}, // NOMBRE
		{3, 120}, // CARGO
		{4, 100}, // FECHA
		{5, 90},  // H_ENTRADA
		{6, 90},  // H_SALIDA
		{7, 70},  // H_NORM
		{8, 70},  // H_EXTRA
		{9, 60},  // H_25
		{10, 60}, // H_35
		{11, 80}, // $_N
		{12, 80}, // $_E
		{13, 100}, // ESTADO
		{14, 200}, // OBSERVACIONES
		{15, 120}, // AUD_USUARIO
		{16, 140}, // AUD_TS
		{17, 100}, // AUD_OLD
		{18, 100}, // AUD_NEW
		{19, 200}, // AUD_HASH
		{20, 100}, // AUD_TAG
		{21, 100}, // SYNC_STATUS
	}
	for _, w := range widths {
		reqs = append(reqs, map[string]any{
			"updateDimensionProperties": map[string]any{
				"range": map[string]any{
					"sheetId":    1002,
					"dimension":  "COLUMNS",
					"startIndex": w.Col,
					"endIndex":   w.Col + 1,
				},
				"properties": map[string]any{"pixelSize": w.PxSize},
				"fields":     "pixelSize",
			},
		})
	}

	return batchUpdate(tok, spreadsheetID, reqs, 1002)
}

// batchUpdate sends a batchUpdate request. Returns the new sheetId (or 0 if not an addSheet).
func batchUpdate(tok, spreadsheetID string, reqs []any, expectedSheetID int64) (int64, error) {
	body, _ := json.Marshal(map[string]any{"requests": reqs, "includeSpreadsheetInResponse": false})
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("POST: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	return expectedSheetID, nil
}

// updateCells builds a 2D updateCells request.
func updateCells(sheetID int64, startRow, startCol int, values [][]any) map[string]any {
	rows := []any{}
	for _, r := range values {
		cells := []any{}
		for _, c := range r {
			if m, ok := c.(map[string]any); ok {
				cells = append(cells, m)
			} else {
				cells = append(cells, map[string]any{
					"userEnteredValue": map[string]any{"stringValue": fmt.Sprintf("%v", c)},
				})
			}
		}
		rows = append(rows, map[string]any{"values": cells})
	}
	rowCount := len(values)
	colCount := 0
	if rowCount > 0 {
		colCount = len(values[0])
	}
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      startRow + rowCount,
				"startColumnIndex": startCol,
				"endColumnIndex":   startCol + colCount,
			},
			"rows":  rows,
			"fields": "userEnteredValue,userEnteredFormat",
		},
	}
}

// updateCellsColumn writes single-column data starting at (startRow, startCol).
func updateCellsColumn(sheetID int64, startRow, startCol int, values [][]any) map[string]any {
	rows := []any{}
	for _, r := range values {
		if len(r) == 0 {
			continue
		}
		rows = append(rows, map[string]any{"values": []any{r[0]}})
	}
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      startRow + len(values),
				"startColumnIndex": startCol,
				"endColumnIndex":   startCol + 1,
			},
			"rows":  rows,
			"fields": "userEnteredValue",
		},
	}
}

// cfgAnchorLabel returns the semantic label for each anchor row.
func cfgAnchorLabel(idx int) string {
	labels := []string{
		"mes", "anio", "semana", "persona",
		"factor_diurno", "factor_nocturno", "factor_feriado",
		"jornada_max", "last_refresh",
	}
	if idx < len(labels) {
		return labels[idx]
	}
	return ""
}

func decrypt() map[string]any {
	keyB, err := os.ReadFile(setupKey)
	if err != nil {
		die("read key: %v", err)
	}
	encB, err := os.ReadFile(setupEnc)
	if err != nil {
		die("read enc: %v", err)
	}
	block, err := aes.NewCipher(keyB)
	if err != nil {
		die("cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		die("gcm: %v", err)
	}
	pt, err := gcm.Open(nil, encB[:12], encB[12:], nil)
	if err != nil {
		die("decrypt: %v", err)
	}
	creds := map[string]any{}
	if err := json.Unmarshal(pt, &creds); err != nil {
		die("unmarshal creds: %v", err)
	}
	return creds
}

func getToken(c map[string]any) string {
	body, _ := json.Marshal(map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	})
	r, err := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	if err != nil {
		die("token POST: %v", err)
	}
	defer r.Body.Close()
	res := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		die("decode token: %v", err)
	}
	tok, ok := res["access_token"].(string)
	if !ok || tok == "" {
		die("no access_token: %v", res)
	}
	return tok
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}

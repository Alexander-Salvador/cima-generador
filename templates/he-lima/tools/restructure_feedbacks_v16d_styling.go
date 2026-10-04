// restructure_feedbacks_v16d_styling.go — Fase D
//
// Styling premium + properties para DASHBOARD + DATA area
// - Styling: DASHBOARD header, KPI cards, data header, zebra, borders
// - CF rules: TIPO_FEEDBACK (col B), ESTADO (col G), STATUS (col I)
// - Data validation: TIPO_FEEDBACK, ESTADO dropdowns rows 9-30
// - Hyperlink styling (cols A, F)
// - Freeze rows 1-8
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
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetFB       = int64(681240158)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v16d — STYLING + PROPERTIES")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STYLING — Dashboard area (rows 1-7) + Data area (rows 8-15)
	// ═══════════════════════════════════════════════════════════════════

	// Title row 1: dark blue bg + white text
	fmt.Println("━━━ Title row 1 styling")
	reqs = append(reqs, styleRange(0, 0, 1, 10, sheetFB, "#1a3a5c", "#FFFFFF", true, false, 16))

	// Subtitle row 2: light blue bg
	fmt.Println("━━━ Subtitle row 2 styling")
	reqs = append(reqs, styleRange(1, 0, 2, 10, sheetFB, "#e8f0fa", "#1a3a5c", false, true, 11))

	// DASHBOARD title row 4: dark blue bg
	fmt.Println("━━━ DASHBOARD title row 4")
	reqs = append(reqs, styleRange(3, 0, 4, 10, sheetFB, "#1a3a5c", "#FFFFFF", true, false, 12))

	// KPI LABELS row 5: gray bg + dark blue text
	fmt.Println("━━━ KPI labels row 5")
	reqs = append(reqs, styleRange(4, 0, 5, 10, sheetFB, "#f0f4f8", "#1a3a5c", true, false, 11))

	// KPI VALUES row 6: white bg + dark blue text, LARGE font
	fmt.Println("━━━ KPI values row 6 (big font 20pt)")
	reqs = append(reqs, styleRange(5, 0, 6, 10, sheetFB, "#FFFFFF", "#1a3a5c", true, false, 20))

	// DATA HEADER row 8: dark blue bg + white text
	fmt.Println("━━━ Data header row 8")
	reqs = append(reqs, styleRange(7, 0, 8, 10, sheetFB, "#1a3a5c", "#FFFFFF", true, false, 11))

	// DATA rows 9-15: zebra striping
	fmt.Println("━━━ Data rows 9-15 zebra")
	for r := 9; r <= 15; r++ {
		bg := "#FFFFFF"
		if r%2 == 1 {
			bg = "#f5f8fc"
		}
		reqs = append(reqs, styleRange(r-1, 0, r, 10, sheetFB, bg, "#0F2440", false, false, 10))
	}

	// EMPTY rows 16-30: zebra striping (para que se vea ordenado cuando el supervisor agregue nuevos)
	fmt.Println("━━━ Empty rows 16-30 zebra (placeholder)")
	for r := 16; r <= 30; r++ {
		bg := "#FFFFFF"
		if r%2 == 1 {
			bg = "#fafcfe"
		}
		reqs = append(reqs, styleRange(r-1, 0, r, 10, sheetFB, bg, "#0F2440", false, false, 10))
	}

	// HYPERLINK styling en A y F
	fmt.Println("━━━ HYPERLINK styling A, F")
	for r := 9; r <= 15; r++ {
		for _, col := range []string{"A", "F"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetFB, "", "#1a73e8", true, true, 10))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// BORDERS (sutiles, premium minimalista)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ Borders subtle: underline row 8 (header divider)")
	reqs = append(reqs, hBorder(8, 0, 10, sheetFB, 2))

	// ═══════════════════════════════════════════════════════════════════
	// DATA VALIDATIONS rows 9-30
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ Data validations rows 9-30")
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "B9:B30",
		[]string{"Positivo", "Negativo", "Evaluativo", "Amonestador"}))
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "G9:G30",
		[]string{"Pendiente", "En evaluación", "Feedback dado", "Carta de compromiso", "Cumplido", "Vencido"}))

	// ═══════════════════════════════════════════════════════════════════
	// CF rules — rango rows 9-30 (para incluir nuevos agentes)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ CF rules rows 9-30")

	// TIPO_FEEDBACK (col B)
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Positivo", 0.6, 0.9, 0.6},
		{"Negativo", 1.0, 0.6, 0.6},
		{"Evaluativo", 0.6, 0.85, 1.0},
		{"Amonestador", 1.0, 0.5, 0.3},
	} {
		reqs = append(reqs, addCF(sheetFB, 1, 9, 1, 30, c.Val, c.R, c.G, c.B))
	}

	// ESTADO (col G)
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Pendiente", 0.85, 0.85, 0.85},
		{"En evaluación", 1.0, 1.0, 0.6},
		{"Feedback dado", 0.6, 0.85, 1.0},
		{"Carta de compromiso", 1.0, 0.85, 0.4},
		{"Cumplido", 0.5, 0.9, 0.5},
		{"Vencido", 1.0, 0.5, 0.5},
	} {
		reqs = append(reqs, addCF(sheetFB, 6, 9, 6, 30, c.Val, c.R, c.G, c.B))
	}

	// STATUS_VIS (col I) — TEXT_CONTAINS rules
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"✓", 0.6, 0.9, 0.6},
		{"⚠", 1.0, 0.6, 0.6},
		{"📋", 1.0, 0.85, 0.4},
		{"💬", 0.6, 0.85, 1.0},
		{"⏳", 1.0, 1.0, 0.6},
	} {
		reqs = append(reqs, addCFContains(sheetFB, 8, 9, 8, 30, c.Val, c.R, c.G, c.B))
	}

	fmt.Printf("\nTotal: %d requests\n", len(reqs))

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	respB, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	var result map[string]any
	json.Unmarshal(respB, &result)
	replies, _ := result["replies"].([]any)
	fmt.Printf("✅ Sent %d, replies %d\n", len(reqs), len(replies))
	fmt.Printf("🕐 Done: %s\n", time.Now().Format(time.RFC3339))
}

// ═══════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════

func colIdx(cell string) int {
	col := 0
	for _, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		}
	}
	return col - 1
}

func styleRange(startRow, startCol, endRow, endCol int, sheetID int64, bgHex, fgHex string, bold, italic bool, fontSize int) map[string]any {
	textFormat := map[string]any{
		"foregroundColor": map[string]any{
			"red":   hexToR(gbParse(fgHex)),
			"green": hexToG(gbParse(fgHex)),
			"blue":  hexToB(gbParse(fgHex)),
		},
		"fontSize": fontSize,
	}
	if bold {
		textFormat["bold"] = true
	}
	if italic {
		textFormat["italic"] = true
	}
	bgColor := map[string]any{
		"red":   hexToR(gbParse(bgHex)),
		"green": hexToG(gbParse(bgHex)),
		"blue":  hexToB(gbParse(bgHex)),
	}
	return map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgColor,
					"textFormat":      textFormat,
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	}
}

func styleCell(startRow, startCol, endRow, endCol int, sheetID int64, bgHex, fgHex string, bold, underline bool, fontSize int) map[string]any {
	textFormat := map[string]any{
		"foregroundColor": map[string]any{
			"red":   hexToR(gbParse(fgHex)),
			"green": hexToG(gbParse(fgHex)),
			"blue":  hexToB(gbParse(fgHex)),
		},
		"fontSize": fontSize,
	}
	if bold {
		textFormat["bold"] = true
	}
	if underline {
		textFormat["underline"] = true
	}
	bgColor := map[string]any{
		"red":   hexToR(gbParse(bgHex)),
		"green": hexToG(gbParse(bgHex)),
		"blue":  hexToB(gbParse(bgHex)),
	}
	return map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgColor,
					"textFormat":      textFormat,
					"horizontalAlignment": "LEFT",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	}
}

func hBorder(row, startCol, endCol int, sheetID int64, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    row - 1,
				"endRowIndex":      row,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"bottom": map[string]any{"style": "SOLID", "width": width, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
		},
	}
}

func dataValidationOneOfList(sheetID int64, a1Range string, values []string) map[string]any {
	_, startRow, endRow, startCol, endCol := a1ToGrid(a1Range)
	return map[string]any{
		"setDataValidation": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"rule": map[string]any{
				"condition": map[string]any{
					"type":   "ONE_OF_LIST",
					"values": valuesToUserEntered(values),
				},
				"showCustomUi": true,
			},
		},
	}
}

func valuesToUserEntered(values []string) []map[string]any {
	out := make([]map[string]any, len(values))
	for i, v := range values {
		out[i] = map[string]any{
			"userEnteredValue": v,
		}
	}
	return out
}

func a1ToGrid(a1 string) (string, int, int, int, int) {
	parts := strings.Split(a1, ":")
	if len(parts) != 2 {
		return a1, 0, 1, 0, 1
	}
	startColS := strings.TrimLeft(parts[0], "0123456789")
	startRowS := strings.TrimLeft(parts[0], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	startCol := colIdx(startColS)
	startRow, _ := strconvAtoi(startRowS)
	endColS := strings.TrimLeft(parts[1], "0123456789")
	endRowS := strings.TrimLeft(parts[1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	endCol := colIdx(endColS) + 1
	endRow, _ := strconvAtoi(endRowS)
	return a1, startRow - 1, endRow, startCol, endCol
}

func strconvAtoi(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func addCF(sheetID int64, col, startRow, _, endRow int, value string, r, g, b float64) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetID,
						"startRowIndex":    startRow - 1,
						"endRowIndex":      endRow,
						"startColumnIndex": col,
						"endColumnIndex":   col + 1,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type":   "TEXT_EQ",
						"values": []map[string]any{{"userEnteredValue": value}},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{"red": r, "green": g, "blue": b},
					},
				},
			},
			"index": 0,
		},
	}
}

func addCFContains(sheetID int64, col, startRow, _, endRow int, value string, r, g, b float64) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetID,
						"startRowIndex":    startRow - 1,
						"endRowIndex":      endRow,
						"startColumnIndex": col,
						"endColumnIndex":   col + 1,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type":   "TEXT_CONTAINS",
						"values": []map[string]any{{"userEnteredValue": value}},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{"red": r, "green": g, "blue": b},
					},
				},
			},
			"index": 0,
		},
	}
}

type rgb struct{ r, g, b int }

func gbParse(hex string) rgb {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return rgb{255, 255, 255}
	}
	var v int
	fmt.Sscanf(hex, "%x", &v)
	return rgb{
		r: (v >> 16) & 0xFF,
		g: (v >> 8) & 0xFF,
		b: v & 0xFF,
	}
}

func hexToR(c rgb) float64 { return float64(c.r) / 255.0 }
func hexToG(c rgb) float64 { return float64(c.g) / 255.0 }
func hexToB(c rgb) float64 { return float64(c.b) / 255.0 }

func decrypt() map[string]any {
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	out := map[string]any{}
	json.Unmarshal(pt, &out)
	return out
}

func getToken(c map[string]any) string {
	body := map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	}
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok, _ := res["access_token"].(string)
	return tok
}
// restructure_feedbacks_v12_styling.go — Paso 2: styling FEEDBACK master
//
// Aplica:
//   - Header row 4 styling (dark blue bg + white text + bold)
//   - Zebra striping rows 5-11
//   - Borders (frame + interiores)
//   - Col widths
//   - Freeze rows 1-4
//   - HYPERLINK styling (azul + underline) para A, B, D
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
	fmt.Println("FEEDBACKS v12 — STYLING FEEDBACK master")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// Title row 1 styling (centered, larger)
	fmt.Println("  Title row 1: bg dark blue + fg white + bold")
	reqs = append(reqs, styleRange(0, 0, 1, 14, sheetFB, "#1a3a5c", "#FFFFFF", true, true, 14))

	// Subtitle row 2 styling
	fmt.Println("  Subtitle row 2: bg light blue + fg dark + italic")
	reqs = append(reqs, styleRange(1, 0, 2, 14, sheetFB, "#d4e3f3", "#1a3a5c", false, true, 11))

	// Header row 4: dark blue bg + white text + bold + centered
	fmt.Println("  Header row 4: bg dark blue + fg white + bold + centered")
	reqs = append(reqs, styleRange(3, 0, 4, 14, sheetFB, "#1a3a5c", "#FFFFFF", true, false, 11))

	// Zebra striping rows 5-11
	fmt.Println("  Zebra rows 5-11 (odd=light gray, even=white)")
	for _, r := range []int{5, 7, 9, 11} {
		reqs = append(reqs, styleRange(r-1, 0, r, 14, sheetFB, "#f0f4f8", "#0F2440", false, false, 10))
	}
	for _, r := range []int{6, 8, 10} {
		reqs = append(reqs, styleRange(r-1, 0, r, 14, sheetFB, "#FFFFFF", "#0F2440", false, false, 10))
	}

	// Borders
	fmt.Println("  Borders: frame HEADER (1-4) + frame DATA (5-11)")
	reqs = append(reqs, frameBorder(0, 0, 4, 14, sheetFB, 2))
	reqs = append(reqs, frameBorder(4, 0, 11, 14, sheetFB, 2))
	fmt.Println("  Borders: horizontal separators entre rows 5-11")
	for r := 6; r <= 11; r++ {
		reqs = append(reqs, hBorder(r, 0, 14, sheetFB))
	}

	// Col widths
	fmt.Println("  Col widths: ID=30, AGENTE=180, SUP=160, CANT=70, TIPO_FB=110, ...")
	colWidths := []int{30, 180, 160, 70, 110, 100, 90, 90, 110, 300, 110, 110, 130, 70}
	for i, w := range colWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetFB))
	}

	// Freeze
	fmt.Println("  Freeze rows 1-4")
	reqs = append(reqs, setFreeze(4, 0, sheetFB))

	// HYPERLINK styling (azul + underline) para celdas con link
	fmt.Println("  HYPERLINK styling (azul + underline) en A, B, D rows 5-11")
	for r := 5; r <= 11; r++ {
		for _, col := range []string{"B", "D"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetFB, "#FFFFFF", "#1a73e8", false, true, 10))
		}
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

func frameBorder(startRow, startCol, endRow, endCol int, sheetID int64, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":    map[string]any{"style": "SOLID", "width": width},
			"bottom": map[string]any{"style": "SOLID", "width": width},
			"left":   map[string]any{"style": "SOLID", "width": width},
			"right":  map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func hBorder(row, startCol, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    row - 1,
				"endRowIndex":      row,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"bottom": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.7, "green": 0.7, "blue": 0.7}},
		},
	}
}

func setColWidth(startIdx, endIdx, pixelSize int, sheetID int64) map[string]any {
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "COLUMNS",
				"startIndex": startIdx,
				"endIndex":   endIdx,
			},
			"properties": map[string]any{"pixelSize": pixelSize},
			"fields":     "pixelSize",
		},
	}
}

func setFreeze(frozenRowCount, frozenColCount int, sheetID int64) map[string]any {
	return map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": sheetID,
				"gridProperties": map[string]any{
					"frozenRowCount":    frozenRowCount,
					"frozenColumnCount": frozenColCount,
				},
			},
			"fields": "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
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
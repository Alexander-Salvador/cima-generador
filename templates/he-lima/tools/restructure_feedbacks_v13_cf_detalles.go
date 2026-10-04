// restructure_feedbacks_v13_cf_detalles.go — Paso 3: CF rules + DETALLE styling
//
// Aplica:
//   A. CF rules FEEDBACK (4 TIPO_FEEDBACK + 6 ESTADO)
//   B. Styling DETALLE_LLAMADA completo (rows 1-28)
//   C. CF DETALLE (3 ESTADO_AUDITORIA + 4 TIPO_LLAMADA)
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
	sheetDETALLE  = int64(412152155)
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
	fmt.Println("FEEDBACKS v13 — CF RULES + DETALLE STYLING")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// ═══ A. CF rules FEEDBACK ═══
	fmt.Println("\n━━━ A. CF rules FEEDBACK")

	fmt.Println("  TIPO_FEEDBACK (col E rows 5-11)")
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Positivo", 0.6, 0.9, 0.6},
		{"Negativo", 1.0, 0.6, 0.6},
		{"Evaluativo", 0.6, 0.85, 1.0},
		{"Amonestador", 1.0, 0.5, 0.3},
	} {
		reqs = append(reqs, addCF(sheetFB, 4, 5, 4, 11, c.Val, c.R, c.G, c.B))
	}

	fmt.Println("  ESTADO (col M rows 5-11)")
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
		reqs = append(reqs, addCF(sheetFB, 12, 5, 12, 11, c.Val, c.R, c.G, c.B))
	}

	// ═══ B. Styling DETALLE_LLAMADA ═══
	fmt.Println("\n━━━ B. Styling FEEDBACK-DETALLE_LLAMADA")

	// Title row 1
	fmt.Println("  Title row 1: bg dark blue + fg white + bold")
	reqs = append(reqs, styleRange(0, 0, 1, 12, sheetDETALLE, "#1a3a5c", "#FFFFFF", true, true, 14))

	// Header row 2
	fmt.Println("  Header row 2: bg dark blue + fg white + bold")
	reqs = append(reqs, styleRange(1, 0, 2, 12, sheetDETALLE, "#1a3a5c", "#FFFFFF", true, false, 11))

	// Zebra striping data rows 3-28
	fmt.Println("  Zebra rows 3-28 (odd=light gray)")
	for r := 3; r <= 28; r++ {
		bg := "#FFFFFF"
		if r%2 == 1 {
			bg = "#f0f4f8"
		}
		reqs = append(reqs, styleRange(r-1, 0, r, 12, sheetDETALLE, bg, "#0F2440", false, false, 10))
	}

	// Borders
	fmt.Println("  Borders: frame HEADER (1-2) + frame DATA (3-28)")
	reqs = append(reqs, frameBorder(0, 0, 2, 12, sheetDETALLE, 2))
	reqs = append(reqs, frameBorder(2, 0, 28, 12, sheetDETALLE, 2))
	for r := 4; r <= 28; r++ {
		reqs = append(reqs, hBorder(r, 0, 12, sheetDETALLE))
	}

	// Col widths DETALLE
	fmt.Println("  Col widths DETALLE")
	detalleWidths := []int{40, 40, 100, 150, 170, 130, 130, 220, 110, 60, 90, 150}
	for i, w := range detalleWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetDETALLE))
	}

	// Freeze rows 1-2
	fmt.Println("  Freeze rows 1-2")
	reqs = append(reqs, setFreeze(2, 0, sheetDETALLE))

	// HYPERLINK styling DETALLE en B, E, J cols rows 3-28
	fmt.Println("  HYPERLINK styling DETALLE (B, E, J cols rows 3-28)")
	for r := 3; r <= 28; r++ {
		for _, col := range []string{"B", "E", "J"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetDETALLE, "#FFFFFF", "#1a73e8", false, true, 10))
		}
	}

	// ═══ C. CF rules DETALLE ═══
	fmt.Println("\n━━━ C. CF rules DETALLE")

	fmt.Println("  TIPO_LLAMADA (col G rows 3-28)")
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Estados incorrectos", 1.0, 0.7, 0.7},
		{"Abandono llamada", 1.0, 0.85, 0.4},
		{"Speech incompleto", 0.7, 0.85, 1.0},
		{"No devuelve llamada", 1.0, 0.7, 1.0},
	} {
		reqs = append(reqs, addCF(sheetDETALLE, 6, 3, 6, 28, c.Val, c.R, c.G, c.B))
	}

	fmt.Println("  ESTADO_AUDITORIA (col I rows 3-28)")
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Pendiente", 0.85, 0.85, 0.85},
		{"Auditada", 0.6, 0.9, 0.6},
		{"Con feedback", 0.6, 0.85, 1.0},
	} {
		reqs = append(reqs, addCF(sheetDETALLE, 8, 3, 8, 28, c.Val, c.R, c.G, c.B))
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
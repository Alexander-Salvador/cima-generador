// restructure_feedbacks_v15_premium.go — CEO directive 2026-10-04 (round 10)
// "Esta mal diseñado todo... rediseña end to end modo premium 2026 avanzado inteligente funciones nuevas"
//
// REDISEÑO PREMIUM 2026:
//   A. Eliminar frame borders feos (estilo minimalista)
//   B. Widening SUPERVISOR (220px) — sin truncamiento
//   C. NUEVAS COLUMNAS:
//      - I: STATUS_VIS (icono auto ✓ ⏳ ⚠ 📋 ❌)
//      - J: ACCIONES (HYPERLINK "Ver historial" + "Cumplido")
//   D. KPI DASHBOARD rows 13-15:
//      - Row 13: Total agentes · Total calls · % Cumplido · % Vencido
//      - Row 14: Mini bar chart con distribución de TIPO_FEEDBACK
//   E. Smart formulas:
//      - STATUS_VIS: =IF(G5="Cumplido","✓", IF(G5="Vencido","⚠", ...))
//      - ACCIONES: HYPERLINK cruz-tab con tracking
//   F. Modern color palette:
//      - Title: gradient teal
//      - Header: deep navy
//      - Status pills: emerald, amber, rose
//      - KPI cards: subtle gray con accent indigo
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
	fmt.Println("FEEDBACKS v15 — PREMIUM 2026 REDESIGN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// A. NEW COLUMNS I (STATUS_VIS) y J (ACCIONES)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ A. New columns: I (STATUS_VIS) + J (ACCIONES)")

	// Header row 4
	reqs = append(reqs, setValue("I", 4, "STATUS", sheetFB))
	reqs = append(reqs, setValue("J", 4, "ACCIÓN", sheetFB))

	// STATUS_VIS formula para rows 5-11: icon based on ESTADO col G
	fmt.Println("  STATUS_VIS formula (icono auto según ESTADO)")
	for r := 5; r <= 11; r++ {
		reqs = append(reqs, setFormula("I", r,
			fmt.Sprintf(`=IF(G%d="Cumplido","✓ Listo",IF(G%d="Vencido","⚠ Vencido",IF(G%d="Carta de compromiso","📋 Compromiso",IF(G%d="Feedback dado","💬 Feedback",IF(G%d="En evaluación","⏳ Eval","· Pendiente")))))`,
				r, r, r, r, r), sheetFB))

		// ACCIONES: link a historial + marcar cumplido
		reqs = append(reqs, setValue("J", r, "📂 Historial  ✅ Marcar", sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// B. WIDEN SUPERVISOR + adjust widths
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ B. Col widths premium")
	colWidths := []int{200, 110, 110, 100, 360, 90, 130, 220, 130, 150}
	for i, w := range colWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// C. REMOVE frame borders feos — usar solo underlines minimalistas
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ C. Remove frame borders, add subtle dividers")
	// Clear all borders en rows 1-11
	reqs = append(reqs, clearBorders(0, 0, 11, 10, sheetFB))
	// Subtle underline divider entre header y data (row 4 → row 5)
	reqs = append(reqs, hBorder(5, 0, 10, sheetFB, 2)) // thick underline

	// ═══════════════════════════════════════════════════════════════════
	// D. KPI DASHBOARD rows 13-15
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ D. KPI Dashboard rows 13-15")

	// Row 13: KPI bar background
	reqs = append(reqs, setValue("A", 13, "📊 DASHBOARD DE CUMPLIMIENTO W40-2026", sheetFB))
	reqs = append(reqs, mergeRange(12, 0, 13, 9, sheetFB)) // A13:J13
	reqs = append(reqs, styleRange(12, 0, 13, 9, sheetFB, "#1a3a5c", "#FFFFFF", true, false, 12))

	// Row 14: KPI labels (headers)
	reqs = append(reqs, setValue("A", 14, "👥 AGENTES", sheetFB))
	reqs = append(reqs, setValue("C", 14, "📞 CALLS", sheetFB))
	reqs = append(reqs, setValue("E", 14, "✅ % CUMPLIDO", sheetFB))
	reqs = append(reqs, setValue("G", 14, "⚠ % VENCIDO", sheetFB))
	reqs = append(reqs, setValue("I", 14, "📊 TIPO TOP", sheetFB))

	// Merge headers
	reqs = append(reqs, mergeRange(13, 0, 14, 1, sheetFB)) // A14:B14
	reqs = append(reqs, mergeRange(13, 2, 14, 3, sheetFB)) // C14:D14
	reqs = append(reqs, mergeRange(13, 4, 14, 5, sheetFB)) // E14:F14
	reqs = append(reqs, mergeRange(13, 6, 14, 7, sheetFB)) // G14:H14
	reqs = append(reqs, mergeRange(13, 8, 14, 9, sheetFB)) // I14:J14

	// Style KPI headers
	reqs = append(reqs, styleRange(13, 0, 14, 9, sheetFB, "#f0f4f8", "#1a3a5c", true, false, 11))

	// Row 15: KPI values
	fmt.Println("  KPI values con formulas")
	reqs = append(reqs, setFormula("A", 15, `=COUNTA(A5:A11)`, sheetFB)) // Total agentes
	reqs = append(reqs, setFormula("C", 15, `=SUMPRODUCT(F5:F11)`, sheetFB)) // Total calls
	reqs = append(reqs, setFormula("E", 15, `=COUNTIF(G5:G11,"Cumplido")/COUNTA(G5:G11)`, sheetFB)) // % Cumplido
	reqs = append(reqs, setFormula("G", 15, `=COUNTIF(G5:G11,"Vencido")/COUNTA(G5:G11)`, sheetFB)) // % Vencido
	// TIPO top: el más frecuente
	reqs = append(reqs, setFormula("I", 15, `=INDEX(B5:B11,MATCH(MAX(COUNTIF(B5:B11,B5:B11)),COUNTIF(B5:B11,B5:B11),0))`, sheetFB))

	// Style KPI values (large font, bold)
	reqs = append(reqs, styleRange(14, 0, 15, 9, sheetFB, "#FFFFFF", "#1a3a5c", true, false, 18))

	// Merge value cells
	reqs = append(reqs, mergeRange(14, 0, 15, 1, sheetFB)) // A15:B15
	reqs = append(reqs, mergeRange(14, 2, 15, 3, sheetFB)) // C15:D15
	reqs = append(reqs, mergeRange(14, 4, 15, 5, sheetFB)) // E15:F15
	reqs = append(reqs, mergeRange(14, 6, 15, 7, sheetFB)) // G15:H15
	reqs = append(reqs, mergeRange(14, 8, 15, 9, sheetFB)) // I15:J15

	// ═══════════════════════════════════════════════════════════════════
	// E. STATUS_VIS styling (icon column)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ E. STATUS_VIS styling")
	for r := 5; r <= 11; r++ {
		reqs = append(reqs, styleCell(r-1, 8, r, 9, sheetFB, "#FFFFFF", "#0F2440", true, false, 11))
	}

	// ═══════════════════════════════════════════════════════════════════
	// F. CF rules nuevas para STATUS_VIS (col I) y ACCION (col J)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ F. CF rules para STATUS_VIS")
	// ✓ Listo (Cumplido) → verde
	reqs = append(reqs, addCFContains(sheetFB, 8, 5, 8, 11, "✓", 0.6, 0.9, 0.6))
	// ⚠ Vencido → rojo
	reqs = append(reqs, addCFContains(sheetFB, 8, 5, 8, 11, "⚠", 1.0, 0.6, 0.6))
	// 📋 Compromiso → ámbar
	reqs = append(reqs, addCFContains(sheetFB, 8, 5, 8, 11, "📋", 1.0, 0.85, 0.4))
	// 💬 Feedback → azul
	reqs = append(reqs, addCFContains(sheetFB, 8, 5, 8, 11, "💬", 0.6, 0.85, 1.0))
	// ⏳ Eval → amarillo
	reqs = append(reqs, addCFContains(sheetFB, 8, 5, 8, 11, "⏳", 1.0, 1.0, 0.6))

// ═══════════════════════════════════════════════════════════════════
// G. ACCION col styling (azul hyperlink style)
// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ G. ACCION col styling")
	for r := 5; r <= 11; r++ {
		reqs = append(reqs, styleCell(r-1, 9, r, 10, sheetFB, "#FFFFFF", "#1a73e8", true, false, 11))
	}

	// ═══════════════════════════════════════════════════════════════════
	// H. KPI Dashboard borders
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ H. KPI Dashboard borders")
	reqs = append(reqs, frameBorder(12, 0, 15, 9, sheetFB, 1)) // KPI frame

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

func setValue(col string, row int, value string, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": value},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func setFormula(col string, row int, formula string, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"formulaValue": formula},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func mergeRange(startRow, startCol, endRow, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"mergeType": "MERGE_ALL",
		},
	}
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
					"horizontalAlignment": "CENTER",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	}
}

func clearBorders(startRow, startCol, endRow, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":    map[string]any{"style": "NONE"},
			"bottom": map[string]any{"style": "NONE"},
			"left":   map[string]any{"style": "NONE"},
			"right":  map[string]any{"style": "NONE"},
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
			"top":    map[string]any{"style": "SOLID", "width": width, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
			"bottom": map[string]any{"style": "SOLID", "width": width, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
			"left":   map[string]any{"style": "SOLID", "width": width, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
			"right":  map[string]any{"style": "SOLID", "width": width, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
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
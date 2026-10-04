// restructure_feedbacks_v17_kpi5_removal.go — Eliminar KPI 5 (ÚLTIMO TIPO)
//
// CEO directive 2026-10-04 round 12:
// - "Eliminar ULTIMO TIPO EVALUATIVO no aporta nada"
// - "luego corregir y encuadrar bien la tabla de datos todas las row deben estar disponibles"
//
// Cambios:
//   1. Clear KPI 5 label (row 2 col I) y value (row 3 col I)
//   2. Unmerge KPI 5 merge range
//   3. ENCUADRAR rows 13-30 con zebra striping + borders sutiles
//      (estos rows estaban vacíos sin styling después de los shifts de v16)
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
	fmt.Println("FEEDBACKS v17 — Eliminar KPI 5 + Encuadrar tabla")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// A. ELIMINAR KPI 5 (ÚLTIMO TIPO) — rows 2-3 col I (idx 8)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ A. Eliminar KPI 5 (ÚLTIMO TIPO)")

	// A1: Clear I2 (label) y I3 (value)
	reqs = append(reqs, clearCell("I", 2, sheetFB))
	reqs = append(reqs, clearCell("I", 3, sheetFB))

	// A2: Unmerge los 2 merges del KPI 5 (label y value exactos del diag)
	// Merge 1: rows 1-2 cols 8-10 (KPI 5 label "ÚLTIMO TIPO")
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    1,
				"endRowIndex":      2,
				"startColumnIndex": 8,
				"endColumnIndex":   10,
			},
		},
	})
	// Merge 2: rows 2-3 cols 8-10 (KPI 5 value "Evaluativo")
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    2,
				"endRowIndex":      3,
				"startColumnIndex": 8,
				"endColumnIndex":   10,
			},
		},
	})

	// Also unmerge rows 12-13 cols 0-9 (merge residual)
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    12,
				"endRowIndex":      13,
				"startColumnIndex": 0,
				"endColumnIndex":   9,
			},
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// B. ENCUADRAR TABLA: rows 13-30 con zebra + borders
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ B. Encuadrar rows 13-30 con zebra striping + borders")

	for r := 13; r <= 30; r++ {
		// Zebra alterna
		bg := "#FFFFFF"
		if r%2 == 1 {
			bg = "#fafcfe"
		}
		// Aplicar background color + text format (gris claro para "empty placeholder")
		reqs = append(reqs, styleRange(r-1, 0, r, 8, sheetFB, bg, "#0F2440", false, false, 10))
	}

	// Add subtle bottom border to data rows (12) separator
	// And top border to first placeholder row (13)
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    12,
				"endRowIndex":      13,
				"startColumnIndex": 0,
				"endColumnIndex":   8,
			},
			"top": map[string]any{"style": "SOLID", "width": 2, "color": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36}},
		},
	})

	// Subtle horizontal separators between placeholder rows
	for r := 14; r <= 30; r++ {
		reqs = append(reqs, map[string]any{
			"updateBorders": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetFB,
					"startRowIndex":    r - 1,
					"endRowIndex":      r,
					"startColumnIndex": 0,
					"endColumnIndex":   8,
				},
				"bottom": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.92, "green": 0.94, "blue": 0.96}},
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// C. EXPAND COL widths for new layout (8 cols visible)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ C. Col widths optimizadas (8 cols)")
	colWidths := []int{200, 110, 110, 100, 360, 90, 130, 220}
	for i, w := range colWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// D. Free up rows 13-30 para INPUT (clear any leftover CF on those rows)
	// ═══════════════════════════════════════════════════════════════════
	// (CF rules already point to rows 9-27 from v16d, we don't need to clear more)

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

func clearCell(col string, row int, sheetID int64) map[string]any {
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
							"userEnteredValue": map[string]any{"stringValue": ""},
						},
					},
				},
			},
			"fields": "userEnteredValue",
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
					"horizontalAlignment": "LEFT",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
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
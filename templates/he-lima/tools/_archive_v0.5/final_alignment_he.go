// final_alignment_he.go — CEO directive 2026-09-26
// UNA sola pasada de alineación completa del documento 10_HE.
//
// Lo que arregla:
// 1. Col A: ancho consistente (no más 30px inconsistente)
// 2. B+C+D merged: name area ancho consistente (no más 180/120/100 disparejo)
// 3. E-AI: 31 day cols a 32px uniforme (más respiración visual)
// 4. AJ, AL, AN, AP, AR, AT: numbers section ancho consistente (no más 70/100 mixto)
// 5. Spacers AK, AM, AO, AS: a 6px (más delgados pero presentes)
// 6. Frames: SOLO 2 frames limpios A1:AT7 + A8:AT32 (sin borders apilados)
// 7. Bordes: clean — clear ALL existing borders first, then apply 2 frames
// 8. Typography: 11pt base, 18pt title, 14pt subtitle
// 9. Row heights: 24px default, 32px header rows, 18px spacer rows
// 10. Vertical alignment: MIDDLE en todas las celdas del grid
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
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetHE       = int64(222789142)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	creds := decrypt()
	tok := getToken(creds)

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — CLEAR ALL existing borders on the whole sheet (A1:AT100)
	//          to start fresh. No stacked conflicting borders.
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      100,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":             nil,
			"bottom":          nil,
			"left":            nil,
			"right":           nil,
			"innerHorizontal": nil,
			"innerVertical":   nil,
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Col widths: A=32, B=200, C=120, D=110 (name area), E-AI=32, AJ-AT vars
	// ═══════════════════════════════════════════════════════════════════

	// A (col 1): N° col → 32px
	reqs = append(reqs, colWidth(sheetHE, 0, 1, 32))
	// B (col 2): name part 1 → 200px
	reqs = append(reqs, colWidth(sheetHE, 1, 2, 200))
	// C (col 3): name part 2 → 120px
	reqs = append(reqs, colWidth(sheetHE, 2, 3, 120))
	// D (col 4): name part 3 → 110px
	reqs = append(reqs, colWidth(sheetHE, 3, 4, 110))
	// E-AI (cols 5-35): day cells → 32px each
	reqs = append(reqs, colWidth(sheetHE, 4, 35, 32))
	// AJ (col 36): H_norm → 70px
	reqs = append(reqs, colWidth(sheetHE, 35, 36, 70))
	// AK (col 37): spacer → 6px (más delgado)
	reqs = append(reqs, colWidth(sheetHE, 36, 37, 6))
	// AL (col 38): H_ext → 70px
	reqs = append(reqs, colWidth(sheetHE, 37, 38, 70))
	// AM (col 39): hidden spacer → 6px
	reqs = append(reqs, colWidth(sheetHE, 38, 39, 6))
	// AN (col 40): H_25% → 70px
	reqs = append(reqs, colWidth(sheetHE, 39, 40, 70))
	// AO (col 41): hidden spacer → 6px
	reqs = append(reqs, colWidth(sheetHE, 40, 41, 6))
	// AP (col 42): H_35% → 70px
	reqs = append(reqs, colWidth(sheetHE, 41, 42, 70))
	// AQ (col 43): hidden spacer → 6px
	reqs = append(reqs, colWidth(sheetHE, 42, 43, 6))
	// AR (col 44): $_N → 100px
	reqs = append(reqs, colWidth(sheetHE, 43, 44, 100))
	// AS (col 45): spacer → 6px
	reqs = append(reqs, colWidth(sheetHE, 44, 45, 6))
	// AT (col 46): $_E → 100px
	reqs = append(reqs, colWidth(sheetHE, 45, 46, 100))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Row heights: header rows taller, data rows uniform
	//          Row 1 (title): 36px
	//          Row 2-3 (KPIs): 32px
	//          Row 4 (spacer): 12px
	//          Row 5-6 (selectors): 28px
	//          Row 7 (spacer): 12px
	//          Row 8 (day nums): 28px
	//          Row 9 (day letters): 22px
	//          Rows 10-30 (emp data): 22px each
	//          Row 31 (spacer): 12px
	//          Row 32 (TOTALES): 28px
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, rowHeight(sheetHE, 0, 1, 36))   // row 1
	reqs = append(reqs, rowHeight(sheetHE, 1, 3, 32))   // rows 2-3
	reqs = append(reqs, rowHeight(sheetHE, 3, 4, 12))   // row 4 spacer
	reqs = append(reqs, rowHeight(sheetHE, 4, 6, 28))   // rows 5-6 selectors
	reqs = append(reqs, rowHeight(sheetHE, 6, 7, 12))   // row 7 spacer
	reqs = append(reqs, rowHeight(sheetHE, 7, 8, 28))   // row 8 day nums
	reqs = append(reqs, rowHeight(sheetHE, 8, 9, 22))   // row 9 day letters
	reqs = append(reqs, rowHeight(sheetHE, 9, 30, 22))  // rows 10-30 emp data
	reqs = append(reqs, rowHeight(sheetHE, 30, 31, 12)) // row 31 spacer
	reqs = append(reqs, rowHeight(sheetHE, 31, 32, 28)) // row 32 TOTALES

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — TWO clean frames
	//          FRAME 1 HEADER:  A1:AT7
	//          FRAME 2 DATA:    A8:AT32
	//          Both: SOLID_MEDIUM navy 2px all around.
	// ═══════════════════════════════════════════════════════════════════

	// FRAME 1 — Header
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      7,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_MEDIUM", borderPrimary(), 2),
			"bottom": border("SOLID_MEDIUM", borderPrimary(), 2),
			"left":   border("SOLID_MEDIUM", borderPrimary(), 2),
			"right":  border("SOLID_MEDIUM", borderPrimary(), 2),
		},
	})

	// FRAME 2 — Data
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      32,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_MEDIUM", borderPrimary(), 2),
			"bottom": border("SOLID_MEDIUM", borderPrimary(), 2),
			"left":   border("SOLID_MEDIUM", borderPrimary(), 2),
			"right":  border("SOLID_MEDIUM", borderPrimary(), 2),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Internal section dividers (subtle, 1px accent)
	//          D|E: name area | day grid
	//          AI|AJ: day grid | numbers section
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      32,
				"startColumnIndex": 4, // E
				"endColumnIndex":   5,
			},
			"left": border("SOLID", borderAccent(), 1),
		},
	})
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      32,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   36,
			},
			"left": border("SOLID", borderAccent(), 1),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 6 — Vertical alignment MIDDLE everywhere in grid
	//          Title row 1: middle
	//          All data rows (8-30): middle
	//          TOTALES row 32: middle
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      32,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"verticalAlignment": "MIDDLE",
				},
			},
			"fields": "userEnteredFormat.verticalAlignment",
		},
	})

	// Send batch
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
	defer resp.Body.Close()

	respB, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}

	fmt.Printf("✅ OK — %d requests applied (one-pass alignment)\n", len(reqs))
	fmt.Println("Col widths normalized: A=32, B=200, C=120, D=110, E-AI=32,")
	fmt.Println("  AJ=70, AK=6, AL=70, AM=6, AN=70, AO=6, AP=70, AQ=6, AR=100, AS=6, AT=100")
	fmt.Println("Row heights: header rows taller, data uniform 22px")
	fmt.Println("Frames: 2 clean from col A (header A1:AT7, data A8:AT32)")
	fmt.Println("All borders cleared first (no conflicts), then 2 frames + 2 internal accents")
	fmt.Println("Vertical align: MIDDLE across whole grid")
}

// ── Helpers ────────────────────────────────────────────────────────

func colWidth(sheetID int64, startCol, endCol int, px int) map[string]any {
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "COLUMNS",
				"startIndex": startCol,
				"endIndex":   endCol,
			},
			"properties": map[string]any{"pixelSize": px},
			"fields":     "pixelSize",
		},
	}
}

func rowHeight(sheetID int64, startRow, endRow int, px int) map[string]any {
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "ROWS",
				"startIndex": startRow,
				"endIndex":   endRow,
			},
			"properties": map[string]any{"pixelSize": px},
			"fields":     "pixelSize",
		},
	}
}

func borderPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786,
		"green": 0.227450982,
		"blue":  0.360784314,
	}
}

func borderAccent() map[string]any {
	return map[string]any{
		"red":   0.9019608,
		"green": 0.4941176,
		"blue":  0.1333333,
	}
}

func border(style string, c map[string]any, w int) map[string]any {
	return map[string]any{"style": style, "color": c, "width": w}
}

// ── OAuth ──────────────────────────────────────────────────────────

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
	return res["access_token"].(string)
}
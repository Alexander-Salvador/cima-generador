// fix_p2_two_frames_he.go — CEO directive 2026-09-26
// "row 1 a row 7 todo ese contenido enmarcado es como un 'header'
//  de la row 8 hasta lo ultimo enmarcalo bien
//  ambos marcos deben empezar desde la column A de manera inteligente"
//
// Design: TWO clean frames stacked vertically:
//   Frame 1 (HEADER):  A1:AT7   rows 1-7
//   Frame 2 (DATA):    A8:AT32  rows 8-32
//
// Both start at col A, span full width to AT (col 46).
// Thick navy borders all around. No internal clutter.
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
	// STEP 1 — Clear any existing borders on these ranges first.
	//          (Multiple updateBorders requests on same range stack,
//           so we need to explicitly remove old borders.)
	// ═══════════════════════════════════════════════════════════════════

	// Clear borders on A1:AT7 (old header frame from fix_three_bugs_he.go)
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      7,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    nil,
			"bottom": nil,
			"left":   nil,
			"right":  nil,
			"innerHorizontal": nil,
			"innerVertical":   nil,
		},
	})

	// Clear borders on A8:AT34 (old data/numbers frames from polish_and_frame_he.go)
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      34,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    nil,
			"bottom": nil,
			"left":   nil,
			"right":  nil,
			"innerHorizontal": nil,
			"innerVertical":   nil,
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — FRAME 1 (HEADER): A1:AT7, rows 1-7
	//   Thick navy border all around. This is the "header" frame:
	//   - Title (row 1, navy bg already)
	//   - KPI cards (rows 2-3)
	//   - Spacer (row 4)
	//   - Año selector (row 5)
	//   - Mes selector (row 6)
	//   - Spacer (row 7)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0, // row 1
				"endRowIndex":      7, // row 7 (exclusive)
				"startColumnIndex": 0, // col A
				"endColumnIndex":   46, // col AT (exclusive = 46)
			},
			"top":    border("SOLID_MEDIUM", borderPrimary(), 2),
			"bottom": border("SOLID_MEDIUM", borderPrimary(), 2),
			"left":   border("SOLID_MEDIUM", borderPrimary(), 2),
			"right":  border("SOLID_MEDIUM", borderPrimary(), 2),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — FRAME 2 (DATA): A8:AT32, rows 8-32
	//   Thick navy border all around. This is the "data" frame:
	//   - Day numbers (row 8)
	//   - Day letters (row 9)
	//   - Employees with data (rows 10-25)
	//   - Employees without data (rows 26-30)
	//   - Spacer (row 31)
	//   - TOTALES (row 32)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,  // row 8
				"endRowIndex":      32, // row 32 (TOTALES)
				"startColumnIndex": 0,  // col A
				"endColumnIndex":   46, // col AT (exclusive = 46)
			},
			"top":    border("SOLID_MEDIUM", borderPrimary(), 2),
			"bottom": border("SOLID_MEDIUM", borderPrimary(), 2),
			"left":   border("SOLID_MEDIUM", borderPrimary(), 2),
			"right":  border("SOLID_MEDIUM", borderPrimary(), 2),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Subtle internal accent: vertical line between data section
	//          and numbers section inside the data frame.
	//          Col AI (col 35) right edge / col AJ (col 36) left edge.
	//          Subtle accent orange to visually separate, NOT a thick
	//          border (clean look).
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,  // row 8
				"endRowIndex":      32, // row 32 (TOTALES)
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   36,
			},
			"left": border("SOLID", borderAccent(), 1),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Vertical divider between name area and day area (D|E).
	//          Subtle accent (same as step 4) for visual cohesion.
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,  // row 8
				"endRowIndex":      32, // row 32
				"startColumnIndex": 4,  // E
				"endColumnIndex":   5,
			},
			"left": border("SOLID", borderAccent(), 1),
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

	fmt.Printf("✅ OK — %d requests applied\n", len(reqs))
	fmt.Println("Two clean frames:")
	fmt.Println("  Frame 1 (HEADER):  A1:AT7   — title + KPIs + selectors")
	fmt.Println("  Frame 2 (DATA):    A8:AT32  — day headers + 21 emp rows + TOTALES")
	fmt.Println("  Internal accent lines: D|E (name|day) + AI|AJ (day|numbers)")
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
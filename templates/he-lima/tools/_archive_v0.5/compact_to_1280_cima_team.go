// compact_to_1280_cima_team.go — CEO directive 2026-09-27
// Compact CIMA TEAM to fit in ~1280px viewport.
//
// New layout (total ~1302px, fits typical laptop):
//   A (N°):           26px
//   B (name 1):       75px
//   C (name 2):       70px
//   D (name 3):       60px
//   E-AH (30 days):   24px × 30 = 720px
//   AI (spacer):      6px
//   AJ (H_norm):      50px
//   AK (spacer):      5px
//   AL (H_ext):       50px
//   AM (spacer):      5px
//   AN (H_25%):       50px
//   AO (spacer):      5px
//   AP (H_35%):       50px
//   AQ (spacer):      5px
//   AR ($_N):         60px
//   AS (spacer):      5px
//   AT ($_E):         60px
//
// Also reduces font sizes to 10pt for compact cells.
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
	// STEP 1 — Col widths (compact to ~1300px)
	// ═══════════════════════════════════════════════════════════════════
	widths := []struct {
		start, end, px int
	}{
		{0, 1, 26},   // A
		{1, 2, 75},   // B
		{2, 3, 70},   // C
		{3, 4, 60},   // D
		{4, 34, 24},  // E-AH (30 day cols)
		{34, 35, 6},  // AI (spacer)
		{35, 36, 50}, // AJ
		{36, 37, 5},  // AK (spacer)
		{37, 38, 50}, // AL
		{38, 39, 5},  // AM (spacer)
		{39, 40, 50}, // AN
		{40, 41, 5},  // AO (spacer)
		{41, 42, 50}, // AP
		{42, 43, 5},  // AQ (spacer)
		{43, 44, 60}, // AR
		{44, 45, 5},  // AS (spacer)
		{45, 46, 60}, // AT
	}
	for _, w := range widths {
		reqs = append(reqs, map[string]any{
			"updateDimensionProperties": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetHE,
					"dimension":  "COLUMNS",
					"startIndex": w.start,
					"endIndex":   w.end,
				},
				"properties": map[string]any{"pixelSize": w.px},
				"fields":     "pixelSize",
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Reduce font sizes for compact cells
	//          Day numbers (row 8), day letters (row 9), data (10-30), TOTALES (36)
	// ═══════════════════════════════════════════════════════════════════

	// Row 8 (day numbers): 11pt → 10pt
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      8,
				"startColumnIndex": 4,
				"endColumnIndex":   34,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"textFormat": map[string]any{"fontSize": 10, "bold": true},
				},
			},
			"fields": "userEnteredFormat.textFormat",
		},
	})

	// Row 9 (day letters): 11pt → 10pt
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    8,
				"endRowIndex":      9,
				"startColumnIndex": 4,
				"endColumnIndex":   34,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"textFormat": map[string]any{"fontSize": 10, "bold": true},
				},
			},
			"fields": "userEnteredFormat.textFormat",
		},
	})

	// Data rows 10-25: 10pt stays (was 10pt)
	// Empty rows 26-34: 10pt stays

	// TOTALES row 36: 11pt → 11pt (bold)
	// (already correct)

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Title: more compact font (16pt instead of 18pt)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   35,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"textFormat": map[string]any{"fontSize": 16, "bold": true},
				},
			},
			"fields": "userEnteredFormat.textFormat",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Row heights: reduce for compactness
	// ═══════════════════════════════════════════════════════════════════
	heights := []struct {
		start, end, px int
	}{
		{0, 1, 32},   // row 1 title (was 40)
		{1, 3, 24},   // rows 2-3 KPI (was 32)
		{3, 4, 8},    // row 4 spacer (was 12)
		{4, 6, 22},   // rows 5-6 selectors (was 28)
		{6, 7, 8},    // row 7 spacer (was 12)
		{7, 8, 22},   // row 8 day nums (was 28)
		{8, 9, 18},   // row 9 day letters (was 22)
		{9, 34, 18},  // rows 10-33 emp data (was 22)
		{34, 35, 8},  // row 34 spacer
		{35, 36, 22}, // row 36 TOTALES (was 28)
	}
	for _, h := range heights {
		reqs = append(reqs, map[string]any{
			"updateDimensionProperties": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetHE,
					"dimension":  "ROWS",
					"startIndex": h.start,
					"endIndex":   h.end,
				},
				"properties": map[string]any{"pixelSize": h.px},
				"fields":     "pixelSize",
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Make frame borders more visible (different color: accent
	//          orange for HEADER frame so it stands out vs default navy)
	// ═══════════════════════════════════════════════════════════════════
	// Actually, keep navy but increase to 4px for visibility
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      7,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderPrimary(), 4),
			"bottom": border("SOLID_THICK", borderPrimary(), 4),
			"left":   border("SOLID_THICK", borderPrimary(), 4),
			"right":  border("SOLID_THICK", borderPrimary(), 4),
		},
	})
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderAccent(), 4),
			"bottom": border("SOLID_THICK", borderAccent(), 4),
			"left":   border("SOLID_THICK", borderAccent(), 4),
			"right":  border("SOLID_THICK", borderAccent(), 4),
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
	fmt.Println("Layout compacted to ~1302px:")
	fmt.Println("  Col A: 26px | B:75 | C:70 | D:60 | E-AH:24×30=720 | AI:6")
	fmt.Println("  AJ:50 | AL:50 | AN:50 | AP:50 | AR:60 | AT:60")
	fmt.Println("  Spacers (AK,AM,AO,AQ,AS): 5px each")
	fmt.Println("Font sizes: day nums 10pt, day letters 10pt, title 16pt")
	fmt.Println("Row heights: title 32px, KPI 24px, day nums 22px, data 18px")
	fmt.Println("Frame borders: HEADER navy 4px, DATA accent (orange) 4px")
}

// ── Tokens ─────────────────────────────────────────────────────────

func borderPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786, "green": 0.227450982, "blue": 0.360784314,
	}
}
func borderAccent() map[string]any {
	return map[string]any{
		"red":   0.9019608, "green": 0.4941176, "blue": 0.1333333,
	}
}
func border(style string, c map[string]any, w int) map[string]any {
	return map[string]any{"style": style, "color": c, "width": w}
}

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
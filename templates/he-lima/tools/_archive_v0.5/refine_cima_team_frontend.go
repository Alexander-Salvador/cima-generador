// refine_cima_team_frontend.go — CEO directive 2026-09-27
// Refine CIMA TEAM frontend after adding 4 missing employees.
//
// Steps:
// 1. Style new empty rows (31-34): muted bg + italic muted fg
//    (matching existing empty rows 26-30)
// 2. Extend data frame border: was A8:AT32 → A8:AT36 (includes
//    all 25 emp rows + spacer + TOTALES)
// 3. Clear any old borders leftover from previous frames
// 4. Apply clean 2-frame structure:
//    - Frame 1 HEADER: A1:AT7
//    - Frame 2 DATA:   A8:AT36 (covers all 25 emp rows + spacer + TOTALES)
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
	// STEP 1 — Style new empty rows (31-34): muted bg + italic muted fg
	//          Day area (E-AH = cols 4-34, 0-based)
	//          Name area (A-D = cols 0-4)
	//          Numbers area (AJ-AT = cols 35-46)
	// ═══════════════════════════════════════════════════════════════════

	// Day area E-AH (rows 31-34)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    30, // row 31
				"endRowIndex":      34, // row 34 exclusive
				"startColumnIndex": 4,  // E
				"endColumnIndex":   34, // AH
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgMuted(),
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// Name area A-D (rows 31-34)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    30,
				"endRowIndex":      34,
				"startColumnIndex": 0, // A
				"endColumnIndex":   5, // E exclusive = D inclusive
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgMuted(),
					"textFormat": map[string]any{
						"foregroundColor": fgMuted(),
						"italic":          true,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
		},
	})

	// Numbers area AJ-AT (rows 31-34)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    30,
				"endRowIndex":      34,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   46, // AT
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgAlertSoftMuted(),
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — FORCE-CLEAR all borders (per-row) to remove artifacts
	// ═══════════════════════════════════════════════════════════════════
	for row := 0; row < 40; row++ {
		reqs = append(reqs, map[string]any{
			"updateBorders": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row,
					"endRowIndex":      row + 1,
					"startColumnIndex": 0,
					"endColumnIndex":   46,
				},
				"top":             map[string]any{"style": "NONE"},
				"bottom":          map[string]any{"style": "NONE"},
				"left":            map[string]any{"style": "NONE"},
				"right":           map[string]any{"style": "NONE"},
				"innerHorizontal": map[string]any{"style": "NONE"},
				"innerVertical":   map[string]any{"style": "NONE"},
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Apply 2 clean frames (NOW EXTENDED to row 36)
	//          FRAME 1 HEADER: A1:AT7
	//          FRAME 2 DATA:   A8:AT36 (covers 25 emps + spacer + TOTALES)
	// ═══════════════════════════════════════════════════════════════════

	// FRAME 1 — HEADER
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      7,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderPrimary(), 3),
			"bottom": border("SOLID_THICK", borderPrimary(), 3),
			"left":   border("SOLID_THICK", borderPrimary(), 3),
			"right":  border("SOLID_THICK", borderPrimary(), 3),
		},
	})

	// FRAME 2 — DATA (extended to row 36)
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36, // row 36 (TOTALES) inclusive
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderPrimary(), 3),
			"bottom": border("SOLID_THICK", borderPrimary(), 3),
			"left":   border("SOLID_THICK", borderPrimary(), 3),
			"right":  border("SOLID_THICK", borderPrimary(), 3),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Internal accent: subtle line at D|E (name | day grid)
	//          and AI|AJ (day grid | numbers section)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36,
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
				"endRowIndex":      36,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   36,
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
	fmt.Println("Refinements applied:")
	fmt.Println("  - New empty rows (31-34) styled to match 26-30 (muted bg)")
	fmt.Println("  - All borders cleared per-row (no artifacts)")
	fmt.Println("  - Frame 1 HEADER: A1:AT7 (SOLID_THICK navy 3px)")
	fmt.Println("  - Frame 2 DATA:   A8:AT36 (extended for 25 employees + TOTALES)")
	fmt.Println("  - Internal accents at D|E and AI|AJ (subtle 1px)")
}

// ── Tokens ─────────────────────────────────────────────────────────

func bgMuted() map[string]any {
	return map[string]any{
		"red":   0.9176471,
		"green": 0.9294118,
		"blue":  0.9450980,
	}
}

func bgAlertSoftMuted() map[string]any {
	return map[string]any{
		"red":   0.9647059,
		"green": 0.9176471,
		"blue":  0.8588235,
	}
}

func fgMuted() map[string]any {
	return map[string]any{
		"red":   0.4509804,
		"green": 0.5176471,
		"blue":  0.6000000,
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
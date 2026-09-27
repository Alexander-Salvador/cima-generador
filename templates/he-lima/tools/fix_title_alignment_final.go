// fix_title_alignment_final.go — CEO directive 2026-09-27
// "todo mal end to end mal diseñado"
//
// Root cause: Title merge is B1:AI1 (offset by 1 col). Should be A1:AI1.
//
// Also: VALORIZACIÓN should align with H_norm (col AJ), not start at col AK.
//
// This script does ONE clean pass to fix both.
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
	// STEP 1 — Unmerge title (was B1:AI1 = cols 1-35 0-based)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 1, // B
				"endColumnIndex":   35, // AI+1
			},
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Merge title A1:AI1 (cols 0-35 0-based)
	//          This INCLUDES col A so title aligns with N° col below
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0, // A
				"endColumnIndex":   35, // AI
			},
			"mergeType": "MERGE_ALL",
		},
	})

	// Re-write title content at A1
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": "CONTROL DE HORAS EXTRAS"},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	})

	// Re-style title: navy bg + white bold LEFT aligned + vertical middle
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
					"backgroundColor": bgBrandPrimary(),
					"textFormat": map[string]any{
						"foregroundColor": fgInverse(),
						"fontSize":        18,
						"bold":            true,
					},
					"horizontalAlignment": "LEFT",
					"verticalAlignment":   "MIDDLE",
					"padding": map[string]any{
						"left":   12,
						"right":  12,
						"top":    8,
						"bottom": 8,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,padding)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Re-style VALORIZACIÓN header (AJ1:AT1)
	//          Make sure it aligns with H_norm column below (AJ)
	//          Navy bg + white bold CENTER + vertical middle
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 35,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgBrandPrimary(),
					"textFormat": map[string]any{
						"foregroundColor": fgInverse(),
						"fontSize":        14,
						"bold":            true,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"padding": map[string]any{
						"left":   8,
						"right":  8,
						"top":    8,
						"bottom": 8,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,padding)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Re-style CIMA TEAM label (rows 8-9, cols B-D)
	//          Should be visually distinct header, navy bg + white bold
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      9,
				"startColumnIndex": 1, // B
				"endColumnIndex":   5, // E
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgBrandPrimary(),
					"textFormat": map[string]any{
						"foregroundColor": fgInverse(),
						"fontSize":        13,
						"bold":            true,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Apply row 1 height (36px) for title visibility
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetHE,
				"dimension":  "ROWS",
				"startIndex": 0,
				"endIndex":   1,
			},
			"properties": map[string]any{"pixelSize": 40},
			"fields":     "pixelSize",
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
	fmt.Println("Title alignment fix:")
	fmt.Println("  - Unmerged B1:AI1, merged A1:AI1 (title now starts at col A)")
	fmt.Println("  - Navy bg + white bold 18pt LEFT aligned + padding 12px")
	fmt.Println("  - Row 1 height: 40px")
	fmt.Println("VALORIZACIÓN realigned:")
	fmt.Println("  - AJ1:AT1 stays, but restyled navy bg + white bold 14pt CENTER")
	fmt.Println("CIMA TEAM label restyled:")
	fmt.Println("  - Rows 8-9 cols B-E: navy bg + white bold 13pt CENTER")
}

// ── Tokens ─────────────────────────────────────────────────────────

func bgBrandPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786, "green": 0.227450982, "blue": 0.360784314,
	}
}
func fgInverse() map[string]any {
	return map[string]any{"red": 1, "green": 1, "blue": 1}
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
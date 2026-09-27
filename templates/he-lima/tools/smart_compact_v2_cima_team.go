// smart_compact_v2_cima_team.go — CEO directive 2026-09-27
// Smart compaction based on real data analysis.
//
// Issues detected by verify_cima_team.go:
// - WIDTH 1302px > 1280px viewport
// - Duplicate conditional formats (=E$8="" x3, =E$9="dom"/"sáb" x2 each)
// - Day cells show "7.00" instead of "7" (wastes width visually)
//
// SMART CHANGES:
// 1. Reduce day cells 24px → 22px (saves 60px across 30 cols)
// 2. Reduce name area: B 75→60, C 70→55, D 60→50 (saves 40px)
// 3. Reduce spacers 5-6px → 2px (saves ~20px)
// 4. Number format "0" instead of "0.00" (cleaner look, more compact visual)
// 5. Cleanup duplicate conditional formats
//
// New total: ~1183px (fits 1280px viewport with 97px margin)
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

	// Get current rule IDs (in REVERSE order to safely delete)
	ruleIDs := getConditionalFormatRuleIDs(tok)
	// Reverse the slice
	for i, j := 0, len(ruleIDs)-1; i < j; i, j = i+1, j-1 {
		ruleIDs[i], ruleIDs[j] = ruleIDs[j], ruleIDs[i]
	}

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Delete all conditional formats (re-apply cleanly after)
	// ═══════════════════════════════════════════════════════════════════
	for _, id := range ruleIDs {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   id,
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Smart col widths (compact to ~1183px)
	// ═══════════════════════════════════════════════════════════════════
	widths := []struct {
		start, end, px int
	}{
		{0, 1, 26},   // A (N°)
		{1, 2, 60},   // B (name 1) 75→60
		{2, 3, 55},   // C (name 2) 70→55
		{3, 4, 50},   // D (name 3) 60→50
		{4, 34, 22},  // E-AH (days) 24→22
		{34, 35, 2},  // AI spacer 6→2
		{35, 36, 50}, // AJ H_norm
		{36, 37, 2},  // AK spacer 5→2
		{37, 38, 50}, // AL H_ext
		{38, 39, 2},  // AM spacer 5→2
		{39, 40, 50}, // AN H_25%
		{40, 41, 2},  // AO spacer 5→2
		{41, 42, 50}, // AP H_35%
		{42, 43, 2},  // AQ spacer 5→2
		{43, 44, 60}, // AR $_N
		{44, 45, 2},  // AS spacer 5→2
		{45, 46, 60}, // AT $_E
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
	// STEP 3 — Integer format on day cells (no ".00")
	//          Day cells in rows 10-25 (data) and row 8 (numbers)
	// ═══════════════════════════════════════════════════════════════════

	// Day data cells rows 10-25 (cols E-AH)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,
				"endRowIndex":      25,
				"startColumnIndex": 4,
				"endColumnIndex":   34,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"numberFormat": map[string]any{
						"type": "NUMBER",
						"pattern": "0",
					},
				},
			},
			"fields": "userEnteredFormat.numberFormat",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Re-apply clean conditional formats (3 rules only)
	//          a) Ghost day =E$8="" (rows 8-34 cols 5-34)
	//          b) Weekend sáb E$9="sáb" (row 9)
	//          c) Weekend dom E$9="dom" (row 9)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetHE,
						"startRowIndex":    7,
						"endRowIndex":      34,
						"startColumnIndex": 4,
						"endColumnIndex":   34,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": "=E$8=\"\""},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgMuted(),
						"textFormat": map[string]any{
							"foregroundColor": fgMuted(),
							"italic":          true,
							"strikethrough":   true,
						},
					},
				},
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetHE,
						"startRowIndex":    8,
						"endRowIndex":      9,
						"startColumnIndex": 4,
						"endColumnIndex":   34,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": "=E$9=\"sáb\""},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgAlertSoft(),
						"textFormat": map[string]any{
							"foregroundColor": fgPrimary(),
							"bold":            true,
						},
					},
				},
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetHE,
						"startRowIndex":    8,
						"endRowIndex":      9,
						"startColumnIndex": 4,
						"endColumnIndex":   34,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": "=E$9=\"dom\""},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgAlertSoft(),
						"textFormat": map[string]any{
							"foregroundColor": fgPrimary(),
							"bold":            true,
						},
					},
				},
			},
		},
	})

	// Send batch 1 (rules + widths + number format)
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
	fmt.Println("Smart compact v2:")
	fmt.Println("  Col widths compacted:")
	fmt.Println("    Day cells: 24 → 22px (saves 60px)")
	fmt.Println("    Name area: 75/70/60 → 60/55/50 (saves 40px)")
	fmt.Println("    Spacers: 5-6 → 2px (saves ~20px)")
	fmt.Println("    New total: ~1183px (fits 1280px)")
	fmt.Println("  Day data format: 0.00 → 0 (cleaner, integer only)")
	fmt.Println("  Conditional formats: 9 → 3 (no duplicates)")
}

// ── Helpers ────────────────────────────────────────────────────────

func getConditionalFormatRuleIDs(tok string) []int {
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?ranges=CIMA%%20TEAM&fields=sheets(conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)
	ids := []int{}
	sheets := out["sheets"].([]any)
	for _, s := range sheets {
		cf := s.(map[string]any)["conditionalFormats"].([]any)
		for i := range cf {
			ids = append(ids, i)
		}
	}
	return ids
}

func bgMuted() map[string]any {
	return map[string]any{
		"red":   0.9176471, "green": 0.9294118, "blue": 0.9450980,
	}
}
func bgAlertSoft() map[string]any {
	return map[string]any{
		"red":   0.9921569, "green": 0.9215686, "blue": 0.8156863,
	}
}
func fgPrimary() map[string]any {
	return map[string]any{
		"red":   0.0588235, "green": 0.1411765, "blue": 0.2509804,
	}
}
func fgMuted() map[string]any {
	return map[string]any{
		"red":   0.4509804, "green": 0.5176471, "blue": 0.6000000,
	}
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
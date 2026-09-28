// fix_fase2_formulas_v2_cima_team.go — Fix formulas after first pass.
//
// Issues:
// 1. Day letters are 1-char (D, L, M, X, J, V, S) not 3-char (dom, lun)
//    → H_25 formula "dom" check doesn't match
// 2. Holiday conditional format was placeholder ($E$8=TODAY())
//    → use proper VLOOKUP against 01_FERIADOS_PE
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

	// Get conditional format rule IDs to delete
	ruleIDs := getConditionalFormatRuleIDs(tok)

	reqs := []any{}

	// Delete existing conditional format rules
	for _, id := range ruleIDs {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   id,
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// Re-apply CORRECT conditional formats
	// ═══════════════════════════════════════════════════════════════════

	// 1) Holiday (gold bg + bold red text) - applies to day cells if date is holiday
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
							map[string]any{"userEnteredValue": "=ISNUMBER(VLOOKUP(E$8,INDIRECT(\"'01_FERIADOS_PE'!$A$2:$A$100\"),1,FALSE))"},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgHoliday(),
						"textFormat": map[string]any{
							"foregroundColor": fgHoliday(),
							"bold":            true,
						},
					},
				},
			},
		},
	})

	// 2) Weekend "S" (Sábado) - subtle bg for day letters row
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
							map[string]any{"userEnteredValue": "=E$9=\"S\""},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgWeekend(),
						"textFormat": map[string]any{
							"foregroundColor": fgPrimary(),
							"bold":            true,
						},
					},
				},
			},
		},
	})

	// 3) Weekend "D" (Domingo) - alert bg for day letters row
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
							map[string]any{"userEnteredValue": "=E$9=\"D\""},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgSunday(),
						"textFormat": map[string]any{
							"foregroundColor": fgPrimary(),
							"bold":            true,
						},
					},
				},
			},
		},
	})

	// 4) Ghost day (when day formula returns empty for non-existent days)
	//    E.g., Sept has 30 days, day 31 returns "" → mute
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

	// ═══════════════════════════════════════════════════════════════════
	// FIX H_25 formula (use single-letter "D" not "dom")
	// ═══════════════════════════════════════════════════════════════════
	for row := 10; row <= 30; row++ {
		// AN = H_25 (Sunday hours) — use "D" for single-letter day
		reqs = append(reqs, cellFormula(fmt.Sprintf("AN%d", row),
			fmt.Sprintf(`=SUMPRODUCT((E$9:AI$9="D")*(N(E%d:AI%d)))`, row, row)))
	}

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
	fmt.Println("FIXES:")
	fmt.Println("  - Holiday conditional format: VLOOKUP against 01_FERIADOS_PE (without quotes)")
	fmt.Println("  - Weekend conditional formats: S (Sábado) and D (Domingo) 1-char")
	fmt.Println("  - Ghost day (E$8=\"\" → muted strikethrough)")
	fmt.Println("  - H_25 formula: was \"dom\" now \"D\" (1-char matching)")
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
	// Reverse so we delete from highest index to lowest
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	return ids
}

func cellFormula(cell, formula string) map[string]any {
	ri := rowIdx(cell)
	ci := colIdx(cell)
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
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

func rowIdx(cell string) int {
	row := 0
	for _, c := range cell {
		if c >= '0' && c <= '9' {
			row = row*10 + int(c-'0')
		}
	}
	return row - 1
}

func colIdx(cell string) int {
	col := 0
	for _, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		}
	}
	return col - 1
}

// ── Colors ─────────────────────────────────────────────────────────

func bgHoliday() map[string]any {
	return map[string]any{"red": 1.0, "green": 0.843, "blue": 0.0}
}
func fgHoliday() map[string]any {
	return map[string]any{"red": 0.698, "green": 0.133, "blue": 0.133}
}
func bgWeekend() map[string]any {
	return map[string]any{"red": 0.95, "green": 0.96, "blue": 0.98}
}
func bgSunday() map[string]any {
	return map[string]any{"red": 0.992, "green": 0.922, "blue": 0.816}
}
func bgMuted() map[string]any {
	return map[string]any{"red": 0.917, "green": 0.929, "blue": 0.945}
}
func fgMuted() map[string]any {
	return map[string]any{"red": 0.451, "green": 0.518, "blue": 0.6}
}
func fgPrimary() map[string]any {
	return map[string]any{"red": 0.059, "green": 0.141, "blue": 0.251}
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
	tok, _ := res["access_token"].(string)
	return tok
}
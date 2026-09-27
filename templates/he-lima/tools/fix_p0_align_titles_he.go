// fix_p0_align_titles_he.go — CEO directive 2026-09-26
// P0 fixes: title alignment, VALORIZACIÓN alignment, day letters 1-letter.
//
// Bug 1: Title B1 (data starts A1). Move to A1:AI1 (includes N° col).
// Bug 2: VALORIZACIÓN at AK1 (numbers start AJ). Move to AJ1:AT1.
// Bug 3: Day letters "miércoles" (11pt) overflows 28px cols. Switch to 1-letter.
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
	// BUG 1 — Title alignment: unmerge B1:AI1, merge A1:AI1
	//         Title now starts at col A (same as N° col below)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 1, // B
				"endColumnIndex":   35, // AJ exclusive = AI inclusive
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0, // A — start at col A
				"endColumnIndex":   35,
			},
			"mergeType": "MERGE_ALL",
		},
	})
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   1, // A1
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
	// Style A1:AI1 (whole title row)
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
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// BUG 2 — VALORIZACIÓN alignment: unmerge AK1:AT1, merge AJ1:AT1
	//         Header now starts at AJ (same as H_norm col below)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 36, // AK
				"endColumnIndex":   46, // AT
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   46, // AT
			},
			"mergeType": "MERGE_ALL",
		},
	})
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   36,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": "VALORIZACIÓN"},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	})
	// Style AJ1:AT1 (whole VALORIZACIÓN header)
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
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// BUG 3 — Day letters: rewrite E9:AI9 to use 1-letter Spanish abbrev
	//         Use LEFT(TEXT(...,"DDD"),1) but Spanish needs custom mapping.
	//         Use CHOOSE: dom→D, lun→L, mar→M, mié→X, jue→J, vie→V, sáb→S
	// ═══════════════════════════════════════════════════════════════════

	// First: write new formulas. The original formula was TEXT(DATE,"DDD").
	// New formula: CHOOSE(WEEKDAY(...), "D", "L", "M", "X", "J", "V", "S")
	// Spanish week: Sun=D, Mon=L, Tue=M, Wed=X, Thu=J, Fri=V, Sat=S
	// WEEKDAY default: 1=Sun, 7=Sat
	type cell struct {
		formula string
	}

	// Build 31 formulas, one per column E..AI. Each formula uses the day
	// number in row 8 (col-locked row) to compute weekday letter.
	dayCols := []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}
	for i, col := range dayCols {
		n := i + 1
		formula := fmt.Sprintf(
			`=IF($%s$8="","",CHOOSE(WEEKDAY($%s$8,2),"L","M","X","J","V","S","D"))`,
			col, col,
		)
		_ = formula
		_ = n
	}

	// Simpler: just rewrite E9 formula, then sheet auto-replicates via copy
	// Actually batchUpdate updateCells can write one row of 31 cells with
	// different formulas each.
	row := []any{}
	for _, col := range dayCols {
		formula := fmt.Sprintf(
			`=IF($%s$8="","",CHOOSE(WEEKDAY($%s$8,2),"L","M","X","J","V","S","D"))`,
			col, col,
		)
		row = append(row, map[string]any{
			"userEnteredValue": map[string]any{"formulaValue": formula},
		})
	}

	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    8, // row 9 (0-based)
				"endRowIndex":      9,
				"startColumnIndex": 4, // E
				"endColumnIndex":   35,
			},
			"rows": []any{
				map[string]any{"values": row},
			},
			"fields": "userEnteredValue",
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
	fmt.Println("P0 fixes:")
	fmt.Println("  1. Title alignment: B1:AI1 → A1:AI1 (includes N° col)")
	fmt.Println("  2. VALORIZACIÓN: AK1:AT1 → AJ1:AT1 (aligns with H_norm)")
	fmt.Println("  3. Day letters: 3-char (mié) → 1-letter (L M X J V S D)")
}

func bgBrandPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786,
		"green": 0.227450982,
		"blue":  0.360784314,
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
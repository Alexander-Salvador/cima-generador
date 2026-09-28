// fix_p1_empty_rows_spacers_he.go — P1 fixes: empty emp rows + spacer cols.
//
// Bug 4: Rows 26-30 (LEON, GOMEZ, OSORIO, IMAN, YANA) have name but no
//        day data. Cells look "broken" because they have white bg with
//        nothing in them. Apply muted bg to whole row (E-AI).
//
// Bug 5: Spacer cols AK, AM, AO, AS (8px) inside numbers section show as
//        dark vertical lines. Apply alertSoft bg to blend with numbers
//        section.
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
	// BUG 4 — Empty employee rows: apply muted style to rows 26-30
	//         Day area E-AI: muted bg + muted fg
	//         Name area A-D: also muted bg (consistent with day area)
	//         Numbers area AJ-AT: muted alertSoft
	// ═══════════════════════════════════════════════════════════════════

	// Day area (E-AI, rows 26-30): muted bg
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25, // row 26 (0-based)
				"endRowIndex":      30, // row 30 (0-based, exclusive)
				"startColumnIndex": 4,  // E
				"endColumnIndex":   35, // AJ exclusive = AI inclusive
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgMuted(),
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// Name area (A-D, rows 26-30): same muted bg as day area
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25,
				"endRowIndex":      30,
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

	// Numbers area (AJ-AT, rows 26-30): lighter muted
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25,
				"endRowIndex":      30,
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
	// BUG 5 — Spacer cols (AK, AM, AO, AS): blend with numbers section bg
	//         AK (col 37): between AJ (H_norm) and AL (H_ext)
	//         AM (col 39): hidden, between AL (H_ext) and AN (H_25%)
	//         AO (col 41): hidden, between AN (H_25%) and AP (H_35%)
	//         AS (col 45): between AR ($_N) and AT ($_E)
	//
	//         Apply alertSoft bg (rows 8-31) so they blend visually
	// ═══════════════════════════════════════════════════════════════════
	for _, colIdx := range []int{37, 39, 41, 45} { // AK, AM, AO, AS
		reqs = append(reqs, map[string]any{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    7,  // row 8
					"endRowIndex":      31, // row 31 (TOTALES)
					"startColumnIndex": colIdx,
					"endColumnIndex":   colIdx + 1,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"backgroundColor": bgAlertSoft(),
					},
				},
				"fields": "userEnteredFormat.backgroundColor",
			},
		})
	}

	// Also AJ, AL, AN, AP, AR, AT — make sure numbers section bg is
	// alertSoft throughout (rows 10-30, not just headers)
	// Already done in fix_day31_and_polish_he.go. Skip.

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
	fmt.Println("P1 fixes:")
	fmt.Println("  4. Empty emp rows (26-30): muted bg + muted italic fg")
	fmt.Println("  5. Spacer cols (AK, AM, AO, AS): alertSoft bg (blend)")
}

// ── Tokens ─────────────────────────────────────────────────────────

func bgMuted() map[string]any {
	return map[string]any{
		"red":   0.9176471,
		"green": 0.9294118,
		"blue":  0.9450980,
	}
}

func bgAlertSoft() map[string]any {
	return map[string]any{
		"red":   0.9921569,
		"green": 0.9215686,
		"blue":  0.8156863,
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
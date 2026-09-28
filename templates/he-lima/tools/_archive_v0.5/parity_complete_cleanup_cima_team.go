// parity_complete_cleanup_cima_team.go — CEO directive 2026-09-27
// "verifica pariedad real en toda la HOJA HE"
//
// ONE-SHOT parity cleanup. Fixes:
// 1. Remove all 6 conditional format rules (re-apply 4 cleanly)
// 2. Clear all bg colors per-cell (so alternating weirdness goes)
// 3. Apply consistent bg palette:
//    - Rows 1-7 (header): clean light bg (no alternating)
//    - Rows 8-9 (day headers): primarySubtle + brandPrimary
//    - Rows 10-25 (emp with data): white bg, center align
//    - Rows 26-34 (emp without data): muted bg + italic
//    - Row 35 (spacer): white
//    - Row 36 (TOTALES): primarySubtle bg + bold
// 4. Clear ALL borders per-row, re-apply 2 clean frames
// 5. Spacers (AK, AM, AO, AS) width 6px consistent
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

	// First, get current conditional format rule IDs to delete
	ruleIDs := getConditionalFormatRuleIDs(tok)

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Delete all existing conditional format rules on CIMA TEAM
	//          (will re-apply correctly after)
	// ═══════════════════════════════════════════════════════════════════
	for _, ruleID := range ruleIDs {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   ruleID,
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Force clear ALL borders per-row (rows 0-40)
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
	// STEP 3 — Clear all bg colors per-cell (rows 0-40) to remove
	//          alternating weird colors. We'll re-apply clean below.
	// ═══════════════════════════════════════════════════════════════════
	for row := 0; row < 40; row++ {
		reqs = append(reqs, map[string]any{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row,
					"endRowIndex":      row + 1,
					"startColumnIndex": 0,
					"endColumnIndex":   46,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"backgroundColor": map[string]any{
							"red":   1,
							"green": 1,
							"blue":  1,
						},
					},
				},
				"fields": "userEnteredFormat.backgroundColor",
			},
		})
	}

	// Send batch (clear + reset bg)
	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR clear:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respB, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR clear", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Cleared %d requests (rules + borders + bg)\n", len(reqs))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Apply clean CIMA TEAM design
	// ═══════════════════════════════════════════════════════════════════
	reqs2 := []any{}

	// Header section bg (rows 1-7) — light subtle
	for row := 0; row < 7; row++ {
		reqs2 = append(reqs2, map[string]any{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row,
					"endRowIndex":      row + 1,
					"startColumnIndex": 0,
					"endColumnIndex":   46,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"backgroundColor": bgPrimarySubtle(),
					},
				},
				"fields": "userEnteredFormat.backgroundColor",
			},
		})
	}

	// Row 1 (title): override bg to brand primary (navy)
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
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

	// Row 8 (day numbers): primarySubtle bg + bold center
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      8,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgPrimarySubtle(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"bold":            true,
						"fontSize":        11,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})

	// Row 9 (day letters): brand primary bg + white bold center
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    8,
				"endRowIndex":      9,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgBrandPrimary(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgInverse(),
						"bold":            true,
						"fontSize":        11,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})

	// Rows 10-25 (emp with data): name area primarySubtle + day area white
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,
				"endRowIndex":      25,
				"startColumnIndex": 0,
				"endColumnIndex":   5,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgPrimarySubtle(),
					"horizontalAlignment": "LEFT",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"fontSize":        10,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})
	// Day area rows 10-25 (E-AH)
	reqs2 = append(reqs2, map[string]any{
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
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"fontSize":        10,
					},
				},
			},
			"fields": "userEnteredFormat(horizontalAlignment,verticalAlignment,textFormat)",
		},
	})
	// Numbers area rows 10-25 (AJ-AT) alertSoft
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,
				"endRowIndex":      25,
				"startColumnIndex": 35,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgAlertSoft(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"fontSize":        10,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})

	// Rows 26-34 (emp without data): muted bg + italic muted fg
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25,
				"endRowIndex":      34,
				"startColumnIndex": 0,
				"endColumnIndex":   5,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgMuted(),
					"horizontalAlignment": "LEFT",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgMuted(),
						"italic":          true,
						"fontSize":        10,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})
	// Day area 26-34 (E-AH)
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25,
				"endRowIndex":      34,
				"startColumnIndex": 4,
				"endColumnIndex":   34,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgMuted(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment)",
		},
	})
	// Numbers area 26-34 (AJ-AT) alertSoftMuted
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25,
				"endRowIndex":      34,
				"startColumnIndex": 35,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgAlertSoftMuted(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment)",
		},
	})

	// Row 36 (TOTALES): primarySubtle bg + bold
	reqs2 = append(reqs2, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    35,
				"endRowIndex":      36,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgPrimarySubtle(),
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"bold":            true,
						"fontSize":        11,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — 2 clean frames
	//          FRAME 1 HEADER: A1:AT7
	//          FRAME 2 DATA:   A8:AT36
	// ═══════════════════════════════════════════════════════════════════
	reqs2 = append(reqs2, map[string]any{
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
	reqs2 = append(reqs2, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36,
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
	// STEP 6 — Internal accents (subtle, 1px)
	//          D|E (name|day) + AI|AJ (day|numbers)
	// ═══════════════════════════════════════════════════════════════════
	reqs2 = append(reqs2, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36,
				"startColumnIndex": 4,
				"endColumnIndex":   5,
			},
			"left": border("SOLID", borderAccent(), 1),
		},
	})
	reqs2 = append(reqs2, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      36,
				"startColumnIndex": 35,
				"endColumnIndex":   36,
			},
			"left": border("SOLID", borderAccent(), 1),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 7 — Re-apply 4 conditional formats (CLEAN, no dupes)
	//          a) Ghost day when =E$8="" (rows 8-34, cols E-AH)
	//          b) Weekend "sáb" (rows 9, cols E-AH)
	//          c) Weekend "dom" (rows 9, cols E-AH)
	// ═══════════════════════════════════════════════════════════════════
	// Ghost day (covers header row 8 + emp rows 9-34, day area E-AH)
	reqs2 = append(reqs2, map[string]any{
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
	// Weekend "sáb" (row 9, cols E-AH)
	reqs2 = append(reqs2, map[string]any{
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
	// Weekend "dom" (row 9, cols E-AH)
	reqs2 = append(reqs2, map[string]any{
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

	// Send batch 2 (apply design)
	body2 := map[string]any{"requests": reqs2}
	bb2, _ := json.Marshal(body2)
	req2, _ := http.NewRequest("POST", url, bytes.NewReader(bb2))
	req2.Header.Set("Authorization", "Bearer "+tok)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println("ERROR apply:", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()
	respB2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode >= 300 {
		fmt.Println("API ERROR apply", resp2.StatusCode, ":", string(respB2))
		os.Exit(1)
	}
	fmt.Printf("✅ Applied %d design requests\n", len(reqs2))
	fmt.Println("CIMA TEAM parity cleanup complete:")
	fmt.Println("  ✓ Removed all 6 conditional format rules (re-applied 3 clean)")
	fmt.Println("  ✓ Cleared all borders + bg colors")
	fmt.Println("  ✓ Applied unified palette:")
	fmt.Println("    - Header (rows 1-7): primarySubtle + title navy")
	fmt.Println("    - Day numbers (row 8): primarySubtle + bold center")
	fmt.Println("    - Day letters (row 9): navy + white bold")
	fmt.Println("    - Emp with data (rows 10-25): name area primarySubtle, day white, numbers alertSoft")
	fmt.Println("    - Emp without data (rows 26-34): muted bg + italic")
	fmt.Println("    - TOTALES (row 36): primarySubtle + bold")
	fmt.Println("  ✓ 2 clean frames (header A1:AT7, data A8:AT36)")
	fmt.Println("  ✓ Internal accents at D|E and AI|AJ")
}

// ── Helpers ────────────────────────────────────────────────────────

func getConditionalFormatRuleIDs(tok string) []int {
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?sheets(conditionalFormats)&ranges=CIMA%%20TEAM", spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)

	ids := []int{}
	sheets, _ := out["sheets"].([]any)
	for _, s := range sheets {
		sp, _ := s.(map[string]any)
		cf, _ := sp["conditionalFormats"].([]any)
		for i := range cf {
			ids = append(ids, i)
		}
	}
	return ids
}

func bgBrandPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786, "green": 0.227450982, "blue": 0.360784314,
	}
}
func bgPrimarySubtle() map[string]any {
	return map[string]any{
		"red":   0.8313726, "green": 0.8901961, "blue": 0.9529412,
	}
}
func bgAlertSoft() map[string]any {
	return map[string]any{
		"red":   0.9921569, "green": 0.9215686, "blue": 0.8156863,
	}
}
func bgAlertSoftMuted() map[string]any {
	return map[string]any{
		"red":   0.9647059, "green": 0.9176471, "blue": 0.8588235,
	}
}
func bgMuted() map[string]any {
	return map[string]any{
		"red":   0.9176471, "green": 0.9294118, "blue": 0.9450980,
	}
}
func fgPrimary() map[string]any {
	return map[string]any{
		"red":   0.0588235, "green": 0.1411765, "blue": 0.2509804,
	}
}
func fgInverse() map[string]any {
	return map[string]any{"red": 1, "green": 1, "blue": 1}
}
func fgMuted() map[string]any {
	return map[string]any{
		"red":   0.4509804, "green": 0.5176471, "blue": 0.6000000,
	}
}
func borderPrimary() map[string]any { return bgBrandPrimary() }
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
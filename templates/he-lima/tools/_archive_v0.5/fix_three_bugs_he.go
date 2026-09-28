// fix_three_bugs_he.go — CEO directive 2026-09-26
// Bug 1: Containers mal generados (title muy corto + duplicate TOTALES row)
// Bug 2: Separacion de nombres no marcada
// Bug 3: Dias 1-31 deben tener espacio fijo + ghost dias que no existen en mes
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
	// BUG 1.1 — Fix title: unmerge B1:T1, merge B1:AI1 (full data section)
	//           B=1, T=19, AI=34 (0-based startColumnIndex/endColumnIndex
	//           are exclusive-end)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 1, // B
				"endColumnIndex":   20, // T+1
			},
		},
	})
	reqs = append(reqs, map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 1,  // B
				"endColumnIndex":   35, // AI+1 (col 34 = AI, exclusive end 35)
			},
			"mergeType": "MERGE_ALL",
		},
	})
	// Re-style B1 with new width
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 1,
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
	// BUG 1.2 — Add frame border around HEADER section (rows 5-9, B-AI)
	//           Selector rows + day headers
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    4,  // row 5
				"endRowIndex":      9,  // row 9 (exclusive)
				"startColumnIndex": 1,  // B
				"endColumnIndex":   35, // AJ exclusive = AI inclusive
			},
			"top":    border("SOLID_MEDIUM", borderPrimary(), 2),
			"bottom": border("SOLID_MEDIUM", borderPrimary(), 2),
			"left":   border("SOLID_MEDIUM", borderPrimary(), 2),
			"right":  border("SOLID_MEDIUM", borderPrimary(), 2),
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// BUG 2 — Name separation: vertical accent border between name area
	//         (cols B-D merged) and day area (col E onward). Plus
	//         primarySubtle bg on name area (rows 10-30, B-D).
	// ═══════════════════════════════════════════════════════════════════

	// Vertical accent border: right of D / left of E (rows 8-32)
	reqs = append(reqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,  // row 8
				"endRowIndex":      31, // row 31 (TOTALES = 32, exclusive 31)
				"startColumnIndex": 4,  // E
				"endColumnIndex":   5,
			},
			"left": border("SOLID_THICK", borderAccent(), 3),
		},
	})

	// Subtle bg for name area (rows 10-30, cols B-D merged area)
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,  // row 10
				"endRowIndex":      30, // row 30 (exclusive end)
				"startColumnIndex": 1,  // B
				"endColumnIndex":   5,  // E exclusive = D inclusive
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgPrimarySubtle(),
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// N° column A: bg matching name area
	reqs = append(reqs, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,
				"endRowIndex":      30,
				"startColumnIndex": 0, // A
				"endColumnIndex":   1,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor":    bgPrimarySubtle(),
					"horizontalAlignment": "CENTER",
					"textFormat": map[string]any{
						"foregroundColor": fgPrimary(),
						"bold":            true,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,horizontalAlignment,textFormat)",
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// BUG 3 — Intelligent day coverage: ghost cols where day formula = ""
	//         Trigger: =E$8="" (E col, row 8 — formula evaluates empty
	//         when day > days_in_month)
	//         Applied to rows 8-31 (headers + emp data) cols E-AI
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetHE,
						"startRowIndex":    7,  // row 8
						"endRowIndex":      31, // row 31 (TOTALES)
						"startColumnIndex": 4,  // E
						"endColumnIndex":   35, // AJ exclusive = AI inclusive
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
	// BUG 1.3 — Delete duplicate TOTALES row (row 33 → index 32)
	//           Wait: row 33 was already deleted via CLI before this script.
	//           Skip if already gone.
	// ═══════════════════════════════════════════════════════════════════
	// (Skipped — CLI already removed row 33)

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
	fmt.Println("Bugs fixed:")
	fmt.Println("  1. Containers: title B1:AI1 + frame rows 5-9 (data section header)")
	fmt.Println("  2. Name separation: thick accent border D|E + primarySubtle bg")
	fmt.Println("  3. Intelligent days: ghost format when =E$8=\"\" (mutes non-existent)")
}

// ── Token helpers (HE LIMA v1.0.1) ─────────────────────────────────

func bgBrandPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786,
		"green": 0.227450982,
		"blue":  0.360784314,
	}
}

func bgPrimarySubtle() map[string]any {
	return map[string]any{
		"red":   0.8313726,
		"green": 0.8901961,
		"blue":  0.9529412,
	}
}

func bgMuted() map[string]any {
	return map[string]any{
		"red":   0.9176471,
		"green": 0.9294118,
		"blue":  0.9450980,
	}
}

func fgPrimary() map[string]any {
	return map[string]any{
		"red":   0.0588235,
		"green": 0.1411765,
		"blue":  0.2509804,
	}
}

func fgInverse() map[string]any {
	return map[string]any{"red": 1, "green": 1, "blue": 1}
}

func fgMuted() map[string]any {
	return map[string]any{
		"red":   0.4509804,
		"green": 0.5176471,
		"blue":  0.6000000,
	}
}

func borderPrimary() map[string]any { return bgBrandPrimary() }

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

// ── OAuth decrypt ──────────────────────────────────────────────────

func decrypt() map[string]any {
	keyB, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	encB, err := os.ReadFile(encPath)
	if err != nil {
		panic(err)
	}
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	nonce := encB[:12]
	ct := encB[12:]
	pt, _ := gcm.Open(nil, nonce, ct, nil)
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
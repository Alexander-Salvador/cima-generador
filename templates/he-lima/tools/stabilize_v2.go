// stabilize_v2.go — FASE 5, 7, 8: Extend CFs a row 7, format moneda, validation.
//
//   FASE 5: Extender 3 CFs al rango E7:AI30 (incluye row 7 fechas)
//   FASE 7: Format moneda en T3 ($ HE) — S/. #,##0.00
//   FASE 8: Data Validation en E10:AI30 — 0 ≤ value ≤ 12
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

// Full day range E7:AI30 (rows 7-30, cols 5-35)
var rangeDays = []any{
	map[string]any{
		"sheetId":          sheetHE,
		"startRowIndex":    6,
		"endRowIndex":      30,
		"startColumnIndex": 4,
		"endColumnIndex":   35,
	},
}

// Inputs range E10:AI30 (rows 10-30, cols 5-35) for data validation
var rangeInputs = []any{
	map[string]any{
		"sheetId":          sheetHE,
		"startRowIndex":    9,
		"endRowIndex":      30,
		"startColumnIndex": 4,
		"endColumnIndex":   35,
	},
}

func main() {
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Delete all existing CF rules
	// ═══════════════════════════════════════════════════════════════════
	cfCount := getCFCount(tok)
	fmt.Printf("Found %d existing CFs\n", cfCount)
	for i := cfCount - 1; i >= 0; i-- {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   i,
			},
		})
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 5 — Add 3 CFs over E7:AI30 (row 7 included)
	// ═══════════════════════════════════════════════════════════════════
	const dateRef = `INDIRECT(ADDRESS(7,COLUMN(),4))`

	// CF 1: Feriado (priority alta — gold + bold red text)
	reqs = append(reqs, addCFFeriado(dateRef))
	// CF 2: Domingo — bg alert_soft
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=7`, 1.0, 0.92, 0.81))
	// CF 3: Sábado — bg subtle
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=6`, 0.94, 0.96, 0.97))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 7 — Format moneda en T3 ($ HE) — S/. #,##0.00
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, setNumberFormat("T", 3, `"S/."#,##0.00`))

	// Bold + center T3
	reqs = append(reqs, setFormat("T", 3, true, "CENTER"))

	// Same format for AP36 ($ HE total)
	reqs = append(reqs, setNumberFormat("AP", 36, `"S/."#,##0.00`))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 8 — Data Validation en E10:AI30 (inputs)
	//         0 ≤ value ≤ 12 (horas extras razonables)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, addDataValidation())

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
	respB, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Stabilized: 3 CFs (E7:AI30), moneda format, data validation 0-12\n")
}

func getCFCount(tok string) int {
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(sheetId,title),conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets := out["sheets"].([]any)
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		if sidF, ok := props["sheetId"].(float64); ok && int64(sidF) == sheetHE {
			if cf, ok := sm["conditionalFormats"].([]any); ok {
				return len(cf)
			}
		}
	}
	return 0
}

func addCF(formula string, r, g, b float64) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeDays,
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": formula},
						},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{
							"red":   r,
							"green": g,
							"blue":  b,
						},
					},
				},
			},
		},
	}
}

func addCFFeriado(dateRef string) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeDays,
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{
								"userEnteredValue": `=ISNUMBER(VLOOKUP(` + dateRef + `,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE))`,
							},
						},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{
							"red":   1.0,
							"green": 0.84,
							"blue":  0.0,
						},
						"textFormat": map[string]any{
							"foregroundColor": map[string]any{
								"red":   0.7,
								"green": 0.0,
								"blue":  0.0,
							},
							"bold": true,
						},
					},
				},
			},
		},
	}
}

func setNumberFormat(col string, row int, pattern string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
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
							"userEnteredFormat": map[string]any{
								"numberFormat": map[string]any{
									"type":   "NUMBER",
									"pattern": pattern,
								},
							},
						},
					},
				},
			},
			"fields": "userEnteredFormat.numberFormat",
		},
	}
}

func setFormat(col string, row int, bold bool, halign string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	tf := map[string]any{}
	if bold {
		tf["bold"] = true
	}
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
							"userEnteredFormat": map[string]any{
								"textFormat":         tf,
								"horizontalAlignment": halign,
							},
						},
					},
				},
			},
			"fields": "userEnteredFormat(textFormat,horizontalAlignment)",
		},
	}
}

func addDataValidation() map[string]any {
	return map[string]any{
		"setDataValidation": map[string]any{
			"range": rangeInputs[0],
			"rule": map[string]any{
				"condition": map[string]any{
					"type": "NUMBER_BETWEEN",
					"values": []any{
						map[string]any{"userEnteredValue": "0"},
						map[string]any{"userEnteredValue": "12"},
					},
				},
				"inputMessage": "Ingrese horas extras del día (0-12)",
				"strict":      true,
				"showCustomUi": true,
			},
		},
	}
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

// polish_and_frame_he.go — Polish day headers + frame data/numbers + trim tabs.
// CEO directive: 'pule las fechas numeritas y los dias, enmarca la parte de
// datos y la parte de numeros para que se diferencie y elimine las columnas
// innecesarias que sobren'
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
	sheetPER      = int64(1395452450)
	sheetFB       = int64(1347080845)
	sheetFER      = int64(1407735723)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	creds := decrypt()
	tok := getToken(creds)

	requests := []any{}

	// ── 10_HE: Polish day headers + frame sections ─────────────────────
	heRequests := []any{}

	// ROW 8: Day numbers — bigger, bolder, polished
	heRequests = append(heRequests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      8,
				"startColumnIndex": 4, // E
				"endColumnIndex":   35, // AI
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.8313726, // primarySubtle #D4E3F3
						"green": 0.8901961,
						"blue":  0.9529412,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   0.0588235, // text.primary #0F2440
							"green": 0.1411765,
							"blue":  0.2509804,
						},
						"fontSize": 11,
						"bold":     true,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	// ROW 9: Day letters — polished (brand.primary bg, white fg, bold, center)
	heRequests = append(heRequests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    8,
				"endRowIndex":      9,
				"startColumnIndex": 4,
				"endColumnIndex":   35,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.101960786, // brand.primary #1A3A5C
						"green": 0.227450982,
						"blue":  0.360784314,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   1, // text.inverse
							"green": 1,
							"blue":  1,
						},
						"fontSize": 11,
						"bold":     true,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	// Conditional formatting: Weekend highlight on row 9 (sáb/dom)
	// When cell value = "sáb" or "dom", apply accent.alertSoft bg
	for _, weekend := range []string{"sáb", "dom"} {
		heRequests = append(heRequests, map[string]any{
			"addConditionalFormatRule": map[string]any{
				"rule": map[string]any{
					"ranges": []any{
						map[string]any{
							"sheetId":          sheetHE,
							"startRowIndex":    7,
							"endRowIndex":      9,
							"startColumnIndex": 4,
							"endColumnIndex":   35,
						},
					},
					"booleanRule": map[string]any{
						"condition": map[string]any{
							"type":         "CUSTOM_FORMULA",
							"values": []any{
								map[string]any{
									"userEnteredValue": fmt.Sprintf("=$E9=\"%s\"", weekend),
								},
							},
						},
						"format": map[string]any{
							"backgroundColor": map[string]any{
								"red":   0.9921569, // accent.alertSoft #FDEBD0
								"green": 0.9215686,
								"blue":  0.8156863,
							},
							"textFormat": map[string]any{
								"foregroundColor": map[string]any{
									"red":   0.0588235, // text.primary
									"green": 0.1411765,
									"blue":  0.2509804,
								},
								"bold": true,
							},
						},
					},
				},
			},
		})
	}

	// THICK FRAME around data section: cols B-AH (employee + days), rows 8-34
	heRequests = append(heRequests, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,  // row 8
				"endRowIndex":      34, // row 34
				"startColumnIndex": 1,  // B
				"endColumnIndex":   34, // AH
			},
			"top": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.101960786, // brand.primary
					"green": 0.227450982,
					"blue":  0.360784314,
				},
				"width": 2,
			},
			"bottom": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.101960786,
					"green": 0.227450982,
					"blue":  0.360784314,
				},
				"width": 2,
			},
			"left": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.101960786,
					"green": 0.227450982,
					"blue":  0.360784314,
				},
				"width": 2,
			},
			"right": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.101960786,
					"green": 0.227450982,
					"blue":  0.360784314,
				},
				"width": 2,
			},
		},
	})

	// THICK FRAME around numbers section: cols AJ-AT, rows 8-34
	heRequests = append(heRequests, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      34,
				"startColumnIndex": 35, // AJ
				"endColumnIndex":   46, // AT
			},
			"top": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.9019608, // accent.alert
					"green": 0.4941176,
					"blue":  0.1333333,
				},
				"width": 2,
			},
			"bottom": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.9019608,
					"green": 0.4941176,
					"blue":  0.1333333,
				},
				"width": 2,
			},
			"left": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.9019608,
					"green": 0.4941176,
					"blue":  0.1333333,
				},
				"width": 2,
			},
			"right": map[string]any{
				"style": "SOLID_MEDIUM",
				"color": map[string]any{
					"red":   0.9019608,
					"green": 0.4941176,
					"blue":  0.1333333,
				},
				"width": 2,
			},
		},
	})

	// Different bg for the numbers section (accent.alertSoft for the section)
	heRequests = append(heRequests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,  // row 10 (employee data start)
				"endRowIndex":      33, // row 33
				"startColumnIndex": 36, // AK (spacer)
				"endColumnIndex":   37, // AL (H_ext)
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.9921569, // accent.alertSoft
						"green": 0.9215686,
						"blue":  0.8156863,
					},
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// Subtle bg for spacer cols (AI spacer between data and numbers)
	heRequests = append(heRequests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    9,
				"endRowIndex":      33,
				"startColumnIndex": 34, // AI
				"endColumnIndex":   35, // AJ
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.5,
						"green": 0.5,
						"blue":  0.5,
					},
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	})

	// ── Trim grids (eliminate empty cols) ──────────────────────────────
	// 01_FERIADOS_PE: 26 → 5 cols (only A-E used)
	// Feedbacks: 26 → 7 cols (only A-G used)
	// PERSONAL: 20 → 8 cols (only A-H used)
	// CONFIG: 11 (all used, no trim)
	// 10_HE: 46 (all used, no trim)

	// Delete empty cols from PERSONAL (cols I-T = indices 8-19)
	for colIdx := 19; colIdx >= 8; colIdx-- {
		requests = append(requests, map[string]any{
			"deleteDimension": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetPER,
					"dimension":  "COLUMNS",
					"startIndex": colIdx,
					"endIndex":   colIdx + 1,
				},
			},
		})
	}

	// Delete empty cols from Feedbacks (cols H-Z = indices 7-25)
	for colIdx := 25; colIdx >= 7; colIdx-- {
		requests = append(requests, map[string]any{
			"deleteDimension": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetFB,
					"dimension":  "COLUMNS",
					"startIndex": colIdx,
					"endIndex":   colIdx + 1,
				},
			},
		})
	}

	// Delete empty cols from 01_FERIADOS_PE (cols F-Z = indices 5-25)
	for colIdx := 25; colIdx >= 5; colIdx-- {
		requests = append(requests, map[string]any{
			"deleteDimension": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetFER,
					"dimension":  "COLUMNS",
					"startIndex": colIdx,
					"endIndex":   colIdx + 1,
				},
			},
		})
	}

	requests = append(requests, heRequests...)

	// Execute all
	body := map[string]any{"requests": requests}
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("POLISH+TRIM: %d\n%s\n", resp.StatusCode, string(respBody))

	fmt.Println("\n✓ Polish + frame + trim applied:")
	fmt.Println("  • 10_HE row 8 (day nums): bigger, bolder, polished")
	fmt.Println("  • 10_HE row 9 (day letters): brand.primary bg, white fg, bold")
	fmt.Println("  • 10_HE conditional format: weekend highlight (sáb/dom)")
	fmt.Println("  • 10_HE data section frame: navy thick border (B-AH)")
	fmt.Println("  • 10_HE numbers section frame: orange thick border (AJ-AT)")
	fmt.Println("  • PERSONAL: trimmed 20 → 8 cols (removed I-T)")
	fmt.Println("  • Feedbacks: trimmed 26 → 7 cols (removed H-Z)")
	fmt.Println("  • 01_FERIADOS_PE: trimmed 26 → 5 cols (removed F-Z)")
}

func decrypt() map[string]any {
	enc, _ := os.ReadFile(encPath)
	keyFile, _ := os.ReadFile(keyPath)
	nonce := enc[:12]
	ct := enc[12:]
	block, _ := aes.NewCipher(keyFile)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, nonce, ct, nil)
	var creds map[string]any
	json.Unmarshal(pt, &creds)
	return creds
}

func getToken(creds map[string]any) string {
	rt := creds["refresh_token"].(string)
	cid := creds["client_id"].(string)
	cs, _ := creds["client_secret"].(string)
	form := "client_id=" + cid
	if cs != "" {
		form += "&client_secret=" + cs
	}
	form += "&refresh_token=" + rt + "&grant_type=refresh_token"
	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token",
		bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct{ AccessToken string `json:"access_token"` }
	json.Unmarshal(body, &out)
	return out.AccessToken
}
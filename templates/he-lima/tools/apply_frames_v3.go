// apply_frames_v3.go — Apply frames + CFs
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
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	creds := map[string]any{}
	json.Unmarshal(pt, &creds)

	body, _ := json.Marshal(map[string]any{
		"client_id":     creds["client_id"],
		"client_secret": creds["client_secret"],
		"refresh_token": creds["refresh_token"],
		"grant_type":    "refresh_token",
	})
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	b, _ := io.ReadAll(r.Body)
	r.Body.Close()
	res := map[string]any{}
	json.Unmarshal(b, &res)
	tok := res["access_token"].(string)

	reqs := []any{}

	// Clear all borders A1:AP36
	reqs = append(reqs, clearAllBorders(0, 0, 35, 41))

	// HEADER frame: rows 1-7, cols 1-42 (0-based: rows 0-6, cols 0-41)
	reqs = append(reqs, frameBorder(0, 0, 6, 41, 3))

	// DATA frame: rows 8-36, cols 1-42 (0-based: rows 7-35, cols 0-41)
	reqs = append(reqs, frameBorder(7, 0, 35, 41, 3))

	// Internal accents
	reqs = append(reqs, internalVertical(7, 0, 35, 1, 1))  // A|B vertical
	reqs = append(reqs, internalVertical(7, 34, 35, 35, 1)) // AI|AJ vertical

	// Re-apply 3 CFs
	// Delete old CFs first
	cfCount := getCFCount(tok)
	fmt.Printf("Found %d CFs to delete\n", cfCount)
	for i := cfCount - 1; i >= 0; i-- {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   i,
			},
		})
	}

	// Add 3 CFs
	const dateRef = `INDIRECT(ADDRESS(7,COLUMN(),4))`
	reqs = append(reqs, addCFFeriado(dateRef))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=7`, 1.0, 0.92, 0.81))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=6`, 0.94, 0.96, 0.97))

	body2, _ := json.Marshal(map[string]any{"requests": reqs})
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body2))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respB, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode >= 300 {
		fmt.Println("ERROR:", string(respB)[:500])
	} else {
		fmt.Printf("✅ Frames + 3 CFs applied\n")
	}
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
	sheets, _ := out["sheets"].([]any)
	for _, sh := range sheets {
		sm, _ := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		if sidF, ok := props["sheetId"].(float64); ok && int64(sidF) == sheetHE {
			cfAny := sm["conditionalFormats"]
			if cfAny == nil {
				return 0
			}
			if cf, ok := cfAny.([]any); ok {
				return len(cf)
			}
		}
	}
	return 0
}

func clearAllBorders(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow, "endRowIndex": endRow,
				"startColumnIndex": startCol, "endColumnIndex": endCol,
			},
			"top":             map[string]any{"style": "NONE"},
			"bottom":          map[string]any{"style": "NONE"},
			"left":            map[string]any{"style": "NONE"},
			"right":           map[string]any{"style": "NONE"},
			"innerHorizontal": map[string]any{"style": "NONE"},
			"innerVertical":   map[string]any{"style": "NONE"},
		},
	}
}

func frameBorder(startRow, startCol, endRow, endCol int, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow, "endRowIndex": endRow,
				"startColumnIndex": startCol, "endColumnIndex": endCol,
			},
			"top":    map[string]any{"style": "SOLID", "width": width},
			"bottom": map[string]any{"style": "SOLID", "width": width},
			"left":   map[string]any{"style": "SOLID", "width": width},
			"right":  map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func internalVertical(startRow, startCol, endRow, endCol int, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow, "endRowIndex": endRow,
				"startColumnIndex": startCol, "endColumnIndex": endCol,
			},
			"right": map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func addCF(formula string, r, g, b float64) map[string]any {
	rangeDays := []any{
		map[string]any{
			"sheetId":          sheetHE,
			"startRowIndex":    6, "endRowIndex": 30,
			"startColumnIndex": 4, "endColumnIndex": 35,
		},
	}
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
							"red": r, "green": g, "blue": b,
						},
					},
				},
			},
		},
	}
}

func addCFFeriado(dateRef string) map[string]any {
	rangeDays := []any{
		map[string]any{
			"sheetId":          sheetHE,
			"startRowIndex":    6, "endRowIndex": 30,
			"startColumnIndex": 4, "endColumnIndex": 35,
		},
	}
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
							"red": 1.0, "green": 0.84, "blue": 0.0,
						},
						"textFormat": map[string]any{
							"foregroundColor": map[string]any{
								"red": 0.7, "green": 0.0, "blue": 0.0,
							},
							"bold": true,
						},
					},
				},
			},
		},
	}
}

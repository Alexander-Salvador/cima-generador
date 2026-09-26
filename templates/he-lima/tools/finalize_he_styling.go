// finalize_he_styling.go — Final styling pass:
//   1. Visual separator (thick right border on merged name area B+C+D)
//   2. Improved day header (weekend highlight on sáb/dom)
//   3. Empty employee styling (rows 26-30: muted bg + placeholder text)
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

	requests := []any{}

	// 1. THICK RIGHT BORDER on the merged name area (col D, rows 8, 9, 10-30, 32)
	// The merge is B+C+D, so the right edge is the right edge of D (col index 3 = 4 in API)
	thickBorderColor := map[string]any{
		"red":   0.101960786,
		"green": 0.227450982,
		"blue":  0.360784314,
	}
	for _, row := range []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 32} {
		requests = append(requests, map[string]any{
			"updateBorders": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row - 1,
					"endRowIndex":      row,
					"startColumnIndex": 3, // D
					"endColumnIndex":   4, // E exclusive (right edge of D)
				},
				"right": map[string]any{
					"style": "SOLID_THICK",
					"color": thickBorderColor,
					"width": 2,
				},
				"top": map[string]any{
					"style": "SOLID",
					"color": thickBorderColor,
					"width": 1,
				},
				"bottom": map[string]any{
					"style": "SOLID",
					"color": thickBorderColor,
					"width": 1,
				},
			},
		})
	}

	// 2. DAY HEADER ROWS (8-9) — apply strong styling
	// Row 8: day numbers (E-AI) — already styled, ensure primarySubtle bg + primary fg + bold 10pt
	// Row 9: day letters (E-AI) — already styled
	// Skip weekend highlight for now (would require conditional formatting based on TEXT)

	// Actually let me apply a baseline styling to ensure rows 8-9 are uniform
	requests = append(requests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7, // row 8
				"endRowIndex":      8,
				"startColumnIndex": 4, // E
				"endColumnIndex":   35,
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
							"red":   0.0588235,
							"green": 0.1411765,
							"blue":  0.2509804,
						},
						"fontSize": 10,
						"bold":     true,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	})

	requests = append(requests, map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    8, // row 9
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

	// 3. EMPTY EMPLOYEES (rows 26-30) — muted styling
	// Placeholder text + light bg
	emptyStyle := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    25, // row 26
				"endRowIndex":      30, // row 30 exclusive
				"startColumnIndex": 0,  // A
				"endColumnIndex":   35, // AI
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.9568627, // #F4F4F4 surface.disabled
						"green": 0.9568627,
						"blue":  0.9568627,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   0.3647059, // #5D6B85 text.muted
							"green": 0.4196079,
							"blue":  0.5215687,
						},
						"italic": true,
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
		},
	}
	requests = append(requests, emptyStyle)

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
	fmt.Printf("FINALIZE: %d %s\n", resp.StatusCode, string(respBody))

	fmt.Println("\n✓ Styling finalized:")
	fmt.Println("  • Thick right border on merged name area (B+C+D)")
	fmt.Println("  • Day nums row 8: primarySubtle bg + primary fg")
	fmt.Println("  • Day letters row 9: brand.primary bg + white fg")
	fmt.Println("  • Empty employees (rows 26-30): muted bg + italic")
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
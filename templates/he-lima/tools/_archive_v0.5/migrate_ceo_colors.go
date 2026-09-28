// fix_colors_he.go — migrate CEO-introduced colors to existing tokens.json palette
// per CEO directive "Migrar a tokens existentes" (no new tokens added).
//
// CEO introduced these non-token colors in 10_HE:
//   #0F172A (dark navy) — title right extension + KPI hero fg
//   #10B981 (vivid green) — success sub-label
//   #F59E0B (amber) — neutral sub-label
//   #F1F5F9 (light cyan bg) — sub-label backgrounds
//
// Migration:
//   #0F172A → #1A3A5C (brand.primary) for title bg, #0F2440 (text.primary) for KPI fg
//   #10B981 → #1FB6A8 (accent.success)
//   #F59E0B → #E67E22 (accent.alert)
//   #F1F5F9 → #D4F1ED (accent.successSoft) for K3, #FDEBD0 (accent.alertSoft) for S3
//   Remove dark navy bg from U1:AH1 (monochromatic title per CEO decision)
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

	// 1. Remove dark navy bg from U1:AH1 (set to default = no bg)
	//    This makes title bar monochromatic (B1:S1 brand.primary only)
	removeTitleStrip := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 20, // U
				"endColumnIndex":   34, // AH (exclusive)
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   1.0,
						"green": 1.0,
						"blue":  1.0,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   0.0,
							"green": 0.0,
							"blue":  0.0,
						},
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
		},
	}

	// 2. K2 "S/ 0" — fg #0F172A → #0F2440 (text.primary)
	// 3. S2 "21" — fg #0F172A → #0F2440 (text.primary)
	migrateK2S2 := []map[string]any{
		// K2 (col 10, row 1)
		{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    1,
					"endRowIndex":      2,
					"startColumnIndex": 10,
					"endColumnIndex":   11,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"textFormat": map[string]any{
							"foregroundColor": map[string]any{
								"red":   0.0588235, // #0F2440
								"green": 0.1411765,
								"blue":  0.2509804,
							},
						},
					},
				},
				"fields": "userEnteredFormat.textFormat.foregroundColor",
			},
		},
		// S2 (col 18, row 1)
		{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    1,
					"endRowIndex":      2,
					"startColumnIndex": 18,
					"endColumnIndex":   19,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"textFormat": map[string]any{
							"foregroundColor": map[string]any{
								"red":   0.0588235,
								"green": 0.1411765,
								"blue":  0.2509804,
							},
						},
					},
				},
				"fields": "userEnteredFormat.textFormat.foregroundColor",
			},
		},
	}

	// 4. K3 "+0% vs mes anterior" — bg #F1F5F9 → #D4F1ED (successSoft), fg #10B981 → #1FB6A8 (success)
	migrateK3 := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    2,
				"endRowIndex":      3,
				"startColumnIndex": 10,
				"endColumnIndex":   11,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.8313726, // #D4F1ED accent.successSoft
						"green": 0.9450980,
						"blue":  0.9294118,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   0.1215686, // #1FB6A8 accent.success
							"green": 0.7137255,
							"blue":  0.6588235,
						},
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
		},
	}

	// 5. S3 "TODOS EN PLANILLA" — bg #F1F5F9 → #FDEBD0 (alertSoft), fg #F59E0B → #E67E22 (alert)
	migrateS3 := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    2,
				"endRowIndex":      3,
				"startColumnIndex": 18,
				"endColumnIndex":   19,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": map[string]any{
						"red":   0.9921569, // #FDEBD0 accent.alertSoft
						"green": 0.9215686,
						"blue":  0.8156863,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   0.9019608, // #E67E22 accent.alert
							"green": 0.4941176,
							"blue":  0.1333333,
						},
					},
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
		},
	}

	all := []map[string]any{removeTitleStrip}
	all = append(all, migrateK2S2...)
	all = append(all, migrateK3, migrateS3)

	for i, req := range all {
		body := map[string]any{"requests": []any{req}}
		raw, _ := json.Marshal(body)
		post, _ := http.NewRequest("POST",
			"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
			bytes.NewReader(raw))
		post.Header.Set("Authorization", "Bearer "+tok)
		post.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(post)
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("[%d] %d %s\n", i+1, resp.StatusCode, string(respBody))
	}

	fmt.Println("\n✓ Color migration applied")
	fmt.Println("Tokens used:")
	fmt.Println("  • #1A3A5C brand.primary  (title bar B1:S1, unchanged)")
	fmt.Println("  • #0F2440 text.primary    (KPI hero numbers K2/S2)")
	fmt.Println("  • #D4F1ED accent.successSoft (K3 success bg)")
	fmt.Println("  • #1FB6A8 accent.success   (K3 +0% green text)")
	fmt.Println("  • #FDEBD0 accent.alertSoft (S3 warning bg)")
	fmt.Println("  • #E67E22 accent.alert    (S3 neutral text)")
	fmt.Println("  • U1:AH1 cleared (monochromatic title)")
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
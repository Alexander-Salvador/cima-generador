// fix_yana2.go — Re-add YANA in row 33 (with fresh token)
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

	// Get fresh token
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
	tok, _ := res["access_token"].(string)
	fmt.Printf("Got fresh token: %s...\n", tok[:20])

	// Set YANA in row 33
	reqs := []any{
		map[string]any{
			"updateCells": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    32,
					"endRowIndex":      33,
					"startColumnIndex": 0,
					"endColumnIndex":   1,
				},
				"rows": []any{
					map[string]any{
						"values": []any{
							map[string]any{
								"userEnteredValue": map[string]any{"numberValue": float64(25)},
							},
						},
					},
				},
				"fields": "userEnteredValue",
			},
		},
		map[string]any{
			"updateCells": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    32,
					"endRowIndex":      33,
					"startColumnIndex": 1,
					"endColumnIndex":   2,
				},
				"rows": []any{
					map[string]any{
						"values": []any{
							map[string]any{
								"userEnteredValue": map[string]any{"stringValue": "YANA JUNCO ROSALIA ROCIO"},
							},
						},
					},
				},
				"fields": "userEnteredValue",
			},
		},
	}

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
		fmt.Println("ERROR:", string(respB))
	} else {
		fmt.Println("✅ YANA re-added in row 33")
	}
}

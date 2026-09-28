// clear_all_bgs.go — Clear all background colors
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

var allCols = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q",
	"R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI",
	"AJ", "AK", "AL", "AM", "AN", "AO", "AP",
}

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
	fmt.Println("Token:", tok[:20], "...")

	reqs := []any{}
	// Clear all bgs
	for row := 1; row <= 36; row++ {
		for _, col := range allCols {
			ci := colIdx(col)
			ri := row - 1
			reqs = append(reqs, map[string]any{
				"updateCells": map[string]any{
					"range": map[string]any{
						"sheetId": sheetHE,
						"startRowIndex": ri, "endRowIndex": ri + 1,
						"startColumnIndex": ci, "endColumnIndex": ci + 1,
					},
					"rows": []any{
						map[string]any{"values": []any{
							map[string]any{
								"userEnteredFormat": map[string]any{
									"backgroundColor": map[string]any{"red": 1.0, "green": 1.0, "blue": 1.0},
								},
							},
						}},
					},
					"fields": "userEnteredFormat.backgroundColor",
				},
			})
		}
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
		fmt.Println("ERROR:", string(respB)[:500])
	} else {
		fmt.Printf("✅ Cleared all bgs (%d requests)\n", len(reqs))
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

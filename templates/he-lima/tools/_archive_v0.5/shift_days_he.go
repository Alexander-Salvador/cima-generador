// shift_days_he.go — Clean re-runnable script to shift day data D:AH → E:AI.
// This complements merge_name_cells_he.go by writing the actual day DATA.
// Use this if data was lost or not persisted.
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
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	creds := decrypt()
	tok := getToken(creds)

	// Step 1: Read original D10:AH30 data (may be empty if already shifted)
	get := func(rng string) [][]any {
		req, _ := http.NewRequest("GET",
			"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+"/values/"+rng, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var raw map[string]any
		json.Unmarshal(body, &raw)
		vs, _ := raw["values"].([]any)
		out := [][]any{}
		for _, v := range vs {
			row, _ := v.([]any)
			out = append(out, row)
		}
		return out
	}

	// If source D10:AH30 is empty (already shifted), read from E10:AI30 instead.
	srcData := get("10_HE!D10:AH30")
	if len(srcData) == 0 {
		fmt.Println("Source D10:AH30 is empty (already shifted?) — reading from E10:AI30")
		srcData = get("10_HE!E10:AI30")
	}
	fmt.Printf("Got %d rows of source data\n", len(srcData))

	// Pad to 21 rows
	for len(srcData) < 21 {
		srcData = append(srcData, []any{})
	}

	// Pad each row to 31 cells (some rows may have fewer trailing empty cells)
	for i, row := range srcData {
		for len(row) < 31 {
			srcData[i] = append(row, "")
		}
	}

	// Step 2: Write to E10:AI30
	put := func(rng string, rows [][]any) {
		body := map[string]any{
			"range":          rng,
			"majorDimension": "ROWS",
			"values":         rows,
		}
		raw, _ := json.Marshal(body)
		url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
			"/values/" + rng + "?valueInputOption=USER_ENTERED"
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("PUT %s: %d\n", rng, resp.StatusCode)
	}

	put("10_HE!E10:AI30", srcData)

	// Step 3: Clear D10:AH30
	clear := func(rng string) {
		url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
			"/values/" + rng + ":clear"
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("CLEAR %s: %d\n", rng, resp.StatusCode)
	}
	clear("10_HE!D10:AH30")
	clear("10_HE!D10:AH30")

	// Apply number format 0.00 to E10:AI30
	batchBody := map[string]any{
		"requests": []any{
			map[string]any{
				"repeatCell": map[string]any{
					"range": map[string]any{
						"sheetId":          222789142,
						"startRowIndex":    9,
						"endRowIndex":      30,
						"startColumnIndex": 4, // E
						"endColumnIndex":   35, // AI exclusive
					},
					"cell": map[string]any{
						"userEnteredFormat": map[string]any{
							"numberFormat": map[string]any{
								"type":    "NUMBER",
								"pattern": "0.00",
							},
						},
					},
					"fields": "userEnteredFormat.numberFormat",
				},
			},
		},
	}
	bRaw, _ := json.Marshal(batchBody)
	bReq, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(bRaw))
	bReq.Header.Set("Authorization", "Bearer "+tok)
	bReq.Header.Set("Content-Type", "application/json")
	bResp, _ := http.DefaultClient.Do(bReq)
	_, _ = io.ReadAll(bResp.Body)
	bResp.Body.Close()
	fmt.Printf("FORMAT: %d\n", bResp.StatusCode)

	fmt.Println("\n✓ Day data shifted D:AH → E:AI")
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
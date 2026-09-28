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
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	r.Body.Close()
	tok, _ := res["access_token"].(string)

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/HORAS%%20EXTRAS!A33:AP36",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)
	values, _ := out["values"].([]any)
	if values == nil {
		fmt.Println("No values returned")
		return
	}
	for _, row := range values {
		cells := row.([]any)
		for ci, v := range cells {
			if v == nil {
				continue
			}
			letter := ""
			idx := ci + 1
			for idx > 0 {
				idx--
				letter = string(rune('A'+idx%26)) + letter
				idx = idx / 26
			}
			fmt.Printf("  %s: %v\n", letter, v)
		}
	}
}

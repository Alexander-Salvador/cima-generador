// check_cols_b_d.go — Check what's in cols B, C, D
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
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	r.Body.Close()
	tok := res["access_token"].(string)

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?includeGridData=true&ranges=HORAS%%20EXTRAS!A9:E10&fields=sheets(data(rowData(values(formattedValue,userEnteredValue,effectiveFormat(backgroundColor,horizontalAlignment,wrapStrategy)))))",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets := out["sheets"].([]any)
	rowData := sheets[0].(map[string]any)["data"].([]any)[0].(map[string]any)["rowData"].([]any)
	for ri, r := range rowData {
		rm := r.(map[string]any)
		cells := rm["values"].([]any)
		actualRow := ri + 9
		fmt.Printf("=== ROW %d ===\n", actualRow)
		for ci, cAny := range cells {
			if cAny == nil {
				continue
			}
			c, ok := cAny.(map[string]any)
			if !ok {
				continue
			}
			letter := ""
			idx := ci + 1
			for idx > 0 {
				idx--
				letter = string(rune('A'+idx%26)) + letter
				idx = idx / 26
			}
			fv, _ := c["formattedValue"].(string)
			ue := c["userEnteredValue"]
			ef := c["effectiveFormat"]
			bgStr := ""
			if ef != nil {
				if efM, ok := ef.(map[string]any); ok {
					if bg, ok := efM["backgroundColor"].(map[string]any); ok {
						if r, ok := bg["red"]; ok && r != nil {
							bgStr = fmt.Sprintf(" bg=(%d,%d,%d)", int(r.(float64)*255), int(bg["green"].(float64)*255), int(bg["blue"].(float64)*255))
						}
					}
				}
			}
			halign := ""
			if ef != nil {
				if efM, ok := ef.(map[string]any); ok {
					if ha, ok := efM["horizontalAlignment"].(string); ok {
						halign = " h=" + ha
					}
				}
			}
			fmt.Printf("  %s: %q ue=%v%s%s\n", letter, fv, ue, halign, bgStr)
		}
	}
}

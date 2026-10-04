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
	encPath = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath = "/home/braka/.config/ovav/vault.key"
	ssID    = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
)

func main() {
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	c := map[string]any{}
	json.Unmarshal(pt, &c)
	body := map[string]any{"client_id": c["client_id"], "client_secret": c["client_secret"], "refresh_token": c["refresh_token"], "grant_type": "refresh_token"}
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok := res["access_token"].(string)

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(merges,properties(title,sheetId),protectedRanges,data(rowData(values(userEnteredValue,formattedValue))))", ssID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	var data map[string]any
	json.Unmarshal(rb, &data)
	sheets := data["sheets"].([]any)
	for _, s := range sheets {
		sm := s.(map[string]any)
		props := sm["properties"].(map[string]any)
		title := props["title"].(string)
		if title == "CONFIG" {
			fmt.Printf("=== %s (sheetId=%v) ===\n", title, props["sheetId"])
			merges, _ := sm["merges"].([]any)
			fmt.Printf("Merges: %d\n", len(merges))
			for _, m := range merges {
				fmt.Printf("  %v\n", m)
			}
			prot, _ := sm["protectedRanges"].([]any)
			fmt.Printf("ProtectedRanges: %d\n", len(prot))
			for _, p := range prot {
				fmt.Printf("  %v\n", p)
			}
			// Dump raw values row 4
			datas, _ := sm["data"].([]any)
			for _, d := range datas {
				dm := d.(map[string]any)
				rd, _ := dm["rowData"].([]any)
				if len(rd) >= 4 {
					r4 := rd[3].(map[string]any)
					values, _ := r4["values"].([]any)
					fmt.Printf("\nRow 4 raw (%d cells):\n", len(values))
					for i, v := range values {
						fmt.Printf("  Col %d: %v\n", i+1, v)
					}
				}
			}
		}
	}
}
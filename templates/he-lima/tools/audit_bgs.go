// audit_bgs.go — Audit all background colors in HORAS EXTRAS
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
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	b, _ := io.ReadAll(r.Body)
	r.Body.Close()
	res := map[string]any{}
	json.Unmarshal(b, &res)
	tok, _ := res["access_token"].(string)

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?includeGridData=true&ranges=HORAS%%20EXTRAS!A1:AP36&fields=sheets(data(rowData(values(effectiveFormat(backgroundColor),formattedValue))))",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ = io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets := out["sheets"].([]any)
	rowData := sheets[0].(map[string]any)["data"].([]any)[0].(map[string]any)["rowData"].([]any)

	// For each row, count unique background colors
	for ri, r := range rowData {
		rm := r.(map[string]any)
		cells := rm["values"].([]any)
		actualRow := ri + 1
		bgCounts := map[string]int{}
		for _, cAny := range cells {
			if cAny == nil {
				continue
			}
			c, ok := cAny.(map[string]any)
			if !ok {
				continue
			}
			efAny, ok := c["effectiveFormat"].(map[string]any)
			if !ok {
				continue
			}
			bgAny, ok := efAny["backgroundColor"].(map[string]any)
			if !ok {
				continue
			}
			r, okR := bgAny["red"].(float64)
			g, okG := bgAny["green"].(float64)
			b, okB := bgAny["blue"].(float64)
			if !okR || !okG || !okB {
				continue
			}
			key := fmt.Sprintf("(%d,%d,%d)", int(r*255), int(g*255), int(b*255))
			bgCounts[key]++
		}
		if len(bgCounts) > 0 {
			fmt.Printf("R%d: ", actualRow)
			for k, v := range bgCounts {
				fmt.Printf("%s=%d ", k, v)
			}
			fmt.Println()
		}
	}
}

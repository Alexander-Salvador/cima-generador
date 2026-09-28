// fix_kpi_ranges.go — Fix KPI card formulas to use full range B9:B33 (25 employees)
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
	tok, _ := res["access_token"].(string)

	reqs := []any{
		// KPI Card 1: B3 = SUM(AJ9:AJ33) H_extra totales
		map[string]any{
			"updateCells": map[string]any{
				"range": map[string]any{
					"sheetId": sheetHE, "startRowIndex": 2, "endRowIndex": 3,
					"startColumnIndex": 1, "endColumnIndex": 2,
				},
				"rows": []any{
					map[string]any{"values": []any{
						map[string]any{"userEnteredValue": map[string]any{"formulaValue": "=SUM(AJ9:AJ33)"}},
					}},
				},
				"fields": "userEnteredValue",
			},
		},
		// KPI Card 2: J3 = COUNTA(B9:B33) count
		map[string]any{
			"updateCells": map[string]any{
				"range": map[string]any{
					"sheetId": sheetHE, "startRowIndex": 2, "endRowIndex": 3,
					"startColumnIndex": 9, "endColumnIndex": 10,
				},
				"rows": []any{
					map[string]any{"values": []any{
						map[string]any{"userEnteredValue": map[string]any{"formulaValue": "=COUNTA(B9:B33)"}},
					}},
				},
				"fields": "userEnteredValue",
			},
		},
		// KPI Card 3: T3 = SUM(AP9:AP33) $_E totales
		map[string]any{
			"updateCells": map[string]any{
				"range": map[string]any{
					"sheetId": sheetHE, "startRowIndex": 2, "endRowIndex": 3,
					"startColumnIndex": 19, "endColumnIndex": 20,
				},
				"rows": []any{
					map[string]any{"values": []any{
						map[string]any{"userEnteredValue": map[string]any{"formulaValue": "=SUM(AP9:AP33)"}},
					}},
				},
				"fields": "userEnteredValue",
			},
		},
		// Also fix individual formulas AJ9:AJ33, AL9:AL33, AN9:AN33, AP9:AP33
		// (rows 9-33 for the 25 agents)
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
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		fmt.Println("ERROR:", string(b))
	} else {
		fmt.Println("✅ KPI formulas updated to use rows 9-33 (25 employees)")
	}
}

// fix_widths_cima_team.go — Corregir widths de HORAS EXTRAS post-transform.
//
// Estado actual (después de transform_cima_to_he_only):
//   E=50 (anomalía, debería 22)
//   AI=22 (debería 6)
//   AJ=2, AK=50, AL=2, AM=50 (alternados visible)
//   AN=2(hidden), AO=50(hidden), AP=2(hidden), AQ=50(hidden) (alternados hidden)
//   AR,AS,AT = eliminadas
//
// Estado deseado:
//   E=22, F-AH=22 (días 1-29)
//   AI=6 (spacer)
//   AJ=55 (H_extra), AK=2, AL=55 (H_25), AM=2, AN=55 (H_35), AO=2, AP=65 ($ HE)
//   Sin hidden
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
	reqs := []any{}

	// E=22 (was 50 anomaly)
	reqs = append(reqs, resizeCol("E", 22))

	// AI=6 (was 22)
	reqs = append(reqs, resizeCol("AI", 6))

	// AJ=55 (H_extra)
	reqs = append(reqs, resizeCol("AJ", 55))
	// AK=2 (spacer)
	reqs = append(reqs, resizeCol("AK", 2))
	// AL=55 (H_25)
	reqs = append(reqs, resizeCol("AL", 55))
	// AM=2 (spacer)
	reqs = append(reqs, resizeCol("AM", 2))
	// AN=55 (H_35)
	reqs = append(reqs, resizeCol("AN", 55))
	// AO=2 (spacer)
	reqs = append(reqs, resizeCol("AO", 2))
	// AP=65 ($ HE)
	reqs = append(reqs, resizeCol("AP", 65))
	// AQ=0 (eliminar col vacía)
	reqs = append(reqs, resizeCol("AQ", 0))

	// Un-hide AN, AO, AP, AQ (and reset widths just in case)
	reqs = append(reqs, unHideCol("AN"))
	reqs = append(reqs, unHideCol("AO"))
	reqs = append(reqs, unHideCol("AP"))
	reqs = append(reqs, unHideCol("AQ"))

	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	respB, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Sent %d requests\n", len(reqs))
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

func resizeCol(col string, pixelSize int) map[string]any {
	ci := colIdx(col)
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetHE,
				"dimension":  "COLUMNS",
				"startIndex": ci,
				"endIndex":   ci + 1,
			},
			"properties": map[string]any{
				"pixelSize": pixelSize,
			},
			"fields": "pixelSize",
		},
	}
}

func unHideCol(col string) map[string]any {
	ci := colIdx(col)
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetHE,
				"dimension":  "COLUMNS",
				"startIndex": ci,
				"endIndex":   ci + 1,
			},
			"properties": map[string]any{
				"hiddenByUser": false,
			},
			"fields": "hiddenByUser",
		},
	}
}

func decrypt() map[string]any {
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	out := map[string]any{}
	json.Unmarshal(pt, &out)
	return out
}

func getToken(c map[string]any) string {
	body := map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	}
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok, _ := res["access_token"].(string)
	return tok
}

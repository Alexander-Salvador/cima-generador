// rewrite_date_formulas_he.go — Rewrite day formulas in E8:AI8 + E9:AI9
// (after the col-shift D:AH → E:AI).
// Same logic as connect_dates_he.go but with new ranges.
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

	monthList := `{"Enero";"Febrero";"Marzo";"Abril";"Mayo";"Junio";"Julio";"Agosto";"Septiembre";"Octubre";"Noviembre";"Diciembre"}`

	startCol := 4 // E
	numCols := 31

	dayFormulas := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		dayNum := i + 1
		dayFormulas[i] = fmt.Sprintf(
			`=IFERROR(IF(MONTH(DATE($C$5,MATCH($C$6,%s,0),%d))=MATCH($C$6,%s,0),DATE($C$5,MATCH($C$6,%s,0),%d),""),"")`,
			monthList, dayNum, monthList, monthList, dayNum)
	}

	letterFormulas := make([]string, numCols)
	colLetter := func(idx int) string {
		result := ""
		for idx >= 0 {
			result = string(rune('A'+idx%26)) + result
			idx = idx/26 - 1
		}
		return result
	}
	for i := 0; i < numCols; i++ {
		col := colLetter(startCol + i)
		letterFormulas[i] = fmt.Sprintf(`=IF(%s8="","",TEXT(%s8,"DDD"))`, col, col)
	}

	dayMatrix := [][]interface{}{}
	for _, f := range dayFormulas {
		dayMatrix = append(dayMatrix, []interface{}{f})
	}
	letterMatrix := [][]interface{}{}
	for _, f := range letterFormulas {
		letterMatrix = append(letterMatrix, []interface{}{f})
	}

	put := func(rng string, matrix [][]interface{}) {
		body := map[string]any{
			"range":          rng,
			"majorDimension": "COLUMNS",
			"values":         matrix,
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

	put("10_HE!E8", dayMatrix)
	put("10_HE!E9", letterMatrix)

	// Number format 'd' on E8:AI8
	batchBody := map[string]any{
		"requests": []any{
			map[string]any{
				"repeatCell": map[string]any{
					"range": map[string]any{
						"sheetId":          222789142,
						"startRowIndex":    7,
						"endRowIndex":      8,
						"startColumnIndex": 4,  // E
						"endColumnIndex":   35, // AI exclusive
					},
					"cell": map[string]any{
						"userEnteredFormat": map[string]any{
							"numberFormat": map[string]any{
								"type":    "DATE",
								"pattern": "d",
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

	fmt.Println("✓ Day formulas rewritten at E8:AI8 + E9:AI9")
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
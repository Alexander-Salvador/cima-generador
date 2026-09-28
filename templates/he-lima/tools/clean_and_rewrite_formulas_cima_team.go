// clean_and_rewrite_formulas_cima_team.go — Clean ALL formulas and rewrite fresh.
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

	// STEP 1: Clear ALL formulas in summary columns (AJ10:AT30)
	// Write empty string to clear formulas
	for row := 10; row <= 30; row++ {
		for _, col := range []string{"AJ", "AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT"} {
			reqs = append(reqs, clearCell(col, row))
		}
	}

	// STEP 2: Clear TOTALES row 36
	for _, col := range []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI", "AJ", "AL", "AN", "AP", "AR", "AT"} {
		reqs = append(reqs, clearCell(col, 36))
	}

	// Send batch 1 (clear)
	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR clear:", err)
		os.Exit(1)
	}
	respB, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR clear", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Cleared %d requests\n", len(reqs))

	// STEP 3: Re-write formulas in correct columns
	reqs = []any{}
	const rate = 15.0
	for row := 10; row <= 30; row++ {
		// AJ = H_norm
		reqs = append(reqs, setFormula("AJ", row,
			fmt.Sprintf(`=SUMPRODUCT((N(E%d:AI%d))*(N(E%d:AI%d)<=8))`, row, row, row, row)))
		// AL = H_extra
		reqs = append(reqs, setFormula("AL", row,
			fmt.Sprintf(`=SUMPRODUCT((N(E%d:AI%d))*(N(E%d:AI%d)>8))`, row, row, row, row)))
		// AN = H_25 (Sunday hours)
		reqs = append(reqs, setFormula("AN", row,
			fmt.Sprintf(`=SUMPRODUCT((E$9:AI$9="D")*(N(E%d:AI%d)))`, row, row)))
		// AP = H_35 (Holiday hours)
		reqs = append(reqs, setFormula("AP", row,
			fmt.Sprintf(`=SUMPRODUCT((ISNUMBER(VLOOKUP(E$8,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE)))*(N(E%d:AI%d)))`, row, row)))
		// AR = $_N (normal pay)
		reqs = append(reqs, setFormula("AR", row,
			fmt.Sprintf(`=AJ%d*%g`, row, rate)))
		// AT = $_E (extra pay)
		reqs = append(reqs, setFormula("AT", row,
			fmt.Sprintf(`=AL%d*%g*1.25+AN%d*%g*1.25+AP%d*%g*1.35`,
				row, rate, row, rate, row, rate)))
	}

	// TOTALES row 36
	for _, col := range []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"} {
		reqs = append(reqs, setFormula(col, 36, fmt.Sprintf(`=SUM(%s10:%s30)`, col, col)))
	}
	reqs = append(reqs, setFormula("AJ", 36, `=SUM(AJ10:AJ30)`))
	reqs = append(reqs, setFormula("AL", 36, `=SUM(AL10:AL30)`))
	reqs = append(reqs, setFormula("AN", 36, `=SUM(AN10:AN30)`))
	reqs = append(reqs, setFormula("AP", 36, `=SUM(AP10:AP30)`))
	reqs = append(reqs, setFormula("AR", 36, `=SUM(AR10:AR30)`))
	reqs = append(reqs, setFormula("AT", 36, `=SUM(AT10:AT30)`))

	// Send batch 2 (write)
	body2 := map[string]any{"requests": reqs}
	bb2, _ := json.Marshal(body2)
	req2, _ := http.NewRequest("POST", url, bytes.NewReader(bb2))
	req2.Header.Set("Authorization", "Bearer "+tok)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println("ERROR write:", err)
		os.Exit(1)
	}
	respB2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if resp2.StatusCode >= 300 {
		fmt.Println("API ERROR write", resp2.StatusCode, ":", string(respB2))
		os.Exit(1)
	}
	fmt.Printf("✅ Wrote %d requests\n", len(reqs))
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

func rowIdx(cell string) int {
	row := 0
	for _, c := range cell {
		if c >= '0' && c <= '9' {
			row = row*10 + int(c-'0')
		}
	}
	return row - 1
}

func clearCell(col string, row int) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": ""},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func setFormula(col string, row int, formula string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"formulaValue": formula},
						},
					},
				},
			},
			"fields": "userEnteredValue",
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
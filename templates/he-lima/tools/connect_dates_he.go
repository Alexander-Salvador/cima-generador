// connect_dates_he.go — connect Año/Mes selectors to dynamic dates + day letters
// CEO directive: 'primero conecta el año y mes con la tabla debe funcionar y
// cambiar dinamicamente'
//
// Layout:
//   C5 = Año (dropdown 2025-2030)
//   C6 = Mes (dropdown Enero...Diciembre)
//
//   D8 = first day of month (DATE)
//   D9 = day-letter (TEXT DDD)
//
//   E8:AH8 = subsequent days (D8+1, D8+2, ...)
//   E9:AH9 = corresponding day-letters
//
// Edge cases:
//   - Feb non-leap (28 days): cells AF8:AH8 stay empty (no Mar 1, 2)
//   - 30-day months: cell AH8 (day 30) shown, no day 31
//   - 31-day months: cell AH8 shows day 31
//
// Implementation:
//   Use IF(MONTH(D8+n) = MONTH(D8), D8+n, "") to gate per-month.
//   This means DATE values can extend past month-end but display empty.
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

	// Spanish month names → MATCH formula
	// Note: in Spanish Sheets, array literal uses ; for rows.
	monthList := `{"Enero";"Febrero";"Marzo";"Abril";"Mayo";"Junio";"Julio";"Agosto";"Septiembre";"Octubre";"Noviembre";"Diciembre"}`

	// D8 = first day of (C5, C6, 1)
	// E8:AH8 = D8+1, D8+2, ... but only if still in same month as D8
	// D9:AH9 = day letter, "" if D8:AH8 is ""

	// Build the formulas matrix
	// Column D is index 3 (0-based). Range D8:AH8 is 31 cells (cols 3..33).
	// Row 7 (0-based) = row 8 (1-based).

	const startCol = 3   // D
	const endCol = 34    // AH (exclusive)
	const numCols = endCol - startCol // 31
	const rowDayNums = 7  // 0-based row 7 = 1-based row 8
	const rowDayLetters = 8

	// Day-of-month for D8 = 1, E8 = 2, ..., AH8 = 31
	dayFormulas := make([]string, numCols)
	letterFormulas := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		dayNum := i + 1
		// D8 = DATE(C5, MATCH(C6, list, 0), 1)
		// E8 = DATE(C5, month_num, 2) — using DATE with day-arg guarantees correct date
		// For day 29, 30, 31: if invalid (e.g., Feb 30), DATE returns the next month's date
		// To guard: wrap with IF(MONTH(...)=MATCH(C6,list,0), DATE(...), "")
		if dayNum == 1 {
			dayFormulas[i] = fmt.Sprintf(
				`=IFERROR(IF(MONTH(DATE($C$5,MATCH($C$6,%s,0),1))=MATCH($C$6,%s,0),DATE($C$5,MATCH($C$6,%s,0),1),""),"")`,
				monthList, monthList, monthList)
		} else {
			dayFormulas[i] = fmt.Sprintf(
				`=IFERROR(IF(MONTH(DATE($C$5,MATCH($C$6,%s,0),%d))=MATCH($C$6,%s,0),DATE($C$5,MATCH($C$6,%s,0),%d),""),"")`,
				monthList, dayNum, monthList, monthList, dayNum)
		}
		letterFormulas[i] = fmt.Sprintf(`=IF(D%d="","",TEXT(D%d,"DDD"))`, rowDayNums+1, rowDayNums+1)
		// Use the actual col letter instead
		col := colLetterFromIndex(startCol + i)
		letterFormulas[i] = fmt.Sprintf(`=IF(%s%d="","",TEXT(%s%d,"DDD"))`,
			col, rowDayNums+1, col, rowDayNums+1)
	}

	// Build the values matrix
	dayMatrix := [][]interface{}{}
	for _, f := range dayFormulas {
		dayMatrix = append(dayMatrix, []interface{}{f})
	}

	letterMatrix := [][]interface{}{}
	for _, f := range letterFormulas {
		letterMatrix = append(letterMatrix, []interface{}{f})
	}

	// Write D8:AH8 formulas
	writeFormula(tok, "D8", dayMatrix)
	writeFormula(tok, "D9", letterMatrix)

	// Apply number format "d" (day-of-month) to D8:AH8 via batchUpdate
	applyNumFormat(tok)

	fmt.Println("\n✓ Dynamic dates + day letters connected")
	fmt.Println("Try changing dropdowns C5 (Año) and C6 (Mes) — D8:AH9 update automatically.")
}

func colLetterFromIndex(idx int) string {
	result := ""
	for idx >= 0 {
		result = string(rune('A'+idx%26)) + result
		idx = idx/26 - 1
	}
	return result
}

func writeFormula(tok string, startRange string, matrix [][]interface{}) {
	body := map[string]any{
		"range":          "10_HE!" + startRange + ":AH" + startRange[1:],
		"majorDimension": "COLUMNS",
		"values":         matrix,
	}
	raw, _ := json.Marshal(body)
	url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
		"/values/10_HE!" + startRange + ":AH" + startRange[1:] + "?valueInputOption=USER_ENTERED"
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("PUT %s: %d %s\n", startRange, resp.StatusCode, string(respBody))
}

func applyNumFormat(tok string) {
	req := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7, // row 8
				"endRowIndex":      8,
				"startColumnIndex": 3, // D
				"endColumnIndex":   34,
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
	}
	body := map[string]any{"requests": []any{req}}
	raw, _ := json.Marshal(body)
	post, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(raw))
	post.Header.Set("Authorization", "Bearer "+tok)
	post.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(post)
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("Number format: %d %s\n", resp.StatusCode, string(respBody))
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
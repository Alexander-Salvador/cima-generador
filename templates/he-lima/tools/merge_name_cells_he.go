// merge_name_cells_he.go — Merge B+C+D per row in 10_HE to fit long employee names.
// Also: shift day columns from D:AH → E:AI, add N° in col A, improve day headers.
//
// CEO directive: 'Row 10, tomamos en la row 10 column b,c,d y combinamos celdas'
// We apply this to rows 8, 9, 10-30, 32 so headers + data + totales all align.
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

	// Step 1: Read current values/formulas in D8:AH9 (headers) + D10:AH30 (data) + D32:AH32 (totales)
	readRange := func(rng string) [][]any {
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

	headersDayNums := readRange("10_HE!D8:AH8")
	headersDayLetters := readRange("10_HE!D9:AH9")
	dataRows := readRange("10_HE!D10:AH30")
	totalesRow := readRange("10_HE!D32:AH32")

	fmt.Printf("Read %d day-num cells in row 8\n", len(headersDayNums[0]))
	fmt.Printf("Read %d day-letter cells in row 9\n", len(headersDayLetters[0]))
	fmt.Printf("Read %d data rows\n", len(dataRows))
	fmt.Printf("Read %d totales cells\n", func() int {
		if len(totalesRow) == 0 {
			return 0
		}
		return len(totalesRow[0])
	}())

	// Step 2: Write shifted data to E:AI
	// Format: each row is a list of values. The shifted row has 31 cells (matching D:AH).
	shiftAndWrite := func(rng string, rows [][]any) {
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
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("PUT %s: %d %s\n", rng, resp.StatusCode, string(respBody))
	}

	// Headers (rows 8-9): just shift formulas
	// Data (rows 10-30): shift values
	// Totales (row 32): shift formula

	shiftAndWrite("10_HE!E8:AI8", headersDayNums)
	shiftAndWrite("10_HE!E9:AI9", headersDayLetters)

	// Pad data rows to 21 (rows 26-30 are empty in current state)
	for len(dataRows) < 21 {
		dataRows = append(dataRows, []any{})
	}
	shiftAndWrite("10_HE!E10:AI30", dataRows)

	// Totales row may be empty — write empty matrix (just clears E32:AI32)
	if len(totalesRow) == 0 {
		totalesRow = [][]any{{}}
	}
	shiftAndWrite("10_HE!E32:AI32", totalesRow)

	// Step 3: Clear D:AH in rows 8, 9, 10-30, 32
	clearRange := func(rng string) {
		body := map[string]any{}
		raw, _ := json.Marshal(body)
		url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
			"/values/" + rng + ":clear"
		req, _ := http.NewRequest("POST", url, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("CLEAR %s: %d %s\n", rng, resp.StatusCode, string(respBody))
	}

	clearRange("10_HE!D8:AH8")
	clearRange("10_HE!D9:AH9")
	clearRange("10_HE!D10:AH30")
	clearRange("10_HE!D32:AH32")

	// Step 4: Merge B+C+D in rows 8, 9, 10-30, 32
	mergeRequests := []any{}
	for _, row := range []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 32} {
		mergeRequests = append(mergeRequests, map[string]any{
			"mergeCells": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row - 1,
					"endRowIndex":      row,
					"startColumnIndex": 1, // B
					"endColumnIndex":   4, // E (exclusive)
				},
				"mergeType": "MERGE_ALL",
			},
		})
	}

	mergeBody := map[string]any{"requests": mergeRequests}
	mergeRaw, _ := json.Marshal(mergeBody)
	mergeReq, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(mergeRaw))
	mergeReq.Header.Set("Authorization", "Bearer "+tok)
	mergeReq.Header.Set("Content-Type", "application/json")
	mergeResp, _ := http.DefaultClient.Do(mergeReq)
	mergeBody2, _ := io.ReadAll(mergeResp.Body)
	mergeResp.Body.Close()
	fmt.Printf("MERGE B+C+D: %d %s\n", mergeResp.StatusCode, string(mergeBody2))

	// Step 5: Set col widths
	// A=30, B=180, C=120, D=100, E-AH=28, AI=8
	widthReqs := []any{}
	widths := map[int]int{
		0:   30,  // A
		1:   180, // B
		2:   120, // C
		3:   100, // D (merged into name)
		4:   28,  // E day 1
		5:   28,
		6:   28,
		7:   28,
		8:   28,
		9:   28,
		10:  28,
		11:  28,
		12:  28,
		13:  28,
		14:  28,
		15:  28,
		16:  28,
		17:  28,
		18:  28,
		19:  28,
		20:  28,
		21:  28,
		22:  28,
		23:  28,
		24:  28,
		25:  28,
		26:  28,
		27:  28,
		28:  28,
		29:  28,
		30:  28,
		31:  28,
		32:  28,
		33:  28,  // AH day 31
		34:  8,   // AI spacer
	}
	for col, w := range widths {
		widthReqs = append(widthReqs, map[string]any{
			"updateDimensionProperties": map[string]any{
				"range": map[string]any{
					"sheetId":   sheetHE,
					"dimension": "COLUMNS",
					"startIndex": col,
					"endIndex":   col + 1,
				},
				"properties": map[string]any{
					"pixelSize": w,
				},
				"fields": "pixelSize",
			},
		})
	}

	widthBody := map[string]any{"requests": widthReqs}
	widthRaw, _ := json.Marshal(widthBody)
	widthReq, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(widthRaw))
	widthReq.Header.Set("Authorization", "Bearer "+tok)
	widthReq.Header.Set("Content-Type", "application/json")
	widthResp, _ := http.DefaultClient.Do(widthReq)
	widthBody2, _ := io.ReadAll(widthResp.Body)
	widthResp.Body.Close()
	fmt.Printf("WIDTHS: %d %s\n", widthResp.StatusCode, string(widthBody2))

	// Step 6: Add N° 1-21 in A10:A30
	numMatrix := [][]any{}
	for i := 1; i <= 21; i++ {
		numMatrix = append(numMatrix, []any{fmt.Sprintf("%d", i)})
	}
	body := map[string]any{
		"range":          "10_HE!A10:A30",
		"majorDimension": "ROWS",
		"values":         numMatrix,
	}
	numRaw, _ := json.Marshal(body)
	numUrl := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
		"/values/10_HE!A10:A30?valueInputOption=USER_ENTERED"
	numReq, _ := http.NewRequest("PUT", numUrl, bytes.NewReader(numRaw))
	numReq.Header.Set("Authorization", "Bearer "+tok)
	numReq.Header.Set("Content-Type", "application/json")
	numResp, _ := http.DefaultClient.Do(numReq)
	numBody, _ := io.ReadAll(numResp.Body)
	numResp.Body.Close()
	fmt.Printf("N°: %d %s\n", numResp.StatusCode, string(numBody))

	fmt.Println("\n✓ Layout restructured:")
	fmt.Println("  • N° in col A (1-21)")
	fmt.Println("  • B+C+D merged per row → wider name (180+120+100 = 400px)")
	fmt.Println("  • Days shifted D:AH → E:AI (1 col right)")
	fmt.Println("  • Col widths tuned for new layout")
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
// full_audit.go — Audit completo del sheet HORAS EXTRAS
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
	"strconv"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetHE       = int64(222789142)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func colLetter(idx int) string {
	result := ""
	for idx > 0 {
		idx--
		result = string(rune('A'+idx%26)) + result
		idx = idx / 26
	}
	return result
}

func getStr(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func main() {
	creds := decrypt()
	tok := getToken(creds)

	// 1. Get CF rules + grid properties
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(gridProperties),conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets := out["sheets"].([]any)
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		sidF, _ := props["sheetId"].(float64)
		if int64(sidF) != sheetHE {
			continue
		}
		fmt.Printf("=== SHEET: %v ===\n", props["title"])
		gp, _ := props["gridProperties"].(map[string]any)
		fmt.Printf("Grid: rows=%v cols=%v freezeRowCount=%v frozenColumnCount=%v\n",
			gp["rowCount"], gp["columnCount"], gp["freezeRowCount"], gp["frozenColumnCount"])

		cf, _ := sm["conditionalFormats"].([]any)
		fmt.Printf("\n=== CONDITIONAL FORMATS (%d) ===\n", len(cf))
		for i, r := range cf {
			rm := r.(map[string]any)
			rangesAny, _ := rm["ranges"].([]any)
			brAny, _ := rm["booleanRule"].(map[string]any)
			condAny, _ := brAny["condition"].(map[string]any)
			ctype := getStr(condAny, "type")
			valsAny, _ := condAny["values"].([]any)
			valStrs := []string{}
			for _, v := range valsAny {
				valStrs = append(valStrs, getStr(v.(map[string]any), "userEnteredValue"))
			}
			for _, rng := range rangesAny {
				rngMap := rng.(map[string]any)
				sidF2 := getFloat(rngMap, "sheetId")
				if int64(sidF2) != sheetHE {
					continue
				}
				fmt.Printf("  [%d] R%d-%d C%d-%d type=%s val=%s\n", i,
					int(getFloat(rngMap, "startRowIndex"))+1,
					int(getFloat(rngMap, "endRowIndex")),
					int(getFloat(rngMap, "startColumnIndex"))+1,
					int(getFloat(rngMap, "endColumnIndex")),
					ctype, valStrs[0])
			}
		}
	}

	// 2. Get column widths, merges, content
	url2 := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?includeGridData=true&ranges=HORAS%%20EXTRAS!A1:AP40&fields=sheets(data(columnMetadata(pixelSize),rowData(values(effectiveFormat(backgroundColor,textFormat),formattedValue,userEnteredValue))),merges)",
		spreadsheetID)
	req2, _ := http.NewRequest("GET", url2, nil)
	req2.Header.Set("Authorization", "Bearer "+tok)
	resp2, _ := http.DefaultClient.Do(req2)
	defer resp2.Body.Close()
	b2, _ := io.ReadAll(resp2.Body)
	out2 := map[string]any{}
	json.Unmarshal(b2, &out2)
	sheets2 := out2["sheets"].([]any)
	for _, sh := range sheets2 {
		sm := sh.(map[string]any)
		data := sm["data"].([]any)
		colMeta := data[0].(map[string]any)["columnMetadata"].([]any)

		total := 0
		for _, c := range colMeta {
			pwF := c.(map[string]any)["pixelSize"].(float64)
			total += int(pwF)
		}
		fmt.Printf("\n=== COLUMN WIDTHS (total: %dpx) ===\n", total)
		for i, c := range colMeta {
			cm := c.(map[string]any)
			pw := int(cm["pixelSize"].(float64))
			fmt.Printf("  %s: %dpx\n", colLetter(i+1), pw)
		}

		merges, _ := sm["merges"].([]any)
		fmt.Printf("\n=== MERGES (%d) ===\n", len(merges))
		for i, m := range merges {
			mm := m.(map[string]any)
			sr := int(getFloat(mm, "startRowIndex")) + 1
			er := int(getFloat(mm, "endRowIndex"))
			sc := int(getFloat(mm, "startColumnIndex")) + 1
			ec := int(getFloat(mm, "endColumnIndex"))
			fmt.Printf("  [%d] R%d-%d C%d-%d (%s:%s)\n", i, sr, er, sc, ec, colLetter(sc), colLetter(ec))
		}

		rowData := data[0].(map[string]any)["rowData"].([]any)
		for _, ri := range []int{6, 7, 8, 9, 35} {
			if ri >= len(rowData) {
				continue
			}
			rm := rowData[ri].(map[string]any)
			cellsAny := rm["values"].([]any)
			fmt.Printf("\n=== ROW %d content (cells=%d) ===\n", ri+1, len(cellsAny))
			for ci := 0; ci < len(cellsAny); ci++ {
				cAny := cellsAny[ci]
				if cAny == nil {
					continue
				}
				c, ok := cAny.(map[string]any)
				if !ok {
					continue
				}
				ueAny := c["userEnteredValue"]
				if ueAny == nil {
					continue
				}
				ue, _ := ueAny.(map[string]any)
				if ue == nil {
					continue
				}
				var v string
				if n := getFloat(ue, "numberValue"); n != 0 {
					v = strconv.FormatFloat(n, 'f', 0, 64)
				} else if s := getStr(ue, "stringValue"); s != "" {
					v = s
				} else if f := getStr(ue, "formulaValue"); f != "" {
					if len(f) > 50 {
						f = f[:50] + "..."
					}
					v = "=" + f
				}
				if v == "" {
					continue
				}
				bgStr := ""
				if efAny := c["effectiveFormat"]; efAny != nil {
					if ef, ok := efAny.(map[string]any); ok {
						if bg, ok := ef["backgroundColor"].(map[string]any); ok {
							if r := bg["red"]; r != nil {
								bgStr = fmt.Sprintf(" bg=(%d,%d,%d)", int(r.(float64)*255), int(bg["green"].(float64)*255), int(bg["blue"].(float64)*255))
							}
						}
					}
				}
				fmt.Printf("  %s: %s%s\n", colLetter(ci+1), v, bgStr)
			}
		}
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

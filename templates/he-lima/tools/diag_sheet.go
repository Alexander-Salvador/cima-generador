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

	target := "FEEDBACK"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?includeGridData=true&ranges=%s!A1:N50&fields=sheets(merges,properties(title,sheetId),protectedRanges,conditionalFormats,data(rowData(values(formattedValue,userEnteredValue,hyperlink))))", ssID, target)
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
		if title != target {
			continue
		}
		fmt.Printf("=== %s (sheetId=%v) ===\n", title, props["sheetId"])

		merges, _ := sm["merges"].([]any)
		fmt.Printf("MERGES: %d\n", len(merges))
		for _, m := range merges {
			fmt.Printf("  %v\n", m)
		}

		cf, _ := sm["conditionalFormats"].([]any)
		if cf == nil {
			cf = []any{}
		}
		fmt.Printf("CONDITIONAL FORMATS: %d\n", len(cf))
		for i, c := range cf {
			fmt.Printf("  [%d] %v\n", i, c)
		}

		prot, _ := sm["protectedRanges"].([]any)
		if prot == nil {
			prot = []any{}
		}
		fmt.Printf("PROTECTED RANGES: %d\n", len(prot))
		for _, p := range prot {
			fmt.Printf("  %v\n", p)
		}

		datas, _ := sm["data"].([]any)
		for _, d := range datas {
			dm := d.(map[string]any)
			rd, _ := dm["rowData"].([]any)
			fmt.Printf("ROWS WITH DATA: %d\n", len(rd))
			for i, row := range rd {
				rm := row.(map[string]any)
				vals, _ := rm["values"].([]any)
				if len(vals) > 0 {
					rowNum := i + 1
					fmt.Printf("  Row %d: %d cells\n", rowNum, len(vals))
					for j, v := range vals {
						vm := v.(map[string]any)
						fv, _ := vm["formattedValue"].(string)
						uv := vm["userEnteredValue"]
						hl, _ := vm["hyperlink"].(string)
						if fv != "" || hl != "" {
							short := fv
							if len(short) > 40 {
								short = short[:40] + "..."
							}
							fmt.Printf("    Col %d (%s): val=%q uv=%v hl=%q\n", j+1, colLetter(j+1), short, uv, hl)
						}
					}
				}
			}
		}
	}
}

func colLetter(n int) string {
	s := ""
	for n > 0 {
		n--
		s = string(rune('A'+n%26)) + s
		n /= 26
	}
	return s
}
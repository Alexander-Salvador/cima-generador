// cleanup_totals.go — Limpiar inputs restantes + mover TOTALES a row 36.
//
//   FASE 1: Limpiar inputs row 9 (ALBA que mi script anterior no tocó)
//   FASE 2: Limpiar row 34 (TOTALES label + datos) y row 35 (vacía)
//   FASE 3: Re-aplicar TOTALES en row 36 (final del DATA frame)
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

var dayCols = []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}

func main() {
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// FASE 1 — Limpiar inputs row 9 (ALBA, no fue tocado antes)
	for _, col := range dayCols {
		reqs = append(reqs, clearCell(col, 9))
	}

	// FASE 2 — Limpiar rows 33, 34, 35 (TOTALES viejos + buffer)
	for row := 33; row <= 35; row++ {
		for _, col := range dayCols {
			reqs = append(reqs, clearCell(col, row))
		}
		// Summary cols
		reqs = append(reqs, clearCell("AJ", row))
		reqs = append(reqs, clearCell("AL", row))
		reqs = append(reqs, clearCell("AN", row))
		reqs = append(reqs, clearCell("AP", row))
		// Label cols
		reqs = append(reqs, clearCell("A", row))
		reqs = append(reqs, clearCell("B", row))
	}

	// FASE 3 — Re-aplicar TOTALES en row 36
	//   A36 = "TOTALES" label
	//   E36:AI36 = SUM(E9:E33) por día
	//   AJ36 = SUM(AJ9:AJ33) H_extra
	//   AL36 = SUM(AL9:AL33) H_25
	//   AN36 = SUM(AN9:AN33) H_35
	//   AP36 = SUM(AP9:AP33) $_E
	reqs = append(reqs, setValue("A", 36, "TOTALES"))

	// Merge A36:B36 for label space
	reqs = append(reqs, mergeRange(35, 0, 36, 2))

	for _, col := range dayCols {
		reqs = append(reqs, setFormula(col, 36, fmt.Sprintf(`=SUM(%s9:%s33)`, col, col)))
	}
	reqs = append(reqs, setFormula("AJ", 36, `=SUM(AJ9:AJ33)`))
	reqs = append(reqs, setFormula("AL", 36, `=SUM(AL9:AL33)`))
	reqs = append(reqs, setFormula("AN", 36, `=SUM(AN9:AN33)`))
	reqs = append(reqs, setFormula("AP", 36, `=SUM(AP9:AP33)`))

	// Format AP36 as moneda
	reqs = append(reqs, setNumberFormat("AP", 36, `"S/."#,##0.00`))

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
	fmt.Printf("✅ Cleaned: row 9 inputs, rows 33-35, re-applied TOTALES in row 36\n")
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

func setValue(col string, row int, value string) map[string]any {
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
							"userEnteredValue": map[string]any{"stringValue": value},
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

func setNumberFormat(col string, row int, pattern string) map[string]any {
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
							"userEnteredFormat": map[string]any{
								"numberFormat": map[string]any{
									"type":    "NUMBER",
									"pattern": pattern,
								},
							},
						},
					},
				},
			},
			"fields": "userEnteredFormat.numberFormat",
		},
	}
}

func mergeRange(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"mergeType": "MERGE_ALL",
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

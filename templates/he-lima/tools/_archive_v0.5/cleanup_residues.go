// cleanup_residues.go — Limpiar residuos en row 8, KPI cards, y E10.
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

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Limpiar residuos en row 8 (debajo de day letters)
	//          AJ8 = "#VALUE!", AK8 = "H_norm", AM8 = "H_ext" etc.
	// ═══════════════════════════════════════════════════════════════════
	for _, col := range []string{"AJ", "AK", "AL", "AM", "AN", "AO", "AP"} {
		reqs = append(reqs, clearCell(col, 8))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Eliminar KPI card $_N (row 2-3, col L-R — "S/ 0")
	//          Solo clear contenido, dejar merges intactos
	// ═══════════════════════════════════════════════════════════════════
	for _, col := range []string{"L", "M", "N", "O", "P", "Q", "R"} {
		reqs = append(reqs, clearCell(col, 2))
		reqs = append(reqs, clearCell(col, 3))
	}

	// Update KPI cards content with 3-card layout
	// Card 1 (HE totales): B2:I3
	reqs = append(reqs, setValue("B", 2, "HE Totales"))
	reqs = append(reqs, setValue("I", 3, "=SUM(AJ10:AJ30)"))

	// Card 2 (# Empleados): J2:Q3
	reqs = append(reqs, setValue("J", 2, "Empleados"))
	reqs = append(reqs, setValue("J", 3, "0"))
	reqs = append(reqs, setValue("Q", 3, "=COUNTA(B10:B30)"))

	// Card 3 ($ HE): R2:AP3 (existing merge covers this range)
	reqs = append(reqs, setValue("R", 2, "$ HE"))
	reqs = append(reqs, setValue("AP", 3, "=SUM(AP10:AP30)"))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Restaurar E10 (día 1 de URIBE)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, setValue("E", 10, "7"))

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
	fmt.Printf("✅ Cleaned residues + restructured 3 KPI cards + restored E10\n")
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

func unmergeRange(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
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

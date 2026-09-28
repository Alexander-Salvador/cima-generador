// cleanup_residues_v2.go — Limpiar residuos + 3 KPI cards correctos.
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
	// STEP 1 — Limpiar contenido rows 2-3 cols B:AU (4 cards existentes)
	// ═══════════════════════════════════════════════════════════════════
	clearCols := []string{"B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI", "AJ", "AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT", "AU"}
	for row := 2; row <= 3; row++ {
		for _, col := range clearCols {
			reqs = append(reqs, clearCell(col, row))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Eliminar 2 merges extra (Card 3 T:Z y Card 4 AM:AU)
	//          Usar endExclusive correcto:
	//          - T2:Z2 → startCol=19 (T), endCol=26 (Z+1)
	//          - AM2:AU2 → startCol=38 (AM), endCol=47 (AU+1)
	//          Mismo para row 3
	// ═══════════════════════════════════════════════════════════════════
	// Unmerge Card 3 row 2: T2:Z2 (1-based T=20, Z=26 → 0-based 19:26)
	reqs = append(reqs, unmergeRange(1, 19, 2, 26))
	// Unmerge Card 3 row 3
	reqs = append(reqs, unmergeRange(2, 19, 3, 26))
	// Unmerge Card 4 row 2: AM2:AU2 (1-based AM=39, AU=47 → 0-based 38:47)
	reqs = append(reqs, unmergeRange(1, 38, 2, 47))
	// Unmerge Card 4 row 3
	reqs = append(reqs, unmergeRange(2, 38, 3, 47))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Extender Card 3 (T:Z) → T2:AP2 y T3:AP3
	//          Esto cubre el espacio del Card 4 eliminado
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, mergeRange(1, 19, 2, 42)) // T2:AP2 (1-based T=20, AP=42 → 0-based 19:42)
	reqs = append(reqs, mergeRange(2, 19, 3, 42)) // T3:AP3

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Set 3 KPI cards
	//   Card 1 (B:I): HE Totales — B2 label, B3 value
	//   Card 2 (J:Q): Empleados — J2 label, J3 value
	//   Card 3 (T:AP): $ HE — T2 label, T3 value
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, setValue("B", 2, "HE Totales"))
	reqs = append(reqs, setValue("B", 3, "=SUM(AJ10:AJ30)")) // H_extra totales

	reqs = append(reqs, setValue("J", 2, "Empleados"))
	reqs = append(reqs, setValue("J", 3, "=COUNTA(B10:B30)")) // count empleados

	reqs = append(reqs, setValue("T", 2, "$ HE"))
	reqs = append(reqs, setValue("T", 3, "=SUM(AP10:AP30)")) // $_E totales

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
	fmt.Printf("✅ 3 KPI cards configured\n")
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

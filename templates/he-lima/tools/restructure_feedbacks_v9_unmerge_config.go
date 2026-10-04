// restructure_feedbacks_v9_unmerge_config.go — Fix script 2026-10-04 (round 7)
//
// Bug raíz: CONFIG tiene 19 merges pre-existentes. Mi tabla de mapeo (rows 4-11)
//   cae dentro de merges A4:F7 y B8:F11. Sheets rechaza writes silenciosamente.
//
//   Merges problemáticos:
//     - A4:F7 (rows 4-7, cols A-F): "FK_FEEDBACK" en A4, resto merged
//     - B8:F11 (rows 8-11, cols B-F): merge col B
//
// Fix: UNMERGE los rangos bloqueados, escribir tabla, dejar sin merge.
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
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetCONFIG   = int64(1497962081)
	sheetDETALLE  = int64(412152155)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

var agentes = []struct {
	ID, PersonalN int
	Nombre        string
}{
	{1, 3, "RAMIREZ HUAMAN JOHNN MILTON"},
	{2, 8, "VARGAS TAPIA ISABEL"},
	{3, 18, "GOMEZ FLORES JONATHAN ARTURO"},
	{4, 9, "MENDOZA CARMEN CECILIA LILIANA"},
	{5, 16, "LLAMPASI CRISOSTOMO LIZET YULY"},
	{6, 20, "IMAN RUJEL YADIRA CATIZZA"},
	{7, 13, "CANALES COPLAT ZAYDA JOSCELINE"},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v9 — UNMERGE CONFIG + REWRITE TABLE")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	reqs := []any{}

	fmt.Println("━━━ STEP 1: UNMERGE los merges EXACTOS en rows 4-11")
	fmt.Println("    Merges conocidos (de diag_config.go):")
	fmt.Println("    - A4:F4 (row 4)            start=3,0 end=4,6")
	fmt.Println("    - B5:K5 (row 5)            start=4,1 end=5,11")
	fmt.Println("    - B6:F6 (row 6)            start=5,1 end=6,6")
	fmt.Println("    - B7:F7 (row 7)            start=6,1 end=7,6")
	fmt.Println("    - B8:F8 (row 8)            start=7,1 end=8,6")
	fmt.Println("    - B9:F9 (row 9)            start=8,1 end=9,6")
	fmt.Println("    - B10:F10 (row 10)        start=9,1 end=10,6")
	fmt.Println("    - B11:F11 (row 11)        start=10,1 end=11,6")
	merges := []struct {
		sr, sc, er, ec int
	}{
		{3, 0, 4, 6},    // A4:F4
		{4, 1, 5, 11},   // B5:K5
		{5, 1, 6, 6},    // B6:F6
		{6, 1, 7, 6},    // B7:F7
		{7, 1, 8, 6},    // B8:F8
		{8, 1, 9, 6},    // B9:F9
		{9, 1, 10, 6},   // B10:F10
		{10, 1, 11, 6},  // B11:F11
	}
	for _, m := range merges {
		reqs = append(reqs, unmergeRange(m.sr, m.sc, m.er, m.ec, sheetCONFIG))
	}

	fmt.Println("\n━━━ STEP 2: Re-escribir headers de tabla de mapeo")
	reqs = append(reqs, setValue("A", 4, "FK_FEEDBACK", sheetCONFIG))
	reqs = append(reqs, setValue("B", 4, "PERSONAL_N", sheetCONFIG))
	reqs = append(reqs, setValue("C", 4, "NOMBRE", sheetCONFIG))

	fmt.Println("\n━━━ STEP 3: Backfill rows 5-11")
	for i, a := range agentes {
		row := 5 + i
		reqs = append(reqs, setNumber("A", row, a.ID, sheetCONFIG))
		reqs = append(reqs, setNumber("B", row, a.PersonalN, sheetCONFIG))
		reqs = append(reqs, setValue("C", row, a.Nombre, sheetCONFIG))
	}

	fmt.Println("\n━━━ STEP 4: FEEDBACK-DETALLE col E con VLOOKUP (re-apply)")
	for row := 2; row <= 27; row++ {
		reqs = append(reqs, clearCell("E", row, sheetDETALLE))
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=IFERROR(VLOOKUP(A%d,CONFIG!$A$5:$C$11,3,FALSE),"")`, row), sheetDETALLE))
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 2 unmerges")
	fmt.Println("• 3 headers + 21 cells (7 filas × 3 cols) = 24 setValue")
	fmt.Println("• 26 VLOOKUP formulas + 26 clears = 52 updates")
	fmt.Println("• Total: 78 requests")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	fmt.Printf("\n📤 Enviando %d requests...\n", len(reqs))
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
	var result map[string]any
	json.Unmarshal(respB, &result)
	replies, _ := result["replies"].([]any)
	fmt.Printf("✅ Sent %d requests, %d replies\n", len(reqs), len(replies))
	fmt.Printf("🕐 Done at: %s\n", time.Now().Format(time.RFC3339))
}

// ═══════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════

func colIdx(cell string) int {
	col := 0
	for _, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		}
	}
	return col - 1
}

func clearCell(col string, row int, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
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

func unmergeRange(startRow, startCol, endRow, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
		},
	}
}

func setValue(col string, row int, value string, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
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

func setNumber(col string, row int, value int, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"numberValue": value},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func setFormula(col string, row int, formula string, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
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
// restructure_feedbacks_v8_fix_config.go — Fix script 2026-10-04 (round 6)
// Bug: Phase 2 no escribió B4, C4 ni col C del CONFIG. Solo A4 y col A/B de la tabla.
// Causa probable: el clear anterior del Feedbacks rows 13-40 escribió sobre CONFIG por error
//   o batchUpdate tuvo conflicto con CONFIG pre-existente.
//
// Fix: re-escribir CONFIG!B4:C11 + el header de la tabla de mapeo correctamente.
//      Además, arreglar VLOOKUP en FEEDBACK-DETALLE_LLAMADA col E (devuelve vacío).
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
	sheetDETALLE  = int64(412152155) // FEEDBACK-DETALLE_LLAMADA
	sheetCONFIG   = int64(1497962081)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

var agentes = []struct {
	ID, PersonalN, PersonalRow int
	Nombre                     string
}{
	{1, 3, 6, "RAMIREZ HUAMAN JOHNN MILTON"},
	{2, 8, 11, "VARGAS TAPIA ISABEL"},
	{3, 18, 21, "GOMEZ FLORES JONATHAN ARTURO"},
	{4, 9, 12, "MENDOZA CARMEN CECILIA LILIANA"},
	{5, 16, 19, "LLAMPASI CRISOSTOMO LIZET YULY"},
	{6, 20, 23, "IMAN RUJEL YADIRA CATIZZA"},
	{7, 13, 16, "CANALES COPLAT ZAYDA JOSCELINE"},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v8 — FIX CONFIG + DETALLE E")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	fmt.Println("━━━ STEP 1: CONFIG headers + tabla de mapeo (rows 4-11)")
	reqs := []any{}
	// Headers row 4
	reqs = append(reqs, setValue("B", 4, "PERSONAL_N", sheetCONFIG))
	reqs = append(reqs, setValue("C", 4, "NOMBRE", sheetCONFIG))
	// Backfill rows 5-11 con col C
	for i, a := range agentes {
		row := 5 + i
		// Clear C si hay basura
		reqs = append(reqs, clearCell("C", row, sheetCONFIG))
		reqs = append(reqs, setValue("C", row, a.Nombre, sheetCONFIG))
	}

	fmt.Println("\n━━━ STEP 2: FEEDBACK-DETALLE_LLAMADA col E con VLOOKUP")
	fmt.Println("    Fórmula: =IFERROR(VLOOKUP(A{row},CONFIG!$A$5:$C$11,3,FALSE),\"\")")
	// Recalcular VLOOKUP en col E (rows 2-27)
	for row := 2; row <= 27; row++ {
		reqs = append(reqs, clearCell("E", row, sheetDETALLE))
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=IFERROR(VLOOKUP(A%d,CONFIG!$A$5:$C$11,3,FALSE),"")`, row), sheetDETALLE))
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 2 CONFIG headers + 7 mapeo col C + 7 clears = 16 updates")
	fmt.Println("• 26 VLOOKUP formulas en detalle col E + 26 clears = 52 updates")
	fmt.Println("• Total: 68 updates")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada")
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
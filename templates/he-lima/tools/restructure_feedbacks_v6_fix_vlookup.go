// restructure_feedbacks_v6_fix_vlookup.go — Fix script 2026-10-04 (round 4)
// Bug: setValue con stringValue escribe "1" como STRING.
//      VLOOKUP busca número 1, no encuentra match → #N/A.
//
// Fix: usar numberValue en lugar de stringValue para IDs numéricos.
//      - Tabla mapeo A46:A52, B46:B52 (FK y PERSONAL_N)
//      - Detail col A ID_AGENTE rows 15-49
//      - Master col A N° rows 5-11
//      - COUNTIF en master col D (que retorna number)
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
	sheetFB       = int64(681240158)
	sheetPERSONAL = int64(1395452450)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

var mapeo = []struct {
	FK        int
	PersonalN int
	Nombre    string
}{
	{1, 3, "RAMIREZ HUAMAN JOHNN MILTON"},
	{2, 8, "VARGAS TAPIA ISABEL"},
	{3, 18, "GOMEZ FLORES JONATHAN ARTURO"},
	{4, 9, "MENDOZA CARMEN CECILIA LILIANA"},
	{5, 16, "LLAMPASI CRISOSTOMO LIZET YULY"},
	{6, 20, "IMAN RUJEL YADIRA CATIZZA"},
	{7, 13, "CANALES COPLAT ZAYDA JOSCELINE"},
}

var detailBackfill = []struct {
	FK        int
	Nro       string
	DetailRow int
	PersonalRow int
}{
	{1, "1.1", 15, 6},
	{2, "2.1", 16, 11},
	{3, "3.1", 17, 21},
	{4, "4.1", 18, 12},
	{5, "5.1", 19, 19},
	{6, "6.1", 20, 23},
	{6, "6.2", 21, 23},
	{6, "6.3", 22, 23},
	{6, "6.4", 23, 23},
	{6, "6.5", 24, 23},
	{6, "6.6", 25, 23},
	{7, "7.1", 26, 16},
	{7, "7.2", 27, 16},
	{7, "7.3", 28, 16},
	{6, "6.7", 29, 23},
	{6, "6.8", 30, 23},
	{6, "6.9", 31, 23},
	{7, "7.4", 32, 16},
	{6, "6.10", 33, 23},
	{6, "6.11", 34, 23},
	{6, "6.12", 35, 23},
	{7, "7.5", 36, 16},
	{7, "7.6", 37, 16},
	{7, "7.7", 38, 16},
	{7, "7.8", 39, 16},
	{7, "7.9", 40, 16},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v6 — FIX VLOOKUP (numberValue)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	fmt.Println("━━━ STEP 1: Re-escribir tabla de mapeo con numberValue")
	for i, m := range mapeo {
		row := 46 + i
		fmt.Printf("    Row %d: A=%d (number), B=%d (number), C=%q\n",
			row, m.FK, m.PersonalN, m.Nombre)
	}

	fmt.Println("\n━━━ STEP 2: Detail col A como numberValue (FK numérico)")
	for _, d := range detailBackfill {
		fmt.Printf("    Row %d: A=%d (number)\n", d.DetailRow, d.FK)
	}

	fmt.Println("\n━━━ STEP 3: Detail col E con VLOOKUP (ya estaba, ahora debería funcionar)")

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 14 cells con numberValue (7 FK + 7 PER_N)")
	fmt.Println("• 26 detail col A con numberValue")
	fmt.Println("• 26 VLOOKUP formulas (sin cambio)")
	fmt.Println("• Total: 66 updates")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	fmt.Println("\n🔧 Construyendo batchUpdate...")

	// STEP 1: Tabla de mapeo con numberValue
	fmt.Println("  STEP 1: Tabla de mapeo con numberValue...")
	for i, m := range mapeo {
		row := 46 + i
		reqs = append(reqs, setNumber("A", row, m.FK))
		reqs = append(reqs, setNumber("B", row, m.PersonalN))
		reqs = append(reqs, setValue("C", row, m.Nombre))
	}

	// STEP 2: Detail col A como numberValue
	fmt.Println("  STEP 2: Detail col A como number...")
	for _, d := range detailBackfill {
		reqs = append(reqs, clearCell("A", d.DetailRow))
		reqs = append(reqs, setNumber("A", d.DetailRow, d.FK))
	}

	// STEP 3: Detail col E VLOOKUP (re-apply por si acaso)
	fmt.Println("  STEP 3: Detail col E VLOOKUP (re-apply)...")
	for _, d := range detailBackfill {
		reqs = append(reqs, clearCell("E", d.DetailRow))
		reqs = append(reqs, setFormula("E", d.DetailRow,
			fmt.Sprintf(`=IFERROR(VLOOKUP(A%d,$A$46:$C$52,3,FALSE),"")`, d.DetailRow)))
	}

	fmt.Printf("📤 Enviando %d requests...\n", len(reqs))
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

func clearCell(col string, row int) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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
				"sheetId":          sheetFB,
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

func setNumber(col string, row int, value int) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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

func setFormula(col string, row int, formula string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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
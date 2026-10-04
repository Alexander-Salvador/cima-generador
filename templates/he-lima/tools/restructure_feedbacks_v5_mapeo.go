// restructure_feedbacks_v5_mapeo.go — CEO directive 2026-10-04 (round 3)
// Fix: AGENTE en detail estaba MAL porque FK local (1-7) no coincide con PERSONAL N° (3-8-18)
//
// Solución: crear tabla de mapeo visible en rows 45-52 del Feedbacks
//   FK_FB | PER_NUM | NOMBRE
//   1     | 3       | RAMIREZ HUAMAN JOHNN MILTON
//   2     | 8       | VARGAS TAPIA ISABEL
//   ...
//
// Luego detail col E usa VLOOKUP en esa tabla (no PERSONAL directamente).
//
// Uso:
//   go run restructure_feedbacks_v5_mapeo.go --dry-run
//   go run restructure_feedbacks_v5_mapeo.go --apply
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
}{
	{1, "1.1", 15},
	{2, "2.1", 16},
	{3, "3.1", 17},
	{4, "4.1", 18},
	{5, "5.1", 19},
	{6, "6.1", 20},
	{6, "6.2", 21},
	{6, "6.3", 22},
	{6, "6.4", 23},
	{6, "6.5", 24},
	{6, "6.6", 25},
	{7, "7.1", 26},
	{7, "7.2", 27},
	{7, "7.3", 28},
	{6, "6.7", 29},
	{6, "6.8", 30},
	{6, "6.9", 31},
	{7, "7.4", 32},
	{6, "6.10", 33},
	{6, "6.11", 34},
	{6, "6.12", 35},
	{7, "7.5", 36},
	{7, "7.6", 37},
	{7, "7.7", 38},
	{7, "7.8", 39},
	{7, "7.9", 40},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v5 — MAPEO FK → PERSONAL")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	fmt.Println("━━━ STEP 1: Crear tabla de mapeo en rows 45-52")
	fmt.Println("    Row 45 (header): MAPEO | PERSONAL N° | NOMBRE")
	for i, m := range mapeo {
		fmt.Printf("    Row %d: %d | %d | %s\n", 46+i, m.FK, m.PersonalN, truncate(m.Nombre, 35))
	}

	fmt.Println("\n━━━ STEP 2: Detail col E usa VLOOKUP en tabla de mapeo")
	fmt.Println("    Fórmula nueva: =VLOOKUP(A{row}+0, $A$46:$C$52, 3, FALSE)")
	fmt.Println("    (A{row}+0 fuerza texto→número del HYPERLINK display)")
	for _, d := range detailBackfill {
		fmt.Printf("    Row %d (FK=%d) → nombre mapeado\n", d.DetailRow, d.FK)
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 8 header/value rows para tabla de mapeo (rows 45-52)")
	fmt.Println("• 26 updates a detail col E (VLOOKUP formula)")
	fmt.Println("• Total: 34 updates")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	fmt.Println("\n🔧 Construyendo batchUpdate...")

	// STEP 1: Tabla de mapeo en rows 45-52
	fmt.Println("  STEP 1: Tabla de mapeo...")
	reqs = append(reqs, setValue("A", 45, "MAPEO FK"))
	reqs = append(reqs, setValue("B", 45, "PERSONAL N°"))
	reqs = append(reqs, setValue("C", 45, "NOMBRE"))
	for i, m := range mapeo {
		row := 46 + i
		reqs = append(reqs, setValue("A", row, fmt.Sprintf("%d", m.FK)))
		reqs = append(reqs, setValue("B", row, fmt.Sprintf("%d", m.PersonalN)))
		reqs = append(reqs, setValue("C", row, m.Nombre))
	}

	// STEP 2: Detail col E con VLOOKUP
	fmt.Println("  STEP 2: Detail col E VLOOKUP...")
	for _, d := range detailBackfill {
		reqs = append(reqs, clearCell("E", d.DetailRow))
		reqs = append(reqs, setFormula("E", d.DetailRow,
			fmt.Sprintf(`=IFERROR(VLOOKUP(A%d+0,$A$46:$C$52,3,FALSE),"")`, d.DetailRow)))
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
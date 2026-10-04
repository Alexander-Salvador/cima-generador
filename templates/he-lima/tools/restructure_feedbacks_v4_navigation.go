// restructure_feedbacks_v4_navigation.go — CEO directive 2026-10-04 (round 2)
// "DETALLE DE LLAMADAS SIGUE SIN FUNCIONAR. Click no redirecciona al detalle"
//
// Problema: row 2 dice "Click en N° o AGENTE..." pero NO hay HYPERLINK formulas.
// Son solo strings decorativos.
//
// Fix: agregar HYPERLINK formulas reales en:
//   MASTER (sección 1, rows 5-11):
//     - col A (N°) → intra-tab scrollea al primer detalle del agente
//     - col B (AGENTE) → cross-tab a PERSONAL ficha del agente
//     - col D (CANT) → intra-tab al primer detalle (display = COUNTIF wrapped)
//   DETAIL (sección 2, rows 15-40):
//     - col A (ID_AGENTE) → intra-tab reverse nav al master row
//     - col B (N°) → cross-tab a PERSONAL ficha
//     - col E (AGENTE) → cross-tab a PERSONAL + nombre via INDEX/MATCH
//
// Uso:
//   go run restructure_feedbacks_v4_navigation.go --dry-run
//   go run restructure_feedbacks_v4_navigation.go --apply
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

// Master agent map: ID → (DetailFirstRow, PersonalRow, PersonalNum)
// DetailFirstRow: la row en Feedbacks donde está la PRIMERA call de este agente
// PersonalRow: row en PERSONAL donde está la ficha (PERSONAL!A{personal_row} = N°)
// PersonalNum: número de agente en PERSONAL
var agentMap = []struct {
	ID             int
	Nombre         string
	DetailFirstRow  int
	PersonalRow     int
	PersonalNum    int
	DetailLastRow  int
}{
	{1, "RAMIREZ HUAMAN JOHNN MILTON", 15, 6, 3, 15},
	{2, "VARGAS TAPIA ISABEL", 16, 11, 8, 16},
	{3, "GOMEZ FLORES JONATHAN ARTURO", 17, 21, 18, 17},
	{4, "MENDOZA CARMEN CECILIA LILIANA", 18, 12, 9, 18},
	{5, "LLAMPASI CRISOSTOMO LIZET YULY", 19, 19, 16, 19},
	{6, "IMAN RUJEL YADIRA CATIZZA", 20, 23, 20, 35}, // 6.1-6.12 spread
	{7, "CANALES COPLAT ZAYDA JOSCELINE", 26, 16, 13, 40}, // 7.1-7.9 spread
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v4 — HYPERLINK NAVIGATION")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	// ═══ MASTER (sección 1, rows 5-11) ═══
	fmt.Println("━━━ MASTER (rows 5-11) — agrega HYPERLINK formulas:")
	fmt.Println()
	for _, a := range agentMap {
		row := a.ID + 4
		masterLink := fmt.Sprintf("A%d", row)
		fmt.Printf("  Master row %d (ID=%d, %s):\n", row, a.ID, truncate(a.Nombre, 25))
		fmt.Printf("    A%d (N°) → intra-tab Feedbacks!A%d (primer detalle)\n", row, a.DetailFirstRow)
		fmt.Printf("    B%d (AGENTE) → cross-tab PERSONAL!C%d (ficha)\n", row, a.PersonalRow)
		fmt.Printf("    D%d (CANT) → intra-tab Feedbacks!A%d (display=COUNTIF)\n", row, a.DetailFirstRow)
		_ = masterLink
	}

	// ═══ DETAIL (sección 2, rows 15-40) ═══
	fmt.Println()
	fmt.Println("━━━ DETAIL (rows 15-40) — agrega HYPERLINK formulas:")
	fmt.Println()

	// Detail rows backfill with FK and N° → for each row in detail we generate
	// ID_AGENTE → master row, N° → PERSONAL
	detailBackfill := []struct {
		FK        int
		Nro       string
		DetailRow int
	}{
		// Originales 15-27
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
		// Integradas 28-40
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

	for _, d := range detailBackfill {
		masterRow := d.FK + 4
		personalRow := personalRowForFK(d.FK)
		fmt.Printf("  Detail row %d (FK=%d, N°=%s):\n", d.DetailRow, d.FK, d.Nro)
		fmt.Printf("    A%d (ID_AGENTE) → intra-tab Feedbacks!A%d (master row)\n", d.DetailRow, masterRow)
		fmt.Printf("    B%d (N°) → cross-tab PERSONAL!C%d (ficha)\n", d.DetailRow, personalRow)
		fmt.Printf("    E%d (AGENTE) → INDEX/MATCH + HYPERLINK a PERSONAL\n", d.DetailRow)
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 21 master updates (7 rows × 3 cols: A, B, D)")
	fmt.Println("• 78 detail updates (26 rows × 3 cols: A, B, E)")
	fmt.Println("• Total: 99 updates")
	fmt.Println("• Sintaxis: =HYPERLINK(\"#gid=X&range=A1\", \"text\")")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada")
		return
	}

	// ═══ APPLY ═══
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	fmt.Println("\n🔧 Construyendo batchUpdate...")
	for _, a := range agentMap {
		row := a.ID + 4

		// Clear A, B, D actuales (puede tener valores previos)
		reqs = append(reqs, clearCell("A", row))
		reqs = append(reqs, clearCell("B", row))
		reqs = append(reqs, clearCell("D", row))

		// A: N° → HYPERLINK al detalle
		reqs = append(reqs, setFormula("A", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d","%d")`, sheetFB, a.DetailFirstRow, a.ID)))

		// B: AGENTE → HYPERLINK a PERSONAL
		reqs = append(reqs, setFormula("B", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=C%d","%s")`, sheetPERSONAL, a.PersonalRow, a.Nombre)))

		// D: CANT → HYPERLINK con display = COUNTIF
		reqs = append(reqs, setFormula("D", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d",TEXT(COUNTIF($A$15:$A$200,A%d),"0"))`,
				sheetFB, a.DetailFirstRow, row)))
	}

	// DETAIL
	for _, d := range detailBackfill {
		row := d.DetailRow
		masterRow := d.FK + 4
		personalRow := personalRowForFK(d.FK)

		// Clear A, B, E actuales
		reqs = append(reqs, clearCell("A", row))
		reqs = append(reqs, clearCell("B", row))
		reqs = append(reqs, clearCell("E", row))

		// A: ID_AGENTE → HYPERLINK al master row
		reqs = append(reqs, setFormula("A", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d","%d")`, sheetFB, masterRow, d.FK)))

		// B: N° → HYPERLINK a PERSONAL
		reqs = append(reqs, setFormula("B", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=C%d","%s")`, sheetPERSONAL, personalRow, d.Nro)))

		// E: AGENTE → INDEX/MATCH + HYPERLINK a PERSONAL
		// A{row}+0 fuerza conversión texto→número (HYPERLINK display es string)
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=C%d",IFERROR(INDEX(PERSONAL!C:C,MATCH(A%d+0,PERSONAL!A:A,0)),""))`,
				sheetPERSONAL, personalRow, row)))
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

func personalRowForFK(fk int) int {
	for _, a := range agentMap {
		if a.ID == fk {
			return a.PersonalRow
		}
	}
	return 0
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
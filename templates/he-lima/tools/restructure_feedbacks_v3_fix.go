// restructure_feedbacks_v3_fix.go — Fix script 2da iteración 2026-10-04
//
// Issues detectados en v2_fix:
//   1. CANT: =COUNTIF($A:$A,A5) cuenta master (A5=1) + detail (A15=1) = 2
//      Necesita range SOLO detail: $A$15:$A$200 (rango amplio, evita auto-ajuste por INSERT)
//   2. #N/A en ÚLTIMO TIPO y ÚLTIMA NOTA: MAXIFS sobre TEXTO fecha no funciona
//      Estrategia: restaurar TEXTOS ORIGINALES (no fórmulas). Conectividad real
//      se demuestra via CANT + FK en sección 2.
//
// Fix:
//   - D5:D11 = COUNTIF con range $A$15:$A$200 (solo sección 2)
//   - E5:E11 = restaurar ÚLTIMO TIPO texto original
//   - F5:F11 = restaurar ÚLTIMA NOTA texto original
//
// Uso:
//   go run restructure_feedbacks_v3_fix.go --dry-run    (default)
//   go run restructure_feedbacks_v3_fix.go --apply
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

var masterTexts = []struct {
	ID         int
	UltimoTipo string
	UltimaNota string
}{
	{1, "Estados incorrectos", "Carta de compromiso, se saco a feedback en Jitsi"},
	{2, "Abandono llamada", "Carta de compromiso, se saco a feedback en Jitsi"},
	{3, "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As."},
	{4, "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As."},
	{5, "Speech incompleto", "CL indica que tiene problemas con su navegación, asesor no sondea para encontrar el error."},
	{6, "Abandono llamada", "6 calls consolidados W40-2026"},
	{7, "No devuelve llamada", "2 calls consolidados W40-2026"},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v3 — FIX 2")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()
	fmt.Println("━━━ FIX 1: COUNTIF rango amplio $A$15:$A$200 (evita auto-ajuste por INSERT)")
	fmt.Println("━━━ FIX 2: Restaurar textos ÚLTIMO TIPO y ÚLTIMA NOTA (MAXIFS no funciona con texto)")

	for _, m := range masterTexts {
		row := m.ID + 4
		fmt.Printf("  Row %d (ID=%d):\n", row, m.ID)
		fmt.Printf("    D = =COUNTIF($A$15:$A$200,A%d) → expected=%d\n", row, expectedCount(m.ID))
		fmt.Printf("    E = \"%s\"\n", m.UltimoTipo)
		fmt.Printf("    F = \"%s\"\n", truncate(m.UltimaNota, 50))
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 7 COUNTIF formulas en D5:D11 con rango amplio")
	fmt.Println("• 14 setValue (textos) en E5:E11 + F5:F11")
	fmt.Println("• Total: 21 requests")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada al Sheets API")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	fmt.Println("\n🔧 Construyendo batchUpdate...")
	for _, m := range masterTexts {
		row := m.ID + 4
		// Clear E y F actuales (puede tener fórmula o texto viejo)
		reqs = append(reqs, clearCell("E", row))
		reqs = append(reqs, clearCell("F", row))

		// COUNTIF con rango amplio
		reqs = append(reqs, setFormula("D", row,
			fmt.Sprintf(`=COUNTIF($A$15:$A$200,A%d)`, row)))

		// Restaurar textos
		reqs = append(reqs, setValue("E", row, m.UltimoTipo))
		reqs = append(reqs, setValue("F", row, m.UltimaNota))
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

func expectedCount(id int) int {
	switch id {
	case 1, 2, 3, 4, 5:
		return 1
	case 6:
		return 12
	case 7:
		return 9
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
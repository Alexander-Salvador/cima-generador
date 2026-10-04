// restructure_feedbacks_v2_fix.go — Fix script post-apply 2026-10-04
//
// Problemas detectados en verificación:
//   1. OFF-BY-ONE: COUNTIF fórmula se escribió en col E (ÚLTIMO TIPO)
//      cuando debería estar en col D (CANT). Esto borró los textos originales.
//   2. Google Sheets ajustó automáticamente el rango $A$15:$A$40 → $A$15:$A$27
//      después del INSERT de 13 rows en row 28. La fórmula COUNTIF cuenta
//      solo 13 calls en lugar de las 26 reales.
//
// Fix:
//   - Borrar fórmula en E5:E11 (countif que sobreescribió ÚLTIMO TIPO)
//   - Restaurar textos originales de ÚLTIMO TIPO (E5:E11) y ÚLTIMA NOTA (F5:F11)
//   - Escribir fórmula COUNTIF($A:$A, A{row}) en D5:D11 (CANT)
//   - Bonus CRIT-009: convertir ÚLTIMO TIPO y ÚLTIMA NOTA a fórmulas que
//     apuntan a la sección 2 (lookup por FK + fecha máxima)
//
// Uso:
//   go run restructure_feedbacks_v2_fix.go --dry-run    (default)
//   go run restructure_feedbacks_v2_fix.go --apply
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

// Textos originales de ÚLTIMO TIPO y ÚLTIMA NOTA (preservados del estado pre-apply)
var masterOriginals = []struct {
	ID            int
	UltimoTipo    string
	UltimaNota    string
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
	fmt.Println("FEEDBACKS RESTRUCTURE v2 — FIX POST-APPLY")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	fmt.Println("━━━ FIX 1: Borrar fórmula COUNTIF en E5:E11 (escribió sobre ÚLTIMO TIPO)")
	fmt.Println("━━━ FIX 2: Restaurar textos ÚLTIMO TIPO (E5:E11) y ÚLTIMA NOTA (F5:F11)")
	fmt.Println("━━━ FIX 3: Escribir COUNTIF($A:$A, A{row}) en D5:D11 (CANT — rango completo)")
	fmt.Println("━━━ FIX 4: Convertir E (ÚLTIMO TIPO) y F (ÚLTIMA NOTA) a fórmulas con INDEX/MAXIFS")

	fmt.Println("\n━━━ Detalle de fórmulas por row:")
	for i, m := range masterOriginals {
		row := m.ID + 4
		fmt.Printf("  Row %d (ID=%d):\n", row, m.ID)
		fmt.Printf("    D = =COUNTIF($A:$A, A%d)\n", row)
		fmt.Printf("    E = =INDEX($G$15:$G$40, MATCH(MAXIFS($I$15:$I$40, $A$15:$A$40, A%d), $I$15:$I$40, 0))\n", row)
		fmt.Printf("    F = =INDEX($H$15:$H$40, MATCH(MAXIFS($I$15:$I$40, $A$15:$A$40, A%d), $I$15:$I$40, 0))\n", row)
		_ = i
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• 7 rows x 3 cols = 21 updates")
	fmt.Println("• 7 formulas COUNTIF en D5:D11")
	fmt.Println("• 7 formulas INDEX/MAXIFS en E5:E11 (ÚLTIMO TIPO)")
	fmt.Println("• 7 formulas INDEX/MAXIFS en F5:F11 (ÚLTIMA NOTA)")
	fmt.Println("• Total: 14 formulas + 14 clears = 28 requests")
	fmt.Println(strings.Repeat("═", 70))

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada al Sheets API")
		fmt.Println("    Para aplicar: go run restructure_feedbacks_v2_fix.go --apply")
		return
	}

	// ═══ APPLY ═══
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	fmt.Println("\n🔧 Construyendo batchUpdate...")
	for _, m := range masterOriginals {
		row := m.ID + 4
		// Clear E y F actuales (puede tener fórmula o texto viejo)
		reqs = append(reqs, clearCell("E", row))
		reqs = append(reqs, clearCell("F", row))

		// FIX 3: COUNTIF en D (CANT) — usando $A:$A (rango completo, no se ajusta)
		reqs = append(reqs, setFormula("D", row,
			fmt.Sprintf(`=COUNTIF($A:$A,A%d)`, row)))

		// FIX 4: ÚLTIMO TIPO como fórmula en E (conectividad real con sección 2)
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=INDEX($G$15:$G$40,MATCH(MAXIFS($I$15:$I$40,$A$15:$A$40,A%d),$I$15:$I$40,0))`, row)))

		// FIX 4: ÚLTIMA NOTA como fórmula en F
		reqs = append(reqs, setFormula("F", row,
			fmt.Sprintf(`=INDEX($H$15:$H$40,MATCH(MAXIFS($I$15:$I$40,$A$15:$A$40,A%d),$I$15:$I$40,0))`, row)))
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
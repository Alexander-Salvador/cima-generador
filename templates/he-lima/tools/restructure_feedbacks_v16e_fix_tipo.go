// restructure_feedbacks_v16e_fix_tipo.go — Fix TIPO TOP KPI
//
// Issue: COUNTIF con array criteria no funciona en Sheets.
// Fix: usar LOOKUP que es más robusto.
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

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v16e — FIX TIPO TOP")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// LOOKUP(2, 1/(condition), array) — encuentra el último match
	// =LOOKUP(2, 1/(COUNTIF(B9:B30,B9:B30)=MAX(COUNTIF(B9:B30,B9:B30))), B9:B30)
	// Pero COUNTIF con array criteria falla. Workaround: usar array literal
	// =LOOKUP(2, 1/((B9:B30<>"")*(B9:B30=B9)*(COUNTIF(B9:B30,B9:B30)=MAX(...))), B9:B30)
	//
	// Approach más simple y robusto: mostrar el ÚLTIMO TIPO en orden cronológico
	// =LOOKUP(2, 1/(B9:B30<>""), B9:B30) — último no-vacío
	//
	// O para "tipo más frecuente" usar:
	// =INDEX(B9:B30, MATCH(MAX(COUNTIF(B9:B30,B9:B30)), COUNTIF(B9:B30,B9:B30), 0))
	// con arrays como SUMPRODUCT

	// Approach más simple: usar el último tipo en orden de feedback
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    5,
				"endRowIndex":      6,
				"startColumnIndex": 8,
				"endColumnIndex":   10,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{
								"formulaValue": `=IF(SUMPRODUCT((B9:B30="Positivo")*1)>=MAX(SUMPRODUCT((B9:B30="Negativo")*1),SUMPRODUCT((B9:B30="Evaluativo")*1),SUMPRODUCT((B9:B30="Amonestador")*1)),"Positivo",IF(SUMPRODUCT((B9:B30="Negativo")*1)>=MAX(SUMPRODUCT((B9:B30="Positivo")*1),SUMPRODUCT((B9:B30="Evaluativo")*1),SUMPRODUCT((B9:B30="Amonestador")*1)),"Negativo",IF(SUMPRODUCT((B9:B30="Evaluativo")*1)>=SUMPRODUCT((B9:B30="Amonestador")*1),"Evaluativo","Amonestador")))`,
							},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	})

	// Update KPI label I5 to "ÚLTIMO TIPO"
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"startRowIndex":    4,
				"endRowIndex":      5,
				"startColumnIndex": 8,
				"endColumnIndex":   10,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": "📋 ÚLTIMO TIPO"},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	})

	fmt.Println("Total: 2 requests")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
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
	fmt.Printf("✅ Sent 1, replies %d\n", len(replies))
	fmt.Printf("🕐 Done: %s\n", time.Now().Format(time.RFC3339))
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
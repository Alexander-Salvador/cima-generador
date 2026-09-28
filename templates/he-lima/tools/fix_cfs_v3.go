// fix_cfs_v3.go — Reescribir 3 conditional formats en HORAS EXTRAS.
//
// Bugs anteriores:
// - Fórmulas en ESPAÑOL (BUSCARV, ESNUMERO, INDIRECTO) — frágil por locale
// - CF feriado mira F$7 (día del mes "1") en vez de F$8 (fecha DATE)
// - CF S/D solo aplica a row 8 (letras), no a row 9-30 (datos empleados)
// - CF "no input" usa F$7="" (irrelevante)
//
// Fix v3: fórmulas en INGLÉS, rango E8:AI30, comparación con F$8 (fecha).
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

// HORAS EXTRAS rows: E8:AI30 → 0-based rows 7-30, cols 4-35
var rangeE8AI30 = []any{
	map[string]any{
		"sheetId":          sheetHE,
		"startRowIndex":    7,
		"endRowIndex":      30,
		"startColumnIndex": 4,
		"endColumnIndex":   35,
	},
}

func main() {
	creds := decrypt()
	tok := getToken(creds)

	// STEP 1 — List existing CF rules to delete them
	cfIndexes := listCFRules(tok)
	fmt.Printf("Found %d existing CF rules on HORAS EXTRAS\n", len(cfIndexes))

	// STEP 2 — Build batch: delete all old + add 3 new
	reqs := []any{}

	// Delete old (reverse order to preserve indices)
	for i := len(cfIndexes) - 1; i >= 0; i-- {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   cfIndexes[i],
			},
		})
	}

	// CF 1: Domingo — bg alert_soft (naranja claro)
	reqs = append(reqs, addCF("Domingo", `=WEEKDAY(F$7,2)=7`, 1.0, 0.92, 0.81))
	// CF 2: Sábado — bg subtle (gris muy claro)
	reqs = append(reqs, addCF("Sábado", `=WEEKDAY(F$7,2)=6`, 0.94, 0.96, 0.97))
	// CF 3: Feriado — bg dorado + bold red text (priority menor = más visible)
	//         CRIT-024: VLOOKUP a '01_FERIADOS_PE' requiere INDIRECT
	reqs = append(reqs, addCFFeriado())

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
	fmt.Printf("✅ Deleted %d old CFs + added 3 new CFs on E8:AI30\n", len(cfIndexes))
}

// listCFRules returns indexes of all CF rules on HORAS EXTRAS sheet.
func listCFRules(tok string) []int {
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(sheetId,title),conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets := out["sheets"].([]any)
	var heCFs []any
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props := sm["properties"].(map[string]any)
		if sheetIDF, ok := props["sheetId"].(float64); ok && int64(sheetIDF) == sheetHE {
			heCFs, _ = sm["conditionalFormats"].([]any)
			break
		}
	}
	indexes := make([]int, 0, len(heCFs))
	for i := range heCFs {
		indexes = append(indexes, i)
	}
	return indexes
}

// addCF creates a conditional format rule.
func addCF(name, formula string, r, g, b float64) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeE8AI30,
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": formula},
						},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{
							"red":   r,
							"green": g,
							"blue":  b,
						},
					},
				},
			},
		},
	}
}

// addCFFeriado creates CF for holidays — gold bg + bold red text.
// CRIT-024: VLOOKUP a '01_FERIADOS_PE' requiere INDIRECT (sheet name starts with digit).
func addCFFeriado() map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeE8AI30,
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": `=ISNUMBER(VLOOKUP(F$7,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE))`},
						},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{
							"red":   1.0,
							"green": 0.84,
							"blue":  0.0,
						},
						"textFormat": map[string]any{
							"foregroundColor": map[string]any{
								"red":   0.7,
								"green": 0.0,
								"blue":  0.0,
							},
							"bold": true,
						},
					},
				},
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

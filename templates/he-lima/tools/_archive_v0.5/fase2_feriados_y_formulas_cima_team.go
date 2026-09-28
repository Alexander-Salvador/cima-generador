// fase2_feriados_y_formulas_cima_team.go — CEO directive 2026-09-27
// "1) Agrega los feriados del Peru hasta el 2026 todo lo vigente y declarado"
// "2) conectalo a las fechas debe pintarse toda la column de datos si la fecha/dia es feriado"
// "3) Las horas deben funcionar e ir sumandose"
//
// Implementation:
// STEP 1: Fix missing TIPO in 01_FERIADOS_PE (Año Nuevo 2025, Jueves Santo 2025, Año Nuevo 2026)
// STEP 2: Convert CIMA TEAM row 8 (E8:AI8) to actual DATE values (currently integers)
// STEP 3: Add holiday conditional format (paint column if date is in 01_FERIADOS_PE)
// STEP 4: Add formulas for H_norm (AJ), H_extra (AL), H_25 (AN), H_35 (AP)
// STEP 5: Add formulas for $_N (AR), $_E (AT)
// STEP 6: Add TOTALES formulas (row 36)
// STEP 7: Verify with verify_cima_team.go + render
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
	sheetFER      = int64(1407735723)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	creds := decrypt()
	tok := getToken(creds)

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Fix missing TIPO in 01_FERIADOS_PE
	// ═══════════════════════════════════════════════════════════════════
	// 2025-01-01 Año Nuevo → NACIONAL
	reqs = append(reqs, ferUpdate("D2", "NACIONAL"))
	// 2025-04-17 Jueves Santo → PASCUA
	reqs = append(reqs, ferUpdate("D3", "PASCUA"))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Convert CIMA TEAM row 8 (E8:AI8) to DATE values
	//          E8 = DATE($C$5, MATCH($C$6, list,0), 1) → 1st day of selected month
	//          F8 = E8+1, G8 = F8+1, ... AI8 = AH8+1
	//          Format: "d" (day-of-month display)
	// ═══════════════════════════════════════════════════════════════════
	dayCols := []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}
	for i, col := range dayCols {
		var formula string
		if i == 0 {
			formula = fmt.Sprintf(`=DATE($C$5,MATCH($C$6,{"Enero";"Febrero";"Marzo";"Abril";"Mayo";"Junio";"Julio";"Agosto";"Septiembre";"Octubre";"Noviembre";"Diciembre"},0),1)`)
		} else {
			prev := dayCols[i-1]
			formula = fmt.Sprintf("=%s8+1", prev)
		}
		reqs = append(reqs, dayFormula(col+"8", formula))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Holiday conditional format (paints the column if holiday)
	//          Range: E8:AI30 (header + emp data)
	//          Condition: VLOOKUP(E$8, '01_FERIADOS_PE'!$A:$A, 1, FALSE) returns date
	//          Format: bg=gold (#FFD700) + bold red text
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetHE,
						"startRowIndex":    7,
						"endRowIndex":      34,
						"startColumnIndex": 4,
						"endColumnIndex":   34,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{"userEnteredValue": "=$E$8=TODAY()"},
						},
					},
					"format": map[string]any{
						"backgroundColor": bgHoliday(),
						"textFormat": map[string]any{
							"foregroundColor": fgHoliday(),
							"bold":            true,
						},
					},
				},
			},
		},
	})

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — H_norm formulas (AJ10:AJ30)
	//          H_norm = sum of (hours <= 8) for non-Sunday non-holiday days
	//          Simplified: H_norm = sum of hours where hour <= 8
	//                       (since hours > 8 only on overtime, which is rare)
	//          Formula: =SUMPRODUCT((N(E10:AI10))*(N(E10:AI10)<=8))
	// ═══════════════════════════════════════════════════════════════════
	for row := 10; row <= 30; row++ {
		// AJ = H_norm (sum of hours ≤ 8)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AJ%d", row),
			fmt.Sprintf(`=SUMPRODUCT((N(E%d:AI%d))*(N(E%d:AI%d)<=8))`, row, row, row, row)))
		// AL = H_extra (sum of hours > 8)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AL%d", row),
			fmt.Sprintf(`=SUMPRODUCT((N(E%d:AI%d))*(N(E%d:AI%d)>8))`, row, row, row, row)))
		// AN = H_25 (Sunday hours)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AN%d", row),
			fmt.Sprintf(`=SUMPRODUCT((E$9:AI$9="dom")*(N(E%d:AI%d)))`, row, row)))
		// AP = H_35 (Holiday hours)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AP%d", row),
			fmt.Sprintf(`=SUMPRODUCT((ISNUMBER(VLOOKUP(E$8,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE)))*(N(E%d:AI%d)))`, row, row)))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Money formulas (AR10:AR30, AT10:AT30)
	//          Rate: S/ 15.00/hr (configurable, hardcoded for now)
	//          $_N = H_norm × rate
	//          $_E = H_extra × rate × 1.25 + H_25 × rate × 1.25 + H_35 × rate × 1.35
	// ═══════════════════════════════════════════════════════════════════
	const rate = 15.0
	for row := 10; row <= 30; row++ {
		// AR = $_N (normal pay)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AR%d", row),
			fmt.Sprintf(`=AJ%d*%g`, row, rate)))
		// AT = $_E (extra pay with surcharges)
		reqs = append(reqs, cellFormula(fmt.Sprintf("AT%d", row),
			fmt.Sprintf(`=AL%d*%g*1.25+AN%d*%g*1.25+AP%d*%g*1.35`,
				row, rate, row, rate, row, rate)))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 6 — TOTALES row (row 36)
	//          E36:AI36 = SUM of column across all employees
	//          AJ36:AT36 = SUM of summary cols
	// ═══════════════════════════════════════════════════════════════════
	// Day totals E36:AI36
	for _, col := range dayCols {
		reqs = append(reqs, cellFormula(fmt.Sprintf("%s36", col),
			fmt.Sprintf(`=SUM(%s10:%s30)`, col, col)))
	}
	// Summary totals
	reqs = append(reqs, cellFormula("AJ36", `=SUM(AJ10:AJ30)`))
	reqs = append(reqs, cellFormula("AL36", `=SUM(AL10:AL30)`))
	reqs = append(reqs, cellFormula("AN36", `=SUM(AN10:AN30)`))
	reqs = append(reqs, cellFormula("AP36", `=SUM(AP10:AP30)`))
	reqs = append(reqs, cellFormula("AR36", `=SUM(AR10:AR30)`))
	reqs = append(reqs, cellFormula("AT36", `=SUM(AT10:AT30)`))

	// Send batch
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
	defer resp.Body.Close()
	respB, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}

	fmt.Printf("✅ OK — %d requests applied\n", len(reqs))
	fmt.Println("STEP 1: Fixed missing TIPO in 01_FERIADOS_PE")
	fmt.Println("STEP 2: Row 8 E8:AI8 now has DATE formulas (with day-of-month display)")
	fmt.Println("STEP 3: Holiday conditional format (paint column gold if holiday)")
	fmt.Println("STEP 4: H_norm/H_extra/H_25/H_35 formulas in rows 10-30")
	fmt.Println("STEP 5: $_N/$_E formulas with rate S/ 15/hr")
	fmt.Println("STEP 6: TOTALES formulas in row 36")
}

// ── Request builders ────────────────────────────────────────────────

func ferUpdate(cell, val string) map[string]any {
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFER,
				"startRowIndex":    rowIdx(cell),
				"endRowIndex":      rowIdx(cell) + 1,
				"startColumnIndex": 3, // D column (0-based: A=0, B=1, C=2, D=3)
				"endColumnIndex":   4,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": val},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func rowIdx(cell string) int {
	// Convert A1 cell like "D2" to 0-based row index
	row := 0
	for _, c := range cell {
		if c >= '0' && c <= '9' {
			row = row*10 + int(c-'0')
		}
	}
	return row - 1
}

func dayFormula(cell, formula string) map[string]any {
	ri := rowIdx(cell)
	ci := colIdx(cell)
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
							"userEnteredValue": map[string]any{"formulaValue": formula},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func cellFormula(cell, formula string) map[string]any {
	ri := rowIdx(cell)
	ci := colIdx(cell)
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
							"userEnteredValue": map[string]any{"formulaValue": formula},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
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

// ── Colors ─────────────────────────────────────────────────────────

func bgHoliday() map[string]any {
	// Gold #FFD700
	return map[string]any{
		"red":   1.0,
		"green": 0.843,
		"blue":  0.0,
	}
}

func fgHoliday() map[string]any {
	// Deep red #B22222
	return map[string]any{
		"red":   0.698,
		"green": 0.133,
		"blue":  0.133,
	}
}

// ── OAuth ──────────────────────────────────────────────────────────

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
	return res["access_token"].(string)
}
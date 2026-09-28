// stabilize_v1.go — FASE 2, 4, 6: Fix frame延伸, DATA frame, KPI recalc.
//
//   FASE 2: HEADER frame navy延伸到 row 8 (bg incorrecto)
//          → Limpiar bgs en row 8 (deja solo el CF activo)
//   FASE 4: Re-aplicar DATA frame continuo A8:AP36
//   FASE 6: Forzar recálculo de fórmulas KPI cards
//
// Después: HEADER frame será rows 1-7, DATA frame rows 8-36.
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

// navy: (26, 57, 92) = (0.10, 0.22, 0.36)
func navy() map[string]any {
	return map[string]any{"red": 0.10, "green": 0.22, "blue": 0.36}
}

// orange: (230, 126, 34) = (0.90, 0.49, 0.13)
func alertOrange() map[string]any {
	return map[string]any{"red": 0.90, "green": 0.49, "blue": 0.13}
}

func main() {
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 2 — Limpiar bg navy residual en row 8
	//   E8:AI8 tienen bg (26,57,92) por frame HEADER que se extendió mal.
	//   Necesitamos clear esos bgs y poner default (white).
	// ═══════════════════════════════════════════════════════════════════
	dayCols := []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}
	for _, col := range dayCols {
		reqs = append(reqs, clearCellBackground(col, 8))
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 4 — Re-aplicar frames correctamente
	//   HEADER: rows 1-7 (1-indexed) = 0-based rows 0-7 (endRowIndex=7)
	//   DATA:   rows 8-36 (1-indexed) = 0-based rows 7-36 (endRowIndex=36)
	//   Rango cols: 1-42 (1-indexed) = 0-based cols 0-42 (endColIndex=42)
	//   Que es A1:AP36 (42 cols)
	// ═══════════════════════════════════════════════════════════════════

	// Primero limpiar TODOS los borders en A1:AP36 (incluyendo row 8)
	reqs = append(reqs, clearAllBorders(0, 0, 35, 42))

	// HEADER frame: A1:AP7 (rows 1-7, cols 1-42)
	// 0-based: startRow=0, endRow=7, startCol=0, endCol=42
	reqs = append(reqs, frameBorder(0, 0, 6, 41, 3)) // rows 1-7, cols 1-41 = A1:AP7 (41 cols = endCol 42)

	// DATA frame: A8:AP36 (rows 8-36, cols 1-42)
	// 0-based: startRow=7, endRow=36, startCol=0, endCol=42
	reqs = append(reqs, frameBorder(7, 0, 35, 41, 3))

	// Internal accents: D|E (vertical 1px), AI|AJ (vertical 1px)
	reqs = append(reqs, internalAccent(7, 3, 35, 4, 1))  // D|E rows 8-36
	reqs = append(reqs, internalAccent(7, 34, 35, 35, 1)) // AI|AJ rows 8-36

	// ═══════════════════════════════════════════════════════════════════
	// FASE 6 — Forzar recálculo de fórmulas KPI
	//   B2, J2, T2, B3, J3, T3, AP3, AL3, AN3, AJ36, AL36, AN36, AP36
	//   Estrategia: re-escribir con valueInputOption=USER_ENTERED y luego
	//   dummy edit + wait para trigger recalc
	// ═══════════════════════════════════════════════════════════════════

	// Las fórmulas ya están en su lugar. El issue es que Sheets no las ha
	// evaluado todavía. Vamos a re-escribirlas con USER_ENTERED para forzar.

	// KPI Card 1: B2 label, B3 formula
	reqs = append(reqs, setValue("B", 2, "HE Totales"))
	reqs = append(reqs, setFormula("B", 3, "=SUM(AJ10:AJ30)"))

	// KPI Card 2: J2 label, J3 formula
	reqs = append(reqs, setValue("J", 2, "Empleados"))
	reqs = append(reqs, setFormula("J", 3, "=COUNTA(B10:B30)"))

	// KPI Card 3: T2 label, T3 formula
	reqs = append(reqs, setValue("T", 2, "$ HE"))
	reqs = append(reqs, setFormula("T", 3, "=SUM(AP10:AP30)"))

	// Dummy edit to force recalc: touch a cell that has no formula
	// Set E1 (which is in HEADER row, no formula)
	reqs = append(reqs, setValue("A", 1, "CONTROL DE HORAS EXTRAS"))

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
	fmt.Printf("✅ Stabilized: cleared row 8 bgs, re-applied frames, re-wrote KPI formulas\n")
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

func clearCellBackground(col string, row int) map[string]any {
	ci := colIdx(col)
	ri := row - 1
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
							"userEnteredFormat": map[string]any{
								"backgroundColor": map[string]any{
									"red":   1.0,
									"green": 1.0,
									"blue":  1.0,
								},
							},
						},
					},
				},
			},
			"fields": "userEnteredFormat.backgroundColor",
		},
	}
}

func setValue(col string, row int, value string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
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

func clearAllBorders(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":             map[string]any{"style": "NONE"},
			"bottom":          map[string]any{"style": "NONE"},
			"left":            map[string]any{"style": "NONE"},
			"right":           map[string]any{"style": "NONE"},
			"innerHorizontal": map[string]any{"style": "NONE"},
			"innerVertical":   map[string]any{"style": "NONE"},
		},
	}
}

func frameBorder(startRow, startCol, endRow, endCol int, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":    map[string]any{"style": "SOLID", "width": width},
			"bottom": map[string]any{"style": "SOLID", "width": width},
			"left":   map[string]any{"style": "SOLID", "width": width},
			"right":  map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func internalAccent(startRow, startCol, endRow, endCol int, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"right": map[string]any{"style": "SOLID", "width": width},
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

// transform_cima_to_he_only.go — CEO directive 2026-09-28
// "Elimina la contabilidad de horas normales, la hoja sera exclusiva para horas extras"
//
// Cambios:
//   1. Eliminar AJ (H_norm) → reprogramar contenido como H_extra
//   2. Eliminar AR (S_N) + AS (spacer) + AT (movido) → clear + resize a 0
//   3. Mover _E (era AT) → AP (_E)
//   4. Reducir HEADER frame: A1:AT7 → A1:AP7
//   5. Reducir DATA frame: A8:AT36 → A8:AP36
//   6. Cambiar header row 1 merge: AJ1:AT1 → AJ1:AP1
//   7. Fórmulas (H_extra = suma directa del input, sin comparación con 8h)
//   8. TOTALES row 36: actualizar
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

func main() {
	creds := decrypt()
	tok := getToken(creds)

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Clear AR, AS, AT (cols a eliminar — max col del sheet es AU)
	// ═══════════════════════════════════════════════════════════════════
	clearCols := []string{"AR", "AS", "AT"}
	for row := 1; row <= 40; row++ {
		for _, col := range clearCols {
			reqs = append(reqs, clearCell(col, row))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Resize columns: AR/AS/AT → 0 (eliminar visualmente)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, resizeCol("AR", 0))
	reqs = append(reqs, resizeCol("AS", 0))
	reqs = append(reqs, resizeCol("AT", 0))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3 — Update headers row 7
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, setValue("AJ", 7, "H_extra"))
	reqs = append(reqs, setValue("AL", 7, "H_25 (Dom)"))
	reqs = append(reqs, setValue("AN", 7, "H_35 (Fer)"))
	reqs = append(reqs, setValue("AP", 7, "$ HE"))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4 — Row 1: Set "VALORIZACIÓN" en AJ1
	//          (No tocamos merges existentes — solo cambiamos el contenido.
	//          El label aparecerá centrado en el merge existente.)
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, setValue("AJ", 1, "VALORIZACIÓN"))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5 — Formulas filas 10-30 (HE exclusivo)
	//          AJ = H_extra = SUM simple (sin comparar con 8)
	//          AL = H_25 (Dom) — mantiene
	//          AN = H_35 (Fer) — mantiene
	//          AP = _E — movido desde AT
	// ═══════════════════════════════════════════════════════════════════
	const rate = 15.0
	for row := 10; row <= 30; row++ {
		reqs = append(reqs, setFormula("AJ", row,
			fmt.Sprintf(`=SUMPRODUCT((N(E%d:AI%d)))`, row, row)))
		reqs = append(reqs, setFormula("AL", row,
			fmt.Sprintf(`=SUMPRODUCT((E$9:AI$9="D")*(N(E%d:AI%d)))`, row, row)))
		reqs = append(reqs, setFormula("AN", row,
			fmt.Sprintf(`=SUMPRODUCT((ISNUMBER(VLOOKUP(E$8,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE)))*(N(E%d:AI%d)))`, row, row)))
		reqs = append(reqs, setFormula("AP", row,
			fmt.Sprintf(`=(AJ%d+AL%d)*%g*1.25+AN%d*%g*1.35`,
				row, row, rate, row, rate)))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 6 — TOTALES row 36
	// ═══════════════════════════════════════════════════════════════════
	dayCols := []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}
	for _, col := range dayCols {
		reqs = append(reqs, setFormula(col, 36, fmt.Sprintf(`=SUM(%s10:%s30)`, col, col)))
	}
	reqs = append(reqs, setFormula("AJ", 36, `=SUM(AJ10:AJ30)`))
	reqs = append(reqs, setFormula("AL", 36, `=SUM(AL10:AL30)`))
	reqs = append(reqs, setFormula("AN", 36, `=SUM(AN10:AN30)`))
	reqs = append(reqs, setFormula("AP", 36, `=SUM(AP10:AP30)`))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 7 — Frame borders
	//          Limpiar borders viejos A1:AT7 y A8:AT36
	//          Crear nuevos A1:AP7 y A8:AP36
	// ═══════════════════════════════════════════════════════════════════
	// Limpiar rango amplio para asegurar
	reqs = append(reqs, clearBorders(0, 0, 6, 45))   // rows 0-6, cols 0-44 (A1:AT7)
	reqs = append(reqs, clearBorders(7, 0, 35, 45))  // rows 7-35, cols 0-44 (A8:AT36)

	// Nuevos frames
	reqs = append(reqs, frameSolidThick(0, 0, 6, 41))   // A1:AP7
	reqs = append(reqs, frameSolidThick(7, 0, 35, 41))  // A8:AP36

	// Internal accents (1px SOLID):
	//   D|E (name area / day grid): col D=3, col E=4
	reqs = append(reqs, innerVerticalBorder(7, 3, 35, 5))
	//   AI|AJ (day grid / summary): col AI=34, col AJ=35
	reqs = append(reqs, innerVerticalBorder(7, 34, 35, 36))

	// ═══════════════════════════════════════════════════════════════════
	// SEND
	// ═══════════════════════════════════════════════════════════════════
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

func resizeCol(col string, pixelSize int) map[string]any {
	ci := colIdx(col)
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetHE,
				"dimension":  "COLUMNS",
				"startIndex": ci,
				"endIndex":   ci + 1,
			},
			"properties": map[string]any{
				"pixelSize": pixelSize,
			},
			"fields": "pixelSize",
		},
	}
}

func mergeRange(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"mergeType": "MERGE_ALL",
		},
	}
}

func unmergeRange(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
		},
	}
}

func clearBorders(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":             borderNone(),
			"bottom":          borderNone(),
			"left":            borderNone(),
			"right":           borderNone(),
			"innerHorizontal": borderNone(),
			"innerVertical":   borderNone(),
		},
	}
}

func frameSolidThick(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":    borderSolid(3),
			"bottom": borderSolid(3),
			"left":   borderSolid(3),
			"right":  borderSolid(3),
		},
	}
}

func innerVerticalBorder(startRow, startCol, endRow, endCol int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"left":  borderSolid(1),
			"right": borderSolid(1),
		},
	}
}

func borderNone() map[string]any {
	return map[string]any{"style": "NONE"}
}

func borderSolid(width int) map[string]any {
	return map[string]any{"style": "SOLID", "width": width}
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

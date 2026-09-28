// stabilize_v3.go — CEO directive 2026-09-28:
// "Header roto. Contenido mal marcado, colores diferentes no se entiende
//  separacion real de datos y contenido"
//
// Cambios:
//   FASE 1: Limpiar todos los bgs en A1:AP36 (volver a default)
//   FASE 2: Aplicar bg gris claro (subtle) al HEADER rows 1-7
//           - Row 1: bg navy (26,57,92) — el título blanco destaca
//           - Rows 2-7: bg gris claro (244,244,244) — sub-header cohesivo
//   FASE 3: Aplicar bg gris claro (244,244,244) a la fila TOTALES (row 36)
//   FASE 4: Mantener Cese marker en ALBA (217,217,217)
//   FASE 5: Mantener bg gris claro muted (234,237,240) en rows sin data 25-33
//   FASE 6: Re-aplicar frame HEADER rows 1-7 + frame DATA rows 8-36
//   FASE 7: Re-aplicar 3 CFs (Feriado, Sabado, Domingo) en E7:AI30
//
// Resultado esperado:
//   - HEADER: bloque cohesivo con bg gris claro
//   - DATA: bg blanco (default), ALBA gris, empleados sin data gris muted
//   - TOTALES: bg gris claro (destaca de data)
//   - CFs pintan feriados/S/D sin contaminar otras celdas
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

// All cols A-AP (1-42 0-based = 0-41)
var allCols = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q",
	"R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI",
	"AJ", "AK", "AL", "AM", "AN", "AO", "AP",
}

func main() {
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 1 — Limpiar todos los bgs en A1:AP36 (volver a default)
	// ═══════════════════════════════════════════════════════════════════
	for row := 1; row <= 36; row++ {
		for _, col := range allCols {
			reqs = append(reqs, setBgColor(col, row, 1.0, 1.0, 1.0))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 2 — HEADER rows 1-7
	// ═══════════════════════════════════════════════════════════════════
	// Row 1: bg navy (26,57,92) — fondo del título
	for _, col := range allCols {
		reqs = append(reqs, setBgColor(col, 1, 0.10, 0.22, 0.36))
	}
	// Rows 2-7: bg gris claro (244,244,244) — sub-header cohesivo
	for row := 2; row <= 7; row++ {
		for _, col := range allCols {
			reqs = append(reqs, setBgColor(col, row, 0.96, 0.96, 0.96))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 3 — TOTALES row 36: bg gris medio (230,230,230) — destaca
	// ═══════════════════════════════════════════════════════════════════
	for _, col := range allCols {
		reqs = append(reqs, setBgColor(col, 36, 0.90, 0.90, 0.90))
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 4 — ALBA CARRILLO (row 9) Cese: bg gris (217,217,217) + italic
	// ═══════════════════════════════════════════════════════════════════
	// A9 y B9 ya tienen formato Cese aplicado en reorganize_agents.go
	// Pero el bg se borró en FASE 1. Re-aplicar.
	for _, col := range []string{"A", "B"} {
		reqs = append(reqs, setBgColor(col, 9, 0.85, 0.85, 0.85))
		reqs = append(reqs, setCeseItalic(col, 9))
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 5 — Empleados sin data (rows 25-33, donde están WONG e YANA)
	//          bg gris claro muted (240,240,240) para distinguir de activos
	// ═══════════════════════════════════════════════════════════════════
	for row := 25; row <= 33; row++ {
		// A, B con bg muted
		for _, col := range []string{"A", "B"} {
			reqs = append(reqs, setBgColor(col, row, 0.95, 0.95, 0.95))
		}
		// AJ, AL, AN, AP con bg muted (antes sandía)
		for _, col := range []string{"AJ", "AL", "AN", "AP"} {
			reqs = append(reqs, setBgColor(col, row, 0.95, 0.95, 0.95))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 6 — Frame borders
	// ═══════════════════════════════════════════════════════════════════
	// Clear all borders en A1:AP36
	reqs = append(reqs, clearAllBorders(0, 0, 35, 41))

	// HEADER frame: rows 1-7, cols 1-42 (0-based: rows 0-6, cols 0-41)
	// 0-based: endRow=7, endCol=42
	reqs = append(reqs, frameBorder(0, 0, 6, 41, 3)) // rows 1-7, cols 1-42 (3px navy)

	// DATA frame: rows 8-36, cols 1-42 (0-based: rows 7-35, cols 0-41)
	// 0-based: endRow=36, endCol=42
	reqs = append(reqs, frameBorder(7, 0, 35, 41, 3)) // rows 8-36, cols 1-42 (3px orange)

	// Internal accents
	reqs = append(reqs, internalVertical(7, 0, 35, 1, 1))  // A|B vertical
	reqs = append(reqs, internalVertical(7, 34, 35, 35, 1)) // AI|AJ vertical

	// Bottom border thicker for TOTALES row
	reqs = append(reqs, internalHorizontal(35, 0, 36, 42, 2)) // row 36 top thicker

	// ═══════════════════════════════════════════════════════════════════
	// FASE 7 — Re-aplicar 3 CFs (Feriado, Sabado, Domingo) en E7:AI30
	// ═══════════════════════════════════════════════════════════════════
	// Delete old CFs first
	cfCount := getCFCount(tok)
	for i := cfCount - 1; i >= 0; i-- {
		reqs = append(reqs, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId": sheetHE,
				"index":   i,
			},
		})
	}
	// Add 3 CFs
	const dateRef = `INDIRECT(ADDRESS(7,COLUMN(),4))`
	reqs = append(reqs, addCFFeriado(dateRef))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=7`, 1.0, 0.92, 0.81))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=6`, 0.94, 0.96, 0.97))

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
	if resp.StatusCode >= 200 { fmt.Printf("DEBUG: status=%d body=%s\n", resp.StatusCode, string(respB)[:min(500, len(string(respB)))]); } if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Stabilized v3: cleared bgs, HEADER navy+gray, DATA white, Cese gray, total gray, CFs\n")
}

func getCFCount(tok string) int {
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(sheetId,title),conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets, _ := out["sheets"].([]any)
	for _, sh := range sheets {
		sm, _ := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		if sidF, ok := props["sheetId"].(float64); ok && int64(sidF) == sheetHE {
			cfAny := sm["conditionalFormats"]
			if cfAny == nil {
				return 0
			}
			if cf, ok := cfAny.([]any); ok {
				return len(cf)
			}
		}
	}
	return 0
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

func setBgColor(col string, row int, r, g, b float64) map[string]any {
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
									"red":   r,
									"green": g,
									"blue":  b,
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

func setCeseItalic(col string, row int) map[string]any {
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
								"textFormat": map[string]any{
									"italic": true,
									"foregroundColor": map[string]any{
										"red":   0.4,
										"green": 0.4,
										"blue":  0.4,
									},
								},
							},
						},
					},
				},
			},
			"fields": "userEnteredFormat.textFormat",
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

func internalVertical(startRow, startCol, endRow, endCol int, width int) map[string]any {
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

func internalHorizontal(startRow, startCol, endRow, endCol int, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top": map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func addCF(formula string, r, g, b float64) map[string]any {
	rangeDays := []any{
		map[string]any{
			"sheetId":          sheetHE,
			"startRowIndex":    6,
			"endRowIndex":      30,
			"startColumnIndex": 4,
			"endColumnIndex":   35,
		},
	}
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeDays,
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

func addCFFeriado(dateRef string) map[string]any {
	rangeDays := []any{
		map[string]any{
			"sheetId":          sheetHE,
			"startRowIndex":    6,
			"endRowIndex":      30,
			"startColumnIndex": 4,
			"endColumnIndex":   35,
		},
	}
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": rangeDays,
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type": "CUSTOM_FORMULA",
						"values": []any{
							map[string]any{
								"userEnteredValue": `=ISNUMBER(VLOOKUP(` + dateRef + `,INDIRECT("'01_FERIADOS_PE'!$A$2:$A$100"),1,FALSE))`,
							},
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
	body, _ := json.Marshal(map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	})
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	res := map[string]any{}
	json.Unmarshal(b, &res)
	tok, _ := res["access_token"].(string)
	return tok
}

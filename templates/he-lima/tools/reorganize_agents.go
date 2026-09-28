// reorganize_agents.go — CEO directive 2026-09-28:
// "Limpia todos los registros de Horas extras, termina de organizar bien a
//  mis 25-26 agentes, organiza la separacion, que tengan un espacio inteligente
//  y se lean los nombres"
//
// Cambios:
//  1. Limpiar inputs E10:AI30 (datos demo del Excel legacy)
//  2. Ajustar widths: A=30, B=260, C=0, D=0 (cols C/D eliminadas)
//  3. Re-ordenar 25 agentes ALFABÉTICAMENTE (Apellido Paterno)
//  4. Marcar ALBA CARRILLO (Cese) con bg gris + italic
//  5. Re-aplicar frame DATA A8:AP36
//  6. Re-aplicar freeze panes: 7 rows + 2 cols
//  7. Re-aplicar CFs (rango E7:AI30)
//
// Total width: 1182px (cabe en 1280px viewport)
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
	"sort"
	"strings"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetHE       = int64(222789142)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

// 25 agents ordered alphabetically by apellido paterno
// Format: (apellido_paterno, apellido_materno, nombres, estado)
var agents = []agent{
	{"ALBA", "CARRILLO", "OSCAR DANIEL", "Cese"},
	{"ALVARADO", "RODRIGUEZ", "JACKELIN CONSUELO", "Activo"},
	{"ARIAS", "CUEVA", "STEPHANNI CECILIA", "Activo"},
	{"ARBIETO", "DIAZ", "PABLO ANTONIO", "Activo"},
	{"CANALES", "COPLAT", "ZAYDA JOSCELINE", "Activo"},
	{"CARRILLO", "SALINAS", "JAZMIN ROSAURA", "Activo"},
	{"ESCUDERO", "GRANDEZ", "CRISTHINA ISABEL", "Activo"},
	{"FLORES", "SUCRE", "SHARON", "Activo"},
	{"GOMEZ", "FLORES", "JONATHAN ARTURO", "Activo"},
	{"HALL", "CASANOVA", "SASHA MIRELLI", "Activo"},
	{"HILARIO", "PURIZACA", "BEATRIZ PATRICIA", "Activo"},
	{"IMAN", "RUJEL", "YADIRA CATIZZA", "Activo"},
	{"IPARRAGUIRRE", "VALVERDE", "LESLIE MAYTE", "Activo"},
	{"LAU", "ROCHA", "CARLOS HERNAN", "Activo"},
	{"LEON", "CASAN", "JAQUELINE EVA", "Activo"},
	{"LLAMPASI", "CRISOSTOMO", "LIZET YULY", "Activo"},
	{"LLANOS", "BRITO", "EITHAN LEANDRO MATHIAS", "Activo"},
	{"MENDOZA", "CARMEN", "CECILIA LILIANA", "Activo"},
	{"OSORIO", "PEREZ", "TATIANA LUCERO", "Activo"},
	{"RAMIREZ", "HUAMAN", "JOHNN MILTON", "Activo"},
	{"URIBE", "DAVILA", "MONICA PAHOLA", "Activo"},
	{"USURIN", "VICENTE", "MIRIAM MARIA", "Activo"},
	{"VARGAS", "TAPIA", "ISABEL", "Activo"},
	{"WONG", "CORREA", "PATRICIA DEL PILAR", "Activo"},
	{"YANA", "JUNCO", "ROSALIA ROCIO", "Activo"},
}

type agent struct {
	Paterno string
	Materno string
	Nombres string
	Estado  string
}

func (a agent) FullName() string {
	return a.Paterno + " " + a.Materno + " " + a.Nombres
}

func main() {
	// Sanity check: sort again
	sort.SliceStable(agents, func(i, j int) bool {
		return agents[i].Paterno < agents[j].Paterno
	})

	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 1 — Limpiar inputs E10:AI30 (datos demo)
	// ═══════════════════════════════════════════════════════════════════
	dayCols := []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI"}
	for _, col := range dayCols {
		for row := 10; row <= 33; row++ {
			reqs = append(reqs, clearCell(col, row))
		}
	}
	// Clear TOTALES row 36 too
	for _, col := range dayCols {
		reqs = append(reqs, clearCell(col, 36))
	}
	reqs = append(reqs, clearCell("AJ", 36))
	reqs = append(reqs, clearCell("AL", 36))
	reqs = append(reqs, clearCell("AN", 36))
	reqs = append(reqs, clearCell("AP", 36))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 2 — Ajustar widths: A=30, B=260, C=0, D=0
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, resizeCol("A", 30))
	reqs = append(reqs, resizeCol("B", 260))
	reqs = append(reqs, resizeCol("C", 0))
	reqs = append(reqs, resizeCol("D", 0))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 3 — Escribir N° y nombres en rows 9-33 (alfabético)
	// ═══════════════════════════════════════════════════════════════════
	for i, ag := range agents {
		row := 9 + i
		// A: N°
		reqs = append(reqs, setNumber("A", row, i+1))
		// B: APELLIDOS Y NOMBRES completo
		reqs = append(reqs, setString("B", row, ag.FullName()))

		// Si es Cese, marcar con bg gris + italic
		if ag.Estado == "Cese" {
			reqs = append(reqs, setCeseFormat("A", row))
			reqs = append(reqs, setCeseFormat("B", row))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// FASE 4 — Re-aplicar frame DATA A8:AP36
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, clearAllBorders(7, 0, 35, 41))
	reqs = append(reqs, frameBorder(7, 0, 35, 41, 3))
	// Internal accent: A|B vertical
	reqs = append(reqs, internalAccent(7, 0, 35, 1, 1))
	// AI|AJ vertical (day grid / summary)
	reqs = append(reqs, internalAccent(7, 34, 35, 35, 1))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 5 — Re-aplicar CFs (range E7:AI30)
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
	// Add 3 CFs (with INDIRECT(ADDRESS(7,COLUMN(),4)))
	const dateRef = `INDIRECT(ADDRESS(7,COLUMN(),4))`
	reqs = append(reqs, addCFFeriado(dateRef))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=7`, 1.0, 0.92, 0.81))
	reqs = append(reqs, addCF(`=WEEKDAY(`+dateRef+`,2)=6`, 0.94, 0.96, 0.97))

	// ═══════════════════════════════════════════════════════════════════
	// FASE 6 — Freeze panes: 1 row + 0 cols
	//         Title (row 1) siempre visible. Sin freeze horizontal
	//         porque los merges A1:AJ1 y AK1:AU1 lo impiden.
	//         Trade-off: aceptable porque el sheet cabe en 1182px
	//         en pantallas 1280px+.
	// ═══════════════════════════════════════════════════════════════════
	reqs = append(reqs, freezePanes(1, 0))

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
	fmt.Printf("✅ Reorganized: cleared inputs, 25 agents alpha, B=260px, freeze 1x0, 3 CFs\n")
	fmt.Printf("Total agents: %d (1 Cese: %s)\n", len(agents), agents[0].FullName())
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
	sheets := out["sheets"].([]any)
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		if sidF, ok := props["sheetId"].(float64); ok && int64(sidF) == sheetHE {
			if cf, ok := sm["conditionalFormats"].([]any); ok {
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

func setString(col string, row int, value string) map[string]any {
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

func setNumber(col string, row int, value int) map[string]any {
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
							"userEnteredValue": map[string]any{"numberValue": float64(value)},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func setCeseFormat(col string, row int) map[string]any {
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
									"red":   0.85,
									"green": 0.85,
									"blue":  0.85,
								},
								"textFormat": map[string]any{
									"italic":       true,
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
			"fields": "userEnteredFormat(backgroundColor,textFormat)",
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

func freezePanes(frozenRowCount, frozenColCount int) map[string]any {
	return map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": sheetHE,
				"gridProperties": map[string]any{
					"frozenRowCount":    frozenRowCount,
					"frozenColumnCount": frozenColCount,
				},
			},
			"fields": "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
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

// Keep imports happy
var _ = strings.HasPrefix

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const finalizeSheetID = "1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw"
const finalizeSheetNum = 294616252

func runCitrixFinalize(repoRoot string, args []string) error {
	v := NewCredStore(repoRoot)
	creds, _ := v.Load()
	if err := EnsureFresh(creds); err != nil {
		return err
	}
	token := creds.AccessToken

	fmt.Println("=== STEP 1: Re-write complete sheet with correct order ===")
	values := [][]interface{}{
		// Row 1: TITLE (will be merged across A1:I1)
		{"SUSTENTO DE ACTUALIZACIÓN DE SISTEMA CITRIX", "", "", "", "", "", "", ""},
		// Row 2: SUBTITLE (will be merged across A2:H2)
		{"Para RRHH/Nómina — Tiempo SIN TRABAJO por mantenimiento", "", "", "", "", "", "", ""},
		// Row 3: empty
		{"", "", "", "", "", "", "", ""},
		// Row 4: SECCIÓN 1 header (merged A4:H4)
		{"═══ SECCIÓN 1: DATOS DEL EVENTO ═══", "", "", "", "", "", "", ""},
		// Row 5: Column headers
		{"CAMPO", "VALOR", "", "", "", "", "", ""},
		// Row 6-10: Metadata
		{"Aplicativo", "CITRIX", "", "", "", "", "", ""},
		{"Motivo", "Actualización de sistema", "", "", "", "", "", ""},
		{"Fecha del evento", "", "", "", "", "", "", ""},
		{"Hora inicio", "", "", "", "", "", "", ""},
		{"Hora fin", "", "", "", "", "", "", ""},
		// Row 11: empty
		{"", "", "", "", "", "", "", ""},
		// Row 12: SECCIÓN 2 header (merged A12:H12)
		{"═══ SECCIÓN 2: AGENTES ═══", "", "", "", "", "", "", ""},
		// Row 13: Column headers
		{"N°", "DNI", "APELLIDOS_Y_NOMBRES", "FECHA", "HORA_INICIO", "HORA_FIN", "DURACIÓN", "OBSERVACIONES", "SUPERVISOR"},
		// Row 14: 1 fila de prueba
		{"1", "'09887772", "URIBE DAVILA MONICA PAHOLA", "2026-09-30", "02:00", "04:00", "=IF(OR(E14=\"\",F14=\"\"),\"\",(F14-E14)*24)", "Actualización CITRIX", "WILLIAMS ALEXANDER HUERTA SALVADOR"},
	}
	updateBody, _ := json.Marshal(map[string]any{"values": values})
	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/CITRIX_UPDATE_LOG!A1:I14?valueInputOption=USER_ENTERED", finalizeSheetID), bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("update: %d %s", resp.StatusCode, string(body))
	}
	fmt.Println("✅ All 14 rows written with correct order")

	fmt.Println("\n=== STEP 2: Apply formatting ===")
	requests := []map[string]any{
		// Cleanup: reset bg of all cells to white first
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 60,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
			}},
			"fields": "userEnteredFormat.backgroundColor",
		}},
		// MERGE row 1 (title)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 1,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"mergeType": "MERGE_ALL",
		}},
		// MERGE row 2 (subtitle)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 1, "endRowIndex": 2,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"mergeType": "MERGE_ALL",
		}},
		// MERGE row 4 (SECCIÓN 1)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 3, "endRowIndex": 4,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"mergeType": "MERGE_ALL",
		}},
		// MERGE row 12 (SECCIÓN 2)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 11, "endRowIndex": 12,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"mergeType": "MERGE_ALL",
		}},
		// TITLE style (row 1)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 1,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 14,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
			}},
			"fields": "userEnteredFormat",
		}},
		// SUBTITLE style (row 2)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 1, "endRowIndex": 2,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.91, "green": 0.91, "blue": 0.91},
				"textFormat": map[string]any{
					"italic": true, "fontSize": 10,
				},
			}},
			"fields": "userEnteredFormat",
		}},
		// SECCIÓN headers (rows 4 and 12)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 3, "endRowIndex": 4,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.29, "green": 0.33, "blue": 0.4},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 12,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
			}},
			"fields": "userEnteredFormat",
		}},
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 11, "endRowIndex": 12,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.29, "green": 0.33, "blue": 0.4},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 12,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
			}},
			"fields": "userEnteredFormat",
		}},
		// CAMPO|VALOR header (row 5)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 4, "endRowIndex": 5,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.18, "green": 0.49, "blue": 0.2},
				"textFormat": map[string]any{
					"bold": true,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
			}},
			"fields": "userEnteredFormat",
		}},
		// Metadata rows 6-10: gray bg
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 5, "endRowIndex": 10,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.97, "green": 0.97, "blue": 0.98},
			}},
			"fields": "userEnteredFormat.backgroundColor",
		}},
		// Yellow on metadata pending cells (rows 8-10 col 1 = B8-B10)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 7, "endRowIndex": 10,
				"startColumnIndex": 1, "endColumnIndex": 2,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1.0, "green": 0.95, "blue": 0.8},
			}},
			"fields": "userEnteredFormat.backgroundColor",
		}},
		// AGENTES column headers (row 13)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 12, "endRowIndex": 13,
				"startColumnIndex": 0, "endColumnIndex": 8,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.18, "green": 0.49, "blue": 0.2},
				"textFormat": map[string]any{
					"bold": true,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
			}},
			"fields": "userEnteredFormat",
		}},
		// Yellow on agentes pending cells (rows 14+, cols D,E,F = indices 3,4,5)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 3, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1.0, "green": 0.95, "blue": 0.8},
			}},
			"fields": "userEnteredFormat.backgroundColor",
		}},
		// DNI column (B): TEXT format
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 1, "endColumnIndex": 2,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "TEXT"},
			}},
			"fields": "userEnteredFormat.numberFormat",
		}},
		// FECHA column (D): DATE format
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 3, "endColumnIndex": 4,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "DATE", "pattern": "yyyy-mm-dd"},
			}},
			"fields": "userEnteredFormat.numberFormat",
		}},
		// HORA columns (E,F): TIME format
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 4, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "TIME", "pattern": "hh:mm"},
			}},
			"fields": "userEnteredFormat.numberFormat",
		}},
		// DURACIÓN (G): NUMBER format
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 6, "endColumnIndex": 7,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "NUMBER", "pattern": "0.00"},
			}},
			"fields": "userEnteredFormat.numberFormat",
		}},
		// Column widths
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 0, "endIndex": 1},
			"properties": map[string]any{"pixelSize": 50}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 1, "endIndex": 2},
			"properties": map[string]any{"pixelSize": 110}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 2, "endIndex": 3},
			"properties": map[string]any{"pixelSize": 280}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 3, "endIndex": 4},
			"properties": map[string]any{"pixelSize": 130}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 4, "endIndex": 5},
			"properties": map[string]any{"pixelSize": 110}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 5, "endIndex": 6},
			"properties": map[string]any{"pixelSize": 110}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 6, "endIndex": 7},
			"properties": map[string]any{"pixelSize": 110}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 7, "endIndex": 8},
			"properties": map[string]any{"pixelSize": 280}, "fields": "pixelSize"}},
		{"updateDimensionProperties": map[string]any{
			"range": map[string]any{"sheetId": finalizeSheetNum, "dimension": "COLUMNS", "startIndex": 8, "endIndex": 9},
			"properties": map[string]any{"pixelSize": 220}, "fields": "pixelSize"}},
		// Freeze rows 13 (AGENTES header always visible)
		{"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": finalizeSheetNum,
				"gridProperties": map[string]any{"frozenRowCount": 13},
			},
			"fields": "gridProperties.frozenRowCount",
		}},
	}

	formatBody, _ := json.Marshal(map[string]any{"requests": requests})
	req2, _ := http.NewRequest("POST", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", finalizeSheetID), bytes.NewReader(formatBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != 200 {
		return fmt.Errorf("format: %d %s", resp2.StatusCode, string(body2))
	}
	fmt.Println("✅ Format + widths + freeze applied")

	fmt.Println("\n=== ✅ LISTO ===")
	fmt.Printf("URL: https://docs.google.com/spreadsheets/d/%s/edit\n", finalizeSheetID)
	return nil
}

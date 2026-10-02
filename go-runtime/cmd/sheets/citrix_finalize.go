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

	// STEP 1: Clear entire sheet
	fmt.Println("=== STEP 1: Clear entire sheet ===")
	clearReq, _ := http.NewRequest("POST", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values:batchClear", finalizeSheetID), bytes.NewReader([]byte(`{"ranges":["CITRIX_UPDATE_LOG!A1:I60"]}`)))
	clearReq.Header.Set("Authorization", "Bearer "+token)
	clearReq.Header.Set("Content-Type", "application/json")
	if clearResp, _ := http.DefaultClient.Do(clearReq); clearResp != nil {
		clearResp.Body.Close()
		fmt.Printf("  Clear status: %d\n", clearResp.StatusCode)
	}

	// STEP 2: Write 14 rows of complete data (9 cols each)
	fmt.Println("=== STEP 2: Write complete data (all 9 cols consistent) ===")
	empty := func() []interface{} {
		return []interface{}{"", "", "", "", "", "", "", "", ""}
	}
	values := [][]interface{}{
		// Row 1: TITLE
		{"SUSTENTO DE ACTUALIZACIÓN DE SISTEMA CITRIX", "", "", "", "", "", "", "", ""},
		// Row 2: SUBTITLE
		{"Para RRHH/Nómina — Tiempo SIN TRABAJO por mantenimiento", "", "", "", "", "", "", "", ""},
		// Row 3: separator
		empty(),
		// Row 4: SECCIÓN 1 header
		{"═══ SECCIÓN 1: DATOS DEL EVENTO ═══", "", "", "", "", "", "", "", ""},
		// Row 5: CAMPO|VALOR headers
		{"CAMPO", "VALOR", "", "", "", "", "", "", ""},
		// Row 6: Aplicativo
		{"Aplicativo", "CITRIX", "", "", "", "", "", "", ""},
		// Row 7: Motivo
		{"Motivo", "Actualización de sistema", "", "", "", "", "", "", ""},
		// Row 8: Fecha
		{"Fecha del evento", "", "", "", "", "", "", "", ""},
		// Row 9: Hora inicio
		{"Hora inicio", "", "", "", "", "", "", "", ""},
		// Row 10: Hora fin
		{"Hora fin", "", "", "", "", "", "", "", ""},
		// Row 11: separator
		empty(),
		// Row 12: SECCIÓN 2 header
		{"═══ SECCIÓN 2: AGENTES ═══", "", "", "", "", "", "", "", ""},
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
	fmt.Println("✅ 14 rows written (9 cols each)")

	// STEP 3: Apply formatting + merges + borders + widths
	fmt.Println("=== STEP 3: Apply formatting ===")
	
	requests := []map[string]any{
		// First: unmerge everything (clean slate)
		// Note: We need to find all current merges and unmerge them
		// We'll skip this since we cleared the data; merges may persist
		// To be safe, we'll unmerge the known ranges:
		
		// Cleanup all bg colors first (set white)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 60,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
			}},
			"fields": "userEnteredFormat.backgroundColor",
		}},
		
		// --- MERGES ---
		// Row 1 TITLE (merge A1:I1)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 1,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"mergeType": "MERGE_ALL",
		}},
		// Row 2 SUBTITLE (merge A2:I2)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 1, "endRowIndex": 2,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"mergeType": "MERGE_ALL",
		}},
		// Row 4 SECCIÓN 1 header (merge A4:I4)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 3, "endRowIndex": 4,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"mergeType": "MERGE_ALL",
		}},
		// Row 12 SECCIÓN 2 header (merge A12:I12)
		{"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 11, "endRowIndex": 12,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"mergeType": "MERGE_ALL",
		}},

		// --- STYLES (with borders for "encuadrar") ---
		// Row 1 TITLE - navy bg, white bold 14
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 0, "endRowIndex": 1,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.1, "green": 0.23, "blue": 0.36},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 14,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
				"borders": map[string]any{
					"bottom": map[string]any{
						"style": "SOLID",
						"width": 2,
						"color": map[string]any{"red": 0.0, "green": 0.0, "blue": 0.0},
					},
				},
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,borders)",
		}},
		// Row 2 SUBTITLE - gray bg italic
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 1, "endRowIndex": 2,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.91, "green": 0.91, "blue": 0.91},
				"textFormat": map[string]any{
					"italic": true, "fontSize": 10,
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		}},

		// Row 4 SECCIÓN 1 - dark gray bg, white bold 12
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 3, "endRowIndex": 4,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.29, "green": 0.33, "blue": 0.4},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 12,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
				"horizontalAlignment": "LEFT",
				"verticalAlignment": "MIDDLE",
				"borders": map[string]any{
					"top": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.5, "green": 0.5, "blue": 0.5}},
					"bottom": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.5, "green": 0.5, "blue": 0.5}},
				},
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,borders)",
		}},
		// Row 12 SECCIÓN 2 - dark gray bg, white bold 12
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 11, "endRowIndex": 12,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.29, "green": 0.33, "blue": 0.4},
				"textFormat": map[string]any{
					"bold": true, "fontSize": 12,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
				"horizontalAlignment": "LEFT",
				"verticalAlignment": "MIDDLE",
				"borders": map[string]any{
					"top": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.5, "green": 0.5, "blue": 0.5}},
					"bottom": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.5, "green": 0.5, "blue": 0.5}},
				},
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,borders)",
		}},

		// Row 5 CAMPO|VALOR - verde bg, white bold
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 4, "endRowIndex": 5,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.18, "green": 0.49, "blue": 0.2},
				"textFormat": map[string]any{
					"bold": true,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		}},
		// Rows 6-7 (Aplicativo, Motivo) - light gray bg
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 5, "endRowIndex": 7,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.97, "green": 0.97, "blue": 0.98},
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,verticalAlignment)",
		}},
		// Rows 8-10 (Fecha, Hora inicio, Hora fin) - light gray bg
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 7, "endRowIndex": 10,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.97, "green": 0.97, "blue": 0.98},
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,verticalAlignment)",
		}},
		// Yellow on metadata pending (col B = index 1, rows 8-10 = index 7-9)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 7, "endRowIndex": 10,
				"startColumnIndex": 1, "endColumnIndex": 2,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1.0, "green": 0.95, "blue": 0.8},
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,verticalAlignment)",
		}},

		// Row 13 AGENTES headers - verde bg, white bold
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 12, "endRowIndex": 13,
				"startColumnIndex": 0, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 0.18, "green": 0.49, "blue": 0.2},
				"textFormat": map[string]any{
					"bold": true,
					"foregroundColor": map[string]any{"red": 1, "green": 1, "blue": 1},
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
				"wrapStrategy": "WRAP",
			}},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment,wrapStrategy)",
		}},

		// Row 14+ AGENTES data - default white bg
		// But apply yellow on pendientes (cols D,E,F = indices 3,4,5)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 3, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1.0, "green": 0.95, "blue": 0.8},
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,verticalAlignment)",
		}},

		// --- FORMATS for cell types ---
		// DNI column (B = index 1, rows 14+) - TEXT
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 1, "endColumnIndex": 2,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "TEXT"},
				"horizontalAlignment": "LEFT",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// FECHA column (D = index 3)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 3, "endColumnIndex": 4,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "DATE", "pattern": "yyyy-mm-dd"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// HORA_INICIO + HORA_FIN (E,F = indices 4,5)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 4, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "TIME", "pattern": "hh:mm"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// DURACIÓN (G = index 6)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 6, "endColumnIndex": 7,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "NUMBER", "pattern": "0.00"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// N° column (A) - CENTER
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 0, "endColumnIndex": 1,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(horizontalAlignment,verticalAlignment)",
		}},
		// OBSERVACIONES (H = index 7) y SUPERVISOR (I = index 8)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 7, "endColumnIndex": 9,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"horizontalAlignment": "LEFT",
				"verticalAlignment": "MIDDLE",
				"wrapStrategy": "WRAP",
			}},
			"fields": "userEnteredFormat(horizontalAlignment,verticalAlignment,wrapStrategy)",
		}},
		// APELLIDOS_Y_NOMBRES (C = index 2)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": finalizeSheetNum, "startRowIndex": 13, "endRowIndex": 60,
				"startColumnIndex": 2, "endColumnIndex": 3,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"horizontalAlignment": "LEFT",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(horizontalAlignment,verticalAlignment)",
		}},

		// --- COLUMN WIDTHS (no comprimido) ---
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

		// --- FREEZE PANES (rows 1-13 frozen, header always visible) ---
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
	fmt.Println("✅ Format + widths + freeze + merges applied")

	fmt.Println("\n=== ✅ LISTO ===")
	fmt.Printf("URL: https://docs.google.com/spreadsheets/d/%s/edit\n", finalizeSheetID)
	return nil
}

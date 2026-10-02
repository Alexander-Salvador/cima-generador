package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const formulasSheetID = "1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw"
const formulasSheetNum = 294616252

func runCitrixFormulas(repoRoot string, args []string) error {
	v := NewCredStore(repoRoot)
	creds, _ := v.Load()
	if err := EnsureFresh(creds); err != nil {
		return err
	}
	token := creds.AccessToken

	// STEP 1: Clear rows 4-100 (keep rows 1-3 intact)
	fmt.Println("=== STEP 1: Clear rows 4-100 ===")
	clearReq, _ := http.NewRequest("POST", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values:batchClear", formulasSheetID), bytes.NewReader([]byte(`{"ranges":["CITRIX_UPDATE_LOG!A4:I100"]}`)))
	clearReq.Header.Set("Authorization", "Bearer "+token)
	clearReq.Header.Set("Content-Type", "application/json")
	if clearResp, _ := http.DefaultClient.Do(clearReq); clearResp != nil {
		clearResp.Body.Close()
		fmt.Printf("  Clear status: %d\n", clearResp.StatusCode)
	}

	// STEP 2: Update G3 with smart 24h formula (fix the -4.00 bug)
	fmt.Println("=== STEP 2: Update G3 with smart 24h formula ===")
	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/CITRIX_UPDATE_LOG!G3?valueInputOption=USER_ENTERED", formulasSheetID), bytes.NewReader([]byte(`{"values":[["=IF(OR(E3=\"\",F3=\"\"),\"\",IF(F3<E3,(F3-E3+1)*24,(F3-E3)*24))"]]}`)))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")
	if updateResp, _ := http.DefaultClient.Do(updateReq); updateResp != nil {
		body, _ := io.ReadAll(updateResp.Body)
		fmt.Printf("  update status: %d, body: %s\n", updateResp.StatusCode, string(body))
		updateResp.Body.Close()
	}

	// STEP 3: Apply formats and formulas to rows 4-100
	fmt.Println("=== STEP 3: Apply formats + smart formulas (rows 4-100) ===")

	// Formula for G4:G100 (smart 24h with wrap-around)

	requests := []map[string]any{
		// Yellow background for pendiente cells (D4:F100)
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": formulasSheetNum, "startRowIndex": 3, "endRowIndex": 100,
				"startColumnIndex": 3, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{"red": 1.0, "green": 0.95, "blue": 0.8},
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(backgroundColor,verticalAlignment)",
		}},
		// DATE format on D4:D100
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": formulasSheetNum, "startRowIndex": 3, "endRowIndex": 100,
				"startColumnIndex": 3, "endColumnIndex": 4,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "DATE", "pattern": "yyyy-mm-dd"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// TIME format on E4:F100
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": formulasSheetNum, "startRowIndex": 3, "endRowIndex": 100,
				"startColumnIndex": 4, "endColumnIndex": 6,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "TIME", "pattern": "hh:mm"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// NUMBER format on G4:G100
		{"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId": formulasSheetNum, "startRowIndex": 3, "endRowIndex": 100,
				"startColumnIndex": 6, "endColumnIndex": 7,
			},
			"cell": map[string]any{"userEnteredFormat": map[string]any{
				"numberFormat": map[string]any{"type": "NUMBER", "pattern": "0.00"},
				"horizontalAlignment": "CENTER",
				"verticalAlignment": "MIDDLE",
			}},
			"fields": "userEnteredFormat(numberFormat,horizontalAlignment,verticalAlignment)",
		}},
		// Formulas in G4:G100 (smart 24h with wrap-around)
		{"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          formulasSheetNum,
				"startRowIndex":    3,
				"endRowIndex":      100,
				"startColumnIndex": 6,
				"endColumnIndex":   7,
			},
			"rows": func() []map[string]any {
				rows := []map[string]any{}
				for i := 4; i <= 100; i++ {
					// Formula: IF(OR(E[i]="",F[i]=""),"",IF(F[i]<E[i],(F[i]-E[i]+1)*24,(F[i]-E[i])*24))
					formula := fmt.Sprintf(`=IF(OR(E%d="",F%d=""),"",IF(F%d<E%d,(F%d-E%d+1)*24,(F%d-E%d)*24))`, i, i, i, i, i, i, i, i)
					rows = append(rows, map[string]any{
						"values": []map[string]any{
							{"userEnteredValue": map[string]any{"formulaValue": formula}},
						},
					})
				}
				return rows
			}(),
			"fields": "userEnteredValue",
		}},
	}

	formatBody, _ := json.Marshal(map[string]any{"requests": requests})
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", formulasSheetID), bytes.NewReader(formatBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("format: %d %s", resp.StatusCode, string(body))
	}
	fmt.Println("✅ Formats + formulas applied to rows 4-100")

	fmt.Println("\n=== ✅ LISTO ===")
	fmt.Println("Smart formula applied:")
	fmt.Println("  =IF(OR(E[i]=\"\",F[i]=\"\"),\"\",IF(F[i]<E[i],(F[i]-E[i]+1)*24,(F[i]-E[i])*24))")
	fmt.Println()
	fmt.Println("Logic:")
	fmt.Println("  - Si E o F vacías → muestra vacío")
	fmt.Println("  - Si F < E (cambio de día) → suma 24h")
	fmt.Println("  - Si F >= E → resta directo")
	fmt.Println()
	fmt.Printf("URL: https://docs.google.com/spreadsheets/d/%s/edit\n", formulasSheetID)
	return nil
}

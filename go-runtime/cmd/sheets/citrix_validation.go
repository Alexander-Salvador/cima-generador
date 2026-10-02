package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const valSheetID = "1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw"
const valSheetNum = 294616252

func runCitrixValidation(repoRoot string, args []string) error {
	v := NewCredStore(repoRoot)
	creds, _ := v.Load()
	if err := EnsureFresh(creds); err != nil {
		return err
	}
	token := creds.AccessToken

	fmt.Println("=== Apply Data Validation for Date picker ===")
	fmt.Println("Note: Google Sheets API no soporta TIME validation nativo.")
	fmt.Println("      Para time picker: doble-click en celdas con formato TIME.")
	fmt.Println()

	requests := []map[string]any{
		// DATE picker for FECHA (D4:D100) - DATE_IS_VALID shows calendar
		{
			"setDataValidation": map[string]any{
				"range": map[string]any{
					"sheetId":          valSheetNum,
					"startRowIndex":    3,
					"endRowIndex":      100,
					"startColumnIndex": 3,
					"endColumnIndex":   4,
				},
				"rule": map[string]any{
					"condition": map[string]any{
						"type": "DATE_IS_VALID",
					},
					"inputMessage": "Click to open date picker",
					"showCustomUi": true,
					"strict":      false,
				},
			},
		},
	}

	body, _ := json.Marshal(map[string]any{"requests": requests})
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", valSheetID),
		bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body2, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body2))
	}
	fmt.Println("✅ DATE picker validation applied (D4:D100)")
	fmt.Println()
	fmt.Println("Ahora al hacer click en D4-D100 → aparece calendar picker ✅")
	fmt.Println()
	fmt.Println("Para TIME picker:")
	fmt.Println("  → Doble-click en celdas E o F (formato TIME ya aplicado)")
	fmt.Println("  → Aparece icono de reloj para seleccionar hora")
	fmt.Println()
	fmt.Printf("URL: https://docs.google.com/spreadsheets/d/%s/edit\n", valSheetID)
	return nil
}
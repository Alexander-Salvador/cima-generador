package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func runCitrixInspect(repoRoot string, args []string) error {
	v := NewCredStore(repoRoot)
	creds, _ := v.Load()
	if err := EnsureFresh(creds); err != nil {
		return err
	}
	token := creds.AccessToken
	sheetID := "1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw"

	url := "https://sheets.googleapis.com/v4/spreadsheets/" + sheetID + "?fields=sheets.data.rowData.values(dataValidation.condition,formattedValue,userEnteredFormat.numberFormat)"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Sheets []struct {
			Data []struct {
				RowData []struct {
					Values []struct {
						FormattedValue    string `json:"formattedValue"`
						DataValidation    map[string]any `json:"dataValidation"`
						UserEnteredFormat struct {
							NumberFormat struct {
								Type string `json:"type"`
							} `json:"numberFormat"`
						} `json:"userEnteredFormat"`
					} `json:"values"`
				} `json:"rowData"`
			} `json:"data"`
		} `json:"sheets"`
	}
	json.Unmarshal(body, &result)
	for _, s := range result.Sheets {
		for i, row := range s.Data[0].RowData {
			if i >= 8 {
				break
			}
			for j, v := range row.Values {
				if j >= 6 {
					break
				}
				if v.DataValidation != nil {
					fmt.Printf("R%dC%d: val=%v\n", i+1, j+1, v.DataValidation)
				}
			}
		}
	}
	fmt.Println("Formats check (rows 3-7):")
	for _, s := range result.Sheets {
		for i, row := range s.Data[0].RowData {
			if i >= 7 || i < 2 {
				continue
			}
			fmt.Printf("Row %d:\n", i+1)
			for j, v := range row.Values {
				if j >= 9 {
					break
				}
				nf := v.UserEnteredFormat.NumberFormat
				if nf.Type != "" {
					fmt.Printf("  C%d: format=%s\n", j+1, nf.Type)
				}
			}
		}
	}
	return nil
}

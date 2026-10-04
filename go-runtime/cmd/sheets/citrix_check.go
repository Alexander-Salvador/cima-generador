package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func runCitrixCheck(repoRoot string, args []string) error {
	v := NewCredStore(repoRoot)
	creds, _ := v.Load()
	if err := EnsureFresh(creds); err != nil {
		return err
	}
	token := creds.AccessToken
	sheetID := "1imL6DD1ygTRyX95F3bwD1IIUQjV8YIlO_8umDTg3KEw"

	// Read all 6 rows with values
	url := "https://sheets.googleapis.com/v4/spreadsheets/" + sheetID + "/values/CITRIX_UPDATE_LOG!A1:I10?valueRenderOption=FORMATTED_VALUE"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Values [][]interface{} `json:"values"`
	}
	json.Unmarshal(body, &result)
	fmt.Println("=== A1:I10 FORMATTED ===")
	for i, row := range result.Values {
		fmt.Printf("Row %d: ", i+1)
		for j, cell := range row {
			if j > 0 {
				fmt.Print(" | ")
			}
			if s, ok := cell.(string); ok && s != "" {
				fmt.Printf("%s", s)
			} else {
				fmt.Printf("_")
			}
		}
		fmt.Println()
	}
	return nil
}

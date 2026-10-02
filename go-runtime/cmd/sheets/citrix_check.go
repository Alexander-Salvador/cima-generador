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
	if len(args) > 0 && args[0] != "" {
		sheetID = args[0]
	}

	// Drive metadata
	url := "https://www.googleapis.com/drive/v3/files/" + sheetID + "?fields=name,mimeType,webViewLink,permissions"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Name         string `json:"name"`
		MimeType     string `json:"mimeType"`
		WebViewLink  string `json:"webViewLink"`
		Permissions  []struct {
			Type string `json:"type"`
			Role string `json:"role"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshal: %w body=%s", err, string(body))
	}

	fmt.Println("=== ✅ SPREADSHEET VERIFICADO EN DRIVE ===")
	fmt.Printf("Name:        %s\n", result.Name)
	fmt.Printf("MimeType:    %s\n", result.MimeType)
	fmt.Printf("WebViewLink: %s\n", result.WebViewLink)
	fmt.Printf("Permissions:\n")
	for _, p := range result.Permissions {
		fmt.Printf("  - type=%s role=%s\n", p.Type, p.Role)
	}

	// Read first 14 rows
	readURL := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/CITRIX_UPDATE_LOG!A1:I14", sheetID)
	req2, _ := http.NewRequest("GET", readURL, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)

	if resp2.StatusCode != 200 {
		return fmt.Errorf("read status %d: %s", resp2.StatusCode, string(body2))
	}

	var values struct {
		Values [][]interface{} `json:"values"`
	}
	json.Unmarshal(body2, &values)

	fmt.Println("\n=== DATOS LEÍDOS (A1:I14) ===")
	for i, row := range values.Values {
		fmt.Printf("Row %d: ", i+1)
		for j, cell := range row {
			if j > 0 {
				fmt.Print(" | ")
			}
			if s, ok := cell.(string); ok {
				fmt.Printf("%s", s)
			} else {
				fmt.Printf("%v", cell)
			}
		}
		fmt.Println()
	}

	return nil
}

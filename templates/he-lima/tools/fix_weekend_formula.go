// fix_weekend_formula.go — Fix the conditional formatting formula for weekend highlight.
// Wrong: =$E9 (column E is fixed for ALL cells, so only E9 actually triggers)
// Right: =E$9 (column is relative, so each column checks ITS row-9 cell)
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

	// First, find and delete the existing weekend rules (with the wrong formula)
	// Then re-add with correct formula

	// Get current rules
	url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
		"?fields=sheets(conditionalFormats)"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]any
	json.Unmarshal(body, &raw)

	sheets := raw["sheets"].([]any)
	cfs := sheets[0].(map[string]any)["conditionalFormats"].([]any)

	requests := []any{}

	// Find and delete weekend rules (we need their indices)
	deleteRules := []int{}
	for i, cf := range cfs {
		cfm, _ := cf.(map[string]any)
		ranges, _ := cfm["ranges"].([]any)
		for _, r := range ranges {
			rm, _ := r.(map[string]any)
			sheetId, _ := rm["sheetId"].(float64)
			if int64(sheetId) == sheetHE {
				// Check if this is our weekend rule
				rule, _ := cfm["booleanRule"].(map[string]any)
				if rule != nil {
					cond, _ := rule["condition"].(map[string]any)
					values, _ := cond["values"].([]any)
					if len(values) > 0 {
						first, _ := values[0].(map[string]any)
						formula, _ := first["userEnteredValue"].(string)
						if formula == "=$E9=\"sáb\"" || formula == "=$E9=\"dom\"" {
							deleteRules = append(deleteRules, i)
						}
					}
				}
			}
		}
	}

	// Delete weekend rules (in reverse to preserve indices)
	for idx := len(deleteRules) - 1; idx >= 0; idx-- {
		requests = append(requests, map[string]any{
			"deleteConditionalFormatRule": map[string]any{
				"sheetId":          sheetHE,
				"index":            deleteRules[idx],
			},
		})
	}

	// Re-add with correct formula: =E$9 (relative column, fixed row 9)
	for _, weekend := range []string{"sáb", "dom"} {
		requests = append(requests, map[string]any{
			"addConditionalFormatRule": map[string]any{
				"rule": map[string]any{
					"ranges": []any{
						map[string]any{
							"sheetId":          sheetHE,
							"startRowIndex":    7,
							"endRowIndex":      9,
							"startColumnIndex": 4,
							"endColumnIndex":   35,
						},
					},
					"booleanRule": map[string]any{
						"condition": map[string]any{
							"type": "CUSTOM_FORMULA",
							"values": []any{
								map[string]any{
									"userEnteredValue": fmt.Sprintf("=E$9=\"%s\"", weekend),
								},
							},
						},
						"format": map[string]any{
							"backgroundColor": map[string]any{
								"red":   0.9921569,
								"green": 0.9215686,
								"blue":  0.8156863,
							},
							"textFormat": map[string]any{
								"foregroundColor": map[string]any{
									"red":   0.0588235,
									"green": 0.1411765,
									"blue":  0.2509804,
								},
								"bold": true,
							},
						},
					},
				},
			},
		})
	}

	body2 := map[string]any{"requests": requests}
	raw2, _ := json.Marshal(body2)
	req2, _ := http.NewRequest("POST",
		"https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+":batchUpdate",
		bytes.NewReader(raw2))
	req2.Header.Set("Authorization", "Bearer "+tok)
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := http.DefaultClient.Do(req2)
	respBody2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	fmt.Printf("FIX: %d\n%s\n", resp2.StatusCode, string(respBody2))
}

func decrypt() map[string]any {
	enc, _ := os.ReadFile(encPath)
	keyFile, _ := os.ReadFile(keyPath)
	nonce := enc[:12]
	ct := enc[12:]
	block, _ := aes.NewCipher(keyFile)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, nonce, ct, nil)
	var creds map[string]any
	json.Unmarshal(pt, &creds)
	return creds
}

func getToken(creds map[string]any) string {
	rt := creds["refresh_token"].(string)
	cid := creds["client_id"].(string)
	cs, _ := creds["client_secret"].(string)
	form := "client_id=" + cid
	if cs != "" {
		form += "&client_secret=" + cs
	}
	form += "&refresh_token=" + rt + "&grant_type=refresh_token"
	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token",
		bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct{ AccessToken string `json:"access_token"` }
	json.Unmarshal(body, &out)
	return out.AccessToken
}
// restructure_feedbacks_v11_cleanup.go — Paso 1: solo limpieza
//
// Borra rows 45-50 del FEEDBACK (legacy mapping table)
// Borra todas las CF rules existentes
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
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetFB       = int64(681240158)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v11 — LIMPIEZA (DELETE rows legacy + CF rules)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// Delete rows 45-50 reverse
	fmt.Println("━━━ DELETE rows 45-50")
	for r := 50; r >= 45; r-- {
		reqs = append(reqs, deleteRow(r, sheetFB))
	}

	// Delete existing CF rules — first query how many exist
	// Per diag actual: 10 rules. Delete reverse 9 → 0
	fmt.Println("━━━ DELETE 10 existing CF rules (reverse)")
	for i := 9; i >= 0; i-- {
		reqs = append(reqs, deleteConditionalFormat(i, sheetFB))
	}

	fmt.Printf("\nTotal requests: %d\n", len(reqs))

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
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
	var result map[string]any
	json.Unmarshal(respB, &result)
	replies, _ := result["replies"].([]any)
	fmt.Printf("✅ Sent %d, replies %d\n", len(reqs), len(replies))
	fmt.Printf("🕐 Done: %s\n", time.Now().Format(time.RFC3339))
}

func deleteRow(row int, sheetID int64) map[string]any {
	ri := row - 1
	return map[string]any{
		"deleteDimension": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "ROWS",
				"startIndex": ri,
				"endIndex":   ri + 1,
			},
		},
	}
}

func deleteConditionalFormat(index int, sheetID int64) map[string]any {
	return map[string]any{
		"deleteConditionalFormatRule": map[string]any{
			"sheetId": sheetID,
			"index":   index,
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
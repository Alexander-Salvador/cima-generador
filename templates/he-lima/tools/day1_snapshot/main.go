// day1_snapshot.go — CIMA Day 1: Snapshot full TEST state to JSON.
//
// Dumps complete spreadsheet metadata + grid data to .ovav/state/test_snapshot_2026-09-28.json
// for rollback safety. Read-only operation.
//
// Authority: andres · area=platform_engineering · lead=thavren

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
)

const (
	snapPath  = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state/test_snapshot_2026-09-28.json"
	snapState = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state"
	snapEnc   = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	snapKey   = "/home/braka/.config/ovav/vault.key"
)

func main() {
	if err := os.MkdirAll(snapState, 0o755); err != nil {
		die("mkdir state: %v", err)
	}

	// Read TEST spreadsheet ID from state
	idBytes, err := os.ReadFile(snapState + "/test_sheet_id")
	if err != nil {
		die("read test_sheet_id: %v (run day1_test_setup.go first)", err)
	}
	spreadsheetID := strings.TrimSpace(string(idBytes))
	fmt.Printf("Snapshotting TEST sheet: %s\n", spreadsheetID)

	tok := getToken(decrypt())

	// Fetch full spreadsheet metadata + grid data for 00_Config and 10_Registro_HE
	fields := strings.Join([]string{
		"spreadsheetId",
		"properties",
		"sheets(properties,data(rowData(values(formattedValue,userEnteredValue)),startRow,startColumn))",
	}, ",")
	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s?fields=%s&includeGridData=true",
		spreadsheetID, fields,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		die("GET: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		die("HTTP %d: %s", resp.StatusCode, string(b))
	}

	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		die("unmarshal: %v", err)
	}

	// Pretty print
	pretty, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile(snapPath, pretty, 0o644); err != nil {
		die("write snapshot: %v", err)
	}

	// Stats
	sheets, _ := data["sheets"].([]any)
	rowCount := 0
	sheetNames := []string{}
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props, _ := sm["properties"].(map[string]any)
		title, _ := props["title"].(string)
		sheetNames = append(sheetNames, title)
		grid, _ := sm["data"].([]any)
		for _, gd := range grid {
			gdm := gd.(map[string]any)
			rd, _ := gdm["rowData"].([]any)
			rowCount += len(rd)
		}
	}

	fmt.Printf("✅ Snapshot saved: %s\n", snapPath)
	fmt.Printf("   Size: %d bytes\n", len(pretty))
	fmt.Printf("   Sheets: %d (%s)\n", len(sheetNames), strings.Join(sheetNames, ", "))
	fmt.Printf("   Total grid rows captured: %d\n", rowCount)
}

func decrypt() map[string]any {
	keyB, _ := os.ReadFile(snapKey)
	encB, _ := os.ReadFile(snapEnc)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	creds := map[string]any{}
	json.Unmarshal(pt, &creds)
	return creds
}

func getToken(c map[string]any) string {
	body, _ := json.Marshal(map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	})
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok, _ := res["access_token"].(string)
	return tok
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}

// Keep imports satisfied
var _ = io.Discard

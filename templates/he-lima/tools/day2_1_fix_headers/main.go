// day2_1_fix_headers.go — CIMA Day 2.1: Fix 10_Registro_HE headers to canonical.
//
// SCOPE (Day 2.1 SOLO):
//   - Reads TEST sheet ID from .ovav/state/test_sheet_id
//   - Fetches spreadsheet, finds current sheetId of "10_Registro_HE"
//   - batchUpdate: deleteSheet(sheetId) + addSheet with canonical REGISTRO_HEADERS
//   - NO production changes (different spreadsheet ID entirely)
//   - NO Apps Script deploy (Day 2.2)
//   - NO visual styling (column widths/freeze kept minimal/best-effort)
//
// CANONICAL SOURCE: .ovav/templates/he-lima/apps-script/03_RegistroEngine.gs
// const REGISTRO_HEADERS (lines 30-53). 22 headers, byte-for-byte.
//
// Exit codes: 0 = success (headers fixed), 1 = failure.
//
// Authority: andres · area=platform_engineering · lead=thavren · CEO=braka

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
	"reflect"
	"strings"
)

const (
	stateDir     = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state"
	encPath      = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath      = "/home/braka/.config/ovav/vault.key"
	prodSheetID  = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y" // NEVER touch
	registroID   = 1002                                              // stable sheetId for re-creation
	registroRows = 110
	registroCols = 22
)

// canonicalRegistroHeaders — VERBATIM copy of
// .ovav/templates/he-lima/apps-script/03_RegistroEngine.gs const REGISTRO_HEADERS (lines 30-53).
// Day 2.1 fix: byte-for-byte, no translation.
var canonicalRegistroHeaders = []string{
	"Fecha",
	"ISO_Week",
	"Dia_Semana",
	"Es_Feriado",
	"Tipo_Dia",
	"ID_Persona",
	"Apellidos_Nombres",
	"Turno",
	"Hora_Ingreso",
	"Hora_Salida",
	"Horas_Total",
	"Horas_Normales",
	"Horas_Extras",
	"Factor_Aplicado",
	"Horas_Ponderadas",
	"Pago_HE_PEN",
	"Observaciones",
	"Aud_Usuario",
	"Aud_Timestamp",
	"Estado",
	"Mes_Año",
	"Sync_Status",
}

func main() {
	if len(canonicalRegistroHeaders) != 22 {
		die("FATAL: canonicalRegistroHeaders must be exactly 22, got %d", len(canonicalRegistroHeaders))
	}

	tok := getToken(decrypt())
	fmt.Println("✅ OAuth token refreshed")

	// Read TEST sheet ID
	idBytes, err := os.ReadFile(stateDir + "/test_sheet_id")
	if err != nil {
		die("read test_sheet_id: %v", err)
	}
	testID := strings.TrimSpace(string(idBytes))
	if testID == "" {
		die("empty test_sheet_id")
	}
	if testID == prodSheetID {
		die("SAFETY: test_sheet_id == prodSheetID. ABORTING — would touch production.")
	}
	fmt.Printf("✅ TEST sheet ID: %s\n", testID)

	// Step 1: Fetch spreadsheet to find current sheetId of 10_Registro_HE
	fmt.Println("\n=== STEP 1: Fetch spreadsheet to find current 10_Registro_HE sheetId ===")
	currentSheetID, err := findSheetID(tok, testID, "10_Registro_HE")
	if err != nil {
		die("findSheetID: %v", err)
	}
	fmt.Printf("✅ Current 10_Registro_HE sheetId: %d\n", currentSheetID)

	// Step 2: Verify production untouched (sanity check via GET)
	fmt.Println("\n=== STEP 2: Production safety check ===")
	prodTabs, err := listSheetTitles(tok, prodSheetID)
	if err != nil {
		die("list prod sheets: %v", err)
	}
	fmt.Printf("✅ Production tabs (%d): %v\n", len(prodTabs), prodTabs)
	if len(prodTabs) != 5 {
		fmt.Printf("⚠️  WARNING: production tabs expected 5, got %d\n", len(prodTabs))
	}

	// Step 3: batchUpdate — deleteSheet + addSheet with canonical headers
	fmt.Println("\n=== STEP 3: batchUpdate (delete + re-create with canonical headers) ===")
	if err := rebuildRegistroSheet(tok, testID, currentSheetID); err != nil {
		die("rebuildRegistroSheet: %v", err)
	}
	fmt.Println("✅ batchUpdate complete")

	// Step 4: Verify byte-for-byte
	fmt.Println("\n=== STEP 4: Verify byte-for-byte match ===")
	actual, err := readHeaderRow(tok, testID, "10_Registro_HE")
	if err != nil {
		die("readHeaderRow: %v", err)
	}
	fmt.Printf("Actual headers (%d):\n", len(actual))
	for i, h := range actual {
		fmt.Printf("  [%2d] %q\n", i+1, h)
	}
	fmt.Printf("\nCanonical (%d):\n", len(canonicalRegistroHeaders))
	for i, h := range canonicalRegistroHeaders {
		fmt.Printf("  [%2d] %q\n", i+1, h)
	}

	if !reflect.DeepEqual(actual, canonicalRegistroHeaders) {
		fmt.Println("\n❌ BYTE-FOR-BYTE MISMATCH")
		for i := 0; i < len(actual) || i < len(canonicalRegistroHeaders); i++ {
			if i >= len(actual) {
				fmt.Printf("  [%2d] actual=<missing> canonical=%q\n", i+1, canonicalRegistroHeaders[i])
				continue
			}
			if i >= len(canonicalRegistroHeaders) {
				fmt.Printf("  [%2d] actual=%q canonical=<missing>\n", i+1, actual[i])
				continue
			}
			if actual[i] != canonicalRegistroHeaders[i] {
				fmt.Printf("  [%2d] DIFF actual=%q canonical=%q\n", i+1, actual[i], canonicalRegistroHeaders[i])
			}
		}
		os.Exit(1)
	}
	fmt.Printf("\n✅ BYTE-FOR-BYTE MATCH: %d/%d headers\n", len(actual), len(canonicalRegistroHeaders))

	// Step 5: Verify production still untouched (post-fix)
	fmt.Println("\n=== STEP 5: Re-verify production safety (post-fix) ===")
	prodTabsAfter, err := listSheetTitles(tok, prodSheetID)
	if err != nil {
		die("list prod sheets (post): %v", err)
	}
	if !reflect.DeepEqual(prodTabs, prodTabsAfter) {
		fmt.Printf("❌ PRODUCTION DRIFT DETECTED: before=%v after=%v\n", prodTabs, prodTabsAfter)
		os.Exit(1)
	}
	fmt.Println("✅ Production tabs unchanged (post-fix)")

	fmt.Println("\n=== DAY 2.1 COMPLETE ===")
	fmt.Println("Run tools/day1_verify/main.go for full gate verification.")
}

// ─── API helpers ───

func findSheetID(tok, spreadsheetID, title string) (int64, error) {
	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(sheetId,title))",
		spreadsheetID,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GET: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets, _ := out["sheets"].([]any)
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props := sm["properties"].(map[string]any)
		if stringOf(props["title"]) == title {
			f, _ := props["sheetId"].(float64)
			return int64(f), nil
		}
	}
	return 0, fmt.Errorf("sheet %q not found in spreadsheet %s", title, spreadsheetID)
}

func listSheetTitles(tok, spreadsheetID string) ([]string, error) {
	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(title))",
		spreadsheetID,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	out := map[string]any{}
	json.Unmarshal(b, &out)
	sheets, _ := out["sheets"].([]any)
	titles := []string{}
	for _, sh := range sheets {
		sm := sh.(map[string]any)
		props := sm["properties"].(map[string]any)
		titles = append(titles, stringOf(props["title"]))
	}
	return titles, nil
}

func readHeaderRow(tok, spreadsheetID, sheetTitle string) ([]string, error) {
	// Get sheetId first
	sheetID, err := findSheetID(tok, spreadsheetID, sheetTitle)
	if err != nil {
		return nil, err
	}
	// Read A1:V1
	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s!A1:%s%d",
		spreadsheetID, sheetTitle, colLetter(registroCols), 1,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	out := map[string]any{}
	json.Unmarshal(b, &out)
	vr, _ := out["values"].([]any)
	if len(vr) == 0 {
		return nil, fmt.Errorf("no values returned for %s", sheetTitle)
	}
	row, _ := vr[0].([]any)
	out2 := make([]string, 0, len(row))
	for _, c := range row {
		if s, ok := c.(string); ok {
			out2 = append(out2, s)
		} else {
			out2 = append(out2, fmt.Sprintf("%v", c))
		}
	}
	_ = sheetID
	return out2, nil
}

func rebuildRegistroSheet(tok, spreadsheetID string, currentSheetID int64) error {
	reqs := []any{}

	// 1. Delete existing 10_Registro_HE
	reqs = append(reqs, map[string]any{
		"deleteSheet": map[string]any{
			"sheetId": currentSheetID,
		},
	})

	// 2. Re-add with canonical headers, same properties as day1_setup
	reqs = append(reqs, map[string]any{
		"addSheet": map[string]any{
			"properties": map[string]any{
				"sheetId": registroID,
				"title":   "10_Registro_HE",
				"index":   1,
				"tabColor": map[string]any{
					"red":   0.961,
					"green": 0.620,
					"blue":  0.043,
				}, // #f59e0b orange
				"gridProperties": map[string]any{
					"rowCount":           registroRows,
					"columnCount":        registroCols,
					"frozenRowCount":     1,
					"frozenColumnCount":  5,
				},
			},
		},
	})

	// 3. Update properties to hide
	reqs = append(reqs, map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": registroID,
				"hidden":   true,
			},
			"fields": "hidden",
		},
	})

	// 4. Write canonical headers A1:V1
	hdrRow := []any{}
	for _, h := range canonicalRegistroHeaders {
		hdrRow = append(hdrRow, map[string]any{
			"userEnteredValue": map[string]any{"stringValue": h},
			"userEnteredFormat": map[string]any{
				"backgroundColor": map[string]any{
					"red":   0.118,
					"green": 0.227,
					"blue":  0.541,
				},
				"textFormat": map[string]any{
					"foregroundColor": map[string]any{
						"red": 1.0, "green": 1.0, "blue": 1.0,
					},
					"bold":     true,
					"fontSize": 10,
				},
				"horizontalAlignment": "CENTER",
				"verticalAlignment":   "MIDDLE",
				"wrapStrategy":        "WRAP",
			},
		})
	}
	cells := []any{}
	for _, c := range hdrRow {
		cells = append(cells, c)
	}
	row := []any{
		map[string]any{"values": cells},
	}
	reqs = append(reqs, map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          registroID,
				"startRowIndex":    0,
				"endRowIndex":      1,
				"startColumnIndex": 0,
				"endColumnIndex":   registroCols,
			},
			"rows":   row,
			"fields": "userEnteredValue,userEnteredFormat",
		},
	})

	// Send batchUpdate
	body, _ := json.Marshal(map[string]any{
		"requests":                   reqs,
		"includeSpreadsheetInResponse": false,
	})
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// ─── Auth helpers ───

func decrypt() map[string]any {
	keyB, err := os.ReadFile(keyPath)
	if err != nil {
		die("read key: %v", err)
	}
	encB, err := os.ReadFile(encPath)
	if err != nil {
		die("read enc: %v", err)
	}
	block, err := aes.NewCipher(keyB)
	if err != nil {
		die("cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		die("gcm: %v", err)
	}
	pt, err := gcm.Open(nil, encB[:12], encB[12:], nil)
	if err != nil {
		die("decrypt: %v", err)
	}
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
	r, err := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	if err != nil {
		die("token POST: %v", err)
	}
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok, _ := res["access_token"].(string)
	if tok == "" {
		die("no access_token: %v", res)
	}
	return tok
}

// ─── Misc helpers ───

func stringOf(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func colLetter(n int) string {
	// n=1 → A, n=22 → V, n=27 → AA
	s := ""
	for n > 0 {
		n--
		s = string(rune('A'+(n%26))) + s
		n /= 26
	}
	return s
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}

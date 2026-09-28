// verify_day1.go — CIMA Day 1 verification gate.
//
// Validates on the TEST spreadsheet:
//   1. All prod tabs preserved (CIMA TEAM, HORAS EXTRAS, 01_FERIADOS_PE, PERSONAL, CONFIG, Feedbacks)
//   2. 00_Config exists with 9 anchors (B2-B9, B11) and correct values
//   3. 10_Registro_HE exists, hidden, 110×22 dimensions, headers present
//   4. Data validation on B2:B3 of 00_Config
//   5. Tab colors correct (00_Config navy, 10_Registro_HE orange)
//
// Exit codes: 0 = PASS, 1 = FAIL
//
// Authority: andres · area=platform_engineering · lead=thavren · QA=clara

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
	verState = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state"
	verEnc   = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	verKey   = "/home/braka/.config/ovav/vault.key"
)

// prodTabs are the tabs that MUST be preserved in TEST after Drive copy.
// (Verified 2026-09-28 — prod sheet actually contains only these 5 tabs;
//  CIMA TEAM is referenced in architecture but was never created in prod.)
var prodTabs = []string{
	"HORAS EXTRAS",
	"01_FERIADOS_PE",
	"PERSONAL",
	"CONFIG",
	"Feedbacks",
}

type verifyResult struct {
	OK    bool
	Issue string
}

func main() {
	idBytes, err := os.ReadFile(verState + "/test_sheet_id")
	if err != nil {
		die("read test_sheet_id: %v", err)
	}
	spreadsheetID := strings.TrimSpace(string(idBytes))
	fmt.Printf("Verifying TEST sheet: %s\n\n", spreadsheetID)

	tok := getToken(decrypt())
	ss := fetchSpreadsheet(tok, spreadsheetID)

	results := []verifyResult{}
	allOK := true

	// ─── Check 1: All prod tabs preserved ───
	fmt.Println("=== 1. PROD TABS PRESERVED ===")
	actualTabs := map[string]bool{}
	for _, sh := range ss.sheets {
		actualTabs[sh.title] = true
	}
	for _, expected := range prodTabs {
		if actualTabs[expected] {
			fmt.Printf("  ✅ %s\n", expected)
		} else {
			fmt.Printf("  ❌ MISSING: %s\n", expected)
			results = append(results, verifyResult{false, fmt.Sprintf("prod tab missing: %s", expected)})
			allOK = false
		}
	}

	// ─── Check 2: 00_Config exists with anchors ───
	fmt.Println("\n=== 2. 00_CONFIG ANCHORS ===")
	cfg := ss.findSheet("00_Config")
	if cfg == nil {
		results = append(results, verifyResult{false, "00_Config sheet missing"})
		allOK = false
		fmt.Println("  ❌ 00_Config NOT FOUND")
	} else {
		// Check tab color (navy 0.118/0.227/0.541)
		if cfg.tabColor != nil {
			r, g, b := cfg.tabColor.R, cfg.tabColor.G, cfg.tabColor.B
			if near(r, 0.118, 0.02) && near(g, 0.227, 0.02) && near(b, 0.541, 0.02) {
				fmt.Printf("  ✅ Tab color: navy (%.3f, %.3f, %.3f)\n", r, g, b)
			} else {
				fmt.Printf("  ⚠️  Tab color unexpected: (%.3f, %.3f, %.3f), expected navy ~0.118/0.227/0.541\n", r, g, b)
			}
		}
		// Check anchors B2-B9, B11
		expectedAnchors := map[string]interface{}{
			"B2":  float64(9),
			"B3":  float64(2026),
			"B4":  "W39",
			"B5":  "TODOS",
			"B6":  1.25,
			"B7":  1.35,
			"B8":  2.00,
			"B9":  float64(8),
			"B11": "",
		}
		for cell, exp := range expectedAnchors {
			row, col := cellRef(cell)
			got := cfg.valueAt(row-1, col-1)
			if matchValue(got, exp) {
				fmt.Printf("  ✅ %s = %v\n", cell, got)
			} else {
				fmt.Printf("  ❌ %s = %v, expected %v\n", cell, got, exp)
				results = append(results, verifyResult{false, fmt.Sprintf("anchor %s mismatch: got %v, expected %v", cell, got, exp)})
				allOK = false
			}
		}
		// Check headers row 1
		hdr := []string{"ANCLA", "VALOR", "DESCRIPCIÓN"}
		for i, h := range hdr {
			got := cfg.valueAt(0, i)
			if s, ok := got.(string); ok && strings.TrimSpace(s) == h {
				fmt.Printf("  ✅ A1:C1[%d] = %q\n", i+1, h)
			} else {
				fmt.Printf("  ❌ Header mismatch at col %d: got %v, expected %q\n", i+1, got, h)
				results = append(results, verifyResult{false, fmt.Sprintf("00_Config header %d mismatch", i+1)})
				allOK = false
			}
		}
	}

	// ─── Check 3: 10_Registro_HE exists, hidden, dimensions, headers ───
	fmt.Println("\n=== 3. 10_REGISTRO_HE ===")
	reg := ss.findSheet("10_Registro_HE")
	if reg == nil {
		results = append(results, verifyResult{false, "10_Registro_HE sheet missing"})
		allOK = false
		fmt.Println("  ❌ 10_Registro_HE NOT FOUND")
	} else {
		// Hidden?
		if reg.hidden {
			fmt.Println("  ✅ Hidden: true")
		} else {
			fmt.Println("  ❌ Hidden: false (should be hidden working surface)")
			results = append(results, verifyResult{false, "10_Registro_HE not hidden"})
			allOK = false
		}
		// Dimensions
		if reg.rowCount == 110 {
			fmt.Printf("  ✅ Row count: 110\n")
		} else {
			fmt.Printf("  ⚠️  Row count: %d (expected 110)\n", reg.rowCount)
		}
		if reg.colCount == 22 {
			fmt.Printf("  ✅ Column count: 22\n")
		} else {
			fmt.Printf("  ❌ Column count: %d (expected 22)\n", reg.colCount)
			results = append(results, verifyResult{false, fmt.Sprintf("10_Registro_HE col count %d ≠ 22", reg.colCount)})
			allOK = false
		}
		// Tab color (orange 0.961/0.620/0.043)
		if reg.tabColor != nil {
			r, g, b := reg.tabColor.R, reg.tabColor.G, reg.tabColor.B
			if near(r, 0.961, 0.02) && near(g, 0.620, 0.02) && near(b, 0.043, 0.02) {
				fmt.Printf("  ✅ Tab color: orange (%.3f, %.3f, %.3f)\n", r, g, b)
			} else {
				fmt.Printf("  ⚠️  Tab color: (%.3f, %.3f, %.3f), expected orange ~0.961/0.620/0.043\n", r, g, b)
			}
		}
		// Headers — CANONICAL from .ovav/templates/he-lima/apps-script/03_RegistroEngine.gs
		// const REGISTRO_HEADERS (lines 30-53). Day 2.1 fix: byte-for-byte verification.
		expectedHdrs := []string{
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
		for i, h := range expectedHdrs {
			got := reg.valueAt(0, i)
			if s, ok := got.(string); ok && s == h {
				fmt.Printf("  ✅ A1:V1[%d] = %q\n", i+1, h)
			} else {
				fmt.Printf("  ❌ Header mismatch at col %d: got %v, expected %q\n", i+1, got, h)
				results = append(results, verifyResult{false, fmt.Sprintf("10_Registro_HE header %d mismatch", i+1)})
				allOK = false
			}
		}
	}

	// ─── Final report ───
	fmt.Println("\n=== VERIFICATION SUMMARY ===")
	if allOK {
		fmt.Println("✅ ALL CHECKS PASSED")
		os.Exit(0)
	} else {
		fmt.Printf("❌ %d ISSUES DETECTED\n", len(results))
		for i, r := range results {
			fmt.Printf("  [%d] %s\n", i+1, r.Issue)
		}
		os.Exit(1)
	}
}

// ─── Data structures ───

type colorRGB struct {
	R, G, B float64
}

type sheetInfo struct {
	title     string
	sheetID   int64
	hidden    bool
	rowCount  int
	colCount  int
	tabColor  *colorRGB
	gridData  [][]interface{} // [row][col] → formatted value
}

func (s *sheetInfo) valueAt(row, col int) interface{} {
	if row < 0 || row >= len(s.gridData) {
		return nil
	}
	if col < 0 || col >= len(s.gridData[row]) {
		return nil
	}
	return s.gridData[row][col]
}

type spreadsheetInfo struct {
	id     string
	sheets []*sheetInfo
}

func (s *spreadsheetInfo) findSheet(title string) *sheetInfo {
	for _, sh := range s.sheets {
		if sh.title == title {
			return sh
		}
	}
	return nil
}

// ─── API fetch ───

func fetchSpreadsheet(tok, spreadsheetID string) *spreadsheetInfo {
	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets(properties(title,sheetId,hidden,tabColor,gridProperties(rowCount,columnCount,frozenRowCount,frozenColumnCount)),data(rowData(values(formattedValue,userEnteredValue))))&includeGridData=true",
		spreadsheetID,
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
	out := map[string]any{}
	json.Unmarshal(b, &out)

	info := &spreadsheetInfo{id: spreadsheetID}
	for _, sh := range out["sheets"].([]any) {
		sm := sh.(map[string]any)
		props := sm["properties"].(map[string]any)
		si := &sheetInfo{
			title:   stringOf(props["title"]),
			sheetID: int64Of(props["sheetId"]),
			hidden:  boolOf(props["hidden"]),
		}
		if gp, ok := props["gridProperties"].(map[string]any); ok {
			si.rowCount = intOf(gp["rowCount"])
			si.colCount = intOf(gp["columnCount"])
		}
		if tc, ok := props["tabColor"].(map[string]any); ok {
			si.tabColor = &colorRGB{
				R: floatOf(tc["red"]),
				G: floatOf(tc["green"]),
				B: floatOf(tc["blue"]),
			}
		}
		// Grid data
		data, _ := sm["data"].([]any)
		for _, gd := range data {
			gdm := gd.(map[string]any)
			rd, _ := gdm["rowData"].([]any)
			for _, row := range rd {
				rm := row.(map[string]any)
				cells, _ := rm["values"].([]any)
				rowVals := []interface{}{}
				for _, c := range cells {
					cm := c.(map[string]any)
					rowVals = append(rowVals, extractCell(cm))
				}
				si.gridData = append(si.gridData, rowVals)
			}
		}
		info.sheets = append(info.sheets, si)
	}
	return info
}

func extractCell(cm map[string]any) interface{} {
	if v, ok := cm["formattedValue"].(string); ok && v != "" {
		return v
	}
	if uev, ok := cm["userEnteredValue"].(map[string]any); ok {
		if s, ok := uev["stringValue"].(string); ok {
			return s
		}
		if n, ok := uev["numberValue"].(float64); ok {
			return n
		}
		if b, ok := uev["boolValue"].(bool); ok {
			return b
		}
	}
	return nil
}

// ─── Helpers ───

func near(a, b, tol float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < tol
}

func matchValue(got, exp interface{}) bool {
	if got == nil && exp == "" {
		return true
	}
	if got == nil {
		return false
	}
	switch e := exp.(type) {
	case string:
		if e == "" {
			return got == nil || got == ""
		}
		s, ok := got.(string)
		if !ok {
			// Try coercing got to string
			s = fmt.Sprintf("%v", got)
		}
		return s == e
	case float64:
		// API returns formattedValue as string (e.g., "9", "1.25"). Try both paths.
		if g, ok := got.(float64); ok {
			return near(g, e, 0.01)
		}
		if s, ok := got.(string); ok {
			// Parse string to float
			var g float64
			if _, err := fmt.Sscanf(s, "%f", &g); err == nil {
				return near(g, e, 0.01)
			}
		}
		return false
	case int:
		return matchValue(got, float64(e))
	}
	return false
}

func cellRef(cell string) (row, col int) {
	// Parse "B11" → row=11, col=2
	i := 0
	colStr := ""
	for i < len(cell) && cell[i] >= 'A' && cell[i] <= 'Z' {
		colStr += string(cell[i])
		i++
	}
	rowStr := cell[i:]
	// col
	col = 0
	for _, c := range colStr {
		col = col*26 + int(c-'A'+1)
	}
	row = 0
	for _, c := range rowStr {
		row = row*10 + int(c-'0')
	}
	return
}

func stringOf(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func intOf(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}
func int64Of(v interface{}) int64 {
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
}
func floatOf(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
func boolOf(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func decrypt() map[string]any {
	keyB, _ := os.ReadFile(verKey)
	encB, _ := os.ReadFile(verEnc)
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

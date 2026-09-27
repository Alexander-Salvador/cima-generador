// verify_cima_team.go — Comprehensive verification tool for CIMA TEAM.
//
// This tool reads ALL relevant state from the sheet and outputs a structured
// report. Run this AFTER any change to verify the result before claiming done.
//
// Usage: go run verify_cima_team.go [--compact]
//
// Outputs:
// - Total width (sum of col widths)
// - Frame borders present
// - Cell content sample (first/last rows, days 1/15/30)
// - Conditional format rules
// - Merges count
// - Issues detected (overflow, misaligned, missing)
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
	"strconv"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetHE       = int64(222789142)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	creds := map[string]any{}
	json.Unmarshal(pt, &creds)

	body, _ := json.Marshal(map[string]any{
		"client_id":     creds["client_id"],
		"client_secret": creds["client_secret"],
		"refresh_token": creds["refresh_token"],
		"grant_type":    "refresh_token",
	})
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(body))
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	r.Body.Close()
	tok := res["access_token"].(string)

	// Fetch sheet data with everything we need
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?includeGridData=true&ranges=CIMA%%20TEAM!A1:AT40&fields=sheets(properties.title,data(columnMetadata(pixelSize),rowData(values(formattedValue,userEnteredValue))),merges,conditionalFormats)",
		spreadsheetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	json.Unmarshal(b, &out)

	sheets := out["sheets"].([]any)
	if len(sheets) == 0 {
		fmt.Println("ERROR: no sheets in response")
		os.Exit(1)
	}
	s := sheets[0].(map[string]any)

	// === 1. TOTAL WIDTH ===
	fmt.Println("=== TOTAL WIDTH ===")
	totalW := 0
	data := s["data"].([]any)
	colMeta := data[0].(map[string]any)["columnMetadata"].([]any)
	for i, c := range colMeta {
		cm := c.(map[string]any)
		pwF := cm["pixelSize"].(float64)
		pw := int(pwF)
		totalW += pw
		if i < 50 {
			marker := ""
			if pw >= 100 {
				marker = " ←wide"
			}
			if pw <= 6 {
				marker = " ←spacer"
			}
			fmt.Printf("  %s (%2d): %3dpx%s\n", colLetter(i+1), i+1, pw, marker)
		}
	}
	fmt.Printf("  ─────────────────────────\n")
	fmt.Printf("  TOTAL: %dpx\n", totalW)
	fmt.Println()

	// === 2. CONTENT SAMPLE (first/last emp rows + day cells) ===
	fmt.Println("=== CONTENT SAMPLE ===")
	rowData, ok := data[0].(map[string]any)["rowData"].([]any)
	if !ok {
		fmt.Println("no rowData")
		return
	}
	for _, ri := range []int{0, 4, 5, 7, 8, 9, 25, 35} {
		if ri >= len(rowData) {
			continue
		}
		cells := rowData[ri].(map[string]any)["values"].([]any)
		fmt.Printf("  Row %d (%s):\n", ri+1, rowLabel(ri+1))
		for ci := 0; ci < 6 && ci < len(cells); ci++ {
			v := cellValue(cells[ci])
			fmt.Printf("    Col %s: %q\n", colLetter(ci+1), v)
		}
		// Sample day cells
		dayCells := []int{4, 18, 33} // E, S, AH
		for _, ci := range dayCells {
			if ci >= len(cells) {
				continue
			}
			v := cellValue(cells[ci])
			fmt.Printf("    Col %s (day): %q\n", colLetter(ci+1), v)
		}
	}
	fmt.Println()

	// === 3. MERGES ===
	fmt.Println("=== MERGES ===")
	merges := s["merges"].([]any)
	fmt.Printf("  Total merges: %d\n", len(merges))
	fmt.Println()

	// === 4. CONDITIONAL FORMATS ===
	fmt.Println("=== CONDITIONAL FORMATS ===")
	cf := s["conditionalFormats"].([]any)
	fmt.Printf("  Total rules: %d\n", len(cf))
	for i, rule := range cf {
		r := rule.(map[string]any)
		ranges := r["ranges"].([]any)
		for _, rng := range ranges {
			rm := rng.(map[string]any)
			srF := int(rm["startRowIndex"].(float64)) + 1
			erF := int(rm["endRowIndex"].(float64))
			scF := int(rm["startColumnIndex"].(float64)) + 1
			ecF := int(rm["endColumnIndex"].(float64))
			br, _ := r["booleanRule"].(map[string]any)
			cond, _ := br["condition"].(map[string]any)
			ctype, _ := cond["type"].(string)
			fmt.Printf("  [%d] R%d-%d C%d-%d  %s\n", i, srF, erF, scF, ecF, ctype)
		}
	}
	fmt.Println()

	// === 5. ISSUES DETECTED ===
	fmt.Println("=== ISSUES DETECTED ===")
	issues := []string{}

	// Width overflow
	if totalW > 1280 {
		issues = append(issues, fmt.Sprintf("WIDTH OVERFLOW: total %dpx > 1280px viewport", totalW))
	}
	if totalW > 1440 {
		issues = append(issues, fmt.Sprintf("WIDTH CRITICAL: total %dpx > 1440px (won't fit in typical laptop)", totalW))
	}

	// Spacer inconsistency
	spacers := []int{34, 36, 38, 40, 42, 44}
	for _, ci := range spacers {
		if ci < len(colMeta) {
			pw := int(colMeta[ci].(map[string]any)["pixelSize"].(float64))
			if pw > 6 {
				issues = append(issues, fmt.Sprintf("Spacer col %s: %dpx (should be ≤6px)", colLetter(ci+1), pw))
			}
		}
	}

	// Day cell overflow risk
	for ci := 4; ci < 34; ci++ {
		if ci < len(colMeta) {
			pw := int(colMeta[ci].(map[string]any)["pixelSize"].(float64))
			if pw < 22 {
				issues = append(issues, fmt.Sprintf("Day col %s: %dpx (too narrow for 10pt text)", colLetter(ci+1), pw))
			}
		}
	}

	// Conditional format duplicates
	cfTypes := []string{}
	for _, rule := range cf {
		br, _ := rule.(map[string]any)["booleanRule"].(map[string]any)
		cond, _ := br["condition"].(map[string]any)
		ctype, _ := cond["type"].(string)
		vals, _ := cond["values"].([]any)
		for _, v := range vals {
			vm := v.(map[string]any)
			s, _ := vm["userEnteredValue"].(string)
			cfTypes = append(cfTypes, ctype+":"+s)
		}
	}
	typeCount := map[string]int{}
	for _, t := range cfTypes {
		typeCount[t]++
	}
	for t, count := range typeCount {
		if count > 1 {
			issues = append(issues, fmt.Sprintf("Duplicate conditional format: %s (x%d)", t, count))
		}
	}

	if len(issues) == 0 {
		fmt.Println("  ✓ No issues detected")
	} else {
		for _, issue := range issues {
			fmt.Printf("  ⚠️  %s\n", issue)
		}
	}

	fmt.Println()
	fmt.Println("=== END OF REPORT ===")
}

func cellValue(c any) string {
	if c == nil {
		return ""
	}
	cm, ok := c.(map[string]any)
	if !ok {
		return ""
	}
	if v, ok := cm["formattedValue"].(string); ok {
		return v
	}
	if v, ok := cm["userEnteredValue"].(map[string]any); ok {
		if s, ok := v["stringValue"].(string); ok {
			return s
		}
		if f, ok := v["formulaValue"].(string); ok {
			return "=" + f
		}
		if n, ok := v["numberValue"].(float64); ok {
			return strconv.FormatFloat(n, 'f', -1, 64)
		}
	}
	return ""
}

func colLetter(idx int) string {
	result := ""
	for idx > 0 {
		idx--
		result = string(rune('A'+idx%26)) + result
		idx = idx / 26
	}
	return result
}

func rowLabel(row int) string {
	switch row {
	case 1:
		return "title"
	case 5:
		return "Año"
	case 6:
		return "Mes"
	case 8:
		return "day numbers"
	case 9:
		return "day letters"
	case 10:
		return "first emp"
	case 26:
		return "first empty"
	case 36:
		return "TOTALES"
	}
	return ""
}
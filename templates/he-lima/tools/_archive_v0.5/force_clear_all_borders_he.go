// force_clear_all_borders_he.go — Aggressive border cleanup.
//
// Issue: updateBorders with nil fields doesn't clear borders in Google Sheets.
// The API requires explicit {"style": "NONE"} to remove borders.
//
// This script:
// 1. Sets EVERY border on EVERY cell in A1:AT100 to {style: NONE} explicitly
// 2. Then re-applies the 2 clean frames
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

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1 — Force-clear ALL borders by setting them to {style: NONE}
	//          Apply per-row to avoid massive single request.
	// ═══════════════════════════════════════════════════════════════════
	for row := 0; row < 35; row++ {
		reqs = append(reqs, map[string]any{
			"updateBorders": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetHE,
					"startRowIndex":    row,
					"endRowIndex":      row + 1,
					"startColumnIndex": 0,
					"endColumnIndex":   46,
				},
				"top":             map[string]any{"style": "NONE"},
				"bottom":          map[string]any{"style": "NONE"},
				"left":            map[string]any{"style": "NONE"},
				"right":           map[string]any{"style": "NONE"},
				"innerHorizontal": map[string]any{"style": "NONE"},
				"innerVertical":   map[string]any{"style": "NONE"},
			},
		})
	}

	// Send batch (clear)
	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR clear:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respB, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR clear", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	fmt.Printf("✅ Cleared %d rows of borders\n", len(reqs))

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2 — Apply ONLY 2 clean frames (NO internal accents, NO dividers)
	//          FRAME 1 HEADER: A1:AT7
	//          FRAME 2 DATA:   A8:AT32
	// ═══════════════════════════════════════════════════════════════════
	frameReqs := []any{}

	// FRAME 1 — HEADER
	frameReqs = append(frameReqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    0,
				"endRowIndex":      7,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderPrimary(), 3),
			"bottom": border("SOLID_THICK", borderPrimary(), 3),
			"left":   border("SOLID_THICK", borderPrimary(), 3),
			"right":  border("SOLID_THICK", borderPrimary(), 3),
		},
	})

	// FRAME 2 — DATA
	frameReqs = append(frameReqs, map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetHE,
				"startRowIndex":    7,
				"endRowIndex":      32,
				"startColumnIndex": 0,
				"endColumnIndex":   46,
			},
			"top":    border("SOLID_THICK", borderPrimary(), 3),
			"bottom": border("SOLID_THICK", borderPrimary(), 3),
			"left":   border("SOLID_THICK", borderPrimary(), 3),
			"right":  border("SOLID_THICK", borderPrimary(), 3),
		},
	})

	// Send batch (frames)
	body2 := map[string]any{"requests": frameReqs}
	bb2, _ := json.Marshal(body2)

	req2, _ := http.NewRequest("POST", url, bytes.NewReader(bb2))
	req2.Header.Set("Authorization", "Bearer "+tok)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println("ERROR frames:", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	respB2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode >= 300 {
		fmt.Println("API ERROR frames", resp2.StatusCode, ":", string(respB2))
		os.Exit(1)
	}
	fmt.Printf("✅ Applied 2 frames (SOLID_THICK navy 3px)\n")
	fmt.Println("Result: ONE clean header frame + ONE clean data frame, no leftover borders")
}

func borderPrimary() map[string]any {
	return map[string]any{
		"red":   0.101960786,
		"green": 0.227450982,
		"blue":  0.360784314,
	}
}

func border(style string, c map[string]any, w int) map[string]any {
	return map[string]any{"style": style, "color": c, "width": w}
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
	return res["access_token"].(string)
}
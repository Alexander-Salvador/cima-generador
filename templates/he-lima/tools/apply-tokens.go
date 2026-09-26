// Package main — apply-tokens: bulk-style tabs of CIMA HE LIMA sheet
// using canonical design tokens from .ovav/templates/he-lima/tokens.json.
//
// CRIT-009 baseline: one canonical source of truth for colors/typography.
// CRIT-002 applied: each tab is applied + verified separately, no batching.
//
// Usage:
//   go run main.go                    # apply all 5 tabs (with HE row-2 deletion)
//   go run main.go -tab=PERSONAL      # apply only one tab
//   go run main.go -dry-run           # show what would change, no writes
//   go run main.go -skip-he           # apply 4 simple tabs, skip HE (safer)
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
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	tokensPath    = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/templates/he-lima/tokens.json"
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

// ── Token resolution ──────────────────────────────────────────────────────────

func loadTokens() (map[string]string, error) {
	raw, err := os.ReadFile(tokensPath)
	if err != nil {
		return nil, fmt.Errorf("read tokens: %w", err)
	}
	var t struct {
		Color map[string]map[string]struct {
			Hex string `json:"hex"`
		} `json:"color"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("parse tokens: %w", err)
	}
	resolved := map[string]string{}
	for cat, names := range t.Color {
		for name, v := range names {
			key := cat + "." + name
			resolved[key] = strings.ToUpper(strings.TrimPrefix(v.Hex, "#"))
		}
	}
	return resolved, nil
}

func resolve(token string, m map[string]string) (string, error) {
	token = strings.TrimPrefix(token, "token:")
	if token == "" {
		return "", nil
	}
	if hex, ok := m[token]; ok {
		return hex, nil
	}
	if len(token) == 6 || len(token) == 8 {
		return strings.ToUpper(token), nil
	}
	return "", fmt.Errorf("unknown token: %s", token)
}

// ── OAuth + Sheets API ──────────────────────────────────────────────────────

func decryptCreds() (map[string]any, error) {
	enc, err := os.ReadFile(encPath)
	if err != nil {
		return nil, err
	}
	keyFile, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	if len(enc) < 28 {
		return nil, fmt.Errorf("encrypted blob too short: %d", len(enc))
	}
	nonce := enc[:12]
	ct := enc[12:]
	block, err := aes.NewCipher(keyFile)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	var creds map[string]any
	if err := json.Unmarshal(pt, &creds); err != nil {
		return nil, err
	}
	return creds, nil
}

func getAccessToken(creds map[string]any) (string, error) {
	rt, _ := creds["refresh_token"].(string)
	cid, _ := creds["client_id"].(string)
	cs, _ := creds["client_secret"].(string)
	if rt == "" || cid == "" {
		return "", fmt.Errorf("missing refresh_token or client_id")
	}
	form := "client_id=" + cid
	if cs != "" {
		form += "&client_secret=" + cs
	}
	form += "&refresh_token=" + rt + "&grant_type=refresh_token"
	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token",
		bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	json.Unmarshal(body, &out)
	if out.AccessToken == "" {
		return "", fmt.Errorf("token exchange failed: %s · body: %s", out.Error, string(body))
	}
	return out.AccessToken, nil
}

type sheetsAPI struct {
	tok    string
	client *http.Client
}

func newSheetsAPI() (*sheetsAPI, error) {
	creds, err := decryptCreds()
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	tok, err := getAccessToken(creds)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	return &sheetsAPI{tok: tok, client: &http.Client{Timeout: 60 * time.Second}}, nil
}

func (a *sheetsAPI) get(path string, out any) error {
	req, _ := http.NewRequest("GET", "https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+path, nil)
	req.Header.Set("Authorization", "Bearer "+a.tok)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GET %s · %d · %s", path, resp.StatusCode, string(respBody))
	}
	return json.Unmarshal(respBody, out)
}

func (a *sheetsAPI) post(path string, payload any, out any) error {
	raw, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://sheets.googleapis.com/v4/spreadsheets/"+spreadsheetID+path,
		bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+a.tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("POST %s · %d · %s", path, resp.StatusCode, string(respBody))
	}
	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

// ── Sheet metadata ──────────────────────────────────────────────────────────

type SheetMeta struct {
	SheetID int64  `json:"sheetId"`
	Title   string `json:"title"`
}

func (a *sheetsAPI) listSheets() ([]SheetMeta, error) {
	var out struct {
		Sheets []struct {
			Properties SheetMeta `json:"properties"`
		} `json:"sheets"`
	}
	if err := a.get("?fields=sheets(properties)", &out); err != nil {
		return nil, err
	}
	result := []SheetMeta{}
	for _, s := range out.Sheets {
		result = append(result, s.Properties)
	}
	return result, nil
}

// ── A1 → grid range conversion ──────────────────────────────────────────────

func colLetterToIndex(letters string) int {
	n := 0
	for _, c := range letters {
		n = n*26 + int(c-'A'+1)
	}
	return n - 1
}

func parseA1(a1 string) (row, col int, err error) {
	for i, c := range a1 {
		if c >= '0' && c <= '9' {
			colStr := a1[:i]
			rowStr := a1[i:]
			col = colLetterToIndex(colStr)
			r, _ := strconv.Atoi(rowStr)
			row = r - 1
			return
		}
	}
	return 0, 0, fmt.Errorf("invalid A1: %s", a1)
}

func parseRange(rng string) (startRow, startCol, endRow, endCol int, err error) {
	parts := strings.SplitN(rng, ":", 2)
	startRow, startCol, err = parseA1(parts[0])
	if err != nil {
		return
	}
	if len(parts) == 1 {
		endRow, endCol = startRow, startCol
		return
	}
	endRow, endCol, err = parseA1(parts[1])
	return
}

// ── Style operations ────────────────────────────────────────────────────────

type StyleOp struct {
	Range    string
	Bg       string
	Fg       string
	Bold     bool
	Italic   bool
	FontSize int
	Align    string
}

type MergeOp struct {
	Range string
}

type DeleteRowOp struct {
	StartRow int // 0-based
	EndRow   int // exclusive
}

type NumberFormatOp struct {
	Range  string
	Format string
	Type   string
}

type TabSpec struct {
	Title         string
	DeleteRows    []DeleteRowOp
	Styles        []StyleOp
	Merges        []MergeOp
	Unmerges      []MergeOp
	NumberFormats []NumberFormatOp
}

func hex(token string, m map[string]string) string {
	h, err := resolve(token, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  token resolve failed: %s (%v)\n", token, err)
		return "FFFFFF"
	}
	if os.Getenv("DEBUG_TOKEN") != "" {
		fmt.Fprintf(os.Stderr, "    token %s → %s\n", token, h)
	}
	return h
}

func hexToFloat(hex string, idx int) float64 {
	if len(hex) < 6 {
		return 0
	}
	hex = strings.ToUpper(hex)
	v, err := strconv.ParseUint(hex[2*idx:2*idx+2], 16, 8)
	if err != nil {
		if os.Getenv("DEBUG_TOKEN") != "" {
			fmt.Fprintf(os.Stderr, "    ParseUint err hex=%q idx=%d: %v\n", hex, idx, err)
		}
		return 0
	}
	return float64(v) / 255.0
}

// ── batchUpdate builders ────────────────────────────────────────────────────

func buildRequests(spec TabSpec, sheetID int64, m map[string]string) []any {
	if os.Getenv("DEBUG_REQ") != "" {
		fmt.Fprintf(os.Stderr, "DEBUG_REQ enabled\n")
	}
	reqs := []any{}

	// 1. Deletes (must run first — shifts all subsequent rows)
	for _, op := range spec.DeleteRows {
		reqs = append(reqs, map[string]any{
			"deleteDimension": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetID,
					"dimension":  "ROWS",
					"startIndex": op.StartRow,
					"endIndex":   op.EndRow,
				},
			},
		})
	}

	// 2. Unmerges
	for _, op := range spec.Unmerges {
		sr, sc, er, ec, err := parseRange(op.Range)
		if err != nil {
			continue
		}
		reqs = append(reqs, map[string]any{
			"unmergeCells": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetID,
					"startRowIndex":    sr,
					"endRowIndex":      er + 1,
					"startColumnIndex": sc,
					"endColumnIndex":   ec + 1,
				},
			},
		})
	}

	// 3. Merges
	for _, op := range spec.Merges {
		sr, sc, er, ec, err := parseRange(op.Range)
		if err != nil {
			continue
		}
		reqs = append(reqs, map[string]any{
			"mergeCells": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetID,
					"startRowIndex":    sr,
					"endRowIndex":      er + 1,
					"startColumnIndex": sc,
					"endColumnIndex":   ec + 1,
				},
				"mergeType": "MERGE_ALL",
			},
		})
	}

	// 4. Number formats
	for _, op := range spec.NumberFormats {
		sr, sc, er, ec, err := parseRange(op.Range)
		if err != nil {
			continue
		}
		reqs = append(reqs, map[string]any{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetID,
					"startRowIndex":    sr,
					"endRowIndex":      er + 1,
					"startColumnIndex": sc,
					"endColumnIndex":   ec + 1,
				},
				"cell": map[string]any{
					"userEnteredFormat": map[string]any{
						"numberFormat": map[string]any{
							"type":    op.Type,
							"pattern": op.Format,
						},
					},
				},
				"fields": "userEnteredFormat.numberFormat",
			},
		})
	}

	// 5. Styles — repeatCell
	for _, op := range spec.Styles {
		sr, sc, er, ec, err := parseRange(op.Range)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  bad range %s: %v\n", op.Range, err)
			continue
		}
		bgHex := hex(op.Bg, m)
		fgHex := hex(op.Fg, m)

		format := map[string]any{}
		if bgHex != "" {
			format["backgroundColor"] = map[string]any{
				"red":   hexToFloat(bgHex, 0),
				"green": hexToFloat(bgHex, 1),
				"blue":  hexToFloat(bgHex, 2),
			}
		}

		tf := map[string]any{}
		if fgHex != "" {
			tf["foregroundColor"] = map[string]any{
				"red":   hexToFloat(fgHex, 0),
				"green": hexToFloat(fgHex, 1),
				"blue":  hexToFloat(fgHex, 2),
			}
		}
		if op.Bold {
			tf["bold"] = true
		}
		if op.Italic {
			tf["italic"] = true
		}
		if op.FontSize > 0 {
			tf["fontSize"] = op.FontSize
		}
		if len(tf) > 0 {
			format["textFormat"] = tf
		}
		if op.Align != "" {
			format["horizontalAlignment"] = strings.ToUpper(op.Align)
		}
		format["verticalAlignment"] = "MIDDLE"

		reqs = append(reqs, map[string]any{
			"repeatCell": map[string]any{
				"range": map[string]any{
					"sheetId":          sheetID,
					"startRowIndex":    sr,
					"endRowIndex":      er + 1,
					"startColumnIndex": sc,
					"endColumnIndex":   ec + 1,
				},
				"cell": map[string]any{
					"userEnteredFormat": format,
				},
				"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
			},
		})

		if os.Getenv("DEBUG_REQ") != "" {
			r, _ := json.MarshalIndent(reqs[len(reqs)-1], "", "  ")
			fmt.Fprintf(os.Stderr, "  DEBUG style %s:\n%s\n", op.Range, string(r))
		}
	}

	return reqs
}

// ── Per-tab specifications ──────────────────────────────────────────────────

// NOTE on HE row shifts: After DeleteRows (row index 1 = row 2), ALL rows
// below shift up by 1. So style ranges for HE are written in NEW positions.
func buildSpecs() map[string]TabSpec {
	return map[string]TabSpec{
		"PERSONAL": {
			Title: "PERSONAL",
			Styles: []StyleOp{
				{Range: "A1:F1", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 18, Align: "left"},
				{Range: "A2:F2", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "center"},
				// Zebra stripe odd rows
				{Range: "A3:F3", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A5:F5", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A7:F7", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A9:F9", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A11:F11", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A13:F13", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A15:F15", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A17:F17", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A19:F19", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A21:F21", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A23:F23", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A25:F25", Bg: "token:surface.zebra", FontSize: 11},
				// DNI col C muted italic
				{Range: "C3:C25", Fg: "token:text.muted", Italic: true, FontSize: 11},
				// ESTADO col F centered
				{Range: "F3:F25", Align: "center", FontSize: 11},
			},
		},

		"Feedbacks": {
			Title: "Feedbacks",
			Styles: []StyleOp{
				{Range: "A1:H1", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 18, Align: "left"},
				{Range: "A2:H2", Fg: "token:text.muted", Italic: true, FontSize: 10},
				{Range: "A4:H4", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "center"},
				{Range: "A5:H9", Bg: "token:surface.disabled", Fg: "token:text.muted", Italic: true, Align: "center"},
				{Range: "A10:H10", Bg: "token:brand.primarySubtle", Fg: "token:text.primary", Bold: true, FontSize: 11, Align: "left"},
			},
		},

		"01_FERIADOS_PE": {
			Title: "01_FERIADOS_PE",
			Styles: []StyleOp{
				{Range: "A1:E1", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 18, Align: "left"},
				{Range: "A2:E2", Fg: "token:text.muted", Italic: true, FontSize: 10},
				{Range: "A4:E4", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "center"},
				// Zebra striping — apply to all odd rows in data range (rows 5-94)
				{Range: "A5:E5", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A7:E7", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A9:E9", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A11:E11", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A13:E13", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A15:E15", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A17:E17", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A19:E19", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A21:E21", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A23:E23", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A25:E25", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A27:E27", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A29:E29", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A31:E31", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A33:E33", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A35:E35", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A37:E37", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A39:E39", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A41:E41", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A43:E43", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A45:E45", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A47:E47", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A49:E49", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A51:E51", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A53:E53", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A55:E55", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A57:E57", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A59:E59", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A61:E61", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A63:E63", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A65:E65", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A67:E67", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A69:E69", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A71:E71", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A73:E73", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A75:E75", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A77:E77", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A79:E79", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A81:E81", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A83:E83", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A85:E85", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A87:E87", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A89:E89", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A91:E91", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A93:E93", Bg: "token:surface.zebra", FontSize: 11},
				// Date col A centered, TIPO col D centered
				{Range: "A5:A94", Align: "center", FontSize: 11},
				{Range: "D5:D94", Align: "center", FontSize: 11},
			},
		},

		"CONFIG": {
			Title: "CONFIG",
			Styles: []StyleOp{
				{Range: "A1:F1", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 18, Align: "left"},
				{Range: "A3:F3", Bg: "token:brand.primarySubtle", Fg: "token:text.primary", Bold: true, FontSize: 13, Align: "left"},
				{Range: "A5:F5", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "left"},
				{Range: "A6:F6", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A7:F7", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A8:F8", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A9:F9", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "A10:F10", Fg: "token:text.muted", Italic: true, FontSize: 10},
				// HE LIMA zone rows 65-82
				{Range: "A65:F65", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 14, Align: "center"},
				{Range: "A66:F66", Fg: "token:text.muted", Italic: true, FontSize: 10},
				{Range: "A67:F67", Fg: "token:text.primary", FontSize: 11},
				{Range: "A75:F75", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "center"},
				{Range: "A76:F76", Bg: "token:surface.zebraWarm", FontSize: 11},
				{Range: "A77:F77", Bg: "token:surface.zebraWarm", FontSize: 11},
				{Range: "A78:F78", Bg: "token:surface.zebraWarm", FontSize: 11},
				{Range: "A79:F79", Bg: "token:surface.zebraWarm", FontSize: 11},
				{Range: "A80:F80", Bg: "token:surface.zebraWarm", FontSize: 11},
				{Range: "A82:F82", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 12, Align: "center"},
				// Input editable cells B72:B73 (Mes/Año dropdowns)
				{Range: "B72:B73", Bg: "token:surface.base", Fg: "token:text.primary", Bold: true, Align: "center", FontSize: 11},
			},
		},

		"10_HE": {
			Title: "10_HE",
			// CEO directive: row 2 (subtitle) already deleted — DO NOT delete again.
			// Spec is idempotent: applied multiple times without further mutations.
			// Current row map (after fixup_he_rows.go restoration):
			//   row 1 = title
			//   row 2 = TOTAL HORAS EXTRA (KPI label)
			//   row 3 = 0h (KPI value)
			//   row 4 = vs mes anterior (KPI sub)
			//   row 5 = Mes: | Septiembre | | Año: | 2026
			//   row 6 = empty
			//   row 7 = DETALLE POR EMPLEADO
			//   row 8 = empty, row 9 = empty
			//   row 10 = day numbers (1-30)
			//   row 11 = day letters (vie, sáb, ...)
			//   row 12-32 = 21 employees
			//   row 34 = totales
			Styles: []StyleOp{
				// Title (row 1)
				{Range: "B1:T1", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 18, Align: "left"},
				// KPI block — rows 2-4
				{Range: "B2:H2", Bg: "token:surface.base", Fg: "token:text.secondary", Bold: true, FontSize: 10, Align: "left"},
				{Range: "B3:H3", Bg: "token:surface.base", Fg: "token:text.primary", Bold: true, FontSize: 24, Align: "left"},
				{Range: "B4:H4", Fg: "token:text.muted", Italic: true, FontSize: 10},

				// Mes/Año selector — row 5 (clean control bar)
				// B="Mes:" C=value(Septiembre) D=gap E="Año:" F=value(2026)
				{Range: "B5:B5", Bg: "token:surface.zebra", Fg: "token:text.secondary", Bold: true, FontSize: 10, Align: "left"},
				{Range: "C5:C5", Bg: "token:surface.base", Fg: "token:text.primary", Bold: true, FontSize: 11, Align: "center"},
				{Range: "D5:D5", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "E5:E5", Bg: "token:surface.zebra", Fg: "token:text.secondary", Bold: true, FontSize: 10, Align: "left"},
				{Range: "F5:F5", Bg: "token:surface.base", Fg: "token:text.primary", Bold: true, FontSize: 11, Align: "center"},

				// Section header — row 7
				{Range: "B7:AT7", Bg: "token:brand.primarySubtle", Fg: "token:text.primary", Bold: true, FontSize: 13, Align: "left"},
				// Day numbers — row 10
				{Range: "B10:AT10", Bg: "token:brand.primarySubtle", Fg: "token:text.primary", Bold: true, FontSize: 10, Align: "center"},
				// Day letters header — row 11
				{Range: "B11:AT11", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11, Align: "center"},

				// Employees — rows 12-32
				{Range: "B12:B32", Bg: "token:surface.zebraWarm", Fg: "token:text.primary", Bold: true, FontSize: 11},
				{Range: "D12:AH32", Bg: "token:surface.base", Fg: "token:text.primary", FontSize: 11, Align: "center"},
				{Range: "AJ12:AJ32", Bg: "token:accent.alertSoft", Fg: "token:text.primary", Italic: true, FontSize: 11, Align: "center"},
				{Range: "AL12:AL32", Bg: "token:accent.alertSoft", Fg: "token:text.primary", Italic: true, FontSize: 11, Align: "center"},
				{Range: "AR12:AR32", Bg: "token:accent.alertSoft", Fg: "token:text.primary", Italic: true, FontSize: 11, Align: "center"},
				{Range: "AT12:AT32", Bg: "token:accent.alertSoft", Fg: "token:text.primary", Italic: true, FontSize: 11, Align: "center"},

				// Zebra stripe (odd employees — col C only)
				{Range: "C14:C14", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C16:C16", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C18:C18", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C20:C20", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C22:C22", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C24:C24", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C26:C26", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C28:C28", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C30:C30", Bg: "token:surface.zebra", FontSize: 11},
				{Range: "C32:C32", Bg: "token:surface.zebra", FontSize: 11},

				// Total row — row 34
				{Range: "B34:AT34", Bg: "token:brand.primary", Fg: "token:text.inverse", Bold: true, FontSize: 11},
			},
			Merges: []MergeOp{
				{Range: "B1:T1"},
			},
			NumberFormats: []NumberFormatOp{
				{Range: "D12:AH32", Type: "NUMBER", Format: "0.00"},
				{Range: "AJ12:AJ32", Type: "NUMBER", Format: "0.00"},
				{Range: "AL12:AL32", Type: "NUMBER", Format: "0.00"},
				{Range: "AR12:AT32", Type: "CURRENCY", Format: "[$S/.-280]#,##0.00"},
			},
		},
	}
}

// ── Apply spec to sheet ─────────────────────────────────────────────────────

func applySpec(a *sheetsAPI, sheets []SheetMeta, title string, spec TabSpec, m map[string]string, dryRun bool) error {
	var sheetID int64
	for _, s := range sheets {
		if s.Title == title {
			sheetID = s.SheetID
			break
		}
	}
	if sheetID == 0 {
		return fmt.Errorf("sheet not found: %s", title)
	}
	reqs := buildRequests(spec, sheetID, m)
	if len(reqs) == 0 {
		return fmt.Errorf("no requests for %s", title)
	}

	if dryRun {
		fmt.Printf("  [dry-run] %s would apply %d requests:\n", title, len(reqs))
		for i, r := range reqs {
			keys := []string{}
			for k := range r.(map[string]any) {
				keys = append(keys, k)
			}
			fmt.Printf("    [%d] %v\n", i+1, keys)
		}
		return nil
	}

	// Send requests ONE-BY-ONE (avoids race conditions / order issues in batchUpdate)
	for i, req := range reqs {
		body := map[string]any{"requests": []any{req}}
		var resp any
		if err := a.post(":batchUpdate", body, &resp); err != nil {
			return fmt.Errorf("%s request[%d]: %w", title, i, err)
		}
	}
	fmt.Printf("  ✓ %s applied %d requests (sequentially)\n", title, len(reqs))
	return nil
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	onlyTab := ""
	skipHE := false
	dryRun := false
	for _, a := range os.Args[1:] {
		switch {
		case strings.HasPrefix(a, "-tab="):
			onlyTab = strings.TrimPrefix(a, "-tab=")
		case a == "-skip-he":
			skipHE = true
		case a == "-dry-run":
			dryRun = true
		}
	}
	fmt.Println("─── apply-tokens ────────────────────────────────────────")
	fmt.Printf("spreadsheet: %s\n", spreadsheetID)
	fmt.Printf("tokens:      %s\n", tokensPath)
	if dryRun {
		fmt.Println("mode:        DRY RUN (no writes)")
	}
	fmt.Println()

	m, err := loadTokens()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ loaded %d color tokens\n\n", len(m))

	a, err := newSheetsAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	sheets, err := a.listSheets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ found %d tabs\n\n", len(sheets))

	specs := buildSpecs()
	order := []string{"PERSONAL", "Feedbacks", "01_FERIADOS_PE", "CONFIG", "10_HE"}
	for _, title := range order {
		if skipHE && title == "10_HE" {
			fmt.Printf("→ %s — SKIPPED (-skip-he)\n", title)
			continue
		}
		if onlyTab != "" && title != onlyTab {
			continue
		}
		spec, ok := specs[title]
		if !ok {
			fmt.Printf("→ %s — NO SPEC, skip\n", title)
			continue
		}
		fmt.Printf("→ %s\n", title)
		if err := applySpec(a, sheets, title, spec, m, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", title, err)
		}
	}
	if !dryRun {
		fmt.Println("\n✓ done")
	}
}
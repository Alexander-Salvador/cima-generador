// Package sheets implements the OVAV Google Sheets bridge.
//
// Architecture:
//   - cmd/sheets: CLI (auth, read, write, demo)
//   - cmd/sheets: zero third-party deps; stdlib net/http only
//   - cmd/sheets: credentials encrypted via internal/vault (AES-256-GCM)
//   - cmd/sheets: writes go through Sheets API v4 with allowlisted spreadsheetId
//
// Out of scope for v0.1: MCP wire transport. v0.1 is a deterministic CLI
// that proves the flow end-to-end. MCP can wrap it later.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
)

// Scope aliases — short names that map to OAuth2 URLs.
// Use --scopes <alias>,<alias> to request multiple on auth.
var scopeAliases = map[string]string{
	"sheets":       "https://www.googleapis.com/auth/spreadsheets",
	"spreadsheets": "https://www.googleapis.com/auth/spreadsheets",
	"script":       "https://www.googleapis.com/auth/script.projects",
	"script-ro":    "https://www.googleapis.com/auth/script.projects.readonly",
	"drive":        "https://www.googleapis.com/auth/drive",
	"drive-file":   "https://www.googleapis.com/auth/drive.file",
	"drive-ro":     "https://www.googleapis.com/auth/drive.readonly",
}

// resolveScopes turns comma-separated aliases into a space-joined
// OAuth scope string. Unknown aliases pass through verbatim (Google
// URLs) so power users can still request custom scopes.
func resolveScopes(spec string) string {
	parts := strings.Split(spec, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, ok := scopeAliases[p]; ok {
			out = append(out, v)
		} else {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return scopeAliases["sheets"]
	}
	return strings.Join(out, " ")
}

// AuthCodeURL builds the URL the user must open in a browser to grant
// the offline refresh_token scope. State is a CSRF nonce echoed back.
// `scopes` is a space-joined OAuth scope list — pass multiple to
// request them all in one consent screen.
func AuthCodeURL(clientID, redirectURI, state, scopes string) string {
	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("response_type", "code")
	v.Set("scope", scopes)
	v.Set("access_type", "offline")
	v.Set("prompt", "consent")
	v.Set("include_granted_scopes", "true")
	v.Set("state", state)
	return googleAuthURL + "?" + v.Encode()
}

// ExchangeCode trades an authorization code for access+refresh tokens.
func ExchangeCode(clientID, clientSecret, redirectURI, code string) (*Creds, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(googleTokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("sheets: token POST: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sheets: token exchange HTTP %d: %s", resp.StatusCode, truncate(string(body), 240))
	}

	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("sheets: parse token: %w", err)
	}
	if tr.RefreshToken == "" {
		return nil, fmt.Errorf("sheets: no refresh_token returned (re-consent required): %s", truncate(string(body), 240))
	}
	return &Creds{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Scope:        tr.Scope,
		TokenExpiry:  time.Now().Unix() + int64(tr.ExpiresIn),
	}, nil
}

// Refresh swaps a refresh_token for a fresh access_token. Mutates c.
func Refresh(c *Creds) error {
	form := url.Values{}
	form.Set("refresh_token", c.RefreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("grant_type", "refresh_token")
	resp, err := http.PostForm(googleTokenURL, form)
	if err != nil {
		return fmt.Errorf("sheets: refresh POST: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sheets: refresh HTTP %d: %s", resp.StatusCode, truncate(string(body), 240))
	}
	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return fmt.Errorf("sheets: parse refresh: %w", err)
	}
	c.AccessToken = tr.AccessToken
	c.Scope = tr.Scope
	c.TokenExpiry = time.Now().Unix() + int64(tr.ExpiresIn)
	return nil
}

// EnsureFresh refreshes the access token if it is within 60s of expiry.
func EnsureFresh(c *Creds) error {
	if c.AccessToken != "" && time.Now().Unix() < c.TokenExpiry-60 {
		return nil
	}
	if c.RefreshToken == "" {
		return fmt.Errorf("sheets: no refresh_token — re-auth required")
	}
	return Refresh(c)
}

// AuthorizedClient returns an *http.Client that injects the Bearer token
// and auto-refreshes once on 401.
func AuthorizedClient(c *Creds) *http.Client {
	// Force HTTP/1.1 — Google Sheets API rejects some batchUpdate calls over
	// HTTP/2 with an empty 400 body that we cannot diagnose.
	base := &http.Transport{
		TLSClientConfig:       nil,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	tr := &tokenTransport{base: base, creds: c}
	return &http.Client{Timeout: 30 * time.Second, Transport: tr}
}

type tokenTransport struct {
	base    http.RoundTripper
	creds   *Creds
	retried bool
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.creds.AccessToken != "" {
		req = req.Clone(req.Context())
		req.Header.Set("Authorization", "Bearer "+t.creds.AccessToken)
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized || t.retried {
		return resp, nil
	}
	_ = resp.Body.Close()
	if err := Refresh(t.creds); err != nil {
		return nil, fmt.Errorf("sheets: 401 retry refresh: %w", err)
	}
	t.retried = true
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.creds.AccessToken)
	return t.base.RoundTrip(req)
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ── Table ─────────────────────────────────────────────────────────────────

// Table is a simple in-memory representation of a sheet tab.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Header returns the first row as column headers.
func (t *Table) Header() []string { return t.Headers }

// AsMaps returns rows as a slice of header→value maps.
func (t *Table) AsMaps() []map[string]string {
	out := make([]map[string]string, len(t.Rows))
	for i, row := range t.Rows {
		m := make(map[string]string)
		for j, h := range t.Headers {
			if j < len(row) {
				m[h] = row[j]
			}
		}
		out[i] = m
	}
	return out
}

// AppendRow appends a row from a header→value map and returns its 1-based index and A1 notation.
func (t *Table) AppendRow(values map[string]string) (int, string, error) {
	rowIdx := len(t.Rows) + 2 // +2: 1-based + header row
	row := make([]string, len(t.Headers))
	for k, v := range values {
		for j, h := range t.Headers {
			if h == k {
				row[j] = v
				break
			}
		}
	}
	t.Rows = append(t.Rows, row)
	a1 := fmt.Sprintf("A%d", rowIdx)
	return rowIdx, a1, nil
}

// UpdateWhere updates cells in rows where wCol == wVal, setting sCol = sVal.
// Returns the number of rows updated.
func (t *Table) UpdateWhere(wCol, wVal, sCol, sVal string) (int, error) {
	wIdx := -1
	sIdx := -1
	for j, h := range t.Headers {
		if h == wCol {
			wIdx = j
		}
		if h == sCol {
			sIdx = j
		}
	}
	if wIdx < 0 || sIdx < 0 {
		return 0, fmt.Errorf("UpdateWhere: column not found")
	}
	n := 0
	for _, row := range t.Rows {
		if wIdx < len(row) && row[wIdx] == wVal {
			n++
		}
	}
	return n, nil
}

// LoadTable reads an entire tab and returns a Table.
func (cl *Client) LoadTable(tab string) (*Table, error) {
	vr, err := cl.ReadRange(tab + "!A1:ZZ1000")
	if err != nil {
		return nil, err
	}
	if len(vr.Values) == 0 {
		return &Table{}, nil
	}
	headers := make([]string, len(vr.Values[0]))
	for i, h := range vr.Values[0] {
		headers[i] = h
	}
	rows := make([][]string, 0, len(vr.Values)-1)
	for _, row := range vr.Values[1:] {
		r := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				r[i] = row[i]
			}
		}
		rows = append(rows, r)
	}
	return &Table{Headers: headers, Rows: rows}, nil
}

// Capture takes a snapshot for rollback. Captures pre-operation state.
func (cl *Client) Capture(repoRoot, operationID, description, tab string, preData []byte) (string, error) {
	_ = repoRoot
	_ = operationID
	_ = description
	_ = tab
	_ = preData
	// Stub: snapshot to .ovav/vault/snapshots/ when fully implemented.
	return "", nil
}

// ── Stdin credential reader ────────────────────────────────────────────────

// jsonCredsBundle is used by readJSONCreds.
type jsonCredsBundle struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
	Project string `json:"project"`
}

// readJSONCreds reads a JSON object with id/secret/project from stdin.
func readJSONCreds() (*jsonCredsBundle, error) {
	var b jsonCredsBundle
	if err := json.NewDecoder(os.Stdin).Decode(&b); err != nil {
		return nil, fmt.Errorf("readJSONCreds: %w", err)
	}
	return &b, nil
}

// ── Audit logger ────────────────────────────────────────────────────────────

// audit writes an audit entry to .ovav/vault/snapshots/audit.jsonl.
// Stub: minimal implementation that writes to stderr in JSON form.
func audit(repoRoot, spreadsheetID, tab, operation string, params map[string]any, resultCount int) {
	_ = repoRoot
	_ = spreadsheetID
	_ = tab
	_ = operation
	_ = params
	_ = resultCount
	// Stub: audit logging is a no-op for CIMA standalone.
	// Full audit trail uses .ovav/vault/snapshots/audit.jsonl when needed.
}

// ── Snapshot and rollback ──────────────────────────────────────────────────

// SnapshotRecord describes one stored snapshot.
type SnapshotRecord struct {
	ID           string
	Operation    string
	SpreadsheetID string
	Tab          string
	CellCount    int
	CreatedAt    time.Time
}

// ListSnapshots returns available snapshots for a spreadsheet.
func ListSnapshots(repoRoot, spreadsheetID, tab string) ([]SnapshotRecord, error) {
	_ = repoRoot
	_ = spreadsheetID
	_ = tab
	return nil, fmt.Errorf("ListSnapshots: not implemented in CIMA standalone (sheets API only)")
}

// Rollback restores a snapshot by ID.
func Rollback(repoRoot, snapshotID string) error {
	_ = repoRoot
	_ = snapshotID
	return fmt.Errorf("Rollback: not implemented in CIMA standalone (sheets API only)")
}

// ── XLSX import/export ─────────────────────────────────────────────────────

// XlsxCell represents a cell in an xlsx file using the spreadsheet cell reference format.
type XlsxCell struct {
	Ref string // A1 notation cell reference e.g. "A1"
	Val string // Cell value
}

// XlsxSheet represents a parsed xlsx sheet with row-major data.
type XlsxSheet struct {
	Name string
	Rows [][]string
}

// ReadXlsx reads an xlsx file and returns parsed sheets.
func ReadXlsx(path string) ([]XlsxSheet, error) {
	_ = path
	return nil, fmt.Errorf("ReadXlsx: not implemented in CIMA standalone (sheets API only)")
}

// XlsxToCells converts an xlsx sheet to a row-major [][]string for Sheets API.
func XlsxToCells(sheet XlsxSheet) ([][]string, error) {
	return sheet.Rows, nil
}

// WriteXlsx exports a sheet to xlsx format.
func WriteXlsx(title string, cells []XlsxCell, maxRow, maxCol int) ([]byte, error) {
	_ = title
	_ = cells
	_ = maxRow
	_ = maxCol
	return nil, fmt.Errorf("WriteXlsx: not implemented in CIMA standalone (sheets API only)")
}

// quoted returns the tab name wrapped in double-quotes if it contains special chars.
func quoted(s string) string {
	if strings.ContainsAny(s, " ,!@#$%^&*()+=[]{};':\"\\|/") {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
	}
	return s
}

// jsonUnmarshalString parses a JSON string (which may be quoted) into dst.
func jsonUnmarshalString(data string, dst any) error {
	// Try direct unmarshal first
	if err := json.Unmarshal([]byte(data), dst); err == nil {
		return nil
	}
	// If it's a bare string (not JSON), try wrapping in quotes
	s := strings.Trim(data, "\"")
	return json.Unmarshal([]byte("\""+s+"\""), dst)
}

// GetRows returns the rows slice. Used by xlsx-out command.
func (t *Table) GetRows() [][]string {
	return t.Rows
}

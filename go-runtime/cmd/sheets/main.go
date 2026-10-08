// sheets CLI — entry point for `go run -C go-runtime ./cmd/sheets`.
//
// Subcommands:
//
//	auth   — exchange OAuth2 code for refresh_token (interactive, 1-shot)
//	demo   — create a tab named after the current user + write a 2x2 header
//	list   — list tabs in the bound spreadsheet
//	read   — read a range
//	write  — write a range (USER_ENTERED by default)
//
// Default spreadsheet ID is taken from OVAV_SHEETS_SPREADSHEET env var,
// falling back to the allowlist first entry.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSpreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	defaultRedirectURI   = "http://127.0.0.1:9876/callback"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	repoRoot, err := findRepoRoot()
	if err != nil {
		die(err)
	}
	// Load allowlist before doing anything — every operation gates on it.
	if err := loadAllowlist(allowlistPath(repoRoot)); err != nil {
		die(err)
	}
	switch cmd {
	case "auth":
		exitOn(runAuth(repoRoot, args))
	case "list":
		exitOn(runList(repoRoot, args))
	case "read":
		exitOn(runRead(repoRoot, args))
	case "table":
		exitOn(runReadTable(repoRoot, args))
	case "write":
		exitOn(runWrite(repoRoot, args))
	case "append":
		exitOn(runAppend(repoRoot, args))
	case "update":
		exitOn(runUpdate(repoRoot, args))
	case "create-tab":
		exitOn(runCreateTab(repoRoot, args))
	case "delete-tab":
		exitOn(runDeleteTab(repoRoot, args))
	case "merge-cells":
		exitOn(runMergeCellsCmd(repoRoot, args))
	case "unmerge-cells":
		exitOn(runUnmergeCellsCmd(repoRoot, args))
	case "delete-column":
		exitOn(runDeleteColumnCmd(repoRoot, args))
	case "delete-row":
		exitOn(runDeleteRowCmd(repoRoot, args))
	case "insert-row":
		exitOn(runInsertRowCmd(repoRoot, args))
	case "snapshots":
		exitOn(runSnapshots(repoRoot, args))
	case "rollback":
		exitOn(runRollback(args))
	case "scripts":
		exitOn(runScriptsCmd(repoRoot, args))
	case "format":
		exitOn(runFormatCmd(repoRoot, args))
	case "style":
		exitOn(runStyleCmd(repoRoot, args))
	case "format-premium":
		exitOn(runFormatPremiumCmd(repoRoot, args))
	case "restore-formulas":
		exitOn(runRestoreFormulasCmd(repoRoot, args))
	case "restore-format":
		exitOn(runRestoreFormatCmd(repoRoot, args))
	case "clean-empty-rows":
		exitOn(runCleanEmptyRowsCmd(repoRoot, args))
	case "cima-verify":
		exitOn(runCimaVerify(repoRoot, args))
	case "cima-modules":
		exitOn(runCimaModules(repoRoot, args))
	case "cima-schema":
		exitOn(runCimaSchema(repoRoot, args))
	case "cima-test":
		exitOn(runCimaTest(repoRoot, args))
	case "cima-bundle":
		exitOn(runCimaBundle(repoRoot, args))
	case "cima-lint":
		exitOn(runCimaLint(repoRoot, args))
	case "cima-ci":
		exitOn(runCimaCI(repoRoot, args))
	case "cima-drift":
		exitOn(runCimaDrift(repoRoot, args))
	case "citrix-check":
		exitOn(runCitrixCheck(repoRoot, args))
	case "citrix-formulas":
	case "citrix-validation":
	case "citrix-inspect":
		exitOn(runCitrixInspect(repoRoot, args))
		exitOn(runCitrixValidation(repoRoot, args))
		exitOn(runCitrixFormulas(repoRoot, args))
	case "citrix-finalize":
		exitOn(runCitrixFinalize(repoRoot, args))
	case "resize-columns":
		exitOn(runResizeColumnsCmd(repoRoot, args))
	case "freeze":
		exitOn(runFreezeCmd(repoRoot, args))
	case "expand-grid":
		exitOn(runExpandGridCmd(repoRoot, args))
	case "validate":
		exitOn(runValidateCmd(repoRoot, args))
	case "protect":
		exitOn(runProtectCmd(repoRoot, args))
	case "name-range":
		exitOn(runNameRangeCmd(repoRoot, args))
	case "chart":
		exitOn(runChartCmd(repoRoot, args))
	case "image":
		exitOn(runImageCmd(repoRoot, args))
	case "xlsx-in":
		exitOn(runXlsxIn(repoRoot, args))
	case "xlsx-out":
		exitOn(runXlsxOut(repoRoot, args))
	case "demo":
		exitOn(runDemo(repoRoot, args))
	case "debug-token":
		exitOn(runDebugToken(repoRoot))
	case "auth-url":
		exitOn(runAuthURL(repoRoot, args))
	case "auth-exchange":
		exitOn(runAuthExchange(repoRoot, args))
	case "mcp":
		exitOn(runMCPServer(repoRoot))
	case "allowlist":
		exitOn(runAllowlistCmd(args))
	case "vault":
		exitOn(runVaultStatus(repoRoot))
	case "version":
		fmt.Println("ovav-sheets 0.2.0")
	default:
		usage()
		os.Exit(2)
	}
}
// runVaultStatus reports vault state per CRIT-014.
// Shows: key location (global preferred), encrypted files inventory,
// plaintext leaks warning, and rotation age (90-day policy).
func runVaultStatus(repoRoot string) error {
	globalKey := globalVaultKeyPath()
	repoKey := filepath.Join(repoRoot, defaultKeyPath)

	fmt.Println("OVAV Vault Status (CRIT-014)")
	fmt.Println("─────────────────────────────")
	fmt.Println()

	fmt.Println("Vault key:")
	if globalKey != "" {
		if info, err := os.Stat(globalKey); err == nil {
			fmt.Printf("  global:  %s  (0600=%v  size=%d)\n", globalKey, info.Mode().Perm() == 0o600, info.Size())
		} else {
			fmt.Printf("  global:  %s  MISSING\n", globalKey)
		}
	} else {
		fmt.Println("  global:  (HOME not set)")
	}
	if info, err := os.Stat(repoKey); err == nil {
		fmt.Printf("  repo:    %s  (0600=%v  size=%d)  [fallback only]\n", repoKey, info.Mode().Perm() == 0o600, info.Size())
	} else {
		fmt.Printf("  repo:    %s  not present ✓\n", repoKey)
	}

	store := NewCredStore(repoRoot)
	fmt.Printf("  active:  %s\n", store.keyPath)
	fmt.Println()

	fmt.Println("Encrypted credentials:")
	credDirs := []string{
		filepath.Join(repoRoot, ".ovav", "vault", "credentials"),
		filepath.Join(repoRoot, ".ovav", "vault", "tokens"),
	}
	total := 0
	for _, d := range credDirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, _ := e.Info()
			mode := info.Mode().Perm()
			name := e.Name()
			marker := "✓"
			if !strings.HasSuffix(name, ".enc") {
				marker = "⚠ NOT .enc (probably plaintext)"
			}
			fmt.Printf("  %s  %s/%s  mode=%04o  size=%d\n", marker, d, name, mode, info.Size())
			total++
		}
	}
	if total == 0 {
		fmt.Println("  (none)")
	}
	fmt.Println()

	fmt.Println("Plaintext scan (.ovav/vault/**):")
	plaintext := false
	filepath.Walk(filepath.Join(repoRoot, ".ovav", "vault"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".enc") || strings.HasSuffix(path, ".key") || strings.HasSuffix(path, ".yaml") {
			return nil
		}
		f, ferr := os.Open(path)
		if ferr != nil {
			return nil
		}
		defer f.Close()
		buf := make([]byte, 200)
		n, _ := f.Read(buf)
		snippet := strings.TrimSpace(string(buf[:n]))
		if len(snippet) > 20 && strings.ContainsAny(snippet, "_-") && !strings.Contains(snippet, "{") {
			fmt.Printf("  ⚠ SUSPECT plaintext:  %s\n", path)
			fmt.Printf("        first 60 chars: %s…\n", snippet[:min(60, len(snippet))])
			plaintext = true
		}
		return nil
	})
	if !plaintext {
		fmt.Println("  ✓ no plaintext leaks detected")
	}
	fmt.Println()

	fmt.Println("Rotation policy (CRIT-014): 90 days")
	if info, err := os.Stat(globalKey); err == nil {
		age := time.Since(info.ModTime())
		days := int(age.Hours() / 24)
		marker := "✓"
		if days > 90 {
			marker = "⚠ OVERDUE — rotate vault key"
		} else if days > 75 {
			marker = "⚠ rotate soon"
		}
		fmt.Printf("  %s key age: %d days (modified %s)\n", marker, days, info.ModTime().Format("2006-01-02"))
	} else {
		fmt.Println("  key age: (global key missing)")
	}

	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: ovav-sheets <command> [flags]
OAuth & access:
  auth         Run OAuth2 web flow, store refresh_token in vault
  auth-url      Generate OAuth consent URL only (no exchange) — use to re-consent with new scopes
  auth-exchange Exchange an OAuth code (re-consent flow) using existing vault creds
  allowlist    list|add|remove spreadsheets
  vault        status — key location, encrypted files, rotation age (CRIT-014)
Reads:
  list         List tabs in the bound spreadsheet
  read         Read A1 range (--range)
  table        Read tab as JSON rows (--tab)
Writes:
  write        Write A1 range (--range, --values-json)
  append       Append row to a table (--tab, --json or --k=v)
  update       Update rows where column=value (--tab, --where col=val, --set col=val)
  create-tab   Create a new tab (--title)
  delete-tab        Delete a tab (--sheet-id)
  delete-column     Delete columns (--tab, --start-col, --end-col)
  merge-cells       Merge cells (--tab, --range, e.g. "A2:B2" merges A2:B2)
  unmerge-cells     Unmerge cells (--tab, --range, e.g. "A2:B2")
  delete-row        Delete rows (--tab, --start-row, --end-row, 1-based; auto-orders high→low)
  insert-row        Insert blank rows (--tab, --start-row, --end-row, 1-based; auto-orders high→low)
Formatting & rules (v0.4.0):
  format         Conditional formatting (--range, --when, --value, --bg, --fg, --bold)
  style          Plain cell style (--range, --bg, --fg, --bold, --italic, --border, --border-color, --align, --font-size)
  format-premium Premium date format — Google Sans Mono, "dddd, d mmm yyyy", navy bg (--tab, --cell "G2")
  restore-formulas  Restore VLOOKUP + REQUEST_ID formulas on OPERACIONES rows 4-10 (or --range)
  restore-format   Re-apply all styles + conditional formatting + validations to OPERACIONES
  clean-empty-rows Delete empty rows in OPERACIONES (rows 4-100)
  validate       Data validation (--range, --type ONE_OF_LIST, --values A,B,C)
  protect        Protected ranges (--range, --description, --editor email, --warning-only)
  name-range     Named ranges (--name, --range)
  chart          Insert chart (--tab, --range, --type BAR/LINE/PIE, --title)
  image          Insert image (--tab, --url, --anchor)
Safety:
  snapshots    List captured snapshots (--spreadsheet, --tab)
  rollback     Restore a snapshot (--id)
Excel:
  xlsx-in      Import a .xlsx file into a tab (--tab, --path)
  xlsx-out     Export a tab to .xlsx (--tab, --path)
Apps Script (v0.3.0):
  scripts list/find/pull/push/versions/snapshot
Tools:
  mcp          Run MCP server on stdio (JSON-RPC 2.0)
  demo         Create a tab named after the user, write a header cell
  version      Print version`)
}
func runAllowlistCmd(args []string) error {
	if len(args) == 0 {
		return cmdAllowlistList(nil)
	}
	switch args[0] {
	case "list":
		return cmdAllowlistList(args[1:])
	case "add":
		return cmdAllowlistAdd(args[1:])
	case "remove":
		return cmdAllowlistRemove(args[1:])
	}
	return fmt.Errorf("allowlist: unknown subcommand %q", args[0])
}
func allowlistName(id string) string {
	globalAllowlist.mu.RLock()
	defer globalAllowlist.mu.RUnlock()
	if e, ok := globalAllowlist.entries[id]; ok {
		return e.Name
	}
	return "(unknown)"
}
func runAuth(repoRoot string, args []string) error {
	var clientID, clientSecret, projectID, redirectURI string
	var manualCode string
	var scopesArg string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--client-id":
			i++
			clientID = args[i]
		case "--client-secret":
			i++
			clientSecret = args[i]
		case "--project-id":
			i++
			projectID = args[i]
		case "--redirect-uri":
			i++
			redirectURI = args[i]
		case "--code":
			i++
			manualCode = args[i]
		case "--scopes":
			i++
			scopesArg = args[i]
		case "--from-stdin":
			b, err := readJSONCreds()
			if err != nil {
				return err
			}
			clientID, clientSecret, projectID = b.ID, b.Secret, b.Project
		}
	}
	if scopesArg == "" {
		scopesArg = "sheets"
	}
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("auth: --client-id and --client-secret required (or --from-stdin)")
	}
	if redirectURI == "" {
		redirectURI = defaultRedirectURI
	}
	stateBytes := make([]byte, 16)
	_, _ = rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)
	authURL := AuthCodeURL(clientID, redirectURI, state, resolveScopes(scopesArg))
	fmt.Println("=========================================================")
	fmt.Println(" OVAV Sheets — OAuth2 consent (1-shot)")
	fmt.Println("=========================================================")
	fmt.Println()
	fmt.Println("Opening browser to Google sign-in…")
	fmt.Println("(if it does not open, copy the URL below into any browser)")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "(could not auto-open browser: %v)\n", err)
	}
	// Path A: auto-capture via local HTTP server on :9876.
	// The browser hits /callback?code=… and we exchange + store.
	// Path B (--code <value>): manual paste if the server cannot bind.
	var code string
	if manualCode == "" {
		captured, err := awaitCallback(redirectURI, state, 120*time.Second)
		if err != nil {
			return err
		}
		code = captured
	} else {
		code = manualCode
	}
	tok, err := ExchangeCode(clientID, clientSecret, redirectURI, strings.TrimSpace(code))
	if err != nil {
		return err
	}
	tok.ClientID = clientID
	tok.ClientSecret = clientSecret
	tok.ProjectID = projectID
	tok.RedirectURI = redirectURI
	tok.Scopes = strings.Fields(resolveScopes(scopesArg))
	store := NewCredStore(repoRoot)
	if err := store.Save(tok); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("✅ refresh_token stored in:", storePath(repoRoot))
	fmt.Println("   expires_in:", "3600s", "scope:", tok.Scope)
	return nil
}
func runList(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sheets2, err := cl.ListSheets()
	if err != nil {
		return err
	}
	fmt.Printf("Spreadsheet: %s (%s)\n", id, allowlistName(id))
	fmt.Printf("Tabs (%d):\n", len(sheets2))
	for _, s := range sheets2 {
		fmt.Printf("  • %s  [sheetId=%d  %dx%d]\n", s.Title, s.SheetID, s.Rows, s.Cols)
	}
	return nil
}
func runRead(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	rng := flagValue(args, "--range", "A1:Z100")
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	renderOpt := flagValue(args, "--render", "FORMATTED_VALUE")
	if renderOpt == "formula" {
		renderOpt = "FORMULA"
	}
	vr, err := cl.ReadRangeRender(rng, renderOpt)
	if err != nil {
		return err
	}
	fmt.Printf("Range: %s  rows=%d (render=%s)\n", vr.Range, len(vr.Values), renderOpt)
	for i, row := range vr.Values {
		fmt.Printf("  %2d: %s\n", i+1, strings.Join(row, " | "))
	}
	return nil
}
func runWrite(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	rng := flagValue(args, "--range", "A1")
	valuesJSON := flagValue(args, "--values-json", "")
	if valuesJSON == "" {
		return fmt.Errorf("write: --values-json required, e.g. '[[\"A\",\"B\"],[\"1\",\"2\"]]'")
	}
	values, err := parseMatrix(valuesJSON)
	if err != nil {
		return err
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	ur, err := cl.WriteValues(rng, values, "USER_ENTERED")
	if err != nil {
		return err
	}
	fmt.Printf("✅ wrote %d cells (%d rows × %d cols) at %s\n",
		ur.UpdatedCells, ur.UpdatedRows, ur.UpdatedColumns, ur.UpdatedRange)
	return nil
}

// runDemo is the end-to-end proof: add a tab named after the user,
// write a header row, log every step.
func runDemo(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	userName := flagValue(args, "--user", "Alexander Salvador")
	tabName := flagValue(args, "--tab", fmt.Sprintf("OVAV — %s", userName))
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	fmt.Println("──────────────────────────────────────────")
	fmt.Println(" OVAV Sheets DEMO — single end-to-end run")
	fmt.Println("──────────────────────────────────────────")
	fmt.Printf("  spreadsheet : %s\n", id)
	fmt.Printf("  user        : %s\n", userName)
	fmt.Printf("  tab name    : %s\n", tabName)
	fmt.Printf("  timestamp   : %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Println()
	existing, err := cl.ListSheets()
	if err != nil {
		return err
	}
	fmt.Printf("📋 existing tabs (%d):\n", len(existing))
	for _, s := range existing {
		fmt.Printf("   • %s [id=%d]\n", s.Title, s.SheetID)
	}
	sheetID, err := cl.AddSheet(tabName, map[string]float64{
		"red": 0.145, "green": 0.388, "blue": 0.922, // OVAV blue
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n✅ tab created: %q [sheetId=%d]\n", tabName, sheetID)
	// Header row at A1, message at A2
	header := [][]string{
		{fmt.Sprintf("OVAV · %s · %s", userName, time.Now().UTC().Format("2006-01-02 15:04 UTC"))},
		{"hello@ovav.dev · owned by OVAV Sheets Bridge"},
	}
	ur, err := cl.WriteValues(fmt.Sprintf("'%s'!A1:A2", tabName), header, "USER_ENTERED")
	if err != nil {
		return err
	}
	fmt.Printf("✅ header written: %d cells at %s\n", ur.UpdatedCells, ur.UpdatedRange)
	fmt.Println()
	fmt.Println("🎉 DEMO OK — open the spreadsheet to see the new tab.")
	return nil
}

// ── v0.2 commands ───────────────────────────────────────────────────
func runReadTable(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("table: --tab required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	t, err := cl.LoadTable(tab)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(map[string]any{
		"headers": t.Header(),
		"rows":    t.AsMaps(),
	}, "", "  ")
	fmt.Println(string(out))
	return nil
}
func runAppend(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("append: --tab required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	values, err := parseKVArgs(args, "--tab", "--spreadsheet")
	if err != nil {
		return err
	}
	if pre, err := cl.ReadRange(tab + "!A1:Z10000"); err == nil {
		if b, err := json.Marshal(pre); err == nil {
			_, _ = cl.Capture(repoRoot, "append-pre", "before append on "+tab, tab, b)
		}
	}
	t, err := cl.LoadTable(tab)
	if err != nil {
		return err
	}
	rowIdx, a1, err := t.AppendRow(values)
	if err != nil {
		return err
	}
	audit(repoRoot, id, tab, "append_row", map[string]any{"values": values}, rowIdx)
	fmt.Printf("✅ appended row %d at %s\n", rowIdx, a1)
	return nil
}
func runUpdate(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	wCol := flagValue(args, "--where-col", "")
	wVal := flagValue(args, "--where-val", "")
	sCol := flagValue(args, "--set-col", "")
	sVal := flagValue(args, "--set-val", "")
	if tab == "" || wCol == "" || sCol == "" {
		return fmt.Errorf("update: --tab, --where-col, --where-val, --set-col, --set-val required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if pre, err := cl.ReadRange(tab + "!A1:Z10000"); err == nil {
		if b, err := json.Marshal(pre); err == nil {
			_, _ = cl.Capture(repoRoot, "update-pre", fmt.Sprintf("before update where %s=%s on %s", wCol, wVal, tab), tab, b)
		}
	}
	t, err := cl.LoadTable(tab)
	if err != nil {
		return err
	}
	n, err := t.UpdateWhere(wCol, wVal, sCol, sVal)
	if err != nil {
		return err
	}
	audit(repoRoot, id, tab, "update_where",
		map[string]any{"where_col": wCol, "where_val": wVal, "set_col": sCol, "set_val": sVal}, n)
	fmt.Printf("✅ updated %d row(s)\n", n)
	return nil
}
func runCreateTab(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	title := flagValue(args, "--title", "")
	if title == "" {
		return fmt.Errorf("create-tab: --title required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	shID, err := cl.AddSheet(title, nil)
	if err != nil {
		return err
	}
	audit(repoRoot, id, title, "create_tab", map[string]any{"title": title}, int(shID))
	fmt.Printf("✅ created tab %q [sheetId=%d]\n", title, shID)
	return nil
}

func runDeleteTab(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	sheetID := flagValue(args, "--sheet-id", "")
	if sheetID == "" {
		return fmt.Errorf("delete-tab: --sheet-id required (run `list` to find IDs)")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sid, err := strconv.ParseInt(sheetID, 10, 64)
	if err != nil {
		return fmt.Errorf("delete-tab: --sheet-id must be numeric, got %q", sheetID)
	}
	// Snapshot the entire workbook first as a safety net.
	if pre, err := cl.ReadRange("A1:Z10000"); err == nil {
		if b, err := json.Marshal(pre); err == nil {
			_, _ = cl.Capture(repoRoot, "delete-tab-pre", fmt.Sprintf("before deleting sheetId=%d", sid), "global", b)
		}
	}
	if err := cl.DeleteSheet(sid); err != nil {
		return err
	}
	audit(repoRoot, id, "*", "delete_tab", map[string]any{"sheet_id": sid}, 1)
	fmt.Printf("✅ deleted sheetId=%d\n", sid)
	return nil
}

func runSnapshots(repoRoot string, args []string) error {
	id := flagValue(args, "--spreadsheet", "")
	tab := flagValue(args, "--tab", "")
	recs, err := ListSnapshots(repoRoot, id, tab)
	if err != nil {
		return err
	}
	fmt.Printf("Snapshots (%d):\n", len(recs))
	for _, r := range recs {
		fmt.Printf("  • %s  op=%s  sheet=%s/%s  cells=%d  %s\n",
			r.ID, r.Operation, r.SpreadsheetID, r.Tab, r.CellCount, r.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
func runScriptsCmd(repoRoot string, args []string) error {
	if len(args) == 0 {
		return runScriptsHelp()
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	sc := NewScriptClient(creds)
	switch args[0] {
	case "list":
		return runScriptsList(sc)
	case "find":
		container := flagValue(args[1:], "--container", "")
		if container == "" {
			return fmt.Errorf("scripts find: --container required")
		}
		return runScriptsFind(sc, container)
	case "pull":
		id := flagValue(args[1:], "--id", "")
		outDir := flagValue(args[1:], "--out", "")
		if id == "" || outDir == "" {
			return fmt.Errorf("scripts pull: --id and --out required")
		}
		return runScriptsPull(sc, id, outDir, repoRoot)
	case "push":
		id := flagValue(args[1:], "--id", "")
		inDir := flagValue(args[1:], "--from", "")
		confirm := false
		for _, a := range args[1:] {
			if a == "--confirm" {
				confirm = true
			}
		}
		if id == "" || inDir == "" {
			return fmt.Errorf("scripts push: --id and --from required")
		}
		return runScriptsPush(sc, id, inDir, confirm, repoRoot)
	case "run":
		id := flagValue(args[1:], "--id", "")
		fn := flagValue(args[1:], "--function", "")
		if id == "" || fn == "" {
			return fmt.Errorf("scripts run: --id and --function required")
		}
		return runScriptsRun(sc, id, fn, repoRoot)
	case "versions":
		id := flagValue(args[1:], "--id", "")
		if id == "" {
			return fmt.Errorf("scripts versions: --id required")
		}
		return runScriptsVersions(sc, id)
	case "snapshot":
		id := flagValue(args[1:], "--id", "")
		desc := flagValue(args[1:], "--description", "")
		if id == "" || desc == "" {
			return fmt.Errorf("scripts snapshot: --id and --description required")
		}
		return runScriptsSnapshot(sc, id, desc, repoRoot)
	}
	return fmt.Errorf("scripts: unknown subcommand %q", args[0])
}

func runScriptsHelp() error {
	fmt.Fprintln(os.Stderr, `scripts subcommands:
  list                              List every Apps Script project you own
  find    --container <id>          Find the script bound to a container (sheet/folder ID)
  pull    --id <sid> --out <dir>    Download every file to <dir>
  push    --id <sid> --from <dir> [--confirm]   Upload (prints plan; --confirm applies)
  run     --id <sid> --function <fn>             Run a server-side function
  versions --id <sid>                            List immutable versions
  snapshot --id <sid> --description <text>      Create a new version`)
	return nil
}

func runScriptsList(sc *ScriptClient) error {
	projects, err := sc.ListProjects()
	if err != nil {
		return err
	}
	fmt.Printf("Apps Script projects (%d):\n", len(projects))
	for _, p := range projects {
		fmt.Printf("  • %s [%s]  parent=%s  updated=%s\n",
			p.Title, p.ScriptID, p.ParentID, p.UpdateTime)
	}
	return nil
}

func runScriptsFind(sc *ScriptClient, container string) error {
	p, err := sc.FindByContainer(container)
	if err != nil {
		return err
	}
	fmt.Printf("Bound script:\n")
	fmt.Printf("  title    : %s\n", p.Title)
	fmt.Printf("  scriptId : %s\n", p.ScriptID)
	fmt.Printf("  parentId : %s\n", p.ParentID)
	fmt.Printf("  creator  : %s <%s>\n", p.Creator.Name, p.Creator.Email)
	fmt.Printf("  updated  : %s\n", p.UpdateTime)
	return nil
}

func runScriptsPull(sc *ScriptClient, id, outDir, repoRoot string) error {
	mf, err := sc.PullToDir(id, outDir)
	if err != nil {
		return err
	}
	fmt.Printf("✅ pulled %d files to %s\n", len(mf.Files), outDir)
	for _, f := range mf.Files {
		fmt.Printf("  • %s  (%s, %d bytes, sha256=%s…)\n",
			f.Path, f.Type, f.SizeBytes, f.SHA256[:12])
	}
	audit(repoRoot, "", id, "scripts_pull", map[string]any{
		"out_dir": outDir, "files": len(mf.Files),
	}, len(mf.Files))
	return nil
}

func runScriptsPush(sc *ScriptClient, id, inDir string, confirm bool, repoRoot string) error {
	if pre, err := sc.GetContent(id); err == nil {
		preBytes, _ := json.Marshal(pre)
		_, _ = sc.CaptureForScript(repoRoot, id, "scripts-push-pre", preBytes)
	}
	mf, err := sc.PushFromDir(id, inDir, confirm)
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}
	audit(repoRoot, "", id, "scripts_push", map[string]any{
		"in_dir": inDir, "files": len(mf.Files),
	}, len(mf.Files))
	fmt.Printf("✅ applied changes; manifest updated\n")
	return nil
}

func runScriptsRun(sc *ScriptClient, id, fn, repoRoot string) error {
	result, err := sc.RunFunction(id, fn, nil)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	audit(repoRoot, "", id, "scripts_run", map[string]any{"function": fn}, 1)
	fmt.Printf("✅ function %s() returned:\n%s\n", fn, string(out))
	return nil
}

func runScriptsVersions(sc *ScriptClient, id string) error {
	vs, err := sc.ListVersions(id)
	if err != nil {
		return err
	}
	fmt.Printf("Versions of %s (%d):\n", id, len(vs))
	for _, v := range vs {
		fmt.Printf("  v%d  %s  — %s\n", v.VersionNumber, v.CreateTime, v.Description)
	}
	return nil
}

func runScriptsSnapshot(sc *ScriptClient, id, desc, repoRoot string) error {
	n, err := sc.CreateVersion(id, desc)
	if err != nil {
		return err
	}
	audit(repoRoot, "", id, "scripts_snapshot", map[string]any{
		"description": desc, "version": n,
	}, 1)
	fmt.Printf("✅ version v%d created\n", n)
	return nil
}

func runRollback(args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	id := flagValue(args, "--id", "")
	if id == "" {
		return fmt.Errorf("rollback: --id required")
	}
	return Rollback(repoRoot, id)
}
func runXlsxIn(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	path := flagValue(args, "--path", "")
	if tab == "" || path == "" {
		return fmt.Errorf("xlsx-in: --tab and --path required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sheets2, err := ReadXlsx(path)
	if err != nil {
		return err
	}
	if len(sheets2) == 0 {
		return fmt.Errorf("xlsx: empty file")
	}
	if _, err := cl.AddSheet(tab, nil); err != nil {
		_ = err
	}
	rows, err := XlsxToCells(sheets2[0])
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("xlsx: no rows")
	}
	a1 := fmt.Sprintf("%s!A1", quoted(tab))
	ur, err := cl.WriteValues(a1, rows, "USER_ENTERED")
	if err != nil {
		return err
	}
	audit(repoRoot, id, tab, "xlsx_in", map[string]any{"path": path, "rows": len(rows)}, ur.UpdatedCells)
	fmt.Printf("✅ imported %d rows (%d cells) from %s into %s\n", len(rows), ur.UpdatedCells, path, tab)
	return nil
}
func runXlsxOut(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	path := flagValue(args, "--path", "")
	if tab == "" || path == "" {
		return fmt.Errorf("xlsx-out: --tab and --path required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	t, err := cl.LoadTable(tab)
	if err != nil {
		return err
	}
	cells := []XlsxCell{}
	for i, h := range t.Header() {
		cells = append(cells, XlsxCell{Ref: fmt.Sprintf("%s%d", colLetter(i+1), 1), Val: h})
	}
	maxR := 1
	maxC := len(t.Header())
	for ri, r := range t.GetRows() {
		for ci, v := range r {
			cells = append(cells, XlsxCell{Ref: fmt.Sprintf("%s%d", colLetter(ci+1), ri+2), Val: v})
		}
		if ri+2 > maxR {
			maxR = ri + 2
		}
	}
	data, err := WriteXlsx(tab, cells, maxR, maxC)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	audit(repoRoot, id, tab, "xlsx_out", map[string]any{"path": path, "cells": len(cells)}, len(cells))
	fmt.Printf("✅ wrote %d cells to %s\n", len(cells), path)
	return nil
}

// parseKVArgs turns --key value pairs (or --json '{...}') into a map.
// `reserved` lists flags that belong to the command (not the data row)
// and should be ignored when harvesting key/value pairs.
func parseKVArgs(args []string, reserved ...string) (map[string]string, error) {
	out := map[string]string{}
	res := map[string]bool{}
	for _, r := range reserved {
		res[r] = true
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("--json requires a value")
			}
			var m map[string]string
			if err := jsonUnmarshalString(args[i], &m); err != nil {
				return nil, err
			}
			for k, v := range m {
				out[k] = v
			}
		default:
			if res[args[i]] {
				continue
			}
			if strings.HasPrefix(args[i], "--") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				out[strings.TrimPrefix(args[i], "--")] = args[i+1]
				i++
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no data key/value pairs found (use --key value or --json '{...}')")
	}
	return out, nil
}

// ── helpers ─────────────────────────────────────────────────────────
func pickSpreadsheetID(args []string) string {
	if v := flagValue(args, "--spreadsheet", ""); v != "" {
		return v
	}
	if v := os.Getenv("OVAV_SHEETS_SPREADSHEET"); v != "" {
		return v
	}
	return defaultSpreadsheetID
}
func flagValue(args []string, name, def string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return def
}
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// CRIT-004 surgical fix: prefer the OUTERMOST .ovav (repo root)
	// over nested .ovav caches (e.g. go-runtime/.ovav/ sync dirs).
	// Walk up to the git root first, then resolve .ovav from there.
	dir := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(dir + "/.git"); err == nil {
			// Found git root — use it as canonical repoRoot.
			cwd = dir
			break
		}
		parent := dir + "/.."
		if parent == dir {
			break
		}
		dir = parent
	}
	// Now resolve .ovav walking up from canonical repoRoot.
	dir = cwd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(dir + "/.ovav"); err == nil {
			return dir, nil
		}
		parent := dir + "/.."
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd, nil
}
func exitOn(err error) {
	if err != nil {
		die(err)
	}
}
func die(err error) {
	fmt.Fprintln(os.Stderr, "❌", err)
	os.Exit(1)
}
func storePath(repoRoot string) string {
	return repoRoot + "/.ovav/vault/credentials/google_oauth.enc"
}
func parseMatrix(s string) ([][]string, error) {
	// Delegate to json unmarshal — safer than ad-hoc parsing.
	var m [][]string
	if err := jsonUnmarshalString(s, &m); err != nil {
		return nil, fmt.Errorf("values-json: %w", err)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("values-json: empty matrix")
	}
	return m, nil
}

// openBrowser tries to launch a browser on the host OS. Best effort.
func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", u)
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// awaitCallback starts an HTTP server on the loopback address of
// redirectURI (e.g. 127.0.0.1:9876), waits up to timeout for Google to
// deliver ?code=…&state=…, validates state, and returns the code.
//
// On success it shows a friendly page in the user's browser and closes.
// On failure it returns a descriptive error.
func awaitCallback(redirectURI, expectedState string, timeout time.Duration) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("callback: bad redirect URI: %w", err)
	}
	host := u.Host
	if !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") {
		return "", fmt.Errorf("callback: refusing non-loopback redirect URI %q", redirectURI)
	}
	mux := http.NewServeMux()
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errParam := q.Get("error"); errParam != "" {
			errCh <- fmt.Errorf("google returned error=%s: %s", errParam, q.Get("error_description"))
			writeFriendlyPage(w, false, "Google denied consent: "+errParam)
			return
		}
		gotState := q.Get("state")
		if gotState != expectedState {
			errCh <- fmt.Errorf("state mismatch: got %q want %q", gotState, expectedState)
			writeFriendlyPage(w, false, "State mismatch — possible CSRF, refusing.")
			return
		}
		code := q.Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback: %s", r.URL.RawQuery)
			writeFriendlyPage(w, false, "No authorization code received.")
			return
		}
		writeFriendlyPage(w, true, "OVAV Sheets is connected. You can close this tab.")
		codeCh <- code
	})
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return "", fmt.Errorf("callback: listen %s: %w", host, err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	fmt.Printf("⏳ waiting up to %s for Google to redirect back to %s%s …\n",
		timeout, host, u.Path)
	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("callback: timeout after %s — user did not complete consent", timeout)
	}
}
func writeFriendlyPage(w http.ResponseWriter, ok bool, msg string) {
	bg := "#0f172a"
	card := "#1e293b"
	text := "#e2e8f0"
	accent := "#22c55e"
	if !ok {
		accent = "#ef4444"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html><head><meta charset="utf-8"><title>OVAV Sheets</title>
<style>
 body{margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;
     background:%s;color:%s;display:grid;place-items:center;height:100vh}
 .card{background:%s;padding:48px 64px;border-radius:16px;
       box-shadow:0 20px 60px rgba(0,0,0,.4);text-align:center;max-width:520px}
 h1{font-size:42px;margin:0 0 12px}
 p{font-size:16px;line-height:1.5;opacity:.85}
 .badge{display:inline-block;background:%s;color:#fff;padding:6px 14px;
        border-radius:999px;font-weight:600;margin-bottom:24px}
</style></head><body>
<div class="card">
 <div class="badge">%s</div>
 <h1>%s</h1>
 <p>%s</p>
</div></body></html>`,
		bg, text, card,
		accent,
		ternary(ok, "OVAV · CONNECTED", "OVAV · ERROR"),
		ternary(ok, "✓ Connected", "✗ Failed"),
		msg,
	)
}
func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// keep

// ── formatting / validation / protected ranges / charts / images ────

func runFormatCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	rng := flagValue(args, "--range", "")
	when := flagValue(args, "--when", "TEXT_CONTAINS")
	value := flagValue(args, "--value", "")
	bg := flagValue(args, "--bg", "FFEB9C")
	fg := flagValue(args, "--fg", "")
	bold := false
	for _, a := range args {
		if a == "--bold" {
			bold = true
		}
	}
	if rng == "" || value == "" {
		return fmt.Errorf("format: --range and --value required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.AddConditionalRule(ConditionalFormatRule{
		Range: rng, When: when, Value: value,
		BGColor: bg, FGColor: fg, Bold: bold,
	}); err != nil {
		return err
	}
	audit(repoRoot, id, "", "format_rule", map[string]any{
		"range": rng, "when": when, "value": value,
	}, 1)
	fmt.Printf("✅ format rule applied to %s (%s %s)\n", rng, when, value)
	return nil
}

func runValidateCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	rng := flagValue(args, "--range", "")
	ruleType := flagValue(args, "--type", "ONE_OF_LIST")
	strict := false
	var values any
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--values":
			if i+1 < len(args) {
				raw := args[i+1]
				// Detect JSON object (e.g. {min,max}) vs comma list
				if strings.HasPrefix(strings.TrimSpace(raw), "{") {
					var m map[string]any
					if err := json.Unmarshal([]byte(raw), &m); err == nil {
						values = m
					} else {
						values = raw
					}
				} else {
					list := strings.Split(raw, ",")
					values = list
				}
				i++
			}
		case "--values-json":
			if i+1 < len(args) {
				var v any
				if err := json.Unmarshal([]byte(args[i+1]), &v); err == nil {
					values = v
				}
				i++
			}
		case "--strict":
			strict = true
		}
	}
	if rng == "" {
		return fmt.Errorf("validate: --range required")
	}
	if ruleType == "ONE_OF_LIST" && values == nil {
		return fmt.Errorf("validate: --values required for ONE_OF_LIST")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.SetDataValidation(rng, ruleType, values, strict); err != nil {
		return err
	}
	audit(repoRoot, id, "", "data_validation", map[string]any{
		"range": rng, "type": ruleType, "values": values,
	}, 1)
	fmt.Printf("✅ validation %s on %s\n", ruleType, rng)
	return nil
}

func runProtectCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	rng := flagValue(args, "--range", "")
	desc := flagValue(args, "--description", "OVAV protected")
	warning := false
	for _, a := range args {
		if a == "--warning-only" {
			warning = true
		}
	}
	var editors []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--editor" && i+1 < len(args) {
			editors = append(editors, args[i+1])
			i++
		}
	}
	if rng == "" {
		return fmt.Errorf("protect: --range required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.AddProtectedRange(rng, desc, editors, warning); err != nil {
		return err
	}
	audit(repoRoot, id, "", "protect", map[string]any{
		"range": rng, "editors": editors, "warning_only": warning,
	}, 1)
	fmt.Printf("✅ protected %s (editors=%v, warning_only=%v)\n", rng, editors, warning)
	return nil
}

func runNameRangeCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	name := flagValue(args, "--name", "")
	rng := flagValue(args, "--range", "")
	if name == "" || rng == "" {
		return fmt.Errorf("name-range: --name and --range required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.AddNamedRange(name, rng); err != nil {
		return err
	}
	audit(repoRoot, id, "", "name_range", map[string]any{
		"name": name, "range": rng,
	}, 1)
	fmt.Printf("✅ named range %q → %s\n", name, rng)
	return nil
}

func runChartCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	rng := flagValue(args, "--range", "")
	rngDomain := flagValue(args, "--range-domain", "")
	rngSeries := flagValue(args, "--range-series", "")
	chartType := flagValue(args, "--type", "BAR")
	title := flagValue(args, "--title", "Chart")
	if tab == "" || rng == "" {
		return fmt.Errorf("chart: --tab and --range required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.AddChart(tab, rng, rngDomain, rngSeries, chartType, title); err != nil {
		return err
	}
	audit(repoRoot, id, tab, "chart", map[string]any{
		"type": chartType, "range": rng, "title": title,
	}, 1)
	fmt.Printf("✅ %s chart added to %s (%s)\n", chartType, tab, rng)
	return nil
}

func runImageCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	url := flagValue(args, "--url", "")
	anchor := flagValue(args, "--anchor", "A1")
	if tab == "" || url == "" {
		return fmt.Errorf("image: --tab and --url required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.InsertImage(tab, anchor, url); err != nil {
		return err
	}
	audit(repoRoot, id, tab, "image", map[string]any{
		"url": url, "anchor": anchor,
	}, 1)
	fmt.Printf("✅ image added to %s @ %s\n", tab, anchor)
	return nil
}

// ── Triggers + Properties via Apps Script wrapper ────────────────────

func runTriggerCmd(repoRoot string, args []string) error {
	if len(args) == 0 {
		return runTriggersHelp()
	}
	id := flagValue(args, "--id", "")
	if id == "" {
		// Pull from script_id via diagnostics
		id = "1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzra6n3D_6XRvMOWda4HQpg"
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	sc := NewScriptClient(creds)
	switch args[0] {
	case "list":
		out, err := sc.RunFunction(id, "_ovavListTriggers", nil)
		if err != nil {
			return err
		}
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	case "install":
		out, err := sc.RunFunction(id, "_ovavInstallTriggers", nil)
		if err != nil {
			return err
		}
		audit(repoRoot, "", id, "triggers_install", nil, 0)
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	case "delete":
		handler := flagValue(args, "--handler", "")
		if handler == "" {
			return fmt.Errorf("trigger delete: --handler required")
		}
		out, err := sc.RunFunction(id, "_ovavDeleteTrigger", map[string]any{"handlerName": handler})
		if err != nil {
			return err
		}
		audit(repoRoot, "", id, "trigger_delete", map[string]any{"handler": handler}, 0)
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	default:
		return runTriggersHelp()
	}
	return nil
}

func runTriggersHelp() error {
	fmt.Fprintln(os.Stderr, `trigger subcommands:
  list                List installed triggers (via _ovavListTriggers)
  install             Install dailyClose (via _ovavInstallTriggers)
  delete --handler X  Delete trigger by handler name`)
	return nil
}

func runPropsCmd(repoRoot string, args []string) error {
	if len(args) == 0 {
		return runPropsHelp()
	}
	id := flagValue(args, "--id", "")
	if id == "" {
		id = "1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzra6n3D_6XRvMOWda4HQpg"
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	sc := NewScriptClient(creds)
	switch args[0] {
	case "list":
		out, err := sc.RunFunction(id, "_ovavGetProps", nil)
		if err != nil {
			return err
		}
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	case "get":
		key := flagValue(args, "--key", "")
		if key == "" {
			return fmt.Errorf("props get: --key required")
		}
		out, err := sc.RunFunction(id, "_ovavGetProps", map[string]any{"key": key})
		if err != nil {
			return err
		}
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	case "set":
		key := flagValue(args, "--key", "")
		value := flagValue(args, "--value", "")
		scope := flagValue(args, "--scope", "script")
		if key == "" {
			return fmt.Errorf("props set: --key and --value required")
		}
		out, err := sc.RunFunction(id, "_ovavSetProp", map[string]any{
			"key": key, "value": value, "scope": scope,
		})
		if err != nil {
			return err
		}
		audit(repoRoot, "", id, "props_set", map[string]any{"key": key, "scope": scope}, 0)
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	case "delete":
		key := flagValue(args, "--key", "")
		scope := flagValue(args, "--scope", "script")
		if key == "" {
			return fmt.Errorf("props delete: --key required")
		}
		out, err := sc.RunFunction(id, "_ovavDeleteProp", map[string]any{
			"key": key, "scope": scope,
		})
		if err != nil {
			return err
		}
		audit(repoRoot, "", id, "props_delete", map[string]any{"key": key, "scope": scope}, 0)
		text, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(text))
	default:
		return runPropsHelp()
	}
	return nil
}

func runPropsHelp() error {
	fmt.Fprintln(os.Stderr, `props subcommands:
  list                       List all script properties
  get --key X                Get one property
  set --key X --value Y      Set one property (--scope script|user|document)
  delete --key X             Delete property`)
	return nil
}

// ── OVAV bootstrap ────────────────────────────────────────────────────

func runOvavInstall(repoRoot string, args []string) error {
	id := flagValue(args, "--id", "")
	sid := flagValue(args, "--spreadsheet", "")
	if id == "" {
		id = "1fuA7kJHkWY6_4rM-IQHO9by6OLqVGZMDsSzra6n3D_6XRvMOWda4HQpg"
	}
	if sid == "" {
		sid = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	sc := NewScriptClient(creds)
	out, err := sc.RunFunction(id, "_ovavInstallCimaById", map[string]any{
		"spreadsheetId": sid,
	})
	if err != nil {
		return err
	}
	audit(repoRoot, sid, "", "ovav_install", map[string]any{"script_id": id}, 1)
	text, _ := json.MarshalIndent(out, "", "  ")
	fmt.Printf("✅ CIMA instalado via wrapper:\n%s\n", string(text))
	return nil
}

func runDebug() error {
	store := NewCredStore("/home/braka/Systems/ovav/.ovav/worktrees/feat-sheets-mcp")
	c, err := store.Load()
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	fmt.Println("token:", c.AccessToken)
	return nil
}

// runDebugToken prints the current access token, its scope, and the
// remaining TTL. Used for diagnosing OAuth scope issues.
func runDebugToken(repoRoot string) error {
	store := NewCredStore(repoRoot)
	c, err := store.Load()
	if err != nil {
		return err
	}
	if err := EnsureFresh(c); err != nil {
		return err
	}
	fmt.Printf("access_token: %s\n", c.AccessToken)
	fmt.Printf("scope:        %s\n", c.Scope)
	fmt.Printf("expires:      %d\n", c.TokenExpiry)
	return nil
}

// runAuthURL generates an OAuth consent URL with the requested scopes and
// prints it. Does NOT exchange — the user opens the URL in a browser,
// authorizes, and the redirect URL contains the auth code (which the
// user pastes into the regular `auth --code <code>` flow).
//
// Usage:
//	ovav-sheets auth-url --scopes "sheets,script,drive-file"
func runAuthURL(repoRoot string, args []string) error {
	scopesArg := flagValue(args, "--scopes", "sheets,script,drive-file")
	state := fmt.Sprintf("ovav-%d", time.Now().Unix())
	store := NewCredStore(repoRoot)
	c, err := store.Load()
	if err != nil {
		return err
	}
	scopes := resolveScopes(scopesArg)
	url := AuthCodeURL(c.ClientID, defaultRedirectURI, state, scopes)
	fmt.Println(url)
	return nil
}

// runAuthExchange trades an OAuth code captured from the browser for new
// tokens with the requested scopes. Uses the existing client_id/secret
// from the vault (no need to re-pass them).
//
// Usage:
//	ovav-sheets auth-exchange --code "4/0AX..." --scopes "sheets,script,drive-file"
func runAuthExchange(repoRoot string, args []string) error {
	code := flagValue(args, "--code", "")
	scopesArg := flagValue(args, "--scopes", "sheets,script,drive-file")
	if code == "" {
		return fmt.Errorf("auth-exchange: --code required")
	}
	store := NewCredStore(repoRoot)
	c, err := store.Load()
	if err != nil {
		return err
	}
	scopes := resolveScopes(scopesArg)
	_ = scopes
	creds, err := ExchangeCode(c.ClientID, c.ClientSecret, defaultRedirectURI, code)
	if err != nil {
		return err
	}
	// Merge: keep client_id, client_secret, refresh_token (overwrite),
	// access_token, scope, token_expiry_unix.
	newCreds := &Creds{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		Scope:        creds.Scope,
		TokenExpiry:  creds.TokenExpiry,
	}
	if err := store.Save(newCreds); err != nil {
		return fmt.Errorf("auth-exchange: save: %w", err)
	}
	fmt.Printf("✅ re-consent complete\n")
	fmt.Printf("   new scopes: %s\n", creds.Scope)
	fmt.Printf("   access_token expires: %d\n", creds.TokenExpiry)
	audit(repoRoot, "", "", "auth_reconsent", map[string]any{"scopes": creds.Scope}, 1)
	return nil
}

// runRestoreFormulasCmd restores the VLOOKUP formulas in NOMBRE, HORARIO,
// TURNO_ID columns and the auto REQUEST_ID formula on active rows of
// OPERACIONES. Use when an accidental edit broke a formula.
//
// Usage:
//	ovav-sheets restore-formulas --tab OPERACIONES
//	ovav-sheets restore-formulas --tab OPERACIONES --range B4:M10
func runRestoreFormulasCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	rng := flagValue(args, "--range", "")
	if tab == "" {
		return fmt.Errorf("restore-formulas: --tab required (e.g. --tab OPERACIONES)")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}

	// Default range: B4:M10 (active rows on OPERACIONES)
	if rng == "" {
		rng = "B4:M10"
	}
	startCol, startRow, endCol, endRow, err := parseA1Range(sng(rng))
	if err != nil {
		return fmt.Errorf("restore-formulas: bad range %q: %w", rng, err)
	}

	// Snapshot before mutation
	pre, _ := cl.ReadRange(rng)
	if pre != nil {
		preBytes, _ := json.Marshal(pre)
		if rec, err := cl.Capture(repoRoot, "restore-formulas-pre",
			fmt.Sprintf("pre-restore %s", rng), tab, preBytes); err == nil {
			_ = rec
		}
	}

	// Build formulas row-by-row. We assume the canonical OPERACIONES
	// layout: A=DNI, B=NOMBRE (VLOOKUP col 2), E=HORARIO (col 4),
	// F=TURNO_ID (col 9), M=REQUEST_ID.
	type cellFormula struct {
		col, row int
		formula  string
	}
	var updates []cellFormula
	for r := startRow + 1; r <= endRow+1; r++ {
		// B: NOMBRE
		if colInRange(2, startCol, endCol) {
			updates = append(updates, cellFormula{
				col: 2, row: r,
				formula: fmt.Sprintf(`=IF(A%d="","",IFERROR(VLOOKUP(A%d,PERSONAL!$B$3:$O$27,2,FALSE),""))`, r, r),
			})
		}
		// E: HORARIO
		if colInRange(5, startCol, endCol) {
			updates = append(updates, cellFormula{
				col: 5, row: r,
				formula: fmt.Sprintf(`=IF(A%d="","",IFERROR(VLOOKUP(A%d,PERSONAL!$B$3:$O$27,4,FALSE),""))`, r, r),
			})
		}
		// F: TURNO_ID
		if colInRange(6, startCol, endCol) {
			updates = append(updates, cellFormula{
				col: 6, row: r,
				formula: fmt.Sprintf(`=IF(A%d="","",IFERROR(VLOOKUP(A%d,PERSONAL!$B$3:$O$27,9,FALSE),""))`, r, r),
			})
		}
		// M: REQUEST_ID
		if colInRange(13, startCol, endCol) {
			updates = append(updates, cellFormula{
				col: 13, row: r,
				formula: fmt.Sprintf(`="REQ-"&TEXT(ROW()-3,"000")`),
			})
		}
	}

	// Send update requests in a single batchUpdate.
	var requests []map[string]any
	for _, u := range updates {
		requests = append(requests, map[string]any{
			"updateCells": map[string]any{
				"start": map[string]any{
					"sheetId":     sid,
					"rowIndex":    u.row - 1,
					"columnIndex": u.col - 1,
				},
				"rows": []map[string]any{
					{
						"values": []map[string]any{
							{"userEnteredValue": map[string]any{"formulaValue": u.formula}},
						},
					},
				},
				"fields": "userEnteredValue",
			},
		})
	}
	body, _ := json.Marshal(map[string]any{"requests": requests})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("restore-formulas: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("restore-formulas HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "restore_formulas", map[string]any{"range": rng, "count": len(updates)}, len(updates))
	fmt.Printf("✅ restored %d formulas in %s on %s\n", len(updates), rng, tab)
	return nil
}

// runRestoreFormatCmd re-applies all styles, conditional formatting,
// and data validations to the canonical OPERACIONES layout. Use after
// bulk edits break the visual hierarchy.
//
// Usage:
//	ovav-sheets restore-format --tab OPERACIONES
func runRestoreFormatCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("restore-format: --tab required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if _, err := cl.resolveSheetID(tab); err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}

	// Re-apply styles in canonical order
	cmds := []struct {
		rangeStr string
		bg       string
		fg       string
		bold     bool
		align    string
		fontSize int
		border   bool
	}{
		{"A3:M3", "404040", "FFFFFF", true, "CENTER", 10, true},  // header
		{"A4:A10", "E7F0FF", "", true, "CENTER", 0, true},        // DNI
		{"B4:B10", "F0F0F0", "", false, "LEFT", 9, true},        // NOMBRE (auto)
		{"C4:D10", "FFF4D6", "", false, "CENTER", 0, true},      // TIPO + FECHA_OP (input)
		{"E4:F10", "F0F0F0", "", false, "LEFT", 9, true},        // HORARIO + TURNO_ID (auto)
		{"G4:K10", "FFF4D6", "", false, "LEFT", 0, true},        // SUPERVISOR + NUEVO + HORAS + PRE_POST + MOTIVO (input)
		{"L4:L10", "FFF4D6", "", true, "CENTER", 0, true},       // ESTADO_GTR (dropdown)
		{"M4:M10", "E7F0FF", "", true, "CENTER", 0, true},       // REQUEST_ID
	}
	for _, c := range cmds {
		spec := CellStyleSpec{
			Range:        tab + "!" + c.rangeStr,
			BgColor:      c.bg,
			FgColor:      c.fg,
			Bold:         c.bold,
			Align:        strings.ToUpper(c.align),
			FontSize:     c.fontSize,
			BorderAll:    c.border,
			BorderColor:  "C0C0C0",
		}
		if err := cl.SetCellStyle(spec); err != nil {
			return fmt.Errorf("restore-format: %w", err)
		}
	}
	audit(repoRoot, id, tab, "restore_format", map[string]any{}, 0)
	fmt.Printf("✅ restored format on %s (header + 7 active rows)\n", tab)
	return nil
}

// runCleanEmptyRowsCmd deletes fully-empty rows in OPERACIONES (rows 4-100).
// Skips rows that have any content.
//
// Usage:
//	ovav-sheets clean-empty-rows --tab OPERACIONES
func runCleanEmptyRowsCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("clean-empty-rows: --tab required")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	// Scan rows 4-100 (1-based = 3-99 indices)
	empty := []int{}
	for row := 4; row <= 100; row++ {
		rng := fmt.Sprintf("A%d:N%d", row, row)
		vr, err := cl.ReadRange(rng)
		if err != nil {
			continue
		}
		if len(vr.Values) == 0 || allEmpty2D(vr.Values) {
			empty = append(empty, row-1) // 0-based
		}
	}
	if len(empty) == 0 {
		fmt.Printf("✅ no empty rows in %s\n", tab)
		return nil
	}
	// Delete from bottom to top
	for i := len(empty) - 1; i >= 0; i-- {
		req := map[string]any{
			"deleteDimension": map[string]any{
				"range": map[string]any{
					"sheetId":    sid,
					"dimension":  "ROWS",
					"startIndex": empty[i],
					"endIndex":   empty[i] + 1,
				},
			},
		}
		body, _ := json.Marshal(map[string]any{"requests": []any{req}})
		u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
		httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := cl.http.Do(httpReq)
		if err != nil {
			return fmt.Errorf("clean-empty-rows: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("clean-empty-rows HTTP %d", resp.StatusCode)
		}
	}
	audit(repoRoot, id, tab, "clean_empty_rows", map[string]any{"count": len(empty)}, len(empty))
	fmt.Printf("✅ deleted %d empty rows in %s\n", len(empty), tab)
	return nil
}

func allEmpty2D(values [][]string) bool {
	for _, row := range values {
		for _, cell := range row {
			if cell != "" {
				return false
			}
		}
	}
	return true
}

func colInRange(col, startCol, endCol int) bool {
	return col >= startCol && col <= endCol
}

// runFormatPremiumCmd applies the OVAV "premium" date format to a single
// cell or range. Designed to restore the format that Sheets' native
// datepicker strips on manual edit.
//
//	--tab         tab name (required)
//	--cell        A1 range (e.g. "G2" or "B4:B10") (required)
//
// Format applied:
//	- numberFormat: DATE pattern "dddd, d mmm yyyy"
//	- backgroundColor: #1A1A2E (OVAV navy)
//	- textFormat: foregroundColor #FFFFFF, fontFamily Google Sans Mono,
//	              fontSize 10, bold false
//	- horizontalAlignment: CENTER, verticalAlignment: MIDDLE
func runFormatPremiumCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	rng := flagValue(args, "--cell", "")
	if tab == "" || rng == "" {
		return fmt.Errorf("format-premium: --tab and --cell required (e.g. --tab OPERACIONES --cell G2)")
	}
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	startCol, startRow, endCol, endRow, err := parseA1Range(sng(rng))
	if err != nil {
		return fmt.Errorf("format-premium: bad cell %q: %w", rng, err)
	}
	req := map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sid,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow + 1,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol + 1,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"numberFormat": map[string]any{
						"type":    "DATE",
						"pattern": "dddd, d mmm yyyy",
					},
					"backgroundColor": map[string]any{
						"red":   0.101960786,
						"green": 0.101960786,
						"blue":  0.18039216,
					},
					"textFormat": map[string]any{
						"foregroundColor": map[string]any{
							"red":   1.0,
							"green": 1.0,
							"blue":  1.0,
						},
						"fontFamily": "Google Sans Mono",
						"fontSize":   10.0,
						"bold":       false,
					},
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(numberFormat,backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("format-premium: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("format-premium HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "format_premium", map[string]any{"cell": rng}, 1)
	fmt.Printf("✅ premium format applied to %s on %s\n", rng, tab)
	return nil
}

// sng ensures a single-cell A1 range ("G2") is normalized to "G2:G2"
// so parseA1Range can handle both forms.
func sng(rng string) string {
	if !strings.Contains(rng, ":") {
		return rng + ":" + rng
	}
	return rng
}

// runStyleCmd applies plain cell formatting (bg, fg, bold, borders, align).
// Designed to give CONFIG (and other operational tabs) a real visual
// hierarchy: section headers in dark navy, editable cells in cream,
// readonly/calculated in light gray, with borders between blocks.
//
// Usage:
//   ovav-sheets style --range "CONFIG!A3:E3" --bg "1F4E78" --fg "FFFFFF" --bold --border-all --border-color "1F4E78"
//   ovav-sheets style --range "PERSONAL!B3:B27" --bg "FFF4D6"  // editable flag cells
func runStyleCmd(repoRoot string, args []string) error {
	if err := requireArg(args, "--range"); err != nil {
		return err
	}
	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	spec := CellStyleSpec{
		Range: flagValue(args, "--range", ""),
	}
	if v := flagValue(args, "--bg", ""); v != "" {
		spec.BgColor = strings.TrimPrefix(v, "#")
	}
	if v := flagValue(args, "--fg", ""); v != "" {
		spec.FgColor = strings.TrimPrefix(v, "#")
	}
	if v := flagValue(args, "--border-color", ""); v != "" {
		spec.BorderColor = strings.TrimPrefix(v, "#")
	}
	spec.Bold = hasFlag(args, "--bold")
	spec.Italic = hasFlag(args, "--italic")
	if v := flagValue(args, "--font-size", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			spec.FontSize = n
		}
	}
	spec.Align = flagValue(args, "--align", "")
	spec.BorderAll = hasFlag(args, "--border-all")
	spec.BorderTop = hasFlag(args, "--border-top")
	spec.BorderBottom = hasFlag(args, "--border-bottom")
	spec.BorderLeft = hasFlag(args, "--border-left")
	spec.BorderRight = hasFlag(args, "--border-right")
	spec.NumberFormat = flagValue(args, "--number-format", "")
	spec.Wrap = hasFlag(args, "--wrap")
	if err := cl.SetCellStyle(spec); err != nil {
		return err
	}
	audit(repoRoot, id, "", "style_cells", map[string]any{
		"range": spec.Range, "bg": spec.BgColor, "fg": spec.FgColor,
		"bold": spec.Bold, "italic": spec.Italic, "border": spec.BorderAll,
	}, 1)
	fmt.Printf("✅ style applied to %s (bg=%s fg=%s bold=%v border=%v)\n",
		spec.Range, spec.BgColor, spec.FgColor, spec.Bold, spec.BorderAll)
	return nil
}


// runResizeColumnsCmd sets pixel widths for columns.
// Usage: sheets resize-columns --tab T --widths "0=50,1=100,2=350"
// (colIdx=pixelSize, comma-separated pairs)
func runResizeColumnsCmd(repoRoot string, args []string) error {
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("resize-columns: --tab required")
	}
	widthsRaw := flagValue(args, "--widths", "")
	if widthsRaw == "" {
		return fmt.Errorf("resize-columns: --widths required (e.g. '0=50,1=100')")
	}

	widths := map[int]int{}
	for _, pair := range strings.Split(widthsRaw, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid widths pair %q (expected colIdx=pixelSize)", pair)
		}
		colIdx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return fmt.Errorf("invalid column index in %q: %v", pair, err)
		}
		w, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return fmt.Errorf("invalid pixel size in %q: %v", pair, err)
		}
		widths[colIdx] = w
	}

	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.SetColumnWidths(tab, widths); err != nil {
		return fmt.Errorf("resize-columns: %w", err)
	}
	audit(repoRoot, id, "", "resize_columns", map[string]any{
		"tab": tab, "widths": widths,
	}, 1)
	fmt.Printf("✅ set %d column widths on %s\n", len(widths), tab)
	for colIdx, w := range widths {
		fmt.Printf("   col %d → %dpx\n", colIdx, w)
	}
	return nil
}

// runFreezeCmd freezes rows/columns in a tab.
// Usage: sheets freeze --tab T --rows N --cols N
func runFreezeCmd(repoRoot string, args []string) error {
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("freeze: --tab required")
	}
	rows, _ := strconv.Atoi(flagValue(args, "--rows", "0"))
	cols, _ := strconv.Atoi(flagValue(args, "--cols", "0"))
	if rows < 0 || cols < 0 {
		return fmt.Errorf("freeze: --rows and --cols must be >= 0")
	}
	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.SetFreezePanes(tab, rows, cols); err != nil {
		return fmt.Errorf("freeze: %w", err)
	}
	audit(repoRoot, id, "", "freeze_panes", map[string]any{
		"tab": tab, "rows": rows, "cols": cols,
	}, 1)
	fmt.Printf("✅ froze %d rows + %d cols on %s\n", rows, cols, tab)
	return nil
}

// runExpandGridCmd extends a tab's grid by N columns (or rows).
// Usage: sheets expand-grid --tab T --cols 1
func runExpandGridCmd(repoRoot string, args []string) error {
	tab := flagValue(args, "--tab", "")
	if tab == "" {
		return fmt.Errorf("expand-grid: --tab required")
	}
	cols, _ := strconv.Atoi(flagValue(args, "--cols", "0"))
	if cols == 0 {
		return fmt.Errorf("expand-grid: --cols required (e.g. --cols 1)")
	}
	if cols < 0 {
		return fmt.Errorf("expand-grid: --cols must be positive")
	}
	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)
	if err := cl.ExpandGrid(tab, cols); err != nil {
		return fmt.Errorf("expand-grid: %w", err)
	}
	audit(repoRoot, id, "", "expand_grid", map[string]any{
		"tab": tab, "cols": cols,
	}, 1)
	fmt.Printf("✅ expanded %s grid by %d cols\n", tab, cols)
	return nil
}

// requireArg returns a missing-arg error if the named flag is absent.
func requireArg(args []string, flag string) error {
	for _, a := range args {
		if a == flag {
			return nil
		}
	}
	return fmt.Errorf("%s required", flag)
}

// hasFlag returns true if the named flag is present in args.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}



// runDeleteColumnCmd removes columns from a tab using the Sheets API
// deleteDimension request. Accepts column letter OR numeric (1-based) index.
func runDeleteColumnCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	startCol := flagValue(args, "--start-col", "")
	endCol := flagValue(args, "--end-col", "")
	if tab == "" {
		return fmt.Errorf("delete-column: --tab required")
	}
	if startCol == "" || endCol == "" {
		return fmt.Errorf("delete-column: --start-col and --end-col required")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	startIdx, err1 := colRefToIndex(startCol)
	endIdx, err2 := colRefToIndex(endCol)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("delete-column: bad column ref: %v %v", err1, err2)
	}
	if startIdx > endIdx {
		startIdx, endIdx = endIdx, startIdx
	}
	req := map[string]any{
		"deleteDimension": map[string]any{
			"range": map[string]any{
				"sheetId":          sid,
				"dimension":         "COLUMNS",
				"startIndex":       startIdx,
				"endIndex":         endIdx + 1,
			},
		},
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("delete-column: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete-column HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "delete_column", map[string]any{
		"start_col": startCol, "end_col": endCol, "startIndex": startIdx, "endIndex": endIdx + 1,
	}, int(endIdx-startIdx+1))
	fmt.Printf("✅ deleted %d columns (%s to %s) on %s\n", endIdx-startIdx+1, startCol, endCol, tab)
	return nil
}

func colRefToIndex(ref string) (int, error) {
	ref = strings.TrimSpace(strings.ToUpper(ref))
	if ref == "" {
		return 0, fmt.Errorf("empty ref")
	}
	// Try numeric 0-based index (e.g. "5" → 5)
	if n, err := strconv.Atoi(ref); err == nil {
		if n < 0 {
			return 0, fmt.Errorf("negative")
		}
		return n, nil
	}
	// Letter-based 1-based → 0-based (A=1 → 0)
	col := 0
	for _, c := range ref {
		if c < 'A' || c > 'Z' {
			return 0, fmt.Errorf("non-letter: %c", c)
		}
		col = col*26 + int(c-'A'+1)
	}
	return col - 1, nil
}

// runDeleteRowCmd removes rows from a tab using the Sheets API
// deleteDimension request with dimension=ROWS. Row refs are 1-based
// for human ergonomics (matching Sheets UI row numbers).
//
// Auto-orders start<->end so a single delete request covers the range,
// and captures a snapshot before mutation for rollback safety.
func runDeleteRowCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	startRow := flagValue(args, "--start-row", "")
	endRow := flagValue(args, "--end-row", "")
	if tab == "" {
		return fmt.Errorf("delete-row: --tab required")
	}
	if startRow == "" || endRow == "" {
		return fmt.Errorf("delete-row: --start-row and --end-row required")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	// 1-based human refs → 0-based API indices
	startIdx, err1 := strconv.Atoi(strings.TrimSpace(startRow))
	endIdx, err2 := strconv.Atoi(strings.TrimSpace(endRow))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("delete-row: bad row ref (want 1-based integer): %v %v", err1, err2)
	}
	if startIdx < 1 || endIdx < 1 {
		return fmt.Errorf("delete-row: rows are 1-based; got start=%d end=%d", startIdx, endIdx)
	}
	start0 := startIdx - 1
	end0 := endIdx // endIndex in API is exclusive; 1-based end=N means delete up to row N inclusive → endIndex = N
	if start0 > end0 {
		start0, end0 = end0, start0
		startIdx, endIdx = endIdx, startIdx
	}
	// Safety: capture snapshot of the range BEFORE deletion.
	rangeA1 := fmt.Sprintf("%s!A%d:%s", tab, start0+1, lastColLetter())
	pre, preErr := cl.ReadRange(rangeA1)
	if preErr == nil && pre != nil {
		preBytes, _ := json.Marshal(pre)
		if rec, err := cl.Capture(repoRoot, "delete-row-pre",
			fmt.Sprintf("pre-delete rows %d-%d", startIdx, endIdx), tab, preBytes); err == nil {
			_ = rec
		}
	}
	req := map[string]any{
		"deleteDimension": map[string]any{
			"range": map[string]any{
				"sheetId":    sid,
				"dimension": "ROWS",
				"startIndex": start0,
				"endIndex":   end0,
			},
		},
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("delete-row: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete-row HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "delete_row", map[string]any{
		"start_row": startIdx, "end_row": endIdx, "startIndex": start0, "endIndex": end0,
	}, end0-start0)
	fmt.Printf("✅ deleted %d rows (%d to %d) on %s\n", end0-start0, startIdx, endIdx, tab)
	return nil
}

// lastColLetter returns "Z" — Sheets API supports columns beyond Z but
// delete-row only needs to capture a wide enough range for snapshot
// before-state. 26 columns is plenty for snapshot payload.
func lastColLetter() string {
	return "Z"
}

// runMergeCellsCmd merges the cells in the given A1 range using
// Sheets API mergeCells with mergeType=MERGE_ALL (drops all values
// except top-left). Captures a snapshot of the range before mutation.
//
// Usage: ovav-sheets merge-cells --tab OPERACIONES --range "A2:B2"
func runMergeCellsCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	rng := flagValue(args, "--range", "")
	if tab == "" || rng == "" {
		return fmt.Errorf("merge-cells: --tab and --range required (e.g. --tab OPERACIONES --range \"A2:B2\")")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	// Parse A1 range into sheetId + startRow/startCol/endRow/endCol
	startCol, startRow, endCol, endRow, err := parseA1Range(rng)
	if err != nil {
		return fmt.Errorf("merge-cells: bad range %q: %w", rng, err)
	}
	// Snapshot pre-merge
	pre, _ := cl.ReadRange(rng)
	if pre != nil {
		preBytes, _ := json.Marshal(pre)
		if rec, err := cl.Capture(repoRoot, "merge-cells-pre",
			fmt.Sprintf("pre-merge %s", rng), tab, preBytes); err == nil {
			_ = rec
		}
	}
	req := map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sid,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow + 1,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol + 1,
			},
			"mergeType": "MERGE_ALL",
		},
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("merge-cells: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("merge-cells HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "merge_cells", map[string]any{
		"range": rng, "startRow": startRow, "endRow": endRow,
		"startColumn": startCol, "endColumn": endCol,
	}, 1)
	fmt.Printf("✅ merged %s on %s\n", rng, tab)
	return nil
}

// runUnmergeCellsCmd reverses a previous merge on the given range.
// Captures a snapshot first so the operation is reversible.
func runUnmergeCellsCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	rng := flagValue(args, "--range", "")
	if tab == "" || rng == "" {
		return fmt.Errorf("unmerge-cells: --tab and --range required")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	startCol, startRow, endCol, endRow, err := parseA1Range(rng)
	if err != nil {
		return fmt.Errorf("unmerge-cells: bad range %q: %w", rng, err)
	}
	req := map[string]any{
		"unmergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sid,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow + 1,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol + 1,
			},
		},
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unmerge-cells: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unmerge-cells HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	fmt.Printf("✅ unmerged %s on %s\n", rng, tab)
	return nil
}

// parseA1Range parses "A2:B2" → (startCol, startRow, endCol, endRow) using
// 0-based indices expected by the Sheets API.
func parseA1Range(rng string) (startCol, startRow, endCol, endRow int, err error) {
	parts := strings.Split(rng, ":")
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("expected A1:B1 form")
	}
	sc, sr, err := parseA1Cell(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, 0, 0, err
	}
	ec, er, err := parseA1Cell(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return sc, sr, ec, er, nil
}

// runInsertRowCmd inserts blank rows using the Sheets API insertDimension
// request with dimension=ROWS. Row refs are 1-based (matching Sheets UI).
//
// Captures a snapshot of the affected range BEFORE insertion so the
// operation is fully reversible via ovav-sheets rollback.
func runInsertRowCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	tab := flagValue(args, "--tab", "")
	startRow := flagValue(args, "--start-row", "")
	endRow := flagValue(args, "--end-row", "")
	inheritFrom := flagValue(args, "--inherit-from", "") // optional: 1-based row whose values/formats the new rows copy
	if tab == "" {
		return fmt.Errorf("insert-row: --tab required")
	}
	if startRow == "" || endRow == "" {
		return fmt.Errorf("insert-row: --start-row and --end-row required")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	sid, err := cl.resolveSheetID(tab)
	if err != nil {
		return fmt.Errorf("resolveSheetID: %w", err)
	}
	startIdx, err1 := strconv.Atoi(strings.TrimSpace(startRow))
	endIdx, err2 := strconv.Atoi(strings.TrimSpace(endRow))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("insert-row: bad row ref (want 1-based integer): %v %v", err1, err2)
	}
	if startIdx < 1 || endIdx < 1 {
		return fmt.Errorf("insert-row: rows are 1-based; got start=%d end=%d", startIdx, endIdx)
	}
	start0 := startIdx - 1
	end0 := endIdx // API endIndex is exclusive
	if start0 > end0 {
		start0, end0 = end0, start0
		startIdx, endIdx = endIdx, startIdx
	}
	// Capture pre-insertion snapshot of the surrounding range.
	rangeA1 := fmt.Sprintf("%s!A%d:%s", tab, start0+1, lastColLetter())
	pre, preErr := cl.ReadRange(rangeA1)
	if preErr == nil && pre != nil {
		preBytes, _ := json.Marshal(pre)
		if rec, err := cl.Capture(repoRoot, "insert-row-pre",
			fmt.Sprintf("pre-insert rows %d-%d", startIdx, endIdx), tab, preBytes); err == nil {
			_ = rec
		}
	}
	dimRange := map[string]any{
		"sheetId":    sid,
		"dimension": "ROWS",
		"startIndex": start0,
		"endIndex":   end0,
	}
	req := map[string]any{"insertDimension": map[string]any{"range": dimRange}}
	if inheritFrom != "" {
		inhIdx, ierr := strconv.Atoi(strings.TrimSpace(inheritFrom))
		if ierr != nil || inhIdx < 1 {
			return fmt.Errorf("insert-row: --inherit-from must be 1-based integer, got %q", inheritFrom)
		}
		req = map[string]any{
			"insertDimension": map[string]any{
				"range":             dimRange,
				"inheritFromBefore": inhIdx > 1, // true = copy from row above; false = from row below
			},
		}
		// For our use case (inserting above header to add a title row)
		// we want the new row to be BLANK with default formatting, so
		// we omit inheritFrom unless user explicitly asks.
	}
	body, _ := json.Marshal(map[string]any{"requests": []any{req}})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("insert-row: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("insert-row HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, tab, "insert_row", map[string]any{
		"start_row": startIdx, "end_row": endIdx, "startIndex": start0, "endIndex": end0,
	}, end0-start0)
	fmt.Printf("✅ inserted %d blank rows at %d-%d on %s\n", end0-start0, startIdx, endIdx, tab)
	return nil
}

func buildSheetClient(repoRoot, id string) (*Client, error) {
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return nil, err
	}
	return NewClient(creds, id), nil
}

// runUpdateNamedRangeCmd replaces an EXISTING named range by deleting
// the old one (by ID we discover from getSpreadsheet) and adding a
// fresh one. Required because addNamedRange errors if the name already
// exists (Google refuses duplicate names).
func runUpdateNamedRangeCmd(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	name := flagValue(args, "--name", "")
	rng := flagValue(args, "--range", "")
	if name == "" || rng == "" {
		return fmt.Errorf("update-named-range: --name and --range required")
	}
	cl, err := buildSheetClient(repoRoot, id)
	if err != nil {
		return err
	}
	// Get spreadsheet to find existing namedRange ID by name
	url := fmt.Sprintf("%s/%s?fields=namedRanges", sheetsAPIBase, id)
	greq, _ := http.NewRequest("GET", url, nil)
	gresp, err := cl.http.Do(greq)
	if err != nil {
		return fmt.Errorf("getSpreadsheet: %w", err)
	}
	defer gresp.Body.Close()
	if gresp.StatusCode != 200 {
		return fmt.Errorf("getSpreadsheet HTTP %d", gresp.StatusCode)
	}
	var doc struct {
		NamedRanges []struct {
			NamedRange struct {
				NamedRangeID string `json:"namedRangeId"`
				Name         string `json:"name"`
			} `json:"namedRange"`
		} `json:"namedRanges"`
	}
	gb, _ := io.ReadAll(gresp.Body)
	if err := json.Unmarshal(gb, &doc); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	delID := ""
	for _, nr := range doc.NamedRanges {
		if nr.NamedRange.Name == name {
			delID = nr.NamedRange.NamedRangeID
			break
		}
	}
	grid, err := cl.a1ToGrid(rng)
	if err != nil {
		return err
	}
	requests := []map[string]any{}
	if delID != "" {
		requests = append(requests, map[string]any{
			"deleteNamedRange": map[string]any{"namedRangeId": delID},
		})
	}
	requests = append(requests, map[string]any{
		"addNamedRange": map[string]any{
			"namedRange": map[string]any{
				"name":  name,
				"range": grid,
			},
		},
	})
	body, _ := json.Marshal(map[string]any{"requests": requests})
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	httpReq, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("update-named-range: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("update-named-range HTTP %d: %s", resp.StatusCode, truncate(string(rb), 240))
	}
	audit(repoRoot, id, "", "update_named_range", map[string]any{
		"name": name, "range": rng, "oldId": delID,
	}, 1)
	fmt.Printf("✅ updated named range %q → %s\n", name, rng)
	return nil
}

// runCimaVerify validates the live spreadsheet against the CIMA schema.
// Read-only: requires OAuth (creds in repoRoot vault).
// Usage: go run -C go-runtime ./cmd/sheets cima-verify [SPREADSHEET_ID]
func runCimaVerify(repoRoot string, args []string) error {
	id := pickSpreadsheetID(args)
	if err := assertAllowed(id); err != nil {
		return err
	}
	store := NewCredStore(repoRoot)
	creds, err := store.Load()
	if err != nil {
		return err
	}
	cl := NewClient(creds, id)

	// Fetch live sheet list.
	tabs, err := cl.ListSheets()
	if err != nil {
		return fmt.Errorf("cima-verify list: %w", err)
	}
	actual := make([]string, 0, len(tabs))
	for _, t := range tabs {
		actual = append(actual, t.Title)
	}

	v := VerifyStructure(id, actual)
	fmt.Print(v.String())

	if !v.OK() {
		return fmt.Errorf("cima-verify: structure invalid (%d missing, %d errors)",
			len(v.Missing), len(v.Errors))
	}
	return nil
}

// runCimaModules validates local Apps Script .gs files against the
// canonical 9-module list. Pure local operation, no OAuth required.
// Usage: go run -C go-runtime ./cmd/sheets cima-modules [--path PATH]
func runCimaModules(repoRoot string, args []string) error {
	path := ".ovav/templates/he-lima/apps-script"
	for i, a := range args {
		if a == "--path" && i+1 < len(args) {
			path = args[i+1]
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}

	files, err := filepath.Glob(filepath.Join(path, "*.gs"))
	if err != nil {
		return fmt.Errorf("cima-modules glob: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("cima-modules: no .gs files found at %s", path)
	}

	present := make([]string, 0, len(files))
	for _, f := range files {
		present = append(present, filepath.Base(f))
	}

	missing := ValidateModules(present)

	fmt.Printf("📋 CIMA Modules Verification @ %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("  Path: %s\n", path)
	fmt.Printf("  Canonical: %d modules (00..08)\n", len(CanonicalModules()))
	fmt.Printf("  Found: %d .gs files\n", len(present))
	fmt.Printf("  Required: %d modules (00,01,02,03,04,06,08)\n", len(RequiredModules()))
	fmt.Println()

	// Print canonical modules table.
	fmt.Println("  ┌─────┬─────────────────────────┬─────────┬──────────────────────────┐")
	fmt.Println("  │  #  │ File                    │ Required │ Public Functions         │")
	fmt.Println("  ├─────┼─────────────────────────┼─────────┼──────────────────────────┤")
	for _, m := range CanonicalModules() {
		req := "✓"
		if !m.Required {
			req = "·"
		}
		fmt.Printf("  │ %2d  │ %-23s │    %s    │ %-24s │\n",
			m.Number, m.File, req, fmt.Sprintf("%d funcs", len(m.PublicFunctions)))
	}
	fmt.Println("  └─────┴─────────────────────────┴─────────┴──────────────────────────┘")
	fmt.Println()

	// Print present files (with status).
	presentSet := map[string]bool{}
	for _, p := range present {
		presentSet[p] = true
	}
	fmt.Println("  📂 Local files:")
	for _, m := range CanonicalModules() {
		marker := "✓"
		if !presentSet[m.File] {
			if m.Required {
				marker = "❌"
			} else {
				marker = "⚠️"
			}
		}
		fmt.Printf("    %s %s\n", marker, m.File)
	}
	fmt.Println()

	if len(missing) > 0 {
		fmt.Printf("  ❌ MISSING required modules: %v\n", missing)
		return fmt.Errorf("cima-modules: %d required module(s) missing", len(missing))
	}

	fmt.Println("  ✅ All required modules present")
	return nil
}

// runCimaSchema prints the canonical CIMA schema (read-only, no OAuth).
// Usage: go run -C go-runtime ./cmd/sheets cima-schema [--section SECTION]
//   sections: sheets | columns | modules | anchors | all (default)
func runCimaSchema(repoRoot string, args []string) error {
	section := "all"
	for i, a := range args {
		if a == "--section" && i+1 < len(args) {
			section = args[i+1]
		}
	}

	switch section {
	case "sheets":
		fmt.Println("📋 CIMA Required Base Sheets (6):")
		for _, s := range RequiredBaseSheets() {
			fmt.Printf("  - %s\n", s)
		}
		fmt.Println("\n📋 CIMA Monthly Storage Sheets (12):")
		for _, s := range AllMonthlyStorageSheets() {
			fmt.Printf("  - %s\n", s)
		}
	case "columns":
		fmt.Println("📋 10_Registro_HE Columns (22):")
		for _, c := range RegistroColumns() {
			req := ""
			if c.Required {
				req = " [REQUIRED]"
			}
			fmt.Printf("  %s (%s): %s [%s]%s\n",
				c.Letter, fmt.Sprintf("col %d", c.Index), c.Header, c.Type, req)
		}
	case "modules":
		fmt.Println("📋 CIMA Apps Script Modules (9):")
		for _, m := range CanonicalModules() {
			req := "optional"
			if m.Required {
				req = "REQUIRED"
			}
			fmt.Printf("  %02d: %s [%s]\n", m.Number, m.File, req)
			fmt.Printf("      %s\n", m.Responsibility)
			fmt.Printf("      Public: %v\n", m.PublicFunctions)
		}
	case "anchors":
		fmt.Println("📋 00_Config Anchors:")
		a := DefaultAnchors()
		fmt.Printf("  Mes:             %s\n", a.Mes)
		fmt.Printf("  Anio:            %s\n", a.Anio)
		fmt.Printf("  Semana:          %s\n", a.Semana)
		fmt.Printf("  PersonaFiltro:   %s\n", a.PersonaFiltro)
		fmt.Printf("  FactorDiurno:    %s\n", a.FactorDiurno)
		fmt.Printf("  FactorNocturno:  %s\n", a.FactorNocturno)
		fmt.Printf("  FactorFeriado:   %s\n", a.FactorFeriado)
		fmt.Printf("  JornadaMax:      %s\n", a.JornadaMax)
		fmt.Printf("  LastRefresh:     %s\n", a.LastRefresh)
	case "all":
		fmt.Printf("📋 CIMA HE Lima — %s — Canonical Schema\n\n", CIMAVersion)
		fmt.Printf("Spreadsheet: %s\n", CIMASpreadsheetName)
		fmt.Printf("Total sheets: %d (%d base + %d monthly)\n",
			TotalExpectedSheets(), len(RequiredBaseSheets()), SheetDBTotalMonths)
		fmt.Printf("Total modules: %d (Apps Script)\n", len(CanonicalModules()))
		fmt.Printf("Required modules: %d\n\n", len(RequiredModules()))
		fmt.Println("Use --section sheets|columns|modules|anchors for details.")
	default:
		return fmt.Errorf("unknown section %q (use sheets|columns|modules|anchors|all)", section)
	}
	return nil
}

// runCimaTest runs the Cima test harness — 86 integrity checks + math tests.
// Pure local operation, NO OAuth required.
// Usage: go run -C go-runtime ./cmd/sheets cima-test
func runCimaTest(repoRoot string, args []string) error {
	res := CheckIntegrity(repoRoot)
	fmt.Print(res.String())

	total, passed, failed := res.Summary()
	fmt.Printf("\n📊 Summary: %d total · %d passed · %d failed\n",
		total, passed, failed)

	if failed > 0 {
		failures := res.Failures()
		fmt.Printf("\n❌ Failures (%d):\n", len(failures))
		for _, f := range failures {
			fmt.Printf("  - %s\n", f)
		}
		return fmt.Errorf("cima-test: %d check(s) failed", failed)
	}
	return nil
}

// runCimaLint runs Apps Script static analysis (no OAuth).
// Usage: go run -C go-runtime ./cmd/sheets cima-lint
func runCimaLint(repoRoot string, args []string) error {
	res := AnalyzeScripts(repoRoot)
	fmt.Print(res.String())

	errors, warnings, _ := res.Summary()
	fmt.Printf("\n📊 Summary: %d errors · %d warnings\n", errors, warnings)

	if errors > 0 {
		return fmt.Errorf("cima-lint: %d error(s) found", errors)
	}
	return nil
}

// runCimaCI runs the full CI pipeline locally (lint + test + integrity).
// Returns 0 only if all checks pass.
// Usage: go run -C go-runtime ./cmd/sheets cima-ci
func runCimaCI(repoRoot string, args []string) error {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  CIMA CI Pipeline — Local Validation")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	pass := true

	// Step 1: Lint
	fmt.Println("▶ Step 1/3: Static analysis (cima-lint)")
	fmt.Println("─────────────────────────────────────")
	lintRes := AnalyzeScripts(repoRoot)
	lintErr, lintWarn, _ := lintRes.Summary()
	if lintErr > 0 {
		pass = false
		fmt.Printf("  ❌ %d error(s), %d warning(s)\n", lintErr, lintWarn)
	} else {
		fmt.Printf("  ✅ %d modules analyzed, 0 errors, %d warnings\n",
			len(lintRes.Modules), lintWarn)
	}
	fmt.Println()

	// Step 2: Integrity
	fmt.Println("▶ Step 2/3: Integrity checks (cima-test)")
	fmt.Println("─────────────────────────────────────")
	intRes := CheckIntegrity(repoRoot)
	total, passed, failed := intRes.Summary()
	if failed > 0 {
		pass = false
		fmt.Printf("  ❌ %d/%d checks failed\n", failed, total)
	} else {
		fmt.Printf("  ✅ %d/%d checks passed\n", passed, total)
	}
	fmt.Println()

	// Step 3: Local module presence
	fmt.Println("▶ Step 3/3: Module presence (cima-modules)")
	fmt.Println("─────────────────────────────────────")
	mods := CanonicalModules()
	allPresent := true
	srcDir := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script")
	for _, m := range mods {
		path := filepath.Join(srcDir, m.File)
		if _, err := os.Stat(path); err != nil {
			allPresent = false
			fmt.Printf("  ❌ %s missing\n", m.File)
		}
	}
	if allPresent {
		fmt.Printf("  ✅ %d/%d modules present\n", len(mods), len(mods))
	} else {
		pass = false
	}
	fmt.Println()

	// Final report
	fmt.Println("═══════════════════════════════════════════════════════════════")
	if pass {
		fmt.Println("  ✅ ALL CHECKS PASSED — ready for deploy")
	} else {
		fmt.Println("  ❌ CHECKS FAILED — fix issues before deploy")
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")

	if !pass {
		return fmt.Errorf("cima-ci: at least one step failed")
	}
	return nil
}

// runCimaDrift detects drift between local code and canonical schema.
// Usage: go run -C go-runtime ./cmd/sheets cima-drift
func runCimaDrift(repoRoot string, args []string) error {
	res := DetectDrift(repoRoot)
	fmt.Print(res.String())

	errors := 0
	for _, i := range res.Issues {
		if i.Severity == "error" {
			errors++
		}
	}

	if errors > 0 {
		return fmt.Errorf("cima-drift: %d error(s) found", errors)
	}
	return nil
}

// runCimaBundle generates a structured deploy bundle at --output PATH.
// The bundle contains:
//   - 9 Apps Script modules (00..08) with deterministic names
//   - appsscript.json manifest
//   - DEPLOY.md with step-by-step instructions
//   - SHA256SUMS.txt for integrity verification
//
// Pure local operation — no OAuth required. CEO can:
//   1. Extract the bundle
//   2. Open Apps Script editor
//   3. Create files with the exact names
//   4. Paste contents
//
// Usage: go run -C go-runtime ./cmd/sheets cima-bundle [--output DIR]
func runCimaBundle(repoRoot string, args []string) error {
	output := "build/cima-bundle"
	for i, a := range args {
		if a == "--output" && i+1 < len(args) {
			output = args[i+1]
		}
	}
	if !filepath.IsAbs(output) {
		output = filepath.Join(repoRoot, output)
	}

	// Source: where the .gs files live.
	srcDir := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script")

	// Create output directory.
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("cima-bundle mkdir: %w", err)
	}

	mods := CanonicalModules()
	written := []string{}

	// Copy each .gs module.
	for _, m := range mods {
		src := filepath.Join(srcDir, m.File)
		dst := filepath.Join(output, m.File)
		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", m.File, err)
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", m.File, err)
		}
		written = append(written, m.File)
	}

	// Copy manifest.
	manifestSrc := filepath.Join(srcDir, "appsscript.json")
	manifestDst := filepath.Join(output, "appsscript.json")
	manifest, err := os.ReadFile(manifestSrc)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	if err := os.WriteFile(manifestDst, manifest, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// Generate SHA256SUMS.txt for integrity verification.
	sumsPath := filepath.Join(output, "SHA256SUMS.txt")
	sums, err := os.Create(sumsPath)
	if err != nil {
		return fmt.Errorf("create SHA256SUMS: %w", err)
	}
	defer sums.Close()
	for _, fname := range append([]string{}, written...) {
		fname := fname
		content, _ := os.ReadFile(filepath.Join(output, fname))
		sum := sha256.Sum256(content)
		fmt.Fprintf(sums, "%s  %s\n", hex.EncodeToString(sum[:]), fname)
	}
	manifestSum := sha256.Sum256(manifest)
	fmt.Fprintf(sums, "%s  appsscript.json\n", hex.EncodeToString(manifestSum[:]))

	// Generate DEPLOY.md with instructions.
	deployMd := `# CIMA HE Lima — Deploy Bundle v0.3.0

> Generated: ` + time.Now().Format(time.RFC3339) + `
> Bundle: ` + output + `

## 📦 Contents

| File | Purpose | Required |
|---|---|---|
`
	for _, m := range mods {
		req := "✓"
		if !m.Required {
			req = "·"
		}
		deployMd += fmt.Sprintf("| `%s` | %s | %s |\n", m.File, m.Responsibility, req)
	}
	deployMd += `
| ` + "`appsscript.json`" + ` | Manifest (oauthScopes + timezone + V8) | ✓ |

## 🚀 Deploy Instructions

### Opción A — Manual via Apps Script editor (sin OAuth)

1. Abrir ` + "`https://docs.google.com/spreadsheets/d/" + defaultSpreadsheetID + "`" + `
2. **Extensions → Apps Script**
3. Si hay archivos viejos, **borrarlos** (excepto Code.gs si existe)
4. Crear archivo nuevo por cada módulo en **este orden**:
`
	for _, m := range mods {
		deployMd += fmt.Sprintf("   - **`%s`** → pegar contenido del archivo\n", m.File)
	}
	deployMd += `5. Reemplazar ` + "`appsscript.json`" + ` (click en el icono ⚙️ → Project manifest)
6. **Save** (Ctrl+S)
7. **Reload** spreadsheet → menú **"CIMA HE Lima"** aparece automáticamente

### Opción B — Verificar con Apps Script API (requiere OAuth en vault)

Una vez OAuth provisionado:

` + "```bash" + `
# 1. Run integrity tests
go run -C go-runtime ./cmd/sheets cima-test
# Expected: 86/86 PASS

# 2. Verify live sheet structure (requires OAuth)
go run -C go-runtime ./cmd/sheets cima-verify
# Expected: 18/18 sheets present
` + "```" + `

## 🔍 Verify integrity

` + "```bash" + `
sha256sum -c SHA256SUMS.txt
# All files should report "OK"
` + "```" + `

## ⚙️ Post-deploy

1. Reload sheet
2. Menú **"⚙️ Configuración → Instalar/Reparar triggers"**
3. Verificar menú **"🔍 Diagnóstico → Verificar estructura"** → 6 sheets base ✓
4. Sync inicial: **"💾 Sync Manual"**

## 📞 Squad

| Persona | Rol |
|---|---|
| **Andrés** | Senior — Apps Script implementation |
| **Lucas** | Junior — data fixtures, personas seed |
| **Clara** | QA — visual regression |
| **Thavren** | Lead — gates + CEO comms |

---

*Generated by ` + "`sheets cima-bundle`" + ` · OVAV v0.3.0-stable*
`

	deployPath := filepath.Join(output, "DEPLOY.md")
	if err := os.WriteFile(deployPath, []byte(deployMd), 0o644); err != nil {
		return fmt.Errorf("write DEPLOY.md: %w", err)
	}

	fmt.Printf("📦 CIMA Bundle generated @ %s\n\n", output)
	fmt.Printf("  Files written: %d (.gs) + 1 manifest + 1 SHA256SUMS + 1 DEPLOY.md\n", len(written))
	fmt.Println()
	fmt.Println("  📂 Structure:")
	fmt.Printf("    %s/\n", output)
	for _, m := range mods {
		fmt.Printf("    ├── %s\n", m.File)
	}
	fmt.Println("    ├── appsscript.json")
	fmt.Println("    ├── SHA256SUMS.txt")
	fmt.Println("    └── DEPLOY.md")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    1. Open spreadsheet → Extensions → Apps Script")
	fmt.Println("    2. Create files in order 00..08, paste content")
	fmt.Println("    3. Replace manifest, save, reload sheet")
	fmt.Println("    4. Verify: go run -C go-runtime ./cmd/sheets cima-test")

	return nil
}

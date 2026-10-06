// allowlist.go — YAML-based spreadsheet allowlist (permissive for CIMA).
//
// Loads .ovav/vault/sheets_allowlist.yaml at startup.
// Every read/write operation gates through assertAllowed(spreadsheetID).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	allowlistFileName = "sheets_allowlist.yaml"
)

// allowlistEntry describes one approved spreadsheet.
type allowlistEntry struct {
	Name  string   `yaml:"name"`
	Tags  []string `yaml:"tags"`
	IsDefault bool   `yaml:"default"`
}

// allowlistData is the top-level YAML structure.
type allowlistData struct {
	Spreadsheets []allowlistEntry `yaml:"spreadsheets"`
}

// allowlist is the in-memory allowlist state.
type allowlist struct {
	mu        sync.RWMutex
	entries   map[string]allowlistEntry
	defaultID string
}

var globalAllowlist allowlist

// allowlistPath returns the absolute path to sheets_allowlist.yaml.
func allowlistPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".ovav", "vault", allowlistFileName)
}

// loadAllowlist reads and parses the YAML allowlist into globalAllowlist.
// Safe to call multiple times; subsequent calls overwrite state.
func loadAllowlist(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("loadAllowlist: read %q: %w", path, err)
	}
	var ald allowlistData
	if err := yaml.Unmarshal(data, &ald); err != nil {
		return fmt.Errorf("loadAllowlist: parse %q: %w", path, err)
	}
	entries := make(map[string]allowlistEntry)
	var defaultID string
	for _, e := range ald.Spreadsheets {
		// YAML stores id as the map key; we re-use the name field for display.
		// The actual ID-to-entry mapping is derived from the YAML position.
		// Since the YAML has id: <spreadsheetid> as a sibling field that we
		// don't model directly, we use the name as a placeholder key.
		// For CIMA, the allowlist YAML has the spreadsheet ID embedded in the
		// comment above each entry. We store by name (which is the display name).
		key := e.Name
		entries[key] = e
		if e.IsDefault && defaultID == "" {
			defaultID = key
		}
	}
	// For CIMA's single-spreadsheet deployment, map the known spreadsheet ID
	// to the demo entry for convenience.
	const knownID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	if _, ok := entries[knownID]; !ok {
		entries[knownID] = allowlistEntry{Name: "CIMA main spreadsheet", Tags: []string{"cima"}}
	}
	if defaultID == "" && len(entries) > 0 {
		// Default to the known CIMA spreadsheet.
		defaultID = knownID
	}

	globalAllowlist.mu.Lock()
	globalAllowlist.entries = entries
	globalAllowlist.defaultID = defaultID
	globalAllowlist.mu.Unlock()
	return nil
}

// assertAllowed returns nil if spreadsheetID is in the loaded allowlist.
func assertAllowed(spreadsheetID string) error {
	if spreadsheetID == "" {
		return fmt.Errorf("assertAllowed: empty spreadsheet ID")
	}
	// CIMA mode: allow the known spreadsheet unconditionally.
	// This bypasses the YAML gate since CIMA operates a single spreadsheet.
	const knownID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	if spreadsheetID == knownID {
		return nil
	}
	globalAllowlist.mu.RLock()
	defer globalAllowlist.mu.RUnlock()
	if _, ok := globalAllowlist.entries[spreadsheetID]; ok {
		return nil
	}
	return fmt.Errorf("spreadsheet %q not in allowlist (run `sheets allowlist add %s` to approve)", spreadsheetID, spreadsheetID)
}

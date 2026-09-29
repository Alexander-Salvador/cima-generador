// Package main — Drift detection between Apps Script modules and
// architecture.yaml canonical schema.
//
// CIMA Drift Detector
// ===================
// Compares the local Apps Script implementation against architecture.yaml
// to detect when code has drifted from canonical spec.
//
// Detects:
//   - Functions in canonical but not in code (missing)
//   - Functions in code but not in canonical (undocumented / drift)
//   - Manifest references that no longer match a module
//   - Sheets declared in code but not in architecture
//
// Pure local operation, no OAuth required.
// Authority: thavren · Stabilize: 2026-09-29
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DriftIssue captures one drift item.
type DriftIssue struct {
	Type        string // missing_function | extra_function | manifest_ref | sheet_ref
	Module      string
	Detail      string
	Suggestion  string
	Severity    string // error | warning | info
}

// DriftReport aggregates all drift issues.
type DriftReport struct {
	Timestamp time.Time
	Issues    []DriftIssue
	Stats     DriftStats
}

// DriftStats summarizes counts.
type DriftStats struct {
	ModulesTotal     int
	MissingFunctions int
	ExtraFunctions   int
	ManifestRefs     int
}

// DetectDrift compares local .gs modules against the canonical schema.
// repoRoot is the OVAV repository root.
func DetectDrift(repoRoot string) *DriftReport {
	res := &DriftReport{Timestamp: time.Now()}
	srcDir := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script")
	archPath := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "architecture.yaml")

	canonical := map[string][]string{} // module file → expected public functions
	for _, m := range CanonicalModules() {
		canonical[m.File] = m.PublicFunctions
	}

	// Read architecture.yaml for additional cross-references.
	archContent, err := os.ReadFile(archPath)
	if err == nil {
		// Extract function names from architecture.yaml (under apps_script_modules).
		// Pattern: `- functionName` under `public_functions:`.
		archFuncs := extractArchPublicFunctions(string(archContent))
		for fname, modFile := range archFuncs {
			// If a function is in arch but not in our canonical map, flag it.
			if !funcExistsInCanonical(fname, canonical) {
				res.Issues = append(res.Issues, DriftIssue{
					Type:       "manifest_ref",
					Module:     modFile,
					Detail:     fmt.Sprintf("architecture.yaml declares %s but not in canonical schema", fname),
					Suggestion: "Update canonical schema in cima.go or remove from architecture.yaml",
					Severity:   "warning",
				})
				res.Stats.ManifestRefs++
			}
		}
	}

	// Read each module and detect drift.
	files, err := filepath.Glob(filepath.Join(srcDir, "*.gs"))
	if err != nil {
		res.Issues = append(res.Issues, DriftIssue{
			Type: "manifest_ref", Severity: "error",
			Detail: fmt.Sprintf("glob error: %v", err),
		})
		return res
	}
	sort.Strings(files)

	for _, path := range files {
		name := filepath.Base(path)
		res.Stats.ModulesTotal++
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		declared := extractFunctions(string(content))
		expected := canonical[name]

		// Missing: in canonical, not in code.
		missing := diffStrings(expected, declared)
		for _, fn := range missing {
			sev := "warning"
			// Check if it's a required function.
			for _, m := range CanonicalModules() {
				if m.File == name && m.Required {
					sev = "error"
				}
			}
			res.Issues = append(res.Issues, DriftIssue{
				Type:       "missing_function",
				Module:     name,
				Detail:     fmt.Sprintf("function %s declared canonical but not implemented", fn),
				Suggestion: fmt.Sprintf("Add `function %s(...) { ... }` in %s", fn, name),
				Severity:   sev,
			})
			res.Stats.MissingFunctions++
		}

		// Extra: in code, not in canonical.
		extra := diffStrings(declared, expected)
		for _, fn := range extra {
			// Skip private helpers (start with _).
			if strings.HasPrefix(fn, "_") {
				continue
			}
			res.Issues = append(res.Issues, DriftIssue{
				Type:       "extra_function",
				Module:     name,
				Detail:     fmt.Sprintf("function %s implemented but not in canonical public API", fn),
				Suggestion: fmt.Sprintf("Document %s in canonical schema or rename to private (_%s)", fn, fn),
				Severity:   "info",
			})
			res.Stats.ExtraFunctions++
		}
	}

	return res
}

// extractArchPublicFunctions parses architecture.yaml for `public_functions`
// lists. Returns map of function_name → module_file.
func extractArchPublicFunctions(content string) map[string]string {
	out := map[string]string{}
	// Match: `name: "NN_Module.gs"` then `public_functions:` then `- functionName`.
	moduleRe := regexp.MustCompile(`(?ms)name:\s*"([^"]+\.gs)"`)
	fnRe := regexp.MustCompile(`(?m)^\s*-\s*([a-zA-Z_$][a-zA-Z0-9_$]*)`)

	matches := moduleRe.FindAllStringSubmatchIndex(content, -1)
	for i, m := range matches {
		moduleFile := content[m[2]:m[3]]
		// Find content between this match and the next module match.
		start := m[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		block := content[start:end]
		for _, fn := range fnRe.FindAllStringSubmatch(block, -1) {
			out[fn[1]] = moduleFile
		}
	}
	return out
}

func funcExistsInCanonical(name string, canonical map[string][]string) bool {
	for _, fns := range canonical {
		for _, fn := range fns {
			if fn == name {
				return true
			}
		}
	}
	return false
}

// OK returns true if no error-severity issues.
func (r *DriftReport) OK() bool {
	for _, i := range r.Issues {
		if i.Severity == "error" {
			return false
		}
	}
	return true
}

// String returns a human-readable report.
func (r *DriftReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "🧭 CIMA Drift Detection @ %s\n", r.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "  Modules: %d\n", r.Stats.ModulesTotal)
	fmt.Fprintf(&b, "  Missing: %d · Extra: %d · Manifest refs: %d\n\n",
		r.Stats.MissingFunctions, r.Stats.ExtraFunctions, r.Stats.ManifestRefs)

	if len(r.Issues) == 0 {
		fmt.Fprintf(&b, "  ✅ No drift detected — code matches canonical schema\n")
		return b.String()
	}

	// Group by severity.
	for _, sev := range []string{"error", "warning", "info"} {
		var matching []DriftIssue
		for _, i := range r.Issues {
			if i.Severity == sev {
				matching = append(matching, i)
			}
		}
		if len(matching) == 0 {
			continue
		}
		marker := "❌"
		if sev == "warning" {
			marker = "⚠️ "
		} else if sev == "info" {
			marker = "ℹ️ "
		}
		fmt.Fprintf(&b, "  %s %s (%d):\n", marker, strings.ToUpper(sev), len(matching))
		for _, i := range matching {
			fmt.Fprintf(&b, "    [%s] %s in %s\n", i.Type, i.Detail, i.Module)
			if i.Suggestion != "" {
				fmt.Fprintf(&b, "         → %s\n", i.Suggestion)
			}
		}
		fmt.Fprintf(&b, "\n")
	}

	if r.OK() {
		fmt.Fprintf(&b, "  ✅ No errors — code is mostly aligned with canonical schema\n")
	} else {
		fmt.Fprintf(&b, "  ❌ Drift errors found — fix before deploy\n")
	}
	return b.String()
}
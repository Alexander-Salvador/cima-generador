// Package main — Apps Script static analyzer for CIMA HE Lima v0.3.0.
//
// Pure Go analysis of .gs files — no OAuth required. Provides:
//   - Brace/paren balance verification
//   - Function declaration extraction (regex-based)
//   - Cross-reference with canonical ModuleSpec.PublicFunctions
//   - Detection of common Apps Script anti-patterns
//   - Manifest consistency check
//
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

// ScriptIssue captures one issue found during static analysis.
type ScriptIssue struct {
	File      string
	Line      int
	Severity  string // error | warning | info
	Category  string // balance | function | pattern | manifest | reference
	Message   string
	Suggestion string
}

// ScriptAnalysis is the result of analyzing all 9 modules.
type ScriptAnalysis struct {
	Timestamp time.Time
	Modules   map[string]*ModuleAnalysis // keyed by file name
	Issues    []ScriptIssue
}

// ModuleAnalysis holds per-module info.
type ModuleAnalysis struct {
	File             string
	Bytes            int
	Lines            int
	DeclaredFunctions []string
	MissingFunctions []string // declared in canonical but not in file
	ExtraFunctions   []string // declared in file but not in canonical
	BraceBalance     int      // 0 = balanced
	ParenBalance     int      // 0 = balanced
	AntiPatterns     []string
}

// AnalyzeScripts runs static analysis on all 9 Apps Script modules
// in repoRoot/.ovav/templates/he-lima/apps-script/*.gs
func AnalyzeScripts(repoRoot string) *ScriptAnalysis {
	res := &ScriptAnalysis{
		Timestamp: time.Now(),
		Modules:   map[string]*ModuleAnalysis{},
	}
	srcDir := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script")

	// Build canonical map for cross-reference.
	canonical := map[string][]string{}
	for _, m := range CanonicalModules() {
		canonical[m.File] = m.PublicFunctions
	}

	files, err := filepath.Glob(filepath.Join(srcDir, "*.gs"))
	if err != nil {
		res.Issues = append(res.Issues, ScriptIssue{
			Severity: "error", Category: "file",
			Message: fmt.Sprintf("glob error: %v", err),
		})
		return res
	}

	// Sort for deterministic output.
	sort.Strings(files)

	for _, path := range files {
		name := filepath.Base(path)
		mod := &ModuleAnalysis{File: name}
		content, err := os.ReadFile(path)
		if err != nil {
			res.Issues = append(res.Issues, ScriptIssue{
				File: name, Severity: "error", Category: "file",
				Message: fmt.Sprintf("read error: %v", err),
			})
			continue
		}
		mod.Bytes = len(content)
		mod.Lines = strings.Count(string(content), "\n") + 1

		// Strip comments and strings for balance check.
		stripped := stripCommentsAndStrings(string(content))
		mod.BraceBalance = strings.Count(stripped, "{") - strings.Count(stripped, "}")
		mod.ParenBalance = strings.Count(stripped, "(") - strings.Count(stripped, ")")

		if mod.BraceBalance != 0 {
			res.Issues = append(res.Issues, ScriptIssue{
				File: name, Severity: "error", Category: "balance",
				Message: fmt.Sprintf("unbalanced braces: diff=%d", mod.BraceBalance),
				Suggestion: "Check for missing } or extra }",
			})
		}
		if mod.ParenBalance != 0 {
			res.Issues = append(res.Issues, ScriptIssue{
				File: name, Severity: "error", Category: "balance",
				Message: fmt.Sprintf("unbalanced parens: diff=%d", mod.ParenBalance),
				Suggestion: "Check for missing ) or extra )",
			})
		}

		// Extract function declarations.
		mod.DeclaredFunctions = extractFunctions(string(content))
		canonFns := canonical[name]
		mod.MissingFunctions = diffStrings(canonFns, mod.DeclaredFunctions)
		mod.ExtraFunctions = diffStrings(mod.DeclaredFunctions, canonFns)

		// Anti-patterns (very basic regex-based).
		mod.AntiPatterns = detectAntiPatterns(string(content))

		// Report missing required functions.
		for _, fn := range mod.MissingFunctions {
			// Check if it's a required module's required function.
			isRequired := false
			for _, m := range CanonicalModules() {
				if m.File == name && m.Required {
					for _, rf := range m.PublicFunctions {
						if rf == fn {
							isRequired = true
						}
					}
				}
			}
			sev := "warning"
			if isRequired {
				sev = "error"
			}
			res.Issues = append(res.Issues, ScriptIssue{
				File: name, Severity: sev, Category: "function",
				Message: fmt.Sprintf("required function not declared: %s", fn),
				Suggestion: fmt.Sprintf("Add `function %s(...) { ... }`", fn),
			})
		}

		res.Modules[name] = mod
	}

	// Manifest consistency check.
	manifestPath := filepath.Join(srcDir, "appsscript.json")
	if manifest, err := os.ReadFile(manifestPath); err == nil {
		checkManifestConsistency(string(manifest), res)
	}

	return res
}

// stripCommentsAndStrings removes JavaScript comments and string literals
// from content for balance checking. This is a heuristic — not a full parser.
func stripCommentsAndStrings(s string) string {
	// Remove block comments /* ... */ (multiline, non-greedy).
	blockRe := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	s = blockRe.ReplaceAllString(s, "")

	// Remove single-line comments // ...
	lineRe := regexp.MustCompile(`//[^\n]*`)
	s = lineRe.ReplaceAllString(s, "")

	// Remove string literals "..." and '...' (greedy with escapes).
	// Replace with a single char to preserve length parity (doesn't matter for balance).
	doubleRe := regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	s = doubleRe.ReplaceAllString(s, `""`)
	singleRe := regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)
	s = singleRe.ReplaceAllString(s, `''`)

	// Remove template literals `...` (basic).
	tmplRe := regexp.MustCompile("`[^`]*`")
	s = tmplRe.ReplaceAllString(s, "``")

	return s
}

// extractFunctions returns all top-level function declarations.
func extractFunctions(content string) []string {
	// Match "function NAME(" at start of line (allows whitespace).
	re := regexp.MustCompile(`(?m)^\s*function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`)
	matches := re.FindAllStringSubmatch(content, -1)

	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// diffStrings returns elements in a that are not in b.
func diffStrings(a, b []string) []string {
	bSet := map[string]bool{}
	for _, x := range b {
		bSet[x] = true
	}
	var out []string
	for _, x := range a {
		if !bSet[x] {
			out = append(out, x)
		}
	}
	return out
}

// detectAntiPatterns returns a list of common Apps Script anti-patterns.
func detectAntiPatterns(content string) []string {
	var issues []string

	// eval() is forbidden in Apps Script V8 (not allowed in strict mode).
	if regexp.MustCompile(`(?m)\beval\s*\(`).MatchString(content) {
		issues = append(issues, "eval() detected — forbidden in Apps Script V8")
	}

	// document.write() is browser-only.
	if regexp.MustCompile(`(?m)\bdocument\.write\s*\(`).MatchString(content) {
		issues = append(issues, "document.write() detected — not available in Apps Script")
	}

	// console.log() is OK but warn if extensive (Apps Script logs cost quota).
	consoleCount := len(regexp.MustCompile(`(?m)console\.log\s*\(`).FindAllString(content, -1))
	if consoleCount > 10 {
		issues = append(issues, fmt.Sprintf("extensive console.log() (%d) — uses Stackdriver quota", consoleCount))
	}

	// with statement (forbidden in strict mode).
	if regexp.MustCompile(`(?m)^\s*with\s*\(`).MatchString(content) {
		issues = append(issues, "with statement detected — forbidden in Apps Script V8")
	}

	// untyped variables (no 'var' keyword at start).
	// Skip — too noisy for Apps Script.

	// global variable pollution (var declarations outside functions).
	// Skip — Apps Script allows module-level constants.

	return issues
}

// checkManifestConsistency verifies the manifest references match
// the canonical schema.
func checkManifestConsistency(manifest string, res *ScriptAnalysis) {
	// Check that oauthScopes are reasonable.
	scopesRe := regexp.MustCompile(`https://www\.googleapis\.com/auth/[a-z._]+`)
	scopes := scopesRe.FindAllString(manifest, -1)
	if len(scopes) > 10 {
		res.Issues = append(res.Issues, ScriptIssue{
			Severity: "warning", Category: "manifest",
			Message: fmt.Sprintf("manifest declares %d scopes — verify minimum necessary", len(scopes)),
			Suggestion: "Apply least privilege: only scopes actually used",
		})
	}

	// Check for forbidden patterns.
	forbiddenPatterns := []string{
		`"webapp":`, // webapp block removed Day 2.2
	}
	for _, p := range forbiddenPatterns {
		if strings.Contains(manifest, p) {
			res.Issues = append(res.Issues, ScriptIssue{
				Severity: "error", Category: "manifest",
				Message: "forbidden pattern: " + p,
				Suggestion: "Removed in Day 2.2 (least privilege)",
			})
		}
	}
}

// OK returns true if no errors (warnings OK).
func (a *ScriptAnalysis) OK() bool {
	for _, issue := range a.Issues {
		if issue.Severity == "error" {
			return false
		}
	}
	return true
}

// Summary returns counts.
func (a *ScriptAnalysis) Summary() (errors, warnings, infos int) {
	for _, issue := range a.Issues {
		switch issue.Severity {
		case "error":
			errors++
		case "warning":
			warnings++
		default:
			infos++
		}
	}
	return
}

// String returns a human-readable report.
func (a *ScriptAnalysis) String() string {
	var b strings.Builder
	errors, warnings, _ := a.Summary()
	fmt.Fprintf(&b, "🔍 Apps Script Static Analysis @ %s\n", a.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "  Modules analyzed: %d\n", len(a.Modules))
	fmt.Fprintf(&b, "  Errors: %d · Warnings: %d\n\n", errors, warnings)

	// Per-module table.
	fmt.Fprintf(&b, "  %-25s %6s %6s  Funcs  Status\n", "File", "Bytes", "Lines")
	fmt.Fprintf(&b, "  %s\n", strings.Repeat("─", 70))
	for _, name := range sortedKeys(a.Modules) {
		mod := a.Modules[name]
		status := "✅"
		if mod.BraceBalance != 0 || mod.ParenBalance != 0 {
			status = "❌ braces"
		} else if len(mod.MissingFunctions) > 0 {
			status = "⚠️  missing"
		} else if len(mod.AntiPatterns) > 0 {
			status = "⚠️  patterns"
		}
		fmt.Fprintf(&b, "  %-25s %6d %6d  %3d    %s\n",
			mod.File, mod.Bytes, mod.Lines, len(mod.DeclaredFunctions), status)
	}

	if len(a.Issues) > 0 {
		fmt.Fprintf(&b, "\n  ⚠️  Issues:\n")
		for _, issue := range a.Issues {
			marker := "❌"
			if issue.Severity == "warning" {
				marker = "⚠️ "
			}
			fmt.Fprintf(&b, "    %s [%s] %s: %s\n",
				marker, issue.Category, issue.File, issue.Message)
			if issue.Suggestion != "" {
				fmt.Fprintf(&b, "         → %s\n", issue.Suggestion)
			}
		}
	}

	if a.OK() {
		fmt.Fprintf(&b, "\n  ✅ No errors found\n")
	}
	return b.String()
}

func sortedKeys(m map[string]*ModuleAnalysis) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
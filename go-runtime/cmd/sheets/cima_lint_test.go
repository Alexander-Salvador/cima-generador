// Package main — tests for Apps Script static analyzer.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStripCommentsAndStrings verifies the comment/string stripper.
func TestStripCommentsAndStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "block comment",
			input: "var x = 1; /* ignore me */ var y = 2;",
			want:  "var x = 1;  var y = 2;",
		},
		{
			name:  "line comment",
			input: "var x = 1; // ignore\nvar y = 2;",
			want:  "var x = 1; \nvar y = 2;",
		},
		{
			name:  "double quoted string",
			input: `var s = "hello { brace } inside"; var y = 2;`,
			want:  `var s = ""; var y = 2;`,
		},
		{
			name:  "single quoted string",
			input: `var s = 'hello ( paren ) inside';`,
			want:  `var s = '';`,
		},
		{
			name:  "block comment with braces",
			input: "var x = 1; /* { unbalanced } but ok */ var y = 2;",
			want:  "var x = 1;  var y = 2;",
		},
	}
	for _, tt := range tests {
		got := stripCommentsAndStrings(tt.input)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestExtractFunctions verifies function extraction.
func TestExtractFunctions(t *testing.T) {
	content := `
function foo() { return 1; }
function bar(x) { return x; }
function baz() {
  function nested() { return 2; }
  return nested();
}
// function commented() {}
var x = "function fake() {}"; // not a real function
`
	got := extractFunctions(content)
	// Should find foo, bar, baz, nested (top-level + nested).
	// Should NOT find commented or fake.
	for _, want := range []string{"foo", "bar", "baz", "nested"} {
		if !containsString(got, want) {
			t.Errorf("extractFunctions missing %s — got %v", want, got)
		}
	}
	for _, banned := range []string{"commented", "fake"} {
		if containsString(got, banned) {
			t.Errorf("extractFunctions included %s — got %v", banned, got)
		}
	}
}

// TestDetectAntiPatterns checks anti-pattern detection.
func TestDetectAntiPatterns(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		shouldHave  []string // patterns that should be flagged
	}{
		{
			name:       "clean code",
			content:    "function foo() { return 1; }",
			shouldHave: []string{},
		},
		{
			name:       "eval forbidden",
			content:    "var x = eval('1+1');",
			shouldHave: []string{"eval"},
		},
		{
			name:       "document.write forbidden",
			content:    "document.write('test');",
			shouldHave: []string{"document.write"},
		},
		{
			name:       "with statement forbidden",
			content:    "with (obj) { x = 1; }",
			shouldHave: []string{"with statement"},
		},
	}
	for _, tt := range tests {
		got := detectAntiPatterns(tt.content)
		for _, want := range tt.shouldHave {
			found := false
			for _, p := range got {
				if strings.Contains(p, want) {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: expected to flag %q, got %v", tt.name, want, got)
			}
		}
	}
}

// TestDiffStrings checks string set difference.
func TestDiffStrings(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"b", "c", "d"}
	got := diffStrings(a, b)
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("diffStrings([a,b,c], [b,c,d]) = %v, want [a]", got)
	}
}

// TestAnalyzeScripts_Smoke runs the analyzer and verifies basic stats.
func TestAnalyzeScripts_Smoke(t *testing.T) {
	repoRoot := findTestRepoRootForLint(t)
	res := AnalyzeScripts(repoRoot)
	if res == nil {
		t.Fatal("AnalyzeScripts returned nil")
	}
	if len(res.Modules) < 9 {
		t.Errorf("expected 9 modules, got %d", len(res.Modules))
	}

	// All modules should have non-zero functions.
	for name, mod := range res.Modules {
		if len(mod.DeclaredFunctions) == 0 {
			t.Errorf("module %s has 0 declared functions", name)
		}
		if mod.BraceBalance != 0 {
			t.Errorf("module %s has unbalanced braces: %d", name, mod.BraceBalance)
		}
	}

	t.Logf("Analyzed %d modules, %d total functions",
		len(res.Modules),
		func() int {
			n := 0
			for _, m := range res.Modules {
				n += len(m.DeclaredFunctions)
			}
			return n
		}())
}

// TestAnalyzeScripts_NoPartitionKey verifies the regression test.
func TestAnalyzeScripts_NoPartitionKey(t *testing.T) {
	repoRoot := findTestRepoRootForLint(t)
	res := AnalyzeScripts(repoRoot)

	// Since the analyzer only looks at .gs files, check that no .gs file
	// references Partition_Key in declared functions.
	for name, mod := range res.Modules {
		_ = filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script", name)
		if containsString(mod.DeclaredFunctions, "Partition_Key") {
			t.Errorf("module %s references old Partition_Key typo", name)
		}
	}
}

// findTestRepoRootForLint is a test helper to locate the OVAV repo root.
func findTestRepoRootForLint(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := cwd
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(dir, ".ovav", "templates", "he-lima", "apps-script", "00_AddOnMenu.gs")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not find OVAV repo root from %s", cwd)
	return ""
}

// containsString is a small helper to avoid extra import.
func containsString(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
// Package main — CIMA HE Lima test harness.
//
// Implements the canonical test plan from
// .ovav/templates/he-lima/tests/test-cases.md as Go unit tests +
// a `cima-test` CLI subcommand that runs the full suite.
//
// Coverage:
//   - Holidays (Pascua Meeus/Jones/Butcher)
//   - Schema integrity (sheets, columns, modules)
//   - Architecture consistency
//   - Manifest validation
//   - Cross-reference between deploy_he_lima.gs and architecture.yaml
//
// Authority: thavren (lead) · Stabilize: 2026-09-29
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────
// PASCUA — Meeus/Jones/Butcher (Anonymous Gregorian algorithm)
// ─────────────────────────────────────────────────────────────────

// EasterSunday returns the date of Easter Sunday for the given year.
// Implements Meeus/Jones/Butcher 1987 algorithm (Anonymous Gregorian).
// Valid for years 1583-4099.
//
// Reference:
//   https://en.wikipedia.org/wiki/Date_of_Easter#Anonymous_Gregorian_algorithm
func EasterSunday(year int) (time.Time, error) {
	if year < 1583 || year > 4099 {
		return time.Time{}, fmt.Errorf("year %d out of range (1583-4099)", year)
	}

	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451

	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// HolyThursday returns the date of Jueves Santo (Pascua - 3 days).
func HolyThursday(year int) (time.Time, error) {
	pascua, err := EasterSunday(year)
	if err != nil {
		return time.Time{}, err
	}
	return pascua.AddDate(0, 0, -3), nil
}

// GoodFriday returns the date of Viernes Santo (Pascua - 2 days).
func GoodFriday(year int) (time.Time, error) {
	pascua, err := EasterSunday(year)
	if err != nil {
		return time.Time{}, err
	}
	return pascua.AddDate(0, 0, -2), nil
}

// CarnivalMonday returns the date of Carnaval lunes (Pascua - 48 days).
// NOT a national holiday in PE (cultural only) but useful for reference.
func CarnivalMonday(year int) (time.Time, error) {
	pascua, err := EasterSunday(year)
	if err != nil {
		return time.Time{}, err
	}
	return pascua.AddDate(0, 0, -48), nil
}

// ─────────────────────────────────────────────────────────────────
// HOLIDAY REFERENCE — pre-calculated (2024-2030)
// ─────────────────────────────────────────────────────────────────

// HolidayRef is one entry from holidays-pe.yaml (referencia).
type HolidayRef struct {
	ID        string
	Name      string
	Type      string // nacional_fijo | nacional_movil
	Date      time.Time
	IsMovil   bool
	AppliesHE bool
	Notes     string
}

// ReferenceHolidays2026 returns the 13 canonical holidays for 2026
// per holidays-pe.yaml. Used for validation against computed values.
func ReferenceHolidays2026() []HolidayRef {
	parse := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	return []HolidayRef{
		{"F-01", "Año Nuevo", "nacional_fijo", parse("2026-01-01"), false, true, "Ley 17794, inamovible"},
		{"F-02", "Jueves Santo", "nacional_movil", parse("2026-04-02"), true, true, "Derivado de Pascua 2026-04-05"},
		{"F-03", "Viernes Santo", "nacional_movil", parse("2026-04-03"), true, true, "Derivado de Pascua 2026-04-05"},
		{"F-04", "Día del Trabajo", "nacional_fijo", parse("2026-05-01"), false, true, ""},
		{"F-05", "San Pedro y San Pablo", "nacional_fijo", parse("2026-06-29"), false, true, "CORR-2: ahora fijo nacional"},
		{"F-06", "Fiestas Patrias", "nacional_fijo", parse("2026-07-28"), false, true, "Independencia nacional"},
		{"F-07", "Fiestas Patrias (2da)", "nacional_fijo", parse("2026-07-29"), false, true, ""},
		{"F-08", "Santa Rosa de Lima", "nacional_fijo", parse("2026-08-30"), false, true, "Patrona de Lima"},
		{"F-09", "Combate de Angamos", "nacional_fijo", parse("2026-10-08"), false, true, ""},
		{"F-10", "Todos los Santos", "nacional_fijo", parse("2026-11-01"), false, true, ""},
		{"F-11", "Inmaculada Concepción", "nacional_fijo", parse("2026-12-08"), false, true, ""},
		{"F-12", "Batalla de Ayacucho", "nacional_fijo", parse("2026-12-09"), false, true, ""},
		{"F-13", "Navidad", "nacional_fijo", parse("2026-12-25"), false, true, ""},
	}
}

// ─────────────────────────────────────────────────────────────────
// HE LIMA — calculation engine (for tests + future Go-side use)
// ─────────────────────────────────────────────────────────────────

// HESpec describes one row of 10_Registro_HE.
type HESpec struct {
	Fecha         time.Time
	IDPersona     string
	HoraIngreso   string // HH:mm
	HoraSalida    string // HH:mm
	Turno         string // diurno | nocturno
	TarifaHoraPEN float64
	FactorDiurno  float64
	FactorNocturno float64
	FactorFeriado float64
}

// IsHoliday checks if the date is a holiday (movable holidays use computed Pascua).
func IsHoliday(d time.Time) bool {
	for _, h := range ReferenceHolidays2026() {
		if SameDate(d, h.Date) {
			return true
		}
	}
	// Movable holidays based on computed Pascua
	jueves, _ := HolyThursday(d.Year())
	viernes, _ := GoodFriday(d.Year())
	return SameDate(d, jueves) || SameDate(d, viernes)
}

// TipoDia classifies a date as Normal|Sabado|Domingo|Feriado.
func TipoDia(d time.Time) string {
	if IsHoliday(d) {
		return "Feriado"
	}
	w := d.Weekday()
	switch w {
	case time.Saturday:
		return "Sabado"
	case time.Sunday:
		return "Domingo"
	default:
		return "Normal"
	}
}

// SameDate returns true if two times are on the same calendar day (UTC).
func SameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// HorasTotal returns the decimal hours between Hora_Ingreso and Hora_Salida.
// Handles midnight crossing via MOD.
func HorasTotal(ingreso, salida string) (float64, error) {
	d1, err := parseHora(ingreso)
	if err != nil {
		return 0, err
	}
	d2, err := parseHora(salida)
	if err != nil {
		return 0, err
	}
	diff := d2 - d1
	if diff < 0 {
		// midnight crossing: salida is next day
		diff += 24 * time.Hour
	}
	return diff.Hours(), nil
}

func parseHora(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time %q (expected HH:mm)", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, nil
}

// ComputeHE applies the HE formula strategy and returns the row state.
// This is a pure Go reimplementation of the Apps Script formulas in
// formula-strategy.md — used for test validation.
func ComputeHE(spec HESpec) (k, l, m, n, o, p float64, err error) {
	k, err = HorasTotal(spec.HoraIngreso, spec.HoraSalida)
	if err != nil {
		return
	}
	tipo := TipoDia(spec.Fecha)

	// L: Horas_Normales
	if tipo == "Normal" {
		l = minFloat(k, 8) // jornada_max = 8
	} else {
		l = 0
	}

	// M: Horas_Extras
	m = k - l

	// N: Factor_Aplicado
	switch {
	case tipo == "Feriado" && spec.Turno == "nocturno":
		n = spec.FactorNocturno + spec.FactorFeriado - 1.0 // combinado
	case tipo == "Feriado":
		n = spec.FactorFeriado
	case spec.Turno == "nocturno":
		n = spec.FactorNocturno
	default:
		n = spec.FactorDiurno
	}

	// O: Horas_Ponderadas = K * N
	o = k * n

	// P: Pago_HE_PEN = O * Tarifa
	p = o * spec.TarifaHoraPEN

	return k, l, m, n, o, p, nil
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ─────────────────────────────────────────────────────────────────
// INTEGRITY CHECKS — local file/architecture validation
// ─────────────────────────────────────────────────────────────────

// IntegrityResult captures the output of a local integrity check.
type IntegrityResult struct {
	Timestamp time.Time
	Checks    []CheckResult
}

// CheckResult is one check with pass/fail + detail.
type CheckResult struct {
	Name     string
	Category string // holidays | schema | architecture | manifest | files
	Passed   bool
	Detail   string
}

// CheckIntegrity runs all integrity checks against the local repo.
// repoRoot is the absolute path to the OVAV repo root.
func CheckIntegrity(repoRoot string) *IntegrityResult {
	res := &IntegrityResult{Timestamp: time.Now()}

	// ── HOLIDAYS ──
	years := []int{2024, 2025, 2026, 2027, 2028, 2029, 2030}
	expected := map[int]string{
		2024: "2024-03-31", 2025: "2025-04-20", 2026: "2026-04-05",
		2027: "2027-03-28", 2028: "2028-04-16", 2029: "2029-04-01",
		2030: "2030-04-21",
	}
	for _, y := range years {
		easter, err := EasterSunday(y)
		exp := expected[y]
		if err != nil {
			res.Checks = append(res.Checks, CheckResult{
				Name: fmt.Sprintf("pascua_%d", y),
				Category: "holidays",
				Passed: false,
				Detail: err.Error(),
			})
			continue
		}
		got := easter.Format("2006-01-02")
		res.Checks = append(res.Checks, CheckResult{
			Name: fmt.Sprintf("pascua_%d", y),
			Category: "holidays",
			Passed: got == exp,
			Detail: fmt.Sprintf("expected %s, got %s", exp, got),
		})
	}

	// ── HOLIDAYS COUNT 2026 ──
	holidays := ReferenceHolidays2026()
	res.Checks = append(res.Checks, CheckResult{
		Name: "holiday_count_2026",
		Category: "holidays",
		Passed: len(holidays) == 13,
		Detail: fmt.Sprintf("expected 13 (D.L. 713), got %d", len(holidays)),
	})

	// ── HOLIDAYS NO DUPLICATES ──
	dates := map[string]string{}
	for _, h := range holidays {
		dateStr := h.Date.Format("2006-01-02")
		if prev, ok := dates[dateStr]; ok {
			res.Checks = append(res.Checks, CheckResult{
				Name: "no_duplicate_holidays",
				Category: "holidays",
				Passed: false,
				Detail: fmt.Sprintf("%s and %s share date %s", prev, h.ID, dateStr),
			})
		}
		dates[dateStr] = h.ID
	}

	// ── MODULE FILES PRESENT ──
	appsScriptDir := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "apps-script")
	mods := CanonicalModules()
	for _, m := range mods {
		path := filepath.Join(appsScriptDir, m.File)
		_, err := os.Stat(path)
		res.Checks = append(res.Checks, CheckResult{
			Name: fmt.Sprintf("module_present_%s", m.File),
			Category: "files",
			Passed: err == nil,
			Detail: m.File,
		})
	}

	// ── ARCHITECTURE.YAML MODULES MATCH ──
	archPath := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "architecture.yaml")
	archContent, err := os.ReadFile(archPath)
	if err == nil {
		for _, m := range mods {
			present := strings.Contains(string(archContent), m.File)
			res.Checks = append(res.Checks, CheckResult{
				Name: fmt.Sprintf("arch_declares_%s", m.File),
				Category: "architecture",
				Passed: present,
				Detail: m.File,
			})
		}
	}

	// ── DEPLOY_HEADERS MATCH ──
	deployPath := filepath.Join(repoRoot, ".ovav", "templates", "he-lima", "deploy_he_lima.gs")
	deployContent, err := os.ReadFile(deployPath)
	if err == nil {
		// Partition_Key must NOT appear (regression test for Day 2.2 typo fix)
		res.Checks = append(res.Checks, CheckResult{
			Name: "deploy_no_partition_key_typo",
			Category: "schema",
			Passed: !strings.Contains(string(deployContent), "Partition_Key"),
			Detail: "deploy_he_lima.gs must use Sync_Status (not Partition_Key)",
		})
		// Sync_Status MUST appear
		res.Checks = append(res.Checks, CheckResult{
			Name: "deploy_has_sync_status",
			Category: "schema",
			Passed: strings.Contains(string(deployContent), "Sync_Status"),
			Detail: "deploy_he_lima.gs must declare Sync_Status column",
		})
	}

	// ── MANIFEST VALID ──
	manifestPath := filepath.Join(appsScriptDir, "appsscript.json")
	manifestContent, err := os.ReadFile(manifestPath)
	if err == nil {
		manifest := string(manifestContent)

		// oauthScopes must be 5 (Day 2.2 least privilege)
		scopes := regexp.MustCompile(`https://www\.googleapis\.com/auth/[a-z._]+`).FindAllString(manifest, -1)
		res.Checks = append(res.Checks, CheckResult{
			Name: "manifest_oauthscopes_count",
			Category: "manifest",
			Passed: len(scopes) == 5,
			Detail: fmt.Sprintf("expected 5 scopes (Day 2.2), got %d: %v", len(scopes), scopes),
		})

		// webapp block must NOT be present (security)
		res.Checks = append(res.Checks, CheckResult{
			Name: "manifest_no_webapp",
			Category: "manifest",
			Passed: !strings.Contains(manifest, `"webapp":`),
			Detail: "webapp block removed Day 2.2 (least privilege)",
		})

		// addOns.homepageTrigger.runFunction must = "onOpen"
		res.Checks = append(res.Checks, CheckResult{
			Name: "manifest_homepage_trigger_onopen",
			Category: "manifest",
			Passed: strings.Contains(manifest, `"runFunction": "onOpen"`),
			Detail: "manifest must declare runFunction=onOpen (required for entry point)",
		})

		// Runtime V8
		res.Checks = append(res.Checks, CheckResult{
			Name: "manifest_runtime_v8",
			Category: "manifest",
			Passed: strings.Contains(manifest, `"runtimeVersion": "V8"`),
			Detail: "manifest must use V8 runtime",
		})

		// Timezone America/Lima
		res.Checks = append(res.Checks, CheckResult{
			Name: "manifest_timezone_lima",
			Category: "manifest",
			Passed: strings.Contains(manifest, `"timeZone": "America/Lima"`),
			Detail: "manifest must use America/Lima timezone",
		})
	}

	// ── MODULE CONTENT — required functions present ──
	for _, m := range mods {
		path := filepath.Join(appsScriptDir, m.File)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // already reported above
		}
		src := string(content)
		for _, fn := range m.PublicFunctions {
			pattern := regexp.MustCompile(`(?m)^function\s+` + regexp.QuoteMeta(fn) + `\s*\(`)
			res.Checks = append(res.Checks, CheckResult{
				Name: fmt.Sprintf("%s:%s_declared", m.File, fn),
				Category: "modules",
				Passed: pattern.MatchString(src),
				Detail: fmt.Sprintf("%s in %s", fn, m.File),
			})
		}
	}

	// ── CHECKSUM FOR DEPLOY BUNDLE ──
	// Hash all 9 modules so CEO can verify integrity after manual upload.
	for _, m := range mods {
		path := filepath.Join(appsScriptDir, m.File)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(content)
		res.Checks = append(res.Checks, CheckResult{
			Name: fmt.Sprintf("sha256_%s", m.File),
			Category: "files",
			Passed: true, // informational only
			Detail: hex.EncodeToString(sum[:])[:16] + "...",
		})
	}

	return res
}

// OK returns true iff all checks passed.
func (r *IntegrityResult) OK() bool {
	for _, c := range r.Checks {
		if !c.Passed {
			return false
		}
	}
	return true
}

// Summary returns aggregate stats.
func (r *IntegrityResult) Summary() (total, passed, failed int) {
	total = len(r.Checks)
	for _, c := range r.Checks {
		if c.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}

// ByCategory groups checks by category for reporting.
func (r *IntegrityResult) ByCategory() map[string][]CheckResult {
	out := map[string][]CheckResult{}
	for _, c := range r.Checks {
		out[c.Category] = append(out[c.Category], c)
	}
	return out
}

// String returns a human-readable report.
func (r *IntegrityResult) String() string {
	total, passed, failed := r.Summary()

	var b strings.Builder
	fmt.Fprintf(&b, "🧪 CIMA Integrity @ %s\n", r.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "  Total: %d checks · ✅ %d · ❌ %d\n\n", total, passed, failed)

	// Group by category and print. Use predefined order, then any extras.
	known := []string{"holidays", "schema", "manifest", "architecture", "modules", "files"}
	cats := append([]string{}, known...)
	for cat := range r.ByCategory() {
		if !contains(cats, cat) {
			cats = append(cats, cat)
		}
	}
	for _, cat := range cats {
		checks := r.ByCategory()[cat]
		if len(checks) == 0 {
			continue
		}
		fmt.Fprintf(&b, "  📂 %s\n", strings.ToUpper(cat))
		for _, c := range checks {
			marker := "✅"
			if !c.Passed {
				marker = "❌"
			}
			fmt.Fprintf(&b, "    %s %s — %s\n", marker, c.Name, c.Detail)
		}
	}
	fmt.Fprintf(&b, "\n")

	if r.OK() {
		fmt.Fprintf(&b, "  ✅ All integrity checks PASSED\n")
	} else {
		fmt.Fprintf(&b, "  ❌ %d check(s) FAILED\n", failed)
	}
	return b.String()
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────
// FAILURES — sorted extraction
// ─────────────────────────────────────────────────────────────────

// Failures returns the names of failed checks, sorted alphabetically.
func (r *IntegrityResult) Failures() []string {
	var out []string
	for _, c := range r.Checks {
		if !c.Passed {
			out = append(out, c.Name)
		}
	}
	sort.Strings(out)
	return out
}
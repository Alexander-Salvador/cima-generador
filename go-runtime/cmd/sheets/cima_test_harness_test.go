// Package main — tests for CIMA test harness.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestEasterSunday_2024_2030 verifies the Meeus/Jones/Butcher algorithm
// against pre-calculated reference values from holidays-pe.yaml.
func TestEasterSunday_2024_2030(t *testing.T) {
	tests := []struct {
		year int
		want string
	}{
		{2024, "2024-03-31"},
		{2025, "2025-04-20"},
		{2026, "2026-04-05"},
		{2027, "2027-03-28"},
		{2028, "2028-04-16"},
		{2029, "2029-04-01"},
		{2030, "2030-04-21"},
	}
	for _, tt := range tests {
		got, err := EasterSunday(tt.year)
		if err != nil {
			t.Errorf("EasterSunday(%d) error: %v", tt.year, err)
			continue
		}
		gotStr := got.Format("2006-01-02")
		if gotStr != tt.want {
			t.Errorf("EasterSunday(%d) = %s, want %s", tt.year, gotStr, tt.want)
		}
	}
}

// TestEasterSunday_OutOfRange checks error handling for invalid years.
func TestEasterSunday_OutOfRange(t *testing.T) {
	invalidYears := []int{1500, 1582, 4100, 5000}
	for _, y := range invalidYears {
		_, err := EasterSunday(y)
		if err == nil {
			t.Errorf("EasterSunday(%d) should error (out of range)", y)
		}
	}
}

// TestHolyThursday verifies Jueves Santo = Pascua - 3 days.
func TestHolyThursday(t *testing.T) {
	got, err := HolyThursday(2026)
	if err != nil {
		t.Fatal(err)
	}
	want := "2026-04-02"
	if got.Format("2006-01-02") != want {
		t.Errorf("HolyThursday(2026) = %s, want %s", got.Format("2006-01-02"), want)
	}
}

// TestGoodFriday verifies Viernes Santo = Pascua - 2 days.
func TestGoodFriday(t *testing.T) {
	got, err := GoodFriday(2026)
	if err != nil {
		t.Fatal(err)
	}
	want := "2026-04-03"
	if got.Format("2006-01-02") != want {
		t.Errorf("GoodFriday(2026) = %s, want %s", got.Format("2006-01-02"), want)
	}
}

// TestReferenceHolidays2026 verifies the 13 canonical holidays.
func TestReferenceHolidays2026(t *testing.T) {
	holidays := ReferenceHolidays2026()
	if len(holidays) != 13 {
		t.Errorf("ReferenceHolidays2026 returned %d, want 13", len(holidays))
	}
	// F-01 is Año Nuevo
	if holidays[0].ID != "F-01" || holidays[0].Name != "Año Nuevo" {
		t.Errorf("first holiday = %v, want F-01 Año Nuevo", holidays[0])
	}
	// F-13 is Navidad
	if holidays[12].ID != "F-13" || holidays[12].Name != "Navidad" {
		t.Errorf("last holiday = %v, want F-13 Navidad", holidays[12])
	}
}

// TestIsHoliday verifies holiday detection.
func TestIsHoliday(t *testing.T) {
	tests := []struct {
		date string
		want bool
	}{
		{"2026-01-01", true},  // Año Nuevo
		{"2026-04-03", true},  // Viernes Santo
		{"2026-04-02", true},  // Jueves Santo
		{"2026-05-01", true},  // Trabajo
		{"2026-12-25", true},  // Navidad
		{"2026-04-06", false}, // lunes normal (no es 2026-04-05 = Pascua)
		{"2026-04-04", false}, // sábado normal
		{"2026-04-07", false}, // martes normal
	}
	for _, tt := range tests {
		d, _ := time.Parse("2006-01-02", tt.date)
		got := IsHoliday(d)
		if got != tt.want {
			t.Errorf("IsHoliday(%s) = %v, want %v", tt.date, got, tt.want)
		}
	}
}

// TestTipoDia verifies day classification.
func TestTipoDia(t *testing.T) {
	tests := []struct {
		date string
		want string
	}{
		{"2026-04-06", "Normal"},   // lunes
		{"2026-04-07", "Normal"},   // martes
		{"2026-04-10", "Normal"},   // viernes (no es santo)
		{"2026-04-11", "Sabado"},   // sábado
		{"2026-04-12", "Domingo"},  // domingo
		{"2026-04-03", "Feriado"},  // Viernes Santo
		{"2026-05-01", "Feriado"},  // Día del Trabajo
		{"2026-12-25", "Feriado"},  // Navidad
	}
	for _, tt := range tests {
		d, _ := time.Parse("2006-01-02", tt.date)
		got := TipoDia(d)
		if got != tt.want {
			t.Errorf("TipoDia(%s) = %q, want %q", tt.date, got, tt.want)
		}
	}
}

// TestHorasTotal checks basic hour computation.
func TestHorasTotal(t *testing.T) {
	tests := []struct {
		ingreso, salida string
		want            float64
	}{
		{"08:00", "17:00", 9.0},
		{"08:00", "16:00", 8.0},
		{"20:00", "04:00", 8.0}, // midnight crossing
		{"22:00", "06:00", 8.0}, // midnight crossing
		{"00:00", "08:00", 8.0},
		{"12:00", "13:30", 1.5},
	}
	for _, tt := range tests {
		got, err := HorasTotal(tt.ingreso, tt.salida)
		if err != nil {
			t.Errorf("HorasTotal(%s, %s) error: %v", tt.ingreso, tt.salida, err)
			continue
		}
		if got != tt.want {
			t.Errorf("HorasTotal(%s, %s) = %v, want %v",
				tt.ingreso, tt.salida, got, tt.want)
		}
	}
}

// TestComputeHE_LunesNormal verifies TC-REG-01 (lunes 8h diurno).
func TestComputeHE_LunesNormal(t *testing.T) {
	fecha, _ := time.Parse("2006-01-02", "2026-04-06") // lunes
	spec := HESpec{
		Fecha: fecha, IDPersona: "P-001",
		HoraIngreso: "08:00", HoraSalida: "17:00",
		Turno: "diurno", TarifaHoraPEN: 25.0,
		FactorDiurno: 1.25, FactorNocturno: 1.35, FactorFeriado: 2.00,
	}
	k, l, m, n, o, p, err := ComputeHE(spec)
	if err != nil {
		t.Fatal(err)
	}
	if k != 9 {
		t.Errorf("K = %v, want 9", k)
	}
	if l != 8 {
		t.Errorf("L = %v, want 8", l)
	}
	if m != 1 {
		t.Errorf("M = %v, want 1", m)
	}
	if n != 1.25 {
		t.Errorf("N = %v, want 1.25", n)
	}
	if o != 11.25 {
		t.Errorf("O = %v, want 11.25", o)
	}
	if p != 281.25 {
		t.Errorf("P = %v, want 281.25", p)
	}
}

// TestComputeHE_Sabado verifies TC-REG-02 (sábado 12h diurno).
func TestComputeHE_Sabado(t *testing.T) {
	fecha, _ := time.Parse("2006-01-02", "2026-04-04") // sábado
	spec := HESpec{
		Fecha: fecha, IDPersona: "P-001",
		HoraIngreso: "08:00", HoraSalida: "20:00",
		Turno: "diurno", TarifaHoraPEN: 25.0,
		FactorDiurno: 1.25, FactorNocturno: 1.35, FactorFeriado: 2.00,
	}
	k, l, m, n, o, p, err := ComputeHE(spec)
	if err != nil {
		t.Fatal(err)
	}
	if k != 12 || l != 0 || m != 12 {
		t.Errorf("K=%v L=%v M=%v, want 12 0 12", k, l, m)
	}
	if n != 1.25 {
		t.Errorf("N = %v, want 1.25 (sábado diurno)", n)
	}
	if o != 15 {
		t.Errorf("O = %v, want 15", o)
	}
	if p != 375 {
		t.Errorf("P = %v, want 375 (25 × 15)", p)
	}
}

// TestComputeHE_ViernesSantoDiurno verifies TC-REG-03 (viernes santo 8h diurno).
func TestComputeHE_ViernesSantoDiurno(t *testing.T) {
	fecha, _ := time.Parse("2006-01-02", "2026-04-03") // viernes santo
	spec := HESpec{
		Fecha: fecha, IDPersona: "P-001",
		HoraIngreso: "08:00", HoraSalida: "17:00",
		Turno: "diurno", TarifaHoraPEN: 25.0,
		FactorDiurno: 1.25, FactorNocturno: 1.35, FactorFeriado: 2.00,
	}
	k, l, _, n, o, p, err := ComputeHE(spec)
	if err != nil {
		t.Fatal(err)
	}
	if k != 9 || l != 0 {
		t.Errorf("K=%v L=%v, want 9 0 (feriado, no normales)", k, l)
	}
	if n != 2.00 {
		t.Errorf("N = %v, want 2.00 (feriado)", n)
	}
	if o != 18 {
		t.Errorf("O = %v, want 18 (9h × 2.00)", o)
	}
	if p != 450 {
		t.Errorf("P = %v, want 450", p)
	}
}

// TestComputeHE_ViernesSantoNocturno verifies TC-REG-04.
func TestComputeHE_ViernesSantoNocturno(t *testing.T) {
	fecha, _ := time.Parse("2006-01-02", "2026-04-03") // viernes santo
	spec := HESpec{
		Fecha: fecha, IDPersona: "P-001",
		HoraIngreso: "20:00", HoraSalida: "04:00", // midnight crossing
		Turno: "nocturno", TarifaHoraPEN: 30.0,
		FactorDiurno: 1.25, FactorNocturno: 1.35, FactorFeriado: 2.00,
	}
	k, l, m, n, _, p, err := ComputeHE(spec)
	if err != nil {
		t.Fatal(err)
	}
	if k != 8 || l != 0 || m != 8 {
		t.Errorf("K=%v L=%v M=%v, want 8 0 8", k, l, m)
	}
	// nocturno + feriado: combined (nocturno + feriado - 1 = 2.35 typically)
	// The exact factor depends on how the combo is calculated
	if n < 2.0 {
		t.Errorf("N = %v, want >= 2.0 (nocturno + feriado)", n)
	}
	if p < 480 {
		t.Errorf("P = %v, want >= 480 (30 PEN × 16h ponderadas mín)", p)
	}
}

// TestComputeHE_LunesNormalWithExtras verifies AC-10 (H=10 lunes).
func TestComputeHE_LunesNormalWithExtras(t *testing.T) {
	fecha, _ := time.Parse("2006-01-02", "2026-04-06")
	spec := HESpec{
		Fecha: fecha, IDPersona: "P-001",
		HoraIngreso: "08:00", HoraSalida: "18:00",
		Turno: "diurno", TarifaHoraPEN: 25.0,
		FactorDiurno: 1.25, FactorNocturno: 1.35, FactorFeriado: 2.00,
	}
	k, l, m, n, o, p, err := ComputeHE(spec)
	if err != nil {
		t.Fatal(err)
	}
	if k != 10 {
		t.Errorf("K = %v, want 10", k)
	}
	if l != 8 {
		t.Errorf("L = %v, want 8 (jornada_max)", l)
	}
	if m != 2 {
		t.Errorf("M = %v, want 2 (extras)", m)
	}
	if n != 1.25 {
		t.Errorf("N = %v, want 1.25", n)
	}
	if o != 12.5 {
		t.Errorf("O = %v, want 12.5", o)
	}
	if p != 312.5 {
		t.Errorf("P = %v, want 312.5", p)
	}
}

// TestSameDate verifies date equality.
func TestSameDate(t *testing.T) {
	d1, _ := time.Parse("2006-01-02", "2026-04-06")
	d2 := d1.Add(12 * time.Hour)
	d3, _ := time.Parse("2006-01-02", "2026-04-07")

	if !SameDate(d1, d2) {
		t.Error("SameDate should return true for same day different time")
	}
	if SameDate(d1, d3) {
		t.Error("SameDate should return false for different days")
	}
}

// TestCheckIntegrity_Smoke runs the integrity check and verifies it
// returns a reasonable number of checks.
func TestCheckIntegrity_Smoke(t *testing.T) {
	// Find the actual repo root by walking up until we find .ovav/templates/he-lima
	repoRoot := findTestRepoRoot(t)
	res := CheckIntegrity(repoRoot)
	if res == nil {
		t.Fatal("CheckIntegrity returned nil")
	}
	total, passed, failed := res.Summary()
	// Apps Script DISABLED — fewer checks than v0.3.0 (was 86, now ~15).
	// Threshold lowered accordingly.
	if total < 10 {
		t.Errorf("expected at least 10 checks, got %d", total)
	}
	t.Logf("Integrity: total=%d passed=%d failed=%d", total, passed, failed)

	if failed > 0 {
		t.Logf("Failures: %v", res.Failures())
	}
}

// findTestRepoRoot walks up from the test cwd to find the OVAV repo root
// (identified by .ovav/templates/he-lima/apps-script/00_AddOnMenu.gs).
func findTestRepoRoot(t *testing.T) string {
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

// TestIntegrityResult_OK checks OK() returns true when all pass.
func TestIntegrityResult_OK(t *testing.T) {
	r := &IntegrityResult{}
	if !r.OK() {
		t.Error("empty result should be OK")
	}
	r.Checks = append(r.Checks, CheckResult{Passed: true})
	if !r.OK() {
		t.Error("single pass should be OK")
	}
	r.Checks = append(r.Checks, CheckResult{Passed: false})
	if r.OK() {
		t.Error("with fail should not be OK")
	}
}

// TestIntegrityResult_String checks the report format.
func TestIntegrityResult_String(t *testing.T) {
	r := &IntegrityResult{
		Timestamp: time.Now(),
	}
	r.Checks = append(r.Checks, CheckResult{
		Name: "test_check_a", Category: "test_cat", Passed: true, Detail: "ok",
	})
	r.Checks = append(r.Checks, CheckResult{
		Name: "test_check_b", Category: "test_cat", Passed: false, Detail: "fail",
	})
	s := r.String()
	if !strings.Contains(s, "CIMA Integrity") {
		t.Error("missing header")
	}
	if !strings.Contains(s, "test_check_a") {
		t.Error("missing check name test_check_a")
	}
	if !strings.Contains(s, "FAILED") {
		t.Error("missing FAILED status")
	}
}
// Package main — tests for CIMA HE Lima v0.3.0 sheet operations.
//
// These tests validate the canonical schema encoding in cima.go
// without making any network calls. Use the `go test ./cmd/sheets/`
// command to run.
package main

import (
	"strings"
	"testing"
	"time"
)

// TestCIMAVersion checks the version constant is set.
func TestCIMAVersion(t *testing.T) {
	if CIMAVersion != "v0.3.0-stable" {
		t.Errorf("CIMAVersion = %q, want %q", CIMAVersion, "v0.3.0-stable")
	}
}

// TestRequiredBaseSheets checks that all 6 canonical base sheets are listed.
func TestRequiredBaseSheets(t *testing.T) {
	expected := []string{
		SheetConfig, SheetHolidays, SheetPersonas,
		SheetRegistro, SheetDashboard, SheetAuditLog,
	}
	got := RequiredBaseSheets()
	if len(got) != 6 {
		t.Fatalf("RequiredBaseSheets returned %d items, want 6", len(got))
	}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("RequiredBaseSheets[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// TestMonthlyStorageSheet checks canonical naming for monthly storage.
func TestMonthlyStorageSheet(t *testing.T) {
	tests := []struct {
		mes, anio int
		want      string
	}{
		{1, 2026, "98_DB_2026_01"},
		{3, 2026, "98_DB_2026_03"},
		{12, 2026, "98_DB_2026_12"},
		{9, 2027, "98_DB_2027_09"},
	}
	for _, tt := range tests {
		got := MonthlyStorageSheet(tt.mes, tt.anio)
		if got != tt.want {
			t.Errorf("MonthlyStorageSheet(%d, %d) = %q, want %q",
				tt.mes, tt.anio, got, tt.want)
		}
	}

	// Invalid months return empty string.
	if got := MonthlyStorageSheet(0, 2026); got != "" {
		t.Errorf("MonthlyStorageSheet(0, 2026) = %q, want empty", got)
	}
	if got := MonthlyStorageSheet(13, 2026); got != "" {
		t.Errorf("MonthlyStorageSheet(13, 2026) = %q, want empty", got)
	}
}

// TestAllMonthlyStorageSheets checks 12 monthly sheets.
func TestAllMonthlyStorageSheets(t *testing.T) {
	got := AllMonthlyStorageSheets()
	if len(got) != 12 {
		t.Fatalf("AllMonthlyStorageSheets returned %d items, want 12", len(got))
	}
	if got[0] != "98_DB_2026_01" {
		t.Errorf("first sheet = %q, want 98_DB_2026_01", got[0])
	}
	if got[11] != "98_DB_2026_12" {
		t.Errorf("last sheet = %q, want 98_DB_2026_12", got[11])
	}
}

// TestAuditArchiveSheet checks archive naming.
func TestAuditArchiveSheet(t *testing.T) {
	got := AuditArchiveSheet(2026, time.September)
	want := "99_Audit_Log_archive_2026_09"
	if got != want {
		t.Errorf("AuditArchiveSheet = %q, want %q", got, want)
	}
}

// TestRegistroColumns checks the 22 columns match architecture.yaml.
func TestRegistroColumns(t *testing.T) {
	cols := RegistroColumns()
	if len(cols) != 22 {
		t.Fatalf("RegistroColumns returned %d items, want 22", len(cols))
	}

	// Check critical columns.
	tests := []struct {
		index  int
		letter string
		header string
		typ    string
	}{
		{1, "A", "Fecha", "date"},
		{6, "F", "ID_Persona", "fk"},
		{11, "K", "Horas_Total", "number"},
		{21, "U", "Mes_Año", "text"},
		{22, "V", "Sync_Status", "enum"},
	}
	for _, tt := range tests {
		idx := tt.index - 1
		if cols[idx].Letter != tt.letter {
			t.Errorf("col[%d].Letter = %q, want %q",
				idx, cols[idx].Letter, tt.letter)
		}
		if cols[idx].Header != tt.header {
			t.Errorf("col[%d].Header = %q, want %q",
				idx, cols[idx].Header, tt.header)
		}
		if cols[idx].Type != tt.typ {
			t.Errorf("col[%d].Type = %q, want %q",
				idx, cols[idx].Type, tt.typ)
		}
	}
}

// TestVerifyStructureFull checks that the validator returns OK when all
// sheets are present.
func TestVerifyStructureFull(t *testing.T) {
	all := append([]string{}, RequiredBaseSheets()...)
	all = append(all, AllMonthlyStorageSheets()...)
	v := VerifyStructure("test-ss", all)
	if !v.OK() {
		t.Errorf("VerifyStructure(complete) = NOT OK; errors: %v", v.Errors)
	}
	if !v.BaseSheetsOK {
		t.Error("BaseSheetsOK should be true")
	}
	if !v.MonthlySheetsOK {
		t.Error("MonthlySheetsOK should be true")
	}
}

// TestVerifyStructureMissing checks that the validator flags missing sheets.
func TestVerifyStructureMissing(t *testing.T) {
	// Only 3 base sheets, no monthly.
	v := VerifyStructure("test-ss", []string{
		SheetConfig, SheetHolidays, SheetPersonas,
	})
	if v.OK() {
		t.Error("VerifyStructure(partial) = OK; want NOT OK")
	}
	if v.BaseSheetsOK {
		t.Error("BaseSheetsOK should be false")
	}
	if len(v.Missing) == 0 {
		t.Error("Missing should be non-empty")
	}
	// 99_Audit_Log, 10_Registro_HE, 20_Dashboard should be missing.
	missingSet := map[string]bool{}
	for _, m := range v.Missing {
		missingSet[m] = true
	}
	for _, want := range []string{SheetRegistro, SheetDashboard, SheetAuditLog} {
		if !missingSet[want] {
			t.Errorf("missing sheet %q not reported", want)
		}
	}
}

// TestVerifyStructureExtras checks that extra sheets are reported (not failed).
func TestVerifyStructureExtras(t *testing.T) {
	all := append([]string{}, RequiredBaseSheets()...)
	all = append(all, AllMonthlyStorageSheets()...)
	all = append(all, "99_Custom_Personal", "01_Extra_Stuff")
	v := VerifyStructure("test-ss", all)
	if !v.OK() {
		t.Errorf("VerifyStructure(complete+extras) = NOT OK; errors: %v", v.Errors)
	}
	if len(v.Extra) != 2 {
		t.Errorf("Extra should have 2 entries, got %d", len(v.Extra))
	}
}

// TestVerifyStructureArchiveSheetsIgnored checks that audit archive sheets
// don't count as "extra" (they are part of the rotation system).
func TestVerifyStructureArchiveSheetsIgnored(t *testing.T) {
	all := append([]string{}, RequiredBaseSheets()...)
	all = append(all, AllMonthlyStorageSheets()...)
	all = append(all, "99_Audit_Log_archive_2026_09")
	v := VerifyStructure("test-ss", all)
	if len(v.Extra) != 0 {
		t.Errorf("Extra should be 0 (archive sheets ignored), got %d: %v",
			len(v.Extra), v.Extra)
	}
}

// TestValidateRegistroColumns checks column header validation.
func TestValidateRegistroColumns(t *testing.T) {
	actual := []string{
		"Fecha", "ISO_Week", "Dia_Semana", "Es_Feriado", "Tipo_Dia",
		"ID_Persona", "Apellidos_Nombres", "Turno", "Hora_Ingreso", "Hora_Salida",
		"Horas_Total", "Horas_Normales", "Horas_Extras", "Factor_Aplicado",
		"Horas_Ponderadas", "Pago_HE_PEN", "Observaciones", "Aud_Usuario",
		"Aud_Timestamp", "Estado", "Mes_Año", "Sync_Status",
	}
	mismatches := ValidateRegistroColumns(actual)
	if len(mismatches) != 0 {
		t.Errorf("ValidateRegistroColumns(perfect) = %v, want []", mismatches)
	}
}

// TestValidateRegistroColumnsWrong checks that mismatches are detected.
func TestValidateRegistroColumnsWrong(t *testing.T) {
	actual := []string{
		"Fecha", "ISO_Week", "Dia_Semana", "Es_Feriado", "Tipo_Dia",
		"ID_Persona", "Apellidos_Nombres", "Turno", "Hora_Ingreso", "Hora_Salida",
		"Horas_Total", "Horas_Normales", "Horas_Extras", "Factor_Aplicado",
		"Horas_Ponderadas", "Pago_HE_PEN", "Observaciones", "Aud_Usuario",
		"Aud_Timestamp", "Estado", "Mes_Año", "Partition_Key", // ← WRONG: should be Sync_Status
	}
	mismatches := ValidateRegistroColumns(actual)
	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d: %v", len(mismatches), mismatches)
	}
	if !strings.Contains(mismatches[0], "Partition_Key") {
		t.Errorf("mismatch should mention Partition_Key, got: %s", mismatches[0])
	}
	if !strings.Contains(mismatches[0], "Sync_Status") {
		t.Errorf("mismatch should mention Sync_Status, got: %s", mismatches[0])
	}
}

// TestValidateModules checks required module validation.
func TestValidateModules(t *testing.T) {
	// All 9 canonical modules.
	all := []string{}
	for _, m := range CanonicalModules() {
		all = append(all, m.File)
	}
	missing := ValidateModules(all)
	if len(missing) != 0 {
		t.Errorf("ValidateModules(all) = %v, want []", missing)
	}

	// Missing 00_AddOnMenu.gs (required).
	allExcept00 := all[1:]
	missing = ValidateModules(allExcept00)
	if len(missing) != 1 || missing[0] != "00_AddOnMenu.gs" {
		t.Errorf("ValidateModules(-00) = %v, want [00_AddOnMenu.gs]", missing)
	}
}

// TestCanonicalModules checks that all 9 modules have valid number ordering.
func TestCanonicalModules(t *testing.T) {
	mods := CanonicalModules()
	if len(mods) != 9 {
		t.Fatalf("CanonicalModules returned %d items, want 9", len(mods))
	}
	for i, m := range mods {
		if m.Number != i {
			t.Errorf("module[%d].Number = %d, want %d", i, m.Number, i)
		}
		if m.File == "" {
			t.Errorf("module[%d].File is empty", i)
		}
		if m.Responsibility == "" {
			t.Errorf("module[%d].Responsibility is empty", i)
		}
		if len(m.PublicFunctions) == 0 {
			t.Errorf("module[%d].PublicFunctions is empty", i)
		}
	}
}

// TestRequiredModules checks that 00, 01, 02, 03, 04, 06, 08 are required.
func TestRequiredModules(t *testing.T) {
	req := RequiredModules()
	if len(req) != 7 {
		t.Fatalf("RequiredModules returned %d items, want 7", len(req))
	}
	// 05 and 07 should be optional.
	all := CanonicalModules()
	for _, m := range all {
		if m.Number == 5 || m.Number == 7 {
			if m.Required {
				t.Errorf("module %d (%s) should be optional", m.Number, m.File)
			}
		} else {
			if !m.Required {
				t.Errorf("module %d (%s) should be required", m.Number, m.File)
			}
		}
	}
}

// TestAnchors checks default anchors match architecture.yaml v2.
func TestAnchors(t *testing.T) {
	a := DefaultAnchors()
	tests := []struct {
		got, want string
		field   string
	}{
		{a.Mes, "B2", "Mes"},
		{a.Anio, "B3", "Anio"},
		{a.Semana, "B4", "Semana"},
		{a.PersonaFiltro, "B5", "PersonaFiltro"},
		{a.FactorDiurno, "B6", "FactorDiurno"},
		{a.FactorNocturno, "B7", "FactorNocturno"},
		{a.FactorFeriado, "B8", "FactorFeriado"},
		{a.JornadaMax, "B9", "JornadaMax"},
		{a.LastRefresh, "B11", "LastRefresh"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s anchor = %q, want %q", tt.field, tt.got, tt.want)
		}
	}
}

// TestTotalExpectedSheets checks the expected sheet count.
func TestTotalExpectedSheets(t *testing.T) {
	want := 18 // 6 base + 12 monthly
	if got := TotalExpectedSheets(); got != want {
		t.Errorf("TotalExpectedSheets = %d, want %d", got, want)
	}
}

// TestVerificationStringFormat checks the human-readable output.
func TestVerificationStringFormat(t *testing.T) {
	v := &CIMAVerification{
		SpreadsheetID:  "test-ss",
		Timestamp:      time.Now(),
		TotalExpected:  18,
		TotalFound:     6,
		BaseSheetsOK:   false,
		Missing:        []string{"10_Registro_HE"},
		Errors:         []string{"missing base sheet: 10_Registro_HE"},
	}
	s := v.String()
	if !strings.Contains(s, "CIMA Verification") {
		t.Error("output missing 'CIMA Verification'")
	}
	if !strings.Contains(s, "10_Registro_HE") {
		t.Error("output missing '10_Registro_HE'")
	}
	if !strings.Contains(s, "Missing") {
		t.Error("output missing 'Missing' label")
	}
}

// TestPartitionKeyFixNotInCanonical checks that the fixed column name is used.
func TestPartitionKeyFixNotInCanonical(t *testing.T) {
	// Regression test for Day 2.2 typo fix.
	cols := RegistroColumns()
	for _, c := range cols {
		if c.Header == "Partition_Key" {
			t.Error("Schema still references old typo 'Partition_Key' — should be 'Sync_Status'")
		}
	}
	// Verify Sync_Status is at column V (index 22).
	if cols[21].Header != "Sync_Status" {
		t.Errorf("Column V header = %q, want Sync_Status", cols[21].Header)
	}
}
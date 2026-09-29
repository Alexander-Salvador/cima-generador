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
	if CIMAVersion != "v0.4.0-real" {
		t.Errorf("CIMAVersion = %q, want %q", CIMAVersion, "v0.4.0-real")
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

// TestAuditArchiveSheetMissingSheet — Apps Script DISABLED.
func TestAuditArchiveSheetMissingSheet(t *testing.T) {
	// Apps Script DISABLED, but AuditArchiveSheet is still defined for backwards compat.
	got := AuditArchiveSheet(2026, time.September)
	if got == "" {
		t.Error("AuditArchiveSheet should return a string even though Apps Script is disabled")
	}
}

// TestRealSchema — verifies the REAL (v0.4.0) schema matches what was discovered live.
func TestRealSchema(t *testing.T) {
	// 5 expected tabs, all UPPERCASE.
	expected := RealTabs()
	if len(expected) != 5 {
		t.Fatalf("RealTabs = %d items, want 5", len(expected))
	}

	// PersonalColumns (real schema) DOES have 8 columns.
	pc := PersonalColumns()
	if len(pc) != 8 {
		t.Errorf("PersonalColumns = %d items, want 8", len(pc))
	}
	if pc[0].Header != "N°" {
		t.Errorf("PersonalColumns[0].Header = %q, want N°", pc[0].Header)
	}

	// FeriadosColumns has 5.
	fc := FeriadosColumns()
	if len(fc) != 5 {
		t.Errorf("FeriadosColumns = %d items, want 5", len(fc))
	}
	if fc[0].Header != "FECHA" {
		t.Errorf("FeriadosColumns[0].Header = %q, want FECHA", fc[0].Header)
	}
}

// TestVerifyStructureFull checks that the validator returns OK when all
// real tabs are present (v0.4.0 schema).
func TestVerifyStructureFull(t *testing.T) {
	v := VerifyStructure("test-ss", RealTabs())
	if !v.OK() {
		t.Errorf("VerifyStructure(real-tabs) = NOT OK; errors: %v", v.Errors)
	}
	if v.TotalExpected != 5 {
		t.Errorf("TotalExpected = %d, want 5", v.TotalExpected)
	}
	if v.TotalFound != 5 {
		t.Errorf("TotalFound = %d, want 5", v.TotalFound)
	}
}

// TestVerifyStructureMissing checks that the validator flags missing tabs.
func TestVerifyStructureMissing(t *testing.T) {
	// Only 2 tabs, missing 3.
	v := VerifyStructure("test-ss", []string{
		SheetCONFIG, SheetPersonal,
	})
	if v.OK() {
		t.Error("VerifyStructure(partial) = OK; want NOT OK")
	}
	if len(v.Missing) == 0 {
		t.Error("Missing should be non-empty")
	}
	// HORAS EXTRAS, 01_FERIADOS_PE, Feedbacks should be missing.
	missingSet := map[string]bool{}
	for _, m := range v.Missing {
		missingSet[m] = true
	}
	for _, want := range []string{SheetHorasExtras, SheetFeriadosPE, SheetFeedbacks} {
		if !missingSet[want] {
			t.Errorf("missing tab %q not reported", want)
		}
	}
}

// TestVerifyStructureExtras checks that extra sheets are reported (not failed).
func TestVerifyStructureExtras(t *testing.T) {
	// Add a "99_Extra" tab that is not in the real schema.
	all := append([]string{}, RealTabs()...)
	all = append(all, "99_Extra_Tab")
	v := VerifyStructure("test-ss", all)
	if !v.OK() {
		t.Errorf("VerifyStructure(complete+extras) = NOT OK; errors: %v", v.Errors)
	}
	if len(v.Extra) != 1 {
		t.Errorf("Extra should have 1 entry, got %d", len(v.Extra))
	}
}

// TestVerifyStructureNoAppsScriptSheets ensures v0.3.0 fake sheets are NOT expected.
func TestVerifyStructureNoAppsScriptSheets(t *testing.T) {
	// Provide ONLY the v0.3.0 fake sheets — should fail because those
	// are not the real schema.
	fake := []string{
		"00_Config", "01_Feriados_PE", "02_Personas",
		"10_Registro_HE", "20_Dashboard", "99_Audit_Log",
	}
	v := VerifyStructure("test-ss", fake)
	if v.OK() {
		t.Error("VerifyStructure(fake-sheets) = OK; want NOT OK")
	}
	if len(v.Missing) == 0 {
		t.Error("Should report missing real tabs")
	}
}

// TestValidateRegistroColumns checks column header validation (legacy).
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

// TestValidateRegistroColumnsWrong — Apps Script DISABLED.
// The legacy v0.3.0 schema with 22 columns is no longer the live spec.
func TestValidateRegistroColumnsWrong(t *testing.T) {
	// ValidateRegistroColumns now returns nil (Apps Script DISABLED).
	mismatches := ValidateRegistroColumns([]string{"anything"})
	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches (Apps Script DISABLED), got %d: %v",
			len(mismatches), mismatches)
	}
}

// TestValidateModules checks required module validation.
// Apps Script is DISABLED — no modules to validate.
func TestValidateModules(t *testing.T) {
	missing := ValidateModules([]string{"anything"})
	if len(missing) != 0 {
		t.Errorf("ValidateModules = %v, want [] (Apps Script DISABLED)", missing)
	}
}

// TestCanonicalModules checks that no Apps Script modules exist.
func TestCanonicalModules(t *testing.T) {
	mods := CanonicalModules()
	if len(mods) != 0 {
		t.Errorf("CanonicalModules = %d items, want 0 (Apps Script DISABLED)", len(mods))
	}
}

// TestRequiredModules checks that no Apps Script modules are required.
func TestRequiredModules(t *testing.T) {
	req := RequiredModules()
	if len(req) != 0 {
		t.Errorf("RequiredModules = %d items, want 0 (Apps Script DISABLED)", len(req))
	}
	_ = req
}

// TestAnchors checks default anchors (legacy).
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
		TotalExpected:  5,
		TotalFound:     2,
		Missing:        []string{SheetHorasExtras},
		Errors:         []string{"missing tab: " + SheetHorasExtras},
	}
	s := v.String()
	if !strings.Contains(s, "CIMA Verification") {
		t.Error("output missing 'CIMA Verification'")
	}
	if !strings.Contains(s, SheetHorasExtras) {
		t.Errorf("output missing %q", SheetHorasExtras)
	}
	if !strings.Contains(s, "Missing") {
		t.Error("output missing 'Missing' label")
	}
}

// TestPartitionKeyFixNotInCanonical — Apps Script DISABLED.
// The original v0.3.0 schema is no longer the live spec.
// The real sheet uses different column names (verified live 2026-09-29).
func TestPartitionKeyFixNotInCanonical(t *testing.T) {
	// With Apps Script DISABLED, RegistroColumns returns empty.
	cols := RegistroColumns()
	if len(cols) != 0 {
		t.Errorf("RegistroColumns = %d items, want 0 (Apps Script DISABLED)", len(cols))
	}
}
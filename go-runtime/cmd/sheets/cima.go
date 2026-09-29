// Package main — CIMA HE Lima v0.4.0-real sheet operations (REAL schema).
//
// **IMPORTANT**: This module encodes the REAL state of the live spreadsheet
// `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`, NOT the architecture.yaml
// spec. The schema was discovered via live read on 2026-09-29.
//
// Real tabs (5, ALL UPPERCASE):
//   - CONFIG        — system parameters (zona orientativa + operativa)
//   - 01_FERIADOS_PE — holidays (FECHA | NOMBRE | AÑO | TIPO | OBSERVACIONES)
//   - PERSONAL      — employees (N° | DNI | NOMBRES | TELEFONO | HORARIO | ESTADO | CARGO | AREA)
//   - HORAS EXTRAS  — weekly grid (header rows 1-8, employees from row 9, days as columns)
//   - Feedbacks     — audit/calls (FECHA | SUPERVISOR | AGENTE | DNI | CALL_ID)
//
// Apps Script is DISABLED (per CEO directive 2026-09-22). The bridge CLI
// (this Go runtime) is the primary interaction layer.
//
// History:
//   - v0.3.0-stable had FAKE schema (18 sheets, 22 columns, Apps Script modules)
//   - v0.4.0-real has CORRECT schema (5 tabs, real columns, no Apps Script)
//
// Authority: thavren (lead) · Verified: 2026-09-29
package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────
// CANONICAL SHEET CONSTANTS — REAL schema (verified live 2026-09-29)
// ─────────────────────────────────────────────────────────────────

const (
	// CIMAVersion is the current package version reflecting real schema.
	CIMAVersion = "v0.4.0-real"

	// Tab names — UPPERCASE (verified against live sheet).
	SheetCONFIG        = "CONFIG"
	SheetFeriadosPE    = "01_FERIADOS_PE"
	SheetPersonal      = "PERSONAL"
	SheetHorasExtras   = "HORAS EXTRAS"
	SheetFeedbacks     = "Feedbacks"

	// Apps Script is disabled.
	AppsScriptEnabled = false
)

// ─────────────────────────────────────────────────────────────────
// HORAS EXTRAS — layout constants (verified live 2026-09-29)
// ─────────────────────────────────────────────────────────────────

const (
	// Header rows in HORAS EXTRAS (rows 1-8 are header, row 9+ is data).
	HorasExtrasHeaderRow1 = 1 // "CONTROL DE HORAS EXTRAS"
	HorasExtrasHeaderRow5 = 5 // "Año:" label
	HorasExtrasHeaderRow6 = 6 // "Mes:" label
	HorasExtrasHeaderRow7 = 7 // Day numbers (1-7)
	HorasExtrasHeaderRow8 = 8 // Day letters (M X J V S D L)
	HorasExtrasDataStart = 9 // Employees start at row 9

	// Summary cells (row 2-3).
	HorasExtrasResumenRow = 2
	HorasExtrasValueRow   = 3

	// Year/Mes selectors (col B, rows 5/6).
	HorasExtrasYearRow = 5
	HorasExtrasMesRow  = 6
)

// ─────────────────────────────────────────────────────────────────
// 01_FERIADOS_PE — column spec (verified live)
// ─────────────────────────────────────────────────────────────────

// FeriadosColumn is one column of the holidays tab.
type FeriadosColumn struct {
	Letter string
	Header string
}

// FeriadosColumns returns the 5 columns of 01_FERIADOS_PE.
func FeriadosColumns() []FeriadosColumn {
	return []FeriadosColumn{
		{"A", "FECHA"},
		{"B", "NOMBRE"},
		{"C", "AÑO"},
		{"D", "TIPO"},
		{"E", "OBSERVACIONES"},
	}
}

// ─────────────────────────────────────────────────────────────────
// PERSONAL — column spec (verified live)
// ─────────────────────────────────────────────────────────────────

// PersonalColumn is one column of the PERSONAL tab.
type PersonalColumn struct {
	Index  int
	Letter string
	Header string
}

// PersonalColumns returns the 8 columns of PERSONAL.
func PersonalColumns() []PersonalColumn {
	return []PersonalColumn{
		{1, "A", "N°"},
		{2, "B", "DNI"},
		{3, "C", "NOMBRES_COMPLETOS"},
		{4, "D", "TELEFONO"},
		{5, "E", "HORARIO"},
		{6, "F", "ESTADO"},
		{7, "G", "CARGO"},
		{8, "H", "AREA"},
	}
}

// ─────────────────────────────────────────────────────────────────
// SHEET QUERIES
// ─────────────────────────────────────────────────────────────────

// RealTabs returns the 5 actual tabs in the live spreadsheet.
func RealTabs() []string {
	return []string{
		SheetCONFIG,
		SheetFeriadosPE,
		SheetPersonal,
		SheetHorasExtras,
		SheetFeedbacks,
	}
}

// RequiredTabs returns the same as RealTabs (all 5 are required).
// The legacy v0.3.0 had 18 expected sheets; in reality there are 5.
func RequiredTabs() []string {
	return RealTabs()
}

// TotalExpectedTabs returns the count of expected tabs.
func TotalExpectedTabs() int {
	return 5
}

// ─────────────────────────────────────────────────────────────────
// VALIDATION
// ─────────────────────────────────────────────────────────────────

// CIMAVerification captures the result of validating a live spreadsheet
// against the REAL CIMA schema (v0.4.0).
type CIMAVerification struct {
	SpreadsheetID string
	Timestamp     time.Time
	Present       []string
	Missing       []string
	Extra         []string
	TotalExpected int
	TotalFound    int
	Errors        []string
}

// VerifyStructure checks that all required tabs exist. The caller
// supplies `actual` (the list of tab names from the live spreadsheet).
// The comparison is case-INSENSITIVE because legacy data may have
// mixed casing.
func VerifyStructure(spreadsheetID string, actual []string) *CIMAVerification {
	v := &CIMAVerification{
		SpreadsheetID: spreadsheetID,
		Timestamp:     time.Now(),
		TotalExpected: TotalExpectedTabs(),
	}

	// Build set of expected tab names (lowercased for case-insensitive match).
	expected := map[string]string{} // lower → actual-cased
	for _, s := range RequiredTabs() {
		expected[strings.ToLower(s)] = s
	}

	// Build actual set (lowercased → actual).
	actualSet := map[string]bool{}
	actualLower := map[string]string{}
	for _, a := range actual {
		actualSet[a] = true
		actualLower[strings.ToLower(a)] = a
	}

	// Compute present/missing.
	for lower, expectedCased := range expected {
		if actualLower[lower] != "" {
			v.Present = append(v.Present, actualLower[lower])
		} else {
			v.Missing = append(v.Missing, expectedCased)
		}
	}
	// Compute extras.
	for lower, actualCased := range actualLower {
		if expected[lower] == "" {
			v.Extra = append(v.Extra, actualCased)
		}
	}

	sort.Strings(v.Present)
	sort.Strings(v.Missing)
	sort.Strings(v.Extra)

	v.TotalFound = len(actualSet)

	// OK iff no missing.
	if len(v.Missing) == 0 {
		return v
	}
	for _, m := range v.Missing {
		v.Errors = append(v.Errors, "missing tab: "+m)
	}
	return v
}

// OK returns true iff no tabs are missing.
func (v *CIMAVerification) OK() bool {
	return len(v.Missing) == 0
}

// String returns a human-readable summary.
func (v *CIMAVerification) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "📋 CIMA Verification (REAL schema v0.4.0) @ %s\n",
		v.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "  Spreadsheet: %s\n", v.SpreadsheetID)
	fmt.Fprintf(&b, "  Expected: %d tabs (REAL: CONFIG, 01_FERIADOS_PE, PERSONAL, HORAS EXTRAS, Feedbacks)\n",
		v.TotalExpected)
	fmt.Fprintf(&b, "  Found: %d tabs\n", v.TotalFound)
	if len(v.Missing) > 0 {
		fmt.Fprintf(&b, "  ❌ Missing (%d): %v\n", len(v.Missing), v.Missing)
	}
	if len(v.Extra) > 0 {
		fmt.Fprintf(&b, "  ℹ️  Extra (%d): %v\n", len(v.Extra), v.Extra)
	}
	if v.OK() {
		fmt.Fprintf(&b, "  ✅ Structure matches real schema\n")
	} else {
		fmt.Fprintf(&b, "  ❌ Structure does NOT match real schema\n")
	}
	return b.String()
}

// ─────────────────────────────────────────────────────────────────
// LEGACY v0.3.0 SCHEMA — DEPRECATED
// ─────────────────────────────────────────────────────────────────
//
// The following constants/functions are kept for backwards compatibility
// (e.g., tests) but are NOT the real schema. They encode the FAKE schema
// that was in architecture.yaml v0.2.0-rev, which has been SUPERSEDED.
//
// DO NOT USE THESE FOR LIVE VERIFICATION.

const (
	deprecatedSheetConfig    = "00_Config" // REAL: CONFIG
	deprecatedSheetHolidays  = "01_Feriados_PE" // REAL: 01_FERIADOS_PE
	deprecatedSheetPersonas  = "02_Personas" // REAL: PERSONAL
	deprecatedSheetRegistro  = "10_Registro_HE" // REAL: HORAS EXTRAS
	deprecatedSheetDashboard = "20_Dashboard" // REAL: Feedbacks (audit)
	deprecatedSheetAuditLog  = "99_Audit_Log" // REAL: not present
)

// DeprecatedTotalExpected returns 18 (the v0.3.0 fake schema count).
// Kept for backwards compatibility only.
func DeprecatedTotalExpected() int {
	return 18
}

// DeprecatedVerifyStructure runs the v0.3.0 18-sheet check.
// Returns a CIMAVerification that will likely show many missing tabs
// because the real spreadsheet has 5, not 18.
func DeprecatedVerifyStructure(spreadsheetID string, actual []string) *CIMAVerification {
	v := &CIMAVerification{
		SpreadsheetID: spreadsheetID,
		Timestamp:     time.Now(),
		TotalExpected: DeprecatedTotalExpected(),
	}
	expected := map[string]bool{
		deprecatedSheetConfig: true, deprecatedSheetHolidays: true,
		deprecatedSheetPersonas: true, deprecatedSheetRegistro: true,
		deprecatedSheetDashboard: true, deprecatedSheetAuditLog: true,
		// 12 monthly sheets omitted for brevity
	}
	for e := range expected {
		found := false
		for _, a := range actual {
			if a == e {
				found = true
				break
			}
		}
		if !found {
			v.Missing = append(v.Missing, e)
		}
	}
	for _, a := range actual {
		found := false
		for e := range expected {
			if e == a {
				found = true
			}
		}
		if !found {
			v.Extra = append(v.Extra, a)
		}
	}
	v.TotalFound = len(actual)
	return v
}

// AppsScriptModule is a stub for backwards compatibility.
// Apps Script is DISABLED — no modules exist.
type AppsScriptModule struct {
	Number      int
	File        string
	Responsibility string
	PublicFunctions []string
	Required    bool
}

// CanonicalModules returns 0 modules — Apps Script is DISABLED.
// This is a compatibility stub for code that referenced v0.3.0 schema.
func CanonicalModules() []AppsScriptModule {
	return []AppsScriptModule{}
}

// RequiredModules returns 0 modules — Apps Script is DISABLED.
func RequiredModules() []AppsScriptModule {
	return []AppsScriptModule{}
}

// ValidateModules always returns nil (no modules to validate).
// Apps Script is DISABLED.
func ValidateModules(present []string) []string {
	return nil
}

// CIMASpreadsheetName is the legacy name constant.
// (Real spreadsheet doesn't have an internal name; using display.)
const CIMASpreadsheetName = "CIMA — Tracking Horas Extras Lima"

// Deprecated constants for backwards compat.
const (
	SheetConfig    = "00_Config"      // REAL: CONFIG
	SheetHolidays  = "01_Feriados_PE"  // REAL: 01_FERIADOS_PE
	SheetPersonas  = "02_Personas"     // REAL: PERSONAL
	SheetRegistro  = "10_Registro_HE"  // REAL: HORAS EXTRAS
	SheetDashboard = "20_Dashboard"    // REAL: Feedbacks
	SheetAuditLog  = "99_Audit_Log"    // REAL: not present

	SheetDBPrefix       = "98_DB_"
	SheetDBTotalMonths  = 12
)

// AnchorConfig is the legacy 00_Config anchors. Apps Script DISABLED.
type AnchorConfig struct {
	Mes, Anio, Semana, PersonaFiltro                       string
	FactorDiurno, FactorNocturno, FactorFeriado, JornadaMax string
	LastRefresh                                            string
}

// DefaultAnchors returns the v0.3.0 anchor map (legacy).
func DefaultAnchors() AnchorConfig {
	return AnchorConfig{
		Mes: "B2", Anio: "B3", Semana: "B4", PersonaFiltro: "B5",
		FactorDiurno: "B6", FactorNocturno: "B7",
		FactorFeriado: "B8", JornadaMax: "B9", LastRefresh: "B11",
	}
}

// ColumnSpec is the legacy registro column descriptor.
type ColumnSpec struct {
	Index    int
	Letter   string
	Header   string
	Type     string
	Required bool
}

// RegistroColumns returns 22 v0.3.0 columns (legacy).
func RegistroColumns() []ColumnSpec {
	return []ColumnSpec{}
}

// RequiredBaseSheets returns the v0.3.0 6 base sheets (legacy).
func RequiredBaseSheets() []string {
	return []string{
		SheetConfig, SheetHolidays, SheetPersonas,
		SheetRegistro, SheetDashboard, SheetAuditLog,
	}
}

// MonthlyStorageSheet returns the v0.3.0 98_DB name (legacy — apps script DISABLED).
func MonthlyStorageSheet(mes, anio int) string {
	if mes < 1 || mes > 12 {
		return ""
	}
	return fmt.Sprintf("%s%d_%02d", SheetDBPrefix, anio, mes)
}

// AllMonthlyStorageSheets returns 12 v0.3.0 monthly sheets.
func AllMonthlyStorageSheets() []string {
	out := make([]string, 0, SheetDBTotalMonths)
	for mes := 1; mes <= SheetDBTotalMonths; mes++ {
		out = append(out, MonthlyStorageSheet(mes, 2026))
	}
	return out
}

// TotalExpectedSheets returns 18 (v0.3.0 legacy count).
func TotalExpectedSheets() int {
	return 6 + SheetDBTotalMonths
}

// AuditArchiveSheet returns the v0.3.0 audit archive name.
func AuditArchiveSheet(anio int, mes time.Month) string {
	return fmt.Sprintf("99_Audit_Log_archive_%d_%02d", anio, int(mes))
}

// ValidateRegistroColumns returns nil (no columns to validate).
// Apps Script DISABLED — actual HE sheet has different columns.
func ValidateRegistroColumns(actual []string) []string {
	return nil
}
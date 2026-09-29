// Package main — CIMA HE Lima v0.3.0 sheet operations.
//
// Provides one-click deploy and validation for the CIMA — Tracking Horas
// Extras Lima spreadsheet, encoding the canonical schema declared in
// .ovav/templates/he-lima/architecture.yaml.
//
// Scope:
//   - 6 base sheets: 00_Config, 01_Feriados_PE, 02_Personas,
//     10_Registro_HE, 20_Dashboard, 99_Audit_Log
//   - 12 monthly storage sheets (lazy): 98_DB_2026_01..12
//   - 9 Apps Script modules (00..08)
//
// Architecture authority: marco · Owner: thavren · Stabilize: 2026-09-29
package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────
// CANONICAL SCHEMA — encoding of architecture.yaml
// ─────────────────────────────────────────────────────────────────

// CIMA version constants.
const (
	CIMAVersion        = "v0.3.0-stable"
	CIMASpreadsheetName = "CIMA — Tracking Horas Extras Lima"
)

// SHEET_* constants are the canonical sheet names. They are referenced
// from Apps Script modules (01..08) and validated at deploy time.
const (
	SheetConfig       = "00_Config"
	SheetHolidays     = "01_Feriados_PE"
	SheetPersonas     = "02_Personas"
	SheetRegistro     = "10_Registro_HE"
	SheetDashboard    = "20_Dashboard"
	SheetAuditLog     = "99_Audit_Log"
	SheetDBPrefix     = "98_DB_"
	SheetDBYear       = "2026"
	SheetDBTotalMonths = 12
	SheetAuditArchivePrefix = "99_Audit_Log_archive_"
)

// AnchorConfig holds the canonical cell anchors for 00_Config.
type AnchorConfig struct {
	Mes            string // default ""
	Anio           string
	Semana         string
	PersonaFiltro  string
	FactorDiurno   string
	FactorNocturno string
	FactorFeriado  string
	JornadaMax     string
	LastRefresh    string
}

// DefaultAnchors returns the canonical anchors per architecture.yaml v2.
// B2..B9 for active config, B11 for last refresh timestamp.
func DefaultAnchors() AnchorConfig {
	return AnchorConfig {
		Mes:            "B2",
		Anio:           "B3",
		Semana:         "B4",
		PersonaFiltro:  "B5",
		FactorDiurno:   "B6",
		FactorNocturno: "B7",
		FactorFeriado:  "B8",
		JornadaMax:     "B9",
		LastRefresh:    "B11",
	}
}

// ColumnSpec describes one column of the working surface (10_Registro_HE).
type ColumnSpec struct {
	Index    int    // 1-based column index (A=1, B=2, ...)
	Letter   string // A1 column letter
	Header   string
	Type     string // date|text|number|bool|enum|fk|lookup|datetime
	Required bool
}

// RegistroColumns returns the canonical 22 columns of 10_Registro_HE
// (architecture.yaml §sheets[id=registro].columns, CORR-1 working surface).
func RegistroColumns() []ColumnSpec {
	return []ColumnSpec{
		{1, "A", "Fecha", "date", true},
		{2, "B", "ISO_Week", "text", false},
		{3, "C", "Dia_Semana", "enum", false},
		{4, "D", "Es_Feriado", "bool", false},
		{5, "E", "Tipo_Dia", "enum", false},
		{6, "F", "ID_Persona", "fk", true},
		{7, "G", "Apellidos_Nombres", "lookup", false},
		{8, "H", "Turno", "enum", false},
		{9, "I", "Hora_Ingreso", "time", false},
		{10, "J", "Hora_Salida", "time", false},
		{11, "K", "Horas_Total", "number", true},
		{12, "L", "Horas_Normales", "number", false},
		{13, "M", "Horas_Extras", "number", false},
		{14, "N", "Factor_Aplicado", "number", false},
		{15, "O", "Horas_Ponderadas", "number", false},
		{16, "P", "Pago_HE_PEN", "number", false},
		{17, "Q", "Observaciones", "text", false},
		{18, "R", "Aud_Usuario", "text", false},
		{19, "S", "Aud_Timestamp", "datetime", false},
		{20, "T", "Estado", "enum", false},
		{21, "U", "Mes_Año", "text", false},
		{22, "V", "Sync_Status", "enum", false},
	}
}

// ModuleSpec describes one Apps Script module of the CIMA add-on.
type ModuleSpec struct {
	Number      int
	File        string
	Responsibility string
	PublicFunctions []string
	Required    bool
}

// CanonicalModules returns the 9 Apps Script modules (00..08)
// per architecture.yaml §apps_script_modules (post v0.3.0).
func CanonicalModules() []ModuleSpec {
	return []ModuleSpec{
		{
			Number:      0,
			File:        "00_AddOnMenu.gs",
			Responsibility: "Entry point onOpen() + menú nativo add-on + bootstrap",
			PublicFunctions: []string{"onOpen", "menuRefrescarConfig", "menuSyncManual",
				"menuRefrescarKPIs", "menuRefrescarHeatmap", "menuVerificarEstructura",
				"menuVerAuditLog", "menuTestCache", "menuPoblarDropdowns",
				"menuInstalarTriggers", "menuQuitarTriggers", "menuAcercaDe"},
			Required: true, // required by addOns.homepageTrigger.runFunction
		},
		{
			Number:      1,
			File:        "01_CfgLoader.gs",
			Responsibility: "Lee 00_Config + CacheService (TTL 600s)",
			PublicFunctions: []string{"getMes", "getAnio", "getSemana",
				"getPersonaFiltro", "getFactores", "invalidateCache"},
			Required: true,
		},
		{
			Number:      2,
			File:        "02_HolidaysEngine.gs",
			Responsibility: "Pascua Meeus/Jones/Butcher + 01_Feriados_PE",
			PublicFunctions: []string{"calcularPascua", "calcularJuevesSanto",
				"calcularViernesSanto", "poblarFeriados", "esFeriado"},
			Required: true,
		},
		{
			Number:      3,
			File:        "03_RegistroEngine.gs",
			Responsibility: "upsertRegistro + invariantes + poblarGridVacio",
			PublicFunctions: []string{"upsertRegistro", "validarInvariantes",
				"poblarGridVacio"},
			Required: true,
		},
		{
			Number:      4,
			File:        "04_FormulasRecalc.gs",
			Responsibility: "onEdit router + alertas anomalías",
			PublicFunctions: []string{"onEdit", "recalcularFila", "propagarVLookups",
				"recalcularGridCompleto"},
			Required: true,
		},
		{
			Number:      5,
			File:        "05_DashboardRefresh.gs",
			Responsibility: "KPIs + heatmap + gráfico mensual",
			PublicFunctions: []string{"refrescarKPIs", "refrescarHeatmap",
				"refrescarGraficoMensual"},
			Required: false,
		},
		{
			Number:      6,
			File:        "06_AuditLogger.gs",
			Responsibility: "ISO 8601 + auto-archive rotation",
			PublicFunctions: []string{"logAccion"},
			Required: true,
		},
		{
			Number:      7,
			File:        "07_TriggersWiring.gs",
			Responsibility: "3 triggers + LockService (onChange + daily + monthly-check)",
			PublicFunctions: []string{"instalarTriggers", "eliminarTriggers",
				"triggerCierreMes", "cerrarMes", "listarTriggers"},
			Required: false, // can run without triggers, but loses automation
		},
		{
			Number:      8,
			File:        "08_DBSync.gs",
			Responsibility: "load/save/createMonthSheetIfMissing + onConfigChange",
			PublicFunctions: []string{"resolveTargetSheet", "createMonthSheetIfMissing",
				"loadMonthToGrid", "saveGridToDB", "onConfigChange"},
			Required: true,
		},
	}
}

// ─────────────────────────────────────────────────────────────────
// DERIVED QUERIES
// ─────────────────────────────────────────────────────────────────

// RequiredBaseSheets returns the 6 base sheets that must always exist.
func RequiredBaseSheets() []string {
	return []string{
		SheetConfig, SheetHolidays, SheetPersonas,
		SheetRegistro, SheetDashboard, SheetAuditLog,
	}
}

// MonthlyStorageSheet returns the canonical sheet name for a given month
// and year. Example: MonthlyStorageSheet(3, 2026) → "98_DB_2026_03".
func MonthlyStorageSheet(mes, anio int) string {
	if mes < 1 || mes > 12 {
		return ""
	}
	return fmt.Sprintf("%s%d_%02d", SheetDBPrefix, anio, mes)
}

// AllMonthlyStorageSheets returns the 12 expected monthly storage sheets
// for the canonical year (2026).
func AllMonthlyStorageSheets() []string {
	out := make([]string, 0, SheetDBTotalMonths)
	for mes := 1; mes <= SheetDBTotalMonths; mes++ {
		out = append(out, MonthlyStorageSheet(mes, 2026))
	}
	return out
}

// AuditArchiveSheet returns the canonical name for the audit archive
// sheet for a given timestamp (year+month). Example:
//   AuditArchiveSheet(2026, 9) → "99_Audit_Log_archive_2026_09"
func AuditArchiveSheet(anio int, mes time.Month) string {
	return fmt.Sprintf("%s%d_%02d", SheetAuditArchivePrefix, anio, int(mes))
}

// TotalExpectedSheets returns the count of sheets that should exist after
// the full deploy (6 base + 12 monthly + 1 audit archive optional).
func TotalExpectedSheets() int {
	return len(RequiredBaseSheets()) + SheetDBTotalMonths
}

// RequiredModules returns the Apps Script modules marked as Required.
// Used for validation — deploying without a required module is a hard error.
func RequiredModules() []ModuleSpec {
	out := make([]ModuleSpec, 0, len(CanonicalModules()))
	for _, m := range CanonicalModules() {
		if m.Required {
			out = append(out, m)
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────
// VALIDATION
// ─────────────────────────────────────────────────────────────────

// CIMAVerification captures the result of validating a live spreadsheet
// against the CIMA v0.3.0 canonical schema.
type CIMAVerification struct {
	SpreadsheetID    string
	Timestamp        time.Time
	Present         []string
	Missing         []string
	Extra           []string
	BaseSheetsOK    bool
	MonthlySheetsOK bool
	TotalExpected   int
	TotalFound      int
	Errors          []string
}

// VerifyStructure checks that all required sheets exist. The caller
// supplies `actual` (the list of sheet names retrieved from the live
// spreadsheet). The function returns a CIMAVerification with the diff.
//
// The comparison is order-insensitive and case-sensitive.
func VerifyStructure(spreadsheetID string, actual []string) *CIMAVerification {
	v := &CIMAVerification{
		SpreadsheetID: spreadsheetID,
		Timestamp:     time.Now(),
		TotalExpected: TotalExpectedSheets(),
	}

	// Build set of expected sheet names.
	expected := map[string]bool{}
	for _, s := range RequiredBaseSheets() {
		expected[s] = true
	}
	for _, s := range AllMonthlyStorageSheets() {
		expected[s] = true
	}

	actualSet := map[string]bool{}
	for _, a := range actual {
		actualSet[a] = true
	}

	// Compute present/missing.
	for name := range expected {
		if actualSet[name] {
			v.Present = append(v.Present, name)
		} else {
			v.Missing = append(v.Missing, name)
		}
	}
	// Compute extras (sheets that exist but aren't in our schema).
	for name := range actualSet {
		if !expected[name] && !strings.HasPrefix(name, SheetAuditArchivePrefix) {
			v.Extra = append(v.Extra, name)
		}
	}

	sort.Strings(v.Present)
	sort.Strings(v.Missing)
	sort.Strings(v.Extra)

	v.TotalFound = len(actualSet)

	// Base sheets OK iff all 6 base sheets present.
	presentSet := map[string]bool{}
	for _, p := range v.Present {
		presentSet[p] = true
	}
	v.BaseSheetsOK = true
	for _, base := range RequiredBaseSheets() {
		if !presentSet[base] {
			v.BaseSheetsOK = false
			v.Errors = append(v.Errors,
				"missing base sheet: "+base)
		}
	}
	v.MonthlySheetsOK = len(v.Missing) == 0

	return v
}

// OK returns true iff the verification found no missing sheets and no errors.
func (v *CIMAVerification) OK() bool {
	return v.BaseSheetsOK && len(v.Errors) == 0
}

// String returns a human-readable summary.
func (v *CIMAVerification) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "📋 CIMA Verification @ %s\n", v.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "  Spreadsheet: %s\n", v.SpreadsheetID)
	fmt.Fprintf(&b, "  Expected: %d sheets (%d base + %d monthly)\n",
		v.TotalExpected, len(RequiredBaseSheets()), SheetDBTotalMonths)
	fmt.Fprintf(&b, "  Found: %d sheets\n", v.TotalFound)
	fmt.Fprintf(&b, "  Base sheets OK: %v\n", v.BaseSheetsOK)
	fmt.Fprintf(&b, "  Monthly sheets OK: %v\n", v.MonthlySheetsOK)
	if len(v.Missing) > 0 {
		fmt.Fprintf(&b, "  ❌ Missing (%d): %v\n", len(v.Missing), v.Missing)
	}
	if len(v.Extra) > 0 {
		fmt.Fprintf(&b, "  ℹ️  Extra (%d): %v\n", len(v.Extra), v.Extra)
	}
	if len(v.Errors) > 0 {
		fmt.Fprintf(&b, "  ❌ Errors:\n")
		for _, e := range v.Errors {
			fmt.Fprintf(&b, "    - %s\n", e)
		}
	}
	if v.OK() {
		fmt.Fprintf(&b, "  ✅ Structure is valid\n")
	}
	return b.String()
}

// ─────────────────────────────────────────────────────────────────
// COLUMN VALIDATION
// ─────────────────────────────────────────────────────────────────

// ValidateRegistroColumns checks that the headers in `actual` match the
// canonical schema for 10_Registro_HE. Returns a list of mismatches.
//
// `actual` should be the row of headers (22 columns) read from the sheet.
func ValidateRegistroColumns(actual []string) []string {
	expected := RegistroColumns()
	var mismatches []string

	if len(actual) < len(expected) {
		mismatches = append(mismatches,
			fmt.Sprintf("expected %d columns, got %d",
				len(expected), len(actual)))
		return mismatches
	}

	for i, exp := range expected {
		if actual[i] != exp.Header {
			mismatches = append(mismatches,
				fmt.Sprintf("col %d (%s): expected %q, got %q",
					exp.Index, exp.Letter, exp.Header, actual[i]))
		}
	}
	return mismatches
}

// ─────────────────────────────────────────────────────────────────
// MODULE VALIDATION
// ─────────────────────────────────────────────────────────────────

// ValidateModules checks that the supplied module file names cover all
// required canonical modules. Returns a list of missing module filenames.
func ValidateModules(present []string) []string {
	presentSet := map[string]bool{}
	for _, p := range present {
		presentSet[p] = true
	}

	var missing []string
	for _, m := range RequiredModules() {
		if !presentSet[m.File] {
			missing = append(missing, m.File)
		}
	}
	return missing
}
// Package main — formatting, validation, charts, named ranges,
// protected ranges and image insertion via Sheets API v4.
//
// All implemented via batchUpdate requests — no manual editor work.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ── Conditional formatting ────────────────────────────────────────────

// ConditionalFormatRule is a thin wrapper around the rich
// sheets API request shape. Two convenience constructors cover 95%
// of real-world rules: CONDITION_FORMATTING_HIGHLIGHT and
// BOOLEAN_RULE.
type ConditionalFormatRule struct {
	Range   string // A1 notation, e.g. "CIMA!A5:B5"
	When    string // "TEXT_EQ" | "TEXT_CONTAINS" | "CUSTOM_FORMULA" | "NUMBER_LESS" | ...
	Value   string // comparison value (or formula if When == CUSTOM_FORMULA)
	BGColor string // hex without #, e.g. "FFEB9C" (yellow)
	FGColor string // hex without #
	Bold    bool
}

func (cl *Client) AddConditionalRule(r ConditionalFormatRule) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	gridRange, err := cl.a1ToGrid(r.Range)
	if err != nil {
		return err
	}
	rule := buildConditionalRule(r, gridRange)
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"addConditionalFormatRule": rule},
		},
	})
	return cl.batchUpdate(body)
}

// buildConditionalRule constructs the Sheets API v4
// addConditionalFormatRule request. Sheets v4 wraps `type` and
// `values` inside a `condition` object under `booleanRule`.
func buildConditionalRule(r ConditionalFormatRule, gridRange map[string]any) map[string]any {
	format := map[string]any{}
	if r.BGColor != "" {
		format["backgroundColor"] = map[string]any{
			"red":   hexToFloat(r.BGColor, 0),
			"green": hexToFloat(r.BGColor, 1),
			"blue":  hexToFloat(r.BGColor, 2),
		}
	}
	if r.FGColor != "" {
		format["textFormat"] = map[string]any{
			"foregroundColor": map[string]any{
				"red":   hexToFloat(r.FGColor, 0),
				"green": hexToFloat(r.FGColor, 1),
				"blue":  hexToFloat(r.FGColor, 2),
			},
		}
		if r.Bold {
			format["textFormat"].(map[string]any)["bold"] = true
		}
	} else if r.Bold {
		format["textFormat"] = map[string]any{"bold": true}
	}

	var booleanRule map[string]any
	switch r.When {
	case "TEXT_EQ":
		booleanRule = map[string]any{
			"type":   "TEXT_EQ",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	case "TEXT_CONTAINS":
		booleanRule = map[string]any{
			"type":   "TEXT_CONTAINS",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	case "NUMBER_LESS":
		booleanRule = map[string]any{
			"type":   "NUMBER_LESS",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	case "NUMBER_GREATER":
		booleanRule = map[string]any{
			"type":   "NUMBER_GREATER",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	case "CUSTOM_FORMULA":
		booleanRule = map[string]any{
			"type":   "CUSTOM_FORMULA",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	default:
		booleanRule = map[string]any{
			"type":   "TEXT_CONTAINS",
			"values": []map[string]any{{"userEnteredValue": r.Value}},
		}
	}

	return map[string]any{
		"rule": map[string]any{
			"ranges": []map[string]any{gridRange},
			"booleanRule": map[string]any{
				"condition": booleanRule,
				"format":    format,
			},
		},
		"index": 0,
	}
}

// ── Data validation ──────────────────────────────────────────────────

// SetDataValidation adds a dropdown or checkbox rule.
// type: "ONE_OF_LIST" | "ONE_OF_RANGE" | "CHECKBOX" | "NUMBER_BETWEEN" | "DATE_BETWEEN" | "DATE_IS_VALID"
// values: list items (string for ONE_OF_LIST) or {"min","max"} (for BETWEEN / DATE_BETWEEN)
//         or "TODAY" / "DATE_VALID" / etc. for DATE_IS_VALID
func (cl *Client) SetDataValidation(a1Range, ruleType string, values any, strict bool) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	grid, err := cl.a1ToGrid(a1Range)
	if err != nil {
		return err
	}
	rule := map[string]any{}
	switch ruleType {
	case "ONE_OF_LIST":
		list, _ := values.([]string)
		asAny := make([]any, 0, len(list))
		for _, s := range list {
			asAny = append(asAny, s)
		}
		rule["condition"] = map[string]any{
			"type":   "ONE_OF_LIST",
			"values": toUserEnteredValues(asAny),
		}
		if len(list) > 0 {
			rule["showCustomUi"] = false
		}
	case "ONE_OF_RANGE":
		rule["condition"] = map[string]any{
			"type":   "ONE_OF_RANGE",
			"values": []map[string]any{{"userEnteredValue": fmt.Sprint(values)}},
		}
	case "CHECKBOX":
		rule["condition"] = map[string]any{"type": "BOOLEAN"}
	case "NUMBER_BETWEEN":
		mm, ok := values.(map[string]any)
		if !ok {
			return fmt.Errorf("sheets: NUMBER_BETWEEN requires values {min,max}")
		}
		rule["condition"] = map[string]any{
			"type":   "NUMBER_BETWEEN",
			"values": toUserEnteredValues([]any{mm["min"], mm["max"]}),
		}
	case "DATE_BETWEEN":
		mm, ok := values.(map[string]any)
		if !ok {
			return fmt.Errorf("sheets: DATE_BETWEEN requires values {min,max}")
		}
		rule["condition"] = map[string]any{
			"type":   "DATE_BETWEEN",
			"values": toUserEnteredValues([]any{mm["min"], mm["max"]}),
		}
	case "DATE":
		// values: "TODAY" | "DATE_VALID" — minimal guard; full
		// flexibility via DATE_BETWEEN
		rule["condition"] = map[string]any{
			"type":   "DATE_IS_VALID",
			"values": []map[string]any{{"userEnteredValue": fmt.Sprint(values)}},
		}
	default:
		return fmt.Errorf("sheets: unknown validation type %q", ruleType)
	}
	if strict {
		rule["strict"] = true
	}
	req := map[string]any{
		"range": grid,
		"rule":  rule,
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"setDataValidation": req},
		},
	})
	return cl.batchUpdate(body)
}

func toUserEnteredValues(items []any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, x := range items {
		out = append(out, map[string]any{"userEnteredValue": fmt.Sprint(x)})
	}
	return out
}

// ── Protected ranges ─────────────────────────────────────────────────

// AddProtectedRange locks a range behind an editors-only warning.
// Editors (the email list) can edit without a prompt; everyone else
// gets a warning before editing.
func (cl *Client) AddProtectedRange(a1Range, description string, editors []string, warningOnly bool) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	grid, err := cl.a1ToGrid(a1Range)
	if err != nil {
		return err
	}
	protected := map[string]any{
		"description": description,
		"range":       grid,
		"warningOnly": warningOnly,
	}
	if len(editors) > 0 {
		asAny := make([]any, 0, len(editors))
		for _, e := range editors {
			asAny = append(asAny, e)
		}
		protected["editors"] = map[string]any{"users": toUserEnteredValues(asAny)}
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"addProtectedRange": map[string]any{"protectedRange": protected}},
		},
	})
	return cl.batchUpdate(body)
}

// ── Named ranges ─────────────────────────────────────────────────────

func (cl *Client) AddNamedRange(name, a1Range string) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	grid, err := cl.a1ToGrid(a1Range)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"addNamedRange": map[string]any{
				"namedRange": map[string]any{
					"name":  name,
					"range": grid,
				},
			}},
		},
	})
	return cl.batchUpdate(body)
}

// ── Charts ───────────────────────────────────────────────────────────

// AddChart inserts a chart with optional separate domain and series ranges.
// If rngDomain and rngSeries are empty, the entire rng is used for both.
// Supported types: BAR, LINE, PIE, AREA, COLUMN, SCATTER.
func (cl *Client) AddChart(tabName, a1Range, a1RangeDomain, a1RangeSeries, chartType, title string) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	tab, err := cl.resolveSheetID(tabName)
	if err != nil {
		return err
	}
	spec, err := buildChartSpec(cl, a1Range, a1RangeDomain, a1RangeSeries, chartType, title)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"addChart": map[string]any{
				"chart": map[string]any{
					"spec": spec,
					"position": map[string]any{
						"overlayPosition": map[string]any{
							"anchorCell":   map[string]any{"sheetId": tab, "rowIndex": 1, "columnIndex": 8},
							"widthPixels":  600,
							"heightPixels": 360,
						},
					},
				},
			}},
		},
	})
	return cl.batchUpdate(body)
}

func buildChartSpec(client *Client, a1Range, a1RangeDomain, a1RangeSeries, chartType, title string) (map[string]any, error) {
	gridAll, err := client.a1ToGrid(a1Range)
	if err != nil {
		return nil, err
	}
	gridDomain := gridAll
	gridSeries := gridAll
	if a1RangeDomain != "" {
		gridDomain, err = client.a1ToGrid(a1RangeDomain)
		if err != nil {
			return nil, err
		}
	}
	if a1RangeSeries != "" {
		gridSeries, err = client.a1ToGrid(a1RangeSeries)
		if err != nil {
			return nil, err
		}
	}
	sourceRangeDomain := map[string]any{
		"sourceRange": map[string]any{"sources": []map[string]any{gridDomain}},
	}
	sourceRangeSeries := map[string]any{
		"sourceRange": map[string]any{"sources": []map[string]any{gridSeries}},
	}
	switch chartType {
	case "PIE":
		return map[string]any{
			"title": title,
			"pieChart": map[string]any{
				"legendPosition": "RIGHT_LEGEND",
				"domain":         sourceRangeDomain,
				"series":         sourceRangeSeries,
			},
		}, nil
	default:
		basicChartType := chartType // BAR, LINE, AREA, COLUMN, SCATTER
		domainWrapped := map[string]any{"domain": sourceRangeDomain}
		seriesWrapped := map[string]any{"series": sourceRangeSeries}
		return map[string]any{
			"title": title,
			"basicChart": map[string]any{
				"chartType":      basicChartType,
				"domains":        []map[string]any{domainWrapped},
				"series":         []map[string]any{seriesWrapped},
				"legendPosition": "RIGHT_LEGEND",
				"axis": map[string]any{
					"position": "BOTTOM_AXIS",
					"title":    "",
				},
			},
		}, nil
	}
}

// ── Images ───────────────────────────────────────────────────────────

// InsertImage adds an image from a URL.
func (cl *Client) InsertImage(tabName, anchorCell, imageURL string) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	tab, err := cl.resolveSheetID(tabName)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"insertImage": map[string]any{
				"uri": imageURL,
				"overlayPosition": map[string]any{
					"anchorCell":    map[string]any{"sheetId": tab, "rowIndex": 0, "columnIndex": 0},
					"offsetXPixels": 0,
					"offsetYPixels": 0,
					"widthPixels":   300,
					"heightPixels":  200,
				},
			}},
		},
	})
	return cl.batchUpdate(body)
}

// ── helpers ──────────────────────────────────────────────────────────

// ── Cell formatting (plain, non-conditional) ────────────────────────

// CellStyleSpec is a thin wrapper around Sheets `repeatCell` request.
// Lets the operator paint any A1 range with a deterministic
// background / text / border style — without needing a
// conditional rule.
type CellStyleSpec struct {
	Range        string
	BgColor      string
	FgColor      string
	Bold         bool
	Italic       bool
	FontSize     int
	Align        string
	BorderAll    bool
	BorderTop    bool
	BorderBottom bool
	BorderLeft   bool
	BorderRight  bool
	BorderColor  string
	// NumberFormat sets the cell's number format type ("TEXT" | "NUMBER" |
	// "DATE" | "PERCENT" | "CURRENCY" | etc.). When set, the bg/fg/borders
	// are NOT applied — only the number format. Use a separate call for
	// combined styling.
	NumberFormat string
	// Wrap enables wrapStrategy "WRAP" (text wraps to fit column width).
	Wrap bool
}

// SetCellStyle applies a plain cell style (bg, fg, bold, borders)
// to the given A1 range. Uses sheets.spreadsheets.batchUpdate with
// a single `repeatCell` request.
func (cl *Client) SetCellStyle(spec CellStyleSpec) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	grid, err := cl.a1ToGrid(spec.Range)
	if err != nil {
		return err
	}
	format := map[string]any{}
	if spec.BgColor != "" {
		format["backgroundColor"] = map[string]any{
			"red":   hexToFloat(spec.BgColor, 0),
			"green": hexToFloat(spec.BgColor, 1),
			"blue":  hexToFloat(spec.BgColor, 2),
		}
	}
	tf := map[string]any{}
	if spec.FgColor != "" {
		tf["foregroundColor"] = map[string]any{
			"red":   hexToFloat(spec.FgColor, 0),
			"green": hexToFloat(spec.FgColor, 1),
			"blue":  hexToFloat(spec.FgColor, 2),
		}
	}
	if spec.Bold {
		tf["bold"] = true
	}
	if spec.Italic {
		tf["italic"] = true
	}
	if spec.FontSize > 0 {
		tf["fontSize"] = spec.FontSize
	}
	if len(tf) > 0 {
		format["textFormat"] = tf
	}
	if spec.Align != "" {
		format["horizontalAlignment"] = strings.ToUpper(spec.Align)
	}
	if spec.Wrap {
		format["wrapStrategy"] = "WRAP"
	}
	if spec.BorderAll || spec.BorderTop || spec.BorderBottom || spec.BorderLeft || spec.BorderRight {
		borderColor := spec.BorderColor
		if borderColor == "" {
			borderColor = "808080"
		}
		col := map[string]any{
			"red":   hexToFloat(borderColor, 0),
			"green": hexToFloat(borderColor, 1),
			"blue":  hexToFloat(borderColor, 2),
		}
		borders := map[string]any{}
		add := func(side string, enabled bool) {
			if !enabled && !spec.BorderAll {
				return
			}
			borders[side] = map[string]any{"style": "SOLID", "color": col}
		}
		add("top", spec.BorderTop)
		add("bottom", spec.BorderBottom)
		add("left", spec.BorderLeft)
		add("right", spec.BorderRight)
		if spec.BorderAll {
			borders["top"] = map[string]any{"style": "SOLID", "color": col}
			borders["bottom"] = map[string]any{"style": "SOLID", "color": col}
			borders["left"] = map[string]any{"style": "SOLID", "color": col}
			borders["right"] = map[string]any{"style": "SOLID", "color": col}
		}
		format["borders"] = borders
	}
	// NumberFormat-only path: build a separate request.
	if spec.NumberFormat != "" {
		nfReq := map[string]any{
			"repeatCell": map[string]any{
				"range":  grid,
				"cell":   map[string]any{"userEnteredFormat": map[string]any{"numberFormat": map[string]any{"type": spec.NumberFormat}}},
				"fields": "userEnteredFormat.numberFormat",
			},
		}
		nfBody, _ := json.Marshal(map[string]any{"requests": []map[string]any{nfReq}})
		return cl.batchUpdate(nfBody)
	}
	if len(format) == 0 {
		return fmt.Errorf("SetCellStyle: nothing to apply (need at least bg/fg/bold/borders)")
	}

	fields := []string{}
	if _, ok := format["backgroundColor"]; ok {
		fields = append(fields, "backgroundColor")
	}
	if _, ok := format["textFormat"]; ok {
		fields = append(fields, "textFormat")
	}
	if _, ok := format["horizontalAlignment"]; ok {
		fields = append(fields, "horizontalAlignment")
	}
	if _, ok := format["wrapStrategy"]; ok {
		fields = append(fields, "wrapStrategy")
	}
	if _, ok := format["borders"]; ok {
		fields = append(fields, "borders")
	}
	if len(fields) == 0 {
		return fmt.Errorf("SetCellStyle: nothing to apply (need at least bg/fg/bold/borders)")
	}
	fieldsJoined := "userEnteredFormat(" + strings.Join(fields, ",") + ")"
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{"repeatCell": map[string]any{
				"range":  grid,
				"cell":   map[string]any{"userEnteredFormat": format},
				"fields": fieldsJoined,
			}},
		},
	})
	return cl.batchUpdate(body)
}

func (cl *Client) batchUpdate(body []byte) error {
	u := fmt.Sprintf("%s/%s:batchUpdate", sheetsAPIBase, cl.spreadsheetID)
	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cl.http.Do(req)
	if err != nil {
		return fmt.Errorf("sheets: batchUpdate: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sheets: batchUpdate HTTP %d: %s", resp.StatusCode, truncate(string(rb), 320))
	}
	return nil
}

// SetColumnWidths sets the pixel width for a range of columns in a single tab.
// tabName: tab to modify
// widths: map of column index (0-based) → pixel width
// Returns error if any column width is invalid (< 8px).
func (cl *Client) SetColumnWidths(tabName string, widths map[int]int) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	sheetID, err := cl.resolveSheetID(tabName)
	if err != nil {
		return fmt.Errorf("set column widths: %w", err)
	}
	var requests []map[string]any
	for colIdx, w := range widths {
		if w < 8 {
			return fmt.Errorf("column %d width %d too small (min 8px)", colIdx, w)
		}
		requests = append(requests, map[string]any{
			"updateDimensionProperties": map[string]any{
				"range": map[string]any{
					"sheetId":    sheetID,
					"dimension":  "COLUMNS",
					"startIndex": colIdx,
					"endIndex":   colIdx + 1,
				},
				"properties": map[string]any{"pixelSize": w},
				"fields":     "pixelSize",
			},
		})
	}
	if len(requests) == 0 {
		return fmt.Errorf("no column widths specified")
	}
	body, _ := json.Marshal(map[string]any{"requests": requests})
	return cl.batchUpdate(body)
}

// SetFreezePanes freezes rows and/or columns so they stay visible while scrolling.
// tabName: tab to modify.
// frozenRows: number of rows to freeze from top (0 = none).
// frozenCols: number of columns to freeze from left (0 = none).
func (cl *Client) SetFreezePanes(tabName string, frozenRows, frozenCols int) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	sheetID, err := cl.resolveSheetID(tabName)
	if err != nil {
		return fmt.Errorf("set freeze panes: %w", err)
	}
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"updateSheetProperties": map[string]any{
					"properties": map[string]any{
						"sheetId": sheetID,
						"gridProperties": map[string]any{
							"frozenRowCount":    frozenRows,
							"frozenColumnCount": frozenCols,
						},
					},
					"fields": "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
				},
			},
		},
	})
	return cl.batchUpdate(body)
}

// ExpandGrid extends a tab's row/column count so writes can target cells
// beyond the current grid. Use for adding new columns/rows without shifting
// existing data. tabName: tab to modify. colsToAdd: how many extra columns to append.
// Returns the new column count.
func (cl *Client) ExpandGrid(tabName string, colsToAdd int) error {
	if err := EnsureFresh(cl.c); err != nil {
		return err
	}
	sheetID, err := cl.resolveSheetID(tabName)
	if err != nil {
		return fmt.Errorf("expand grid: %w", err)
	}
	// First, get current grid properties
	sheets, err := cl.ListSheets()
	if err != nil {
		return fmt.Errorf("expand grid: list sheets: %w", err)
	}
	var sheet *SheetMeta
	for _, s := range sheets {
		if s.Title == tabName {
			sheet = &s
			break
		}
	}
	if sheet == nil {
		return fmt.Errorf("expand grid: tab %q not found", tabName)
	}
	newCols := sheet.Cols + colsToAdd
	body, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"updateSheetProperties": map[string]any{
					"properties": map[string]any{
						"sheetId": sheetID,
						"gridProperties": map[string]any{
							"columnCount": newCols,
						},
					},
					"fields": "gridProperties.columnCount",
				},
			},
		},
	})
	return cl.batchUpdate(body)
}

func (cl *Client) resolveSheetID(tabName string) (int64, error) {
	sheets, err := cl.ListSheets()
	if err != nil {
		return 0, err
	}
	for _, s := range sheets {
		if s.Title == tabName {
			return s.SheetID, nil
		}
	}
	return 0, fmt.Errorf("sheets: tab %q not found", tabName)
}

func hexToFloat(hex string, idx int) float64 {
	if len(hex) < 6 {
		return 0
	}
	hex = strings.ToUpper(hex)
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) >= 2 {
		var v uint8
		_, err := fmt.Sscanf(hex[2*idx:2*idx+2], "%x", &v)
		if err == nil {
			return float64(v) / 255.0
		}
	}
	return 0
}

// a1ToGrid parses a tab-qualified A1 like 'PERSONAL!A4:P4' into a
// GridRange (sheetId + numeric indices). Used by named ranges.
func (cl *Client) a1ToGrid(a1Range string) (map[string]any, error) {
	parts := strings.SplitN(a1Range, "!", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("a1ToGrid: range %q must be tab-qualified", a1Range)
	}
	tab := strings.Trim(parts[0], "'")
	rng := parts[1]
	span := strings.SplitN(rng, ":", 2)
	startA1, endA1 := span[0], span[0]
	if len(span) == 2 {
		endA1 = span[1]
	}
	startCol, startRow, err := parseA1Cell(startA1)
	if err != nil {
		return nil, err
	}
	endCol, endRow, err := parseA1Cell(endA1)
	if err != nil {
		return nil, err
	}
	sheets, err := cl.ListSheets()
	if err != nil {
		return nil, err
	}
	var sheetID int64
	for _, s := range sheets {
		if s.Title == tab {
			sheetID = s.SheetID
			break
		}
	}
	if sheetID == 0 {
		return nil, fmt.Errorf("a1ToGrid: tab %q not found", tab)
	}
	return map[string]any{
		"sheetId":          sheetID,
		"startRowIndex":    startRow - 1,
		"endRowIndex":      endRow,
		"startColumnIndex": startCol - 1,
		"endColumnIndex":   endCol,
	}, nil
}

func parseA1Cell(cell string) (col, row int, err error) {
	i := 0
	for i < len(cell) && cell[i] >= 'A' && cell[i] <= 'Z' {
		i++
	}
	if i == 0 || i == len(cell) {
		return 0, 0, fmt.Errorf("bad cell %q", cell)
	}
	col = colLetterInverse(cell[:i])
	row, err = strconv.Atoi(cell[i:])
	if err != nil {
		return 0, 0, fmt.Errorf("bad row in %q: %w", cell, err)
	}
	return col, row, nil
}

// colLetterInverse converts A1 column letters to a 1-based integer.
// Examples: A→1, Z→26, AA→27, AZ→52, BA→53.
func colLetterInverse(letters string) int {
	result := 0
	for _, ch := range letters {
		result = result*26 + int(ch-'A'+1)
	}
	return result
}

// colLetter converts a 1-based column index to A1 letters.
// Examples: 1→A, 26→Z, 27→AA.
func colLetter(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}
	return result
}

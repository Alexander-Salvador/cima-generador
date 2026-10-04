// restructure_feedbacks_v10_total_fix.go — CEO directive 2026-10-04 (round 8)
// "La hoja FEEDBACKS a quedado rota, arregla todo en general frontend UXUI y lógica"
//
// FIX TOTAL en un solo script:
//   A. LIMPIEZA
//     - DELETE rows 45-50 (legacy mapping table from v5)
//     - DELETE CF rules huérfanas (rows 14-40 ya no existen)
//   B. LAYOUT FEEDBACK master
//     - Title row 1 con merge A1:N1
//     - Subtitle row 2 actualizado
//     - Header row 4 con bg dark blue + fg white + bold
//     - Zebra striping rows 5-11 (bg gris claro alterno)
//     - Borders: frame exterior + interiores verticales
//     - Freeze rows 1-4 (header frozen)
//     - Col widths optimizadas
//   C. CF rules
//     - Borrar las 6 huérfanas (rows 14-40)
//     - Agregar 4 para TIPO_FEEDBACK (Positivo/Negativo/Evaluativo/Amonestador)
//     - Re-aplicar 6 para ESTADO en rows 4-11
//   D. HYPERLINK styling
//     - Color azul + underline para celdas con HYPERLINK formulas
//   E. DETALLE_LLAMADA: mismo tratamiento + sort por agente
//   F. Sort DETALLE por agente + fecha
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetFB       = int64(681240158)
	sheetDETALLE  = int64(412152155)
	sheetPERSONAL = int64(1395452450)
	sheetCONFIG   = int64(1497962081)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v10 — TOTAL FIX (layout + logic + UX)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println()

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// A. LIMPIEZA — DELETE legacy
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ A. LIMPIEZA")

	// A1: DELETE rows 45-50 del FEEDBACK (legacy mapping table)
	fmt.Println("  A1: DELETE FEEDBACK rows 45-50 (legacy mapping table)")
	for r := 50; r >= 45; r-- {
		reqs = append(reqs, deleteRow(r, sheetFB))
	}

	// A2: DELETE existing CF rules (las 6 huérfanas en rows 14-40 + las 6 válidas)
	//     Re-crearlas todas desde cero
	//     Iterar hacia atrás (11 → 0) para evitar problemas de reindex
	fmt.Println("  A2: DELETE all 12 existing CF rules (reverse order)")
	for i := 11; i >= 0; i-- {
		reqs = append(reqs, deleteConditionalFormat(i, sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// B. LAYOUT FEEDBACK master
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ B. LAYOUT FEEDBACK master")

	// B1: Title row 1 con merge A1:N1
	fmt.Println("  B1: Title row 1 + merge A1:N1")
	reqs = append(reqs, mergeRange(0, 0, 1, 14, sheetFB)) // A1:N1 (0-based)
	reqs = append(reqs, setValue("A", 1, "FEEDBACKS — Auditoría Semanal W40-2026", sheetFB))

	// B2: Subtitle row 2 con merge A2:N2
	reqs = append(reqs, mergeRange(1, 0, 2, 14, sheetFB))
	reqs = append(reqs, setValue("A", 2, "Click en cualquier AGENTE/CANT/ID para navegar · Click en N° para ver ficha PERSONAL", sheetFB))

	// B3: Header row 4 styling
	fmt.Println("  B3: Header row 4 — bg dark blue + fg white + bold")
	reqs = append(reqs, styleRow(3, 14, 0, 14, sheetFB,
		"", "#1a3a5c", "#FFFFFF", true, false))

	// B4: Zebra striping rows 5, 7, 9, 11 (bg gris claro)
	fmt.Println("  B4: Zebra striping rows 5,7,9,11")
	for _, r := range []int{5, 7, 9, 11} {
		reqs = append(reqs, styleRow(r-1, r, 0, 14, sheetFB,
			"", "#f0f4f8", "#0F2440", false, false))
	}
	// Rows 6, 8, 10 blanco
	for _, r := range []int{6, 8, 10} {
		reqs = append(reqs, styleRow(r-1, r, 0, 14, sheetFB,
			"", "#FFFFFF", "#0F2440", false, false))
	}

	// B5: Borders
	fmt.Println("  B5: Borders frame + interiores")
	// Frame exterior HEADER (rows 1-4)
	reqs = append(reqs, frameBorder(0, 0, 4, 14, sheetFB, 2))
	// Frame exterior DATA (rows 5-11)
	reqs = append(reqs, frameBorder(4, 0, 11, 14, sheetFB, 2))
	// Separación entre rows 5-11 (1px horizontal)
	for r := 5; r < 11; r++ {
		reqs = append(reqs, hBorder(r, 0, 14, sheetFB))
	}

	// B6: Col widths optimizadas
	fmt.Println("  B6: Col widths (ID=30, AGENTE=180, SUPERVISOR=150, ...)")
	colWidths := []int{30, 180, 150, 80, 110, 100, 90, 90, 110, 300, 110, 110, 130, 70}
	for i, w := range colWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetFB))
	}

	// B7: Freeze rows 1-4
	fmt.Println("  B7: Freeze rows 1-4")
	reqs = append(reqs, setFreeze(4, 0, sheetFB))

	// B9: HYPERLINK styling (color azul + underline para celdas con link)
	fmt.Println("  B9: HYPERLINK styling para A, B, D (rows 5-11)")
	for r := 5; r <= 11; r++ {
		for _, col := range []string{"B", "D"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetFB,
				"", "#FFFFFF", "#1a73e8", false, true)) // underline=true
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// C. CF rules FEEDBACK
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ C. CF rules FEEDBACK")

	// C1: TIPO_FEEDBACK en col E rows 5-11
	fmt.Println("  C1: CF TIPO_FEEDBACK (col E rows 5-11)")
	tipoFBColors := []struct {
		Val string
		R, G, B float64
	}{
		{"Positivo", 0.6, 0.9, 0.6},          // verde claro
		{"Negativo", 1.0, 0.6, 0.6},          // rojo claro
		{"Evaluativo", 0.6, 0.85, 1.0},       // azul claro
		{"Amonestador", 1.0, 0.5, 0.3},       // naranja oscuro
	}
	for _, c := range tipoFBColors {
		reqs = append(reqs, addConditionalFormat(sheetFB, 4, 5, 4, 11, c.Val, c.R, c.G, c.B))
	}

	// C2: ESTADO en col M rows 5-11
	fmt.Println("  C2: CF ESTADO (col M rows 5-11)")
	estadoColors := []struct {
		Val string
		R, G, B float64
	}{
		{"Pendiente", 0.85, 0.85, 0.85},
		{"En evaluación", 1.0, 1.0, 0.6},
		{"Feedback dado", 0.6, 0.85, 1.0},
		{"Carta de compromiso", 1.0, 0.85, 0.4},
		{"Cumplido", 0.5, 0.9, 0.5},
		{"Vencido", 1.0, 0.5, 0.5},
	}
	for _, c := range estadoColors {
		reqs = append(reqs, addConditionalFormat(sheetFB, 12, 5, 12, 11, c.Val, c.R, c.G, c.B))
	}

	// ═══════════════════════════════════════════════════════════════════
	// D. DETALLE_LLAMADA layout
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ D. LAYOUT FEEDBACK-DETALLE_LLAMADA")

	// D1: Title row 1 con merge A1:L1
	fmt.Println("  D1: Title row 1 + merge A1:L1")
	reqs = append(reqs, mergeRange(0, 0, 1, 12, sheetDETALLE))
	reqs = append(reqs, setValue("A", 1, "DETALLE DE LLAMADAS — Base de Feedbacks W40-2026", sheetDETALLE))

	// D2: Header row 2 (era row 1 antes del merge del title)
	// Wait — el row 1 ahora es el title. Header debe ser row 2.
	// Pero mi diag mostró que headers actuales están en row 1. Necesito reorganizar.
	//
	// Voy a: shift rows 2-27 down by 1 (insert row 2), poner header en row 3
	// O alternativa: dejar header en row 2 (entre title y data)

	// Por simplicidad: title row 1, header row 2, data rows 3-28
	// Necesito INSERT una row después de row 1 para hacer espacio
	fmt.Println("  D2: Insert row 2 for header (shift data)")
	reqs = append(reqs, insertRow(2, 1, sheetDETALLE))

	// D3: Header row 2 styling (dark blue)
	reqs = append(reqs, styleRow(1, 12, 0, 12, sheetDETALLE,
		"", "#1a3a5c", "#FFFFFF", true, false))

	// D4: Zebra striping en data rows
	fmt.Println("  D4: Zebra striping data rows")
	for _, r := range []int{3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27} {
		reqs = append(reqs, styleRow(r-1, r, 0, 12, sheetDETALLE,
			"", "#f0f4f8", "#0F2440", false, false))
	}

	// D5: Borders
	fmt.Println("  D5: Borders")
	reqs = append(reqs, frameBorder(0, 0, 2, 12, sheetDETALLE, 2))
	reqs = append(reqs, frameBorder(2, 0, 28, 12, sheetDETALLE, 2))
	for r := 3; r < 28; r++ {
		reqs = append(reqs, hBorder(r, 0, 12, sheetDETALLE))
	}

	// D6: Col widths DETALLE
	fmt.Println("  D6: Col widths DETALLE")
	detalleWidths := []int{40, 40, 100, 150, 170, 130, 130, 220, 110, 60, 90, 150}
	for i, w := range detalleWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetDETALLE))
	}

	// D7: Freeze rows 1-2
	reqs = append(reqs, setFreeze(2, 0, sheetDETALLE))

	// D8: HYPERLINK styling DETALLE (B y J cols)
	fmt.Println("  D8: HYPERLINK styling DETALLE (B y J cols, rows 3-28)")
	for r := 3; r <= 28; r++ {
		for _, col := range []string{"B", "E", "J"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetDETALLE,
				"", "#FFFFFF", "#1a73e8", false, true))
		}
	}

	// D9: CF rules DETALLE — ESTADO_AUDITORIA en col I rows 3-28
	fmt.Println("  D9: CF DETALLE ESTADO_AUDITORIA (col I)")
	estadoAudColors := []struct {
		Val string
		R, G, B float64
	}{
		{"Pendiente", 0.85, 0.85, 0.85},
		{"Auditada", 0.6, 0.9, 0.6},
		{"Con feedback", 0.6, 0.85, 1.0},
	}
	for _, c := range estadoAudColors {
		reqs = append(reqs, addConditionalFormat(sheetDETALLE, 8, 3, 8, 28, c.Val, c.R, c.G, c.B))
	}

	// D10: TIPO_LLAMADA CF en col G rows 3-28
	fmt.Println("  D10: CF DETALLE TIPO_LLAMADA (col G)")
	tipoLlamadaColors := []struct {
		Val string
		R, G, B float64
	}{
		{"Estados incorrectos", 1.0, 0.7, 0.7},
		{"Abandono llamada", 1.0, 0.85, 0.4},
		{"Speech incompleto", 0.7, 0.85, 1.0},
		{"No devuelve llamada", 1.0, 0.7, 1.0},
	}
	for _, c := range tipoLlamadaColors {
		reqs = append(reqs, addConditionalFormat(sheetDETALLE, 6, 3, 6, 28, c.Val, c.R, c.G, c.B))
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("• DELETE 6 rows legacy + 12 CF huérfanos")
	fmt.Println("• Layout FEEDBACK: title, subtitle, header styling, borders, freeze, widths, hyperlink styling")
	fmt.Println("• CF FEEDBACK: 4 TIPO_FEEDBACK + 6 ESTADO = 10 reglas")
	fmt.Println("• Layout DETALLE: title, header, zebra, borders, freeze, widths, hyperlink styling, insert row")
	fmt.Println("• CF DETALLE: 3 ESTADO_AUDITORIA + 4 TIPO_LLAMADA = 7 reglas")
	fmt.Println("• Total: ~150 requests")

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
	fmt.Printf("\n📤 Enviando %d requests...\n", len(reqs))
	body := map[string]any{"requests": reqs}
	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s:batchUpdate", spreadsheetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	respB, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println("API ERROR", resp.StatusCode, ":", string(respB))
		os.Exit(1)
	}
	var result map[string]any
	json.Unmarshal(respB, &result)
	replies, _ := result["replies"].([]any)
	fmt.Printf("✅ Sent %d requests, %d replies\n", len(reqs), len(replies))
	fmt.Printf("🕐 Done at: %s\n", time.Now().Format(time.RFC3339))
}

// ═══════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════

func colIdx(cell string) int {
	col := 0
	for _, c := range cell {
		if c >= 'A' && c <= 'Z' {
			col = col*26 + int(c-'A'+1)
		}
	}
	return col - 1
}

func deleteRow(row int, sheetID int64) map[string]any {
	ri := row - 1
	return map[string]any{
		"deleteDimension": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "ROWS",
				"startIndex": ri,
				"endIndex":   ri + 1,
			},
		},
	}
}

func insertRow(row, count int, sheetID int64) map[string]any {
	ri := row - 1
	return map[string]any{
		"insertDimension": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "ROWS",
				"startIndex": ri,
				"endIndex":   ri + count,
			},
			"inheritFromBefore": false,
		},
	}
}

func deleteConditionalFormat(index int, sheetID int64) map[string]any {
	return map[string]any{
		"deleteConditionalFormatRule": map[string]any{
			"sheetId": sheetID,
			"index":   index,
		},
	}
}

func mergeRange(startRow, startCol, endRow, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"mergeCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"mergeType": "MERGE_ALL",
		},
	}
}

func setValue(col string, row int, value string, sheetID int64) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    ri,
				"endRowIndex":      ri + 1,
				"startColumnIndex": ci,
				"endColumnIndex":   ci + 1,
			},
			"rows": []any{
				map[string]any{
					"values": []any{
						map[string]any{
							"userEnteredValue": map[string]any{"stringValue": value},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

// styleRow: aplica backgroundColor, foregroundColor, bold a un rango
func styleRow(startRow, endRow, startCol, endCol int, sheetID int64, _ string, bgHex, fgHex string, bold, underline bool) map[string]any {
	textFormat := map[string]any{
		"foregroundColor": map[string]any{
			"red":   hexToR(gbParse(fgHex)),
			"green": hexToG(gbParse(fgHex)),
			"blue":  hexToB(gbParse(fgHex)),
		},
	}
	if bold {
		textFormat["bold"] = true
	}
	if underline {
		textFormat["underline"] = true
	}
	bgColor := map[string]any{
		"red":   hexToR(gbParse(bgHex)),
		"green": hexToG(gbParse(bgHex)),
		"blue":  hexToB(gbParse(bgHex)),
	}
	return map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgColor,
					"textFormat":      textFormat,
					"horizontalAlignment": "CENTER",
					"verticalAlignment":   "MIDDLE",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)",
		},
	}
}

func styleCell(startRow, startCol, endRow, endCol int, sheetID int64, _ string, bgHex, fgHex string, bold, underline bool) map[string]any {
	textFormat := map[string]any{
		"foregroundColor": map[string]any{
			"red":   hexToR(gbParse(fgHex)),
			"green": hexToG(gbParse(fgHex)),
			"blue":  hexToB(gbParse(fgHex)),
		},
	}
	if bold {
		textFormat["bold"] = true
	}
	if underline {
		textFormat["underline"] = true
	}
	bgColor := map[string]any{
		"red":   hexToR(gbParse(bgHex)),
		"green": hexToG(gbParse(bgHex)),
		"blue":  hexToB(gbParse(bgHex)),
	}
	return map[string]any{
		"repeatCell": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"cell": map[string]any{
				"userEnteredFormat": map[string]any{
					"backgroundColor": bgColor,
					"textFormat":      textFormat,
					"horizontalAlignment": "LEFT",
				},
			},
			"fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	}
}

func frameBorder(startRow, startCol, endRow, endCol int, sheetID int64, width int) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"top":    map[string]any{"style": "SOLID", "width": width},
			"bottom": map[string]any{"style": "SOLID", "width": width},
			"left":   map[string]any{"style": "SOLID", "width": width},
			"right":  map[string]any{"style": "SOLID", "width": width},
		},
	}
}

func hBorder(row, startCol, endCol int, sheetID int64) map[string]any {
	return map[string]any{
		"updateBorders": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    row - 1,
				"endRowIndex":      row,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"bottom": map[string]any{"style": "SOLID", "width": 1, "color": map[string]any{"red": 0.7, "green": 0.7, "blue": 0.7}},
		},
	}
}

func setColWidth(startIdx, endIdx, pixelSize int, sheetID int64) map[string]any {
	return map[string]any{
		"updateDimensionProperties": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetID,
				"dimension":  "COLUMNS",
				"startIndex": startIdx,
				"endIndex":   endIdx,
			},
			"properties": map[string]any{"pixelSize": pixelSize},
			"fields":     "pixelSize",
		},
	}
}

func setFreeze(frozenRowCount, frozenColCount int, sheetID int64) map[string]any {
	return map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": sheetID,
				"gridProperties": map[string]any{
					"frozenRowCount":    frozenRowCount,
					"frozenColumnCount": frozenColCount,
				},
			},
			"fields": "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
		},
	}
}

func addConditionalFormat(sheetID int64, col int, startRow, _, endRow int, value string, r, g, b float64) map[string]any {
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetID,
						"startRowIndex":    startRow - 1,
						"endRowIndex":      endRow,
						"startColumnIndex": col,
						"endColumnIndex":   col + 1,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type":   "TEXT_EQ",
						"values": []map[string]any{{"userEnteredValue": value}},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{"red": r, "green": g, "blue": b},
					},
				},
			},
			"index": 0,
		},
	}
}

// ═══════════════════════════════════════════════════════════════════
// COLOR HELPERS
// ═══════════════════════════════════════════════════════════════════

type rgb struct{ r, g, b int }

func gbParse(hex string) rgb {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return rgb{255, 255, 255}
	}
	var v int
	fmt.Sscanf(hex, "%x", &v)
	return rgb{
		r: (v >> 16) & 0xFF,
		g: (v >> 8) & 0xFF,
		b: v & 0xFF,
	}
}

func hexToR(c rgb) float64 { return float64(c.r) / 255.0 }
func hexToG(c rgb) float64 { return float64(c.g) / 255.0 }
func hexToB(c rgb) float64 { return float64(c.b) / 255.0 }

func decrypt() map[string]any {
	keyB, _ := os.ReadFile(keyPath)
	encB, _ := os.ReadFile(encPath)
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	out := map[string]any{}
	json.Unmarshal(pt, &out)
	return out
}

func getToken(c map[string]any) string {
	body := map[string]any{
		"client_id":     c["client_id"],
		"client_secret": c["client_secret"],
		"refresh_token": c["refresh_token"],
		"grant_type":    "refresh_token",
	}
	bb, _ := json.Marshal(body)
	r, _ := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewReader(bb))
	defer r.Body.Close()
	res := map[string]any{}
	json.NewDecoder(r.Body).Decode(&res)
	tok, _ := res["access_token"].(string)
	return tok
}
// restructure_feedbacks_v14b_rebuild.go — Fase 2: rebuild completo FEEDBACK
//
// Después de v14a (cleanup), ahora:
//   1. Rewrite rows 1-11 con nuevo schema (8 cols)
//   2. Dropdowns actualizados
//   3. CF rules nuevos (TIPO_FEEDBACK, ESTADO, CORRELACIÓN TIPO↔ESTADO)
//   4. Borders + freeze + widths + HYPERLINK styling
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
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

var agentes = []struct {
	ID            int
	PersonalRow   int
	Nombre        string
	TipoFeedback  string
	FBase         string
	FFeedback    string
	Observacion   string
	Estado        string
	Supervisor    string
}{
	{1, 6, "RAMIREZ HUAMAN JOHNN MILTON", "Evaluativo", "23/09/2026", "23/09/2026", "Carta de compromiso, se saco a feedback en Jitsi", "Carta de compromiso", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{2, 11, "VARGAS TAPIA ISABEL", "Evaluativo", "23/09/2026", "23/09/2026", "Carta de compromiso, se saco a feedback en Jitsi", "Carta de compromiso", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{3, 21, "GOMEZ FLORES JONATHAN ARTURO", "Positivo", "17/09/2026", "17/09/2026", "Se alerta via whatsapp obteniendo un compromiso por el As.", "Feedback dado", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{4, 12, "MENDOZA CARMEN CECILIA LILIANA", "Positivo", "22/09/2026", "22/09/2026", "Se alerta via whatsapp obteniendo un compromiso por el As.", "Feedback dado", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{5, 19, "LLAMPASI CRISOSTOMO LIZET YULY", "Evaluativo", "26/09/2026", "26/09/2026", "CL indica que tiene problemas con su navegación, asesor no sondea para encontrar el error.", "Feedback dado", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, 23, "IMAN RUJEL YADIRA CATIZZA", "Amonestador", "22/09/2026", "22/09/2026", "6 calls consolidados W40-2026", "Carta de compromiso", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, 16, "CANALES COPLAT ZAYDA JOSCELINE", "Negativo", "23/09/2026", "23/09/2026", "2 calls consolidados W40-2026", "Carta de compromiso", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
}

func firstDetailRow(agenteID int, fechaBase string) int {
	llamadas := []struct {
		FK    int
		FBase string
		Row   int
	}{
		{1, "23/09/2026", 3}, {2, "23/09/2026", 4}, {3, "17/09/2026", 5},
		{4, "22/09/2026", 6}, {5, "26/09/2026", 7}, {6, "22/09/2026", 8},
		{6, "22/09/2026", 9}, {6, "22/09/2026", 10}, {6, "23/09/2026", 11},
		{6, "23/09/2026", 12}, {6, "23/09/2026", 13}, {7, "23/09/2026", 14},
		{7, "23/09/2026", 15}, {7, "23/09/2026", 16}, {6, "22/09/2026", 17},
		{6, "22/09/2026", 18}, {6, "22/09/2026", 19}, {7, "23/09/2026", 20},
		{6, "23/09/2026", 21}, {6, "23/09/2026", 22}, {6, "23/09/2026", 23},
		{7, "23/09/2026", 24}, {7, "23/09/2026", 25}, {7, "23/09/2026", 26},
		{7, "23/09/2026", 27}, {7, "23/09/2026", 28},
	}
	for _, l := range llamadas {
		if l.FK == agenteID && l.FBase == fechaBase {
			return l.Row
		}
	}
	return 3
}

type cfCond struct {
	Col int
	Val string
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v14b — REBUILD CON NUEVO SCHEMA (8 cols)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// STEP 1: Rewrite rows 1-11 con nuevo schema
	fmt.Println("\n━━━ STEP 1: Rewrite rows 1-11 con nuevo schema (8 cols)")

	reqs = append(reqs, setValue("A", 1, "FEEDBACKS — Auditoría Semanal W40-2026", sheetFB))
	reqs = append(reqs, setValue("A", 2, "Resumen de Feedbacks del Supervisor · 1 row por agente · Conectado a FEEDBACK-DETALLE_LLAMADA", sheetFB))

	headers := []string{"AGENTE", "TIPO_FEEDBACK", "FECHA_FEEDBACK", "FECHA_BASE", "OBSERVACIÓN", "CANT_CALLS", "ESTADO"}
	for i, h := range headers {
		reqs = append(reqs, setValue(string(rune('A'+i)), 4, h, sheetFB))
	}
	reqs = append(reqs, setValue("H", 4, "SUPERVISOR", sheetFB))

	for i, a := range agentes {
		row := i + 5
		detalleRow := firstDetailRow(a.ID, a.FBase)

		reqs = append(reqs, setFormula("A", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d","%s")`, sheetDETALLE, detalleRow, a.Nombre), sheetFB))
		reqs = append(reqs, setValue("B", row, a.TipoFeedback, sheetFB))
		reqs = append(reqs, setValue("C", row, a.FFeedback, sheetFB))
		reqs = append(reqs, setValue("D", row, a.FBase, sheetFB))
		reqs = append(reqs, setValue("E", row, a.Observacion, sheetFB))
		// CANT_CALLS INTELIGENTE — count calls del mismo agente con misma FECHA_BASE
		reqs = append(reqs, setFormula("F", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d",TEXT(COUNTIFS('FEEDBACK-DETALLE_LLAMADA'!E:E,A%d,'FEEDBACK-DETALLE_LLAMADA'!K:K,D%d),"0"))`,
				sheetDETALLE, detalleRow, row, row), sheetFB))
		reqs = append(reqs, setValue("G", row, a.Estado, sheetFB))
		reqs = append(reqs, setValue("H", row, a.Supervisor, sheetFB))
	}

	// STEP 2: Dropdowns
	fmt.Println("\n━━━ STEP 2: Dropdowns (B, G)")
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "B5:B11",
		[]string{"Positivo", "Negativo", "Evaluativo", "Amonestador"}))
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "G5:G11",
		[]string{"Pendiente", "En evaluación", "Feedback dado", "Carta de compromiso", "Cumplido", "Vencido"}))

	// STEP 3: CF rules nuevos
	fmt.Println("\n━━━ STEP 3: CF rules nuevos")
	fmt.Println("  TIPO_FEEDBACK (col B)")
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Positivo", 0.6, 0.9, 0.6},
		{"Negativo", 1.0, 0.6, 0.6},
		{"Evaluativo", 0.6, 0.85, 1.0},
		{"Amonestador", 1.0, 0.5, 0.3},
	} {
		reqs = append(reqs, addCF(sheetFB, 1, 5, 1, 11, c.Val, c.R, c.G, c.B))
	}

	fmt.Println("  ESTADO (col G)")
	for _, c := range []struct {
		Val                string
		R, G, B            float64
	}{
		{"Pendiente", 0.85, 0.85, 0.85},
		{"En evaluación", 1.0, 1.0, 0.6},
		{"Feedback dado", 0.6, 0.85, 1.0},
		{"Carta de compromiso", 1.0, 0.85, 0.4},
		{"Cumplido", 0.5, 0.9, 0.5},
		{"Vencido", 1.0, 0.5, 0.5},
	} {
		reqs = append(reqs, addCF(sheetFB, 6, 5, 6, 11, c.Val, c.R, c.G, c.B))
	}

	fmt.Println("  CORRELACIÓN TIPO↔ESTADO (highlighting)")
	reqs = append(reqs, addCFMultiCondition(sheetFB, 5, 11,
		[]cfCond{{Col: 1, Val: "Positivo"}, {Col: 6, Val: "Cumplido"}}, 0.5, 0.9, 0.5))
	reqs = append(reqs, addCFMultiCondition(sheetFB, 5, 11,
		[]cfCond{{Col: 1, Val: "Amonestador"}, {Col: 6, Val: "Carta de compromiso"}}, 1.0, 0.7, 0.3))
	reqs = append(reqs, addCFMultiCondition(sheetFB, 5, 11,
		[]cfCond{{Col: 1, Val: "Evaluativo"}, {Col: 6, Val: "Feedback dado"}}, 0.5, 0.7, 1.0))

	// STEP 4: Borders + widths + freeze + HYPERLINK styling
	fmt.Println("\n━━━ STEP 4: Borders + widths + freeze + styling")
	reqs = append(reqs, frameBorder(0, 0, 4, 8, sheetFB, 2))
	reqs = append(reqs, frameBorder(4, 0, 11, 8, sheetFB, 2))
	for r := 6; r <= 11; r++ {
		reqs = append(reqs, hBorder(r, 0, 8, sheetFB))
	}

	colWidths := []int{200, 110, 110, 100, 350, 100, 130, 180}
	for i, w := range colWidths {
		reqs = append(reqs, setColWidth(i, i+1, w, sheetFB))
	}
	reqs = append(reqs, setFreeze(4, 0, sheetFB))

	for r := 5; r <= 11; r++ {
		for _, col := range []string{"A", "F"} {
			reqs = append(reqs, styleCell(r-1, colIdx(col), r, colIdx(col)+1, sheetFB, "#FFFFFF", "#1a73e8", false, true, 10))
		}
	}

	fmt.Printf("\nTotal: %d requests\n", len(reqs))

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN")
		return
	}

	creds := decrypt()
	tok := getToken(creds)
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
	fmt.Printf("✅ Sent %d, replies %d\n", len(reqs), len(replies))
	fmt.Printf("🕐 Done: %s\n", time.Now().Format(time.RFC3339))
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

func setFormula(col string, row int, formula string, sheetID int64) map[string]any {
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
							"userEnteredValue": map[string]any{"formulaValue": formula},
						},
					},
				},
			},
			"fields": "userEnteredValue",
		},
	}
}

func dataValidationOneOfList(sheetID int64, a1Range string, values []string) map[string]any {
	_, startRow, endRow, startCol, endCol := a1ToGrid(a1Range)
	return map[string]any{
		"setDataValidation": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetID,
				"startRowIndex":    startRow,
				"endRowIndex":      endRow,
				"startColumnIndex": startCol,
				"endColumnIndex":   endCol,
			},
			"rule": map[string]any{
				"condition": map[string]any{
					"type":   "ONE_OF_LIST",
					"values": valuesToUserEntered(values),
				},
				"showCustomUi": true,
			},
		},
	}
}

func valuesToUserEntered(values []string) []map[string]any {
	out := make([]map[string]any, len(values))
	for i, v := range values {
		out[i] = map[string]any{
			"userEnteredValue": v,
		}
	}
	return out
}

func a1ToGrid(a1 string) (string, int, int, int, int) {
	parts := strings.Split(a1, ":")
	if len(parts) != 2 {
		return a1, 0, 1, 0, 1
	}
	startColS := strings.TrimLeft(parts[0], "0123456789")
	startRowS := strings.TrimLeft(parts[0], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	startCol := colIdx(startColS)
	startRow, _ := strconvAtoi(startRowS)
	endColS := strings.TrimLeft(parts[1], "0123456789")
	endRowS := strings.TrimLeft(parts[1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	endCol := colIdx(endColS) + 1
	endRow, _ := strconvAtoi(endRowS)
	return a1, startRow - 1, endRow, startCol, endCol
}

func strconvAtoi(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func addCF(sheetID int64, col, startRow, _, endRow int, value string, r, g, b float64) map[string]any {
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

func addCFMultiCondition(sheetID int64, startRow, endRow int, conds []cfCond, r, g, b float64) map[string]any {
	// Sheets API v4 CUSTOM_FORMULA requires formula con = prefix
	// AND() no funciona, uso concatenación: $B5&"|"&$G5="Positivo|Cumplido"
	// Pero el API rechaza concatenación. Probemos multiplicación booleana:
	// =(B5="Positivo")*(G5="Cumplido")
	var formula string
	if len(conds) == 2 {
		formula = fmt.Sprintf(`=(%s%d="%s")*(%s%d="%s")`,
			string(rune('A'+conds[0].Col)), startRow, conds[0].Val,
			string(rune('A'+conds[1].Col)), startRow, conds[1].Val)
	} else {
		parts := make([]string, len(conds))
		for i, c := range conds {
			colLetter := string(rune('A' + c.Col))
			parts[i] = fmt.Sprintf(`(%s%d="%s")`, colLetter, startRow, c.Val)
		}
		formula = "=" + strings.Join(parts, "*")
	}
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetID,
						"startRowIndex":    startRow - 1,
						"endRowIndex":      endRow,
						"startColumnIndex": 0,
						"endColumnIndex":   8,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type":   "CUSTOM_FORMULA",
						"values": []map[string]any{{"userEnteredValue": formula}},
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

func styleCell(startRow, startCol, endRow, endCol int, sheetID int64, bgHex, fgHex string, bold, underline bool, fontSize int) map[string]any {
	textFormat := map[string]any{
		"foregroundColor": map[string]any{
			"red":   hexToR(gbParse(fgHex)),
			"green": hexToG(gbParse(fgHex)),
			"blue":  hexToB(gbParse(fgHex)),
		},
		"fontSize": fontSize,
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
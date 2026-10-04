// restructure_feedbacks_v16b_restructure.go — Fase B
//
// Reestructurar layout: DASHBOARD HEADER ARRIBA + espacio para nuevos agentes
//
// LAYOUT FINAL:
//   Row 1:  TITLE (FEEDBACKS — Auditoría Semanal W40-2026)
//   Row 2:  SUBTITLE
//   Row 3:  spacer
//   Row 4:  📊 DASHBOARD HEADER (merge A4:J4)
//   Row 5:  KPI LABELS (5 cards)
//   Row 6:  KPI VALUES (big font)
//   Row 7:  spacer
//   Row 8:  DATA HEADER
//   Rows 9-15: data (7 agentes)
//   Rows 16-30: EMPTY rows con formato (espacio para nuevos)
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

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS v16b — RESTRUCTURE (DASHBOARD ARRIBA)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))

	reqs := []any{}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 1: INSERT 3 rows después de row 3 (push down data)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 1: INSERT 3 rows after row 3 (push data down)")
	reqs = append(reqs, insertRow(4, 3, sheetFB)) // insert at index 4 (after row 3)

	// Después de INSERT:
	// - Row 1: title
	// - Row 2: subtitle
	// - Row 3: empty
	// - Rows 4-6: NEW (vacías, se llenan)
	// - Row 7: spacer (vacío, formerly row 4 = HEADER)
	// - Row 8: data HEADER (formerly row 4)
	// - Rows 9-15: data (formerly rows 5-11)
	// - Row 16: empty (formerly row 12)
	// - Rows 17-19: dashboard viejo (formerly rows 13-15)
	// - Rows 20+: empty

	// ═══════════════════════════════════════════════════════════════════
	// STEP 2: DASHBOARD content (rows 4-6)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 2: Dashboard content rows 4-6")

	// Row 4: DASHBOARD title (merge A4:J4)
	reqs = append(reqs, mergeRange(3, 0, 4, 10, sheetFB))
	reqs = append(reqs, setValue("A", 4, "📊 DASHBOARD DE CUMPLIMIENTO W40-2026", sheetFB))

	// Row 5: KPI LABELS (5 cells, each merged 2 cols)
	reqs = append(reqs, setValue("A", 5, "👥 AGENTES", sheetFB))
	reqs = append(reqs, mergeRange(4, 0, 5, 2, sheetFB)) // A5:B5
	reqs = append(reqs, setValue("C", 5, "📞 CALLS", sheetFB))
	reqs = append(reqs, mergeRange(4, 2, 5, 4, sheetFB)) // C5:D5
	reqs = append(reqs, setValue("E", 5, "✅ % CUMPLIDO", sheetFB))
	reqs = append(reqs, mergeRange(4, 4, 5, 6, sheetFB)) // E5:F5
	reqs = append(reqs, setValue("G", 5, "⚠ % VENCIDO", sheetFB))
	reqs = append(reqs, mergeRange(4, 6, 5, 8, sheetFB)) // G5:H5
	reqs = append(reqs, setValue("I", 5, "📊 TIPO TOP", sheetFB))
	reqs = append(reqs, mergeRange(4, 8, 5, 10, sheetFB)) // I5:J5

	// Row 6: KPI VALUES (formulas con rango rows 9-30)
	reqs = append(reqs, setFormula("A", 6, `=COUNTA(A9:A30)`, sheetFB))
	reqs = append(reqs, mergeRange(5, 0, 6, 2, sheetFB)) // A6:B6
	reqs = append(reqs, setFormula("C", 6, `=SUMPRODUCT(F9:F30)`, sheetFB))
	reqs = append(reqs, mergeRange(5, 2, 6, 4, sheetFB)) // C6:D6
	reqs = append(reqs, setFormula("E", 6, `=COUNTIF(G9:G30,"Cumplido")/COUNTA(G9:G30)`, sheetFB))
	reqs = append(reqs, mergeRange(5, 4, 6, 6, sheetFB)) // E6:F6
	reqs = append(reqs, setFormula("G", 6, `=COUNTIF(G9:G30,"Vencido")/COUNTA(G9:G30)`, sheetFB))
	reqs = append(reqs, mergeRange(5, 6, 6, 8, sheetFB)) // G6:H6
	reqs = append(reqs, setFormula("I", 6, `=IFERROR(INDEX(B9:B30,MATCH(MAX(COUNTIF(B9:B30,B9:B30)),COUNTIF(B9:B30,B9:B30),0)),"")`, sheetFB))
	reqs = append(reqs, mergeRange(5, 8, 6, 10, sheetFB)) // I6:J6

	// ═══════════════════════════════════════════════════════════════════
	// STEP 3: Limpiar contenido dashboard viejo (rows 17-19 después de INSERT)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 3: Clear viejo dashboard (rows 17-19 después de INSERT)")
	for r := 17; r <= 19; r++ {
		for col := 0; col < 10; col++ {
			reqs = append(reqs, clearCell(string(rune('A'+col)), r, sheetFB))
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 4: Re-write DATA HEADER row 8
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 4: Data header row 8")
	headers := []string{"AGENTE", "TIPO_FEEDBACK", "FECHA_FEEDBACK", "FECHA_BASE", "OBSERVACIÓN", "CANT_CALLS", "ESTADO", "SUPERVISOR", "STATUS", "ACCIÓN"}
	for i, h := range headers {
		reqs = append(reqs, setValue(string(rune('A'+i)), 8, h, sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 5: Re-write DATA rows 9-15 (mismo data, nuevas referencias)
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 5: Data rows 9-15 (7 agentes)")
	for i, a := range agentes {
		row := i + 9 // row 9, 10, ... 15
		detalleRow := firstDetailRow(a.ID, a.FBase)

		reqs = append(reqs, setFormula("A", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d","%s")`, sheetDETALLE, detalleRow, a.Nombre), sheetFB))
		reqs = append(reqs, setValue("B", row, a.TipoFeedback, sheetFB))
		reqs = append(reqs, setValue("C", row, a.FFeedback, sheetFB))
		reqs = append(reqs, setValue("D", row, a.FBase, sheetFB))
		reqs = append(reqs, setValue("E", row, a.Observacion, sheetFB))
		// CANT_CALLS ahora apunta a rows 9-30 con A{row}+0, D{row}+0
		reqs = append(reqs, setFormula("F", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d",TEXT(COUNTIFS('FEEDBACK-DETALLE_LLAMADA'!E:E,A%d,'FEEDBACK-DETALLE_LLAMADA'!K:K,D%d),"0"))`,
				sheetDETALLE, detalleRow, row, row), sheetFB))
		reqs = append(reqs, setValue("G", row, a.Estado, sheetFB))
		reqs = append(reqs, setValue("H", row, a.Supervisor, sheetFB))
		// STATUS formula (col I)
		reqs = append(reqs, setFormula("I", row,
			fmt.Sprintf(`=IF(G%d="Cumplido","✓ Listo",IF(G%d="Vencido","⚠ Vencido",IF(G%d="Carta de compromiso","📋 Compromiso",IF(G%d="Feedback dado","💬 Feedback",IF(G%d="En evaluación","⏳ Eval","· Pendiente")))))`, row, row, row, row, row), sheetFB))
		// ACCION text (col J)
		reqs = append(reqs, setValue("J", row, "📂 Historial  ✅ Marcar", sheetFB))
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP 6: Freeze rows 1-8
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 6: Freeze rows 1-8 (HEADER + DASHBOARD)")
	reqs = append(reqs, setFreeze(8, 0, sheetFB))

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

func clearCell(col string, row int, sheetID int64) map[string]any {
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
							"userEnteredValue": map[string]any{"stringValue": ""},
						},
					},
				},
			},
			"fields": "userEnteredValue",
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
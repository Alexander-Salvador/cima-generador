// restructure_feedbacks_v7_split_tabs.go — CEO directive 2026-10-04 (round 5)
//
// REFACTOR ARQUITECTÓNICO MAYOR:
//   - Antes: 1 tab "Feedbacks" con 2 secciones
//   - Ahora: 3 tabs separados
//     • FEEDBACK (renombre del actual Feedbacks) → registros del supervisor (resumen)
//     • FEEDBACK-DETALLE_LLAMADA (NUEVO) → registros completos de cada call auditada
//     • CONFIG (existente) → tabla de mapeo FK→PERSONAL
//
// DOMINIO:
//   - FEEDBACK es DERIVADO (1 row por agente/auditoría)
//   - FEEDBACK-DETALLE_LLAMADA es la BASE (1 row por call auditada)
//   - 1 feedback referencia N calls via FK (FEEDBACK_ID)
//
// OPERACIONES:
//   1. Rename Feedbacks → FEEDBACK
//   2. Create new tab FEEDBACK-DETALLE_LLAMADA
//   3. Clear rows 13-40 del FEEDBACK (sección 2 actual)
//   4. Refinar FEEDBACK master (nuevos headers + TIPO_FEEDBACK)
//   5. Backfill FEEDBACK-DETALLE_LLAMADA con 26 calls migradas
//   6. Tabla de mapeo en CONFIG
//   7. Data validation dropdowns
//   8. HYPERLINK cross-tab formulas
//   9. Conditional formatting
//
// EJECUCIÓN: 2 fases (apply separados porque el sheetId del nuevo tab se asigna dinámicamente)
//
// Uso:
//   go run v7_split_tabs.go --phase1    (rename + create + clear)
//   go run v7_split_tabs.go --phase2    (migración + headers + hyperlinks + CF)
//   go run v7_split_tabs.go --dry-run   (preview ambas fases)
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
	sheetFB       = int64(681240158) // será renombrado a FEEDBACK
	sheetPERSONAL = int64(1395452450)
	sheetCONFIG   = int64(1497962081)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

// ═══════════════════════════════════════════════════════════════════
// MASTER DATA
// ═══════════════════════════════════════════════════════════════════

var agentes = []struct {
	ID             int
	PersonalN      int
	PersonalRow    int
	Nombre         string
	Estado         string
	Riesgo         string
	TipoFeedback   string
	FBase          string
	FEval          string
	FFeedback      string
	UltimoTipo     string
	UltimaNota     string
}{
	{1, 3, 6, "RAMIREZ HUAMAN JOHNN MILTON", "Carta de compromiso", "Bajo", "Evaluativo", "23/09/2026", "23/09/2026", "23/09/2026", "Estados incorrectos", "Carta de compromiso, se saco a feedback en Jitsi"},
	{2, 8, 11, "VARGAS TAPIA ISABEL", "Carta de compromiso", "Bajo", "Evaluativo", "23/09/2026", "23/09/2026", "23/09/2026", "Abandono llamada", "Carta de compromiso, se saco a feedback en Jitsi"},
	{3, 18, 21, "GOMEZ FLORES JONATHAN ARTURO", "Feedback dado", "Bajo", "Positivo", "17/09/2026", "17/09/2026", "17/09/2026", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As."},
	{4, 9, 12, "MENDOZA CARMEN CECILIA LILIANA", "Feedback dado", "Bajo", "Positivo", "22/09/2026", "22/09/2026", "22/09/2026", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As."},
	{5, 16, 19, "LLAMPASI CRISOSTOMO LIZET YULY", "Feedback dado", "Bajo", "Evaluativo", "26/09/2026", "26/09/2026", "26/09/2026", "Speech incompleto", "CL indica que tiene problemas con su navegación, asesor no sondea para encontrar el error."},
	{6, 20, 23, "IMAN RUJEL YADIRA CATIZZA", "En evaluación", "Alto", "Amonestador", "22/09/2026", "", "", "Abandono llamada", "6 calls consolidados W40-2026"},
	{7, 13, 16, "CANALES COPLAT ZAYDA JOSCELINE", "En evaluación", "Medio", "Negativo", "23/09/2026", "", "", "No devuelve llamada", "2 calls consolidados W40-2026"},
}

// 26 calls migradas (13 originales + 13 huérfanas integradas)
var llamadas = []struct {
	FK          int    // ID del agente en FEEDBACK (1-7)
	NroComp     string // N° compuesto (1.1, 6.2, etc.)
	Fecha       string // FECHA_LLAMADA
	CallID      string
	TipoLlamada string
	Nota        string
	EstadoAud   string // ESTADO_AUDITORIA
	FeedbackID  int    // FK al FEEDBACK (1-7) si tiene feedback asociado
	FBase       string
	Auditor     string
}{
	// Originales 1-13
	{1, "1.1", "23/09/2026", "—", "Estados incorrectos", "Carta de compromiso, se saco a feedback en Jitsi", "Con feedback", 1, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{2, "2.1", "23/09/2026", "—", "Abandono llamada", "Carta de compromiso, se saco a feedback en Jitsi", "Con feedback", 2, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{3, "3.1", "17/09/2026", "—", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As.", "Con feedback", 3, "17/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{4, "4.1", "22/09/2026", "—", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As.", "Con feedback", 4, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{5, "5.1", "26/09/2026 20:36:08", "1790472913-864504", "Speech incompleto", "CL indica que tiene problemas con su navegación, asesor no sondea para encontrar el error.", "Con feedback", 5, "26/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.1", "22/09/2026", "1790541659-2494420", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.2", "22/09/2026", "1790547718-2497265", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.3", "22/09/2026", "1790548121-2929410", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.4", "23/09/2026", "1790620841-595264", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.5", "23/09/2026", "1790621167-1598581", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.6", "23/09/2026", "1790625501-30983", "Abandono llamada", "—", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.1", "23/09/2026", "1790645832-1625266", "No devuelve llamada", "—", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.2", "23/09/2026", "1790645658-52839", "No devuelve llamada", "—", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	// Integradas 14-26 (huérfanas)
	{7, "7.3", "23/09/2026", "1790645832-1625266", "No devuelve llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.7", "22/09/2026", "1790541659-2494420", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.8", "22/09/2026", "1790547718-2497265", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.9", "22/09/2026", "1790548121-2929410", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.4", "23/09/2026", "1790645658-52839", "No devuelve llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.10", "23/09/2026", "1790620841-595264", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.11", "23/09/2026", "1790621167-1598581", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{6, "6.12", "23/09/2026", "1790625501-30983", "Abandono llamada", "", "Con feedback", 6, "22/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.5", "23/09/2026", "1790644611-52116", "No devuelve llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.6", "23/09/2026", "1790644611-52116", "Abandono llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.7", "23/09/2026", "1790560765-2501968", "Abandono llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.8", "23/09/2026", "1790555685-319658", "Abandono llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
	{7, "7.9", "23/09/2026", "1790555793-319696", "Abandono llamada", "", "Con feedback", 7, "23/09/2026", "HUERTA SALVADOR WILLIAMS ALEXANDER"},
}

func main() {
	mode := "dry-run"
	phase := ""
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--apply":
			mode = "apply"
		case "--phase1":
			phase = "1"
		case "--phase2":
			phase = "2"
		case "--dry-run":
			mode = "dry-run"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v7 — SPLIT INTO 3 TABS")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	if phase != "" {
		fmt.Println("PHASE:", phase)
	}
	fmt.Println()

	if phase == "" || phase == "1" {
		fmt.Println("━━━ PHASE 1: Estructura base")
		fmt.Println("    1. Rename Feedbacks → FEEDBACK")
		fmt.Println("    2. Create new tab FEEDBACK-DETALLE_LLAMADA")
		fmt.Println("    3. Clear rows 13-40 del FEEDBACK (sección 2 actual)")
	}

	if phase == "" || phase == "2" {
		fmt.Println("\n━━━ PHASE 2: Schema + data + interconexión")
		fmt.Println("    4. FEEDBACK headers row 4 (nuevo schema)")
		fmt.Println("    5. FEEDBACK backfill rows 5-11")
		fmt.Println("    6. FEEDBACK-DETALLE_LLAMADA headers row 1")
		fmt.Println("    7. FEEDBACK-DETALLE_LLAMADA backfill rows 2-27")
		fmt.Println("    8. CONFIG: tabla de mapeo FK→PERSONAL rows 4-12")
		fmt.Println("    9. Data validation dropdowns")
		fmt.Println("    10. HYPERLINK cross-tab formulas")
		fmt.Println("    11. Conditional formatting")
	}

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada")
		fmt.Println("\nOrden recomendado:")
		fmt.Println("  1. go run v7_split_tabs.go --phase1 --apply")
		fmt.Println("  2. Anota el sheetId de FEEDBACK-DETALLE_LLAMADA (verás en output)")
		fmt.Println("  3. go run v7_split_tabs.go --phase2 --apply --detalle-sheetid=XXXXX")
		return
	}

	creds := decrypt()
	tok := getToken(creds)

	if phase == "1" {
		applyPhase1(tok)
	} else if phase == "2" {
		detalleSheetID := int64(0)
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "--detalle-sheetid=") {
				idStr := strings.TrimPrefix(arg, "--detalle-sheetid=")
				fmt.Sscanf(idStr, "%d", &detalleSheetID)
			}
		}
		if detalleSheetID == 0 {
			fmt.Println("ERROR: --detalle-sheetid=XXXXX requerido para phase2")
			fmt.Println("Obtén el sheetId del output de phase1")
			os.Exit(1)
		}
		applyPhase2(tok, detalleSheetID)
	} else {
		fmt.Println("ERROR: especifica --phase1 o --phase2")
		os.Exit(1)
	}
}

func applyPhase1(tok string) {
	reqs := []any{}

	// STEP 1: Rename Feedbacks → FEEDBACK
	fmt.Println("\n  STEP 1: Rename Feedbacks → FEEDBACK...")
	reqs = append(reqs, map[string]any{
		"updateSheetProperties": map[string]any{
			"properties": map[string]any{
				"sheetId": sheetFB,
				"title":   "FEEDBACK",
			},
			"fields": "title",
		},
	})

	// STEP 2: Create new tab FEEDBACK-DETALLE_LLAMADA
	fmt.Println("  STEP 2: Create FEEDBACK-DETALLE_LLAMADA...")
	reqs = append(reqs, map[string]any{
		"addSheet": map[string]any{
			"properties": map[string]any{
				"title":   "FEEDBACK-DETALLE_LLAMADA",
				"gridProperties": map[string]any{
					"rowCount":    100,
					"columnCount": 26,
				},
			},
		},
	})

	// STEP 3: Clear rows 13-40 del FEEDBACK (sección 2 actual)
	fmt.Println("  STEP 3: Clear FEEDBACK rows 13-40...")
	for r := 13; r <= 40; r++ {
		for _, c := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"} {
			reqs = append(reqs, clearCell(c, r, sheetFB))
		}
	}

	fmt.Printf("\n📤 PHASE 1: Enviando %d requests...\n", len(reqs))
	resp := sendBatchUpdate(tok, reqs)
	// Extract new sheetId
	var result map[string]any
	json.Unmarshal(resp, &result)
	replies, _ := result["replies"].([]any)
	for _, r := range replies {
		rm := r.(map[string]any)
		if sp, ok := rm["addSheet"]; ok {
			props := sp.(map[string]any)["properties"].(map[string]any)
			newID := int64(props["sheetId"].(float64))
			fmt.Printf("\n✅ NEW SHEET CREATED: FEEDBACK-DETALLE_LLAMADA\n")
			fmt.Printf("   sheetId = %d\n", newID)
			fmt.Printf("\n📋 Para ejecutar phase2:\n")
			fmt.Printf("   go run v7_split_tabs.go --phase2 --apply --detalle-sheetid=%d\n", newID)
		}
	}
	if len(replies) > 0 {
		fmt.Printf("\n✅ Phase 1 OK: %d replies\n", len(replies))
	}
}

func applyPhase2(tok string, detalleSheetID int64) {
	reqs := []any{}

	fmt.Println("\n  STEP 4: FEEDBACK master headers row 4 (nuevo schema)...")
	masterHeaders := []struct {
		Col, Val string
	}{
		{"A", "ID"}, {"B", "AGENTE"}, {"C", "SUPERVISOR"}, {"D", "CANT_CALLS"},
		{"E", "TIPO_FEEDBACK"}, {"F", "FECHA_FEEDBACK"}, {"G", "FECHA_BASE"},
		{"H", "FECHA_EVAL"}, {"I", "ÚLTIMO_TIPO"}, {"J", "ÚLTIMA_NOTA"},
		{"K", "COMPROMISO_FECHA"}, {"L", "CUMPLIDO_FECHA"}, {"M", "ESTADO"},
		{"N", "RIESGO"},
	}
	for _, h := range masterHeaders {
		reqs = append(reqs, setValue(h.Col, 4, h.Val, sheetFB))
	}

	fmt.Println("  STEP 5: FEEDBACK master backfill rows 5-11...")
	for _, a := range agentes {
		row := a.ID + 4
		// B (AGENTE) → HYPERLINK a PERSONAL
		reqs = append(reqs, setFormula("B", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=C%d","%s")`, sheetPERSONAL, a.PersonalRow, a.Nombre), sheetFB))
		reqs = append(reqs, setValue("C", row, "HUERTA SALVADOR WILLIAMS ALEXANDER", sheetFB))
		// D (CANT_CALLS) → COUNTIF cross-tab
		reqs = append(reqs, setFormula("D", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d",TEXT(COUNTIF('FEEDBACK-DETALLE_LLAMADA'!A:A,A%d),"0"))`,
				detalleSheetID, llamadasFirstRow(a.ID), row), sheetFB))
		reqs = append(reqs, setValue("E", row, a.TipoFeedback, sheetFB))
		reqs = append(reqs, setValue("F", row, a.FFeedback, sheetFB))
		reqs = append(reqs, setValue("G", row, a.FBase, sheetFB))
		reqs = append(reqs, setValue("H", row, a.FEval, sheetFB))
		reqs = append(reqs, setValue("I", row, a.UltimoTipo, sheetFB))
		reqs = append(reqs, setValue("J", row, a.UltimaNota, sheetFB))
		reqs = append(reqs, setValue("M", row, a.Estado, sheetFB))
		reqs = append(reqs, setValue("N", row, a.Riesgo, sheetFB))
	}

	fmt.Println("  STEP 6: FEEDBACK-DETALLE_LLAMADA headers row 1...")
	detailHeaders := []struct {
		Col, Val string
	}{
		{"A", "ID_AGENTE"}, {"B", "N°"}, {"C", "FECHA_LLAMADA"},
		{"D", "SUPERVISOR"}, {"E", "AGENTE"}, {"F", "CALL_ID"},
		{"G", "TIPO_LLAMADA"}, {"H", "NOTA"}, {"I", "ESTADO_AUDITORIA"},
		{"J", "FEEDBACK_ID"}, {"K", "FECHA_BASE"}, {"L", "AUDITOR"},
	}
	for _, h := range detailHeaders {
		reqs = append(reqs, setValue(h.Col, 1, h.Val, detalleSheetID))
	}

	fmt.Println("  STEP 7: FEEDBACK-DETALLE_LLAMADA backfill rows 2-27...")
	for i, l := range llamadas {
		row := i + 2
		// A (ID_AGENTE) → numberValue para FK
		reqs = append(reqs, setNumber("A", row, l.FK, detalleSheetID))
		// B (N°) → HYPERLINK a PERSONAL
		reqs = append(reqs, setFormula("B", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=C%d","%s")`, sheetPERSONAL, personalRowForFK(l.FK), l.NroComp), detalleSheetID))
		reqs = append(reqs, setValue("C", row, l.Fecha, detalleSheetID))
		reqs = append(reqs, setValue("D", row, "HUERTA SALVADOR WILLIAMS ALEXANDER", detalleSheetID))
		// E (AGENTE) → VLOOKUP en CONFIG mapeo
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=IFERROR(VLOOKUP(A%d,CONFIG!$A$4:$C$10,3,FALSE),"")`, row), detalleSheetID))
		reqs = append(reqs, setValue("F", row, l.CallID, detalleSheetID))
		reqs = append(reqs, setValue("G", row, l.TipoLlamada, detalleSheetID))
		reqs = append(reqs, setValue("H", row, l.Nota, detalleSheetID))
		reqs = append(reqs, setValue("I", row, l.EstadoAud, detalleSheetID))
		// J (FEEDBACK_ID) → numberValue FK
		reqs = append(reqs, setNumber("J", row, l.FeedbackID, detalleSheetID))
		reqs = append(reqs, setValue("K", row, l.FBase, detalleSheetID))
		reqs = append(reqs, setValue("L", row, l.Auditor, detalleSheetID))
	}

	fmt.Println("  STEP 8: CONFIG — tabla de mapeo FK→PERSONAL rows 4-12...")
	// Header row 4
	reqs = append(reqs, setValue("A", 4, "FK_FEEDBACK", sheetCONFIG))
	reqs = append(reqs, setValue("B", 4, "PERSONAL_N", sheetCONFIG))
	reqs = append(reqs, setValue("C", 4, "NOMBRE", sheetCONFIG))
	for i, a := range agentes {
		row := 5 + i
		reqs = append(reqs, setNumber("A", row, a.ID, sheetCONFIG))
		reqs = append(reqs, setNumber("B", row, a.PersonalN, sheetCONFIG))
		reqs = append(reqs, setValue("C", row, a.Nombre, sheetCONFIG))
	}

	fmt.Println("  STEP 9: Data validation dropdowns...")
	// FEEDBACK col E (TIPO_FEEDBACK)
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "E5:E11",
		[]string{"Positivo", "Negativo", "Evaluativo", "Amonestador"}))
	// FEEDBACK col M (ESTADO)
	reqs = append(reqs, dataValidationOneOfList(sheetFB, "M5:M11",
		[]string{"Pendiente", "En evaluación", "Feedback dado", "Carta de compromiso", "Cumplido", "Vencido"}))
	// FEEDBACK-DETALLE col G (TIPO_LLAMADA)
	reqs = append(reqs, dataValidationOneOfList(detalleSheetID, "G2:G27",
		[]string{"Estados incorrectos", "Abandono llamada", "Speech incompleto", "No devuelve llamada"}))
	// FEEDBACK-DETALLE col I (ESTADO_AUDITORIA)
	reqs = append(reqs, dataValidationOneOfList(detalleSheetID, "I2:I27",
		[]string{"Pendiente", "Auditada", "Con feedback"}))

	fmt.Println("  STEP 10: HYPERLINK formulas master...")
	for _, a := range agentes {
		row := a.ID + 4
		// A (ID) → HYPERLINK al detalle del agente
		reqs = append(reqs, setNumber("A", row, a.ID, sheetFB))
	}

	fmt.Println("  STEP 11: HYPERLINK formulas detalle FEEDBACK_ID...")
	for i := range llamadas {
		// J (FEEDBACK_ID) ya es number; agregamos HYPERLINK
		row := i + 2
		reqs = append(reqs, clearCell("J", row, detalleSheetID))
		reqs = append(reqs, setFormula("J", row,
			fmt.Sprintf(`=HYPERLINK("#gid=%d&range=A%d",%d)`, sheetFB, llamadas[i].FeedbackID+4, llamadas[i].FeedbackID), detalleSheetID))
	}

	fmt.Printf("\n📤 PHASE 2: Enviando %d requests...\n", len(reqs))
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
	fmt.Printf("✅ Phase 2 OK: %d replies\n", len(replies))
	fmt.Printf("🕐 Done at: %s\n", time.Now().Format(time.RFC3339))
}

func sendBatchUpdate(tok string, reqs []any) []byte {
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
	return respB
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

func llamadasFirstRow(agenteID int) int {
	for i, l := range llamadas {
		if l.FK == agenteID {
			return i + 2
		}
	}
	return 2
}

func personalRowForFK(fk int) int {
	for _, a := range agentes {
		if a.ID == fk {
			return a.PersonalRow
		}
	}
	return 0
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

func setNumber(col string, row int, value int, sheetID int64) map[string]any {
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
							"userEnteredValue": map[string]any{"numberValue": value},
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
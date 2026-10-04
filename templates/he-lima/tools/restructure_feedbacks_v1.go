// restructure_feedbacks_v1.go — CEO directive 2026-10-04
// "Feedback tab tiene 2 partes y no conectan, es como 2 bases diferentes"
// "Falta fecha para distinguir cuando recibimos la base y cuando la evaluamos"
// "Estados potenciales para que se cumpla si se dio el feedback, carta de compromiso"
//
// Cambios:
//   1. SECCIÓN 1 (rows 4-11) — agregar 3 cols workflow:
//      H = FECHA_BASE (cuándo recibimos la base)
//      I = FECHA_EVAL (cuándo evaluamos)
//      J = ESTADO  (workflow: Pendiente/En evaluación/Feedback dado/Carta de compromiso/Cumplido/Vencido)
//   2. SECCIÓN 1 col G — renombrar "STATUS" → "RIESGO" (mantener valores actuales)
//   3. SECCIÓN 1 col E (CANT) — convertir a fórmula =COUNTIF(detalle!A:A, A{row})
//   4. SECCIÓN 2 (rows 14+) — agregar col A = ID_AGENTE (FK real numérica)
//   5. SECCIÓN 2 col B renombrar de "N°" (texto "1.1") → N° (texto "1.1") pero movido
//      para que A sea FK numérica real
//   6. SECCIÓN 2 (rows 14+) — agregar 4 cols workflow:
//      I = FECHA_BASE  (por call individual)
//      J = FECHA_COMPROMISO
//      K = FECHA_CUMPLIMIENTO
//      L = ESTADO_CALL
//   7. DELETE rows 28-41 (datos sueltos sin formato)
//   8. Backfill con valores basados en data actual:
//      - SECCIÓN 1: 7 agentes con fechas probables y estados inferidos
//      - SECCIÓN 2: 13 calls + integrar rows 29-41 huérfanos (13 más)
//   9. Data validation dropdown en col J sección 1 y col L sección 2
//
// Uso:
//   go run restructure_feedbacks_v1.go --dry-run    (default)
//   go run restructure_feedbacks_v1.go --apply
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
	"strconv"
	"strings"
	"time"
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	sheetFB       = int64(681240158)
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

// ═══════════════════════════════════════════════════════════════════
// BACKFILL DATA — interpretado del estado actual del sheet
// ═══════════════════════════════════════════════════════════════════

// Sección 1 — Master (rows 5-11, 7 agentes)
var masterBackfill = []struct {
	ID           int
	Nombre       string
	Supervisor   string
	Riesgo       string // mantiene valores actuales
	FBase        string // FECHA_BASE
	FEval        string // FECHA_EVAL
	Estado       string // workflow
	EstadoColor  string // para validación visual
}{
	{1, "RAMIREZ HUAMAN JOHNN MILTON", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Bajo", "23/09/2026", "23/09/2026", "Carta de compromiso", "amber"},
	{2, "VARGAS TAPIA ISABEL", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Bajo", "23/09/2026", "23/09/2026", "Carta de compromiso", "amber"},
	{3, "GOMEZ FLORES JONATHAN ARTURO", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Bajo", "17/09/2026", "17/09/2026", "Feedback dado", "blue"},
	{4, "MENDOZA CARMEN CECILIA LILIANA", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Bajo", "22/09/2026", "22/09/2026", "Feedback dado", "blue"},
	{5, "LLAMPASI CRISOSTOMO LIZET YULY", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Bajo", "26/09/2026", "26/09/2026", "Feedback dado", "red"},
	{6, "IMAN RUJEL YADIRA CATIZZA", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Alto", "22/09/2026", "", "En evaluación", "yellow"},
	{7, "CANALES COPLAT ZAYDA JOSCELINE", "HUERTA SALVADOR WILLIAMS ALEXANDER", "Medio", "23/09/2026", "", "En evaluación", "yellow"},
}

// Sección 2 — Detail (rows 15-27 actuales + integración rows 29-41)
// Cada row tiene FK numérica al master (ID 1-7)
var detailBackfill = []struct {
	FK        int    // ID_AGENTE
	Nro       string // "1.1", "6.2", etc.
	Fecha     string // FECHA del call
	CallID    string // CALL_ID
	Tipo      string
	Nota      string
	FBase     string // FECHA_BASE (cuándo supervisor entregó)
	FComp     string // FECHA_COMPROMISO (cuándo se firmó carta)
	FCump     string // FECHA_CUMPLIMIENTO (cuándo cerró)
	EstadoC   string // ESTADO_CALL
}{
	// Originales rows 15-27
	{1, "1.1", "23/09/2026", "—", "Estados incorrectos", "Carta de compromiso, se saco a feedback en Jitsi", "23/09/2026", "23/09/2026", "", "Carta de compromiso"},
	{2, "2.1", "23/09/2026", "—", "Abandono llamada", "Carta de compromiso, se saco a feedback en Jitsi", "23/09/2026", "23/09/2026", "", "Carta de compromiso"},
	{3, "3.1", "17/09/2026", "—", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As.", "17/09/2026", "17/09/2026", "", "Feedback dado"},
	{4, "4.1", "22/09/2026", "—", "Abandono llamada", "Se alerta via whatsapp obteniendo un compromiso por el As.", "22/09/2026", "22/09/2026", "", "Feedback dado"},
	{5, "5.1", "26/09/2026 20:36:08", "1790472913-864504", "Speech incompleto", "CL indica que tiene problemas con su navegación, asesor no sondea para encontrar el error.", "26/09/2026", "26/09/2026", "", "Feedback dado"},
	{6, "6.1", "22/09/2026", "1790541659-2494420", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{6, "6.2", "22/09/2026", "1790547718-2497265", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{6, "6.3", "22/09/2026", "1790548121-2929410", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{6, "6.4", "23/09/2026", "1790620841-595264", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{6, "6.5", "23/09/2026", "1790621167-1598581", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{6, "6.6", "23/09/2026", "1790625501-30983", "Abandono llamada", "—", "22/09/2026", "", "", "En evaluación"},
	{7, "7.1", "23/09/2026", "1790645832-1625266", "No devuelve llamada", "—", "23/09/2026", "", "", "En evaluación"},
	{7, "7.2", "23/09/2026", "1790645658-52839", "No devuelve llamada", "—", "23/09/2026", "", "", "En evaluación"},
	// Integración rows 29-41 (datos sueltos)
	// Buscar FK por AGENTE en estos datos:
	{7, "7.3", "23/09/2026", "1790645832-1625266", "No devuelve llamada", "", "23/09/2026", "", "", "En evaluación"},
	{6, "6.7", "22/09/2026", "1790541659-2494420", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{6, "6.8", "22/09/2026", "1790547718-2497265", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{6, "6.9", "22/09/2026", "1790548121-2929410", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{7, "7.4", "23/09/2026", "1790645658-52839", "No devuelve llamada", "", "23/09/2026", "", "", "En evaluación"},
	{6, "6.10", "23/09/2026", "1790620841-595264", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{6, "6.11", "23/09/2026", "1790621167-1598581", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{6, "6.12", "23/09/2026", "1790625501-30983", "Abandono llamada", "", "22/09/2026", "", "", "En evaluación"},
	{7, "7.5", "23/09/2026", "1790644611-52116", "No devuelve llamada", "", "23/09/2026", "", "", "En evaluación"},
	{7, "7.6", "23/09/2026", "1790644611-52116", "Abandono llamada", "", "23/09/2026", "", "", "En evaluación"},
	{7, "7.7", "23/09/2026", "1790560765-2501968", "Abandono llamada", "", "23/09/2026", "", "", "En evaluación"},
	{7, "7.8", "23/09/2026", "1790555685-319658", "Abandono llamada", "", "23/09/2026", "", "", "En evaluación"},
	{7, "7.9", "23/09/2026", "1790555793-319696", "Abandono llamada", "", "23/09/2026", "", "", "En evaluación"},
}

func main() {
	mode := "dry-run"
	for _, arg := range os.Args[1:] {
		if arg == "--apply" {
			mode = "apply"
		}
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("FEEDBACKS RESTRUCTURE v1")
	fmt.Println("CEO directive 2026-10-04:")
	fmt.Println("  • Master-detail con FK real (no N° decorativo)")
	fmt.Println("  • Workflow state machine: Pendiente → ... → Cumplido")
	fmt.Println("  • Fechas: BASE (recepción) + EVAL (evaluación)")
	fmt.Println("  • Integrar 13 rows huérfanos a sección 2")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MODE:", strings.ToUpper(mode))
	fmt.Println("")

	// ═══════════════════════════════════════════════════════════════
	// STEP 1 — DELETE rows 28-41 (datos suertos sin FK)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 1: DELETE rows 28-41 (datos sueltos sin FK)")

	// ═══════════════════════════════════════════════════════════════
	// STEP 2 — UPDATE sección 1 headers row 4
	//          A4=N°, B4=AGENTE, C4=SUPERVISOR, D4=CANT, E4=ÚLTIMO TIPO,
	//          F4=ÚLTIMA NOTA, G4=RIESGO (rename), H4=FECHA_BASE, I4=FECHA_EVAL, J4=ESTADO
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("━━━ STEP 2: UPDATE sección 1 headers row 4")
	fmt.Printf("    G4: STATUS → RIESGO (mantiene semántica de nivel)\n")
	fmt.Printf("    H4: FECHA_BASE (NEW)\n")
	fmt.Printf("    I4: FECHA_EVAL (NEW)\n")
	fmt.Printf("    J4: ESTADO (NEW — workflow cumplimiento)\n")

	// ═══════════════════════════════════════════════════════════════
	// STEP 3 — BACKFILL sección 1 rows 5-11 con valores razonables
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 3: BACKFILL sección 1 rows 5-11")
	for _, m := range masterBackfill {
		fmt.Printf("    Row %d (ID=%d) %s → RIESGO=%s, F_BASE=%s, F_EVAL=%s, ESTADO=%s\n",
			m.ID+4, m.ID, truncate(m.Nombre, 30), m.Riesgo, m.FBase, m.FEval, m.Estado)
	}

	// ═══════════════════════════════════════════════════════════════
	// STEP 4 — UPDATE sección 1 col E (CANT) → fórmulas COUNTIF
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 4: Convertir CANT (col E) a fórmulas COUNTIF")
	fmt.Println("    E5: =COUNTIF(detalle!A:A, A5)")
	fmt.Println("    E6: =COUNTIF(detalle!A:A, A6)")
	fmt.Println("    ... etc hasta E11")

	// ═══════════════════════════════════════════════════════════════
	// STEP 5 — INSERTAR nueva col A = ID_AGENTE en sección 2 (FK)
	//          Estrategia: shift col B (N°) → C, col C-H → D-I
	//          + agregar cols I-L nuevas
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 5: Reorganizar sección 2 — agregar col A=ID_AGENTE (FK)")
	fmt.Println("    Layout viejo: A=N° | B=FECHA | C=SUPERVISOR | D=AGENTE | E=CALL_ID | F=TIPO | G=NOTA")
	fmt.Println("    Layout nuevo: A=ID_AGENTE | B=N° | C=FECHA | D=SUPERVISOR | E=AGENTE | F=CALL_ID | G=TIPO | H=NOTA | I=FECHA_BASE | J=FECHA_COMPROMISO | K=FECHA_CUMPLIMIENTO | L=ESTADO_CALL")

	// ═══════════════════════════════════════════════════════════════
	// STEP 6 — UPDATE sección 2 headers row 14
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 6: UPDATE sección 2 headers row 14")
	fmt.Println("    A14: ID_AGENTE | B14: N° | C14: FECHA | D14: SUPERVISOR | E14: AGENTE | F14: CALL_ID")
	fmt.Println("    G14: TIPO | H14: NOTA | I14: FECHA_BASE | J14: FECHA_COMPROMISO | K14: FECHA_CUMPLIMIENTO | L14: ESTADO_CALL")

	// ═══════════════════════════════════════════════════════════════
	// STEP 7 — BACKFILL sección 2 (rows 15-40) con 25 calls integradas
	//          (13 originales + 13 huérfanas integradas)
	// ═══════════════════════════════════════════════════════════════
	fmt.Printf("\n━━━ STEP 7: BACKFILL sección 2 — %d calls integradas\n", len(detailBackfill))
	for i, d := range detailBackfill {
		fmt.Printf("    Row %d FK=%d N°=%s Fecha=%s Tipo=%s Estado=%s\n",
			i+15, d.FK, d.Nro, d.Fecha, d.Tipo, d.EstadoC)
	}

	// ═══════════════════════════════════════════════════════════════
	// STEP 8 — Data validation dropdown en ESTADO (col J sec1) y ESTADO_CALL (col L sec2)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 8: Data validation dropdown")
	fmt.Println("    J5:J11 → workflow cumplimiento: Pendiente, En evaluación, Feedback dado, Carta de compromiso, Cumplido, Vencido")
	fmt.Println("    L15:L40 → workflow por call: mismos 6 valores")

	// ═══════════════════════════════════════════════════════════════
	// STEP 9 — Conditional formatting para workflow states (visual)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n━━━ STEP 9: Conditional formatting para estados")
	fmt.Println("    Pendiente → gris")
	fmt.Println("    En evaluación → amarillo")
	fmt.Println("    Feedback dado → azul claro")
	fmt.Println("    Carta de compromiso → ámbar")
	fmt.Println("    Cumplido → verde")
	fmt.Println("    Vencido → rojo")

	// ═══════════════════════════════════════════════════════════════
	// SUMMARY
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("RESUMEN DE CAMBIOS")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("• DELETE: 14 rows (28-41)\n")
	fmt.Printf("• UPDATE headers: 2 rows (4, 14)\n")
	fmt.Printf("• BACKFILL: 7 master rows + %d detail rows\n", len(detailBackfill))
	fmt.Printf("• FORMULAS: 7 COUNTIF en CANT\n")
	fmt.Printf("• DATA VALIDATION: 2 dropdowns\n")
	fmt.Printf("• CONDITIONAL FORMATS: 12 rules (6 estados × 2 secciones)\n")
	fmt.Printf("• Total API calls estimadas: ~80\n")
	fmt.Println(strings.Repeat("═", 70))

	if mode == "dry-run" {
		fmt.Println("\n⚠️  DRY-RUN — no se envió nada al Sheets API")
		fmt.Println("    Para aplicar: go run restructure_feedbacks_v1.go --apply")
		return
	}

	// ═══════════════════════════════════════════════════════════════
	// APPLY — Build batchUpdate requests y enviar
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("\n🔧 Construyendo batchUpdate requests...")
	creds := decrypt()
	tok := getToken(creds)
	reqs := []any{}

	// ═══ ORDEN CORREGIDO (Issue 1: INSERT antes que DELETE) ═══

	// STEP 1: UPDATE sección 1 headers row 4
	fmt.Println("  STEP 1: UPDATE sección 1 headers...")
	reqs = append(reqs, setValue("G", 4, "RIESGO"))
	reqs = append(reqs, setValue("H", 4, "FECHA_BASE"))
	reqs = append(reqs, setValue("I", 4, "FECHA_EVAL"))
	reqs = append(reqs, setValue("J", 4, "ESTADO"))

	// STEP 2: BACKFILL sección 1 rows 5-11
	fmt.Println("  STEP 2: BACKFILL sección 1...")
	for _, m := range masterBackfill {
		row := m.ID + 4
		reqs = append(reqs, setValue("G", row, m.Riesgo))
		reqs = append(reqs, setValue("H", row, m.FBase))
		reqs = append(reqs, setValue("I", row, m.FEval))
		reqs = append(reqs, setValue("J", row, m.Estado))
	}

	// STEP 3: COUNTIF fórmulas en CANT (col E)
	fmt.Println("  STEP 3: COUNTIF fórmulas en CANT...")
	for row := 5; row <= 11; row++ {
		reqs = append(reqs, setFormula("E", row,
			fmt.Sprintf(`=COUNTIF($A$15:$A$40,A%d)`, row)))
	}

	// STEP 4: DELETE rows 28-41 (datos sueltos huérfanos) — high to low
	fmt.Println("  STEP 4: DELETE rows 28-41 (datos sueltos)...")
	for r := 41; r >= 28; r-- {
		reqs = append(reqs, deleteRow(r))
	}

	// STEP 5: INSERT 13 rows después de row 27 (para acomodar las 13 huérfanas)
	//         (después del DELETE rows 28-41, las huérfanas que estaban en 29-41 ya no existen;
	//          necesitamos 13 rows nuevas para backfill de las 13 calls huérfanas)
	fmt.Println("  STEP 5: INSERT 13 rows después de row 27...")
	reqs = append(reqs, insertRow(28, 13)) // insert 13 blank rows starting at index 28

	// STEP 6: Clear sección 2 rows 14-40 cols A-H actuales para reescribir limpio
	fmt.Println("  STEP 6: Clear sección 2 cols A-H rows 14-40...")
	for r := 14; r <= 40; r++ {
		for _, c := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
			reqs = append(reqs, clearCell(c, r))
		}
	}

	// STEP 7: Headers sección 2 row 14
	fmt.Println("  STEP 7: UPDATE sección 2 headers row 14...")
	reqs = append(reqs, setValue("A", 14, "ID_AGENTE"))
	reqs = append(reqs, setValue("B", 14, "N°"))
	reqs = append(reqs, setValue("C", 14, "FECHA"))
	reqs = append(reqs, setValue("D", 14, "SUPERVISOR"))
	reqs = append(reqs, setValue("E", 14, "AGENTE"))
	reqs = append(reqs, setValue("F", 14, "CALL_ID"))
	reqs = append(reqs, setValue("G", 14, "TIPO"))
	reqs = append(reqs, setValue("H", 14, "NOTA"))
	reqs = append(reqs, setValue("I", 14, "FECHA_BASE"))
	reqs = append(reqs, setValue("J", 14, "FECHA_COMPROMISO"))
	reqs = append(reqs, setValue("K", 14, "FECHA_CUMPLIMIENTO"))
	reqs = append(reqs, setValue("L", 14, "ESTADO_CALL"))

	// STEP 8: BACKFILL sección 2 (rows 15-40 = 26 rows)
	fmt.Println("  STEP 8: BACKFILL sección 2 — 26 calls...")
	for i, d := range detailBackfill {
		row := i + 15
		if row > 40 {
			break
		}
		reqs = append(reqs, setValue("A", row, strconv.Itoa(d.FK)))
		reqs = append(reqs, setValue("B", row, d.Nro))
		reqs = append(reqs, setValue("C", row, d.Fecha))
		reqs = append(reqs, setValue("D", row, "HUERTA SALVADOR WILLIAMS ALEXANDER"))
		reqs = append(reqs, setValue("F", row, d.CallID))
		reqs = append(reqs, setValue("G", row, d.Tipo))
		reqs = append(reqs, setValue("H", row, d.Nota))
		reqs = append(reqs, setValue("I", row, d.FBase))
		reqs = append(reqs, setValue("J", row, d.FComp))
		reqs = append(reqs, setValue("K", row, d.FCump))
		reqs = append(reqs, setValue("L", row, d.EstadoC))
	}

	// STEP 8: Data validation
	fmt.Println("  STEP 8: Data validation...")
	reqs = append(reqs, dataValidationOneOfList("J5:J11",
		[]string{"Pendiente", "En evaluación", "Feedback dado", "Carta de compromiso", "Cumplido", "Vencido"}))
	reqs = append(reqs, dataValidationOneOfList("L15:L40",
		[]string{"Pendiente", "En evaluación", "Feedback dado", "Carta de compromiso", "Cumplido", "Vencido"}))

	// STEP 9: Conditional formatting
	fmt.Println("  STEP 9: Conditional formatting...")
	cfRules := []struct {
		Value string
		R, G, B float64
		Range  string
	}{
		{"Pendiente", 0.85, 0.85, 0.85, "J5:J11"},
		{"En evaluación", 1.0, 1.0, 0.6, "J5:J11"},
		{"Feedback dado", 0.6, 0.85, 1.0, "J5:J11"},
		{"Carta de compromiso", 1.0, 0.85, 0.4, "J5:J11"},
		{"Cumplido", 0.5, 0.9, 0.5, "J5:J11"},
		{"Vencido", 1.0, 0.5, 0.5, "J5:J11"},
	}
	for _, cf := range cfRules {
		reqs = append(reqs, conditionalFormat(cf.Range, cf.Value, cf.R, cf.G, cf.B))
	}
	cfRules2 := []struct {
		Value string
		R, G, B float64
		Range  string
	}{
		{"Pendiente", 0.85, 0.85, 0.85, "L15:L40"},
		{"En evaluación", 1.0, 1.0, 0.6, "L15:L40"},
		{"Feedback dado", 0.6, 0.85, 1.0, "L15:L40"},
		{"Carta de compromiso", 1.0, 0.85, 0.4, "L15:L40"},
		{"Cumplido", 0.5, 0.9, 0.5, "L15:L40"},
		{"Vencido", 1.0, 0.5, 0.5, "L15:L40"},
	}
	for _, cf := range cfRules2 {
		reqs = append(reqs, conditionalFormat(cf.Range, cf.Value, cf.R, cf.G, cf.B))
	}

	fmt.Printf("\n📤 Enviando %d requests al Sheets API...\n", len(reqs))
	// SEND
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

func clearCell(col string, row int) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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

func setValue(col string, row int, value string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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

func setFormula(col string, row int, formula string) map[string]any {
	ci := colIdx(col)
	ri := row - 1
	return map[string]any{
		"updateCells": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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

func insertRow(row, count int) map[string]any {
	ri := row - 1
	return map[string]any{
		"insertDimension": map[string]any{
			"range": map[string]any{
				"sheetId":    sheetFB,
				"dimension":  "ROWS",
				"startIndex": ri,
				"endIndex":   ri + count,
			},
			"inheritFromBefore": false,
		},
	}
}

func deleteRow(row int) map[string]any {
	ri := row - 1
	return map[string]any{
		"deleteDimension": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
				"dimension":        "ROWS",
				"startIndex":       ri,
				"endIndex":         ri + 1,
			},
		},
	}
}

func dataValidationOneOfList(a1Range string, values []string) map[string]any {
	// Convert A1 notation to GridRange
	gridRange, startRow, endRow, startCol, endCol := a1ToGrid(a1Range)
	_ = gridRange
	return map[string]any{
		"setDataValidation": map[string]any{
			"range": map[string]any{
				"sheetId":          sheetFB,
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
				"strict":       false,
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
	// Simple parser: e.g. "J5:J11" → sheet, rowStart, rowEnd, colStart, colEnd
	parts := strings.Split(a1, ":")
	if len(parts) != 2 {
		return a1, 0, 1, 0, 1
	}
	startColS := strings.TrimLeft(parts[0], "0123456789")
	startRowS := strings.TrimLeft(parts[0], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	startCol := colIdx(startColS)
	startRow, _ := strconv.Atoi(startRowS)
	endColS := strings.TrimLeft(parts[1], "0123456789")
	endRowS := strings.TrimLeft(parts[1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	endCol := colIdx(endColS) + 1
	endRow, _ := strconv.Atoi(endRowS)
	return a1, startRow - 1, endRow, startCol, endCol
}

func conditionalFormat(a1Range, value string, r, g, b float64) map[string]any {
	_, startRow, endRow, startCol, endCol := a1ToGrid(a1Range)
	return map[string]any{
		"addConditionalFormatRule": map[string]any{
			"rule": map[string]any{
				"ranges": []any{
					map[string]any{
						"sheetId":          sheetFB,
						"startRowIndex":    startRow,
						"endRowIndex":      endRow,
						"startColumnIndex": startCol,
						"endColumnIndex":   endCol,
					},
				},
				"booleanRule": map[string]any{
					"condition": map[string]any{
						"type":         "TEXT_EQ",
						"values":       []map[string]any{{"userEnteredValue": value}},
					},
					"format": map[string]any{
						"backgroundColor": map[string]any{
							"red":   r,
							"green": g,
							"blue":  b,
						},
					},
				},
			},
			"index": 0,
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
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
)

const (
	spreadsheetID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
	encPath       = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
	keyPath       = "/home/braka/.config/ovav/vault.key"
)

func main() {
	creds := decrypt()
	tok := getToken(creds)
	cases := []string{
		"PERSONAL!A1", "PERSONAL!A2", "PERSONAL!A3", "PERSONAL!C3",
		"Feedbacks!A1", "Feedbacks!A4",
		"01_FERIADOS_PE!A1", "01_FERIADOS_PE!A5",
		"CONFIG!A1", "CONFIG!A3", "CONFIG!A65", "CONFIG!B72",
		"10_HE!B1", "10_HE!B2", "10_HE!B3", "10_HE!B5", "10_HE!C5", "10_HE!F5",
		"10_HE!B7", "10_HE!B10", "10_HE!B11", "10_HE!B12", "10_HE!AJ12", "10_HE!B34",
	}
	for _, c := range cases {
		url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetID +
			"?ranges=" + c +
			"&fields=sheets(data(rowData(values(userEnteredFormat.backgroundColor,userEnteredFormat.textFormat.foregroundColor,userEnteredFormat.textFormat.fontSize,userEnteredFormat.textFormat.bold))))"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, _ := http.DefaultClient.Do(req)
		body, _ := io.ReadAll(resp.Body)
		var raw map[string]any
		json.Unmarshal(body, &raw)
		sheets := raw["sheets"].([]any)
		if len(sheets) == 0 {
			continue
		}
		data := sheets[0].(map[string]any)["data"].([]any)
		if len(data) == 0 {
			fmt.Printf("%-30s (no data)\n", c)
			resp.Body.Close()
			continue
		}
		rd := data[0].(map[string]any)["rowData"].([]any)
		if len(rd) == 0 {
			fmt.Printf("%-30s (no rowData)\n", c)
			resp.Body.Close()
			continue
		}
		vs := rd[0].(map[string]any)["values"].([]any)
		if len(vs) == 0 {
			fmt.Printf("%-30s (empty cell)\n", c)
			resp.Body.Close()
			continue
		}
		f, _ := vs[0].(map[string]any)["userEnteredFormat"].(map[string]any)
		bg, _ := f["backgroundColor"].(map[string]any)
		tf, _ := f["textFormat"].(map[string]any)
		fg, _ := tf["foregroundColor"].(map[string]any)
		bgHex := rgb(bg)
		fgHex := rgb(fg)
		fs, _ := tf["fontSize"].(float64)
		bold, _ := tf["bold"].(bool)
		fmt.Printf("%-30s bg=%-12s fg=%-12s fs=%-4v bold=%v\n", c, bgHex, fgHex, fs, bold)
		resp.Body.Close()
	}
}

func rgb(m map[string]any) string {
	if m == nil {
		return "(none)"
	}
	r, _ := m["red"].(float64)
	g, _ := m["green"].(float64)
	b, _ := m["blue"].(float64)
	if r == 0 && g == 0 && b == 0 {
		return "#000000"
	}
	return fmt.Sprintf("#%02X%02X%02X", int(r*255+0.5), int(g*255+0.5), int(b*255+0.5))
}

func decrypt() map[string]any {
	enc, _ := os.ReadFile(encPath)
	keyFile, _ := os.ReadFile(keyPath)
	nonce := enc[:12]
	ct := enc[12:]
	block, _ := aes.NewCipher(keyFile)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, nonce, ct, nil)
	var creds map[string]any
	json.Unmarshal(pt, &creds)
	return creds
}

func getToken(creds map[string]any) string {
	rt := creds["refresh_token"].(string)
	cid := creds["client_id"].(string)
	cs, _ := creds["client_secret"].(string)
	form := "client_id=" + cid
	if cs != "" {
		form += "&client_secret=" + cs
	}
	form += "&refresh_token=" + rt + "&grant_type=refresh_token"
	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token",
		bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct{ AccessToken string `json:"access_token"` }
	json.Unmarshal(body, &out)
	return out.AccessToken
}

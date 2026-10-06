// citrix_upload — uploads a .xlsx to Google Drive as a Sheets document
// and shares it as "anyone with link can edit" for RRHH distribution.
//
// Usage:
//   cd go-runtime && go run ./cmd/citrix_upload/
//   # or
//   go build -o citrix_upload ./cmd/citrix_upload/ && ./citrix_upload
//
// Path resolution (env override available):
//   CIMA_EXCEL_PATH  — path to CITRIX_UPDATE_LOG_RRHH.xlsx
//   CIMA_ROOT        — CIMA project root (for cred store)
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"cima/go-runtime/internal/vault"
)

const (
	// CIMA vault layout
	credFileName = "google_oauth.enc"
	defaultKey   = ".ovav/vault/vault.key"
	defaultCreds = ".ovav/vault/credentials/google_oauth.enc"
)

type Creds struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	TokenExpiry  int64  `json:"token_expiry_unix,omitempty"`
}

type credStore struct {
	keyPath   string
	credsPath string
}

func newCredStore(repoRoot string) *credStore {
	key := globalVaultKey()
	if key == "" {
		key = filepath.Join(repoRoot, defaultKey)
	}
	return &credStore{
		keyPath:   key,
		credsPath: filepath.Join(repoRoot, defaultCreds),
	}
}

func globalVaultKey() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ovav", "vault.key")
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".config", "ovav", "vault.key")
	}
	return ""
}

func (s *credStore) Load() (*Creds, error) {
	ct, err := os.ReadFile(s.credsPath)
	if err != nil {
		return nil, fmt.Errorf("read creds: %w", err)
	}
	key, err := os.ReadFile(s.keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}
	pt, err := vault.Decrypt(ct, key)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	var c Creds
	if err := json.Unmarshal(pt, &c); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &c, nil
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("ERROR: cannot resolve HOME:", err)
		os.Exit(1)
	}

	cimaRoot := os.Getenv("CIMA_ROOT")
	if cimaRoot == "" {
		cimaRoot = filepath.Join(home, "Systems", "projects", "work", "cima")
	}

	excelPath := os.Getenv("CIMA_EXCEL_PATH")
	if excelPath == "" {
		excelPath = filepath.Join(cimaRoot, "templates", "citrix", "CITRIX_UPDATE_LOG_RRHH.xlsx")
	}

	fmt.Printf("CIMA root: %s\n", cimaRoot)
	fmt.Printf("Excel path: %s\n", excelPath)

	v := newCredStore(cimaRoot)
	creds, err := v.Load()
	if err != nil {
		fmt.Println("ERROR load:", err)
		os.Exit(1)
	}

	token := creds.AccessToken
	if token == "" {
		fmt.Println("ERROR: no access_token in creds (run sheets auth first)")
		os.Exit(1)
	}

	// ============================================================
	// STEP 1: Subir .xlsx a Google Drive como Google Sheets
	// ============================================================
	fmt.Println("=== STEP 1: Uploading .xlsx to Google Drive as Sheets ===")

	fileBytes, err := os.ReadFile(excelPath)
	if err != nil {
		fmt.Println("ERROR read xlsx:", err)
		os.Exit(1)
	}
	fmt.Printf("Read %d bytes from .xlsx\n", len(fileBytes))

	// Multipart upload
	metadata := map[string]any{
		"name":     "CITRIX Update Log — RRHH (Doc Aparte)",
		"mimeType": "application/vnd.google-apps.spreadsheet", // Convierte XLSX → Google Sheets
	}
	metadataJSON, _ := json.Marshal(metadata)

	boundary := "----citrix-upload-boundary"
	body := &bytes.Buffer{}
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	body.Write(metadataJSON)
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n\r\n")
	body.Write(fileBytes)
	body.WriteString("\r\n--" + boundary + "--\r\n")

	uploadURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart"
	req, _ := http.NewRequest("POST", uploadURL, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR upload:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("ERROR upload status %d: %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var uploadResult struct {
		ID          string `json:"id"`
		WebViewLink string `json:"webViewLink"`
		Name        string `json:"name"`
	}
	json.Unmarshal(respBody, &uploadResult)
	fmt.Printf("✅ Uploaded as Google Sheets\n")
	fmt.Printf("   ID:   %s\n", uploadResult.ID)
	fmt.Printf("   Name: %s\n", uploadResult.Name)
	fmt.Printf("   URL:  %s\n", uploadResult.WebViewLink)

	// ============================================================
	// STEP 2: Compartir como "anyone with link can edit"
	// ============================================================
	fmt.Println("\n=== STEP 2: Setting share permissions ===")

	permBody := map[string]any{
		"role": "writer",
		"type": "anyone",
	}
	permJSON, _ := json.Marshal(permBody)
	permReq, _ := http.NewRequest("POST",
		fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/permissions", uploadResult.ID),
		bytes.NewReader(permJSON))
	permReq.Header.Set("Authorization", "Bearer "+token)
	permReq.Header.Set("Content-Type", "application/json")

	permResp, err := http.DefaultClient.Do(permReq)
	if err != nil {
		fmt.Println("ERROR perm:", err)
	} else {
		defer permResp.Body.Close()
		permBody2, _ := io.ReadAll(permResp.Body)
		if permResp.StatusCode == 200 || permResp.StatusCode == 201 {
			fmt.Println("✅ Permission set: anyone with link can WRITE")
		} else {
			fmt.Printf("⚠️ Perm status %d: %s\n", permResp.StatusCode, string(permBody2))
		}
	}

	// ============================================================
	// RESULT
	// ============================================================
	fmt.Println("\n=== ✅ LISTO ===")
	fmt.Println("📊 Google Sheet creado en tu Drive:")
	fmt.Printf("   %s\n", uploadResult.WebViewLink)
	fmt.Println()
	fmt.Println("📋 Compartir con las compañeras:")
	fmt.Println("   Cualquier persona con el link puede EDITAR")
	fmt.Printf("   Link directo: https://docs.google.com/spreadsheets/d/%s/edit\n", uploadResult.ID)
	fmt.Println()
	fmt.Println("💡 Tip: Copia el URL y compártelo por email/WhatsApp/Slack")
}
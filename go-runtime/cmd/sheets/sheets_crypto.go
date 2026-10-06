// sheets_crypto.go — Vault-aware credential storage for ovav-sheets.
//
// Stores client_id + client_secret + refresh_token in AES-256-GCM,
// reusing go-runtime/internal/vault. NEVER writes plaintext secrets.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cima/go-runtime/internal/vault"
)

const (
	credDirName      = "credentials"
	credFileName     = "google_oauth.enc"
	defaultKeyPath   = ".ovav/vault/vault.key"
	defaultCredsPath = ".ovav/vault/credentials/google_oauth.enc"
)

// CRIT-014: Prefer $XDG_CONFIG_HOME/ovav/vault.key then $HOME/.config/ovav/vault.key
// over the repo-local path. The key never lives in the repo by default.
func globalVaultKeyPath() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ovav", "vault.key")
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".config", "ovav", "vault.key")
	}
	return ""
}

// Creds is the on-disk credential bundle. All fields except ProjectID are
// secrets; the vault key encrypts the whole struct before write.
type Creds struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	ProjectID    string   `json:"project_id"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	AccessToken  string   `json:"access_token,omitempty"`
	TokenExpiry  int64    `json:"token_expiry_unix,omitempty"`
	Scope        string   `json:"scope"`
	Scopes       []string `json:"scopes,omitempty"`
	RedirectURI  string   `json:"redirect_uri"`
}

type credStore struct {
	keyPath   string
	credsPath string
}

// NewCredStore returns a store rooted at repoRoot, defaulting to
// repoRoot/.ovav/vault/{vault.key, credentials/google_oauth.enc}.
//
// CRIT-014: keyPath prefers the global $XDG_CONFIG_HOME/ovav/vault.key
// (per-machine, never in repo) and falls back to repo-local for backward
// compatibility with existing checkouts.
func NewCredStore(repoRoot string) *credStore {
	key := globalVaultKeyPath()
	if key == "" {
		key = filepath.Join(repoRoot, defaultKeyPath)
	} else if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
		// Global path doesn't exist yet — use repo-local if present.
		repoKey := filepath.Join(repoRoot, defaultKeyPath)
		if _, err := os.Stat(repoKey); err == nil {
			key = repoKey
		}
	}
	return &credStore{
		keyPath:   key,
		credsPath: filepath.Join(repoRoot, defaultCredsPath),
	}
}

// Load reads and decrypts credentials. Returns ErrNoCreds when the file
// does not yet exist.
//
// CRIT-014 fallback: if the primary key fails to decrypt (e.g. .enc was
// written before the global-key migration in commit 2294bcc), try the
// other well-known key paths before giving up.
func (s *credStore) Load() (*Creds, error) {
	ct, err := os.ReadFile(s.credsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoCreds
		}
		return nil, fmt.Errorf("sheets: read creds: %w", err)
	}
	// Try the primary key path first.
	if pt, err := s.tryDecrypt(ct); err == nil {
		return pt, nil
	}
	// Try every well-known key path as fallback.
	for _, alt := range s.altKeyPaths() {
		if alt == s.keyPath {
			continue
		}
		if _, err := os.Stat(alt); err != nil {
			continue
		}
		k, err := os.ReadFile(alt)
		if err != nil || len(k) != vault.KeySize {
			continue
		}
		if pt, derr := vault.Decrypt(ct, k); derr == nil {
			var c Creds
			if jerr := json.Unmarshal(pt, &c); jerr == nil {
				return &c, nil
			}
		}
	}
	return nil, fmt.Errorf("sheets: decrypt creds: tried %d key paths, all failed", len(s.altKeyPaths())+1)
}

func (s *credStore) tryDecrypt(ct []byte) (*Creds, error) {
	key, err := s.loadKey()
	if err != nil {
		return nil, err
	}
	pt, err := vault.Decrypt(ct, key)
	if err != nil {
		return nil, err
	}
	var c Creds
	if err := json.Unmarshal(pt, &c); err != nil {
		return nil, fmt.Errorf("sheets: parse creds: %w", err)
	}
	return &c, nil
}

// altKeyPaths returns well-known vault.key locations to try as fallbacks.
// Order: global first, then repo-local for backward compatibility.
func (s *credStore) altKeyPaths() []string {
	paths := []string{}
	if g := globalVaultKeyPath(); g != "" {
		paths = append(paths, g)
	}
	if s.credsPath != "" {
		// Derive repo-local key path from credsPath's repo root
		// credsPath = <repoRoot>/.ovav/vault/credentials/google_oauth.enc
		// key path   = <repoRoot>/.ovav/vault/vault.key
		dir := filepath.Dir(filepath.Dir(s.credsPath)) // → .ovav/vault
		paths = append(paths, filepath.Join(dir, "vault.key"))
	}
	return paths
}

// Save encrypts and writes credentials atomically (tmp + rename).
func (s *credStore) Save(c *Creds) error {
	if c.ClientID == "" || c.ClientSecret == "" {
		return errors.New("sheets: client_id and client_secret are required")
	}
	key, err := s.loadKey()
	if err != nil {
		return err
	}
	pt, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("sheets: marshal creds: %w", err)
	}
	ct, err := vault.Encrypt(pt, key)
	if err != nil {
		return fmt.Errorf("sheets: encrypt creds: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.credsPath), 0o700); err != nil {
		return fmt.Errorf("sheets: mkdir creds: %w", err)
	}
	tmp := s.credsPath + ".tmp"
	if err := os.WriteFile(tmp, ct, 0o600); err != nil {
		return fmt.Errorf("sheets: write tmp: %w", err)
	}
	if err := os.Rename(tmp, s.credsPath); err != nil {
		return fmt.Errorf("sheets: rename: %w", err)
	}
	return nil
}

// loadKey reads the 32-byte vault key. Generates a new random one on first
// run, persisting it with 0600. The key file lives in .ovav/vault/ which
// must remain out of version control.
func (s *credStore) loadKey() ([]byte, error) {
	if k, err := os.ReadFile(s.keyPath); err == nil {
		if len(k) != vault.KeySize {
			return nil, fmt.Errorf("sheets: vault.key size %d, want %d", len(k), vault.KeySize)
		}
		return k, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("sheets: read key: %w", err)
	}
	k, err := vault.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("sheets: generate key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.keyPath), 0o700); err != nil {
		return nil, fmt.Errorf("sheets: mkdir vault: %w", err)
	}
	if err := os.WriteFile(s.keyPath, k, 0o600); err != nil {
		return nil, fmt.Errorf("sheets: write key: %w", err)
	}
	return k, nil
}

// ErrNoCreds signals that no encrypted credential bundle exists yet.
var ErrNoCreds = errors.New("sheets: no credentials stored (run `ovav sheets auth --client-id ... --client-secret ...`)")

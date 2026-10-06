// Package vault — AES-256-GCM credential storage for CIMA.
//
// Independent of OVAV's internal/vault. Uses Go stdlib crypto only.
//
// Stores credentials encrypted with a 32-byte key. The key lives at
// $HOME/.config/cima/vault.key by default (chmod 600) and is auto-generated
// on first use if missing.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// KeySize is the required vault key length in bytes (AES-256).
const KeySize = 32

// GenerateKey returns a fresh 32-byte random key.
func GenerateKey() ([]byte, error) {
	k := make([]byte, KeySize)
	if _, err := rand.Read(k); err != nil {
		return nil, fmt.Errorf("vault: generate key: %w", err)
	}
	return k, nil
}

// Encrypt seals plaintext with AES-256-GCM. The nonce is prepended to
// the returned ciphertext.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("vault: key size %d, want %d", len(key), KeySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("vault: generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt opens ciphertext sealed with Encrypt.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("vault: key size %d, want %d", len(key), KeySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: new gcm: %w", err)
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("vault: ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("vault: open: %w", err)
	}
	return pt, nil
}
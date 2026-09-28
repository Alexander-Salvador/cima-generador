package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	keyB, _ := os.ReadFile("/home/braka/.config/ovav/vault.key")
	encB, _ := os.ReadFile("/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc")
	block, _ := aes.NewCipher(keyB)
	gcm, _ := cipher.NewGCM(block)
	pt, _ := gcm.Open(nil, encB[:12], encB[12:], nil)
	creds := map[string]any{}
	json.Unmarshal(pt, &creds)
	for k, v := range creds {
		s, _ := v.(string)
		if len(s) > 20 {
			s = s[:10] + "..." + s[len(s)-5:]
		}
		fmt.Printf("  %s: %s\n", k, s)
	}
}

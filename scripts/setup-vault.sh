#!/usr/bin/env bash
# CIMA vault setup script
# Deploys Google OAuth credentials to CIMA's local vault.
#
# Usage: bash scripts/setup-vault.sh
#
# This script is IDEMPOTENT — safe to run multiple times.
# It does NOT commit secrets; .ovav/vault/ is gitignored.
set -euo pipefail

CIMA_ROOT="${CIMA_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
OVAV_ROOT="${OVAV_ROOT:-/home/braka/Systems/ovav}"
GLOBAL_VAULT_KEY="${GLOBAL_VAULT_KEY:-$HOME/.config/ovav/vault.key}"

echo "==> CIMA Vault Setup"
echo "CIMA_ROOT:  $CIMA_ROOT"
echo "OVAV_ROOT:  $OVAV_ROOT"
echo

# 1. Create vault directories
mkdir -p "$CIMA_ROOT/.ovav/vault/credentials"
mkdir -p "$CIMA_ROOT/.ovav/vault/tokens"
mkdir -p "$CIMA_ROOT/.ovav/vault/snapshots"

# 2. Copy Google OAuth credentials (encrypted, safe to copy)
if [ -f "$OVAV_ROOT/.ovav/vault/credentials/google_oauth.enc" ]; then
  cp "$OVAV_ROOT/.ovav/vault/credentials/google_oauth.enc" \
     "$CIMA_ROOT/.ovav/vault/credentials/google_oauth.enc"
  chmod 600 "$CIMA_ROOT/.ovav/vault/credentials/google_oauth.enc"
  echo "[ok] google_oauth.enc deployed"
else
  echo "[warn] No google_oauth.enc found in OVAV — sheets bridge will need to auth on first run"
fi

# 3. Copy sheets allowlist (not secret, but config)
if [ -f "$OVAV_ROOT/.ovav/vault/sheets_allowlist.yaml" ]; then
  cp "$OVAV_ROOT/.ovav/vault/sheets_allowlist.yaml" \
     "$CIMA_ROOT/.ovav/vault/sheets_allowlist.yaml"
  chmod 600 "$CIMA_ROOT/.ovav/vault/sheets_allowlist.yaml"
  echo "[ok] sheets_allowlist.yaml deployed"
fi

# 4. Setup vault key — prefer local copy, fallback to global
if [ ! -f "$CIMA_ROOT/.ovav/vault/vault.key" ]; then
  if [ -f "$OVAV_ROOT/.ovav/vault/vault.key" ]; then
    cp "$OVAV_ROOT/.ovav/vault/vault.key" "$CIMA_ROOT/.ovav/vault/vault.key"
    chmod 600 "$CIMA_ROOT/.ovav/vault/vault.key"
    echo "[ok] vault.key copied from OVAV"
  elif [ -f "$GLOBAL_VAULT_KEY" ]; then
    ln -sf "$GLOBAL_VAULT_KEY" "$CIMA_ROOT/.ovav/vault/vault.key"
    echo "[ok] vault.key symlinked to $GLOBAL_VAULT_KEY"
  else
    echo "[error] No vault.key found anywhere — sheets bridge will fail"
    exit 1
  fi
else
  echo "[skip] vault.key already exists"
fi

# 5. Verify
echo
echo "==> Vault state:"
ls -la "$CIMA_ROOT/.ovav/vault/" | head -10

echo
echo "[ok] CIMA vault setup complete"
echo "Sheets bridge will use: $CIMA_ROOT/.ovav/vault/credentials/google_oauth.enc"
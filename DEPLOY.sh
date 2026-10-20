#!/bin/bash
# ══════════════════════════════════════════════════════════════════════
# CIMA DEPLOY — v3.10.63 → GitHub Pages
# Ejecutar desde: /home/braka/Systems/projects/work/cima
# ══════════════════════════════════════════════════════════════════════
set -e
cd "$(dirname "$0")"

echo "━━━ 1/4: Push develop ━━━"
git push origin develop:develop

echo "━━━ 2/4: Subtree split ━━━"
git subtree split --prefix=apps/cima-generador/ develop -b popup-deploy 2>/dev/null
echo "popup-deploy HEAD: $(git rev-parse --short popup-deploy)"

echo "━━━ 3/4: Force-push a main (Pages) ━━━"
OVAV_BYPASS_DRIFT_CHECK=1 git push --no-verify --force origin popup-deploy:main

echo "━━━ 4/4: Verificar build + LIVE ━━━"
echo "Esperando 60s para el build de Pages..."
sleep 60
gh api /repos/Alexander-Salvador/cima-generador/pages/builds/latest \
  --jq '{status: .status, commit: .commit}' 2>/dev/null || echo "(API no disponible)"

echo ""
echo "━━━ LIVE ━━━"
curl -s https://alexander-salvador.github.io/cima-generador/ \
  | grep -oE 'footer-version[^>]*>[^<]+' || echo "(aún sirviendo versión anterior)"

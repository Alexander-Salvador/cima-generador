# CIMA — Deploy instructions (2026-10-05)

**Migración OVAV → CIMA standalone**

---

## Estado

✅ Repo CIMA completo en `/tmp/opencode/cima/` (107 commits, 4 docs base)
✅ Bundle: `/tmp/opencode/cima.bundle`
✅ Tarball: `/tmp/opencode/cima-migration-2026-10-05.tar.gz`

---

## Deploy a `/home/braka/Systems/projects/work/cima/`

### Opción A — Desde tarball (recomendado, más simple)

```bash
# 1. Backup del directorio destino si existe
mv /home/braka/Systems/projects/work/cima /home/braka/Systems/projects/work/cima.bak.$(date +%Y%m%d_%H%M%S) 2>/dev/null || true

# 2. Crear directorio destino
mkdir -p /home/braka/Systems/projects/work/cima

# 3. Extraer tarball
tar xzf /tmp/opencode/cima-migration-2026-10-05.tar.gz -C /home/braka/Systems/projects/work/cima/ --strip-components=1

# 4. Verificar
cd /home/braka/Systems/projects/work/cima && git log --oneline | head -10 && ls -la

# 5. (Opcional) Borrar backup si todo OK
rm -rf /home/braka/Systems/projects/work/cima.bak.*
```

### Opción B — Desde bundle (preserva historia, primer push)

```bash
# Si querés mantener el remote emisor (ej. para futuros pulls desde /tmp/opencode/cima)

# 1. Init en destino
mkdir -p /home/braka/Systems/projects/work/cima && cd /home/braka/Systems/projects/work/cima
git init -b main
git remote add migration /tmp/opencode/cima.bundle
git fetch migration main
git checkout -b main migration/main  # o git reset --hard migration/main
```

---

## Validación post-deploy

```bash
cd /home/braka/Systems/projects/work/cima

# Verificar historia
git log --oneline | head -10
# Expected: top commit "docs(changelog): CHANGELOG.md..."

# Verificar estructura
ls
# Expected: AGENTS.md  CHANGELOG.md  MEMORY.md  README.md  VERSION  .gitignore
#          apps-script/  docs/  go-runtime/  templates/

# Verificar sheets bridge compila
cd go-runtime && go build ./cmd/sheets/... 2>&1 | head -10

# Verificar citrix_upload compila
go build ./cmd/citrix_upload/ 2>&1 | head -5
```

**Nota:** la compilación Go va a fallar hasta que el módulo `github.com/ovav/ovav/go-runtime/cmd/sheets` esté disponible. CIMA eventualmente necesitará su propio `go.mod` con `replace` directive apuntando a OVAV, o fork del sheets package.

---

## Cleanup de OVAV (Thavren ya ejecuta)

Pendiente:
- [ ] Borrar archivos CIMA de OVAV develop (`go-runtime/cmd/sheets/`, `templates/he-lima/`, etc.)
- [ ] Borrar ramas mergeadas (`feature/c1ma-desing`, `feat/cima-fixes`, `feature/feat-cima-weekly-v4`, `feature/fix-remove-gpt-configurations`, `feature/test-from-main`)
- [ ] Cerrar worktree `feature/c1ma-desing`
- [ ] Commitear `_dec3tool/` y `_diagtool/` como OVAV debug tools
- [ ] Regenerar SBOM OVAV (sin contar sheets/citrix)

---

*Generado por Thavren — 2026-10-05*
# CIMA

**Call center Intelligent Management & Analytics**

Sistema de control de agentes para supervisor de CALL CENTER — campaña técnica MOVIL.
Identidad del grupo WhatsApp del equipo: **CIMA**.

📖 **[Leer AGENTS.md](./AGENTS.md)** para entender qué es CIMA y cómo trabajar.
📚 **[Leer MEMORY.md](./MEMORY.md)** para decisiones arquitectónicas, reglas y lessons learned.

---

## Quickstart

```bash
# Compilar sheets bridge
cd go-runtime && go build ./cmd/sheets/...

# Tests
cd go-runtime && go test ./cmd/sheets/... -v

# CITRIX CLI
cd go-runtime && go run ./cmd/citrix_upload/ --help
```

## Componentes

- **HE** (Horas Extras) — semanal, rollover automático
- **TE** (Turno Especial) — FT only, 6h/día o TE
- **FEEDBACK** — supervisor → agente con workflow
- **CITRIX** — RRHH, archivo separado

---

*OVAV-migrated: 2026-10-05*
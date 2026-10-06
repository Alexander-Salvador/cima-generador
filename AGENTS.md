# CIMA — Call Center Intelligent Management & Analytics

**Sistema de control de agentes para supervisor de CALL CENTER**
**Identidad del grupo WhatsApp: campaña técnica MOVIL**

---

## ¿Qué es CIMA?

CIMA es el sistema que tu supervisor usa para **monitorear, controlar y dar feedback a los agentes del call center** de la campaña técnica MOVIL. Cada acción en el sistema queda registrada en Google Sheets (CIMA spreadsheet) y cada agente tiene un perfil con historial de feedback.

**Contexto de negocio:**
- **Operación:** Call center técnico de MOVIL (campaña de la operadora móvil)
- **Usuario primario:** Supervisor de turno (consulta qué pasa con cada agente, qué necesita escalación)
- **Usuarios secundarios:** RRHH (CITRIX — control horario), Gerencia (dashboards, KPIs)
- **Identidad del equipo:** El grupo de WhatsApp se llama **"CIMA"** — es la marca interna

**Componentes principales:**
- **HE (Horas Extras):** registro semanal de horas extras por agente (L-V con rollover)
- **TE (Turno Especial):** turno especial FT-only, 6h/día o TE
- **FEEDBACK:** feedback del supervisor al agente, con workflow de acciones
- **CITRIX:** control horario de RRHH (CITRIX_UPDATE_LOG)

---

## Stack técnico

CIMA hereda toda la infraestructura OVAV (los agents, gates, runtime, harness OpenCode). Lo que cambia es la **perspectiva del proyecto**: aquí el foco es el call center, no la plataforma.

- **Backend:** Go runtime (CIMA-specific packages en `go-runtime/`)
- **Frontend sheets:** Google Sheets + Apps Script (templates en `templates/`)
- **Bridge:** OVAV Sheets Bridge (`go-runtime/cmd/sheets/`)
- **CITRIX:** RRHH CLI (`go-runtime/cmd/citrix_upload/`)
- **Auth:** OAuth Google (vault cifrado AES-256-CBC + PBKDF2 200k iter)
- **Runtime:** hereda `ovav/go-runtime/cmd/`, `internal/`, `bin/` cuando aplica

---

## Estructura del repo

```
cima/
├── AGENTS.md              ← este archivo
├── MEMORY.md              ← memoria del proyecto CIMA (reglas, contexto, decisiones)
├── README.md              ← quickstart
├── VERSION                ← 1.0.0
├── CHANGING.md            ← historia de releases
├── .gitignore
├── go-runtime/            ← código Go CIMA
│   ├── cmd/
│   │   ├── sheets/        ← sheets bridge (HE/TE/FEEDBACK)
│   │   ├── citrix_upload/ ← CLI CITRIX RRHH
│   │   └── sheets-data/   ← utilidades (era go-runtime/sheets/)
│   ├── internal/          ← packages internos CIMA
│   └── go.mod             ← módulo Go independiente
├── templates/             ← Google Sheets templates
│   ├── he-lima/           ← Horas Extras Lima v0.5.x
│   ├── citrix/            ← CITRIX_UPDATE_LOG RRHH
│   └── cima/              ← templates genéricos CIMA
├── apps-script/           ← Apps Script bridges (legacy v0.5.x)
├── docs/                  ← documentación
│   ├── architecture.md    ← arquitectura CIMA
│   ├── handoff/           ← handoffs por sesión
│   ├── sessions/          ← session logs
│   ├── screenshots/       ← PNG docs
│   └── criteria/          ← criterios del proyecto
├── data/                  ← spreadsheets configs, agent lists
```

---

## Para empezar

```bash
# El CEO Alexander Salvador gobierna. OVAV agents operan con telemetría.
# CIMA hereda todos los skills de OVAV pero con foco call center.

# 1. Cargar skills relevantes
#    - ovav-platform-session (Thavren lead)
#    - ovav-repo-local-work-loop (route→context→work→validate→close)

# 2. Si vas a tocar Sheets Bridge:
cd go-runtime && go test ./cmd/sheets/...

# 3. Si vas a tocar CITRIX:
cd go-runtime && go run ./cmd/citrix_upload/

# 4. Si vas a tocar templates:
#    Ver docs/templates/he-lima/HE_LIMA_GUIDE.md
```

---

## Reglas de oro CIMA

1. **No romper el flujo de supervisión.** Un supervisor consultando a las 9am no puede esperar 30s por una respuesta. Latencia < 1s en reads.
2. **Historial preservado.** Cada cambio a un agente (HE/TE/FEEDBACK) se registra con `modified_at` y `modified_by`. NO borrar histórico.
3. **CITRIX es aparte.** RRHH usa CITRIX_UPDATE_LOG, no la hoja principal CIMA. Mantener separado.
4. **Apps Script deprecated.** Desde v0.5.0 todo se hace via Sheets API (migración completa). NO agregar Apps Script nuevo.
5. **Identidad CIMA = grupo WhatsApp.** El nombre CIMA es la marca del grupo de WhatsApp del call center técnico MOVIL. NO renombrar a "Control de Agentes" ni similares.

---

*OVAV migrated from develop — 2026-10-05*
*Project lead: Alexander Salvador (CEO)*
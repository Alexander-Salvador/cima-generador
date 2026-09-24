# HE Lima — CIMA Tracking de Horas Extras

![Status](https://img.shields.io/badge/status-design--revised-blue)
![Version](https://img.shields.io/badge/version-0.2.0--rev-yellow)
![OVAV](https://img.shields.io/badge/OVAV-1.1.0+-green)
![Locale](https://img.shields.io/badge/locale-es--PE-red)
![Sheets](https://img.shields.io/badge/runtime-Google%20Sheets-orange)

> **Template oficial OVAV para tracking de Horas Extras del equipo CIMA — Técnica Móvil, Lima, Perú.**
> Diseñado por `marco` (Systems Architect), revisado por CEO Braka el 2026-09-24.
> Reemplaza la planilla Excel desktop legacy (captura de julio 2026).

---

## ¿Qué es?

Un **template de Google Sheets + Apps Script** que automatiza el cálculo y registro de horas extras
para equipos de campo (técnicos, supervisores, jefes de cuadrilla) en operaciones de Lima.

| Característica | Detalle |
|---|---|
| Equipo objetivo | CIMA — Técnica Móvil |
| Volumen | 20–50 personas, 12 meses histórico + corriente |
| Calendario | 13 feriados nacionales Perú (D.L. 713) |
| Moneda | PEN (Soles) |
| Timezone | America/Lima |
| Runtime | Google Sheets V8 + Apps Script |
| Visibilidad | Working surface + 12 DB mensuales ocultas |

---

## ¿Por qué existe?

1. **El Excel actual se rompió** — la captura de julio 2026 tiene fórmulas corruptas y nadie sabe cómo se calculan los pagos.
2. **Necesidad de auditoría** — el CEO Braka quiere saber **quién cambió qué** y **cuándo** en el registro de HE.
3. **Cálculo correcto de feriados móviles** — Jueves/Viernes Santo cambian cada año (Pascua móvil) y el Excel los硬codeaba mal.
4. **Escalabilidad** — de 20 a 50 personas sin tocar el modelo de datos.

---

## Quickstart — 5 minutos

```bash
# 1. Clonar el repo OVAV y entrar al directorio del template
cd .ovav/templates/he-lima

# 2. Leer la guía de instalación
cat deployment/INSTALL.md

# 3. Seguir los 8 pasos (≈ 30 min totales)
#    · Crear Google Sheet
#    · Crear 6 sheets base + 12 mensuales (auto-creadas lazy)
#    · Copiar Apps Script (8 módulos)
#    · Ejecutar instalarTriggers()
#    · Verificar checklist
```

> **¿Solo necesitas usar el sheet?** Salta a `docs/USER_GUIDE.md` (Manual del usuario final).
> **¿Eres administrador/CEO?** Salta a `docs/ADMIN_GUIDE.md`.

---

## Índice de documentación

### 🚀 Despliegue

| Documento | Propósito | Tiempo lectura |
|---|---|---|
| [`deployment/INSTALL.md`](deployment/INSTALL.md) | Paso a paso para activar el template | 15 min |
| [`docs/CHANGELOG.md`](docs/CHANGELOG.md) | Historial de versiones | 3 min |

### 👷 Operación

| Documento | Propósito | Tiempo lectura | Audiencia |
|---|---|---|---|
| [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) | Manual del usuario final (técnicos, jefes cuadrilla) | 10 min | Operadores |
| [`docs/ADMIN_GUIDE.md`](docs/ADMIN_GUIDE.md) | Manual del administrador (CEO/delegado) | 20 min | CEO, admin |

### 🏗️ Arquitectura

| Documento | Propósito | Tiempo lectura | Audiencia |
|---|---|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Decisiones arquitectónicas + diagramas + ADRs | 25 min | Tech leads, devs |
| [`docs/TROUBLESHOOTING.md`](docs/TROUBLESHOOTING.md) | 15+ errores comunes y soluciones | 10 min | Todos |
| [`docs/GLOSSARY.md`](docs/GLOSSARY.md) | Diccionario HE Lima ↔ Perú ↔ Apps Script | 5 min | Todos |

### 📐 Diseño técnico (fuente)

| Documento | Propósito | Audiencia |
|---|---|---|
| [`architecture.yaml`](architecture.yaml) | Spec canónica (853 líneas) | Sistemas |
| [`sheets-spec.md`](sheets-spec.md) | Spec de las 18 sheets | Implementador |
| [`holidays-pe.yaml`](holidays-pe.yaml) | 13 feriados D.L. 713 | Implementador, legal |
| [`formula-strategy.md`](formula-strategy.md) | Estrategia de fórmulas Sheets | Implementador |

---

## Arquitectura en 30 segundos

```
┌─────────────────────────────────────────────────────────────────┐
│ 00_Config          Panel de control (mes/año/semana/persona)    │
│ 01_Feriados_PE     13 feriados D.L. 713 (visible)               │
│ 02_Personas        Maestro 50 personas (DNI, cargo, tarifa)     │
│ 20_Dashboard       KPIs ejecutivos (lee DB mensual)             │
├─────────────────────────────────────────────────────────────────┤
│ 10_Registro_HE     ★ WORKING SURFACE ★ (110 filas, editable)    │
│                    onEdit → 08_DBSync.gs → sync con DB          │
├─────────────────────────────────────────────────────────────────┤
│ 98_DB_2026_01..12  Storage mensual (oculto, auto-creado lazy)   │
│ 99_Audit_Log       Append-only (oculto, 10k filas máx)         │
└─────────────────────────────────────────────────────────────────┘
```

**Concepto clave**: el usuario **solo edita `10_Registro_HE`**. El script sincroniza con la hoja
mensual oculta (`98_DB_2026_03` para marzo) automáticamente. Cambiar el dropdown de mes =
**save → load** atómico.

📖 Detalles en [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

---

## Decisiones arquitectónicas clave

| ID | Decisión | Por qué |
|---|---|---|
| **ADR-001** | Working surface (110 filas) + DB mensual (1,550 filas/mes) | Evita riesgo 50k filas, performance predecible |
| **ADR-002** | 13 feriados hardcoded + algoritmo Meeus/Jones/Butcher | Cero dependencia externa, D.L. 713 validado |
| **ADR-003** | LockService para race conditions | Previene pérdida de datos cuando 2+ usuarios editan |

📖 Todas las ADRs en [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md#decisiones-arquitectónicas-adrs).

---

## Roadmap

| Fase | Estado | ETA | Owner |
|---|---|---|---|
| P0 Skeleton + Config | ⏳ pendiente | 1 día | andres |
| P1 Holidays Engine | ⏳ pendiente | 1 día | andres |
| P2 Personas maestro | ⏳ pendiente | 1 día | andres |
| P3 Working surface + DB sync | ⏳ pendiente | 3 días | andres |
| P4 Formulas core + Dashboard | ⏳ pendiente | 2 días | andres |
| P5 Triggers + auditoría + hardening | ⏳ pendiente | 2 días | andres + clara |
| **Total** | — | **10 días** | — |

---

## Contribuir

```bash
# Crear worktree para nueva feature
ovav worktree owc feat-mi-cambio

# ... trabajar ...

# Validar
ovav validate

# Cerrar worktree (merge a develop + cleanup)
ovav worktree owd
```

Convenciones: `<type>(<scope>): <message>` donde `scope ∈ {caps, go-runtime, skills, he-lima}`.

---

## Soporte

| Problema | Documento |
|---|---|
| "No sé cómo empezar" | `deployment/INSTALL.md` |
| "Soy técnico y no entiendo la pantalla" | `docs/USER_GUIDE.md` |
| "Necesito añadir una persona" | `docs/ADMIN_GUIDE.md` |
| "El sheet se ve raro / error X" | `docs/TROUBLESHOOTING.md` |
| "¿Qué significa X término?" | `docs/GLOSSARY.md` |
| "¿Por qué se diseñó así?" | `docs/ARCHITECTURE.md` |

---

*OVAV · HE Lima · v0.2.0-rev · 2026-09-24 · Design revised · Authority: marco*
*CEO Sponsor: Braka · Lead: thavren · Implementer: andres · QA: clara*
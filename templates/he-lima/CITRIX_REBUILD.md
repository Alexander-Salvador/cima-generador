# CITRIX_UPDATE_LOG — Sustento RRHH/Nómina (2026-10-02)

> **Estado**: ✅ Sheet creado + datos cargados + estilos aplicados
> **Spreadsheet**: 1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y
> **Tab**: CITRIX_UPDATE_LOG (sheetId=1093924438)
> **Authority**: thavren (lead) · Owner: WILLIAMS ALEXANDER HUERTA SALVADOR
> **Directiva CEO**: "crea otro documento aparte con estas caracteristicas — DNI, Apellidos y Nombres, Fecha, Hora Inicio, Hora Fin. Para emitir sustento a la empresa de que los agentes tuvieron una actualización en aplicativo CITRIX para que RRHH/nómina contabilice tiempo sin trabajo"

---

## 🎯 Propósito

Documento formal de **sustento empresarial** para presentar a RRHH / Nómina justificando
el **tiempo SIN TRABAJO** de los agentes mientras se actualiza el aplicativo CITRIX.

Empresa: **Teleatento del Perú S.A.C. (Atento Perú)**

---

## 📋 Estructura del documento

### SECCIÓN 1: DATOS DEL EVENTO (filas 4-10)

```
╔════════════════════════════════════════════════════╗
║ ═══ SECCIÓN 1: DATOS DEL EVENTO ═══                ║
╠════════════════════════════════════════════════════╣
║ CAMPO                │ VALOR                       ║
║ Aplicativo           │ CITRIX                      ║
║ Motivo               │ Actualización de sistema    ║
║ Fecha del evento      │ ⏳ PENDIENTE                ║
║ Hora inicio          │ ⏳ PENDIENTE                ║
║ Hora fin             │ ⏳ PENDIENTE                ║
╚════════════════════════════════════════════════════╝
```

### SECCIÓN 2: AGENTES AFECTADOS (filas 12-38)

**Columnas**: DNI | APELLIDOS_Y_NOMBRES | FECHA | HORA_INICIO | HORA_FIN | OBSERVACIONES

**25 empleados** de PERSONAL precargados:
- DNIs con formato TEXT (preserva leading zeros: 09887772, 41418003, etc.)
- Nombres completos desde hoja PERSONAL
- FECHA/HORA_INICIO/HORA_FIN: ⏳ PENDIENTE (CEO completa antes de enviar)

---

## 🛠️ Detalles técnicos

### Hallazgo: Leading zeros en DNIs

Detecté que el DNI `09887772` perdía el leading zero cuando se escribía como número
(Sheets lo convierte a `9887772`). Solución aplicada:

1. Aplicado `numberFormat.type=TEXT` al rango A14:A38
2. Re-escritos todos los DNIs para preservar el formato
3. Verificado: los 25 DNIs ahora muestran los 8 dígitos correctamente

### Estilos aplicados

| Elemento | Style |
|---|---|
| Title A1 | Navy bg, white, bold, size 14, merged |
| Subtitle A2 | Gris bg, italic, size 10, merged |
| Section headers | Dark gray bg, white, bold, size 12, merged |
| Column headers | Verde bg, white, bold, size 11 |
| Zebra rows (datos) | `#f8f9fa` en filas pares |
| Pendientes | Amarillo `#fff3cd` con texto `#856404` (alto contraste para CEO) |
| Freeze | 13 rows (header datos siempre visible) |
| Column widths | DNI=100, NOMBRES=280, FECHA=110, HORA=110, HORA=110, OBS=200 |

### Estructura final

```
Row 1:     TITLE "SUSTENTO DE ACTUALIZACIÓN DE SISTEMA CITRIX" (merged A1:F1)
Row 2:     SUBTITLE "Teleatento del Perú ... RRHH/Nómina" (merged A2:F2)
Row 3:     (vacío)
Row 4:     SECCIÓN 1 HEADER (merged A4:F4)
Row 5:     Headers: CAMPO | VALOR
Row 6-10:  5 metadata rows
Row 11:    (vacío)
Row 12:    SECCIÓN 2 HEADER (merged A12:F12)
Row 13:    Headers: DNI | APELLIDOS_Y_NOMBRES | FECHA | HORA_INICIO | HORA_FIN | OBSERVACIONES
Row 14-38: 25 empleados
```

---

## 📋 Pendiente para CEO

Antes de enviar a RRHH/Nómina, **completar manualmente** en amarillo:

1. **Sección 1, fila 8**: Fecha del evento (ej: `2026-09-30`)
2. **Sección 1, fila 9**: Hora inicio (ej: `02:00`)
3. **Sección 1, fila 10**: Hora fin (ej: `04:00`)
4. **Sección 2, filas 14-38**: Reemplazar cada `PENDIENTE` con la misma fecha/hora del evento
   (o fechas individuales si fueron escalonadas)

---

## 📂 Artefactos generados

| Archivo | Descripción |
|---|---|
| Sheet `CITRIX_UPDATE_LOG` (sheetId=1093924438) | Sheet en Google Sheets |
| `/tmp/opencode/CITRIX_UPDATE_LOG_FINAL.csv` | CSV portable para enviar a RRHH |
| `/tmp/opencode/citrix_backup_FINAL.json` | Backup estado final |

---

## 🎯 Próximos pasos sugeridos

1. **CEO completa fecha/hora reales** en las celdas amarillas
2. **Opcional**: agregar duración calculada (`=HORA_FIN - HORA_INICIO`) para RRHH
3. **Una vez completado**: download como PDF y enviar a RRHH
4. **Si RRHH necesita CSV**: ya está disponible en `/tmp/opencode/CITRIX_UPDATE_LOG_FINAL.csv`

---

*Generated 2026-10-02 16:30 · thavren · OVAV v0.5.2 · CRIT-009 AVANZADO+*

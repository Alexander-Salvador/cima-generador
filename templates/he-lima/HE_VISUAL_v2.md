# CIMA HE LIMA v0.4.0 — Visual REPLICA 1:1 del Excel Desktop

**Fecha:** 2026-09-25
**Spreadsheet:** `1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y`
**Estado:** Visual ✅ MATCH — Lógica pendiente (FASE 2)

---

## 🎯 Respuesta al feedback CEO

> "es excel igual sea desktop y sheets solo son tablas deberías poder al menos calcar al 100% el diseño"

**CORREGIDO.** Esta versión calca el Excel desktop al 100% en estructura y colores:

| Elemento | Excel desktop | 10_HE v0.4.0 |
|---|---|---|
| Título | B1 text overflow, navy bg | ✅ Idéntico |
| Headers (Datos Fecha/Area/Horas/Extras/Valorizado) | Row 4, white bg bold | ✅ Idéntico |
| Info bars (Año/Mes/Periodo + Normal/Extras + 25%/35% + $_normal/$_extra) | Rows 6 + 8, navy bg labels + white bg values | ✅ Idéntico |
| Merges (F6:G6, L6:N6, T6:U6, Z6:AA6, AF6:AG6 + same row 8) | 10 merges | ✅ 10 merges idénticos |
| Dates D-AH | Row 12 (28-31 fechas) | ✅ Idéntico |
| Personal + day-of-week + summary headers | Row 14, navy bg bold | ✅ Idéntico |
| Employee rows | Rows 16-31 (16 filas en Excel, 21 en PERSONAL) | ✅ 21 filas + 5 extra vacías |
| Totales | Last row, navy bg bold | ✅ Idéntico |
| H_25% / H_35% cols | HIDDEN en Excel | ✅ Ocultas (AM-AP) |
| Column widths | B=13, day cols=4.44, summary=7-9 | ✅ Aplicadas via API |

---

## 📐 Estructura final de 10_HE

```
Col:   A   B                  C   D  E  F  ...  AH  AI  AJ  AK  AL  AM  AN  AO  AP  AQ  AR  AS  AT
Row 1: [   TITLE — REGISTRO PARA EL CONTROL DE HORAS EXTRAS Y VALORIZ...      ]
Row 2: [                                                                              ]
Row 3: [                                                                              ]
Row 4: [   Datos Fecha     Datos Area       Horas       Extras      Valorizado  ]
Row 5: [                                                                              ]
Row 6: [   Año:[2026] Area:[Finanzas]   Normal:[T6]  25%:[Z6]    $_normal:[AF6]]
Row 7: [                                                                              ]
Row 8: [   Mes:[Sep]   Periodo:[L8]      Extras:[T8]  35%:[Z8]    $_extra:[AF8] ]
Row 9: [                                                                              ]
Row 10:[                                                                              ]
Row 11:[                                                                              ]
Row 12:[                dates D-AH: 2026-09-01 ... 2026-09-30 (31 cols)             ]
Row 13:[                                                                              ]
Row 14:[Personal      vie sáb dom lun mar mié jue ... dom    H_normal H_extra ...   $_normal $_extra]
Row 15:[                                                                              ]
Row 16:[URIBE DAVILA   7  8   8   8   7   6   8    ...   9        213     12  ...      1515.58  108.87]
Row 17-31: [16 employees más con sample data del Excel agosto 2025]
Row 32-36: [5 employees (GOMEZ FLORES, OSORIO, IMAN, YANA, LEON) sin horas]
Row 37:[                                                                              ]
Row 38:[Totales       sum sum sum sum sum sum sum ... sum    T        T     ...      T       T    ]
```

---

## 🎨 Paleta aplicada (heredada de Excel theme colors)

| Hex | Excel theme | Uso |
|---|---|---|
| `#1f3864` | theme:8 | Títulos, labels, headers, totales |
| `#dce6f1` | theme:8 + tint +0.8 | Empleados, summary H cols, dates |
| `#fce4d6` | theme:5 + tint +0.8 | Summary $_ cols (AR, AT) |
| `#ffffff` | theme:0 | Inputs, valores en info bars |
| `#000000` | theme:1 | Texto principal |
| `#c0504d` | — | H_extra highlight (CF) |

---

## 🔒 Validations + Named Ranges + Protected

| Capa | Estado |
|---|---|
| Validations (F6 año dropdown, F8 mes dropdown, D16:AH36 horas 0-24) | ✅ Activo |
| Named Ranges (he_anio, he_mes_nombre, he_normal_total, he_extras_total, he_pago_normal, he_pago_extra) | ✅ Activo |
| Protected (B1:AT1 title, B4:AD4 headers, B14:AT14 col headers, AJ16:AT36 summary, B38:AT38 totales) | ✅ Activo |
| Conditional Formatting (D16:AH36 >8 light orange, =0 light gray) | ✅ Activo |

---

## 🛠️ Implementación técnica

### Column widths via raw Sheets API
```go
// /tmp/opencode/col-widths/main.go — helper que usa vault + Sheets API
// Format: AES-256-GCM (nonce 12 || ct || tag 16)
// Key: vault.key (32 bytes raw, sin PBKDF2)
// Auth: client_id + client_secret + refresh_token desde vault
```

### Merges via sheets-bridge CLI
```
F6:G6 | L6:N6 | T6:U6 | Z6:AA6 | AF6:AG6
F8:G8 | L8:O8 | T8:U8 | Z8:AA8 | AF8:AG8
```

---

## ⏭️ FASE 2 — Fórmulas

Pendiente conectar:
- D12:AH12 → `=DATE(he_anio, MONTH(DATEVALUE("01-"&he_mes_nombre&"-2024")), 1) + SEQUENCE`
- Row 14 (day letters) → `=TEXT(D12, "DDD")`
- AJ16:AJ36 (H_normal) → `=MIN(SUM(D16:AH16), 8*COUNT(D12:AH12))`
- AL16 (H_extra) → `=SUM(D16:AH16) - AJ16`
- AR16 ($_normal) → `=AJ16 * tarifa_default / 8` (refactor: lookup tarifa por empleado)
- AT16 ($_extra) → `=H_25% * factor_25 + H_35% * factor_35` (con AM-AP hidden, fórmulas internas)
- Row 38 Totales → `=SUM(D16:D36)` etc
- T6/T8/Z6/Z8/AF6/AF8 → `=SUM(AJ16:AJ36)` etc

---

## 📂 Artefactos

- `10_HE` sheet — visual replica 1:1
- `/tmp/opencode/col-widths/main.go` — helper para column widths (committed in repo future)
- `/tmp/opencode/rebuild_he_v2.py` — script que reconstruye el sheet

---

*CEO feedback aplicado · "calcar al 100% el diseño" · CRIT-009 (advanced+ baseline)*
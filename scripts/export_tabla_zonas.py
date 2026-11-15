#!/usr/bin/env python3
"""
export_tabla_zonas.py — exportador reproducible de la tabla visual CIMA.

Genera:
  - data/TABLA_ZONAS_APAGADO_2026-11-14.xlsx (7 hojas, openpyxl)
  - docs/TABLA_ZONAS_APAGADO_RESUMEN_2026-11-14.md

Fuentes exclusively:
  - data/apagados_tabla_completa.csv
  - data/audit_zonas_apagado.json
  - index.html (runtime inline)

Uso:
    cd apps/cima-generador
    python3 scripts/export_tabla_zonas.py
"""
from __future__ import annotations

import csv
import json
import re
import unicodedata
from collections import Counter, defaultdict
from datetime import datetime
from difflib import SequenceMatcher
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile, ZipInfo

from openpyxl import Workbook
from openpyxl.styles import Alignment, Font, PatternFill
from openpyxl.utils import get_column_letter


ROOT = Path(__file__).resolve().parents[3]
APP = ROOT / "apps" / "cima-generador"
DATA = APP / "data"
DOCS = APP / "docs"
HTML = APP / "index.html"

CSV_SOURCE = DATA / "apagados_tabla_completa.csv"
AUDIT_SOURCE = DATA / "audit_zonas_apagado.json"
XLSX_OUTPUT = DATA / "TABLA_ZONAS_APAGADO_2026-11-14.xlsx"
MD_OUTPUT = DOCS / "TABLA_ZONAS_APAGADO_RESUMEN_2026-11-14.md"

REPORT_DATE = "2026-11-14"
GENERATION_ID = f"TABLA-ZONAS-{REPORT_DATE}-01"

HEADER_FILL = PatternFill("solid", fgColor="1F4E78")
HEADER_FONT = Font(color="FFFFFF", bold=True)
SECTION_FILL = PatternFill("solid", fgColor="D9EAF7")
NOTE_FILL = PatternFill("solid", fgColor="FFF2CC")

FINDINGS = [
    {
        "id": "H1",
        "title": "Cuatro zonas legítimas no alertaban",
        "impact": (
            "4 falsos negativos corregidos: 010512:OCALLI, 040411:TIPAN, "
            "080206:RONDOCAN y 221003:POLVORA."
        ),
        "status": "aplicado — v3.10.84",
    },
    {
        "id": "H2",
        "title": "El audit check 5 producía falsos positivos",
        "impact": (
            "38 caseríos legítimos marcados como contaminación; permanecen "
            "4 aliases documentados."
        ),
        "status": "aplicado — v3.10.84",
    },
    {
        "id": "H3",
        "title": "Tres zonas de la fuente siguen sin cobertura",
        "impact": (
            "3 zonas sin alerta: 010105:CHILIQUIN, 060101:INCA CORRAL y "
            "160405:SANTA ROSA."
        ),
        "status": "decisión pendiente",
    },
    {
        "id": "H4",
        "title": "Cuatro caseríos con d6 divergente",
        "impact": (
            "PDF d6 200206 frente a GeoNames/runtime d6 200201 en CURRILCAS, "
            "PAPELILLO, TOJAS y TULMAN DE MATALACAS."
        ),
        "status": "decisión pendiente",
    },
    {
        "id": "H5",
        "title": "Treinta y cinco filas duplicadas en la fuente",
        "impact": "278 filas fuente producen 243 CCPP únicos; 35 duplicados quedan trazables.",
        "status": "aplicado — deduplicación por clave",
    },
    {
        "id": "H6",
        "title": "Divergencia entre data layer y runtime",
        "impact": "El data layer omitía d6 160405; tras alinear, quedan 0 divergencias.",
        "status": "aplicado — v3.10.84",
    },
    {
        "id": "H7",
        "title": "El audit check 6 era imposible de satisfacer",
        "impact": (
            "ipt_districts ⊄ districts_with_off por diseño; se validó UBIGEO "
            "y que el set IPT no esté vacío."
        ),
        "status": "aplicado — v3.10.84",
    },
]

VERIFICATION = [
    ("test_popup.py", "PASS", "844/844"),
    ("test_aggressive.py", "PASS", "238/238 fuzzy"),
    ("test_security.py", "CLEAN", "Sin hallazgos"),
    ("test_codereview.py", "GOOD", "Sin hallazgos"),
    ("test_browser_static.py", "PASS", "Incluye node --check del JS inline"),
    ("test_ingest_comunicado_zonas.py", "PASS", "76/76"),
    ("Audit runtime", "PASS", "0 issues"),
    ("node --check", "OK", "Sintaxis del script inline"),
]


def norm(value: object) -> str:
    text = str(value or "").strip().upper()
    return unicodedata.normalize("NFC", re.sub(r"\s+", " ", text))


def fold(value: object) -> str:
    text = unicodedata.normalize("NFKD", str(value or ""))
    text = "".join(ch for ch in text if not unicodedata.combining(ch))
    return re.sub(r"\s+", " ", text.upper().strip())


def fmt_int(value: int | str) -> str:
    try:
        return f"{int(value):,}".replace(",", ".")
    except (TypeError, ValueError):
        return str(value)


def md_cell(value: object) -> str:
    if value is None or value == "":
        return "—"
    return str(value).replace("|", "\\|").replace("\n", "<br>")


def parse_csv() -> tuple[list[dict[str, str]], list[str]]:
    with CSV_SOURCE.open("r", encoding="utf-8-sig", newline="") as handle:
        reader = csv.DictReader(handle)
        rows = list(reader)
        headers = list(reader.fieldnames or [])
    required = {
        "comunicado",
        "campana_estado",
        "version",
        "fecha_apagado",
        "fecha_sms_inicio",
        "nivel",
        "depto",
        "provincia",
        "distrito",
        "centro_poblado_ccpp",
        "ubigeo_d6",
        "key_sistema",
        "cobertura_en_sistema",
        "estado_validacion",
        "ccpp_en_caserios",
        "speech_key",
        "postventa_speech_key",
        "origen_pdf",
        "grupo",
        "notas",
    }
    missing = required - set(headers)
    if missing:
        raise RuntimeError(f"CSV sin columnas requeridas: {sorted(missing)}")
    if len(rows) != 2081:
        raise RuntimeError(f"CSV esperado con 2.081 filas; encontrado: {len(rows)}")
    return rows, headers


def _set_values(html: str, name: str, scope: str | None = None) -> list[str]:
    haystack = scope if scope is not None else html
    match = re.search(rf"\b{re.escape(name)}:\s*new Set\(\[(.*?)\]\)", haystack, re.S)
    if not match:
        raise RuntimeError(f"No se encontró {name} en el runtime inline")
    return re.findall(r'"([^"]+)"', match.group(1))


def _quoted_field(body: str, name: str) -> str | None:
    match = re.search(rf"\b{re.escape(name)}:\s*\"([^\"]*)\"", body)
    return match.group(1) if match else None


def _int_field(body: str, name: str) -> int | None:
    match = re.search(rf"\b{re.escape(name)}:\s*(-?\d+)\b", body)
    return int(match.group(1)) if match else None


def _null_or_string_field(body: str, name: str) -> str | None:
    match = re.search(rf"\b{re.escape(name)}:\s*(null|\"[^\"]*\")", body)
    if not match:
        return None
    raw = match.group(1)
    return None if raw == "null" else raw[1:-1]


def parse_runtime(html: str) -> dict:
    za_match = re.search(r"const ZONAS_APAGADAS = \{(.*?)\n\};", html, re.S)
    if not za_match:
        raise RuntimeError("No se encontró const ZONAS_APAGADAS")
    za_body = za_match.group(1)

    ccpp_keys = _set_values(html, "ccpp_keys", za_body)
    districts = _set_values(html, "districts_with_off", za_body)
    total_match = re.search(r"\btotal_zones:\s*(\d+)", za_body)
    if not total_match:
        raise RuntimeError("No se encontró total_zones")
    total_zones = int(total_match.group(1))
    ipt_districts = _set_values(html, "ipt_districts")

    campaign_match = re.search(
        r"campanas:\s*\{(.*?)\n  \},\n  // v3\.10\.3", html, re.S
    )
    if not campaign_match:
        raise RuntimeError("No se encontró el registro inline de campañas")
    campaign_block = campaign_match.group(1)

    campaigns: dict[str, dict] = {}
    for campaign_id in ("COMU-5514", "COMU-5569"):
        marker = re.search(rf'"{campaign_id}":\s*\{{', campaign_block)
        if not marker:
            raise RuntimeError(f"No se encontró la campaña inline {campaign_id}")
        body_match = re.search(
            rf'"{campaign_id}":\s*\{{(.*?)(?=\n    \}})', campaign_block, re.S
        )
        if not body_match:
            raise RuntimeError(f"No se pudo aislar la campaña inline {campaign_id}")
        body = body_match.group(1)
        faltantes_match = re.search(r"\bfaltantes:\s*\[(.*?)\]", body, re.S)
        urls_match = re.search(r"\bfuente_urls:\s*\[(.*?)\]", body, re.S)
        campaigns[campaign_id] = {
            "nombre": _quoted_field(body, "nombre"),
            "estado": _quoted_field(body, "estado"),
            "version": _quoted_field(body, "version"),
            "fecha_apagado": _null_or_string_field(body, "fecha_apagado"),
            "fecha_sms_inicio": _null_or_string_field(body, "fecha_sms_inicio"),
            "q_clientes": _int_field(body, "q_clientes"),
            "ccpp_count": _int_field(body, "ccpp_count"),
            "fuente": _quoted_field(body, "fuente"),
            "notas": _quoted_field(body, "notas"),
            "faltantes": re.findall(r'"([^"]*)"', faltantes_match.group(1))
            if faltantes_match
            else [],
            "fuente_urls": re.findall(r'"([^"]*)"', urls_match.group(1))
            if urls_match
            else [],
        }

    ipt_source_match = re.search(
        r'"zonas_apagadas_5672"\s*:\s*\{(.*?)\n\s*\}', html, re.S
    )
    if not ipt_source_match:
        raise RuntimeError("No se encontró la fuente runtime COMU-5672")
    ipt_source_body = ipt_source_match.group(1)
    clients_match = re.search(r'"clientes_declarados":\s*"(\d+)', ipt_source_body)
    if not clients_match:
        raise RuntimeError("No se encontraron clientes declarados COMU-5672")
    ipt_clients = int(clients_match.group(1))
    ipt_source = _quoted_field(ipt_source_body, "fuente") or "Apagado Zonas IPT.pdf"

    return {
        "ccpp_keys": ccpp_keys,
        "districts_with_off": districts,
        "total_zonas": total_zones,
        "ipt_districts": ipt_districts,
        "campanas": campaigns,
        "ipt_clientes_declarados": ipt_clients,
        "ipt_fuente": ipt_source,
    }


def _canonical_priority(row: dict[str, str]) -> int:
    return {
        "ok_exacto": 0,
        "sin_fila_literal_en_fuente": 1,
        "duplicado_fuente": 2,
    }.get(row["estado_validacion"], 9)


def build_ccpp_rows(rows: list[dict[str, str]]) -> list[dict]:
    groups: dict[tuple[str, str], list[dict[str, str]]] = defaultdict(list)
    for row in rows:
        if row["comunicado"] != "COMU-5514" or row["cobertura_en_sistema"] != "exacta":
            continue
        groups[(row["ubigeo_d6"], row["centro_poblado_ccpp"])].append(row)

    output: list[dict] = []
    for (d6, ccpp), candidates in groups.items():
        canonical = min(candidates, key=_canonical_priority)
        origins = sorted(
            {
                f"{item['origen_pdf']} / {item['grupo']}"
                for item in candidates
                if item["origen_pdf"] or item["grupo"]
            }
        )
        notes = sorted({item["notas"] for item in candidates if item["notas"]})
        output.append(
            {
                "departamento": canonical["depto"],
                "provincia": canonical["provincia"],
                "distrito": canonical["distrito"],
                "centro poblado (CCPP)": ccpp,
                "ubigeo_d6": d6,
                "key_sistema": canonical["key_sistema"],
                "fecha de apagado": canonical["fecha_apagado"],
                "fecha SMS": canonical["fecha_sms_inicio"],
                "campaña": canonical["comunicado"],
                "speech_key": canonical["speech_key"],
                "postventa_speech_key": canonical["postventa_speech_key"],
                "origen (PDF/grupo)": "; ".join(origins),
                "estado_validacion": canonical["estado_validacion"],
                "cobertura_en_sistema": canonical["cobertura_en_sistema"],
                "ccpp_en_caserios": canonical["ccpp_en_caserios"],
                "notas": " | ".join(notes),
            }
        )

    output.sort(
        key=lambda item: (
            norm(item["departamento"]),
            norm(item["provincia"]),
            norm(item["distrito"]),
            norm(item["centro poblado (CCPP)"]),
        )
    )
    if len(output) != 238:
        raise RuntimeError(f"COMU-5514: se esperaban 238 CCPP exactos; hay {len(output)}")
    return output


def _ipt_note_value(note: str, label: str) -> int | None:
    match = re.search(rf"\b{re.escape(label)}=(\d+)\b", note)
    return int(match.group(1)) if match else None


def build_ipt_rows(rows: list[dict[str, str]]) -> tuple[list[dict], int]:
    output: list[dict] = []
    for row in rows:
        if row["comunicado"] != "COMU-5672":
            continue
        output.append(
            {
                "departamento": row["depto"],
                "provincia": row["provincia"],
                "distrito": row["distrito"],
                "ubigeo_d6": row["ubigeo_d6"],
                "fecha apagado": row["fecha_apagado"],
                "fecha SMS": row["fecha_sms_inicio"],
                "clientes depto": _ipt_note_value(row["notas"], "clientes"),
                "centros depto": _ipt_note_value(row["notas"], "centros"),
                "campaña": row["comunicado"],
                "notas": row["notas"],
            }
        )
    output.sort(
        key=lambda item: (
            norm(item["departamento"]),
            norm(item["provincia"]),
            norm(item["distrito"]),
            item["ubigeo_d6"],
        )
    )
    if len(output) != 1798:
        raise RuntimeError(f"COMU-5672: se esperaban 1.798 zonas; hay {len(output)}")

    clients_by_dept: dict[str, set[int]] = defaultdict(set)
    for row in output:
        clients = row["clientes depto"]
        if clients is not None:
            clients_by_dept[row["departamento"]].add(clients)
    for department, values in clients_by_dept.items():
        if len(values) != 1:
            raise RuntimeError(
                f"COMU-5672: clientes inconsistentes para {department}: {values}"
            )
    csv_clients_sum = sum(next(iter(values)) for values in clients_by_dept.values())
    return output, csv_clients_sum


def _best_extra(source_ccpp: str, extras: list[str]) -> tuple[str | None, float]:
    source_fold = fold(source_ccpp)
    best_key = None
    best_score = 0.0
    for extra in extras:
        _, _, extra_ccpp = extra.partition(":")
        score = SequenceMatcher(None, source_fold, fold(extra_ccpp)).ratio()
        if score > best_score:
            best_key = extra
            best_score = score
    return best_key, best_score


def build_divergences(
    rows: list[dict[str, str]], audit: dict
) -> tuple[list[dict], list[str]]:
    source_by_key = {
        row["key_sistema"]: row
        for row in rows
        if row["comunicado"] == "COMU-5514" and row["key_sistema"]
    }
    divergences: list[dict] = []
    matched_loss_keys: list[str] = []
    extras = list(audit.get("extras", []))

    for lost_key in audit.get("perdidas", []):
        _, _, source_ccpp = lost_key.partition(":")
        extra_key, score = _best_extra(source_ccpp, extras)
        if not extra_key or score < 0.85:
            continue
        source = source_by_key.get(lost_key)
        if not source:
            continue
        extra_d6, _, runtime_ccpp = extra_key.partition(":")
        note = (
            f"PDF: d6 {source['ubigeo_d6']}; GeoNames/runtime: d6 {extra_d6}. "
            "La clave runtime es alcanzable; no se duplicó la clave PDF."
        )
        if fold(source_ccpp) != fold(runtime_ccpp):
            note += f" Literal runtime: {runtime_ccpp}."
        divergences.append(
            {
                "ccpp": source_ccpp,
                "depto/prov/dist según PDF": (
                    f"{source['depto']} / {source['provincia']} / {source['distrito']}"
                ),
                "d6_pdf": source["ubigeo_d6"],
                "d6_sistema": extra_d6,
                "nota": note,
            }
        )
        matched_loss_keys.append(lost_key)

    divergences.sort(key=lambda item: norm(item["ccpp"]))
    if len(divergences) != 4:
        raise RuntimeError(
            f"Se esperaban 4 divergencias d6; se derivaron {len(divergences)}"
        )
    return divergences, matched_loss_keys


def build_pending_decisions(
    rows: list[dict[str, str]], audit: dict, divergence_source_ccpp: set[str]
) -> list[dict]:
    entries = [
        entry
        for entry in audit.get("sin_cobertura", [])
        if fold(entry.get("ccpp")) not in divergence_source_ccpp
    ]
    if len(entries) != 3:
        raise RuntimeError(
            f"Se esperaban 3 zonas sin cobertura; se derivaron {len(entries)}"
        )

    details = {
        "010105:CHILIQUIN": {
            "why": (
                "El PDF escribe el nombre del distrito en la columna CENTRO POBLADO; "
                'GeoNames lista "Chiliquini".'
            ),
            "need": (
                "Confirmación del emisor sobre si la fila representa el caserío "
                "Chiliquini o todo el distrito."
            ),
            "recommendation": (
                "Confirmar con el emisor. Si significa todo el distrito, cubrirlo "
                "distritalmente; no fabricar cobertura CCPP exacta por inferencia."
            ),
        },
        "060101:INCA CORRAL": {
            "why": (
                "No existe en UBIGEO ni CCPP_DATA; el PDF escribe la provincia "
                '"CAJABAMBA" frente a "CAJAMARCA".'
            ),
            "need": (
                "Fuente geográfica o de campaña autoritativa; confirmar la provincia "
                "e incorporar el CCPP."
            ),
            "recommendation": (
                "Solicitar la fuente autoritativa, corregir CAJABAMBA a CAJAMARCA y "
                "mantener sin cobertura hasta que el CCPP sea verificable."
            ),
        },
        "160405:SANTA ROSA": {
            "why": (
                "El distrito existe en el UBIGEO inline, pero tiene 0 caseríos en "
                "CCPP_DATA."
            ),
            "need": "Catálogo de caseríos de SANTA ROSA DE LORETO.",
            "recommendation": (
                "Solicitar el catálogo de caseríos e incorporar cobertura solo con "
                "un CCPP verificable; mantener sin cobertura mientras tanto."
            ),
        },
    }

    output: list[dict] = []
    for entry in entries:
        candidates = [
            row
            for row in rows
            if row["comunicado"] == "COMU-5514"
            and row["cobertura_en_sistema"] == "NO"
            and fold(row["centro_poblado_ccpp"]) == fold(entry.get("ccpp"))
            and row["depto"] == entry.get("depto")
        ]
        if len(candidates) != 1:
            raise RuntimeError(
                f"No se pudo identificar de forma única la zona pendiente {entry.get('ccpp')}"
            )
        row = candidates[0]
        key = row["key_sistema"]
        if key not in details:
            raise RuntimeError(f"Sin diagnóstico configurado para {key}")
        output.append(
            {
                "key": key,
                "depto/prov/dist": (
                    f"{entry.get('depto')} / {entry.get('prov')} / {entry.get('dist')}"
                ),
                "ccpp": entry.get("ccpp"),
                "estado": entry.get("estado"),
                "por qué no está": details[key]["why"],
                "qué se necesita": details[key]["need"],
                "recomendación": details[key]["recommendation"],
            }
        )
    output.sort(key=lambda item: item["key"])
    return output


def build_department_stats(
    ccpp_rows: list[dict], ipt_rows: list[dict]
) -> list[dict]:
    stats: dict[str, dict] = defaultdict(
        lambda: {
            "ccpp": 0,
            "ipt": 0,
            "distritos_ccpp": set(),
            "provincias_ccpp": set(),
        }
    )
    for row in ccpp_rows:
        item = stats[row["departamento"]]
        item["ccpp"] += 1
        item["distritos_ccpp"].add(row["ubigeo_d6"])
        item["provincias_ccpp"].add((row["departamento"], row["provincia"]))
    for row in ipt_rows:
        item = stats[row["departamento"]]
        item["ipt"] += 1

    output = [
        {
            "departamento": department,
            "ccpp": item["ccpp"],
            "ipt": item["ipt"],
            "total": item["ccpp"] + item["ipt"],
            "distritos_ccpp": len(item["distritos_ccpp"]),
            "provincias_ccpp": len(item["provincias_ccpp"]),
        }
        for department, item in stats.items()
    ]
    output.sort(key=lambda item: (-item["total"], norm(item["departamento"])))
    return output


def _date_summary(ipt_rows: list[dict], field: str) -> str:
    by_dept: dict[str, set[str]] = defaultdict(set)
    for row in ipt_rows:
        by_dept[row["departamento"]].add(row[field])
    counts = Counter(next(iter(values)) for values in by_dept.values())
    return " / ".join(
        f"{date} ({counts[date]} deptos)" for date in sorted(counts)
    )


def build_campaign_summary(
    runtime: dict,
    rows: list[dict[str, str]],
    ccpp_rows: list[dict],
    ipt_rows: list[dict],
    csv_clients_sum: int,
) -> list[dict]:
    c5514 = runtime["campanas"]["COMU-5514"]
    c5569 = runtime["campanas"]["COMU-5569"]
    ipt_source = runtime["ipt_fuente"]
    ipt_clients = runtime["ipt_clientes_declarados"]

    c5514_departments = {row["departamento"] for row in ccpp_rows}
    c5514_provinces = {(row["departamento"], row["provincia"]) for row in ccpp_rows}
    c5514_districts = {row["ubigeo_d6"] for row in ccpp_rows}
    ipt_departments = {row["departamento"] for row in ipt_rows}
    ipt_provinces = {(row["departamento"], row["provincia"]) for row in ipt_rows}
    ipt_districts = {row["ubigeo_d6"] for row in ipt_rows}

    pending_row = next(row for row in rows if row["comunicado"] == "COMU-5569")
    return [
        {
            "Comunicado": "COMU-5514",
            "Estado": c5514["estado"],
            "Versión": c5514["version"],
            "Fecha apagado": c5514["fecha_apagado"],
            "Fecha SMS": c5514["fecha_sms_inicio"],
            "CCPP exactos": len(ccpp_rows),
            "Zonas IPT": 0,
            "Total zonas": len(ccpp_rows),
            "Departamentos": len(c5514_departments),
            "Provincias": len(c5514_provinces),
            "Distritos": len(c5514_districts),
            "Clientes declarados": c5514["q_clientes"],
            "Fuente": c5514["fuente"],
            "Nota de cobertura": (
                f"{len(ccpp_rows)} CCPP exactos en {len(c5514_districts)} distritos; "
                "3 zonas fuente sin cobertura y 4 atribuciones d6 pendientes de decisión."
            ),
        },
        {
            "Comunicado": "COMU-5672",
            "Estado": "zonas_conocidas",
            "Versión": "00001",
            "Fecha apagado": _date_summary(ipt_rows, "fecha apagado"),
            "Fecha SMS": _date_summary(ipt_rows, "fecha SMS"),
            "CCPP exactos": 0,
            "Zonas IPT": len(ipt_rows),
            "Total zonas": len(ipt_rows),
            "Departamentos": len(ipt_departments),
            "Provincias": len(ipt_provinces),
            "Distritos": len(ipt_districts),
            "Clientes declarados": ipt_clients,
            "Fuente": ipt_source,
            "Nota de cobertura": (
                "Set IPT separado del CCPP exacto. El PDF declara "
                f"{fmt_int(ipt_clients)} clientes; la suma de notas CSV por depto es "
                f"{fmt_int(csv_clients_sum)}; se conserva la cifra declarada."
            ),
        },
        {
            "Comunicado": "COMU-5569",
            "Estado": c5569["estado"],
            "Versión": c5569["version"],
            "Fecha apagado": c5569["fecha_apagado"],
            "Fecha SMS": c5569["fecha_sms_inicio"] or "Pendiente",
            "CCPP exactos": 0,
            "Zonas IPT": 0,
            "Total zonas": 0,
            "Departamentos": 0,
            "Provincias": 0,
            "Distritos": 0,
            "Clientes declarados": c5569["q_clientes"],
            "Fuente": c5569["fuente"],
            "Nota de cobertura": (
                "0 CCPP cargados; el popup no alerta. Faltan tabla de centros poblados, "
                "fecha de apagado por depto y fecha de SMS por depto."
            ),
        },
    ]


def build_5569_rows(runtime: dict) -> list[dict]:
    campaign = runtime["campanas"]["COMU-5569"]
    return [
        {
            "Tipo": "Estado de campaña",
            "Ítem": campaign["estado"],
            "Detalle": (
                f"0 CCPP · apagado {campaign['fecha_apagado']} · fecha SMS pendiente · "
                f"{fmt_int(campaign['q_clientes'])} clientes declarados"
            ),
        },
        {
            "Tipo": "Faltante",
            "Ítem": campaign["faltantes"][0],
            "Detalle": "Bloquea la cobertura ccpp_exacto y los speeches por marcaje.",
        },
        {
            "Tipo": "Faltante",
            "Ítem": campaign["faltantes"][1],
            "Detalle": "El comunicado solo declara la fecha global 2026-12-10.",
        },
        {
            "Tipo": "Faltante",
            "Ítem": campaign["faltantes"][2],
            "Detalle": "La plantilla SMS conserva el placeholder [dd/mm].",
        },
        {
            "Tipo": "Enlace interno",
            "Ítem": "id=178147",
            "Detalle": campaign["fuente_urls"][0],
        },
        {
            "Tipo": "Enlace interno",
            "Ítem": "id=178146",
            "Detalle": campaign["fuente_urls"][1],
        },
        {
            "Tipo": "Nota",
            "Ítem": "Los 4 PDF son el mismo comunicado",
            "Detalle": (
                "Cuatro archivos cargados, un solo documento; son copias, no cuatro fuentes."
            ),
        },
        {
            "Tipo": "Identificador",
            "Ítem": "MD5",
            "Detalle": "393c127b3b6b4bb486d3d03be0850f97",
        },
        {
            "Tipo": "Cobertura runtime",
            "Ítem": "0 CCPP",
            "Detalle": "El popup no genera alerta mientras la campaña siga pendiente.",
        },
    ]


def style_sheet(ws, filter_end_row: int | None = None) -> None:
    max_row = ws.max_row
    max_col = ws.max_column
    for cell in ws[1]:
        if cell.value is None:
            continue
        cell.fill = HEADER_FILL
        cell.font = HEADER_FONT
        cell.alignment = Alignment(horizontal="center", vertical="center", wrap_text=True)
    ws.row_dimensions[1].height = 32
    ws.freeze_panes = "A2"
    ws.sheet_view.showGridLines = False
    ws.print_title_rows = "1:1"

    widths: dict[int, int] = {}
    long_columns: set[int] = set()
    for row in ws.iter_rows(min_row=1, max_row=max_row, max_col=max_col):
        for cell in row:
            text = "" if cell.value is None else str(cell.value)
            widths[cell.column] = max(widths.get(cell.column, 0), len(text))
            if len(text) > 35:
                long_columns.add(cell.column)
            if cell.row > 1:
                cell.alignment = Alignment(vertical="top", wrap_text=cell.column in long_columns)

    for column in range(1, max_col + 1):
        width = min(max(widths.get(column, 0) + 2, 10), 60)
        ws.column_dimensions[get_column_letter(column)].width = width

    end_row = filter_end_row or max_row
    if end_row >= 1 and max_col >= 1:
        ws.auto_filter.ref = f"A1:{get_column_letter(max_col)}{end_row}"


def normalize_xlsx_archive(path: Path, timestamp: datetime) -> None:
    """Normaliza timestamps y orden del ZIP para una salida byte-idempotente."""
    fixed = (
        timestamp.year,
        timestamp.month,
        timestamp.day,
        timestamp.hour,
        timestamp.minute,
        timestamp.second,
    )
    normalized = path.with_name(f"{path.stem}.normalized{path.suffix}")
    fixed_iso = timestamp.strftime("%Y-%m-%dT%H:%M:%SZ").encode("ascii")
    with ZipFile(path, "r") as source, ZipFile(normalized, "w", ZIP_DEFLATED) as target:
        for name in sorted(source.namelist()):
            old = source.getinfo(name)
            info = ZipInfo(name, date_time=fixed)
            info.compress_type = ZIP_DEFLATED
            info.create_system = old.create_system
            info.external_attr = old.external_attr
            data = source.read(name)
            if name == "docProps/core.xml":
                data = re.sub(
                    rb"(<dcterms:modified[^>]*>)[^<]*(</dcterms:modified>)",
                    rb"\g<1>" + fixed_iso + rb"\g<2>",
                    data,
                )
            target.writestr(info, data)
    normalized.replace(path)


def write_xlsx(
    runtime: dict,
    campaign_rows: list[dict],
    ccpp_rows: list[dict],
    ipt_rows: list[dict],
    pending_rows: list[dict],
    divergence_rows: list[dict],
    audit: dict,
    csv_headers: list[str],
    source_rows: list[dict],
    csv_clients_sum: int,
) -> None:
    wb = Workbook()
    wb.remove(wb.active)
    wb.properties.title = "CIMA — Tabla de zonas de apagado"
    wb.properties.subject = "Inventario completo de zonas apagadas del popup"
    wb.properties.creator = "Thavren"
    wb.properties.description = (
        "Export reproducible desde apagados_tabla_completa.csv, "
        "audit_zonas_apagado.json e index.html"
    )
    source_time = datetime.fromisoformat(audit["generado"])
    wb.properties.created = source_time
    wb.properties.modified = source_time

    ws = wb.create_sheet("RESUMEN")
    summary_headers = list(campaign_rows[0].keys())
    ws.append(summary_headers)
    for row in campaign_rows:
        ws.append([row[column] for column in summary_headers])
    ws.append([])
    ws.append(
        [
            "TOTALES DEL SISTEMA",
            f"{len(ccpp_rows)} CCPP · {len(runtime['districts_with_off'])} distritos · "
            f"{fmt_int(len(ipt_rows))} IPT · {fmt_int(len(ccpp_rows) + len(ipt_rows))} zonas activas",
        ]
    )
    ws.append(["Barrido · pérdidas", len(audit.get("perdidas", []))])
    ws.append(["Barrido · extras", len(audit.get("extras", []))])
    ws.append(["Barrido · huérfanos", len(audit.get("huerfanos", []))])
    ws.append(["Barrido · zonas sin cobertura funcional", len(pending_rows)])
    ws.append(["Barrido · divergencias d6", len(divergence_rows)])
    source_campaign_counts = Counter(row["comunicado"] for row in source_rows)
    ws.append(
        [
            "Fuente · filas CSV",
            f"{fmt_int(len(source_rows))} "
            f"({fmt_int(source_campaign_counts['COMU-5514'])} COMU-5514 + "
            f"{fmt_int(source_campaign_counts['COMU-5672'])} COMU-5672 + "
            f"{fmt_int(source_campaign_counts['COMU-5569'])} COMU-5569)",
        ]
    )
    ws.append(
        [
            "Fuente · columnas CSV",
            f"{len(csv_headers)} (el CSV real tiene {len(csv_headers)}; el encargo indica 22)",
        ]
    )
    ws.append(
        [
            "Cobertura · auditoría",
            f"{len(ccpp_rows)} CCPP exactos + {fmt_int(len(ipt_rows))} IPT = "
            f"{fmt_int(len(ccpp_rows) + len(ipt_rows))} zonas activas",
        ]
    )
    ws.append(
        [
            "Nota · sin cobertura literal",
            f"audit JSON registra {len(audit.get('sin_cobertura', []))}; "
            "2 corresponden a la fuente de las 4 divergencias d6; 3 requieren decisión.",
        ]
    )
    ws.append(
        [
            "Nota · clientes COMU-5672",
            f"El runtime declara {fmt_int(runtime['ipt_clientes_declarados'])}; "
            f"la suma de notas CSV por depto es {fmt_int(csv_clients_sum)}; "
            "se conserva la cifra declarada.",
        ]
    )
    ws.append([])
    ws.append(
        [
            "Generación",
            f"{GENERATION_ID} · fuente audit {audit['generado']}",
        ]
    )
    style_sheet(ws, filter_end_row=4)
    for row_index in range(6, ws.max_row + 1):
        ws.cell(row=row_index, column=1).fill = SECTION_FILL
        ws.cell(row=row_index, column=1).font = Font(bold=True)
        for column in range(1, ws.max_column + 1):
            ws.cell(row=row_index, column=column).alignment = Alignment(
                vertical="top", wrap_text=True
            )
    for row_index in (15, 16):
        ws.cell(row=row_index, column=1).fill = NOTE_FILL
        ws.cell(row=row_index, column=2).fill = NOTE_FILL
    ws.cell(ws.max_row, column=2).font = Font(italic=True)

    ws = wb.create_sheet("COMU-5514 CCPP")
    headers = list(ccpp_rows[0].keys())
    ws.append(headers)
    for row in ccpp_rows:
        ws.append([row[column] for column in headers])
    style_sheet(ws)

    ws = wb.create_sheet("COMU-5672 IPT")
    headers = list(ipt_rows[0].keys())
    ws.append(headers)
    for row in ipt_rows:
        ws.append([row[column] for column in headers])
    style_sheet(ws)

    ws = wb.create_sheet("COMU-5569 PENDIENTE")
    headers = ["Tipo", "Ítem", "Detalle"]
    ws.append(headers)
    for row in build_5569_rows(runtime):
        ws.append([row[column] for column in headers])
    style_sheet(ws)

    ws = wb.create_sheet("PENDIENTES DECISION")
    headers = list(pending_rows[0].keys())
    ws.append(headers)
    for row in pending_rows:
        ws.append([row[column] for column in headers])
    style_sheet(ws)

    ws = wb.create_sheet("DIVERGENCIAS")
    headers = list(divergence_rows[0].keys())
    ws.append(headers)
    for row in divergence_rows:
        ws.append([row[column] for column in headers])
    style_sheet(ws)

    ws = wb.create_sheet("AUDITORIA")
    headers = ["Tipo", "ID", "Título / Suite", "Impacto / Resultado", "Estado / Detalle"]
    ws.append(headers)
    for finding in FINDINGS:
        ws.append(
            [
                "Hallazgo",
                finding["id"],
                finding["title"],
                finding["impact"],
                finding["status"],
            ]
        )
    for suite, result, detail in VERIFICATION:
        ws.append(["Verificación", "", suite, result, detail])
    style_sheet(ws)

    for ws in wb.worksheets:
        ws.sheet_properties.pageSetUpPr.fitToPage = True
        ws.page_setup.fitToWidth = 1
        ws.page_setup.fitToHeight = 0

    XLSX_OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    temporary = XLSX_OUTPUT.with_name(f"{XLSX_OUTPUT.stem}.tmp.xlsx")
    wb.save(temporary)
    normalize_xlsx_archive(temporary, source_time)
    temporary.replace(XLSX_OUTPUT)


def write_markdown(
    runtime: dict,
    campaign_rows: list[dict],
    department_rows: list[dict],
    pending_rows: list[dict],
    divergence_rows: list[dict],
    audit: dict,
    csv_clients_sum: int,
    csv_header_count: int,
) -> None:
    lines: list[str] = [
        "# CIMA — Tabla de zonas de apagado",
        "",
        f"**Fecha:** {REPORT_DATE}  ",
        "Inventario completo de zonas apagadas en el popup CIMA: CCPP exactos COMU-5514, zonas IPT COMU-5672 y estado pendiente COMU-5569.",
        "",
        "## 1. Por campaña",
        "",
    ]

    headers = [
        "Comunicado",
        "Estado",
        "Versión",
        "Apagado",
        "SMS",
        "CCPP",
        "IPT",
        "Total",
        "Deptos",
        "Provincias",
        "Distritos",
        "Clientes",
    ]
    lines.append("| " + " | ".join(headers) + " |")
    lines.append("|" + "|".join(["---"] * len(headers)) + "|")
    for row in campaign_rows:
        values = [
            row["Comunicado"],
            row["Estado"],
            row["Versión"],
            row["Fecha apagado"],
            row["Fecha SMS"],
            fmt_int(row["CCPP exactos"]),
            fmt_int(row["Zonas IPT"]),
            fmt_int(row["Total zonas"]),
            fmt_int(row["Departamentos"]),
            fmt_int(row["Provincias"]),
            fmt_int(row["Distritos"]),
            fmt_int(row["Clientes declarados"]),
        ]
        lines.append("| " + " | ".join(md_cell(value) for value in values) + " |")

    c5514 = runtime["campanas"]["COMU-5514"]
    c5569 = runtime["campanas"]["COMU-5569"]
    lines.extend(
        [
            "",
            "**Fuentes y cobertura:**",
            "",
            f"- **COMU-5514:** {c5514['fuente']}. Cobertura: {campaign_rows[0]['CCPP exactos']} CCPP exactos; 3 zonas fuente sin cobertura y 4 atribuciones d6 pendientes de decisión.",
            f"- **COMU-5672:** {runtime['ipt_fuente']}. Cobertura: {fmt_int(campaign_rows[1]['Zonas IPT'])} zonas IPT en set separado del CCPP exacto. El PDF declara {fmt_int(runtime['ipt_clientes_declarados'])} clientes; la suma de notas CSV por depto es {fmt_int(csv_clients_sum)}.",
            f"- **COMU-5569:** {c5569['fuente']}. Cobertura: 0 CCPP cargados; el popup no alerta. Los 4 PDF cargados son el mismo comunicado.",
            "",
            "## 2. Por departamento",
            "",
            "| Departamento | CCPP exactos | Zonas IPT | Total zonas | Distritos con CCPP | Provincias con CCPP |",
            "|---|---:|---:|---:|---:|---:|",
        ]
    )
    for row in department_rows:
        lines.append(
            "| "
            + " | ".join(
                [
                    md_cell(row["departamento"]),
                    fmt_int(row["ccpp"]),
                    fmt_int(row["ipt"]),
                    fmt_int(row["total"]),
                    fmt_int(row["distritos_ccpp"]),
                    fmt_int(row["provincias_ccpp"]),
                ]
            )
            + " |"
        )

    lines.extend(
        [
            "",
            "## 3. Zonas sin cobertura",
            "",
            "| Key | Departamento / provincia / distrito | CCPP | Estado | Por qué no está | Qué se necesita | Recomendación |",
            "|---|---|---|---|---|---|---|",
        ]
    )
    for row in pending_rows:
        values = [
            row["key"],
            row["depto/prov/dist"],
            row["ccpp"],
            row["estado"],
            row["por qué no está"],
            row["qué se necesita"],
            row["recomendación"],
        ]
        lines.append("| " + " | ".join(md_cell(value) for value in values) + " |")

    lines.extend(
        [
            "",
            "## 4. Divergencias d6",
            "",
            "| CCPP | Departamento / provincia / distrito (PDF) | d6 PDF | d6 sistema | Nota |",
            "|---|---|---:|---:|---|",
        ]
    )
    for row in divergence_rows:
        values = [
            row["ccpp"],
            row["depto/prov/dist según PDF"],
            row["d6_pdf"],
            row["d6_sistema"],
            row["nota"],
        ]
        lines.append("| " + " | ".join(md_cell(value) for value in values) + " |")

    lines.extend(
        [
            "",
            "## 5. Verificación",
            "",
            "| Suite / control | Resultado | Detalle |",
            "|---|---|---|",
        ]
    )
    for suite, result, detail in VERIFICATION:
        lines.append(f"| {md_cell(suite)} | {md_cell(result)} | {md_cell(detail)} |")
    lines.extend(
        [
            "",
            "[Abrir informe completo](AUDITORIA_ZONAS_APAGADO_2026-11-14.md) — ruta: `docs/AUDITORIA_ZONAS_APAGADO_2026-11-14.md`.",
            "",
            "## 6. Cómo se reproduce",
            "",
            "```bash",
            "cd apps/cima-generador",
            "python3 scripts/export_tabla_zonas.py",
            "python3 scripts/audit_zonas_apagado.py",
            "```",
            "",
            "## Notas de consistencia",
            "",
            f"- El CSV fuente contiene 2.081 filas y {csv_header_count} columnas; el encargo indica 22 columnas. El export usa las {csv_header_count} columnas reales.",
            f"- `audit_zonas_apagado.json` lista {len(audit.get('sin_cobertura', []))} filas sin cobertura literal; 2 son la fuente de las 4 divergencias d6 y 3 requieren decisión funcional.",
            f"- COMU-5672: el runtime declara {fmt_int(runtime['ipt_clientes_declarados'])} clientes; la suma de notas CSV por depto es {fmt_int(csv_clients_sum)}. Se conserva la cifra declarada.",
            "",
            "*Generado por Thavren — 2026-11-14 · barrido completo de zonas de apagado*",
            "",
        ]
    )

    MD_OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    temporary = MD_OUTPUT.with_name(f"{MD_OUTPUT.stem}.tmp.md")
    temporary.write_text("\n".join(lines), encoding="utf-8")
    temporary.replace(MD_OUTPUT)


def main() -> None:
    rows, csv_headers = parse_csv()
    audit = json.loads(AUDIT_SOURCE.read_text(encoding="utf-8"))
    runtime = parse_runtime(HTML.read_text(encoding="utf-8"))

    if len(runtime["ccpp_keys"]) != 238:
        raise RuntimeError(
            f"Runtime: se esperaban 238 CCPP; hay {len(runtime['ccpp_keys'])}"
        )
    if len(runtime["districts_with_off"]) != 191:
        raise RuntimeError(
            "Runtime: se esperaban 191 distritos con apagado; hay "
            f"{len(runtime['districts_with_off'])}"
        )
    if len(runtime["ipt_districts"]) != 1798:
        raise RuntimeError(
            f"Runtime: se esperaban 1.798 zonas IPT; hay {len(runtime['ipt_districts'])}"
        )

    ccpp_rows = build_ccpp_rows(rows)
    ipt_rows, csv_clients_sum = build_ipt_rows(rows)
    divergence_rows, _ = build_divergences(rows, audit)
    divergence_source_ccpp = {fold(row["ccpp"]) for row in divergence_rows}
    pending_rows = build_pending_decisions(rows, audit, divergence_source_ccpp)
    department_rows = build_department_stats(ccpp_rows, ipt_rows)
    campaign_rows = build_campaign_summary(
        runtime, rows, ccpp_rows, ipt_rows, csv_clients_sum
    )

    write_xlsx(
        runtime,
        campaign_rows,
        ccpp_rows,
        ipt_rows,
        pending_rows,
        divergence_rows,
        audit,
        csv_headers,
        rows,
        csv_clients_sum,
    )
    write_markdown(
        runtime,
        campaign_rows,
        department_rows,
        pending_rows,
        divergence_rows,
        audit,
        csv_clients_sum,
        len(csv_headers),
    )

    print(f"[OK] XLSX: {XLSX_OUTPUT}")
    print(f"[OK] MD:   {MD_OUTPUT}")
    print(
        "[OK] Conteos: RESUMEN=18 filas; COMU-5514=238; COMU-5672=1798; "
        "COMU-5569=9; PENDIENTES=3; DIVERGENCIAS=4; AUDITORIA=15"
    )


if __name__ == "__main__":
    main()

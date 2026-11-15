#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ingest_comunicado_zonas.py
===========================
Ingesta de zonas de apagado (CCPP) asociadas a un comunicado Movistar
(COMU-XXXX) hacia dos destinos:

  1) data/zonas_apagadas_<COMUNICADO>.json   -> campaña (fuente de verdad)
  2) apps/cima-generador/index.html           -> bloque inline (opt-in)

REGLAS DURAS (no negociables)
-----------------------------
* DRY-RUN por defecto.  Sin ``--apply-inline`` NO se escribe NADA
  (tampoco data/*.json).  El reporte completo se imprime igual.
* IDEMPOTENCIA.  Dedupe por md5 de la fuente (data/.ingest_dedupe_<COM>.json)
  + dedupe por fila lógica (d6 + ccpp normalizado).  Re-ejecutar con la misma
  fuente = 0 duplicados, 0 cambios.
* HONESTIDAD.  Lo que no resuelve UBIGEO va a ``unresolved`` con motivo.
  NUNCA se fuerza al distrito "más cercano" ni se inventa un d6.
* TRAZABILIDAD.  md5 + sha256 de la fuente van en cada salida.
* SIN RED, SIN SECRETOS.  Solo stdlib + (pdfplumber | openpyxl) si existen.

Uso
---
    python3 scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 \
        --source "tabla_zonas.csv" [--modo union|replace] [--dry-run] \
        [--apply-inline] [--campana-json salida.json]

Códigos de salida
-----------------
    0 = OK (incluye dry-run)
    1 = error de uso / fuente inválida / parseo fallido
    2 = patrón NO encontrado en index.html  -> ABORTA sin escribir
    3 = dependencia opcional ausente (pdfplumber / openpyxl)
"""

from __future__ import annotations

import argparse
import csv
import datetime as _dt
import hashlib
import io
import json
import os
import re
import shutil
import sys
import unicodedata
from difflib import SequenceMatcher, get_close_matches
from pathlib import Path

# ═══════════════════════════════════════════════════════════════════════════
# RUTAS POR DEFECTO
# ═══════════════════════════════════════════════════════════════════════════

APP_DIR = Path(__file__).resolve().parents[1]        # apps/cima-generador
ROOT_DIR = Path(__file__).resolve().parents[3]       # raíz del repo
DEFAULT_DATA_DIR = APP_DIR / 'data'
DEFAULT_HTML_PATH = APP_DIR / 'index.html'
DEFAULT_UBIGEO_PATH = DEFAULT_DATA_DIR / 'ubigeo_consolidated.json'

# ═══════════════════════════════════════════════════════════════════════════
# CONSTANTES DE DOMINIO
# ═══════════════════════════════════════════════════════════════════════════

FUZZY_CUTOFF = 0.78
SUPPORTED_EXTS = ('.csv', '.xlsx', '.xlsm', '.pdf', '.json')

# Alias de encabezados -> campo canónico.  Acepta acentos, guiones y
# espacios porque canon_header() los normaliza a UPPER_SNAKE ASCII.
HEADER_ALIASES = {
    'depto': ('DEPARTAMENTO', 'DEPT', 'DPTO', 'DEPART'),
    'prov': ('PROVINCIA', 'PROV'),
    'dist': ('DISTRITO', 'DIST', 'DISTRITOS'),
    'ccpp': ('CENTRO_POBLADO', 'CENTROPOBLADO', 'CENTRO_POBLADOS', 'CCPP',
             'CCPP_POBLADO', 'CASERIO', 'LOCALIDAD', 'POBLADO', 'CENTRO_POBLADO_NOMBRE'),
    'fecha_apagado': ('FECHA_APAGADO', 'FECHA_DE_APAGADO', 'APAGADO',
                      'FECHA_APAGADO_ZONA', 'FECHA_DE_APAGADO_DE_ZONA'),
    'fecha_sms_inicio': ('FECHA_SMS_INICIO', 'FECHA_SMS', 'SMS_INICIO',
                         'FECHA_INICIO_SMS', 'FECHA_SMS_DE_INICIO'),
}

# Orden de palabras clave para encabezados PDF "pegados" en una sola celda.
PDF_KEYWORD_ORDER = (
    ('DEPARTAMENTO', 'depto'),
    ('PROVINCIA', 'prov'),
    ('DISTRITO', 'dist'),
    ('CENTRO POBLADO', 'ccpp'),
    ('CENTROPOBLADO', 'ccpp'),
    ('CCPP', 'ccpp'),
    ('FECHA APAGADO', 'fecha_apagado'),
    ('FECHA SMS INICIO', 'fecha_sms_inicio'),
)

# Contenedores candidatos dentro de ZONAS_APAGADAS_INTEL (en ese orden).
INTEL_CONTAINERS = ('campanas', 'comunicados')

FIELDS = ('depto', 'prov', 'dist', 'ccpp', 'fecha_apagado', 'fecha_sms_inicio')


class IngestError(Exception):
    """Error con código de salida asociado."""

    def __init__(self, mensaje, code=1):
        super().__init__(mensaje)
        self.code = code


# ═══════════════════════════════════════════════════════════════════════════
# UTILIDADES DE TEXTO / FECHAS / HASHES
# ═══════════════════════════════════════════════════════════════════════════

def normalize(s):
    """Estándar v2.0.9 del repo: trim + UPPER + single-space + NFC."""
    if s is None:
        return ''
    s = str(s).strip().upper()
    s = re.sub(r'\s+', ' ', s)
    return unicodedata.normalize('NFC', s)


def canon_header(s):
    """Normaliza un encabezado de columna: sin acentos, UPPER_SNAKE."""
    if s is None:
        return ''
    s = unicodedata.normalize('NFKD', str(s))
    s = s.encode('ascii', 'ignore').decode('ascii')
    s = re.sub(r'[^A-Za-z0-9]+', '_', s).strip('_')
    return s.upper()


def fuzzy_key(s):
    """Clave de coincidencia difusa: normalize() + NFKD (mantiene la tilde).

    Mismo criterio que scripts/regenerate_data.py para que los typos del PDF
    (OCOÑA, MUÑANI, FERREÑAFE...) se comporten igual que en el pipeline previo.
    """
    return unicodedata.normalize('NFKD', normalize(s))


def norm_fecha(v):
    """Normaliza una fecha a ISO YYYY-MM-DD.  Devuelve '' si no se puede."""
    if v is None:
        return ''
    if isinstance(v, _dt.datetime):
        return v.date().isoformat()
    if isinstance(v, _dt.date):
        return v.isoformat()
    s = str(v).strip()
    if not s:
        return ''
    for fmt in ('%Y-%m-%d', '%d/%m/%Y', '%d-%m-%Y', '%Y/%m/%d', '%d.%m.%Y',
                '%Y-%m-%d %H:%M:%S', '%d/%m/%Y %H:%M:%S'):
        try:
            return _dt.datetime.strptime(s, fmt).date().isoformat()
        except ValueError:
            continue
    return s


def file_hashes(path):
    """(md5, sha256, bytes) del archivo fuente."""
    md5 = hashlib.md5()
    sha = hashlib.sha256()
    total = 0
    with open(path, 'rb') as fh:
        for chunk in iter(lambda: fh.read(1024 * 256), b''):
            md5.update(chunk)
            sha.update(chunk)
            total += len(chunk)
    return md5.hexdigest(), sha.hexdigest(), total


def now_iso():
    return _dt.datetime.now().isoformat(timespec='seconds')


# ═══════════════════════════════════════════════════════════════════════════
# ÍNDICE UBIGEO + RESOLUCIÓN (exacto primero, fuzzy después, nunca inventado)
# ═══════════════════════════════════════════════════════════════════════════

class UbigeoIndex:
    """Índice de UBIGEO consolidado: {DEPTO: {provincias: {PROV: {distritos}}}}."""

    def __init__(self, data):
        self._deptos = {}        # norm(depto)  -> depto real
        self._provs = {}         # norm(depto)  -> {norm(prov): prov real}
        self._provs_fuzzy = {}   # norm(depto)  -> {fuzzy(prov): prov real}
        self._dists = {}         # (norm(depto), norm(prov)) -> {norm(dist): dist real}
        self._dists_fuzzy = {}   # (norm(depto), norm(prov)) -> {fuzzy(dist): dist real}
        self._ubigeo = {}        # (norm(depto), norm(prov), norm(dist)) -> d6

        for depto, dnode in (data or {}).items():
            nd = normalize(depto)
            self._deptos[nd] = depto
            provs = (dnode or {}).get('provincias') or {}
            pmap, pfmap = {}, {}
            for prov, pnode in provs.items():
                np_ = normalize(prov)
                pmap[np_] = prov
                pfmap[fuzzy_key(prov)] = prov
                dists = (pnode or {}).get('distritos') or {}
                dmap, dfmap = {}, {}
                for dist, dnode2 in dists.items():
                    ndist = normalize(dist)
                    dmap[ndist] = dist
                    dfmap[fuzzy_key(dist)] = dist
                    code = (dnode2 or {}).get('_ubigeo')
                    if code:
                        self._ubigeo[(nd, np_, ndist)] = str(code)
                self._dists[(nd, np_)] = dmap
                self._dists_fuzzy[(nd, np_)] = dfmap
            self._provs[nd] = pmap
            self._provs_fuzzy[nd] = pfmap

    # -- helpers ----------------------------------------------------------
    def resolve_depto(self, depto):
        nd = normalize(depto)
        if nd in self._deptos:
            return self._deptos[nd], 'exacto'
        cand = self._best_fuzzy(fuzzy_key(depto), list(self._deptos.keys()))
        if cand:
            return self._deptos[cand], 'fuzzy_depto'
        return None, None

    def resolve_prov(self, depto_real, prov):
        ndep = normalize(depto_real)
        np_ = normalize(prov)
        pmap = self._provs.get(ndep, {})
        if np_ in pmap:
            return pmap[np_], 'exacto'
        pfmap = self._provs_fuzzy.get(ndep, {})
        cand = self._best_fuzzy(fuzzy_key(prov), list(pfmap.keys()))
        if cand:
            return pfmap[cand], 'fuzzy_prov'
        return None, None

    def resolve_dist(self, depto_real, prov_real, dist):
        key = (normalize(depto_real), normalize(prov_real))
        dmap = self._dists.get(key, {})
        ndist = normalize(dist)
        if ndist in dmap:
            return dmap[ndist], 'exacto'
        dfmap = self._dists_fuzzy.get(key, {})
        cand = self._best_fuzzy(fuzzy_key(dist), list(dfmap.keys()))
        if cand:
            return dfmap[cand], 'fuzzy_dist'
        return None, None

    @staticmethod
    def _best_fuzzy(query, candidates):
        """Mejor coincidencia NFKD con cutoff.  None si es ambigua o baja."""
        if not query or not candidates:
            return None
        matches = get_close_matches(query, candidates, n=2, cutoff=FUZZY_CUTOFF)
        if not matches:
            return None
        best = matches[0]
        if len(matches) > 1:
            r1 = SequenceMatcher(None, query, best).ratio()
            r2 = SequenceMatcher(None, query, matches[1]).ratio()
            if abs(r1 - r2) < 1e-9:      # empate -> ambiguo
                return None
        return best

    def resolve(self, depto, prov, dist):
        """Resuelve (depto, prov, dist) -> dict o None + motivo."""
        depto_real, met_depto = self.resolve_depto(depto)
        if not depto_real:
            return None, 'departamento no existe en UBIGEO: %s' % (depto or '(vacío)')
        prov_real, met_prov = self.resolve_prov(depto_real, prov)
        if not prov_real:
            return None, 'provincia no existe en %s: %s' % (depto_real, prov or '(vacía)')
        dist_real, met_dist = self.resolve_dist(depto_real, prov_real, dist)
        if not dist_real:
            return None, 'distrito no existe en %s/%s: %s' % (
                depto_real, prov_real, dist or '(vacío)')
        d6 = self._ubigeo.get((normalize(depto_real), normalize(prov_real),
                               normalize(dist_real)))
        if not d6:
            return None, 'distrito sin código UBIGEO: %s/%s/%s' % (
                depto_real, prov_real, dist_real)
        metodo = 'exacto'
        fuzzy_parts = [m for m in (met_depto, met_prov, met_dist)
                       if m and m != 'exacto']
        if fuzzy_parts:
            metodo = 'fuzzy(%s)' % ','.join(fuzzy_parts)
        return {
            'ubigeo': d6,
            'depto': depto_real,
            'prov': prov_real,
            'dist': dist_real,
            'metodo': metodo,
        }, None


def load_ubigeo(path):
    try:
        with open(path, 'r', encoding='utf-8') as fh:
            data = json.load(fh)
    except FileNotFoundError:
        raise IngestError('No existe el archivo UBIGEO: %s' % path)
    except json.JSONDecodeError as exc:
        raise IngestError('UBIGEO ilegible (%s): %s' % (path, exc))
    return UbigeoIndex(data)


# ═══════════════════════════════════════════════════════════════════════════
# LECTORES DE FUENTE
# ═══════════════════════════════════════════════════════════════════════════

def detect_format(path):
    ext = path.suffix.lower()
    if ext not in SUPPORTED_EXTS:
        raise IngestError(
            'Formato no soportado: %s (soportados: %s)' % (ext or '(sin ext)',
                                                           ', '.join(SUPPORTED_EXTS)))
    if ext in ('.xlsx', '.xlsm'):
        return 'xlsx'
    return ext[1:]


def map_columns(headers):
    """{índice de columna -> campo canónico} a partir de encabezados."""
    canons = [canon_header(h) for h in headers]
    mapping = {}
    taken = set()
    for field, aliases in HEADER_ALIASES.items():
        hit = None
        for i, ch in enumerate(canons):
            if i in taken or not ch:
                continue
            if ch in aliases:
                hit = i
                break
        if hit is None:
            for i, ch in enumerate(canons):
                if i in taken or not ch:
                    continue
                if any(ch.startswith(a) for a in aliases):
                    hit = i
                    break
        if hit is not None:
            mapping[hit] = field
            taken.add(hit)
    return mapping


def _rows_to_records(rows, mapping):
    """Convierte filas (listas) + mapeo de columnas -> lista de dicts."""
    out = []
    for row in rows:
        rec = {}
        for idx, field in mapping.items():
            val = row[idx] if idx < len(row) else ''
            rec[field] = val
        out.append(rec)
    return out


def _find_header_row(rows, max_scan=12):
    """Busca la fila de encabezado y devuelve (índice, mapping)."""
    best = (None, {})
    for i, row in enumerate(rows[:max_scan]):
        mapping = map_columns(row)
        if 'ccpp' in mapping.values():
            score = len(mapping)
            if best[0] is None or score > len(best[1]):
                best = (i, mapping)
    return best


def read_csv_rows(path):
    try:
        raw = path.read_bytes().decode('utf-8-sig')
    except UnicodeDecodeError:
        raw = path.read_bytes().decode('latin-1')
    if not raw.strip():
        raise IngestError('CSV vacío: %s' % path)

    # Autodetección de separador: el que produce más columnas en la 1ª fila.
    sample_lines = [l for l in raw.splitlines() if l.strip()][:20]
    delim, best_cols = ',', 1
    for cand in (',', ';', '\t', '|'):
        try:
            ncols = max(len(r) for r in csv.reader(sample_lines, delimiter=cand))
        except Exception:
            ncols = 1
        if ncols > best_cols:
            delim, best_cols = cand, ncols
    rows = [r for r in csv.reader(io.StringIO(raw), delimiter=delim)]
    rows = [r for r in rows if any((c or '').strip() for c in r)]
    if not rows:
        raise IngestError('CSV sin datos: %s' % path)
    h_idx, mapping = _find_header_row(rows)
    if h_idx is None:
        raise IngestError(
            'CSV sin encabezado reconocible (se esperaron columnas DEPARTAMENTO, '
            'PROVINCIA, DISTRITO y CENTRO POBLADO/CCPP): %s' % path)
    return _rows_to_records(rows[h_idx + 1:], mapping), 'csv(sep=%r)' % delim


def read_xlsx_rows(path):
    try:
        import openpyxl
    except ImportError:
        raise IngestError(
            'openpyxl no está instalado.  Instálalo con: pip install openpyxl',
            code=3)
    try:
        wb = openpyxl.load_workbook(str(path), read_only=True, data_only=True)
    except Exception as exc:
        raise IngestError('No se pudo leer el XLSX %s: %s' % (path, exc))
    try:
        sheet = None
        rows = []
        for ws in wb.worksheets:
            data = [[c for c in r] for r in ws.iter_rows(values_only=True)]
            data = [r for r in data if any(c is not None and str(c).strip() for c in r)]
            if data:
                sheet, rows = ws.title, data
                break
        if not rows:
            raise IngestError('XLSX sin datos: %s' % path)
        h_idx, mapping = _find_header_row(rows)
        if h_idx is None:
            raise IngestError(
                'XLSX sin encabezado reconocible (se esperaron columnas DEPARTAMENTO, '
                'PROVINCIA, DISTRITO y CENTRO POBLADO/CCPP): %s' % path)
        return _rows_to_records(rows[h_idx + 1:], mapping), 'xlsx(hoja=%s)' % sheet
    finally:
        try:
            wb.close()
        except Exception:
            pass


def read_json_rows(path):
    try:
        data = json.loads(path.read_text(encoding='utf-8'))
    except json.JSONDecodeError as exc:
        raise IngestError('JSON ilegible (%s): %s' % (path, exc))
    if isinstance(data, dict):
        for key in ('zonas', 'rows', 'data', 'items', 'registros'):
            if isinstance(data.get(key), list):
                data = data[key]
                break
        else:
            raise IngestError(
                'JSON objeto sin lista de zonas (se probaron: zonas, rows, data, '
                'items, registros): %s' % path)
    if not isinstance(data, list):
        raise IngestError('JSON con estructura no soportada: %s' % path)

    records = []
    for item in data:
        if not isinstance(item, dict):
            continue
        canons = {canon_header(k): v for k, v in item.items()}
        mapping = {}
        for field, aliases in HEADER_ALIASES.items():
            hit = None
            for ch, val in canons.items():
                if ch in aliases:
                    hit = ch
                    break
            if hit is None:
                for ch, val in canons.items():
                    if any(ch.startswith(a) for a in aliases):
                        hit = ch
                        break
            if hit is not None:
                mapping[hit] = field
        if 'ccpp' not in mapping.values():
            continue
        by_field = {field: key for key, field in mapping.items()}
        records.append({f: canons.get(by_field.get(f, ''), '') for f in FIELDS})
    if not records:
        raise IngestError(
            'JSON sin filas con DEPARTAMENTO/PROVINCIA/DISTRITO/CCPP: %s' % path)
    return records, 'json'


def _map_pdf_columns(header_cells, n_cols):
    """Mapea columnas de una tabla PDF -> campo canónico (o None)."""
    cells = [(c or '').strip() for c in header_cells]
    canons = [canon_header(c) for c in cells]

    # (a) encabezado con una celda por columna
    mapping = {}
    taken = set()
    for field, aliases in HEADER_ALIASES.items():
        for i, ch in enumerate(canons):
            if i in taken or not ch:
                continue
            if ch in aliases:
                mapping[i] = field
                taken.add(i)
                break
    if 'ccpp' in mapping.values() and len(mapping) >= 3:
        return mapping

    # (b) encabezado "pegado" en una sola celda (caso real APAGADO.pdf)
    joined = canon_header(' '.join(c for c in cells if c))
    hits = []
    for kw, field in PDF_KEYWORD_ORDER:
        pos = joined.find(canon_header(kw))
        if pos >= 0:
            hits.append((pos, field))
    hits.sort()
    fields = []
    for _pos, field in hits:
        if field not in fields:
            fields.append(field)
    if len(fields) >= 4:
        mapping = {}
        for i, field in enumerate(fields[:n_cols]):
            mapping[i] = field
        return mapping
    return None


def read_pdf_rows(path):
    try:
        import pdfplumber
    except ImportError:
        raise IngestError(
            'pdfplumber no está instalado.  Instálalo con: pip install pdfplumber',
            code=3)
    records = []
    notes = []
    carry = {'depto': '', 'prov': '', 'dist': ''}
    try:
        with pdfplumber.open(str(path)) as pdf:
            for pno, page in enumerate(pdf.pages, 1):
                try:
                    tables = page.extract_tables() or []
                except Exception as exc:                      # pragma: no cover
                    notes.append('pág %d: no se pudo extraer tablas (%s)' % (pno, exc))
                    continue
                for tno, table in enumerate(tables, 1):
                    if not table or len(table) < 2:
                        continue
                    n_cols = max(len(r) for r in table)
                    mapping, h_idx = None, None
                    for i, row in enumerate(table[:6]):
                        cand = _map_pdf_columns(row, n_cols)
                        if cand and 'ccpp' in cand.values():
                            mapping, h_idx = cand, i
                            break
                    if mapping is None:
                        notes.append('pág %d tabla %d: encabezado no reconocido '
                                     '(se omite)' % (pno, tno))
                        continue
                    for row in table[h_idx + 1:]:
                        vals = [(c or '').strip() for c in row]
                        rec = {}
                        for ci, field in mapping.items():
                            rec[field] = vals[ci] if ci < len(vals) else ''
                        # forward-fill de la jerarquía (celdas merged del PDF)
                        for f in ('depto', 'prov', 'dist'):
                            if not rec.get(f):
                                rec[f] = carry[f]
                            else:
                                carry[f] = rec[f]
                        if not normalize(rec.get('ccpp')):
                            continue
                        records.append({f: rec.get(f, '') for f in FIELDS})
    except IngestError:
        raise
    except Exception as exc:
        raise IngestError('No se pudo leer el PDF %s: %s' % (path, exc))
    if not records:
        raise IngestError(
            'PDF sin tabla de zonas reconocible (se esperaron columnas DEPARTAMENTO, '
            'PROVINCIA, DISTRITO y CENTRO POBLADO/CCPP): %s' % path)
    return records, 'pdf', notes


def read_rows(path):
    """Lee la fuente y devuelve (records, formato, notas)."""
    fmt = detect_format(path)
    if fmt == 'csv':
        recs, desc = read_csv_rows(path)
        return recs, desc, []
    if fmt == 'xlsx':
        recs, desc = read_xlsx_rows(path)
        return recs, desc, []
    if fmt == 'json':
        recs, desc = read_json_rows(path)
        return recs, desc, []
    if fmt == 'pdf':
        return read_pdf_rows(path)
    raise IngestError('Formato no soportado: %s' % fmt)


# ═══════════════════════════════════════════════════════════════════════════
# RESOLUCIÓN DE FILAS → ZONAS / UNRESOLVED
# ═══════════════════════════════════════════════════════════════════════════

def resolve_rows(records, ubigeo):
    """Divide filas en (zonas resueltas, unresolved)."""
    zones, unresolved = [], []
    seen = set()
    for rec in records:
        depto = normalize(rec.get('depto'))
        prov = normalize(rec.get('prov'))
        dist = normalize(rec.get('dist'))
        ccpp = normalize(rec.get('ccpp'))
        if not ccpp:
            continue
        res, motivo = ubigeo.resolve(depto, prov, dist)
        if res is None:
            key = ('U', depto, prov, dist, ccpp)
            if key in seen:
                continue
            seen.add(key)
            unresolved.append({
                'depto': depto, 'prov': prov, 'dist': dist, 'ccpp': ccpp,
                'motivo': motivo,
                'fecha_apagado': norm_fecha(rec.get('fecha_apagado')),
                'fecha_sms_inicio': norm_fecha(rec.get('fecha_sms_inicio')),
            })
            continue
        d6 = res['ubigeo']
        key = ('Z', d6, ccpp)
        if key in seen:                      # dedupe por fila lógica
            continue
        seen.add(key)
        zones.append({
            'depto': res['depto'],
            'prov': res['prov'],
            'dist': res['dist'],
            'ccpp': ccpp,
            'ubigeo': d6,
            'fecha_apagado': norm_fecha(rec.get('fecha_apagado')),
            'fecha_sms_inicio': norm_fecha(rec.get('fecha_sms_inicio')),
            'nivel': 'ccpp',
            'metodo_resolucion': res['metodo'],
        })
    return zones, unresolved


def ccpp_key(zone):
    return '%s:%s' % (zone['ubigeo'], normalize(zone['ccpp']))


def _most_common(values):
    vals = [v for v in values if v]
    if not vals:
        return ''
    counts = {}
    for v in vals:
        counts[v] = counts.get(v, 0) + 1
    return max(counts.items(), key=lambda kv: (kv[1], kv[0]))[0]


# ═══════════════════════════════════════════════════════════════════════════
# CAMPAÑA (JSON de salida)
# ═══════════════════════════════════════════════════════════════════════════

def campaign_path(data_dir, comunicado, override=None):
    if override:
        return Path(override)
    return Path(data_dir) / ('zonas_apagadas_%s.json' % comunicado)


def dedupe_path(data_dir, comunicado):
    return Path(data_dir) / ('.ingest_dedupe_%s.json' % comunicado)


def build_campaign(comunicado, version, source_path, md5, sha256, modo,
                   zones, unresolved, total_filas):
    for z in zones:
        z['comunicado'] = comunicado
        z.setdefault('nivel', 'ccpp')
    keys = sorted({ccpp_key(z) for z in zones})
    return {
        'comunicado': comunicado,
        'version': version,
        'fecha_generacion': now_iso(),
        'fuente': {
            'archivo': str(source_path),
            'md5': md5,
            'sha256': sha256,
        },
        'modo': modo,
        'total_filas': total_filas,
        'total_resueltas': len(zones),
        'total_unresolved': len(unresolved),
        'ccpp_keys': keys,
        'zones': zones,
        'unresolved': unresolved,
    }


def merge_campaigns(prev, new, modo):
    """union: conserva lo anterior + lo nuevo.  replace: solo lo nuevo."""
    if modo == 'replace' or not prev:
        return new
    zones = list(prev.get('zones') or [])
    have = {ccpp_key(z) for z in zones}
    for z in new.get('zones') or []:
        if ccpp_key(z) not in have:
            zones.append(z)
            have.add(ccpp_key(z))
    unresolved = list(prev.get('unresolved') or [])
    seen_u = {(u.get('depto'), u.get('prov'), u.get('dist'), u.get('ccpp'))
              for u in unresolved}
    for u in new.get('unresolved') or []:
        k = (u.get('depto'), u.get('prov'), u.get('dist'), u.get('ccpp'))
        if k not in seen_u:
            unresolved.append(u)
            seen_u.add(k)
    merged = dict(new)
    merged['zones'] = zones
    merged['unresolved'] = unresolved
    merged['total_resueltas'] = len(zones)
    merged['total_unresolved'] = len(unresolved)
    merged['ccpp_keys'] = sorted({ccpp_key(z) for z in zones})
    merged['modo'] = 'union'
    return merged


# ═══════════════════════════════════════════════════════════════════════════
# INYECCIÓN INLINE EN index.html
# ═══════════════════════════════════════════════════════════════════════════

ZONAS_MARKER = 'const ZONAS_APAGADAS = {'
INTEL_MARKER = 'const ZONAS_APAGADAS_INTEL = {'


def _scan_balanced(text, i):
    """Índice del '}' que cierra el objeto que abre en text[i] (o None)."""
    if i >= len(text) or text[i] != '{':
        return None
    depth = 0
    n = len(text)
    while i < n:
        ch = text[i]
        if ch in '"\'':
            quote = ch
            i += 1
            while i < n:
                if text[i] == '\\':
                    i += 2
                    continue
                if text[i] == quote:
                    break
                i += 1
            if i >= n:
                return None
        elif ch == '`':
            i += 1
            while i < n:
                if text[i] == '\\':
                    i += 2
                    continue
                if text[i] == '`':
                    break
                i += 1
            if i >= n:
                return None
        elif ch == '{':
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0:
                return i
        i += 1
    return None


def _find_js_object(text, marker):
    """(inicio, fin) del objeto literal JS que sigue al marker."""
    i = text.find(marker)
    if i < 0:
        return None
    b = text.find('{', i)
    if b < 0:
        return None
    e = _scan_balanced(text, b)
    if e is None:
        return None
    return b, e


def _parse_set_literal(block, name):
    """Extrae los items de `name: new Set([...])` dentro de block."""
    m = re.search(r'\b%s\s*:\s*new Set\(\s*\[' % re.escape(name), block)
    if not m:
        return None
    lb = block.find('[', m.start())
    rb = _scan_array(block, lb)
    if rb is None:
        return None
    arr_text = block[lb + 1:rb]
    items = [a or b for a, b in re.findall(r'"([^"]*)"|\'([^\']*)\'', arr_text)]
    return items, (lb, rb), arr_text


def _scan_array(text, i):
    if i >= len(text) or text[i] != '[':
        return None
    depth = 0
    n = len(text)
    while i < n:
        ch = text[i]
        if ch in '"\'':
            quote = ch
            i += 1
            while i < n:
                if text[i] == '\\':
                    i += 2
                    continue
                if text[i] == quote:
                    break
                i += 1
            if i >= n:
                return None
        elif ch == '[':
            depth += 1
        elif ch == ']':
            depth -= 1
            if depth == 0:
                return i
        i += 1
    return None


def _detect_sep(arr_text):
    """Detecta el separador original entre items (', ' vs ',') para no ensuciar el diff."""
    m = re.search(r'"\s*,\s*"', arr_text)
    if not m:
        return ','
    return m.group(0)[1:-1]


def _set_literal(items, sep=','):
    return 'new Set([%s])' % sep.join('"%s"' % it for it in items)


def _replace_in_block(block, name, new_literal):
    """Reemplaza `name: new Set([...])` dentro de block por new_literal."""
    m = re.search(r'(\b%s\s*:\s*)new Set\(\s*\[' % re.escape(name), block)
    if not m:
        return None
    lb = block.find('[', m.start())
    rb = _scan_array(block, lb)
    if rb is None:
        return None
    rp = block.find(')', rb)          # cierre de new Set(...)
    if rp < 0:
        return None
    return block[:m.start(1)] + m.group(1) + new_literal + block[rp + 1:]


def _set_number(block, name, value):
    m = re.search(r'(\b%s\s*:\s*)(\d+)' % re.escape(name), block)
    if not m:
        return None
    return block[:m.start(2)] + str(value) + block[m.end(2):]


def _set_string(block, name, value):
    m = re.search(r'(\b%s\s*:\s*)("(?:[^"\\]|\\.)*"|\'(?:[^\'\\]|\\.)*\')'
                  % re.escape(name), block)
    if not m:
        return None
    return block[:m.start(2)] + json.dumps(value, ensure_ascii=False) + block[m.end(2):]


def _find_container(block):
    """(nombre, inicio_objeto, fin_objeto) del contenedor campanas/comunicados."""
    for name in INTEL_CONTAINERS:
        m = re.search(r'\b%s\s*:\s*' % re.escape(name), block)
        if not m:
            continue
        b = block.find('{', m.end())
        if b < 0:
            continue
        e = _scan_balanced(block, b)
        if e is None:
            continue
        return name, b, e
    return None, None, None


def _merge_campana_entry(existing_literal, payload):
    """Mergea payload dentro del literal de la campaña, preservando el resto."""
    try:
        obj = json.loads(existing_literal)
        if not isinstance(obj, dict):
            raise ValueError('no es objeto')
    except Exception:
        # Literal no parseable (comentarios JS) -> cirugía por campo.
        out = existing_literal
        for k, v in payload.items():
            pat = re.compile(r'("%s"\s*:\s*)("(?:[^"\\]|\\.)*"|\[[^\]]*\]|null|true|false|-?\d+)'
                             % re.escape(k))
            if pat.search(out):
                out = pat.sub(lambda m: m.group(1) + json.dumps(v, ensure_ascii=False),
                              out, count=1)
            else:
                out = out[:-1].rstrip() + ', "%s": %s}' % (
                    k, json.dumps(v, ensure_ascii=False))
        return out
    merged = dict(obj)
    merged.update(payload)
    return json.dumps(merged, ensure_ascii=False)


def apply_inline(html, campaign, modo, fuente, fecha, comunicados_container=None):
    """Aplica los cambios al HTML.  Devuelve (html_nuevo, resumen).

    Lanza IngestError(code=2) si no encuentra un patrón EXACTO.
    """
    resumen = {}
    # ── 1) bloque ZONAS_APAGADAS ─────────────────────────────────────────
    span = _find_js_object(html, ZONAS_MARKER)
    if not span:
        raise IngestError(
            'No se encontró el patrón exacto "%s" en index.html.  '
            'ABORTA sin escribir.' % ZONAS_MARKER, code=2)
    b0, e0 = span
    block = html[b0:e0 + 1]

    parsed = _parse_set_literal(block, 'ccpp_keys')
    if not parsed:
        raise IngestError(
            'No se encontró "ccpp_keys: new Set([...])" dentro de %s.  '
            'ABORTA sin escribir.' % ZONAS_MARKER, code=2)
    ccpp_items = parsed[0]
    sep_keys = _detect_sep(parsed[2])
    parsed_d = _parse_set_literal(block, 'districts_with_off')
    if not parsed_d:
        raise IngestError(
            'No se encontró "districts_with_off: new Set([...])" dentro de %s.  '
            'ABORTA sin escribir.' % ZONAS_MARKER, code=2)
    dist_items = parsed_d[0]
    sep_dists = _detect_sep(parsed_d[2])

    new_keys = list(campaign.get('ccpp_keys') or [])
    if modo == 'replace':
        final_keys = sorted(set(new_keys))
    else:
        final_keys = sorted(set(ccpp_items) | set(new_keys))

    new_d6 = {k.split(':', 1)[0] for k in final_keys if ':' in k}
    if modo == 'replace':
        final_dists = sorted(new_d6)
    else:
        final_dists = sorted(set(dist_items) | new_d6)

    block2 = _replace_in_block(block, 'ccpp_keys', _set_literal(final_keys, sep_keys))
    if block2 is None:
        raise IngestError('Fallo interno reescribiendo ccpp_keys.', code=2)
    block2 = _replace_in_block(block2, 'districts_with_off',
                               _set_literal(final_dists, sep_dists))
    if block2 is None:
        raise IngestError('Fallo interno reescribiendo districts_with_off.', code=2)
    block2 = _set_number(block2, 'total_zones', len(final_keys))
    if block2 is None:
        raise IngestError('No se encontró "total_zones" dentro de %s.  '
                          'ABORTA sin escribir.' % ZONAS_MARKER, code=2)

    # total_districts: SOLO se recalcula si en el bloque original era igual a
    # len(districts_with_off).  Si significa otra cosa (ej. 1821), se preserva
    # para no cambiar la semántica en silencio.
    total_districts_antes = None
    mtd = re.search(r'\btotal_districts\s*:\s*(\d+)', block)
    if mtd:
        total_districts_antes = int(mtd.group(1))
    if total_districts_antes is not None and total_districts_antes == len(dist_items):
        block2 = _set_number(block2, 'total_districts', len(final_dists))
        resumen['total_districts'] = 'recalculado -> %d' % len(final_dists)
    else:
        resumen['total_districts'] = 'preservado (%s) — no equivale a len(districts_with_off)=%d' % (
            total_districts_antes, len(dist_items))

    block2 = _set_string(block2, 'fuente', fuente)
    if block2 is None:
        raise IngestError('No se encontró "fuente" dentro de %s.  '
                          'ABORTA sin escribir.' % ZONAS_MARKER, code=2)
    block2 = _set_string(block2, 'fecha', fecha)
    if block2 is None:
        raise IngestError('No se encontró "fecha" dentro de %s.  '
                          'ABORTA sin escribir.' % ZONAS_MARKER, code=2)

    html2 = html[:b0] + block2 + html[e0 + 1:]

    # ── 2) bloque ZONAS_APAGADAS_INTEL → campanas[COMUNICADO] ────────────
    span2 = _find_js_object(html2, INTEL_MARKER)
    if not span2:
        raise IngestError(
            'No se encontró el patrón exacto "%s" en index.html.  '
            'ABORTA sin escribir.' % INTEL_MARKER, code=2)
    i0, i1 = span2
    iblock = html2[i0:i1 + 1]

    cont_name, cb, ce = _find_container(iblock)
    if comunicados_container:
        cont_name = comunicados_container
    if cont_name is None:
        raise IngestError(
            'Dentro de %s no se encontró ningún contenedor de campañas '
            '(se probaron: %s).  ABORTA sin escribir.'
            % (INTEL_MARKER, ', '.join(INTEL_CONTAINERS)), code=2)
    resumen['contenedor_campanas'] = cont_name

    payload = {
        'ccpp_set': sorted(campaign.get('ccpp_keys') or []),
        'ccpp_count': len(campaign.get('ccpp_keys') or []),
        'estado': 'zonas_conocidas',
        'fecha_apagado': _most_common(
            [z.get('fecha_apagado') for z in campaign.get('zones') or []]),
        'fecha_sms_inicio': _most_common(
            [z.get('fecha_sms_inicio') for z in campaign.get('zones') or []]),
        'origen_pdf': os.path.basename(campaign.get('fuente', {}).get('archivo', '')),
    }
    payload = {k: v for k, v in payload.items()}
    if not payload['fecha_apagado']:
        payload['fecha_apagado'] = fecha
    if not payload['fecha_sms_inicio']:
        payload['fecha_sms_inicio'] = ''

    cont = iblock[cb:ce + 1]
    key_re = re.compile(r'"%s"\s*:' % re.escape(campaign['comunicado']))
    km = key_re.search(cont)
    if km:
        lb = cont.find('{', km.end())
        if lb >= 0:
            rb = _scan_balanced(cont, lb)
            if rb is None:
                raise IngestError('No se pudo cerrar el objeto de la campaña %s.  '
                                  'ABORTA sin escribir.' % campaign['comunicado'], code=2)
            old_lit = cont[lb:rb + 1]
            new_lit = _merge_campana_entry(old_lit, payload)
            cont_new = cont[:lb] + new_lit + cont[rb + 1:]
            resumen['campana'] = 'actualizada (%s)' % campaign['comunicado']
        else:
            raise IngestError('La campaña %s existe pero no es un objeto.  '
                              'ABORTA sin escribir.' % campaign['comunicado'], code=2)
    else:
        entry = '"%s": %s,' % (campaign['comunicado'],
                               json.dumps(payload, ensure_ascii=False))
        cont_new = cont[:1] + entry + cont[1:]
        resumen['campana'] = 'insertada (%s)' % campaign['comunicado']

    iblock2 = iblock[:cb] + cont_new + iblock[ce + 1:]
    html3 = html2[:i0] + iblock2 + html2[i1 + 1:]

    resumen['ccpp_keys_antes'] = len(ccpp_items)
    resumen['ccpp_keys_despues'] = len(final_keys)
    resumen['ccpp_keys_nuevas'] = len(set(final_keys) - set(ccpp_items))
    resumen['ccpp_keys_quitadas'] = len(set(ccpp_items) - set(final_keys))
    resumen['districts_with_off_antes'] = len(dist_items)
    resumen['districts_with_off_despues'] = len(final_dists)
    return html3, resumen


def make_backup(html_path, comunicado):
    """Copia atómica de index.html en .backup-pre-ingest/."""
    html_path = Path(html_path)
    ts = _dt.datetime.now().strftime('%Y%m%d-%H%M%S')
    bdir = html_path.parent / '.backup-pre-ingest'
    bdir.mkdir(parents=True, exist_ok=True)
    dest = bdir / ('ingest-%s-%s.html' % (comunicado, ts))
    tmp = dest.with_suffix(dest.suffix + '.tmp')
    shutil.copy2(str(html_path), str(tmp))
    os.replace(str(tmp), str(dest))
    return dest


def write_atomic(path, text):
    """Escritura atómica (tmp + os.replace)."""
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_name(path.name + '.tmp-ingest')
    with open(tmp, 'w', encoding='utf-8') as fh:
        fh.write(text)
        fh.flush()
        os.fsync(fh.fileno())
    os.replace(str(tmp), str(path))


# ═══════════════════════════════════════════════════════════════════════════
# CLI
# ═══════════════════════════════════════════════════════════════════════════

def build_parser():
    p = argparse.ArgumentParser(
        prog='ingest_comunicado_zonas.py',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=(
            'Ingesta zonas de apagado (CCPP) de un comunicado COMU-XXXX.\n'
            'Por defecto NO escribe nada (dry-run).  Use --apply-inline para\n'
            'escribir la campaña en data/ y actualizar index.html.'),
        epilog=(
            'Ejemplos:\n'
            '  # reporte sin escribir nada\n'
            '  python3 scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 \\\n'
            '      --source "Recursos CIMA/Bases y Zonas de Apagado/APAGADO.pdf"\n\n'
            '  # aplicar (escribe data/ + index.html con backup previo)\n'
            '  python3 scripts/ingest_comunicado_zonas.py --comunicado COMU-5569 \\\n'
            '      --source zonas.csv --apply-inline\n\n'
            'Códigos de salida: 0=ok · 1=error · 2=patrón no encontrado (aborta)\n'
            '                    3=dependencia opcional ausente\n'))
    p.add_argument('--comunicado', required=True,
                   help='Código del comunicado, ej. COMU-5569')
    p.add_argument('--source', required=True,
                   help='Archivo fuente: .csv · .xlsx · .pdf · .json')
    p.add_argument('--modo', choices=('union', 'replace'), default='union',
                   help='union (default) suma a lo existente; replace reemplaza')
    p.add_argument('--dry-run', action='store_true',
                   help='Fuerza dry-run (no escribe nada).  Es el default.')
    p.add_argument('--apply-inline', action='store_true',
                   help='ESCRIBE: data/zonas_apagadas_<COM>.json + index.html '
                        '(con backup) + estado de dedupe')
    p.add_argument('--campana-json', default=None,
                   help='Ruta alterna para el JSON de campaña '
                        '(default: data/zonas_apagadas_<COMUNICADO>.json).  '
                        'Solo se escribe con --apply-inline.')
    p.add_argument('--comunicado-version', default='00001',
                   help='Versión del comunicado (default: 00001)')
    p.add_argument('--fecha-apagado', default=None,
                   help='Fecha de apagado forzada (YYYY-MM-DD)')
    p.add_argument('--fecha-sms-inicio', default=None,
                   help='Fecha de inicio de SMS forzada (YYYY-MM-DD)')
    p.add_argument('--data-dir', default=str(DEFAULT_DATA_DIR),
                   help='Directorio de datos (default: %(default)s)')
    p.add_argument('--html-path', default=str(DEFAULT_HTML_PATH),
                   help='index.html a actualizar (default: %(default)s)')
    p.add_argument('--ubigeo-path', default=str(DEFAULT_UBIGEO_PATH),
                   help='ubigeo_consolidated.json (default: %(default)s)')
    p.add_argument('--contenedor-campanas', choices=INTEL_CONTAINERS, default=None,
                   help='Forzar el contenedor dentro de ZONAS_APAGADAS_INTEL '
                        '(default: auto-detecta campanas/comunicados)')
    return p


def main(argv=None):
    argv = list(sys.argv[1:] if argv is None else argv)
    parser = build_parser()
    args = parser.parse_args(argv)

    if args.dry_run and args.apply_inline:
        parser.error('--dry-run y --apply-inline son excluyentes')

    source = Path(args.source).expanduser()
    if not source.exists():
        print('ERROR: no existe el archivo fuente: %s' % source, file=sys.stderr)
        return 1
    if not source.is_file():
        print('ERROR: la fuente no es un archivo: %s' % source, file=sys.stderr)
        return 1

    data_dir = Path(args.data_dir)
    html_path = Path(args.html_path)
    out_json = campaign_path(data_dir, args.comunicado, args.campana_json)
    dedupe_file = dedupe_path(data_dir, args.comunicado)

    try:
        md5, sha256, size = file_hashes(source)
        ubigeo = load_ubigeo(Path(args.ubigeo_path))

        # ── dedupe por md5 ───────────────────────────────────────────────
        state = {}
        if dedupe_file.exists():
            try:
                state = json.loads(dedupe_file.read_text(encoding='utf-8'))
            except Exception:
                state = {}
        ya_ingerido = (state.get('md5') == md5)
        campana_previa = None
        if out_json.exists():
            try:
                Campana_previa = json.loads(out_json.read_text(encoding='utf-8'))
            except Exception:
                campana_previa = None

        notas = []
        if ya_ingerido and campana_previa:
            # Idempotencia: la misma fuente ya fue procesada.  Se reutiliza la
            # campaña previa (no se re-parsea el PDF) -> 0 duplicados.
            campana = campana_previa
            campana['modo'] = args.modo
            formato = 'cache(md5)'
            filas = campana.get('total_filas', 0)
            reuso = True
        else:
            records, formato, notas = read_rows(source)
            filas = len(records)
            zones, unresolved = resolve_rows(records, ubigeo)
            if args.fecha_apagado:
                for z in zones:
                    z['fecha_apagado'] = norm_fecha(args.fecha_apagado)
                for u in unresolved:
                    u['fecha_apagado'] = norm_fecha(args.fecha_apagado)
            if args.fecha_sms_inicio:
                for z in zones:
                    z['fecha_sms_inicio'] = norm_fecha(args.fecha_sms_inicio)
                for u in unresolved:
                    u['fecha_sms_inicio'] = norm_fecha(args.fecha_sms_inicio)
            campana = build_campaign(
                args.comunicado, args.comunicado_version, source, md5, sha256,
                args.modo, zones, unresolved, filas)
            campana = merge_campaigns(campana_previa, campana, args.modo)
            reuso = False

        # ── reporte ──────────────────────────────────────────────────────
        print('═' * 72)
        print('INGESTA DE ZONAS DE APAGADO — %s' % args.comunicado)
        print('═' * 72)
        print('Fuente            : %s' % source)
        print('Formato           : %s%s' % (formato, '  (reuso por md5)' if reuso else ''))
        print('Tamaño            : %d bytes' % size)
        print('md5               : %s' % md5)
        print('sha256            : %s' % sha256)
        print('Modo              : %s' % args.modo)
        print('Filas leídas      : %d' % filas)
        print('Resueltas         : %d' % campana.get('total_resueltas', 0))
        print('Unresolved        : %d' % campana.get('total_unresolved', 0))
        print('ccpp_keys         : %d' % len(campana.get('ccpp_keys') or []))
        if ya_ingerido:
            print('Dedupe md5        : FUENTE YA INGERIDA (idempotente, 0 duplicados)')
        else:
            print('Dedupe md5        : fuente nueva')
        for n in notas:
            print('Nota              : %s' % n)
        if campana.get('unresolved'):
            print('─' * 72)
            print('UNRESOLVED (no se inventaron, quedan con motivo):')
            for u in campana['unresolved']:
                print('  · %s/%s/%s · %s → %s' % (u.get('depto'), u.get('prov'),
                                                  u.get('dist'), u.get('ccpp'),
                                                  u.get('motivo')))
        if campana.get('ccpp_keys'):
            print('─' * 72)
            print('CCPP_KEYS (%d):' % len(campana['ccpp_keys']))
            for k in campana['ccpp_keys']:
                print('  · %s' % k)

        # ── aplicación inline (en memoria primero) ───────────────────────
        html_nuevo = None
        resumen = {}
        if args.apply_inline:
            if not html_path.exists():
                raise IngestError('No existe index.html: %s' % html_path)
            html_actual = html_path.read_text(encoding='utf-8')
            fecha_camp = (args.fecha_apagado
                          or _most_common([z.get('fecha_apagado')
                                           for z in campana.get('zones') or []])
                          or _dt.date.today().isoformat())
            if args.modo == 'replace':
                fuente_txt = '%s v%s (%s)' % (args.comunicado,
                                              args.comunicado_version, fecha_camp)
            else:
                m = re.search(r'fuente\s*:\s*"([^"]*)"', html_actual)
                base = m.group(1) if m else 'Zonas de apagado'
                tag = '%s v%s (%s)' % (args.comunicado, args.comunicado_version,
                                      fecha_camp)
                # idempotencia: no duplicar el tag si ya está en la fuente
                fuente_txt = base if tag in base else '%s + %s' % (base, tag)
            html_nuevo, resumen = apply_inline(
                html_actual, campana, args.modo, fuente_txt, fecha_camp,
                args.contenedor_campanas)
            print('─' * 72)
            print('INLINE (aplicado en memoria, aún NO escrito):')
            for k in ('contenedor_campanas', 'campana', 'ccpp_keys_antes',
                      'ccpp_keys_despues', 'ccpp_keys_nuevas', 'ccpp_keys_quitadas',
                      'districts_with_off_antes', 'districts_with_off_despues',
                      'total_districts'):
                if k in resumen:
                    print('  %-28s : %s' % (k, resumen[k]))
            if html_nuevo == html_actual:
                print('  (sin cambios: la campaña ya estaba inyectada)')

        print('─' * 72)
        if args.apply_inline:
            print('MODO: APPLY — se escribirán:')
            print('  · %s' % out_json)
            print('  · %s' % dedupe_file)
            print('  · %s' % html_path)
            # escritura
            backup = make_backup(html_path, args.comunicado)
            print('  · backup: %s' % backup)
            write_atomic(html_path, html_nuevo)
            write_atomic(out_json, json.dumps(campana, ensure_ascii=False, indent=2) + '\n')
            state = {
                'comunicado': args.comunicado,
                'md5': md5,
                'sha256': sha256,
                'fecha': now_iso(),
                'filas': filas,
                'ccpp_agregados': len(campana.get('ccpp_keys') or []),
                'archivo': str(source),
                'historial': (state.get('historial') or []) + [{
                    'md5': md5, 'fecha': now_iso(), 'filas': filas,
                    'ccpp_agregados': len(campana.get('ccpp_keys') or []),
                }],
            }
            write_atomic(dedupe_file, json.dumps(state, ensure_ascii=False, indent=2) + '\n')
            print('✓ ESCRITO.')
        else:
            print('MODO: DRY-RUN — no se escribió NADA.')
            print('  (para escribir: agrega --apply-inline)')
            print('  destinos previstos: %s | %s | %s'
                  % (out_json, dedupe_file, html_path))
        print('═' * 72)
        return 0

    except IngestError as exc:
        print('ERROR: %s' % exc, file=sys.stderr)
        return exc.code
    except KeyboardInterrupt:
        print('ERROR: interrumpido por el usuario', file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())

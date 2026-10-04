#!/usr/bin/env python3
"""Render Feedbacks tab as PNG with new master-detail layout."""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/feedbacks_render_v3.png"
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

# Workflow state colors
ESTADO_COLORS = {
    "Pendiente":            (235, 235, 235),  # gris
    "En evaluación":        (255, 255, 150),  # amarillo
    "Feedback dado":        (180, 220, 255),  # azul claro
    "Carta de compromiso":  (255, 220, 130),  # ámbar
    "Cumplido":             (170, 230, 170),  # verde
    "Vencido":              (255, 150, 150),  # rojo
}

C = {
    "white":     (255, 255, 255),
    "header":    (26, 58, 92),
    "header_fg": (255, 255, 255),
    "title":     (212, 227, 243),
    "text":      (15, 36, 64),
    "muted":     (240, 240, 240),
    "border":    (180, 180, 180),
}

# Layout post-restructure:
# A=N°(30) | B=AGENTE(160) | C=SUPERVISOR(160) | D=CANT(40) | E=ULT_TIPO(110)
# F=ULT_NOTA(180) | G=RIESGO(60) | H=F_BASE(70) | I=F_EVAL(70) | J=ESTADO(120)
COL_WIDTHS = [30, 160, 160, 40, 110, 180, 60, 70, 70, 120]
ROW_HEIGHTS = {1: 32, 2: 22, 3: 8, 4: 28, 13: 28, 14: 28}

# Detail section starts at row 13 (title) + 14 (header), data 15-28 + master 5-11
# Detail widths: A=ID(30) | B=N°(30) | C=FECHA(80) | D=SUPERVISOR(160) | E=AGENTE(160)
# F=CALL_ID(110) | G=TIPO(110) | H=NOTA(180) | I=F_BASE(70) | J=F_COMPROMISO(70) | K=F_CUMP(70) | L=ESTADO(120)
DETAIL_COL_WIDTHS = [30, 30, 80, 160, 160, 110, 110, 180, 70, 70, 70, 120]


def get_token():
    from cryptography.hazmat.primitives.ciphers.aead import AESGCM
    key = open(KEY_PATH, "rb").read()
    enc = open(VAULT_PATH, "rb").read()
    pt = AESGCM(key).decrypt(enc[:12], enc[12:], None)
    creds = json.loads(pt)
    body = json.dumps({
        "client_id": creds["client_id"],
        "client_secret": creds["client_secret"],
        "refresh_token": creds["refresh_token"],
        "grant_type": "refresh_token",
    }).encode()
    req = urllib.request.Request("https://oauth2.googleapis.com/token", data=body,
        headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["access_token"]


def fetch(token):
    fields = ",".join([
        "sheets.data.rowData.values(formattedValue,userEnteredValue,userEnteredFormat.backgroundColor,userEnteredFormat.textFormat,userEnteredFormat.horizontalAlignment)",
    ])
    url = (f"https://sheets.googleapis.com/v4/spreadsheets/{SPREADSHEET_ID}"
           f"?includeGridData=true&ranges=Feedbacks!A1:L45&fields={fields}")
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["sheets"][0]


def get_font(size, bold=False):
    candidates = [
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
        "/usr/share/fonts/TTF/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/TTF/DejaVuSans.ttf",
    ]
    for path in candidates:
        if os.path.exists(path):
            return ImageFont.truetype(path, size)
    return ImageFont.load_default()


def truncate(draw, text, font, max_w):
    if not text:
        return ""
    s = str(text)
    while draw.textlength(s, font=font) > max_w and len(s) > 1:
        s = s[:-1]
    if s != str(text):
        s = s[:-1] + "…"
    return s


def render():
    token = get_token()
    sheet = fetch(token)
    data = sheet["data"][0]
    rows_data = data.get("rowData", [])

    # Calculate dimensions
    sec1_w = sum(COL_WIDTHS)
    detail_w = sum(DETAIL_COL_WIDTHS)
    total_w = max(sec1_w, detail_w) + 20
    total_h = 0
    for i in range(1, 46):
        total_h += ROW_HEIGHTS.get(i, 18)
    total_h += 30  # padding

    img = Image.new("RGB", (total_w, total_h), C["white"])
    draw = ImageDraw.Draw(img)

    fonts = {
        "title": get_font(16, bold=True),
        "subtitle": get_font(11),
        "section": get_font(13, bold=True),
        "header": get_font(10, bold=True),
        "data": get_font(10),
        "data_bold": get_font(10, bold=True),
    }

    y = 10

    # === SECCIÓN 1: Master (rows 1-11) ===
    for row_idx in range(0, 11):
        actual_row = row_idx + 1
        h = ROW_HEIGHTS.get(actual_row, 18)
        x = 10

        row_cells = (rows_data[row_idx].get("values", []) if row_idx < len(rows_data) else [])

        if actual_row == 1:
            # Title merged across all cols
            text = row_cells[0]["formattedValue"] if row_cells else "FEEDBACKS — Auditoría semanal"
            draw.rectangle([x, y, x + sec1_w, y + h], fill=C["title"])
            draw.text((x + 6, y + 8), str(text), fill=C["header"], font=fonts["title"])
        elif actual_row == 2:
            # Subtitle
            text = row_cells[0]["formattedValue"] if row_cells else ""
            draw.rectangle([x, y, x + sec1_w, y + h], fill=C["white"])
            draw.text((x + 6, y + 4), str(text)[:120], fill=C["text"], font=fonts["subtitle"])
        elif actual_row == 3:
            draw.rectangle([x, y, x + sec1_w, y + h], fill=C["white"])
        elif actual_row == 4:
            # Header row
            draw.rectangle([x, y, x + sec1_w, y + h], fill=C["header"])
            for col_idx, w in enumerate(COL_WIDTHS):
                cell_text = row_cells[col_idx]["formattedValue"] if col_idx < len(row_cells) else ""
                tx = truncate(draw, str(cell_text), fonts["header"], w - 8)
                draw.text((x + 4, y + 6), tx, fill=C["header_fg"], font=fonts["header"])
                draw.rectangle([x, y, x + w, y + h], outline=C["border"], width=1)
                x += w
        else:
            # Data row
            # Find Estado (col J idx=9) for color
            estado_text = ""
            if len(row_cells) > 9 and row_cells[9].get("formattedValue"):
                estado_text = str(row_cells[9]["formattedValue"])
            row_bg = ESTADO_COLORS.get(estado_text, C["white"])

            for col_idx, w in enumerate(COL_WIDTHS):
                cell = row_cells[col_idx] if col_idx < len(row_cells) else {}
                val = cell.get("formattedValue", "")
                # Highlight ESTADO col
                cell_bg = row_bg if col_idx == 9 else C["white"]
                draw.rectangle([x, y, x + w, y + h], fill=cell_bg, outline=C["border"], width=1)
                tx = truncate(draw, str(val), fonts["data"], w - 8)
                font = fonts["data_bold"] if col_idx in [0, 9] else fonts["data"]
                draw.text((x + 4, y + 4), tx, fill=C["text"], font=font)
                x += w
        y += h

    # Gap
    y += 12
    draw.text((10, y), "DETALLE DE LLAMADAS (master-detail con FK real):", fill=C["header"], font=fonts["section"])
    y += 22

    # === SECCIÓN 2: Detail (rows 13-40) ===
    # row 13 is title, row 14 is header, rows 15-40 are data
    for row_idx in range(12, 45):  # 0-based
        actual_row = row_idx + 1
        if actual_row > 45:
            break
        h = ROW_HEIGHTS.get(actual_row, 18)
        x = 10

        row_cells = (rows_data[row_idx].get("values", []) if row_idx < len(rows_data) else [])

        if actual_row == 13:
            text = row_cells[0]["formattedValue"] if row_cells else ""
            draw.rectangle([x, y, x + detail_w, y + h], fill=C["title"])
            draw.text((x + 6, y + 4), str(text)[:120], fill=C["header"], font=fonts["header"])
        elif actual_row == 14:
            draw.rectangle([x, y, x + detail_w, y + h], fill=C["header"])
            for col_idx, w in enumerate(DETAIL_COL_WIDTHS):
                cell = row_cells[col_idx] if col_idx < len(row_cells) else {}
                val = cell.get("formattedValue", "")
                tx = truncate(draw, str(val), fonts["header"], w - 8)
                draw.rectangle([x, y, x + w, y + h], outline=C["border"], width=1)
                draw.text((x + 4, y + 6), tx, fill=C["header_fg"], font=fonts["header"])
                x += w
        elif 15 <= actual_row <= 40:
            # Data row
            estado_text = ""
            if len(row_cells) > 11 and row_cells[11].get("formattedValue"):
                estado_text = str(row_cells[11]["formattedValue"])
            row_bg = ESTADO_COLORS.get(estado_text, C["white"])

            for col_idx, w in enumerate(DETAIL_COL_WIDTHS):
                cell = row_cells[col_idx] if col_idx < len(row_cells) else {}
                val = cell.get("formattedValue", "")
                cell_bg = row_bg if col_idx == 11 else C["white"]
                draw.rectangle([x, y, x + w, y + h], fill=cell_bg, outline=C["border"], width=1)
                tx = truncate(draw, str(val), fonts["data"], w - 8)
                font = fonts["data_bold"] if col_idx in [0, 1, 11] else fonts["data"]
                draw.text((x + 4, y + 4), tx, fill=C["text"], font=font)
                x += w
        else:
            # Empty
            draw.rectangle([x, y, x + detail_w, y + h], fill=C["white"], outline=C["border"], width=1)

        y += h

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT)
    print(f"✅ Rendered: {OUTPUT}")
    print(f"   Dimensions: {img.size[0]}x{img.size[1]} px")
    return OUTPUT


if __name__ == "__main__":
    render()
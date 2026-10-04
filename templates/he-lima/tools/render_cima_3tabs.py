#!/usr/bin/env python3
"""Render the 3-tab Feedbacks architecture (FEEDBACK + FEEDBACK-DETALLE_LLAMADA + CONFIG)."""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/cima_3tabs_render_v3.png"
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

ESTADO_COLORS = {
    "Pendiente":            (235, 235, 235),
    "En evaluación":        (255, 255, 150),
    "Feedback dado":        (180, 220, 255),
    "Carta de compromiso":  (255, 220, 130),
    "Cumplido":             (170, 230, 170),
    "Vencido":              (255, 150, 150),
    "Con feedback":         (180, 220, 255),
    "Auditada":             (200, 230, 200),
    "Positivo":             (170, 230, 170),
    "Negativo":             (255, 150, 150),
    "Evaluativo":           (180, 220, 255),
    "Amonestador":          (255, 165, 110),
    "Estados incorrectos":  (255, 180, 180),
    "Abandono llamada":     (255, 220, 130),
    "Speech incompleto":    (180, 220, 255),
    "No devuelve llamada":  (255, 180, 255),
}

C = {
    "white":      (255, 255, 255),
    "title_bg":   (26, 58, 92),       # dark blue
    "title_fg":   (255, 255, 255),
    "subtitle_bg":(212, 227, 243),     # light blue
    "subtitle_fg":(26, 58, 92),
    "header_bg":  (26, 58, 92),       # dark blue
    "header_fg":  (255, 255, 255),
    "zebra":      (240, 244, 248),     # light gray
    "white_alt":  (255, 255, 255),
    "text":       (15, 36, 64),
    "border":     (180, 180, 180),
    "hyperlink":   (26, 115, 232),     # blue for links
    "tab_master": (26, 110, 110),
    "tab_detail": (140, 80, 160),
    "tab_config": (110, 110, 26),
}

# Master layout (FEEDBACK): A-N (14 cols)
MASTER_COL_WIDTHS = [30, 180, 160, 70, 110, 100, 90, 90, 110, 300, 110, 110, 130, 70]
MASTER_HEADERS = ["ID", "AGENTE", "SUPERVISOR", "CANT_CALLS", "TIPO_FEEDBACK", "FECHA_FEEDBACK", "FECHA_BASE", "FECHA_EVAL", "ÚLTIMO_TIPO", "ÚLTIMA_NOTA", "COMPROMISO_FECHA", "CUMPLIDO_FECHA", "ESTADO", "RIESGO"]

# Detail layout (FEEDBACK-DETALLE_LLAMADA): A-L (12 cols)
DETAIL_COL_WIDTHS = [40, 40, 100, 150, 170, 130, 130, 220, 110, 60, 90, 150]
DETAIL_HEADERS = ["ID_AGENTE", "N°", "FECHA_LLAMADA", "SUPERVISOR", "AGENTE", "CALL_ID", "TIPO_LLAMADA", "NOTA", "ESTADO_AUDITORIA", "FEEDBACK_ID", "FECHA_BASE", "AUDITOR"]


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


def fetch_sheet(token, tab_name, range_a1):
    fields = "sheets.data.rowData.values(formattedValue)"
    url = (f"https://sheets.googleapis.com/v4/spreadsheets/{SPREADSHEET_ID}"
           f"?includeGridData=true&ranges={tab_name}!{range_a1}&fields={fields}")
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["sheets"][0]["data"][0]


def get_font(size, bold=False):
    paths = [
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/usr/share/fonts/TTF/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/TTF/DejaVuSans.ttf",
    ]
    for p in paths:
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
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


def render_master(draw, x_start, y_start, data_rows):
    fonts = {
        "title": get_font(16, bold=True),
        "subtitle": get_font(10),
        "header": get_font(9, bold=True),
        "data": get_font(9),
        "data_bold": get_font(9, bold=True),
    }
    y = y_start
    row_h = 22
    col_widths = MASTER_COL_WIDTHS
    ncols = len(col_widths)
    total_w = sum(col_widths)

    # Tab label
    draw.rectangle([x_start, y, x_start + 220, y + 26], fill=C["tab_master"])
    draw.text((x_start + 10, y + 6), "FEEDBACK — Master (1 row/agente)", fill=(255, 255, 255), font=fonts["header"])
    y += 32

    # Title row 1
    draw.rectangle([x_start, y, x_start + total_w, y + 32], fill=C["title_bg"])
    title_text = data_rows[0]["values"][0]["formattedValue"] if data_rows[0]["values"] else "FEEDBACKS"
    draw.text((x_start + 10, y + 8), title_text, fill=C["title_fg"], font=fonts["title"])
    y += 32

    # Subtitle row 2
    sub_text = data_rows[1]["values"][0]["formattedValue"] if len(data_rows) > 1 and data_rows[1]["values"] else ""
    draw.rectangle([x_start, y, x_start + total_w, y + 24], fill=C["subtitle_bg"])
    draw.text((x_start + 10, y + 5), sub_text, fill=C["subtitle_fg"], font=fonts["subtitle"])
    y += 24

    # Spacer row 3
    y += 4

    # Header row 4
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], fill=C["header_bg"])
    x = x_start
    for col_idx, w in enumerate(col_widths):
        header_text = MASTER_HEADERS[col_idx]
        tx = truncate(draw, header_text, fonts["header"], w - 6)
        draw.text((x + 4, y + 6), tx, fill=C["header_fg"], font=fonts["header"])
        x += w
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], outline=C["border"], width=1)
    y += row_h + 2

    # Data rows 5-11
    for row_idx in range(4, min(11, len(data_rows))):
        actual_row = row_idx + 1
        cells = data_rows[row_idx].get("values", [])
        is_zebra = (actual_row % 2 == 1)
        bg = C["zebra"] if is_zebra else C["white_alt"]

        x = x_start
        for col_idx, w in enumerate(col_widths):
            cell = cells[col_idx] if col_idx < len(cells) else {}
            val = cell.get("formattedValue", "")

            # Highlight ESTADO col (col 12, 0-based)
            cell_bg = bg
            if col_idx == 12 and val in ESTADO_COLORS:
                cell_bg = ESTADO_COLORS[val]
            # Highlight TIPO_FEEDBACK col (col 4)
            if col_idx == 4 and val in ESTADO_COLORS:
                cell_bg = ESTADO_COLORS[val]

            # HYPERLINK styling
            fg = C["text"]
            font = fonts["data"]
            underline = False
            if col_idx in [1, 3]:  # B (AGENTE), D (CANT)
                fg = C["hyperlink"]
                underline = True

            draw.rectangle([x, y, x + w, y + row_h], fill=cell_bg, outline=C["border"], width=1)
            tx = truncate(draw, str(val), font, w - 6)
            draw.text((x + 4, y + 5), tx, fill=fg, font=font)
            x += w
        y += row_h

    return y + 10


def render_detail(draw, x_start, y_start, data_rows):
    fonts = {
        "title": get_font(16, bold=True),
        "header": get_font(9, bold=True),
        "data": get_font(9),
        "data_bold": get_font(9, bold=True),
    }
    y = y_start
    row_h = 20
    col_widths = DETAIL_COL_WIDTHS
    ncols = len(col_widths)
    total_w = sum(col_widths)

    # Tab label
    draw.rectangle([x_start, y, x_start + 280, y + 26], fill=C["tab_detail"])
    draw.text((x_start + 10, y + 6), "FEEDBACK-DETALLE_LLAMADA — Base (26 calls)", fill=(255, 255, 255), font=fonts["header"])
    y += 32

    # Title row 1
    draw.rectangle([x_start, y, x_start + total_w, y + 32], fill=C["title_bg"])
    title_text = data_rows[0]["values"][0]["formattedValue"] if data_rows[0]["values"] else "DETALLE"
    draw.text((x_start + 10, y + 8), title_text, fill=C["title_fg"], font=fonts["title"])
    y += 32

    # Header row 2
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], fill=C["header_bg"])
    x = x_start
    for col_idx, w in enumerate(col_widths):
        header_text = DETAIL_HEADERS[col_idx]
        tx = truncate(draw, header_text, fonts["header"], w - 6)
        draw.text((x + 4, y + 5), tx, fill=C["header_fg"], font=fonts["header"])
        x += w
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], outline=C["border"], width=1)
    y += row_h + 2

    # Data rows 3-28
    for row_idx in range(2, min(28, len(data_rows))):
        actual_row = row_idx + 1
        cells = data_rows[row_idx].get("values", [])
        is_zebra = (actual_row % 2 == 1)
        bg = C["zebra"] if is_zebra else C["white_alt"]

        x = x_start
        for col_idx, w in enumerate(col_widths):
            cell = cells[col_idx] if col_idx < len(cells) else {}
            val = cell.get("formattedValue", "")

            # Highlight ESTADO_AUDITORIA col (col 8)
            cell_bg = bg
            if col_idx == 8 and val in ESTADO_COLORS:
                cell_bg = ESTADO_COLORS[val]
            # Highlight TIPO_LLAMADA col (col 6)
            if col_idx == 6 and val in ESTADO_COLORS:
                cell_bg = ESTADO_COLORS[val]

            fg = C["text"]
            font = fonts["data"]
            if col_idx in [1, 4, 9]:  # B (N°), E (AGENTE), J (FEEDBACK_ID)
                fg = C["hyperlink"]

            draw.rectangle([x, y, x + w, y + row_h], fill=cell_bg, outline=C["border"], width=1)
            tx = truncate(draw, str(val), font, w - 6)
            draw.text((x + 4, y + 4), tx, fill=fg, font=font)
            x += w
        y += row_h

    return y + 10


def main():
    token = get_token()
    fb_data = fetch_sheet(token, "FEEDBACK", "A1:N11")
    detalle_data = fetch_sheet(token, "FEEDBACK-DETALLE_LLAMADA", "A1:L28")

    w = max(sum(MASTER_COL_WIDTHS), sum(DETAIL_COL_WIDTHS)) + 40
    h = 1300
    img = Image.new("RGB", (w, h), C["white"])
    draw = ImageDraw.Draw(img)

    y = 20
    title_font = get_font(20, bold=True)
    subtitle_font = get_font(11)
    draw.text((20, y), "CIMA — Sistema de Feedbacks (3 tabs integrados)", fill=C["title_bg"], font=title_font)
    y += 32
    draw.text((20, y), "FEEDBACK master + FEEDBACK-DETALLE_LLAMADA base + CONFIG metadata", fill=C["text"], font=subtitle_font)
    y += 30

    # Render master
    y = render_master(draw, 20, y, fb_data.get("rowData", []))
    y += 10

    # Render detail
    y = render_detail(draw, 20, y, detalle_data.get("rowData", []))

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT)
    print(f"✅ Rendered: {OUTPUT}")
    print(f"   Dimensions: {img.size[0]}x{img.size[1]} px")


if __name__ == "__main__":
    main()
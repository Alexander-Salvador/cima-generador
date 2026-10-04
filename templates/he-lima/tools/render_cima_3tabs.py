#!/usr/bin/env python3
"""Render the 3-tab Feedbacks architecture (FEEDBACK + FEEDBACK-DETALLE_LLAMADA + CONFIG)."""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/cima_3tabs_render_v1.png"
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
}

C = {
    "white":      (255, 255, 255),
    "header":     (26, 58, 92),
    "header_fg":  (255, 255, 255),
    "title":      (212, 227, 243),
    "text":       (15, 36, 64),
    "tab_label":  (240, 240, 250),
    "border":     (180, 180, 180),
    "tab_master": (26, 110, 110),  # teal para FEEDBACK (master)
    "tab_detail": (140, 80, 160),  # purple para FEEDBACK-DETALLE
    "tab_config": (110, 110, 26),  # olive para CONFIG
}

# Master layout (FEEDBACK): A-N (14 cols)
MASTER_COL_WIDTHS = [30, 150, 150, 60, 100, 90, 80, 80, 100, 200, 100, 100, 110, 60]
# Detail layout (FEEDBACK-DETALLE_LLAMADA): A-L (12 cols)
DETAIL_COL_WIDTHS = [40, 35, 90, 150, 160, 110, 130, 200, 110, 60, 90, 150]


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


def render_table(draw, x_start, y_start, data_rows, col_widths, tab_color, title_text, max_rows=20):
    fonts = {
        "title": get_font(14, bold=True),
        "tab": get_font(11, bold=True),
        "header": get_font(9, bold=True),
        "data": get_font(9),
        "data_bold": get_font(9, bold=True),
    }
    y = y_start
    row_h = 18

    # Tab label
    tab_text = title_text
    draw.rectangle([x_start, y, x_start + 200, y + 22], fill=tab_color)
    draw.text((x_start + 8, y + 4), tab_text, fill=(255, 255, 255), font=fonts["tab"])
    y += 28

    for row_idx in range(min(len(data_rows), max_rows)):
        actual_row = row_idx
        cells = data_rows[row_idx].get("values", []) if row_idx < len(data_rows) else []
        is_header = (row_idx == 0)

        if is_header:
            draw.rectangle([x_start, y, x_start + sum(col_widths), y + row_h + 4], fill=C["header"])
            for col_idx, w in enumerate(col_widths):
                cell = cells[col_idx] if col_idx < len(cells) else {}
                val = cell.get("formattedValue", "")
                tx = truncate(draw, str(val), fonts["header"], w - 6)
                draw.text((x_start + 3, y + 4), tx, fill=(256, 256, 256), font=fonts["header"])
                x_start += w
        else:
            # Find Estado col for highlighting
            estado_text = ""
            # Try col 9 first (detail ESTADO_AUDITORIA)
            for col_idx in [9, 12, 13]:
                if col_idx < len(cells) and cells[col_idx].get("formattedValue"):
                    val = str(cells[col_idx]["formattedValue"])
                    if val in ESTADO_COLORS:
                        estado_text = val
                        break
            row_bg = ESTADO_COLORS.get(estado_text, C["white"])

            x = x_start
            for col_idx, w in enumerate(col_widths):
                cell = cells[col_idx] if col_idx < len(cells) else {}
                val = cell.get("formattedValue", "")
                # Highlight ESTADO col
                cell_bg = row_bg if col_idx in [12, 13, 9] else C["white"]
                draw.rectangle([x, y, x + w, y + row_h], fill=cell_bg, outline=C["border"], width=1)
                tx = truncate(draw, str(val), fonts["data"], w - 6)
                font = fonts["data_bold"] if col_idx in [0, 1] else fonts["data"]
                draw.text((x + 3, y + 3), tx, fill=C["text"], font=font)
                x += w
        y += row_h + (4 if is_header else 0)
        x_start -= sum(col_widths)

    return y


def main():
    token = get_token()

    # Fetch all 3 sheets
    fb_data = fetch_sheet(token, "FEEDBACK", "A1:N11")
    detalle_data = fetch_sheet(token, "FEEDBACK-DETALLE_LLAMADA", "A1:L27")
    config_data = fetch_sheet(token, "CONFIG", "A4:C11")

    # Build image
    w = 1600
    h = 1000
    img = Image.new("RGB", (w, h), C["white"])
    draw = ImageDraw.Draw(img)

    title_font = get_font(20, bold=True)
    subtitle_font = get_font(12)

    y = 20
    # Main title
    draw.text((20, y), "CIMA — Sistema de Feedbacks (3 tabs integrados)", fill=C["header"], font=title_font)
    y += 30
    draw.text((20, y), "Feedback tab refactorizado: FEEDBACK (master) + FEEDBACK-DETALLE_LLAMADA (base) + CONFIG (metadata)", fill=C["text"], font=subtitle_font)
    y += 25

    # TAB 1: FEEDBACK (master)
    y = render_table(draw, 20, y, fb_data.get("rowData", []), MASTER_COL_WIDTHS, C["tab_master"], "FEEDBACK — Master (1 row/agente)")
    y += 20

    # TAB 2: FEEDBACK-DETALLE_LLAMADA (detail)
    y = render_table(draw, 20, y, detalle_data.get("rowData", []), DETAIL_COL_WIDTHS, C["tab_detail"], "FEEDBACK-DETALLE_LLAMADA — Base (26 calls)")
    y += 20

    # TAB 3: CONFIG (metadata)
    y = render_table(draw, 20, y, config_data.get("rowData", []), [120, 90, 280], C["tab_config"], "CONFIG — Tabla de mapeo FK→PERSONAL")
    y += 20

    # Footer
    note_font = get_font(10)
    draw.text((20, y), "Interconexión:", fill=C["header"], font=get_font(11, bold=True))
    y += 16
    draw.text((40, y), "• Click FEEDBACK.ID → cross-tab a FEEDBACK-DETALLE_LLAMADA", fill=C["text"], font=note_font)
    y += 14
    draw.text((40, y), "• Click FEEDBACK-DETALLE_LLAMADA.FEEDBACK_ID → reverse nav al master", fill=C["text"], font=note_font)
    y += 14
    draw.text((40, y), "• Click AGENTE → cross-tab a PERSONAL ficha", fill=C["text"], font=note_font)
    y += 14
    draw.text((40, y), "• VLOOKUP en CONFIG provee el mapeo FK↔PersonalNum↔Nombre", fill=C["text"], font=note_font)
    y += 14

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT)
    print(f"✅ Rendered: {OUTPUT}")
    print(f"   Dimensions: {img.size[0]}x{img.size[1]} px")


if __name__ == "__main__":
    main()
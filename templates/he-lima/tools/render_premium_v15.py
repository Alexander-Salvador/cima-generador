#!/usr/bin/env python3
"""Render premium FEEDBACK layout (v15 — redesigned end-to-end)."""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/cima_premium_v15.png"
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

C = {
    "white":       (255, 255, 255),
    "title_bg":    (26, 58, 92),
    "title_fg":    (255, 255, 255),
    "subtitle_bg": (232, 240, 250),
    "subtitle_fg": (26, 58, 92),
    "header_bg":   (26, 58, 92),
    "header_fg":   (255, 255, 255),
    "zebra":       (245, 248, 252),
    "zebra_alt":    (255, 255, 255),
    "text":        (15, 36, 64),
    "border":      (220, 225, 232),
    "hyperlink":   (26, 115, 232),
    "kpi_bg":      (250, 252, 254),
    "kpi_label_bg":(240, 244, 248),
    "kpi_value_bg":(255, 255, 255),
    "kpi_fg":      (26, 58, 92),
}

ESTADO_ICONS = {
    "Carta de compromiso": ("📋 Compromiso", (255, 220, 130)),
    "Feedback dado":       ("💬 Feedback",   (180, 220, 255)),
    "Cumplido":            ("✓ Listo",       (170, 230, 170)),
    "Vencido":             ("⚠ Vencido",     (255, 150, 150)),
    "En evaluación":       ("⏳ Eval",       (255, 255, 150)),
    "Pendiente":           ("· Pendiente",   (235, 235, 235)),
}

ESTADO_BG = {
    "Carta de compromiso": (255, 220, 130),
    "Feedback dado":       (180, 220, 255),
    "Cumplido":            (170, 230, 170),
    "Vencido":             (255, 150, 150),
    "En evaluación":       (255, 255, 150),
    "Pendiente":           (235, 235, 235),
    "Positivo":            (170, 230, 170),
    "Negativo":            (255, 150, 150),
    "Evaluativo":          (180, 220, 255),
    "Amonestador":         (255, 165, 110),
}

# v15: 10 cols
COL_WIDTHS = [200, 110, 110, 100, 360, 90, 130, 220, 130, 150]
HEADERS = ["AGENTE", "TIPO_FEEDBACK", "FECHA_FEEDBACK", "FECHA_BASE", "OBSERVACIÓN", "CANT_CALLS", "ESTADO", "SUPERVISOR", "STATUS", "ACCIÓN"]


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


def fetch(token, range_a1):
    fields = "sheets.data.rowData.values(formattedValue)"
    url = (f"https://sheets.googleapis.com/v4/spreadsheets/{SPREADSHEET_ID}"
           f"?includeGridData=true&ranges=FEEDBACK!{range_a1}&fields={fields}")
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


def render(draw, x_start, y_start, data_rows):
    fonts = {
        "title":    get_font(16, bold=True),
        "subtitle": get_font(10),
        "header":   get_font(9, bold=True),
        "data":     get_font(9),
        "data_b":   get_font(9, bold=True),
        "kpi_h":    get_font(10, bold=True),
        "kpi_v":    get_font(18, bold=True),
    }

    y = y_start
    row_h = 22
    total_w = sum(COL_WIDTHS)

    # Title row 1
    draw.rectangle([x_start, y, x_start + total_w, y + 32], fill=C["title_bg"])
    title_text = data_rows[0]["values"][0]["formattedValue"] if data_rows[0]["values"] else "FEEDBACKS"
    draw.text((x_start + 10, y + 8), title_text, fill=C["title_fg"], font=fonts["title"])
    y += 32

    # Subtitle row 2
    sub_text = data_rows[1]["values"][0]["formattedValue"] if len(data_rows) > 1 and data_rows[1]["values"] else ""
    draw.rectangle([x_start, y, x_start + total_w, y + 24], fill=C["subtitle_bg"])
    draw.text((x_start + 10, y + 5), sub_text, fill=C["subtitle_fg"], font=fonts["subtitle"])
    y += 28  # extra space

    # Header row 4
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], fill=C["header_bg"])
    x = x_start
    for col_idx, w in enumerate(COL_WIDTHS):
        tx = truncate(draw, HEADERS[col_idx], fonts["header"], w - 6)
        draw.text((x + 4, y + 6), tx, fill=C["header_fg"], font=fonts["header"])
        x += w
    y += row_h + 2

    # Underline divider (premium feel)
    draw.line([(x_start, y), (x_start + total_w, y)], fill=C["title_bg"], width=2)
    y += 4

    # Data rows 5-11
    for row_idx in range(4, min(11, len(data_rows))):
        actual_row = row_idx + 1
        cells = data_rows[row_idx].get("values", [])
        is_zebra = (actual_row % 2 == 1)
        bg = C["zebra"] if is_zebra else C["zebra_alt"]

        x = x_start
        for col_idx, w in enumerate(COL_WIDTHS):
            cell = cells[col_idx] if col_idx < len(cells) else {}
            val = cell.get("formattedValue", "")

            # STATUS col (8) - icon pill
            cell_bg = bg
            if col_idx == 6 and val in ESTADO_BG:
                cell_bg = ESTADO_BG[val]
            if col_idx == 1 and val in ESTADO_BG:
                cell_bg = ESTADO_BG[val]
            if col_idx == 8:  # STATUS visual
                cell_bg = bg
                for ev_prefix, (icon, color) in ESTADO_ICONS.items():
                    if val.startswith(icon[:2]):
                        cell_bg = color
                        break

            # Text styling
            fg = C["text"]
            font = fonts["data"]
            bold = False
            if col_idx in [0, 5]:  # AGENTE, CANT_CALLS
                fg = C["hyperlink"]
                bold = True
            if col_idx == 8:  # STATUS_VIS
                bold = True
                fg = C["text"]

            draw.rectangle([x, y, x + w, y + row_h], fill=cell_bg, outline=C["border"], width=1)
            tx = truncate(draw, str(val), font, w - 6)
            draw.text((x + 4, y + 5), tx, fill=fg, font=font)
            x += w
        y += row_h

    # Spacer
    y += 20

    # KPI Dashboard
    # Row 13: title (dark bg)
    draw.rectangle([x_start, y, x_start + total_w, y + 30], fill=C["title_bg"])
    kpi_title = data_rows[12]["values"][0]["formattedValue"] if len(data_rows) > 12 and data_rows[12]["values"] else "📊 DASHBOARD"
    draw.text((x_start + 10, y + 8), kpi_title, fill=C["title_fg"], font=fonts["title"])
    y += 30

    # Row 14: KPI labels
    kpi_labels = ["👥 AGENTES", "📞 CALLS", "✅ % CUMPLIDO", "⚠ % VENCIDO", "📊 TIPO TOP"]
    kpi_widths = [total_w // 5] * 4 + [total_w - (total_w // 5) * 4]

    x = x_start
    for i, (label, w) in enumerate(zip(kpi_labels, kpi_widths)):
        draw.rectangle([x, y, x + w, y + 24], fill=C["kpi_label_bg"])
        draw.text((x + w // 2 - len(label) * 3, y + 6), label, fill=C["title_bg"], font=fonts["kpi_h"])
        draw.rectangle([x, y, x + w, y + 24], outline=C["border"], width=1)
        x += w
    y += 24

    # Row 15: KPI values (big)
    kpi_values = []
    if len(data_rows) > 14:
        kpi_row = data_rows[14].get("values", [])
        for col_idx in [0, 2, 4, 6, 8]:
            if col_idx < len(kpi_row):
                kpi_values.append(kpi_row[col_idx].get("formattedValue", "0"))
            else:
                kpi_values.append("0")
    else:
        kpi_values = ["0"] * 5

    x = x_start
    for i, (val, w) in enumerate(zip(kpi_values, kpi_widths)):
        draw.rectangle([x, y, x + w, y + 40], fill=C["kpi_value_bg"])
        # Format value
        display = val
        if i == 2 or i == 3:  # %
            try:
                display = f"{float(val) * 100:.0f}%"
            except:
                pass
        draw.text((x + w // 2 - 20, y + 8), display, fill=C["kpi_fg"], font=fonts["kpi_v"])
        draw.rectangle([x, y, x + w, y + 40], outline=C["border"], width=1)
        x += w
    y += 44

    return y


def main():
    token = get_token()
    fb_data = fetch(token, "A1:J15").get("rowData", [])

    w = sum(COL_WIDTHS) + 40
    h = 600
    img = Image.new("RGB", (w, h), C["white"])
    draw = ImageDraw.Draw(img)

    y = 20
    title_font = get_font(20, bold=True)
    subtitle_font = get_font(11)
    draw.text((20, y), "CIMA — FEEDBACK Premium 2026 (v15)", fill=C["title_bg"], font=title_font)
    y += 28
    draw.text((20, y), "Rediseño end-to-end · Master con STATUS_VIS + ACCIONES + KPI Dashboard", fill=C["text"], font=subtitle_font)
    y += 24

    y = render(draw, 20, y, fb_data)

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT)
    print(f"✅ Rendered: {OUTPUT}")
    print(f"   Dimensions: {img.size[0]}x{img.size[1]} px")


if __name__ == "__main__":
    main()
#!/usr/bin/env python3
"""Render FEEDBACK v17 — sin KPI 5 + rows 13-30 encuadradas."""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/cima_v17_kpi5removed.png"
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

C = {
    "white":       (255, 255, 255),
    "title_bg":    (26, 58, 92),
    "title_fg":    (255, 255, 255),
    "kpi_label_bg":(240, 244, 248),
    "kpi_value_bg":(255, 255, 255),
    "header_bg":   (26, 58, 92),
    "header_fg":   (255, 255, 255),
    "zebra":       (245, 248, 252),
    "zebra_alt":   (255, 255, 255),
    "empty_bg":    (250, 252, 254),
    "empty_bg_alt":(255, 255, 255),
    "text":        (15, 36, 64),
    "hyperlink":   (26, 115, 232),
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

# v17: 8 cols (sin STATUS, sin ACCION)
COL_WIDTHS = [200, 110, 110, 100, 360, 90, 130, 220]
HEADERS = ["AGENTE", "TIPO_FEEDBACK", "FECHA_FEEDBACK", "FECHA_BASE", "OBSERVACIÓN", "CANT_CALLS", "ESTADO", "SUPERVISOR"]


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
        "title":  get_font(16, bold=True),
        "dash":   get_font(13, bold=True),
        "kpi_h":  get_font(11, bold=True),
        "kpi_v":  get_font(20, bold=True),
        "header": get_font(10, bold=True),
        "data":   get_font(9),
        "data_b": get_font(9, bold=True),
        "empty":  get_font(9),
    }

    y = y_start
    row_h = 22
    total_w = sum(COL_WIDTHS)

    # Row 1: Title
    draw.rectangle([x_start, y, x_start + total_w, y + 32], fill=C["title_bg"])
    title_text = data_rows[0]["values"][0]["formattedValue"] if data_rows[0]["values"] else "FEEDBACKS"
    draw.text((x_start + 10, y + 8), title_text, fill=C["title_fg"], font=fonts["title"])
    y += 32

    # Row 2-3: 4 KPI cards (sin TIPO TOP)
    kpi_labels = ["👥 AGENTES", "📞 CALLS", "✅ % CUMPLIDO", "⚠ % VENCIDO"]
    kpi_w = total_w // 4
    kpi_extra = total_w - (kpi_w * 3)
    kpi_widths = [kpi_w, kpi_w, kpi_w, kpi_extra]

    # Row 2: KPI labels
    for i, label in enumerate(kpi_labels):
        w = kpi_widths[i]
        x = x_start + i * kpi_w
        draw.rectangle([x, y, x + w, y + 24], fill=C["kpi_label_bg"], outline=(200, 200, 200), width=1)
        bbox = draw.textbbox((0, 0), label, font=fonts["kpi_h"])
        text_w = bbox[2] - bbox[0]
        draw.text((x + (w - text_w) // 2, y + 6), label, fill=C["text"], font=fonts["kpi_h"])
    y += 24

    # Row 3: KPI values
    kpi_vals = []
    if len(data_rows) > 2:
        kpi_row = data_rows[2].get("values", [])
        for col_idx in [0, 2, 4, 6]:
            if col_idx < len(kpi_row):
                kpi_vals.append(kpi_row[col_idx].get("formattedValue", "0"))
            else:
                kpi_vals.append("0")
    else:
        kpi_vals = ["0"] * 4

    for i, val in enumerate(kpi_vals):
        w = kpi_widths[i]
        x = x_start + i * kpi_w
        draw.rectangle([x, y, x + w, y + 40], fill=C["kpi_value_bg"], outline=(200, 200, 200), width=1)
        display = val
        if i in [2, 3]:
            try:
                display = f"{float(val) * 100:.0f}%"
            except:
                pass
        bbox = draw.textbbox((0, 0), display, font=fonts["kpi_v"])
        text_w = bbox[2] - bbox[0]
        draw.text((x + (w - text_w) // 2, y + 8), display, fill=C["text"], font=fonts["kpi_v"])
    y += 40

    # Row 4: spacer
    y += 12

    # Row 5: Data header
    draw.rectangle([x_start, y, x_start + total_w, y + row_h + 2], fill=C["header_bg"])
    x = x_start
    for col_idx, w in enumerate(COL_WIDTHS):
        tx = truncate(draw, HEADERS[col_idx], fonts["header"], w - 6)
        draw.text((x + 4, y + 6), tx, fill=C["header_fg"], font=fonts["header"])
        x += w
    y += row_h + 2

    # Rows 6-30: data + empty placeholder rows
    for row_idx in range(5, min(30, len(data_rows))):
        actual_row = row_idx + 1
        cells = data_rows[row_idx].get("values", [])
        is_zebra = (actual_row % 2 == 1)

        has_data = len(cells) > 0 and cells[0].get("formattedValue", "") != ""

        if has_data:
            bg = C["zebra"] if is_zebra else C["zebra_alt"]
        else:
            bg = C["empty_bg"] if is_zebra else C["empty_bg_alt"]

        x = x_start
        for col_idx, w in enumerate(COL_WIDTHS):
            cell = cells[col_idx] if col_idx < len(cells) else {}
            val = cell.get("formattedValue", "")

            cell_bg = bg
            fg = C["text"] if val else (180, 180, 180)
            font = fonts["data"]

            # Highlight ESTADO col (col 6)
            if col_idx == 6 and val in ESTADO_BG:
                cell_bg = ESTADO_BG[val]
            # Highlight TIPO_FEEDBACK col (col 1)
            if col_idx == 1 and val in ESTADO_BG:
                cell_bg = ESTADO_BG[val]

            if col_idx in [0, 5] and val:
                fg = C["hyperlink"]
                font = fonts["data_b"]

            draw.rectangle([x, y, x + w, y + row_h], fill=cell_bg, outline=(220, 225, 232), width=1)
            tx = truncate(draw, str(val), font, w - 6)
            draw.text((x + 4, y + 5), tx, fill=fg, font=font)
            x += w
        y += row_h

    return y


def main():
    token = get_token()
    fb_data = fetch(token, "A1:H30").get("rowData", [])

    w = sum(COL_WIDTHS) + 40
    h = 850
    img = Image.new("RGB", (w, h), C["white"])
    draw = ImageDraw.Draw(img)

    y = 20
    title_font = get_font(20, bold=True)
    subtitle_font = get_font(11)
    draw.text((20, y), "CIMA — FEEDBACK Premium v17 (KPI 5 eliminado)", fill=C["title_bg"], font=title_font)
    y += 28
    draw.text((20, y), "Layout: Title → 4 KPIs → DATA → 18 rows encuadradas (vacías + zebra)", fill=C["text"], font=subtitle_font)
    y += 24

    y = render(draw, 20, y, fb_data)

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT)
    print(f"✅ Rendered: {OUTPUT}")
    print(f"   Dimensions: {img.size[0]}x{img.size[1]} px")


if __name__ == "__main__":
    main()
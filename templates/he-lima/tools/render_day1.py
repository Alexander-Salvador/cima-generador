#!/usr/bin/env python3
"""render_day1.py — CIMA Day 1: PIL render of 00_Config from TEST sheet.

Renders 00_Config to PNG via Google Sheets API → PIL.
Saves to .ovav/state/test_render_2026-09-28.png
"""
import json
import os
import urllib.request
from PIL import Image, ImageDraw, ImageFont

STATE_DIR = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/state"
OUTPUT = os.path.join(STATE_DIR, "test_render_2026-09-28.png")
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

C = {
    "navy":       (30, 58, 138),     # #1e3a8a
    "white":      (255, 255, 255),
    "surface":    (244, 247, 251),   # #f4f7fb
    "text":       (15, 36, 64),
    "muted":      (115, 132, 153),
    "border":     (180, 190, 210),
    "orange":     (245, 158, 11),    # #f59e0b (accent for 10_Registro_HE)
    "bg":         (245, 245, 247),
}

# 00_Config layout: A=ANCLA (140px), B=VALOR (120px), C=DESCRIPCIÓN (320px)
COL_WIDTHS = [140, 120, 320]
ROW_HEIGHT = 28
HEADER_ROW_HEIGHT = 36


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
    req = urllib.request.Request(
        "https://oauth2.googleapis.com/token", data=body,
        headers={"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["access_token"]


def fetch_cfg(token, spreadsheet_id):
    """Fetch 00_Config grid data + formatting."""
    fields = "sheets.properties.title,sheets.data.rowData.values(formattedValue,userEnteredValue,userEnteredFormat.backgroundColor,userEnteredFormat.textFormat,userEnteredFormat.horizontalAlignment)"
    url = (
        f"https://sheets.googleapis.com/v4/spreadsheets/{spreadsheet_id}"
        f"?includeGridData=true&ranges=00_Config!A1:C11&fields={fields}"
    )
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read())
    sheets = data["sheets"]
    if not sheets:
        return None
    if "data" not in sheets[0]:
        return None
    return sheets[0]["data"][0]


def get_font(size, bold=False):
    paths = [
        f"/usr/share/fonts/truetype/dejavu/DejaVuSans{'-Bold' if bold else ''}.ttf",
        f"/usr/share/fonts/truetype/liberation/LiberationSans{'-Bold' if bold else '-Regular'}.ttf",
    ]
    for p in paths:
        if os.path.exists(p):
            try:
                return ImageFont.truetype(p, size)
            except Exception:
                pass
    return ImageFont.load_default()


def rgb_from_api(color):
    """Convert API color dict (0..1 floats) to 0..255 tuple."""
    return (
        int(color.get("red", 1) * 255),
        int(color.get("green", 1) * 255),
        int(color.get("blue", 1) * 255),
    )


def render():
    # Read TEST sheet ID
    id_path = os.path.join(STATE_DIR, "test_sheet_id")
    spreadsheet_id = open(id_path).read().strip()
    print(f"Rendering 00_Config from {spreadsheet_id}", flush=True)

    sheet = fetch_cfg(get_token(), spreadsheet_id)
    if not sheet or "rowData" not in sheet:
        print("ERROR: no data")
        return 1
    rows = sheet["rowData"]
    print(f"Fetched {len(rows)} rows", flush=True)

    margin = 40
    title_h = 70
    img_w = sum(COL_WIDTHS) + 2 * margin
    img_h = HEADER_ROW_HEIGHT + ROW_HEIGHT * (len(rows) - 1) + 2 * margin + title_h
    img = Image.new("RGB", (img_w, img_h), C["bg"])
    draw = ImageDraw.Draw(img)

    # Title
    draw.text((margin, 10), "HE LIMA — CIMA [TEST 2026-09-28]", fill=C["navy"], font=get_font(20, bold=True))
    draw.text((margin, 36), "00_Config panel — render via Google Sheets API + PIL",
              fill=C["muted"], font=get_font(11))

    y0 = margin + title_h

    # Render rows
    for ri, row in enumerate(rows):
        cells = row.get("values", [])
        x = margin
        if ri == 0:
            h = HEADER_ROW_HEIGHT
            y = y0
        else:
            h = ROW_HEIGHT
            y = y0 + HEADER_ROW_HEIGHT + (ri - 1) * ROW_HEIGHT

        for ci in range(len(COL_WIDTHS)):
            cw = COL_WIDTHS[ci]
            cell = cells[ci] if ci < len(cells) else None

            # Determine bg color
            bg = C["surface"] if ri > 0 else C["navy"]
            text_color = C["text"] if ri > 0 else C["white"]
            bold = ri == 0
            align = "LEFT"

            if cell:
                uef = cell.get("userEnteredFormat", {})
                if "backgroundColor" in uef:
                    bg = rgb_from_api(uef["backgroundColor"])
                if "textFormat" in uef:
                    tf = uef["textFormat"]
                    if tf.get("bold"):
                        bold = True
                    if "foregroundColor" in tf:
                        text_color = rgb_from_api(tf["foregroundColor"])
                if "horizontalAlignment" in uef:
                    align = uef["horizontalAlignment"]

            # Draw cell
            draw.rectangle([x, y, x + cw, y + h], fill=bg, outline=C["border"], width=1)

            # Draw text
            text = ""
            if cell:
                text = cell.get("formattedValue", "")
                if not text and "userEnteredValue" in cell:
                    uev = cell["userEnteredValue"]
                    if "stringValue" in uev:
                        text = uev["stringValue"]
                    elif "numberValue" in uev:
                        text = str(uev["numberValue"])

            if text or ri == 0:
                # Allow descriptions to wrap (col C is wide)
                max_chars = cw // 7
                display = text[:max_chars] if len(text) > max_chars else text
                if len(text) > max_chars:
                    display = display.rstrip() + "…"

                font_size = 13 if ri == 0 else 11
                font = get_font(font_size, bold=bold)
                try:
                    bbox = draw.textbbox((0, 0), display, font=font)
                    tw = bbox[2] - bbox[0]
                    th = bbox[3] - bbox[1]
                except AttributeError:
                    tw, th = draw.textsize(display, font=font)

                if align == "CENTER":
                    tx = x + (cw - tw) // 2
                elif align == "RIGHT":
                    tx = x + cw - tw - 8
                else:
                    tx = x + 10

                ty = y + (h - th) // 2 - 1
                draw.text((tx, ty), display, fill=text_color, font=font)

            x += cw

    # Save
    os.makedirs(STATE_DIR, exist_ok=True)
    img.save(OUTPUT, "PNG", optimize=True)
    print(f"OK saved {OUTPUT} ({os.path.getsize(OUTPUT)} bytes, {img_w}x{img_h})", flush=True)
    return 0


if __name__ == "__main__":
    exit(render())

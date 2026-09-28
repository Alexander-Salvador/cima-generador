#!/usr/bin/env python3
"""Render CIMA TEAM sheet as PNG with merge support."""
import json
import os
import urllib.request
import time
from PIL import Image, ImageDraw, ImageFont

SPREADSHEET_ID = "1MQ3wts_cEG_4Dp5U6X4gehEn6u7F61tIl7KHzSzKw0Y"
OUTPUT = "/tmp/opencode/screenshots/cima_team_render_v2.png"
VAULT_PATH = "/home/braka/Systems/ovav/.ovav/worktrees/feature-c1ma-desing/.ovav/vault/credentials/google_oauth.enc"
KEY_PATH = "/home/braka/.config/ovav/vault.key"

C = {
    "navy":      (26, 58, 92),
    "subtle":    (212, 227, 243),
    "alert":     (230, 126, 34),
    "alert_soft":(253, 235, 208),
    "muted":     (234, 237, 241),
    "muted_dark":(115, 132, 153),
    "white":     (255, 255, 255),
    "text":      (15, 36, 64),
}

COL_WIDTHS = [26, 60, 55, 50] + [22]*30 + [6, 55, 2, 55, 2, 55, 2, 65, 0, 0, 0, 0, 60]  # post-fix-widths: A-D, days E-AH, AI spacer, AJ-AP summary, AQ-AT eliminated, AU
ROW_HEIGHTS = {1: 32, 2: 24, 3: 24, 4: 8, 5: 22, 6: 22, 7: 8, 8: 22, 9: 18}
DEFAULT_ROW_H = 18


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
        "sheets.data.rowData.values(formattedValue,userEnteredValue,userEnteredFormat.numberFormat,userEnteredFormat.textFormat,userEnteredFormat.backgroundColor,userEnteredFormat.horizontalAlignment,userEnteredFormat.verticalAlignment)",
        "sheets.merges"
    ])
    url = (f"https://sheets.googleapis.com/v4/spreadsheets/{SPREADSHEET_ID}"
           f"?includeGridData=true&ranges=HORAS%20EXTRAS!A1:AU40&fields={fields}")
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["sheets"][0]


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


def get_x(idx):
    x = 0
    for w in COL_WIDTHS[:idx]:
        x += w
    return x


def get_y(idx):
    y = 0
    for r in range(idx):
        y += ROW_HEIGHTS.get(r + 1, DEFAULT_ROW_H)
    return y


def render():
    print("Fetching...", flush=True)
    sheet = fetch(get_token())
    rows = sheet["data"][0]["rowData"]
    merges = sheet["merges"]

    # Build merge map: (startRow, endRow, startCol, endCol) -> info
    merge_map = {}  # (row, col) -> merge info
    for m in merges:
        sr = int(m["startRowIndex"])
        er = int(m["endRowIndex"])
        sc = int(m["startColumnIndex"])
        ec = int(m["endColumnIndex"])
        # Anchor is top-left cell
        merge_map[(sr, sc)] = (sr, er, sc, ec)
        # Hide other cells
        for r in range(sr, er):
            for c in range(sc, ec):
                if (r, c) != (sr, sc):
                    merge_map[(r, c)] = ("HIDDEN", None)

    total_w = sum(COL_WIDTHS)
    total_h = sum(ROW_HEIGHTS.get(r + 1, DEFAULT_ROW_H) for r in range(36))
    margin = 40

    img_w = total_w + 2 * margin
    img_h = total_h + 2 * margin + 70
    img = Image.new("RGB", (img_w, img_h), (245, 245, 247))
    draw = ImageDraw.Draw(img)

    # Title
    draw.text((margin, 10), "HORAS EXTRAS — Visual Verification Render", fill=C["navy"], font=get_font(20, bold=True))
    draw.text((margin, 36), f"Total width: {total_w}px | Tab: HORAS EXTRAS | Generated from live API data",
              fill=C["muted_dark"], font=get_font(11))

    y0 = margin + 70
    x0 = margin

    # Frame border around header (rows 1-7) — navy
    hx = x0
    hy = y0 + get_y(0)
    hw = sum(COL_WIDTHS)
    hh = sum(ROW_HEIGHTS.get(r + 1, DEFAULT_ROW_H) for r in range(7))
    draw.rectangle([hx, hy, hx + hw, hy + hh], outline=C["navy"], width=3)

    # Frame border around data (rows 8-36) — orange accent
    dx = x0
    dy = y0 + get_y(7)
    dh = sum(ROW_HEIGHTS.get(r + 1, DEFAULT_ROW_H) for r in range(7, 36))
    draw.rectangle([dx, dy, dx + hw, dy + dh], outline=C["alert"], width=3)

    # Render cells
    for ri, row in enumerate(rows):
        if ri >= 36:
            break
        row_h = ROW_HEIGHTS.get(ri + 1, DEFAULT_ROW_H)
        cy = y0 + get_y(ri)

        cells = row.get("values", [])
        for ci in range(47):
            cx = x0 + get_x(ci)

            # Check merge state
            merge_state = merge_map.get((ri, ci))
            if merge_state == "HIDDEN":
                continue

            # Determine cell bounds (might span multiple cols/rows if merged)
            if isinstance(merge_state, tuple) and merge_state[0] != "HIDDEN":
                sr, er, sc, ec = merge_state
                # Cell covers from (sr, sc) to (er-1, ec-1)
                cell_w = sum(COL_WIDTHS[sc:ec])
                cell_h = sum(ROW_HEIGHTS.get(r + 1, DEFAULT_ROW_H) for r in range(sr, er))
            else:
                cell_w = COL_WIDTHS[ci]
                cell_h = row_h

            cell_bg = C["white"]
            text_color = C["text"]
            bold = False
            align = "LEFT"
            italic = False
            strikethrough = False

            cell = cells[ci] if ci < len(cells) else None
            text = ""
            if cell:
                text = cell.get("formattedValue", "")
                uef = cell.get("userEnteredFormat", {})
                bg = uef.get("backgroundColor")
                if bg:
                    r = int(bg.get("red", 1) * 255)
                    g = int(bg.get("green", 1) * 255)
                    b = int(bg.get("blue", 1) * 255)
                    cell_bg = (r, g, b)
                tf = uef.get("textFormat", {})
                if tf.get("bold"):
                    bold = True
                if tf.get("italic"):
                    italic = True
                if tf.get("strikethrough"):
                    strikethrough = True
                fg = tf.get("foregroundColor")
                if fg:
                    r = int(fg.get("red", 0) * 255)
                    g = int(fg.get("green", 0) * 255)
                    b = int(fg.get("blue", 0) * 255)
                    text_color = (r, g, b)
                if "horizontalAlignment" in uef:
                    align = uef["horizontalAlignment"]

            # Draw cell bg
            draw.rectangle([cx, cy, cx + cell_w, cy + cell_h], fill=cell_bg)

            # Draw text
            if text and text.strip():
                # Font size based on cell
                if ri == 0:
                    font_size = 16
                elif cell_h < 16:
                    font_size = 8
                elif cell_w < 25:
                    font_size = 9
                else:
                    font_size = 10
                font = get_font(font_size, bold=bold)

                # Truncate text to fit cell width
                max_chars = max(3, cell_w // (font_size // 2 + 2))
                display = text[:max_chars]
                if len(text) > max_chars:
                    display = display.rstrip() + "…"

                try:
                    bbox = draw.textbbox((0, 0), display, font=font)
                    tw = bbox[2] - bbox[0]
                    th = bbox[3] - bbox[1]
                except AttributeError:
                    tw, th = draw.textsize(display, font=font)

                if align == "CENTER":
                    tx = cx + (cell_w - tw) // 2
                elif align == "RIGHT":
                    tx = cx + cell_w - tw - 2
                else:
                    tx = cx + 3

                ty = cy + (cell_h - th) // 2 - 1
                draw.text((tx, ty), display, fill=text_color, font=font)

                # Strikethrough overlay
                if strikethrough:
                    sx1 = tx
                    sy = ty + th // 2
                    sx2 = tx + tw
                    draw.line([sx1, sy, sx2, sy], fill=text_color, width=1)

    # Save
    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    img.save(OUTPUT, "PNG", optimize=True)
    print(f"OK saved {OUTPUT} ({os.path.getsize(OUTPUT)} bytes)", flush=True)


if __name__ == "__main__":
    render()

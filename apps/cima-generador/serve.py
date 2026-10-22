#!/usr/bin/env python3
"""
CIMA — Servidor local con anti-cache.

POR QUÉ EXISTE (CRIT-004: cortar la causa raíz, no parchear)
-------------------------------------------------------------
`python3 -m http.server` sirve HTTP/1.0 SIN cabecera `Cache-Control`.
Para un archivo de 1.19 MB, los browsers aplican "heuristic freshness"
(10% de la edad del documento) y sirven HTML VIEJO desde cache aunque el
disco ya cambió. Eso produjo el bug "cambié el código y no se ve".

Este wrapper sirve el MISMO directorio con:
  - Cache-Control: no-store, no-cache, must-revalidate, max-age=0
  - Pragma: no-cache
  - Expires: 0
  - ETag: 0  (desactiva If-None-Match → siempre 200 con contenido fresco)

Uso:
  python3 serve.py [port]      # default 8765
  python3 serve.py 8766        # puerto alternativo

Mismo comportamiento que http.server, pero CADA reload trae el archivo real.
"""

import http.server
import socketserver
import sys
import os
import re
import hashlib
import json
import time

PORT = int(sys.argv[1]) if len(sys.argv) > 1 else 8765
DIR = os.path.dirname(os.path.abspath(__file__))
INDEX = os.path.join(DIR, "index.html")
STARTED = time.time()


def build_id():
    """Extrae la versión embebida en el HTML (build-chip / footerVersion)."""
    try:
        src = open(INDEX, "r", encoding="utf-8").read()
    except OSError:
        return "unknown"
    for pat in (r'id="buildChip"[^>]*>(v[0-9.]+)<',
                r'id="footerVersion"[^>]*>(v[0-9.]+)<'):
        m = re.search(pat, src)
        if m:
            return m.group(1)
    return "unknown"


def file_facts():
    """MD5 + tamaño + mtime + build del index.html en disco."""
    if not os.path.exists(INDEX):
        return {"exists": False}
    data = open(INDEX, "rb").read()
    st = os.stat(INDEX)
    return {
        "exists": True,
        "build": build_id(),
        "md5": hashlib.md5(data).hexdigest(),
        "bytes": len(data),
        "mtime": time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(st.st_mtime)),
    }


class NoCacheHandler(http.server.SimpleHTTPRequestHandler):
    """SimpleHTTPRequestHandler + cabeceras anti-cache + endpoints de auditoría."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIR, **kwargs)

    def _json(self, payload, code=200):
        body = json.dumps(payload, indent=2, ensure_ascii=False).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store, no-cache, must-revalidate")
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = self.path.split("?")[0]

        if path == "/__health":
            return self._json({
                "pid": os.getpid(),
                "serving_dir": DIR,
                "index": file_facts(),
                "uptime_s": round(time.time() - STARTED, 1),
                "cache_control": "no-store",
            })

        # Handshake de build: el popup lo consulta al cargar y se autorrecarga
        # si el build del servidor != el build embebido en el documento.
        if path == "/__build":
            return self._json({"build": build_id(), "facts": file_facts()})

        return super().do_GET()

    def end_headers(self):
        self.send_header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
        self.send_header("Pragma", "no-cache")
        self.send_header("Expires", "0")
        self.send_header("ETag", 'W/"nocache"')
        super().end_headers()

    def log_message(self, fmt, *args):
        sys.stderr.write("  %s\n" % (fmt % args))


class ReusableServer(socketserver.ThreadingTCPServer):
    """Threading: una conexión colgada de un browser NO bloquea el server.

    Antes usaba TCPServer (single-thread): con HTTP/1.0 y un browser que
    mantiene la conexión abierta, el server se quedaba sin poder atender
    el siguiente request → curl devolvía 000 aunque el proceso seguía vivo.
    """
    allow_reuse_address = True
    daemon_threads = True
    request_queue_size = 64


if __name__ == "__main__":
    facts = file_facts()
    with ReusableServer(("127.0.0.1", PORT), NoCacheHandler) as httpd:
        print("=" * 64)
        print("CIMA dev server (NO-CACHE)")
        print("  URLpopup : http://127.0.0.1:{}/index.html?popup=1".format(PORT))
        print("  Debug    : http://127.0.0.1:{}/index.html?popup=1&debug=1".format(PORT))
        print("  Salud    : http://127.0.0.1:{}/__health".format(PORT))
        print("  Sirviendo: {}".format(DIR))
        print("  index.html: {} bytes  md5={}".format(
            facts.get("bytes"), facts.get("md5", "?")[:12]))
        print("  PID       : {}".format(os.getpid()))
        print("=" * 64)
        sys.stdout.flush()
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nServidor detenido.")
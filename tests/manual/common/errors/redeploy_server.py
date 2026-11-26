#!/usr/bin/env python3
"""Tiny Cumulocity stand-in for an application which is already deployed.

The application exists, its new version is already the active one (409 when it
is activated) and the microservice is already subscribed (409 when it is
subscribed). Everything else succeeds.

Usage: redeploy_server.py <port-file>
It picks a free port, writes it to <port-file> and serves until killed.
"""
import json
import sys
from urllib.parse import parse_qs, urlparse
from http.server import BaseHTTPRequestHandler, HTTPServer

SELF = "http://127.0.0.1:{port}/application/applications/5"


class Handler(BaseHTTPRequestHandler):
    def _send(self, status, doc=None):
        body = json.dumps(doc).encode() if doc is not None else b""
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _app(self, name="app", app_type="HOSTED", owner="t100"):
        return {
            "id": "5", "name": name, "key": name + "-key", "type": app_type,
            "contextPath": name, "activeVersionId": "bin2",
            "owner": {"tenant": {"id": owner}},
            "self": SELF.format(port=self.server.server_address[1]),
        }

    def _reply(self):
        length = int(self.headers.get("Content-Length") or 0)
        raw = self.rfile.read(length) if length else b""
        url = urlparse(self.path)
        query = parse_qs(url.query)
        path = url.path
        conflict = {"error": "general/conflict", "message": "already exists"}
        if self.command == "GET" and path == "/application/applications":
            name = (query.get("name") or ["app"])[0]
            app_type = (query.get("type") or ["HOSTED"])[0]
            owner = (query.get("owner") or ["t100"])[0]
            return self._send(200, {"applications": [self._app(name, app_type, owner)],
                                    "statistics": {"currentPage": 1, "pageSize": 5}})
        if self.command == "GET" and path == "/application/applications/5":
            return self._send(200, self._app())
        if self.command == "GET" and path == "/application/applications/5/binaries":
            return self._send(200, {"attachments": [
                {"id": "bin1", "created": "2026-01-01T00:00:00.000Z"},
                {"id": "bin2", "created": "2026-01-02T00:00:00.000Z"}]})
        if self.command == "GET" and path == "/tenant/currentTenant":
            return self._send(200, {"name": "t100"})
        if self.command == "POST" and path == "/application/applications/5/binaries":
            return self._send(201, {"id": "bin2"})
        if self.command == "PUT" and path == "/application/applications/5":
            if b"activeVersionId" in raw:
                return self._send(409, conflict)
            return self._send(200, self._app("app", "MICROSERVICE"))
        if self.command == "POST" and path.startswith("/tenant/tenants/") and path.endswith("/applications"):
            return self._send(409, conflict)
        if self.command == "DELETE":
            return self._send(404, {"error": "general/notFound", "message": "gone"})
        return self._send(404, {"error": "general/notFound", "message": self.command + " " + path})

    do_GET = do_POST = do_PUT = do_DELETE = _reply

    def log_message(self, *args):
        pass


server = HTTPServer(("127.0.0.1", 0), Handler)
with open(sys.argv[1], "w") as f:
    f.write(str(server.server_address[1]))
server.serve_forever()

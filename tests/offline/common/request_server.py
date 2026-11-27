#!/usr/bin/env python3
"""Tiny Cumulocity stand-in which records the requests it receives.

Usage: request_server.py <port-file> <log-file> [mode]

It picks a free port, writes it to <port-file> and serves until killed. Every
request is appended to <log-file> as a block: the request line ("GET /path?query"),
one line per header ("Name: value"), the body (latin-1) and a "--" line.

Modes:
  ok           answers with JSON (default)
  slow         waits 2 seconds before it answers
  unavailable  answers 503
  escaped      answers a list with the "<", ">" and "&" of the names escaped (as Go's
               encoding/json writes them: the six characters backslash-u-0-0-3-c etc.)
"""
import json
import sys
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

port_file, log_file = sys.argv[1], sys.argv[2]
mode = sys.argv[3] if len(sys.argv) > 3 else "ok"


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _body(self):
        if self.headers.get("Transfer-Encoding", "").lower() == "chunked":
            data = b""
            while True:
                size = int(self.rfile.readline().strip() or b"0", 16)
                if size == 0:
                    self.rfile.readline()
                    return data
                data += self.rfile.read(size)
                self.rfile.readline()
        length = int(self.headers.get("Content-Length") or 0)
        return self.rfile.read(length) if length else b""

    def _reply(self):
        body = self._body()
        with open(log_file, "a") as f:
            f.write("%s %s\n" % (self.command, self.path))
            for name, value in self.headers.items():
                f.write("%s: %s\n" % (name, value))
            f.write(body.decode("latin-1") + "\n--\n")
        if mode == "slow":
            time.sleep(2)
        status = 200
        path = self.path.split("?")[0]
        if mode == "unavailable":
            status, out = 503, {}
        elif path.endswith("/emptyarr"):
            out = []
        elif path.endswith("/arr"):
            out = [{"a": 1, "b": 2}, {"a": 3, "b": 4}]
        elif self.command == "POST":
            status, out = 201, {"id": "1", "name": "created"}
        elif path.endswith("/managedObjects"):
            out = {"managedObjects": [], "statistics": {"currentPage": 1, "pageSize": 1}}
        elif path.endswith("/events") or path.endswith("/alarms") or path.endswith("/operations") or path.endswith("/measurements"):
            out = {path.rsplit("/", 1)[1]: [], "statistics": {"currentPage": 1, "pageSize": 1}}
        else:
            out = {"id": "1", "name": "x"}
        payload = json.dumps(out).encode()
        if mode == "escaped":
            # a list which has the characters escaped, as Go's encoding/json writes them
            base = "http://127.0.0.1:%d/inventory/managedObjects" % server.server_address[1]
            payload = json.dumps({
                "managedObjects": [{"id": "1", "name": "amp&lt<gt>"}],
                "next": base + "?pageSize=1&currentPage=2",
                "statistics": {"currentPage": 1, "pageSize": 1},
            }).encode()
            for char, code in (("<", "003c"), (">", "003e"), ("&", "0026")):
                payload = payload.replace(char.encode(), (chr(92) + "u" + code).encode())
        try:
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
        except OSError:
            pass

    do_GET = do_POST = do_PUT = do_DELETE = _reply

    def log_message(self, *args):
        pass

    def handle(self):
        try:
            super().handle()
        except OSError:
            pass


server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
with open(port_file, "w") as f:
    f.write(str(server.server_address[1]))
server.serve_forever()

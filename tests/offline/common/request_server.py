#!/usr/bin/env python3
"""Tiny Cumulocity stand-in which records the requests it receives.

Usage: request_server.py <port-file> <log-file> [mode]

It picks a free port, writes it to <port-file> and serves until killed. Every
request is appended to <log-file> as a block: the request line ("GET /path?query"),
one line per header ("Name: value"), the body (latin-1) and a "--" line.
The time (seconds since the epoch) of each request is appended to <log-file>.times.

Modes:
  ok           answers with JSON (default)
  slow         waits 2 seconds before it answers
  unavailable  answers 503
  notfound     answers every GET with 404 and the error of the platform
  data         answers the collections (events, alarms, measurements, operations,
               audit records and managed objects) with 8 items each, filtered by
               dateFrom/dateTo (inclusive) and by an "_id gt 'N'" in the query, newest
               first (oldest first with revert=true or an "$orderby=_id asc"), and
               paged by currentPage and pageSize, like the platform does
  wsfail-<status>
               answers every /notification2/* request with the status (and the error of
               the platform), and every other request with 200 {} (so no websocket
               upgrade: the websocket handshake of the realtime commands fails)
  escaped      answers a list with the "<", ">" and "&" of the names escaped (as Go's
               encoding/json writes them: the six characters backslash-u-0-0-3-c etc.)
"""
import json
import re
import sys
import time
from urllib.parse import parse_qs, urlencode, urlparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

port_file, log_file = sys.argv[1], sys.argv[2]
mode = sys.argv[3] if len(sys.argv) > 3 else "ok"


COLLECTIONS = {
    "/event/events": "events",
    "/alarm/alarms": "alarms",
    "/measurement/measurements": "measurements",
    "/devicecontrol/operations": "operations",
    "/audit/auditRecords": "auditRecords",
    "/inventory/managedObjects": "managedObjects",
}
TOTAL = 8


def stamp(n):
    return "2024-01-01T10:00:%02d.000Z" % n


def data_page(path, query):
    """The page of a collection: 8 items with the ids 1 to 8, the time of item n is n seconds."""
    q = {k: v[0] for k, v in parse_qs(query).items()}
    items = list(range(1, TOTAL + 1))
    text = q.get("query", "") + " " + q.get("q", "")
    # time filters (the platform keeps the items with dateFrom <= time <= dateTo)
    if q.get("dateTo"):
        items = [n for n in items if stamp(n) <= q["dateTo"]]
    if q.get("dateFrom"):
        items = [n for n in items if stamp(n) >= q["dateFrom"]]
    m = re.search(r"_id gt '(\d+)'", text)
    if m:
        items = [n for n in items if n > int(m.group(1))]
    ascending = q.get("revert") == "true" or "_id asc" in text
    items.sort(reverse=not ascending)
    total = len(items)
    size = int(q.get("pageSize", 5))
    page = int(q.get("currentPage", 1))
    items = items[(page - 1) * size:page * size]
    key = COLLECTIONS[path]
    out = []
    for n in items:
        item = {"id": str(n), "time": stamp(n), "creationTime": stamp(n), "type": "t", "source": {"id": "1"}, "name": "n%d" % n}
        if key == "managedObjects":
            item.pop("time")
        out.append(item)
    base = "http://127.0.0.1:%d%s" % (server.server_address[1], path)
    result = {"next": None, "self": base + "?" + urlencode(q), key: out, "statistics": {"currentPage": page, "pageSize": size}}
    result.pop("next")
    if page * size < total:
        nxt = dict(q)
        nxt["currentPage"] = str(page + 1)
        result["next"] = base + "?" + urlencode(nxt)
    return result


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
        with open(log_file + ".times", "a") as f:
            f.write("%f\n" % time.time())
        if mode == "slow":
            time.sleep(2)
        status = 200
        path = self.path.split("?")[0]
        if mode.startswith("wsfail-"):
            if path.startswith("/notification2/"):
                status = int(mode.split("-", 1)[1])
                out = {"error": "security/Forbidden", "message": "Access is denied"}
            else:
                out = {}
        elif mode == "unavailable":
            status, out = 503, {}
        elif mode == "notfound" and self.command == "GET":
            status, out = 404, {"error": "inventory/Not Found", "message": "Finding device data from database failed", "info": "https://cumulocity.com/api"}
        elif mode == "data" and self.command == "GET" and path in COLLECTIONS:
            out = data_page(path, urlparse(self.path).query)
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

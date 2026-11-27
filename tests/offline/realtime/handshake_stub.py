"""A websocket server which upgrades the connection but never answers the
realtime (Bayeux) handshake, or a server which never answers at all.

usage: handshake_stub.py <portfile> <upgrade|silent>
"""
import base64
import hashlib
import socket
import sys
import threading

portfile, mode = sys.argv[1], sys.argv[2]
server = socket.socket()
server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
server.bind(("127.0.0.1", 0))
server.listen(50)
with open(portfile, "w") as f:
    f.write(str(server.getsockname()[1]))


def handle(conn):
    if mode == "silent":
        while conn.recv(4096):
            pass
        return
    data = b""
    while b"\r\n\r\n" not in data:
        chunk = conn.recv(4096)
        if not chunk:
            return
        data += chunk
    key = [l.split(b":", 1)[1].strip() for l in data.split(b"\r\n") if l.lower().startswith(b"sec-websocket-key")][0]
    accept = base64.b64encode(hashlib.sha1(key + b"258EAFA5-E914-47DA-95CA-C5AB0DC85B11").digest())
    conn.sendall(
        b"HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "
        + accept + b"\r\n\r\n"
    )
    while conn.recv(4096):
        pass


while True:
    c, _ = server.accept()
    threading.Thread(target=handle, args=(c,), daemon=True).start()

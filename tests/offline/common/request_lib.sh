# Helpers for request_parity.yaml (source it from a test command).
#
# start_server [mode] starts request_server.py and sets
#   C8Y_HOST  the server, which is also the session's host
#   LOG       the file with the requests it received
# start_other [mode] starts a second server and sets OTHER_HOST and OTHER_LOG.
# The servers are stopped when the test shell exits.
_servers=""
_start() {
    local portfile
    portfile=$(mktemp)
    _log=$(mktemp)
    python3 ./offline/common/request_server.py "$portfile" "$_log" "${1:-ok}" &
    _servers="$_servers $!"
    for _ in $(seq 1 50); do [ -s "$portfile" ] && break; sleep 0.1; done
    _host="http://127.0.0.1:$(cat "$portfile")"
    trap 'kill $_servers 2>/dev/null' EXIT
}

start_server() {
    _start "$@"
    C8Y_HOST=$_host
    LOG=$_log
    export C8Y_HOST LOG
}

start_other() {
    _start "$@"
    OTHER_HOST=$_host
    OTHER_LOG=$_log
}

# requests prints the request line of every request in the log
requests() {
    grep -E '^(GET|POST|PUT|DELETE) ' "${1:-$LOG}"
}

# count_requests prints the number of requests in the log
count_requests() {
    requests "$@" | wc -l | tr -d ' '
}

# check_waits <seconds>... prints "ok" if the server got the requests with the given waits
# between them (up to 0.2s longer), otherwise the waits it measured
check_waits() {
    python3 - "$LOG.times" "$@" <<'PY'
import sys
t = [float(l) for l in open(sys.argv[1])]
waits = [b - a for a, b in zip(t, t[1:])]
want = [float(w) for w in sys.argv[2:]]
ok = len(waits) == len(want) and all(w <= got < w + 0.2 for got, w in zip(waits, want))
print("ok" if ok else "waits: " + " ".join("%.3f" % w for w in waits))
PY
}

# verbose_log runs a command and prints its log (stderr) without the time and level of each line,
# and without the lines which depend on the environment. A duration is printed as Nms.
# The log lines are "<time><tab><level><tab><message>" (POSIX sh: the test shell can be dash).
verbose_log() {
    "$@" 2>&1 >/dev/null |
        cut -f3- |
        sed -E 's/in [0-9]+ms/in Nms/; s/^Response time: [0-9]+ms/Response time: Nms/' |
        grep -v -E '^(Binding authorization|Failed to read configuration|command:|activityLog path|activityLog is disabled)'
}

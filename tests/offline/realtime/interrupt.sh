#!/bin/bash
# Ctrl-C while the realtime websocket connects: the command ends at once (the handshake
# timeout of the realtime websocket is 10 seconds, and the realtime handshake waits 10 more).
# Checked against a server which upgrades the websocket but never answers the realtime
# handshake, and against one which never answers.
set -eou pipefail

dir=$(mktemp -d)
pids=""
trap 'kill $pids 2>/dev/null; rm -rf "$dir"' EXIT

ok=true
for mode in upgrade silent; do
    python3 ./offline/realtime/handshake_stub.py "$dir/$mode.port" "$mode" &
    pids="$pids $!"
    for _ in $(seq 1 50); do [ -s "$dir/$mode.port" ] && break; sleep 0.1; done
    result=$(C8Y_HOST="http://127.0.0.1:$(cat "$dir/$mode.port")" python3 ./offline/realtime/interrupt.py 1.5 c8y realtime subscribe --channel /alarms/1 --duration 60s)
    seconds=${result% *}
    # well below the 10 seconds of the timeout
    if ! python3 -c "import sys; sys.exit(0 if float('$seconds') < 6 else 1)"; then
        ok=false
        echo "$mode: $result" >&2
    fi
done
echo "interrupted_in_time=$ok"

#!/bin/bash
# Run the dry-run CLI tests (tests/auto) against the go-c8y fake server, so no
# Cumulocity tenant is needed.
#
# The fake server (github.com/reubenmiller/go-c8y/v2/cmd/c8y-fakeserver) is an
# in-memory Cumulocity built from the go-c8y version in go.mod. The dry-run
# tests still look up names (devices, groups, users, ...) for real, so the
# usual test data is created on it first with tests/scripts/setup.sh.
#
# Usage:
#   tests/run-offline.sh [test-folder] [commander args...]
#
# C8Y_OFFLINE_SUITE=manual runs a tests/manual folder instead of tests/auto
# (for manual tests that the fake server supports, e.g. alarms/subscribe).
#
# The c8y binary under test is the first one on PATH.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORK_DIR="$(mktemp -d)"

FAKESERVER="${C8Y_FAKESERVER:-$WORK_DIR/c8y-fakeserver}"
if [ ! -x "$FAKESERVER" ]; then
    echo "Building the go-c8y fake server"
    (cd "$ROOT_DIR" && go build -o "$FAKESERVER" github.com/reubenmiller/go-c8y/v2/cmd/c8y-fakeserver)
fi

if ! command -v c8y >/dev/null; then
    echo "could not find c8y in PATH" >&2
    exit 1
fi

"$FAKESERVER" --addr 127.0.0.1:0 >"$WORK_DIR/fakeserver.log" 2>&1 &
FAKESERVER_PID=$!
cleanup() {
    kill "$FAKESERVER_PID" 2>/dev/null || true
    rm -rf "$WORK_DIR"
}
trap cleanup EXIT

HOST=""
for _ in $(seq 1 50); do
    HOST="$(sed -n 's/^C8Y_HOST=//p' "$WORK_DIR/fakeserver.log" | head -n1)"
    [ -n "$HOST" ] && break
    sleep 0.1
done
if [ -z "$HOST" ]; then
    echo "the fake server did not start" >&2
    cat "$WORK_DIR/fakeserver.log" >&2
    exit 1
fi
echo "Fake server: $HOST"

# Isolate the CLI from the user's sessions and settings
unset C8Y_SESSION C8Y_TOKEN C8Y_USERNAME
export C8Y_SESSION_HOME="$WORK_DIR/session-home"
mkdir -p "$C8Y_SESSION_HOME"
export C8Y_HOST="$HOST"
export C8Y_TENANT=t12345
export C8Y_USER=admin
export C8Y_PASSWORD=admin-pass
export C8Y_SETTINGS_CI=true

echo "Creating test data"
if ! "$SCRIPT_DIR/scripts/setup.sh" >"$WORK_DIR/setup.log" 2>&1 </dev/null; then
    echo "test data setup failed" >&2
    cat "$WORK_DIR/setup.log" >&2
    exit 1
fi

if [ "${C8Y_OFFLINE_SUITE:-auto}" = manual ]; then
    folder="${1:-}"
    [ $# -gt 0 ] && shift
    cd "$SCRIPT_DIR"
    export TEST_SHELL=bash
    commander test --config ./config.manual.yaml "$@" --dir "manual/$folder" </dev/null
else
    "$SCRIPT_DIR/run-auto.sh" "$@" </dev/null
fi

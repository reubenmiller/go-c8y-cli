#!/bin/bash
# Streams realtime alarms with a subscribe command and prints how many arrived.
#
# usage: stream.sh <alarms|realtime|subscribeAll|notification2> [extra flags...]
#
# The subscribe command runs for 4s (--duration). An alarm is created while it
# runs. TLS: with the fake server in TLS mode (C8Y_OFFLINE_TLS=1) the commands
# which stream use the insecure setting of the environment as given.
set -eou pipefail

KIND="$1"
shift
DURATION=4

device=$(c8y devices create --name "stream-$KIND-$RANDOM" --select id --output csv)
out=$(mktemp)
trap 'rm -f "$out"' EXIT

case "$KIND" in
    alarms) subscribe=(c8y alarms subscribe --device "$device") ;;
    realtime) subscribe=(c8y realtime subscribe --channel "/alarms/$device") ;;
    subscribeAll) subscribe=(c8y realtime subscribeAll --device "$device") ;;
    *)
        echo "unknown kind: $KIND" >&2
        exit 2
        ;;
esac

start=$(date +%s)
"${subscribe[@]}" --duration "${DURATION}s" "$@" >"$out" 2>"$out.err" </dev/null &
pid=$!
sleep 1.5
c8y alarms create --device "$device" --type c8y_StreamTest --text "stream test" --severity MINOR --time now >/dev/null
rc=0
wait "$pid" || rc=$?
elapsed=$(( $(date +%s) - start ))

received=$(grep -c c8y_StreamTest "$out" || true)
echo "exit=$rc received=$received elapsed_ok=$([ "$elapsed" -ge "$((DURATION - 1))" ] && echo true || echo false)"
rm -f "$out.err"

#!/bin/bash
set -eou pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

debug () { >&2 echo "$@"; }
command -v c8y >/dev/null || exit 11
command -v nohup >/dev/null || exit 12

export C8Y_SETTINGS_DEFAULTS_DRY=false

mo_id=$( c8y agents create --template "{name: 'agent_' + _.Char(10)}" --select id --output csv )
debug "mo_id: $mo_id"
[[ "$mo_id" =~ ^[0-9]+$ ]] || exit 10

TASK_PID=""
cleanup () {
    kill -9 $TASK_PID 2>&1 >/dev/null || true
    wait $TASK_PID 2>/dev/null || true
    c8y inventory delete --id $mo_id >/dev/null 2>&1 || true
}
trap cleanup EXIT

nohup ./create.operations.sh $mo_id 60 >/dev/null 2>&1 &
TASK_PID="$!"
debug "TASK_PID: $TASK_PID"

values=$( c8y operations subscribe --device $mo_id --duration 10s || true )
item_count=$( echo "$values" | grep "^{" | wc -l )

# result:
debug "line_count: $item_count"
echo "{\"itemCount\":\"$item_count\"}"

# error is 100 + actual line count (to make debugging easier)
[[ $item_count -gt 0 ]] || exit $(( 100 + $item_count ))

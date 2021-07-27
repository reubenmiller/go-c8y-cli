#!/bin/bash
set -eou pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

mo_id=$( c8y agents create --template "{name: 'agent_' + _.Char(10)}" --select id --output csv )

TASK_PID=""
cleanup () {
    kill -9 $TASK_PID 2>&1 >/dev/null || true
    wait $TASK_PID 2>/dev/null || true
    c8y inventory delete --id $mo_id >/dev/null 2>&1 || true
}
trap cleanup EXIT

nohup ./create.operations.sh $mo_id 60 >/dev/null 2>&1 &
TASK_PID="$!"

# starttime=$( date +%s )
values=$( c8y operations subscribe --device $mo_id --duration 10s )
line_count=$( echo "$values" | grep "^{" | wc -l )
[[ $line_count -gt 1 ]] || exit 2

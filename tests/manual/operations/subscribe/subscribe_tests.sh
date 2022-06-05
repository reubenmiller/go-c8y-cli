#!/bin/bash
set -eou pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

TEMPLATE="./template.jsonnet"

debug () { >&2 echo "$@"; }
command -v c8y >/dev/null || exit 11
command -v nohup >/dev/null || exit 12

export C8Y_SETTINGS_DEFAULTS_DRY=false
export C8Y_SETTINGS_DEFAULTS_ABORTONERRORS=1

TASK_PID=()
start_task () {
    local COMMAND="$@"
    debug "Executing: $COMMAND"
    bash -c "$COMMAND >/dev/null 2>&1" &
    last_task="$!"
    TASK_PID+=("$last_task")
    debug "$TASK_PID"
}

stop_task () {
    # Kill parent/children (add - before pid)
    for pid in "${TASK_PID[@]}"
    do
        kill -9 "-${pid}" || true
    done
}

mo_id=

setup () {
    mo_id=$( c8y agents create --template "{name: 'agent_' + _.Char(10)}" --select id --output csv )
    debug "mo_id: $mo_id"
    echo "$mo_id" | c8y assert text --regex "^\d+$" || exit 10
    trap cleanup EXIT
}

cleanup () {
    stop_task
    c8y inventory delete --id "$mo_id" --silentStatusCodes 404 > /dev/null 2>&1 || true
}

test01 () {
    # Watch for a time period
    start_task "seq 1 10 | c8y $SUBCOMMAND create --device $mo_id --template $TEMPLATE --delay 2s --force"

    values=$( c8y $SUBCOMMAND subscribe --device "$mo_id" --duration 10s || true )
    item_count=$( echo "$values" | c8y util show --select id -o csv | wc -l )

    # result:
    c8y template execute --template "{itemCount: '$item_count', deviceId: '$mo_id', taskId: '${TASK_PID}'}"

    # error is 100 + actual line count (to make debugging easier)
    [[ $item_count -gt 0 ]] || exit $(( 100 + $item_count ))
}

test02 () {
    # Watch for a number of objects
    # subscribe to count
    start_task "seq 1 10 | c8y $SUBCOMMAND create --device $mo_id --template $TEMPLATE --delay 2s --force"

    values=$( c8y $SUBCOMMAND subscribe --device "$mo_id" --count 2 || true )
    item_count=$( echo "$values" | c8y util show --select id -o csv | wc -l )

    # result:
    c8y template execute --template "{itemCount: '$item_count', deviceId: '$mo_id', taskId: '${TASK_PID}'}"

    # error is 100 + actual line count (to make debugging easier)
    [[ "$item_count" -eq 2 ]] || exit $(( 100 + $item_count ))
}

test03 () {
    # Watch for a number of objects
    # subscribe to count
    start_task "seq 1 10 | c8y $SUBCOMMAND create --device $mo_id --template $TEMPLATE --delay 2s --force"

    values=$( c8y $SUBCOMMAND subscribe --count 2 || true )
    item_count=$( echo "$values" | c8y util show --select id -o csv | wc -l )

    # result:
    c8y template execute --template "{itemCount: '$item_count', deviceId: '$mo_id', taskId: '${TASK_PID}'}"

    # error is 100 + actual line count (to make debugging easier)
    [[ "$item_count" -eq 2 ]] || exit $(( 100 + $item_count ))
}

SUBCOMMAND=$1
shift

while [ $# -gt 0 ]
do
    RUN=$1
    shift
    case "$RUN" in

        1)
            setup
            test01
            ;;

        2)
            setup
            test02
            ;;

        3)
            setup
            test03
            ;;
        *)
            echo "Importing files"
    esac
done

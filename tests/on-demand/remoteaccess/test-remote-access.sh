#!/usr/bin/env bash
set -ex

#
# On demand test script to check the remoteaccess feature
# Due to dependencies it is hard to automate this, however this may change in the future
# by using thin-edge.io as a test client to automate again in a containerized environment.
#

if [ $# -lt 1 ]; then
    fail "Missing required argument"
fi

DEVICE="$1"

SSH_USER=root
if [ $# -gt 1 ]; then
    SSH_USER="$2"
fi

fail() {
    echo "FAILED: $*" >&2
    exit 1
}

test_configuration() {
    echo "Running tests: c8y remoteaccess configurations" >&2

    OUTPUT=$(c8y remoteaccess configurations list --device "$DEVICE" --select id,name,protocol -o csv)
    [ -n "$OUTPUT" ] || fail "Expected device to have at least 1 configuration"

    #
    # completions
    #
    # list
    LINE_COUNT=$(c8y __complete remoteaccess configurations list --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    # get
    LINE_COUNT=$(c8y __complete remoteaccess configurations get --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    LINE_COUNT=$(c8y __complete remoteaccess configurations get --device "$DEVICE" --id "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    # delete
    LINE_COUNT=$(c8y __complete remoteaccess configurations delete --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    LINE_COUNT=$(c8y __complete remoteaccess configurations delete --device "$DEVICE" --id "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    # create-*
    LINE_COUNT=$(c8y __complete remoteaccess configurations create-passthrough --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    LINE_COUNT=$(c8y __complete remoteaccess configurations create-telnet --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    LINE_COUNT=$(c8y __complete remoteaccess configurations create-vnc --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"

    LINE_COUNT=$(c8y __complete remoteaccess configurations create-webssh --device "" | wc -l | xargs)
    [ "$LINE_COUNT" -gt 2 ] || fail "Expected completion to have greater than 2 lines. got=$LINE_COUNT"
}

test_connect_ssh() {
    echo "Running tests: c8y remoteaccess connect ssh" >&2
    set +e
    OUTPUT=$(c8y remoteaccess connect ssh --device "$DEVICE" --user "$SSH_USER" -- exit 32)
    LAST_EXIT_CODE="$?"
    [[ "$LAST_EXIT_CODE" -eq 32 ]] || fail "Exit code did not match. got=$LAST_EXIT_CODE, expected=32"
    set -e

    # Only messages from the ssh command are included in stdout
    OUTPUT=$(c8y remoteaccess connect ssh --device "$DEVICE" --user "$SSH_USER" -- echo hello world)
    [[ "$OUTPUT" = "hello world" ]] || fail "Output did not match. got=$OUTPUT, expected='hello world'"

    # stderr is not returned by default
    OUTPUT=$(c8y remoteaccess connect ssh --device "$DEVICE" --user "$SSH_USER" -- "sh -c 'echo hello world >&2'")
    [[ -z "$OUTPUT" ]] || fail "Output did not match. got=$OUTPUT, expected=<empty>"

    # stderr can be redirected to stdout
    OUTPUT=$(c8y remoteaccess connect ssh --device "$DEVICE" --user "$SSH_USER" -- "sh -c 'echo hello world >&2'" 2>&1)
    if ! echo "$OUTPUT" | grep "hello world"; then
        fail "Expected output to contain 'hello world'. got=$OUTPUT"
    fi

    # The default user can be set via env variables
    eval "$(c8y settings update remoteaccess.sshuser root --shell auto)"
    c8y remoteaccess connect ssh --device "$DEVICE" -- exit 0
    unset C8Y_SETTINGS_REMOTEACCESS_SSHUSER
}

test_connect_run() {
    echo "Running tests: c8y remoteaccess connect run" >&2
    #
    # Run a custom command
    #
    OUTPUT=$(c8y remoteaccess connect run --device "$DEVICE" -- echo host=%h,port=%p)
    [[ "$OUTPUT" =~ ^host=127.0.0.1,port=[0-9]+$ ]] || fail "Did not match host/port pattern"

    # shellcheck disable=SC2016
    OUTPUT=$(c8y remoteaccess connect run --device "$DEVICE" -- sh -c 'echo host=$TARGET,port=$PORT')
    [[ "$OUTPUT" =~ ^host=127.0.0.1,port=[0-9]+$ ]] || fail "Did not match host/port pattern"

    # Run custom ssh command which also runs another command on the device
    OUTPUT=$(c8y remoteaccess connect run --device "$DEVICE" -- ssh -p %p "$SSH_USER@%h" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -- cat /etc/os-release)    
    [ -n "$OUTPUT" ] || fail "Empty ssh response"

    # Check if exit code is propogated
    set +e
    OUTPUT=$(c8y remoteaccess connect run --device "$DEVICE" -- sh -c 'exit 31')
    LAST_EXIT_CODE="$?"
    [[ "$LAST_EXIT_CODE" -eq 31 ]] || fail "Exit code did not match. got=$LAST_EXIT_CODE, expected=31"
    set -e

    # Exit code 0 does not produce an error
    c8y remoteaccess connect run --device "$DEVICE" -- sh -c 'exit 0'
}


main() {
    test_configuration
    test_connect_ssh
    test_connect_run
}

main

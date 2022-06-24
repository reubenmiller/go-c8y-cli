#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false
export C8Y_SETTINGS_CI=true
export C8Y_SETTINGS_DEFAULTS_CACHE=false

createdir () {
    # Cross-platform compatible
    local name="${1:-"c8y-temp"}"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t "$name")
    echo "$tmpdir"
}

export TEMP_DIR=$(createdir)
IDS=()

cleanup () {
    exit_status=$?
    rm -Rf "$TEMP_DIR"
    for i in "${IDS[@]}"; do
        c8y events delete --id "$i" || true
    done
    exit "$exit_status"
}

trap cleanup EXIT

test01 () {

    file1="$TEMP_DIR/file_original.txt"
    file2="$TEMP_DIR/file_updated.txt"
    printf "Line 1\nLine 2\n" > "$file1"
    printf "Line 3\nLine 4\n" > "$file2"

    event_id=$(c8y events create --text "with attachment" --type "ci_events_createBinary" --device "device01" --select id -o csv)
    c8y events createBinary --id "$event_id" --file "$file1"
    IDS+=("$event_id")

    # Download first time
    c8y events downloadBinary --id "$event_id" --outputFileRaw "$TEMP_DIR/downloaded.file1" > /dev/null
    cmp --silent "$file1" "$TEMP_DIR/downloaded.file1"

    # Update and download again
    c8y events updateBinary --id "$event_id" --file "$file2"

    c8y events downloadBinary --id "$event_id" --outputFileRaw "$TEMP_DIR/downloaded.file2" > /dev/null
    cmp --silent "$file2" "$TEMP_DIR/downloaded.file2"
}

test01

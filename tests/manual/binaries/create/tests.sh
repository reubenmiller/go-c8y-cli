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
    for i in "${IDS[@]}"; do
        c8y binaries delete --id "$i" || true
    done
    exit "$exit_status"
}

trap cleanup EXIT

test01 () {

    file1="$TEMP_DIR/file_original.txt"
    file2="$TEMP_DIR/file_updated.txt"
    printf "Line 1\nLine 2\n" > "$file1"
    printf "Line 3\nLine 4\n" > "$file2"

    binary_id=$(c8y binaries create --file "$file1" --select id -o csv)
    IDS+=("$binary_id")

    # Download first time
    c8y binaries get --id "$binary_id" --outputFileRaw "$TEMP_DIR/downloaded.file1" > /dev/null
    cmp --silent "$file1" "$TEMP_DIR/downloaded.file1"

    # Update and download again
    binary2_id=$(c8y binaries update --id "$binary_id" --file "$file2" --select id -o csv)
    c8y binaries get --id "$binary2_id" --outputFileRaw "$TEMP_DIR/downloaded.file2" > /dev/null
    cmp --silent "$file2" "$TEMP_DIR/downloaded.file2"
}

test01

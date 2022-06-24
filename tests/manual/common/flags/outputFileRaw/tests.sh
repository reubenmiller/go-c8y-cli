#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

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


create_inventory_binary () {
    local name="$1"
    local binary_id=
    echo "Dummy content" > "$TEMP_DIR/binary.txt"
    binary_id=$(c8y binaries create --file "$TEMP_DIR/binary.txt" --name "$name" --select id -o csv)
    IDS+=("$binary_id")
    echo "$binary_id"    
}

test01 () {
    binary_id=$(create_inventory_binary "mycustomfilename.py")
    c8y binaries get --id "$binary_id" --outputFileRaw "$TEMP_DIR/output/prefix-{id}.{filename}" > /dev/null
    test -f "$TEMP_DIR/output/prefix-$binary_id.mycustomfilename.py"
}

test01

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
    c8y binaries get --id "$binary_id" --outputFileRaw "$TEMP_DIR/output/prefix-{id}.{filename}" --verbose
    test -f "$TEMP_DIR/output/prefix-$binary_id.mycustomfilename.py"
}

test02_non_api_commands () {
    printf '{"name":"foo1","type":"type1","owner":"owner1"}' | c8y util show --outputFileRaw "$TEMP_DIR/output/prefix-{name}-{type}-{owner}.json" > /dev/null
    test -f "$TEMP_DIR/output/prefix-foo1-type1-owner1.json"

    printf '{"name":"foo2","type":"type2","owner":"owner2"}' | c8y template execute --template 'input.value' --outputFileRaw "$TEMP_DIR/output/prefix-{name}-{type}-{owner}.json" > /dev/null
    test -f "$TEMP_DIR/output/prefix-foo2-type2-owner2.json"
}

test01
test02_non_api_commands

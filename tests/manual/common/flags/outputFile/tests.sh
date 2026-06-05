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

create_inventory () {
    local name="$1"
    local type="$2"
    local mo_id=

    mo_id=$( c8y inventory create -n -f --name "$name" --type "$type" --select id --output csv )
    IDS+=("$mo_id")
    echo "$mo_id"    
}

test01_api_command () {
    NAME=$(c8y template execute -n --template "'ci_' + _.Char(10)")
    TYPE=$(c8y template execute -n --template "'ci_type_' + _.Char(10)")
    mo_id=$(create_inventory "$NAME" "$TYPE")
    c8y inventory get -n --id "$mo_id" --select id,name,type --outputFile "$TEMP_DIR/output/prefix-{id}.{name}.{type}.json" --verbose
    test -f "$TEMP_DIR/output/prefix-${mo_id}.${NAME}.${TYPE}.json"
}

test02_non_api_commands () {
    printf '{"name":"foo1","type":"type1","owner":"owner1"}' | c8y util show --outputFile "$TEMP_DIR/output/prefix-{name}-{type}-{owner}.json" > /dev/null
    test -f "$TEMP_DIR/output/prefix-foo1-type1-owner1.json"

    printf '{"name":"foo2","type":"type2","owner":"owner2"}' | c8y template execute --template 'input.value' --outputFile "$TEMP_DIR/output/prefix-{name}-{type}-{owner}.json" > /dev/null
    test -f "$TEMP_DIR/output/prefix-foo2-type2-owner2.json"
}

test02_non_api_commands

#!/bin/bash

BIN_DIR="./output"

export C8Y_SETTINGS_DEFAULTS_FORCE=true
export PATH="$(cd "$BIN_DIR"; pwd):$PATH"

if ! command -v c8y; then
    if [[ -f "$BIN_DIR/c8y.linux" ]]; then
        cp "$BIN_DIR/c8y.linux" "$BIN_DIR/c8y"
    fi
fi

setup () {
    echo "Setting up c8y dependencies"
    create_user "peterpi@example.com"
    create_user "benhologram@example.com"
    create_user "tomwillow@example.com"

    create_smartgroup "my smartgroup"

    create_app "my-example-app"
}

create_user () {
    local username="$1"
    
    c8y users get --id $username --silentStatusCodes 404 || c8y users create \
        --email $username \
        --userName $username \
        --template "{password: _.Password()}"
}

create_app () {
    local name="$1"
    c8y applications get --id "$name" --silentStatusCodes 404 ||
        c8y applications create \
            --name "$name" \
            --type HOSTED \
            --key "$name-key" \
            --contextPath "$name"
}

create_smartgroup () {
    local name="$1"
    c8y smartgroups get --id "$name" --silentStatusCodes 404 ||
        c8y smartgroups create \
            --name "$name" \
            --query "name eq '*'"
}

setup

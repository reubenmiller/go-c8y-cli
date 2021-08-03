#!/bin/bash

BIN_DIR="./output"

export C8Y_SETTINGS_DEFAULTS_FORCE=true

if ! command -v c8y; then
    echo "could not find c8y in path. PATH=$PATH"
    exit 1
fi

setup () {
    echo "Setting up c8y dependencies"
    create_user "peterpi@example.com"
    create_user "benhologram@example.com"
    create_user "tomwillow@example.com"

    create_agent "device01"
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

create_agent () {
    local name="$1"
    c8y agents get --id "$name" --silentStatusCodes 404 ||
        c8y agents create \
            --name "$name"
}

setup

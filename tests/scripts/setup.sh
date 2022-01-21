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

    create_agent "agent01"
    create_agent "device01"
    create_smartgroup "my smartgroup"

    create_app "my-example-app"

    create_firmware "iot-linux"
    create_firmware_version "iot-linux" "1.0.0" "https://example.com"

    create_software "my-app"
    create_software_version "my-app" "1.2.3" "https://example.com/debian/my-app-1.2.3.deb"

    create_device_profile "profile01"
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

create_firmware () {
    local name="$1"
    c8y firmware get --id "$name" --silentStatusCodes 404 ||
        c8y firmware create --name "$name"
}

create_firmware_version () {
    local name="$1"
    local version="$2"
    local url="$3"
    c8y firmware versions get --firmware "$name" --id "$version" --silentStatusCodes 404 ||
        c8y firmware versions create --firmware "$name" --version "$version" --url "$url"
}

create_software () {
    local name="$1"
    c8y software get --id "$name" --silentStatusCodes 404 ||
        c8y software create --name "$name"
}

create_software_version () {
    local name="$1"
    local version="$2"
    local url="$3"
    c8y software versions get --software "$name" --id "$version" --silentStatusCodes 404 ||
        c8y software versions create --software "$name" --version "$version" --url "$url"
}

create_device_profile () {
    local name="$1"
    c8y deviceprofiles get --id "$name" --silentStatusCodes 404 ||
        c8y deviceprofiles create --name "$name"
}

setup

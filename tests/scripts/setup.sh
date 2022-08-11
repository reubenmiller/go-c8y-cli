#!/bin/bash

set -e

BIN_DIR="./output"

export C8Y_SETTINGS_DEFAULTS_FORCE=true
export C8Y_SETTINGS_DEFAULTS_VERBOSE=true

if ! command -v c8y; then
    echo "could not find c8y in path. PATH=$PATH"
    exit 1
fi

setup () {
    echo "Setting up c8y dependencies"
    create_user "peterpi@example.com"
    create_user "benhologram@example.com"
    create_user "tomwillow@example.com"

    create_usergroup "powerusers"
    create_usergroup "control-center"

    create_agent "agent01"
    create_agent "device01"
    create_smartgroup "my smartgroup"

    create_child_device "agentParent01" "child"
    create_device_with_assets "agentAssetInfo01" "childAsset"
    create_device_with_additions "agentAdditionInfo01" "childAddition"

    create_app "my-example-app"
    create_service_user "technician"

    create_firmware "iot-linux"
    create_firmware_version "iot-linux" "1.0.0" "https://example.com"

    create_software "my-app"
    create_software_version "my-app" "1.2.3" "https://example.com/debian/my-app-1.2.3.deb"

    create_device_profile "profile01"

    create_device_and_user "mobile-device01"

    create_devicecert "MyCert"
}

create_user () {
    local username="$1"
    
    c8y users get --id $username --silentStatusCodes 404 || c8y users create \
        --email $username \
        --userName $username \
        --template "{password: _.Password()}"
}

create_usergroup () {
    local name="$1"
    
    c8y usergroups get --id $name --silentStatusCodes 404 || c8y usergroups create \
        --name "$name"
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

create_service_user () {
    local appname="$1"

    local tenant=$(c8y currenttenant get --select name -o csv)
    c8y microservices get --id "$appname" --silentStatusCodes 404 ||
        c8y microservices serviceusers create \
            --name "$appname" \
            --tenants "$tenant"
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

create_mo_with_name () {
    local name="$1"

    existing_mo=$(c8y inventory find --query "name eq '$name'")

    if [[ -n "$existing_mo" ]]; then
        echo "$existing_mo"
        return
    fi

    c8y inventory create --name "$name"
}

create_child_device () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_agent "${childNamePrefix}01" | c8y inventory update --data 'type=customdevice' | c8y devices children assign --device "$parent" --silentStatusCodes 409 --silentExit
    create_agent "${childNamePrefix}02" | c8y inventory update --data 'type=customdevice' | c8y devices children assign --device "$parent" --silentStatusCodes 409 --silentExit
}

create_device_with_assets () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_mo_with_name "${childNamePrefix}01" | c8y inventory update --data 'type=custominfo' | c8y devices assets assign --device "$parent" --silentStatusCodes 409 --silentExit
    create_mo_with_name "${childNamePrefix}02" | c8y inventory update --data 'type=custominfo' | c8y devices assets assign --device "$parent" --silentStatusCodes 409 --silentExit
}

create_device_with_additions () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_mo_with_name "${childNamePrefix}01" | c8y inventory update --data 'type=custominfo' | c8y devices additions assign --device "$parent" --silentStatusCodes 409 --silentExit
    create_mo_with_name "${childNamePrefix}02" | c8y inventory update --data 'type=custominfo' | c8y devices additions assign --device "$parent" --silentStatusCodes 409 --silentExit
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

create_device_and_user () {
    local name="$1"
    local extType="c8y_Serial"

    c8y deviceregistration register --id "$name" || true
    c8y deviceregistration getCredentials --id "$name" --sessionUsername "$DEVICE_BOOTSTRAP_USER" --sessionPassword "$DEVICE_BOOTSTRAP_PASSWORD" || true
    c8y deviceregistration approve --id "$name"
    
    creds=$(
        c8y deviceregistration getCredentials \
            --id "$name" \
            --sessionUsername "$DEVICE_BOOTSTRAP_USER" \
            --sessionPassword "$DEVICE_BOOTSTRAP_PASSWORD" \
            --select username,password \
            --output csv
    )
    device_user=$( echo "$creds" | cut -d, -f1 )
    device_password=$( echo "$creds" | cut -d, -f2 )

    if ! c8y identity get --name "$name" --type "$extType" --silentStatusCodes 404; then
        c8y devices create \
            --name "$name" \
            --sessionUsername "$device_user" \
            --sessionPassword "$device_password" \
        | c8y identity create --type "$extType" --name "$name"
    fi

    c8y devices availability set --id "$name" --interval 15
}

create_devicecert () {
    local name="$1"

    c8y devicemanagement certificates get --id "$name" --silentStatusCodes 404 ||
        c8y devicemanagement certificates create --name "$name" --file tests/testdata/trustedcert.pem
}

setup

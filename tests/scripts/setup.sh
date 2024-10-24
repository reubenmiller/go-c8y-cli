#!/bin/bash

set -e

if [ -n "$CI" ]; then
    set -x
fi

BIN_DIR="./output"

export C8Y_SETTINGS_DEFAULTS_FORCE=true
export C8Y_SETTINGS_DEFAULTS_VERBOSE=false
export C8Y_SETTINGS_DEFAULTS_CACHE=false

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
    create_devicegroup "My Group"

    create_child_device "agentParent01" "child"
    create_device_with_assets "agentAssetInfo01" "childAsset"
    create_device_with_additions "agentAdditionInfo01" "childAddition"
    create_device_with_service "device02" "sshd" "systemd" "up"

    create_app "my-example-app"
    create_service_user "technician"

    create_configuration "example-config" "agentConfig" "https://test.com/content/raw/app.json"
    create_firmware "iot-linux"
    create_firmware_version "iot-linux" "1.0.0" "https://example.com"
    create_firmware_patch_version "iot-linux" "1.0.1" "https://example.com/patch1"

    create_software "my-app"
    create_software_version "my-app" "1.2.3" "https://example.com/debian/my-app-1.2.3.deb"

    create_device_profile "profile01"

    create_device_and_user "mobile-device01"

    create_devicecert "MyCert"
}

create_user () {
    local username="$1"
    
    c8y users get -n --id $username --silentStatusCodes 404 || c8y users create -n \
        --email $username \
        --userName $username \
        --template "{password: _.Password()}"
}

create_usergroup () {
    local name="$1"
    
    c8y usergroups get -n --id $name --silentStatusCodes 404 || c8y usergroups create -n \
        --name "$name"
}

create_app () {
    local name="$1"
    c8y applications get -n --id "$name" --silentStatusCodes 404 ||
        c8y applications create -n \
            --name "$name" \
            --type HOSTED \
            --key "$name-key" \
            --contextPath "$name"
}

create_service_user () {
    local appname="$1"

    local tenant=$(c8y currenttenant get -n --select name -o csv)
    c8y microservices get -n --id "$appname" --silentStatusCodes 404 ||
        c8y microservices serviceusers create -n \
            --name "$appname" \
            --tenants "$tenant"
}

create_smartgroup () {
    local name="$1"
    c8y smartgroups get -n --id "$name" --silentStatusCodes 404 ||
        c8y smartgroups create -n \
            --name "$name" \
            --query "name eq '*'"
}

create_devicegroup () {
    local name="$1"
    c8y devicegroups get -n --id "$name" --silentStatusCodes 404 ||
        c8y devicegroups create -n \
            --name "$name"
}

create_agent () {
    local name="$1"
    c8y agents get -n --id "$name" --silentStatusCodes 404 ||
        c8y agents create -n \
            --name "$name"
}

create_mo_with_name () {
    local name="$1"

    existing_mo=$(c8y inventory find -n --query "name eq '$name'")

    if [[ -n "$existing_mo" ]]; then
        echo "$existing_mo"
        return
    fi

    c8y inventory create -n --name "$name"
}

create_child_device () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_agent "${childNamePrefix}01" | c8y inventory update --data 'type=customdevice' | c8y devices children assign --childType device --id "$parent" --silentStatusCodes 409 --silentExit
    create_agent "${childNamePrefix}02" | c8y inventory update --data 'type=customdevice' | c8y devices children assign --childType device --id "$parent" --silentStatusCodes 409 --silentExit
}

create_device_with_assets () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_mo_with_name "${childNamePrefix}01" | c8y inventory update --data 'type=custominfo' | c8y devices children assign --childType asset --id "$parent" --silentStatusCodes 409 --silentExit
    create_mo_with_name "${childNamePrefix}02" | c8y inventory update --data 'type=custominfo' | c8y devices children assign --childType asset --id "$parent" --silentStatusCodes 409 --silentExit
}

create_device_with_additions () {
    local parentName=$1
    local childNamePrefix=$2
    local parent=

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    create_mo_with_name "${childNamePrefix}01" | c8y inventory update --data 'type=custominfo' | c8y devices children assign --childType addition --id "$parent" --silentStatusCodes 409 --silentExit
    create_mo_with_name "${childNamePrefix}02" | c8y inventory update --data 'type=custominfo' | c8y devices children assign --childType addition --id "$parent" --silentStatusCodes 409 --silentExit
}

create_device_with_service () {
    local parentName="$1"
    local serviceName="$2"
    local serviceType="$3"
    local serviceStatus="$4"

    parent=$(create_agent "$parentName" | c8y util show --select id --output csv )
    c8y devices services get -n --device "$parent" --id "$serviceName" --silentStatusCodes 404 ||
        c8y devices services create -n --device "$parent" --name "$serviceName" --serviceType "$serviceType" --status "$serviceStatus"
}

create_firmware () {
    local name="$1"
    c8y firmware get -n --id "$name" --silentStatusCodes 404 ||
        c8y firmware create -n --name "$name"
}

create_firmware_version () {
    local name="$1"
    local version="$2"
    local url="$3"
    c8y firmware versions get -n --firmware "$name" --id "$version" --silentStatusCodes 404 ||
        c8y firmware versions create -n --firmware "$name" --version "$version" --url "$url"
}

create_firmware_patch_version () {
    local name="$1"
    local version="$2"
    local dep_version="$3"
    local url="$4"
    c8y firmware patches get -n --firmware "$name" --id "$version" --silentStatusCodes 404 ||
        c8y firmware patches create -n --firmware "$name" --version "$version" --dependencyVersion "$dep_version" --url "$url"
}

create_configuration () {
    local name="$1"
    local configurationType="$2"
    local url="$3"
    c8y configuration get -n --id "$name" --silentStatusCodes 404 ||
        c8y configuration create -n \
            --name "$name" \
            --description "Example config" \
            --configurationType "$configurationType" \
            --url "$url"
}

create_software () {
    local name="$1"
    c8y software get -n --id "$name" --silentStatusCodes 404 ||
        c8y software create -n --name "$name"
}

create_software_version () {
    local name="$1"
    local version="$2"
    local url="$3"
    c8y software versions get -n --software "$name" --id "$version" --silentStatusCodes 404 ||
        c8y software versions create -n --software "$name" --version "$version" --url "$url"
}

create_device_profile () {
    local name="$1"
    c8y deviceprofiles get -n --id "$name" --silentStatusCodes 404 ||
        c8y deviceprofiles create -n --name "$name"
}

create_device_and_user () {
    local name="$1"
    local extType="c8y_Serial"

    c8y deviceregistration register -n --id "$name" || true
    c8y deviceregistration getCredentials -n --id "$name" --sessionUsername "$DEVICE_BOOTSTRAP_USER" --sessionPassword "$DEVICE_BOOTSTRAP_PASSWORD" || true
    c8y deviceregistration approve -n --id "$name"
    
    creds=$(
        c8y deviceregistration getCredentials -n \
            --id "$name" \
            --sessionUsername "$DEVICE_BOOTSTRAP_USER" \
            --sessionPassword "$DEVICE_BOOTSTRAP_PASSWORD" \
            --select username,password \
            --output csv
    )
    device_user=$( echo "$creds" | cut -d, -f1 )
    device_password=$( echo "$creds" | cut -d, -f2 )

    if ! c8y identity get -n --name "$name" --type "$extType" --silentStatusCodes 404; then
        c8y devices create -n \
            --name "$name" \
            --sessionUsername "$device_user" \
            --sessionPassword "$device_password" \
        | c8y identity create --type "$extType" --name "$name"
    fi

    c8y devices availability set -n --id "$name" --interval 15
}

create_devicecert () {
    local name="$1"

    c8y devicemanagement certificates get -n --id "$name" --silentStatusCodes 404 ||
        c8y devicemanagement certificates create -n --name "$name" --file tests/testdata/trustedcert.pem --silentStatusCodes 409 --silentExit
}

# Remove cache to avoid stale/invalid cache which makes debugging tests very difficult
c8y cache delete

setup

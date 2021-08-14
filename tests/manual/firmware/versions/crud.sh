#!/bin/bash

export C8Y_SETTINGS_DEFAULTS_DRY=false
# export C8Y_SETTINGS_DEFAULTS_FORCE=true

NAME=${1:-linux-firmware-typea}
VERSION=${2:-0.8.6}

# create
ID=$( c8y firmware create --name "$NAME" | c8y firmware versions create --version "$VERSION" --url "test.com" --select id --output csv )

# list
firmware=$( c8y firmware versions list --firmwareId "$NAME" --select "id,c8y_Firmware.version" --output csv )
echo "$firmware" | grep "$ID,$VERSION"

# get > delete
c8y firmware versions get --id "$ID" | c8y firmware versions delete

# TODO: get firmware versions get --id <name>: lookup fails ()
# c8y firmware versions get --id "$ID" | c8y firmware versions delete

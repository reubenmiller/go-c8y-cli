#!/bin/bash

export C8Y_SETTINGS_DEFAULTS_DRY=false
export C8Y_SETTINGS_DEFAULTS_FORCE=true

FIRMWARE_ID=$1
FIRMWARE_VERSION=$2
FIRMWARE_PATCH=$3
FIRMWARE_URL=$4

resp=$( c8y firmware patches create \
    --firmware "$FIRMWARE_ID" \
    --dependencyVersion "$FIRMWARE_VERSION" \
    --version "$FIRMWARE_PATCH" \
    --url "$FIRMWARE_URL" )
echo "$resp"
echo "$resp" | c8y inventory delete > /dev/null

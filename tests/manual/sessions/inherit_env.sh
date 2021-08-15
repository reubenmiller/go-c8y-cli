#!/bin/bash

set -e

export C8Y_SESSION=
export C8Y_HOST=
export C8Y_USERNAME=
export C8Y_PASSWORD=

export C8Y_USERNAME=dummyuser
export C8Y_PASSWORD=dummypassword
export C8Y_BASEURL=http://127.0.0.1:5000
export C8Y_SETTINGS_DEFAULTS_DRYFORMAT=json
export C8Y_SETTINGS_DEFAULTS_DRY=true

fail () {
    echo "$1"
    exit 1
}

# Session should not throw an error
[[ $( c8y sessions get ) ]] || fail "Session should not thrown an error"

# API calls should also work
resp=$( c8y inventory get --id 12345 )

[[ "$(echo "$resp" | c8y util show --select host --output csv )" == "$C8Y_BASEURL" ]] || fail "url did not match"
[[ "$(echo "$resp" | c8y util show --select pathEncoded --output csv )" == "/inventory/managedObjects/12345" ]] || fail "pathEncoded did not match"

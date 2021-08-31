#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

NAME=${1:-""}
VERSION=${2:-0.8.6}

if [[ -z "$NAME" ]]; then
    NAME=$( c8y template execute --template "{name: 'linux-deviceprofiles-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using deviceprofiles name: $NAME"

# create
ID=$( c8y deviceprofiles create --name "$NAME" --select id --output csv )

# get deviceprofiles
DEVICEPROFILE_ID=$( c8y deviceprofiles get --id "$NAME" --select id --output csv )
echo "$NAME" | c8y deviceprofiles get --select id --output csv | grep "^$DEVICEPROFILE_ID$"


# update deviceprofiles
c8y deviceprofiles update --id "$NAME" --deviceType "c8y_Linux" --select c8y_Filter.type --output csv | grep "^c8y_Linux$"


# completion (deviceprofiles)
c8y __complete deviceprofiles get --id "$NAME" | grep id:
c8y __complete deviceprofiles update --id "$NAME" | grep id:
c8y __complete deviceprofiles delete --id "$NAME" | grep id:


# list by pipeline
echo "$NAME" | c8y deviceprofiles list | c8y deviceprofiles get --select "id,name" --output csv | grep "^$ID,$NAME$"

# get and pipe to delete
c8y deviceprofiles get --id "$ID" | c8y deviceprofiles delete

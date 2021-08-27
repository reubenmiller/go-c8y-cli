#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

NAME=${1:-""}

if [[ -z "$NAME" ]]; then
    NAME=$( c8y template execute --template "{name: 'linux-configuration-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using configuration name: $NAME"

# create from url
CONFIG1=$( c8y configuration create --name "$NAME" --description "My custom config" --deviceType "CUSTOM_CONFIG" --url "https://test.com" --select id --output csv )

#
# create from file
#
package_file=$(mktemp /tmp/package-XXXXXX.json)
echo "dummy file" > "$package_file"
trap "rm -f $package_file" EXIT

CONFIG2=$( c8y configuration create --name "$NAME" --file "$package_file" --select "id,url" --output csv )
echo "$CONFIG2" | grep "^[0-9]\+,.*/inventory/binaries/[0-9]\+$"

# download
echo "$CONFIG2" | c8y configuration get | c8y api | grep "^dummy file$"


# update configuration
c8y configuration update --id "$NAME" --description "Example description" --select description --output csv | grep "^Example description$"
c8y configuration update --id "$NAME" --deviceType "myType" --select deviceType --output csv | grep "^myType$"


# completion
c8y __complete configuration get --id "$NAME" | grep id:

# list
c8y configuration list --select "id,url" --output csv | grep "$CONFIG1"

# delete
c8y configuration get --id "$NAME" | c8y configuration delete

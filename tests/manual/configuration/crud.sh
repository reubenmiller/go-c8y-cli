#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

NAME=${1:-""}

if [[ -z "$NAME" ]]; then
    NAME=$( c8y template execute --template "{name: 'linux configuration-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using configuration name: $NAME"

# create from url
CONFIG1=$( c8y configuration create --name "$NAME" --description "My custom config" --configurationType "CUSTOM_CONFIG" --deviceType "myDeviceType" --url "https://test.com" )
CONFIG1_ID=$( echo "$CONFIG1" | c8y util show --select id --output csv  )
echo "$CONFIG1" | c8y util show --select name -o csv | grep "$NAME"
echo "$CONFIG1" | c8y util show --select url -o csv | grep "https://test.com"
echo "$CONFIG1" | c8y util show --select deviceType -o csv | grep "myDeviceType"
echo "$CONFIG1" | c8y util show --select description -o csv | grep "My custom config"

#
# create from file
#
package_file=$(mktemp /tmp/package-XXXXXX.json)
echo "dummy file" > "$package_file"
trap "rm -f $package_file" EXIT

CONFIG2=$( c8y configuration create --name "${NAME}_2" --file "$package_file" --select "id,url" --output csv )
echo "$CONFIG2" | grep "^[0-9]\+,.*/inventory/binaries/[0-9]\+$"

# download
echo "$CONFIG2" | c8y configuration get | c8y api | grep "^dummy file$"


# update configuration
c8y configuration update --id "$NAME" --description "Example description" --select description --output csv | grep "^Example description$"
c8y configuration update --id "$NAME" --deviceType "myType" --select deviceType --output csv | grep "^myType$"

# Update configuration binary
package_file2=$(mktemp /tmp/package-XXXXXX.json)
echo "dummy file 2" > "$package_file2"
trap "rm -f $package_file2" EXIT
CONFIG2_ID=$( echo "$CONFIG2" | cut -d, -f1 )
echo "$CONFIG2" | c8y configuration update --file $package_file2 --select id --output csv | grep "^$CONFIG2_ID$"
echo "$CONFIG2" | c8y configuration get | c8y api | grep "^dummy file 2$"


# completion
c8y __complete configuration get --id "$NAME" | grep id:

# list
c8y configuration list --name "$NAME" --select "id,url" --pageSize 100 --output csv | grep "$CONFIG1_ID"

# delete
c8y configuration get --id "$NAME" | c8y configuration delete

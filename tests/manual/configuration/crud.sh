#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

createdir () {
    # Cross-platform compatible
    local name="${1:-"c8y-temp"}"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t "$name")
    echo "$tmpdir"
}

export TEMP_DIR=$(createdir)
trap "rm -Rf $TEMP_DIR" EXIT

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
package_file="$TEMP_DIR/package-1.json"
echo "dummy file" > "$package_file"

CONFIG2=$( c8y configuration create --name "${NAME}_2" --file "$package_file" --configurationType dummytype --select "id,url" --output csv )
echo "$CONFIG2" | grep "^[0-9]\+,.*/inventory/binaries/[0-9]\+$"

# download
echo "$CONFIG2" | c8y configuration get | c8y api | grep "^dummy file$"


# update configuration
c8y configuration update --id "$NAME" --description "Example description" --select description --output csv | grep "^Example description$"
c8y configuration update --id "$NAME" --deviceType "myType" --select deviceType --output csv | grep "^myType$"


# send configuration via id
CONFIG2_ID=$( echo "$CONFIG2" | cut -d, -f1 )
CONFIG2_URL=$( echo "$CONFIG2" | cut -d, -f2 )
c8y configuration send --device 1234 --configuration "$CONFIG2_ID" --dry --dryFormat json -c | grep -F "\"url\":\"$CONFIG2_URL\"" | grep -F '"type":"dummytype"'

# send configuration via name
CONFIG2_ID=$( echo "$CONFIG2" | cut -d, -f1 )
CONFIG2_URL=$( echo "$CONFIG2" | cut -d, -f2 )
c8y configuration send --device 1234 --configuration "${NAME}_2" --dry --dryFormat json -c | grep -F "\"url\":\"$CONFIG2_URL\"" | grep -F '"type":"dummytype"'


# Update configuration binary
package_file2="$TEMP_DIR/package-2.json"
echo "dummy file 2" > "$package_file2"
CONFIG2_ID=$( echo "$CONFIG2" | cut -d, -f1 )
echo "$CONFIG2" | c8y configuration update --file "$package_file2" --select id --output csv | grep "^$CONFIG2_ID$"
echo "$CONFIG2" | c8y configuration get | c8y api | grep "^dummy file 2$"


# completion
c8y __complete configuration get --id "$NAME" | grep id:

# list
c8y configuration list --name "$NAME" --select "id,url" --pageSize 100 --output csv | grep "$CONFIG1_ID"

# delete
c8y configuration get --id "$NAME" | c8y configuration delete

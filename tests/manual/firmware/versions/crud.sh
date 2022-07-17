#!/bin/bash

set -e

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
VERSION=${2:-0.8.6}

if [[ -z "$NAME" ]]; then
    NAME=$( c8y template execute --template "{name: 'linux firmware-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using firmware name: $NAME"

# create
ID=$( c8y firmware create --name "$NAME" | c8y firmware versions create --version "$VERSION" --url "test.com" --select id --output csv )

#
# create version by file (get details from package name)
#
package_file="$TEMP_DIR/package-XXXXXX-10.2.3.deb"
echo "dummy file" > "$package_file"

VERSION2_ID=$( c8y firmware versions create --firmware "$NAME" --file "$package_file" --select "id,c8y_Firmware.version" --output csv )
echo "$VERSION2_ID" | grep "^[0-9]\+,10.2.3$"

# download
echo "$VERSION2_ID" | c8y firmware versions get | c8y api | grep "^dummy file$"


# update firmware
c8y firmware update --id "$NAME" --description "Example description" --select description --output csv | grep "^Example description$"

# completion (firmware and version)
c8y __complete firmware get --id "$NAME" | grep id:
c8y __complete firmware versions get --firmware "$NAME" --id "$VERSION" | grep id:

# list versions by pipeline
c8y firmware get --id "$NAME" | c8y firmware versions list --select "id,c8y_Firmware.version" --output csv | grep "$ID,$VERSION"

# list
c8y firmware versions list --firmware "$NAME" --select "id,c8y_Firmware.version" --output csv | grep "$ID,$VERSION"

# get > delete
c8y firmware versions get --id "$ID" | c8y firmware versions delete

# delete parent
c8y firmware get --id "$NAME" | c8y firmware delete

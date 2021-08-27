#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

FIRMWARE=${1:-""}
VERSION=${2:-0.8.0}
PATCH=${2:-0.8.1}

if [[ -z "$FIRMWARE" ]]; then
    FIRMWARE=$( c8y template execute --template "{name: 'linux-firmware-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using firmware name: $FIRMWARE"

# create patch
FIRMWARE_ID=$( c8y firmware create --name "$FIRMWARE" --select id --output csv )
VERSION_ID=$( echo "$FIRMWARE_ID" | c8y firmware versions create --version "$VERSION" --url "test.com" --select id --output csv )

# create patch with url
PATCH1_ID=$( c8y firmware get --id $FIRMWARE_ID | c8y firmware patches create --version "$PATCH" --dependencyVersion "$VERSION" --url "https://test.com" --select id --output csv )
# PATCH_ID=$( c8y firmware patches create --firmwareId $FIRMWARE_ID --dependencyVersion $VERSION_ID --version $PATCH )


#
# create patch from file (get details from package name)
#
package_file=$(mktemp /tmp/package-XXXXXX-10.2.3.deb)
echo "dummy file" > "$package_file"
trap "rm -f $package_file" EXIT

PATCH2_ID=$( c8y firmware patches create --firmwareId "$FIRMWARE" --file "$package_file" --dependencyVersion "$VERSION" --select "id,c8y_Patch.dependency,c8y_Firmware.version" --output csv )
echo "$PATCH2_ID" | grep "^[0-9]\+,$VERSION,10.2.3$"

# download
echo "$PATCH2_ID" | c8y firmware patches get | c8y api | grep "^dummy file$"


# completion (firmware and version)
c8y __complete firmware patches get --id "$PATCH" | grep id:
c8y __complete firmware patches get --firmwareId "$FIRMWARE" | grep id:

c8y __complete firmware patches delete --id "$PATCH" | grep id:
c8y __complete firmware patches delete --firmwareId "$FIRMWARE" | grep id:

c8y __complete firmware patches create --firmwareId "$FIRMWARE" | grep id:
c8y __complete firmware patches create --dependencyVersion "$VERSION" | grep id:

c8y __complete firmware patches list --firmwareId "$FIRMWARE" | grep id:


# list patches by pipeline
c8y firmware get --id "$FIRMWARE" | c8y firmware patches list --select "id,c8y_Firmware.version" --output csv | grep "$PATCH1_ID,$PATCH"

# list patches with dependency filter
c8y firmware get --id "$FIRMWARE" | c8y firmware patches list --select "id,c8y_Firmware.version" --output csv | grep "$PATCH1_ID,$PATCH"

# list
c8y firmware patches list --firmwareId "$FIRMWARE" --select "id,c8y_Firmware.version" --output csv | grep "$PATCH1_ID,$PATCH"

# list with dependency filter
c8y firmware patches list --firmwareId "$FIRMWARE" --dependency "$VERSION*" --select "id,c8y_Firmware.version" --output csv | grep "$PATCH1_ID,$PATCH"

# get > delete
c8y firmware patches get --id "$PATCH1_ID" | c8y firmware patches delete

# delete parent
c8y firmware get --id "$FIRMWARE" | c8y firmware delete

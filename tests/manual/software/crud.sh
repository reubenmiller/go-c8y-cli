#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

NAME=${1:-""}
VERSION=${2:-0.8.6}

if [[ -z "$NAME" ]]; then
    NAME=$( c8y template execute --template "{name: 'linux-software-typea_' + _.Char(8)}" --select name --output csv )
fi

echo "Using software name: $NAME"

# create
ID=$( c8y software create --name "$NAME" | c8y software versions create --version "$VERSION" --url "https://test.com" --select id --output csv )

# get software
SOFTWARE_ID=$( c8y software get --id "$NAME" --select id --output csv )
echo "$NAME" | c8y software get --select id --output csv | grep "^$SOFTWARE_ID$"


# update software
c8y software update --id "$NAME" --description "New description" --select description --output csv | grep "^New description$"

#
# create version by file (get details from package name)
#
package_file=$(mktemp /tmp/package-XXXXXX-10.2.3.deb)
echo "dummy file" > "$package_file"
trap "rm -f $package_file" EXIT

VERSION2_ID=$( c8y software versions create --softwareId "$NAME" --file "$package_file" --select "id,c8y_Software.version" --output csv )
echo "$VERSION2_ID" | grep "^[0-9]\+,10.2.3$"

# download
c8y software versions list --softwareId "$NAME" | c8y api | grep "^dummy file$"


# update software
c8y software update --id "$NAME" --description "Example description" --select description --output csv | grep "^Example description$"

# completion (software and version)
c8y __complete software get --id "$NAME" | grep id:
c8y __complete software update --id "$NAME" | grep id:
c8y __complete software delete --id "$NAME" | grep id:
c8y __complete software versions list --softwareId "$NAME" | grep id:
c8y __complete software versions delete --softwareId "$NAME" | grep id:

# list versions by pipeline
c8y software get --id "$NAME" | c8y software versions list --select "id,c8y_Software.version" --output csv | grep "$ID,$VERSION"

# list
c8y software versions list --softwareId "$NAME" --select "id,c8y_Software.version" --output csv | grep "$ID,$VERSION"

# list versions and delete them
c8y software versions list --softwareId "$ID" | c8y software versions delete

# delete parent
c8y software get --id "$NAME" | c8y software delete

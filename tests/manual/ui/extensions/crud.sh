#!/bin/bash
set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

VERSION1_URL="https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.2/tedge-container-plugin-ui.zip"    # version=1.0.1
VERSION2_URL="https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.3/tedge-container-plugin-ui_1.0.2.zip"  # version=1.0.2

createdir () {
    # Cross-platform compatible
    local name="${1:-"c8y-temp"}"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t "$name")
    echo "$tmpdir"
}

TEMP_DIR=$(createdir "c8y-temp")
export TEMP_DIR

cleanup() {
    rm -Rf "$TEMP_DIR"

    if [ -n "$NAME" ]; then
        c8y ui extensions delete --id "$NAME" --silentStatusCodes 404 ||:
    fi
}

trap cleanup EXIT

NAME=${1:-""}

if [[ -z "$NAME" ]]; then
    NAME=$(c8y template execute --template "'ext_' + _.Char(10)")
fi

echo "Using extension name: $NAME"

echo "Creating extension from url"
c8y ui extensions create --file "$VERSION1_URL" --name "$NAME"

echo "Creating extension from file"
EXTENSION_FILE="${TEMP_DIR}/${NAME}.zip" 
wget -O "$EXTENSION_FILE" "$VERSION2_URL"
c8y ui extensions create --file "$EXTENSION_FILE" --tags latest

echo "List extensions"
[ -n "$(c8y ui extensions list)" ]

echo "Get extension"
[ "$(c8y ui extensions get --id "$NAME" --select name -o csv)" = "$NAME" ]

echo "Get extension version by name"
[ "$(c8y ui extensions versions get --extension "$NAME" --version "1.0.1" --select version -o csv)" = "1.0.1" ]

echo "Get extension version by tag"
[ "$(c8y ui extensions versions get --extension "$NAME" --tag "latest" --select version -o csv)" = "1.0.2" ]

echo "List versions"
[ "$(c8y ui extensions versions list --extension "$NAME" | wc -l | xargs)" = "2" ]

echo "Update version tags"
c8y ui extensions versions update --extension "$NAME" --version "1.0.1" --tags latest,v1-info
[ "$(c8y ui extensions versions get --extension "$NAME" --tag v1-info --select version -o csv)" = "1.0.1" ]
[ "$(c8y ui extensions versions get --extension "$NAME" --tag latest --select version -o csv)" = "1.0.1" ]

echo "Delete version by version"
# TODO: Check if --version/--tag is present
c8y ui extensions versions delete --extension "$NAME" --version "1.0.2"
[ -z "$(c8y ui extensions versions get --extension "$NAME" --tag "1.0.2" || true)" ]

# completion (extension and version)
c8y __complete ui extensions get --id "$NAME" | grep id:
c8y __complete ui extensions delete --id "$NAME" | grep id:
c8y __complete ui extensions update --id "$NAME" | grep id:

c8y __complete ui extensions versions get --extension "$NAME" | grep id:
c8y __complete ui extensions versions delete --extension "$NAME" | grep id:
c8y __complete ui extensions versions list --extension "$NAME" | grep id:
c8y __complete ui extensions versions update --extension "$NAME" | grep id:
c8y __complete ui extensions versions create --extension "$NAME" | grep id:

echo "Delete extension"
[ -z "$(c8y ui extensions delete --id "$NAME")" ]

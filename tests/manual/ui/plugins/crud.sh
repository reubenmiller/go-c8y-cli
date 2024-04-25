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
        c8y ui plugins delete --id "$NAME" --silentStatusCodes 404 ||:
    fi
}

trap cleanup EXIT

NAME=${1:-""}

if [[ -z "$NAME" ]]; then
    NAME=$(c8y template execute --template "'ext_' + _.Char(10)")
fi

echo "Using plugins name: $NAME"

echo "Creating plugins from url"
c8y ui plugins create --file "$VERSION1_URL" --name "$NAME"

echo "Creating plugin from file"
PLUGIN_FILE="${TEMP_DIR}/${NAME}.zip" 
wget -O "$PLUGIN_FILE" "$VERSION2_URL"
c8y ui plugins create --file "$PLUGIN_FILE" --tags latest

echo "List plugins"
[ -n "$(c8y ui plugins list)" ]

echo "Get plugin"
[ "$(c8y ui plugins get --id "$NAME" --select name -o csv)" = "$NAME" ]

echo "Get plugin version by name"
[ "$(c8y ui plugins versions get --plugin "$NAME" --version "1.0.1" --select version -o csv)" = "1.0.1" ]

echo "Get plugin version by tag"
[ "$(c8y ui plugins versions get --plugin "$NAME" --tag "latest" --select version -o csv)" = "1.0.2" ]

echo "List versions"
[ "$(c8y ui plugins versions list --plugin "$NAME" | wc -l | xargs)" = "2" ]

echo "Update version tags"
c8y ui plugins versions update --plugin "$NAME" --version "1.0.1" --tags latest,v1-info
[ "$(c8y ui plugins versions get --plugin "$NAME" --tag v1-info --select version -o csv)" = "1.0.1" ]
[ "$(c8y ui plugins versions get --plugin "$NAME" --tag latest --select version -o csv)" = "1.0.1" ]

echo "Delete version by version"
# TODO: Check if --version/--tag is present
c8y ui plugins versions delete --plugin "$NAME" --version "1.0.2"
[ -z "$(c8y ui plugins versions get --plugin "$NAME" --tag "1.0.2" || true)" ]

# completion (plugin and version)
c8y __complete ui plugins get --id "$NAME" | grep id:
c8y __complete ui plugins delete --id "$NAME" | grep id:
c8y __complete ui plugins update --id "$NAME" | grep id:

c8y __complete ui plugins versions get --plugin "$NAME" | grep id:
c8y __complete ui plugins versions delete --plugin "$NAME" | grep id:
c8y __complete ui plugins versions list --plugin "$NAME" | grep id:
c8y __complete ui plugins versions update --plugin "$NAME" | grep id:
c8y __complete ui plugins versions create --plugin "$NAME" | grep id:

echo "Delete plugin"
[ -z "$(c8y ui plugins delete --id "$NAME")" ]

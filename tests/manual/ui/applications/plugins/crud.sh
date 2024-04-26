#!/bin/bash
set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

# Custom plugins (for local plugin testing)
PLUGIN_VERSION1_URL="https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.2/tedge-container-plugin-ui.zip"    # version=1.0.1
PLUGIN_VERSION2_URL="https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.3/tedge-container-plugin-ui_1.0.2.zip"  # version=1.0.2

# Note: The plugin must exist in the management tenant
SHARED_PLUGIN_NAME="Cumulocity community plugins"
SHARED_PLUGIN_CONTEXT_PATH="sag-pkg-community-plugins"

fail() {
    echo "FAIL: $*"
    exit 1
}

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
        c8y applications delete --id "$NAME" --silentStatusCodes 404 ||:
    fi

    if [ -n "$PLUGIN_NAME" ]; then
        c8y ui plugins delete --id "$PLUGIN_NAME" --silentStatusCodes 404 ||:
    fi
}

trap cleanup EXIT

NAME=${1:-""}
PLUGIN_NAME=${2:-""}

if [ -z "$NAME" ]; then
    NAME=$(c8y template execute --template "'app_' + _.Char(10)")
fi

if [ -z "$PLUGIN_NAME" ]; then
    PLUGIN_NAME=$(c8y template execute --template "'plugin_' + _.Char(10)")
fi 

echo "Using application name: $NAME"
c8y applications create --name "$NAME" --type HOSTED --template "{key: $.name + '-key', contextPath: $.name}"

#
# Create a custom dummy with two versions
#
c8y ui plugins create --name "$PLUGIN_NAME" --file "$PLUGIN_VERSION1_URL" --version "1.0.1" --tags latest
c8y ui plugins create --name "$PLUGIN_NAME" --file "$PLUGIN_VERSION2_URL" --version "1.0.2"


#
# List
#
echo "Checking installed plugins (should be empty)"
[ -z "$(c8y ui applications plugins list --application "$NAME")" ] || fail "Plugin list should be empty"

#
# Install
#
echo "Installing plugins to UI application"
c8y ui applications plugins install --application "$NAME" --plugin "$SHARED_PLUGIN_NAME@latest" --plugin "$PLUGIN_NAME"

echo "Installing unknown plugin version should fail"
if c8y ui applications plugins update --application "$NAME" --plugin "$SHARED_PLUGIN_NAME@909.909.99"; then
    fail "Installing unknown plugin versions should fail"
fi

echo "Installing unknown plugin version should fail"
if c8y ui applications plugins update --application "$NAME" --plugin "Unknown Plugin"; then
    fail "Installing unknown plugins should fail"
fi

echo "Checking installed plugins"
[ "$(c8y ui applications plugins list --application "$NAME" --filter "name like '$SHARED_PLUGIN_NAME'" --select name -o csv)" = "$SHARED_PLUGIN_NAME" ] || fail "'$SHARED_PLUGIN_NAME' should be included in the remotes"
[ "$(c8y ui applications plugins list --application "$NAME" --filter "name like '$PLUGIN_NAME'" --select name -o csv)" = "$PLUGIN_NAME" ] || fail "'$PLUGIN_NAME' should be included in the remotes"
[ "$(c8y ui applications plugins list --application "$NAME" | wc -l | xargs)" -eq 2 ] || fail "Expected 2 plugins"


echo "Installing the same plugin should not add duplicate entries"
c8y ui applications plugins install --application "$NAME" --plugin "$SHARED_PLUGIN_NAME@latest"
[ "$(c8y ui applications plugins list --application "$NAME" --filter "name like '$SHARED_PLUGIN_NAME'" | wc -l | xargs)" -eq 1 ] || fail "'$SHARED_PLUGIN_NAME' should only appear once"

#
# Update
#
assert_application_remotes_contains() {
    app_id_name="$1"
    shift
    APP_REMOTES=$(c8y applications get --id "$app_id_name" | jq -r '.config.remotes | keys | .[]')
    while [ "$#" -gt 0 ]; do
        pattern="$1"
        shift
        if ! echo "$APP_REMOTES" | grep -q "$pattern"; then
            fail "Expected $app_id_name application config.remotes to include '$pattern'"
        fi
    done
}

assert_application_remotes_not_contains() {
    app_id_name="$1"
    shift
    APP_REMOTES=$(c8y applications get --id "$app_id_name" | jq -r '.config.remotes | keys | .[]')
    while [ "$#" -gt 0 ]; do
        pattern="$1"
        shift
        if echo "$APP_REMOTES" | grep -q "$pattern"; then
            fail "Expected $app_id_name application config.remotes to include '$pattern'"
        fi
    done
}

echo "Updating specific plugins"
c8y ui applications plugins update --application "$NAME" --plugin "$PLUGIN_NAME"

assert_application_remotes_contains "$NAME" "$SHARED_PLUGIN_CONTEXT_PATH@*" "$PLUGIN_NAME@1.0.1"

echo "Marking $PLUGIN_NAME 1.0.2 as the latest version"
echo "$PLUGIN_NAME" | c8y ui plugins versions update --tags "latest" --version "1.0.2"

echo "Updating all plugins"
c8y ui applications plugins update --application "$NAME" --all
assert_application_remotes_contains "$NAME" "$SHARED_PLUGIN_CONTEXT_PATH@*" "$PLUGIN_NAME@1.0.2"

echo "Updating unknown plugins should fail"
if c8y ui applications plugins update --application "$NAME" --plugin "Invalid Plugin Example"; then
    fail "Updating unknown plugins should fail"
fi

#
# Delete invalid plugins (both orphaned and revoked plugins and plugin versions)
#
echo "Adding an invalid plugin name (which should be deleted later)"
c8y ui applications plugins install --application "$NAME" --template "{
    config:{
        remotes:{
            'manualplugin@1.0.0':['Module1','Module2'],
            '$PLUGIN_NAME@99.99.99':['OtherModule1']
        }
    }
}
"
assert_application_remotes_contains "$NAME" "manualplugin@1.0.0" "$PLUGIN_NAME@99.99.99"

echo "$NAME" | c8y ui applications plugins delete --invalid
assert_application_remotes_not_contains "$NAME" "manualplugin@1.0.0" "$PLUGIN_NAME@99.99.99"

#
# Delete
#
c8y ui applications plugins delete --application "$NAME" --plugin "$SHARED_PLUGIN_NAME" || fail "Failed to delete plugin"
assert_application_remotes_not_contains "$NAME" "$SHARED_PLUGIN_CONTEXT_PATH"

c8y ui applications plugins delete --application "$NAME" --all || fail "Failed to delete all plugins"
[ -z "$(c8y ui applications plugins list --application "$NAME")" ] || fail "Plugin list should be empty"


# completions
c8y __complete ui applications plugins list --application "$NAME" | grep id:
c8y __complete ui applications plugins install --application "$NAME" | grep id:
c8y __complete ui applications plugins replace --application "$NAME" | grep id:
c8y __complete ui applications plugins update --application "$NAME" | grep id:
c8y __complete ui applications plugins delete --application "$NAME" | grep id:

c8y __complete ui applications plugins install --plugin "" | grep id:
# Shared plugins should be included in the list
c8y __complete ui applications plugins install --plugin "" | grep "$SHARED_PLUGIN_NAME"


c8y __complete ui applications plugins replace --plugin "" | grep id:
c8y __complete ui applications plugins install --plugin "" | grep id:
c8y __complete ui applications plugins update --plugin "" | grep id:
c8y __complete ui applications plugins delete --plugin "" | grep id:

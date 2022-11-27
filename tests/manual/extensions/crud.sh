#!/bin/bash

set -ex

export C8Y_SETTINGS_DEFAULTS_DRY=false

createdir () {
    # Cross-platform compatible
    local name="${1:-"c8y-temp"}"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t "$name")
    echo "$tmpdir"
}

C8Y_SETTINGS_EXTENSIONS_DATADIR=$(createdir "datadir")
export C8Y_SETTINGS_EXTENSIONS_DATADIR
TEMP_DIR=$(createdir "extension")

trap "rm -Rf $TEMP_DIR; rm -Rf $C8Y_SETTINGS_EXTENSIONS_DATADIR" EXIT

cd "$TEMP_DIR"
EXTNAME="customext01"

# Create
echo "Using extension name: $EXTNAME"
c8y extensions create "$EXTNAME"

# Install from local repo
c8y extensions install "c8y-$EXTNAME"

# Install from remote repo
c8y extensions install reubenmiller/c8y-defaults

# List
c8y extensions list --select name -o csv --filter "name eq $EXTNAME" | grep -E "^$EXTNAME$"
c8y extensions list --select name -o csv --filter "name eq defaults" | grep -E "^defaults$"

# Use command
OUTPUT=$(c8y customext01 list 2>&1)
echo "$OUTPUT" | grep "Running custom list command"

# Use template
c8y inventory create --template "${EXTNAME}::customCommand.jsonnet" --dry
c8y __complete extensions delete "" | grep -E "^$EXTNAME$"


# Completion
c8y __complete extensions delete "" | grep -E "^$EXTNAME$"
c8y __complete extensions update "" | grep -E "^$EXTNAME$"
c8y __complete inventory create --template "$EXTNAME::" | grep "^$EXTNAME::customCommand.jsonnet$"
c8y __complete inventory list --view "$EXTNAME::" | grep "^$EXTNAME::customDevice"

# Update
c8y extensions update "$EXTNAME" 2>&1 | grep -E "Failed updating extension $EXTNAME: local extensions can not be updated"
c8y extensions update --all

# ls -l "$C8Y_SETTINGS_EXTENSIONS_DATADIR/extensions"

# Delete
c8y extensions delete "$EXTNAME"
c8y extensions list --select name -o csv | grep -v "$EXTNAME"

# TODO: Currently not supported
c8y extensions list | c8y extensions delete
[[ -z $(c8y extensions list --select name -o csv ) ]]

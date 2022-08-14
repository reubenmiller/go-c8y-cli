#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Convert the yaml specs to json
if ! command -v yq 2>&1 >/dev/null; then
    echo "Missing yq, trying to install it now"
    go install github.com/mikefarah/yq/v4@latest
fi

echo "Converting yaml specs to json"
SOURCE_FILES="$SCRIPT_DIR/../api/spec/yaml/*yaml"
DEST_BASE="$SCRIPT_DIR/../api/spec/json"

# Delete any existing specs (as they will be re-generated)
# This ensures there are no orphaned specs if the yaml is deleted and not the json spec
rm -f "$DEST_BASE"/*.json

for filepath in $SOURCE_FILES ; do
    name=$(basename "$filepath")
    dest="$DEST_BASE/${name%.*}.json"
    echo "Converting yaml spec ${filepath}"
    yq -P -o json "$filepath" > "$dest"

    if [[ $? -ne 0 ]]; then
        echo "Could not convert yaml spec to json. $?, file=$filepath"
    fi
done

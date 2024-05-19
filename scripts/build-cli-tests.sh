#!/usr/bin/env bash
set -e
#--------------------------------------------
# Generate the command tests from the specs
#--------------------------------------------

# Delete existing folders to ensure there are no stale test files left behind
rm -Rf ./tests/auto

while IFS= read -r -d '' file
do
    name=$( basename "$file" | sed -E 's/.ya?ml//g' | tr '[:upper:]' '[:lower:]' )
    if ! go run cmd/gen-tests/main.go "./tests/mocks.yaml" "$file" "./tests/auto/$name/tests"; then
        echo "Failed to generate cli tests: $file" >&2
        exit 1
    fi
done <   <( find ./api/spec \( -name "*.yml" -or -name "*.yaml" \) -print0 )

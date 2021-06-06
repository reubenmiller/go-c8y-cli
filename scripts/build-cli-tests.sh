#!/bin/bash
#--------------------------------------------
# Generate the command tests from the specs
#--------------------------------------------

# Delete existing folders to ensure there are no stale test files left behind
rm -Rf ./tests/auto

for file in $( find ./api/spec \( -name "*.yml" -or -name "*.yaml" \) )
do
    name=$( basename "$file" | sed -E 's/.ya?ml//g' | tr A-Z a-z )
    go run cmd/gen-tests/main.go "./tests/mocks.yaml" "$file" "./tests/auto/$name/tests" 
    code=$?
    if [[ $code -ne 0 ]]; then
        echo "Error code: $file, exit_code=$code"
    fi

done

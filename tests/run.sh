#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$SCRIPT_DIR"

folder=$1
shift

export TEST_SHELL=bash
commander test --config ./config.yaml $@ --dir auto/$folder
# commander test --config ./config.yaml $@ --dir auto/tests

# for file in $( find ./ -type d -maxdepth 1 -mindepth 1 \( ! -name "scripts" -and ! -name "dev" \) | sort )
# do
#     name="$file"
#     commander test --config ./config.yaml $@ --dir $name/tests
# done
# commander test --config ./config.yaml $@ --dir ./inventory/tests

popd

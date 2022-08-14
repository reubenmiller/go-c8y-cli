#!/bin/bash

if ! command -v commander; then
    echo "commander is not installed!"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$SCRIPT_DIR"

folder=
if [ $# -gt 0 ]; then
    folder=$1
    shift
fi

export TEST_SHELL=bash
echo "Checking commander help"
commander test --help
commander test --config ./config.manual.yaml $@ --dir manual/$folder
code=$?

popd

exit $code

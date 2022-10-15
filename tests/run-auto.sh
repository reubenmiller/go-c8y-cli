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
commander test --config ./config.yaml "$@" --dir auto/$folder
code=$?

popd

exit $code

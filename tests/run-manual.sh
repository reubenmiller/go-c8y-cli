#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$SCRIPT_DIR"

folder=$1
shift

export TEST_SHELL=bash
commander test --config ./config.manual.yaml $@ --dir manual/$folder
code=$?

popd

exit $code

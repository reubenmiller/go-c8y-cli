#!/bin/bash
####################################################################
# Run bash tests for c8y. A non-zero exit code indicates an error
# Usage:
#   ./test.sh
####################################################################

DIR=$( dirname $( readlink -f "$0" ) )
pushd "$DIR"

echo "Running check: c8y aliases"
./test_aliases.sh || exit 1

echo "Running check: c8y tests"
bats ./
TEST_RESULT=$?

popd
exit $TEST_RESULT

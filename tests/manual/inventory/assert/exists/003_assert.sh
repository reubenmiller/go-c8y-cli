#!/bin/bash
# Regression test: assert exists --not --duration should poll until timeout
# when the managed object exists, not return immediately.
set -eou pipefail

DURATION=5
TOLERANCE=2

mo_id=$( c8y inventory create -f -n --select id --output csv )
cleanup () {
    c8y inventory delete -f --id "$mo_id" > /dev/null 2>&1 || true
}
trap cleanup EXIT

start=$(date +%s)

# The managed object exists, so --not can never be satisfied.
# The command should poll for the full duration before giving up.
c8y inventory assert exists --not --id "$mo_id" --duration "${DURATION}s" --interval 1s --strict || true

elapsed=$(( $(date +%s) - start ))

if [[ $elapsed -lt $(( DURATION - TOLERANCE )) ]]; then
    echo "FAIL: command returned after ${elapsed}s, expected at least $(( DURATION - TOLERANCE ))s" >&2
    exit 1
fi

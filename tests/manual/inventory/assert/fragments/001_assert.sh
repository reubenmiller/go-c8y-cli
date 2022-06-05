#!/bin/bash
set -eou pipefail

mo_id=$( c8y inventory create --name "device01" -n --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo "$mo_id" | c8y inventory assert fragments --fragments "name=example01" --strict

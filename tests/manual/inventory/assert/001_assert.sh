#!/bin/bash
set -eou pipefail

mo_id=$( c8y inventory create -n --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo "$mo_id" | c8y inventory assert --exists | grep "^${mo_id}$"

# Combine with a c8y get Pipe json objects
echo "$mo_id" | c8y inventory get | c8y inventory assert --exists --select id --output csv | grep "^${mo_id}$"

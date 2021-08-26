#!/bin/bash
set -eou pipefail

mo_id=$( c8y inventory create -n --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo -e "$mo_id" | c8y inventory assert exists --strict
echo -e "1\n$mo_id" | c8y inventory assert exists --strict

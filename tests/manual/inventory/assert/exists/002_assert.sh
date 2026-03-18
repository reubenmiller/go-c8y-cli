#!/bin/bash
set -exou pipefail

mo_id=$( c8y inventory create -n --select id --output csv )
cleanup () {
    c8y inventory delete -f -n --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo -e "$mo_id" | c8y inventory assert exists --strict
echo -e "1\n$mo_id" | c8y inventory assert exists --strict --attempts 1 --interval 1s --duration 5s

#!/bin/bash
set -eou pipefail

mo_id=$( c8y inventory create -n --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

nohup c8y inventory update --id $mo_id --data "myTemp=1" --delayBefore 2s >/dev/null 2>&1 &

c8y inventory wait \
    --id $mo_id \
    --fragments "myTemp" \
    --interval 500ms \
    --duration 10s \
    --select id,myTemp \
    --output json

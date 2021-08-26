#!/bin/bash
# set -eou pipefail

mo_id=$( c8y devices create --name "device_with_measurements" --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

# case 1: Strict assertion
c8y measurements assert count --device $mo_id --minimum 1 --strict
if [[ $? -ne 112 ]]; then
    exit 1
fi

c8y measurements assert count --device $mo_id --minimum 1
if [[ $? -ne 0 ]]; then
    exit 2
fi

echo "$mo_id" | c8y measurements create --type c8y_TestMeasurement --time "-0s" --template "{c8y_Signal:{sensor01:{value:1.0,unit:'°C'}}}"

# case 2
echo "$mo_id" | c8y measurements assert count --minimum 1 | grep "^${mo_id}$"

# case 3: Filter by type
echo "$mo_id" | c8y measurements assert count --type c8y_TestMeasurement --minimum 1 | grep "^${mo_id}$"

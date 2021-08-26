#!/bin/bash
# set -eou pipefail

mo_id=$( c8y agents create --name "device_with_operations" --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

# case 1: Strict assertion
c8y operations assert count --device $mo_id --minimum 1 --strict
if [[ $? -ne 112 ]]; then
    exit 1
fi

c8y operations assert count --device $mo_id --minimum 1
if [[ $? -ne 0 ]]; then
    exit 2
fi


echo "$mo_id" | c8y operations create --data "c8y_TestOperation.command='example'"

# case 2
echo "$mo_id" | c8y operations assert count --minimum 1 | grep "^${mo_id}$"

# case 3: Filter by type
echo "$mo_id" | c8y operations assert count --fragmentType c8y_TestOperation --minimum 1 | grep "^${mo_id}$"

echo "device_with_operations" | c8y operations assert count --fragmentType c8y_TestOperation --minimum 1 | grep "^device_with_operations$"

#!/bin/bash
# set -eou pipefail

NAME=$(c8y template execute -n --template "'device_with_operations' + _.Char(10)")
mo_id=$( c8y agents create -n -f --name "$NAME" --select id --output csv )
cleanup () {
    c8y inventory delete -n -f --id $mo_id > /dev/null 2>&1 || true
}
trap cleanup EXIT

# case 1: Strict assertion
c8y operations assert count --device $mo_id --minimum 1 --strict --attempts 2 --interval 1s --duration 5s
if [[ $? -ne 112 ]]; then
    exit 1
fi

c8y operations assert count --device $mo_id --minimum 1 --attempts 2 --interval 1s --duration 5s
if [[ $? -ne 0 ]]; then
    exit 2
fi


echo "$mo_id" | c8y operations create -f --data "c8y_TestOperation.command='example'"

# case 2
echo "$mo_id" | c8y operations assert count --minimum 1 --attempts 2 --interval 1s --duration 5s | grep "^${mo_id}$"

# case 3: Filter by type
echo "$mo_id" | c8y operations assert count --fragmentType c8y_TestOperation --minimum 1  --attempts 2 --interval 1s --duration 5s | grep "^${mo_id}$"

echo "$NAME" | c8y operations assert count --fragmentType c8y_TestOperation --strict --minimum 1 --attempts 2 --interval 1s --duration 5s | grep "^${NAME}$"

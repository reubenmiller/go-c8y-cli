#!/bin/bash
# set -eou pipefail

NAME=$(c8y template execute -n --template "'device_with_events' + _.Char(10)")
mo_id=$( c8y devices create -f --name "$NAME" --select id --output csv )
cleanup () {
    c8y inventory delete -n -f --id "$mo_id" > /dev/null 2>&1 || true
}
trap cleanup EXIT

# case 1: Strict assertion
c8y events assert count --device "$mo_id" --minimum 1 --strict --attempts 2 --interval 1s
if [[ $? -ne 112 ]]; then
    exit 1
fi

c8y events assert count --device "$mo_id" --minimum 1 --attempts 2 --interval 1s
if [[ $? -ne 0 ]]; then
    exit 2
fi

echo "$mo_id" | c8y events create -f --type c8y_TestEvent --time "-0s" --text "Test alarm"

# case 2
echo "$mo_id" | c8y events assert count --minimum 1 --attempts 2 --interval 1s | grep "^${mo_id}$"

# case 3: Filter by type
echo "$mo_id" | c8y events assert count --type c8y_TestEvent --minimum 1 --attempts 2 --interval 1s | grep "^${mo_id}$"

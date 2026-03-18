#!/bin/bash
# set -eou pipefail

NAME=$(c8y template execute -n --template "'device_with_alarms' + _.Char(10)")
mo_id=$( c8y devices create -n -f --name "$NAME" --select id --output csv )
cleanup () {
    c8y inventory delete -n -f --id "$mo_id" > /dev/null 2>&1 || true
}
trap cleanup EXIT

# case 1: Strict assertion
c8y alarms assert count --device "$mo_id" --minimum 1 --strict --attempts 1 --interval 1s
if [[ $? -ne 112 ]]; then
    exit 1
fi

c8y alarms assert count --device "$mo_id" --minimum 1 --attempts 1 --interval 1s
if [[ $? -ne 0 ]]; then
    exit 2
fi

echo "$mo_id" | c8y alarms create -f --type c8y_TestAlarm --time "-0s" --text "Test alarm" --severity MAJOR

# case 2
echo "$mo_id" | c8y alarms assert count --minimum 1 --attempts 1 --interval 1s | grep "^${mo_id}$"

# case 3: Filter by type
echo "$mo_id" | c8y alarms assert count --type c8y_TestAlarm --minimum 1 --attempts 1 --interval 1s | grep "^${mo_id}$"

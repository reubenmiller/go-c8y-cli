#!/bin/bash
set -eou pipefail

mo_id=$( c8y inventory create -n --select id --output csv )
cleanup () {
    c8y inventory delete --id $mo_id --cascade > /dev/null 2>&1 || true
}
trap cleanup EXIT

child1=$( c8y inventory create -n --select id --output csv )
child2=$( c8y inventory create -n --select id --output csv )

c8y inventory children assign --id "$mo_id" --child "$child1" --childType addition > /dev/null
c8y inventory children assign --id "$child1" --child "$child2" --childType addition > /dev/null

[[ $(echo "$child1" | c8y inventory parents get --type addition --select id --output csv) == "$mo_id" ]] || exit 2
[[ $(echo "$child2" | c8y inventory parents get --type addition --select id --output csv) == "$child1" ]] || exit 3
[[ $(echo "$child2" | c8y inventory parents get --type addition --level -1 --select id --output csv) == "$mo_id" ]] || exit 4
[[ $(echo "$child2" | c8y inventory parents get --type addition --level 0 --select id --output csv) == "$child2" ]] || exit 5

# all parents
output=$(echo "$child2" | c8y inventory parents get --type addition --all --select id --output csv)
[[ $(echo "$output" | head -1) == "$child1" ]] || exit 6
[[ $(echo "$output" | tail -1) == "$mo_id" ]] || exit 7

# all parents (reverse)
output=$(echo "$child2" | c8y inventory parents get --type addition --all --reverse --select id --output csv)
[[ $(echo "$output" | head -1) == "$mo_id" ]] || exit 8
[[ $(echo "$output" | tail -1) == "$child1" ]] || exit 9

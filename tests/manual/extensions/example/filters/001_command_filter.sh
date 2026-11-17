#!/bin/bash
# Check that the client side filter of an extension API command is applied to the response,
# that it is combined with the user's --filter values, and that other commands are not affected
set -eou pipefail

export C8Y_SETTINGS_DEFAULTS_DRY=false
export C8Y_SETTINGS_DEFAULTS_CACHE=false

TYPE=$(c8y template execute -n --template "'ci_commandFilter_' + _.Char(10)")

ids=()
cleanup () {
    for id in "${ids[@]}"; do
        c8y inventory delete -n -f --id "$id" > /dev/null 2>&1 || true
    done
}
trap cleanup EXIT

ids+=("$(c8y inventory create -n -f --type "$TYPE" --name modelA --select id -o csv)")
ids+=("$(c8y inventory create -n -f --type "$TYPE" --name templateA --template "{c8y_Template: {}}" --select id -o csv)")
ids+=("$(c8y inventory create -n -f --type "$TYPE" --name modelB --select id -o csv)")

assert_names () {
    local name="$1"
    local expected="$2"
    local actual="$3"
    if [[ "$actual" != "$expected" ]]; then
        echo "FAIL: $name. expected=[$expected], got=[$actual]" >&2
        exit 1
    fi
}

# case 1: the command's filter removes the template
actual=$(c8y kitchensink filters list -n --type "$TYPE" --select name -o csv | sort | paste -sd, -)
assert_names "command filter" "modelA,modelB" "$actual"

# case 2: the user's filter is applied in addition to the command's filter
actual=$(c8y kitchensink filters list -n --type "$TYPE" --filter "name like *B" --select name -o csv | sort | paste -sd, -)
assert_names "command and user filter" "modelB" "$actual"

# case 3: the user can't include the filtered items
actual=$(c8y kitchensink filters list -n --type "$TYPE" --filter "name like template*" --select name -o csv | sort | paste -sd, -)
assert_names "user filter can't undo the command filter" "" "$actual"

# case 4: other commands which call the same endpoint are not affected
actual=$(c8y inventory list -n --type "$TYPE" --select name -o csv | sort | paste -sd, -)
assert_names "other commands are not filtered" "modelA,modelB,templateA" "$actual"

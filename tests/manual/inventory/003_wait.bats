#!/usr/bin/env bats
# set -eou pipefail
load ./load

setup () {
    export mo_id=$( c8y inventory create --data "myCustom.value=two" --select id --output csv )
}
teardown () {
    c8y inventory delete --id $mo_id > /dev/null 2>&1 || true
}

@test "wait for a fragment to be removed" {
    nohup c8y inventory update --id $mo_id --template "{myCustom: null}" --delayBefore 2s >/dev/null 2>&1 &

    # output=$(
        
    # )
    run c8y inventory wait \
            --id $mo_id \
            --fragments '!myCustom.value' \
            --interval 500ms \
            --duration 5s \
            --select "myCustom.**" \
            --output json
    
    assert_output "{}"
    # [ "${output}" == "{}" ]
}

@test "wait for a fragment to be added" {
    nohup c8y inventory update --id $mo_id --template "{myCustom: {value2: 2}}" --delayBefore 2s >/dev/null 2>&1 &

    output=$(
        c8y inventory wait \
            --id $mo_id \
            --fragments 'myCustom.value2=2' \
            --interval 500ms \
            --duration 5s \
            --select "myCustom.**" \
            --output json
    )
    
    assert_output "{\"myCustom\": 2}"
    # [ "${output}" == "{\"myCustom\": 2}" ]
}

#!/bin/bash

#
# Helpers
#
# Control typing speed
export DEMO_TYPE_SPEED_FACTOR=10000

#
# Show a banner with an information box about the demo or
# a specific step
#
showbanner () {
    echo "$@" | boxes -d unicornsay | lolcat -d 1
    sleep 0.800
}

showtitle () {
    echo "$@" | boxes | lolcat -d 1
    sleep 0.800
}

#
# Simulate typing and then run the command
#
runCommand () {
    echo -n "go-c8y-cli % "
    echo -e "$@" | randtype -m 1 -n "%\t" -t 10,$DEMO_TYPE_SPEED_FACTOR
    sleep 0.250
    cmd="$( echo -e "$@" | tr -d '\t\n\\' )"
    eval "$cmd"
}

runCommandWithAlternative () {
    echo -n "go-c8y-cli % "
    local fakeCommand="$1"
    shift
    echo -e "$fakeCommand" | randtype -m 1 -n "%\t" -t 10,$DEMO_TYPE_SPEED_FACTOR
    sleep 0.250
    # Add in --view off before the pipe to jq, as the views are not automatically turned
    # off when using eval, bug in go-c8y-cli?
    # cmd="$( echo -e "$@" | tr -d '\t\n\\' | sed 's/| jq/--view off | jq/g' )"
    cmd="$( echo -e "$@" | tr -d '\t\n\\' )"
    eval "$cmd"
}

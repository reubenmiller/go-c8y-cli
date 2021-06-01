#!/bin/bash

dir=$( dirname $(readlink -f "$0") )
source "$dir/helpers.sh"

demo () {
    local lb=" \\ \\\n \\\t "
    local nlb="\\\n\\\t "

    showtitle "Using Views to focus on your data"

    showbanner "Views are used when outputing to console"
    runCommand c8y applications list
    sleep 7

    showbanner "They are also applied to other output formats (json,csv,csvheader)"
    runCommand c8y applications list --pageSize 2 --output json
    sleep 7

    clear

    showbanner "You can manually specify a view rather than using auto detection"
    runCommand c8y applications list -p 1 --view ignoreinbuilt --output json
    sleep 7

    showbanner "Or just turn them off completely"
    runCommand c8y applications list -p 1 --view off --output json -c
    sleep 7

    clear

    showbanner "Views are automatically turned off when assigning, piping or redirecting"
    runCommandWithAlternative "c8y applications list -p 1 | jq" "c8y applications list -p 1 --view off | jq"
    sleep 7

    showbanner "Views are defined as json files, and can be customized"
    runCommand "cat \"\$HOME/.go-c8y-cli/views/default/devices.json\" | jq"
    sleep 7

    exit
}

demo

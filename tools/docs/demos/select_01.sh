#!/bin/bash

dir=$( dirname $(readlink -f "$0") )
source "$dir/helpers.sh"

demo () {
    local lb=" \\ \\\n \\\t "
    local nlb="\\\n\\\t "

    showtitle "Using global --select parameter"

    showbanner "Only include fields that you want"
    runCommand c8y applications list --select \'id,name,type\'
    sleep 5

    showbanner "And in the data format that you want "
    runCommand c8y applications list --select \'id,name,type\' --output csv
    sleep 5

    clear

    showbanner "Use * to match against unknown fragment names"
    runCommand c8y applications list --select \'id,name,avail*\'
    sleep 5

    showbanner "Get all non-nested fragments"
    runCommand c8y applications list --select \'*\' -p 1 -o json
    sleep 5

    showbanner "Get all fragments"
    runCommand c8y applications list --select \'**\' -p 1 -o json
    sleep 5

    clear

    showbanner "Use dotnotation to reference nested fragments"
    runCommand c8y applications list --select \'id,name,owner.tenant.id\'
    sleep 3

    showbanner "Use globstar '**' if you don't know where the fragment is..."
    runCommand c8y applications list --select \'id,name,**.tenant.id\'
    sleep 5

    showbanner "Use '!' to exclude fragments"
    runCommand c8y applications list --select \'!id,name,**.tenant.id\'
    sleep 5

    showbanner "Use <alias>:<fragment> to rename fragments"
    runCommand c8y applications get --id administration -o json --select \'appName:name,appOwner:owner.**\'
    sleep 5

    exit
}

demo

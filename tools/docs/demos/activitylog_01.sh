#!/bin/bash

dir=$( dirname $(readlink -f "$0") )
source "$dir/helpers.sh"

demo () {
    local lb=" \\ \\\n \\\t "
    local nlb="\\\n\\\t "
    # runCommand $( echo -e "c8y activitylog list $lb--dateFrom -5min $lb--filter \"method eq POST\" --filter \"path like *managedObjects*\" $lb--select responseTimeMS -o csv | $nlb datamash -H min 1 max 1 mean 1 | column -t" )
    # exit
    showbanner "Create some example agents"
    runCommand "c8y util repeat 5 --format \"agent_%s%03s\" | c8y agents create --force --delay 250ms"
    sleep 3

    clear

    showbanner "Check the activity log for a record of what was done"
    runCommand "c8y activitylog list --dateFrom -5min"
    sleep 5

    showbanner "Check the response times and status codes"
    runCommand "c8y activitylog list --dateFrom -5min --select \"method,path,*responsetime*,statusCode\""
    sleep 5

    clear
    showbanner "Calculate performance stats for POST requests by using other tools (like datamash or miller)"
    runCommand $( echo -e "c8y activitylog list $lb--dateFrom -5min $lb--filter \"method eq POST\" --filter \"path like *managedObjects*\" $lb--select responseTimeMS -o csv | $nlb datamash -H min 1 max 1 mean 1 | column -t" )
    sleep 5
    exit
}

demo

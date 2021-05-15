#!/bin/bash

shortdemo1 () {
    local run_once="$1"
    local running=true
    DEMO_STEP=0
    local speed_factor=10000

    showbanner () {
        echo "$@" | boxes -d unicornsay | lolcat -d 1
        sleep 0.5
    }

    runCommand () {
        echo -n "go-c8y-cli % "
        echo "$@" | randtype -m 1 -n "%" -t 10,$speed_factor
        sleep 0.250
        eval "$@"
    }


    while $running
    do
        clear
        case $DEMO_STEP in
        2)
            figlet "go-c8y-cli" | lolcat -d 1
            showbanner "Create/update/delete stuff easily"
            runCommand "seq -f 'demo_%03g' 1 100 | c8y devices create --force --workers 5 --delay 100ms --progress"
            ;;

        3)
            showbanner "Supports piping data"
            runCommand "c8y devices list -p 10 | c8y devices update --template \"{name: 'live' + input.value.name}\""
            ;;
        
        4)
            showbanner "Only show the data that you want (and in either table, json or csv)"
            runCommand "c8y devices list --query \"name eq '*live*'\" --orderBy 'creationTime.date' --select 'id,name' --output csvheader"
            ;;

        5)
            showbanner "Your requests are locally logged for traceability"
            runCommand "c8y activitylog list --dateFrom -5min --filter \"method eq PUT\" --select path,method,responseTimeMS"
            ;;
        
        99)
            showbanner "And completion for almost everything :)"
            # runCommand "c8y activitylog list --dateFrom -10min --filter \"responseSelf like *managedObjects*\" --filter \"method eq POST\" | c8y api --method DELETE"
            ;;
        
        6)
            running=false
            DEMO_STEP=0
            sleep 1
            break
            ;;

        *)
            clear
            DEMO_STEP=1
            # figlet "go-c8y-cli" | lolcat -d 1
            # showbanner "Hi, let's show off what c8y can do"
            ;;

        esac

        DEMO_STEP=$(( $DEMO_STEP + 1 ))

        if [[ -n "$run_once" ]]; then
            break
        fi
        sleep 1
    done

    exit
}

shortdemo1

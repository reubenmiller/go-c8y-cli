#!/bin/bash
DEVICE=$1
END=${2:-60}

for i in $(seq 1 $END); do
    c8y alarms create --device "$DEVICE" --template ./test.alarm.jsonnet -n --force
    sleep 2
done

#!/bin/bash
DEVICE=$1
END=${2:-60}

for i in $(seq 1 $END); do
    c8y operations create --device "$DEVICE" --template '{description: "Test operation", c8y_Restart: {}}' -n --force
    # c8y operations create --device "$DEVICE" --template test.operation.jsonnet -n --force
    sleep 2
done

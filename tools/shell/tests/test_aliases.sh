#!/bin/bash

shopt -s expand_aliases
source ../c8y.plugin.sh

resp=$(alarms || exit 1)
resp=$(apps || exit 2)
resp=$(devices || exit 3)
resp=$(events || exit 4)
resp=$(fmo "name eq 'example*'" || exit 5)
resp=$(measurements || exit 6)
resp=$(ops || exit 7)
# resp=$(series || exit 8)

#resp=$(alarm 1 || exit 9)
#resp=$(app 1 || exit 10)
#resp=$(event 1 || exit 11)
#resp=$(m 1 || exit 12)
#resp=$(mo 1 || exit 13)
#resp=$(op 1 || exit 14)

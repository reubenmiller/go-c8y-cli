#!/bin/bash

mo1_id="$( c8y inventory create --select id --output csv )"

cleanup () {
    echo "$mo1_id" | c8y inventory delete --silentStatusCodes 404 --silentExit 2>/dev/null
}

trap cleanup EXIT

echo -e "0\n$mo1_id" | c8y inventory delete --silentStatusCodes 404
[[ "$?" -ne 4 ]] && exit 1

echo -e "0\n0" | c8y inventory delete --silentStatusCodes 404 --silentExit
[[ "$?" -ne 0 ]] && exit 2

echo -e "0\n0" | c8y inventory delete
[[ "$?" -ne 104 ]] && exit 3

exit 0

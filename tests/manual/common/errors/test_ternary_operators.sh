#!/bin/bash

set -e

name="$( c8y template execute --template "{name: _.AlphaNumeric(16)}" --select name -o csv )"

# check if device exists by name if not create it
mo1=$( c8y devices get --id "$name" --silentStatusCodes 404 || c8y devices create --name "$name" )

cleanup () {
    echo "$mo1" | c8y devices delete
}

trap cleanup EXIT

echo "$mo1" | c8y util show --select name -o csv | c8y assert text --exact "$name"
mo1_id=$( echo "$mo1" | c8y util show --select id -o csv )

# run the command again, this time the name should exist so it should not be created
mo2=$( c8y devices get --id "$name" --silentStatusCodes 404 || c8y devices create --name "$name" )

echo "$mo2" | c8y util show --select id -o csv | c8y assert text --exact "$mo1_id"

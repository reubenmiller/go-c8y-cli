#!/bin/bash

set -e

echo "path: $(which c8y)"

ID1=$( c8y inventory create --name cached_device01 --cacheBodyPaths name,id,type --cache --cacheTTL 10min --template "{now: _.Now()}" --select id --output csv -f )
ID2=$( c8y inventory create --name cached_device01 --cacheBodyPaths name,id,type --cache --cacheTTL 10min --template "{now: _.Now()}" --select id --output csv -f )

cleanup () {
    echo "$ID1" | c8y inventory delete --silentStatusCodes 404 -f || true
    echo "$ID2" | c8y inventory delete --silentStatusCodes 404 -f || true
}
trap cleanup EXIT

[[ "$ID1" == "$ID2" ]] || exit 10

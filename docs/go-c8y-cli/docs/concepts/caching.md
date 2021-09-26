---
title: Caching
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Demo

<Video
  videoSrcURL="https://asciinema.org/a/416566/embed?speed=1.0&autoplay=false&size=small&rows=30"
  videoTitle="Views example"
  width="90%"
  height="550px"
  ></Video>

## Overview

Caching provides an easy way to reduce the number of requests sent to the server by caching the responses locally on disk.

The responses from large requests can be cached so that when the same request is sent again (within the cache's time-to-live (TTL) setting), instead of sending the request to the server, the response is retrieved from file (assuming the cache has not expired). This improves the performance of repetitive queries by reusing already existing responses across commands.

The caching works be intercepting every HTTP request and writing the HTTP response to file using a unique hash based on the URL, path, authorization header and body (if set).

The commands can be cached via control via global parameters `cache`, `noCache` and `cacheTTL`

```sh
c8y devices list --type "c8y_Linux" --cache
# => Sends a request to the server the first time

c8y devices list --type "c8y_Linux" --cache
# => A cached response is used as the same command is being used within the cache TTL

echo agent01 | c8y util repeat 2 | c8y devices get --cache --cacheTTL 60s -v
# => Control time-to-live for single commands 
```

## Cache expiration

The cache's validity is controlled via a time-to-live (TTL) setting where if the cached file last modified time is older than the TTL setting, then the cached file is ignored and the request is sent to the server.


### Renewing cache

The cached files can be renewed (i.e. updating the last modified time to current timestamp).

```sh
c8y cache renew
```

Alternatively you can always just increase the time-to-live (TTL) cache to a value which should cover all of the cached items.

```sh
c8y devices list --cache --cacheTTL 30d
```

## Deleting cache

The cache will grow from time to time as it is not actively deleted by go-c8y-cli during normal operation

To delete the cache, run the following command

```sh
c8y cache delete
```

If you only want to cache for files which are only older than a time period, then the `age` parameter can be used: 

```sh
c8y cache delete --age 1h
```

### Performance improvements

The performance increase of using caching can be easily demonstrated with a simple example.

Let's assume that you want to p

```sh title="example.sh"
#!/bin/bash
echo agent01 | c8y util repeat 2 | c8y devices get --cache --cacheTTL 60s
```

## Usage

### Enable/disable for individual commands

The cache is controlled via the global `cache/noCache` parameter. The global parameter can also be set via the configuration.

```sh
# Enable permanent caching by updating session settings
c8y settings update defaults.cache true
c8y settings update defaults.cacheTTL 10m

c8y alarms list
# => Sends a response the first time, then caches the response

c8y alarms list
# => Cached response is used
```

Caching can be enabled or disabled for individual commands

```sh
c8y devices list --cache
# => Enable caching

c8y devices list --noCache
# => Disable caching
```

## Using in scripts

Caching can be helpful in scripts to reduce the amount of code to write an efficient script, as you don't need to save the whole command response to variables, you can just reuse the same command with the `--cache` parameter and let go-c8y-cli look after it for you.

### Retrieving list of devices



```sh title="./operation-summary.sh"
#!/bin/bash

export C8Y_SETTINGS_CI=true
export C8Y_SETTINGS_DEFAULTS_CACHE=true
export C8Y_SETTINGS_DEFAULTS_CACHETTL=30m

#
# Perform some task (i.e. gather statistics)
#
input_file="$1"

while read name; do

    # Note. c8y devices get will use cache so no need to assign it to a variable.
    echo -n "Operation Summary: Device - "
    echo "$name" | c8y devices get | c8y template execute --template "{label: input.value.name + '(' + input.value.id + ')'}" --select label --output csv
    
    # Get stats about completed operations
    echo -n "Total SUCCESSFUL: "
    echo "$name" | c8y devices get | c8y operations list --status SUCCESSFUL --pageSize 1 --withTotalPages --select statistics.totalPages -o csv

    echo -n "Total FAILED: "
    echo "$name" | c8y devices get | c8y operations list --status FAILED --pageSize 1 --withTotalPages --select statistics.totalPages -o csv

    echo ""
done <"$input_file"

# cleanup (to save space)
c8y cache delete --age 2h
```



```sh title="usage"
echo -e "agent01\nagent01\nagent01" > devices.list

./operation-summary.sh devices.list
```

```sh title="output"
Operation Summary: Device - agent01(761834)
Total SUCCESSFUL: 0
Total FAILED: 0

Operation Summary: Device - agent01(761834)
Total SUCCESSFUL: 0
Total FAILED: 0

Operation Summary: Device - agent01(761834)
Total SUCCESSFUL: 0
Total FAILED: 0

✓ Deleted cache /tmp/go-c8y-cli-cache
```

Since caching is enabled, the requests on the server are reduced because the same GET request is not sent twice, instead it just retrieves the response from cache. The handling of the response does not change so it can fit very nicely into existing scripts. 

### Using cache to prevent creating duplicate items

Caching can also be applied to all HTTP Methods, not just GET requests. This enables the ability to write idempotent scripts to prevent accidentally creating duplicates.

Let's say you want to copy measurements from one device to another, but since it could take a long time (depending on how many measurements there are), you want to be able to stop a script and restart it without creating duplicates.

Normally, this would require the tracking of measurements by keeping track of the source measurement id and before creating it on the target device, you would have to check if the id already exists in the locally managed list.

However because of in-built caching support, the script becomes very simple. The following shows how simple the script really is:

```sh title="./copy_measurements.sh"
#!/bin/bash

#
# Input arguments to set source and target device
source_device=$1
target_device=$2

#
# Enable caching by default for all commands
export C8Y_SETTINGS_DEFAULTS_CACHE="true"

# Enable caching of POST commands as well as GET
export C8Y_SETTINGS_CACHE_METHODS="GET POST"

# Use a really large TTL, as we want to prevent creating it for a long time
export C8Y_SETTINGS_DEFAULTS_CACHETTL="100d"

#
# Copy the measurements from source => target device
# If the measurement was already created then the cached response will be returned
c8y measurements list \
    --device "$source_device" \
    --includeAll \
    --view ignoreinbuilt_id \
| c8y measurements create \
    --device "$target_device" \
    --template input.value \
```

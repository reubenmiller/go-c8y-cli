---
title: Caching
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

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


## Configuration

Caching is disabled by default, and only set for `GET` https methods. However these defaults can be changed to suite your

* Enabled/Disabled
* Time-to-Live (TTL)
* Cache key generation (include auth/host etc.)


### defaults.cache: boolean

Enable caching. Defaults to `false`.

Caching can be turned on for individual commands by using the `--cache` global parameter.

When the setting is activated, then all requests which match the `cache.methods` settings will be cached. Users can opt out of caching for individual commands by using the `--noCache` global parameter.

### defaults.cacheTTL: string

Cache time-to-live (TTL) settings to control the maximum age that a cached response. If the cached response is older than the TTL settings, then the cached item will be ignored.

### cache.keyauth: boolean

Include the request's Authorization header value when generating the cache key. Defaults to `true`.

This can be useful if you want to share caching across multiple users. Normally the authorization header is used in the cache key generation, so the cached item will be a different file key for multiple users

### cache.keyhost: boolean

Include the host name when generating the cache key. Defaults to `true`.

This can be useful when setting up a mock system where you want to fake server response by using cached response regardless of the server. This is a more advanced settings, so just ignore it if you don't understand it.

### cache.methods: string

Space separated list of HTTP methods which should be cached. By default only `GET` is configured.

<CodeExample>

```bash
# Enable more http methods to be cached
c8y settings update cache.methods "GET PUT POST"
```

</CodeExample>

### cache.path: string

Location of the cache directory. If the path does not exist it was be created when the first cached item is created.

Defaults to:

* Linux/MacOS: `{tmp}/go-c8y-cli-cache`, where `{tmp}` is set to `$TMPDIR` if not empty otherwise `/tmp`
* Windows: `{tmp}/go-c8y-cli-cache`, where `{tmp}` is set to the first non-empty value of `%TMP%`, `%TEMP%`, `%USERPROFILE%`

:::note
If go-c8y-cli is being by multiple users on the same computer/server and you need to prevent users from accessing the local cached files, then you should change the default location of the cache, and apply the appropriate folder permissions to prevent other users from accessing the response from other users.
:::

## Cache expiration

The cache's validity is controlled via a time-to-live (TTL) setting where if the cached file last modified time is older than the TTL setting, then the cached file is ignored and the request is sent to the server.


### Renewing cache

The cached files can be renewed (i.e. updating the last modified time to current timestamp).

```sh
c8y cache renew
```

Alternatively you can always just increase the time-to-live (TTL) cache to a value which should cover all of the cached items.

```sh
c8y devices list --cache --cacheTTL 720h
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

## Checking if a response is cached or not

The cac

```sh
c8y devices list --cache -v
c8y devices list --cache -v
```

```sh title="output"
√ go-c8y-cli % c8y devices list -p 1 --cache -v
2021-09-27T05:50:24.723Z        INFO    Binding authorization environment variables
2021-09-27T05:50:24.724Z        INFO    activityLog path: /home/vscode/.cumulocity/activitylog/c8y.activitylog.2021-09-27.json
2021-09-27T05:50:24.727Z        INFO    Loaded session: /workspaces/go-c8y-cli/.cumulocity/example.json
2021-09-27T05:50:24.728Z        INFO    command: c8y devices list -p 1 --cache -v 
2021-09-27T05:50:24.733Z        INFO    Max jobs: 0
2021-09-27T05:50:24.734Z        INFO    worker 1: started job 1
2021-09-27T05:50:24.734Z        INFO    Current username: {tenant}/{username}
2021-09-27T05:50:24.735Z        INFO    Headers: map[Accept:[application/json] Authorization:[Basic  {base64 tenant/username:password}] User-Agent:[go-client] X-Application:[go-client]]
2021-09-27T05:50:24.735Z        INFO    Sending request: GET https://{host}/inventory/managedObjects?pageSize=1&q=$filter=+$orderby=name
2021-09-27T05:50:24.735Z        INFO    Using cached response. file: /tmp/go-c8y-cli-cache/5a/f6/f2d746aba15b47d3f2dfc517c782f030bb28cfc1acca6fc3698ed532948b, age: 2.2051073s, ttl: 1m0s
2021-09-27T05:50:24.735Z        INFO    Status code: 200
2021-09-27T05:50:24.735Z        INFO    Response time: 0ms
2021-09-27T05:50:24.735Z        INFO    Response Content-Type: application/vnd.com.nsn.cumulocity.managedobjectcollection+json;charset=UTF-8;ver=0.9
2021-09-27T05:50:24.735Z        INFO    Response Length: 1.3KB
2021-09-27T05:50:24.735Z        INFO    Unfiltered array size. len=1
2021-09-27T05:50:24.736Z        INFO    View mode: auto
2021-09-27T05:50:24.737Z        INFO    Detected view: id, name, type, owner, lastUpdated, c8y_Availability.status
| id          | name            | type       | owner         | lastUpdated                   | c8y_availability.status |
|-------------|-----------------|------------|---------------|-------------------------------|-------------------------|
| 764806      | 74xsyz1veu      |            | ciuser01      | 2021-09-24T12:41:33.385Z      |                         |
2021-09-27T05:50:24.738Z        INFO    worker 1: finished job 1 in 4ms
```

An additional line is included when a cached response is being used which shows the cache file being used along with the age and the time-to-live (TTL).

```sh
Using cached response. file: /tmp/go-c8y-cli-cache/5a/f6/f2d746aba15b47d3f2dfc517c782f030bb28cfc1acca6fc3698ed532948b, age: 2.2051073s, ttl: 1m0s
```

In addition, the activity log now has additional fields, `.cached` and `.etag` to show if the response was cached and the hashed key of the request. The `.responseTimeMS` is also a key indicator that a cached response is being used, as the response time is generally under 5 milliseconds when a cached response is being used as the response is being read from local disk instead of being sent to the server.

```sh
c8y activity log --dateFrom -5m | jq
```

```sh title="output"
{
  "accept": "application/json",
  "cached": false,
  "ctx": "qPgGLFHs",
  "etag": "",
  "host": "main.example.c8y.io",
  "method": "GET",
  "path": "/inventory/managedObjects",
  "processingMode": "",
  "query": "pageSize=1&q=$filter= $orderby=name",
  "responseSelf": "https://t12345.example.c8y.io/inventory/managedObjects?q=$filter%3D%20$orderby%3Dname&pageSize=1&currentPage=1",
  "responseTimeMS": 194,
  "statusCode": 200,
  "time": "2021-09-27T05:50:22.5373115Z",
  "type": "request"
}
{
  "accept": "application/json",
  "cached": true,
  "ctx": "kTtcTBBb",
  "etag": "5af6f2d746aba15b47d3f2dfc517c782f030bb28cfc1acca6fc3698ed532948b",
  "host": "main.example.c8y.io",
  "method": "GET",
  "path": "/inventory/managedObjects",
  "processingMode": "",
  "query": "pageSize=1&q=$filter= $orderby=name",
  "responseSelf": "https://t12345.example.c8y.io/inventory/managedObjects?q=$filter%3D%20$orderby%3Dname&pageSize=1&currentPage=1",
  "responseTimeMS": 0,
  "statusCode": 200,
  "time": "2021-09-27T05:50:24.735405Z",
  "type": "request"
}
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

Or caching can be enabled or disabled for individual commands

```sh
c8y devices list --cache
# => Enable caching

c8y devices list --noCache
# => Disable caching
```

## Using in scripts

Caching can be helpful in scripts to reduce the amount of code to write an efficient script, as you don't need to save the whole command response to variables, you can just reuse the same command with the `--cache` parameter and let go-c8y-cli look after it for you.

### Getting complete operations overview for multiple devices

Below is a small script to get the an overview of the completed operations for a list of devices. It excepts an input file to contain a list of device ids or names, and prints out the total number of SUCCESSFUL and FAILED operations for the device.

By using cache, the `c8y devices get` command can be repeated without worrying about spamming the server with duplicated requests.

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

Below shows the script being used, and some example output:

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

---
title: Common Parameters
---

import CodeExample from '@site/src/components/CodeExample';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


## Overview

`c8y` supports common parameters to handle a range of different scenarios.

For example:

* only include selected properties in the response
* force/confirm a command (disable/enable prompts)
* verbose/debug output
* control pagination
* change output format

Common parameters are also supported in the `PSc8y` PowerShell module, however the parameters have a slightly different format. Below shows a few examples of how they map to PSc8y. 

|c8y|PSc8y|
|---|-----|
|`--verbose`|`-Verbose`|
|`--pageSize`|`-PageSize`|

:::info
* In PowerShell use a single dash `-` before parameter instead of two dashes `--`.
* `c8y` parameters are **case-sensitive**, however PowerShell is not
:::

Below is an example showing the differences between `c8y` and `PSc8y` 

<CodeExample>

```bash
c8y devices list --verbose --pageSize 1 --withTotalPages
```

</CodeExample>

All common parameters are detailed in the next section along with some examples to show its use case.

## Common Parameters

---

### abortOnErrors

Abort batch when reaching specified number of errors (default 10)

A command is set to exit early if the number of errors reaches a certain value. For example if the user is trying to a large number of alarms for a non-existent device, then the command will stop after 10 errors, giving the user a change to fix their mistake without spamming the platform.

Also if the server is not available due to an ongoing upgrade, then the command will stop early as their is no point sending requests to a platform which is currently unreachable

#### Example: Abort after receiving 2 errors

Create 10 events for the same device. Though for the sake of the example, we will use a non-existent device id (i.e. id=0), which will throw a HTTP 422 error.

The non-existent id will be piped 10 times using the utility command `c8y util repeat` which will be piped to try to create 10 events, where a delay of 1 second is used between creating events.

<CodeExample transform="false">

```bash
c8y util repeat 10 --format 0 |
  c8y events create --type "io_simulation_Raw" --text "Received raw input value" --delay 1000 --abortOnErrors 2
```

```powershell
c8y util repeat 10 --format 0 |
  c8y events create --type "io_simulation_Raw" --text "Received raw input value" --delay 1000 --abortOnErrors 2
```

</CodeExample>

```json title="output"
2021-04-24T11:11:13.638Z        ERROR   commandError: aborted batch as error count has been exceeded. totalErrors=2
```

The exact error can be viewed in the activity log which is accessible either by viewing the json file or using the command:

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom "-10min" --filter "method eq POST" --select "time,method,path,statusCode,responseError"
```

```powershell
c8y activitylog list --dateFrom "-10min" --filter "method eq POST" --select "time,method,path,statusCode,responseError"
```

</CodeExample>

```text title="output"
| time                              | method     | path               | statusCode | responseError.error             | responseError.info                                                                    | responseError.message                           |
|-----------------------------------|------------|--------------------|------------|---------------------------------|---------------------------------------------------------------------------------------|-------------------------------------------------|
| 2021-04-24T11:11:12.6073087Z      | POST       | /event/events      | 422        | event/Unprocessable Entity      | https://cumulocity.com/guides/reference/rest-implementation/#a-name-error-reporting-… | Source object does not exist in inventory.      |
| 2021-04-24T11:11:13.6367223Z      | POST       | /event/events      | 422        | event/Unprocessable Entity      | https://cumulocity.com/guides/reference/rest-implementation/#a-name-error-reporting-… | Source object does not exist in inventory.      |
```

---

### compact

Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal.

<CodeExample>

```bash
c8y devices list --output json --compact --view auto
```

```powershell
Get-DeviceCollection -Output json -Compact -View auto
```

</CodeExample>

````json  title="output"
{"id":"494210","lastUpdated":"2021-04-24T08:44:28.080Z","name":"device_0001","owner":"ciuser01","type":"ci_Test"}
{"id":"480957","lastUpdated":"2021-04-24T08:43:25.058Z","name":"device_0002","owner":"ciuser01","type":"ci_Test"}
{"id":"481037","lastUpdated":"2021-04-24T08:43:25.095Z","name":"device_0003","owner":"ciuser01","type":"ci_Test"}
{"id":"480861","lastUpdated":"2021-04-24T08:43:25.122Z","name":"device_0004","owner":"ciuser01","type":"ci_Test"}
{"id":"480862","lastUpdated":"2021-04-24T08:43:25.148Z","name":"device_0005","owner":"ciuser01","type":"ci_Test"}
````

---

### confirm

Prompt for confirmation

By default only specific commands require a confirmation, i.e. by default commands which just send a GET HTTP request do not require confirmation, however confirmation can be enforced by using `confirm`.

<CodeExample>

```bash
c8y devices list --confirm
```


```powershell
Get-DeviceCollection -Confirm
```

</CodeExample>

````bash  title="output"
? Confirm (job: 1)
Get device collection on tenant {tenant}
[y] Yes  [a] Yes to All  [n] No  [l] No to All (y)
````

:::info
Confirmation will take priority over the force parameter. Useful in production environments when you want to be sure that you will be prompted before the command is sent
:::

---

### confirmText

Custom confirmation text

Control the confirmation message that will be shown to the user. This can be used when creating your own wrapper scripts.

<CodeExample>

```bash
c8y devices list --confirm --confirmText "Hey there, are you sure you want to get a device list"
```

</CodeExample>

````bash  title="output"
? Confirm (job: 1)
Hey there, are you sure you want to get a device list on tenant {tenant}
[y] Yes  [a] Yes to All  [n] No  [l] No to All (y)
````

---

### currentPage

Current page which should be returned

:::tip
Using large page number can be slow, so it is better to use either date filtering or the inventory query where possible

* `c8y devices list --query "name eq 'device*' and lastUpdated.date gt '2021-04-24T09:18:00'"`
* `c8y operations list --dateFrom "2021-04-24T09:18:00"`
:::

**Example: Get the second page of results instead of the first**

<CodeExample>

```bash
c8y devices list --pageSize 5 --currentPage 2
```

</CodeExample>

````json  title="output"
| id          | name             | type       | owner         | lastUpdated                   | c8y_availability.status |
|-------------|------------------|------------|---------------|-------------------------------|-------------------------|
| 461902      | device_0006      |            | ciuser01      | 2021-04-24T12:19:12.803Z      |                         |
| 461903      | device_0007      |            | ciuser01      | 2021-04-24T12:19:12.882Z      |                         |
| 484203      | device_0008      |            | ciuser01      | 2021-04-24T12:19:12.919Z      |                         |
| 453083      | device_0009      |            | ciuser01      | 2021-04-24T12:19:12.951Z      |                         |
| 453084      | device_0010      |            | ciuser01      | 2021-04-24T12:19:12.995Z      |                         |
````

---

### debug

Set very verbose log messages

Debug will show both verbose and debug messages on standard error output. The debug output can be helpful to diagnose unexpected any behavior. But generally it should not be turned on when running many commands otherwise your output will be spammed with messages.

#### Example: Get debug output of a command so it can be included in a bug ticket

Sensitive information should be hidden before sharing it with anyone as it contains the authorization header

<CodeExample transform="false">

```bash
# hide sensitive info
eval $( c8y settings update logger.hideSensitive true --shell auto )

c8y alarms list --debug
```

```powershell
# hide sensitive info
c8y settings update logger.hideSensitive true --shell auto | Out-String | Invoke-Expression

Get-AlarmCollection -Debug
```

</CodeExample>

---

### delay

delay in milliseconds after each request

Adding a delay after each request is useful to apply some basic rate limiting to your commands (especially when using pipelines).


:::info
* Delay is ignored if only one request is being sent. i.e. `c8y devices list --delay 1000` will not apply the 1 second delay
* Delay is applied even for a dry run. This is helpful to check that the delay works as you expect
:::

#### Example: Rate limiting

Using a `delay` of 500 milliseconds to ensure that the requests do not overload the platform.

<CodeExample>

```bash
seq 1 10 | c8y devices create --workers 2 --delay 500
```

```powershell
1..10 | batch | New-Device -Workers 2 -Delay 500
```

</CodeExample>

---

### delayBefore

delay in milliseconds before each request is sent

#### Example 1: Simulate a device (agent) processing operations 

A realtime client is used to subscribe to created operations for a device and it forwards the operation down the pipeline where downstream commands process before passing it on to its downstream command. The following shows the sequence followed by each operation:

1. Wait for newly created operations (using realtime client with websockets)
2. Wait 5000 milliseconds (after receiving an operation from step 1), then set the operation's status to EXECUTING
3. Wait 10000 milliseconds (after receiving operation from step 2), then set the operation's status to SUCCESSFUL

The whole process runs for 180 seconds as the subscription duration is set to `180` in the first command. Alternatively, if you only want to process a fixed count of operations then you can use the `count` parameter, but the duration will still be used to control how long the realtime client listens to operations for.

<CodeExample>

```bash
c8y operations subscribe --device 1234 --duration 180 --actionTypes CREATE |
  c8y operations update --delayBefore 5000 --status EXECUTING |
  c8y operations update --delayBefore 10000 --status SUCCESSFUL
```

</CodeExample>

---

### dry

Dry run. Don't send any data to the server

`dry` displays the REST request that would have been sent if `dry` was not used. This parameter does NOT send a request to Cumulocity, it just displays the contents of the request on the console. This is useful if you want to see exact REST request and body being sent so you can copy it for use in other applications.

**Example: Check what request would be sent without creating the device**

<CodeExample>

```bash
c8y devices create --name "my-test" --dry --dryFormat markdown
```

</CodeExample>

````markdown  title="output"
What If: Sending [POST] request to [https://example.com/inventory/managedObjects]

### POST /inventory/managedObjects

| header            | value
|-------------------|---------------------------
| Accept            | application/json
| Authorization     | Basic  {base64 tenant/username:password}
| Content-Type      | application/json


#### Body

```json
{
  "c8y_IsDevice": {},
  "name": "my-test"
}
```
````

---

### dryFormat

Dry run output format. i.e. json, dump, markdown or curl (default "markdown")

See the [dry run](dryrun) concept page for more details and examples

---

### filter

Apply a client side filter to response before returning it to the user. It supports a simple filter language making it easy for wildcards, regular expression matches

Multiple filters can be applied and they will be applied together with a logical AND.

:::caution
You want to ensure that the `pageSize` is large enough that it will include some the response that you are interested in, otherwise the filter will not display any results.
:::

**Example: Get application list and filter using a wildcard search by name and type**

<CodeExample>

```bash
c8y applications list --pageSize 100 --filter "name like *cockpit*" --filter "type match HOSTED|MICROSERVICE"
```

</CodeExample>

````text  title="output"
| id         | name         | key                          | type        | availability |
|------------|--------------|------------------------------|-------------|--------------|
| 8          | cockpit      | cockpit-application-key      | HOSTED      | MARKET       |
````

See the [filtering](filtering) concept page for more details and examples

---

### flatten

flatten json output by replacing nested json properties with properties where their names are represented by dot notation.

Useful when you want to make it easier to access highly nested data.

:::info
Only used used when the output is set to json
:::

**Example: Get the first device in a list and flatten the json result**

The `select` parameter is used here along with the globstar `**` to select all properties of the response. Alternatively `--view off` could be used which would have the same effect.

<CodeExample>

```bash
c8y devices list --pageSize 1 --output json --flatten --select "**"
```

</CodeExample>

````json  title="output"
{
  "additionParents.references": [],
  "additionParents.self": "https://{tenant}.example.com/inventory/managedObjects/494210/additionParents",
  "assetParents.references": [],
  "assetParents.self": "https://{tenant}.example.com/inventory/managedObjects/494210/assetParents",
  "c8y_IsDevice": {},
  "childAdditions.references": [],
  "childAdditions.self": "https://{tenant}.example.com/inventory/managedObjects/494210/childAdditions",
  "childAssets.references": [],
  "childAssets.self": "https://{tenant}.example.com/inventory/managedObjects/494210/childAssets",
  "childDevices.references": [],
  "childDevices.self": "https://{tenant}.example.com/inventory/managedObjects/494210/childDevices",
  "creationTime": "2021-04-18T18:10:12.940Z",
  "deviceParents.references": [],
  "deviceParents.self": "https://{tenant}.example.com/inventory/managedObjects/494210/deviceParents",
  "id": "494210",
  "lastUpdated": "2021-04-24T08:44:28.080Z",
  "name": "device_0001",
  "owner": "ciuser01",
  "self": "https://{tenant}.example.com/inventory/managedObjects/494210",
  "testme": {},
  "type": "ci_Test"
}
````

---

### force

Do not prompt for confirmation. The force can be used in scripts or if you are sure you know what you are doing.

:::tip
It is recommended to avoid using `force` where possible, as the confirmation text is there to ensure the user is interacting with the correct tenant before doing a potentially destructive command.

It is recommended to limit usage to within automated scripts.
:::

:::info
Force will be ignore if the `confirm` parameter is also provided
:::

---

### header

custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"

Add custom headers to the output going HTTP request. Useful when a custom microservice utilizes headers in custom api calls, or there is a new Cumulocity header which is not yet supported by go-c8y-cli.

**Example: Add custom header when requesting a list of measurements**

In this example a custom accept header is used `text/csv` when requesting a list of measurements, the header instructs Cumulocity to return the results in a csv format instead of the default json format. `c8y measurements list` already has parameters which can control this (`csvFormat` and `excelFormat`) however for the sake of the exercise, we will add header manually. 

<CodeExample>

```bash
c8y measurements list --device 494210 --header "Accept: text/csv"

# equivalent to (for c8y measurements list only)
c8y measurements list --device 494210 --csvFormat
```

```powershell
Get-MeasurementCollection -Device 494210 -Header "Accept: text/csv" -AsPSObject:$false

# equivalent to (for c8y measurements list only)
Get-MeasurementCollection -Device 494210 -CsvFormat
```

</CodeExample>

````csv  title="output"
time,source,device_name,fragment.series,value,unit
2021-04-24T13:19:04.144Z,494210,device_0001,io.sensor01,55,total
2021-04-24T13:19:04.145Z,494210,device_0001,io.sensor01,12,total
2021-04-24T13:19:04.275Z,494210,device_0001,io.sensor01,75,total
2021-04-24T13:19:05.318Z,494210,device_0001,io.sensor01,87,total
2021-04-24T13:19:06.352Z,494210,device_0001,io.sensor01,9,total
2021-04-24T13:19:07.376Z,494210,device_0001,io.sensor01,82,total
2021-04-24T13:19:08.401Z,494210,device_0001,io.sensor01,17,total
2021-04-24T13:19:09.424Z,494210,device_0001,io.sensor01,17,total
2021-04-24T13:19:10.449Z,494210,device_0001,io.sensor01,76,total
2021-04-24T13:19:11.477Z,494210,device_0001,io.sensor01,2,total
````

---

### includeAll

Include all results by iterating through each page

Each page is written to standard output as the response is returned, so and downstream commands (via the pipeline) will not have to wait until all of the pages have been retrieved.

:::info
Commands which use the inventory query language (i.e. `inventory/managedObjects?q=` or `inventory/managedObjects?query=`) will be automatically optimized to use the technique of reformating the query to skip pagination via modifying the query search space by adding `_id gt <last_result_id>` to the query. If the tenant has a large about of devices (>20K), this technique drastically decreases the time it takes to retrieve all results instead of iterating over the pages via the `currentPage` query parameter. Some of the commands which use the technic are: `c8y device list` and `c8y inventory find --query "name eq '*'"`.

This optimization happens without user intervention.
:::

**Example: Rename all of the devices by adding a common prefix to the existing device name**

Renaming all devices (regardless of how many) can be done by chaining two command together via the pipeline.

The example will follow the following steps:

1. Get a list of devices but use a query to exclude devices with a specific fragment (which will be added when updating the name)

2. Update the device name using a template which does two things

  a. Add the "ci-" prefix to the existing device name (using the automatic variable `input.value`)

  b. Add a custom fragment to mark the device as being completed so if the command is run again then it will not be included in the device list in step 1.

A delay of 200 milliseconds is used between each request to rate limit the requests to reduce load on the platform.

<CodeExample>

```bash
c8y devices list --query "not(has(goc8ycli_Rename))" --includeAll |
  c8y devices update --delay 200 --template "{ name: 'ci-' + input.value.name, goc8ycli_Rename: { lastUpdated: _.Now('0s') } }"
```

</CodeExample>

````text  title="output"
| id          | name                   | type       | owner         | lastUpdated                   | c8y_availability.status |
|-------------|------------------------|------------|---------------|-------------------------------|-------------------------|
| 451797      | ci-testdevice_002      |            | ciuser01      | 2021-04-24T13:46:38.021Z      |                         |
| 451798      | ci-testdevice_003      |            | ciuser01      | 2021-04-24T13:46:38.060Z      |                         |
| 451799      | ci-testdevice_006      |            | ciuser01      | 2021-04-24T13:46:38.097Z      |                         |
| 451957      | ci-testdevice_005      |            | ciuser01      | 2021-04-24T13:46:38.137Z      |                         |
| 451958      | ci-testdevice_008      |            | ciuser01      | 2021-04-24T13:46:38.168Z      |                         |
| 451961      | ci-testdevice_014      |            | ciuser01      | 2021-04-24T13:46:38.199Z      |                         |
| 451962      | ci-testdevice_013      |            | ciuser01      | 2021-04-24T13:46:38.227Z      |                         |
| 451963      | ci-testdevice_018      |            | ciuser01      | 2021-04-24T13:46:38.258Z      |                         |
| 451964      | ci-testdevice_019      |            | ciuser01      | 2021-04-24T13:46:38.286Z      |                         |
| 452010      | ci-testdevice_001      |            | ciuser01      | 2021-04-24T13:46:38.311Z      |                         |
| 452011      | ci-testdevice_004      |            | ciuser01      | 2021-04-24T13:46:38.340Z      |                         |
| 452012      | ci-testdevice_007      |            | ciuser01      | 2021-04-24T13:46:38.373Z      |                         |
| 452013      | ci-testdevice_010      |            | ciuser01      | 2021-04-24T13:46:38.401Z      |                         |
````

---

### logMessage

Add custom message to the activity log

It is useful if you want to add a description to the command that you are running, i.e. a task description, so that your future self can more easily remember why you ran the command.

The message will be under the `.message` property of the entry type `command`. The HTTP request related to a command are linked via the `.ctx` (context) property.

**Example: Add custom message to the activity log entry related to executing a command**

<CodeExample>

```bash
c8y devices update --id 494210 --data "ci_data.usage=100" --logMessage "setting data usage to 100%"
```

</CodeExample>

```json  title="file: activitylog entry"
{"time":"2021-04-24T14:33:08.9024232Z","ctx":"CFWmWgDk","type":"command","arguments":["devices","update","--id","494210","--data","ci_data.usage=100","--logMessage","setting data usage to 100%"],"message":"setting data usage to 100%"}
{"time":"2021-04-24T14:33:09.0949105Z","ctx":"CFWmWgDk","type":"request","method":"PUT","host":"example.com","path":"/inventory/managedObjects/494210","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":149,"responseSelf":"https://{tenant}.example.com/inventory/managedObjects/494210"}
```

---

### maxJobs

Maximum number of jobs. 0 = unlimited (use with caution!)

One job represents one outgoing HTTP request (not including an lookup by name requests). If the job count exceeds the given `maxJobs` value, then the command will exit with a non-zero exit code.

The hard limit can be used to protect against unexpected inputs which are feed from the pipeline. i.e. if you are piping in a list of device ids and to update a fragment on each device, and you want to add a protection against accidentally using the the wrong file which contains a much large list of devices. Setting the `maxJobs` to the expected number of jobs is a safe guard in case if the input contains more than what you expect. 

:::tip
Use `maxJobs` in production environments to protect yourself against modifying more devices than you expect when using pipelines
:::

**Example: Prevent unexpected requests when updating number of devices from a file**

You have an input file, `my-device-list.txt`, containing device ids (one per line) and you want to set the required availability interval for these. You expect the input file to only contain 5 devices. To ensure that only 5 devices will be processed at time of running the command, you set the maxJobs limit.

```json  title="file: my-device-list.txt"
1111
2222
3333
4444
5555
6666
7777
```

<CodeExample>

```bash
cat my-device-list.txt | c8y devices setRequiredAvailability --interval 30 --maxJobs 5
```

</CodeExample>

```json title="output"
| id          | name                    | type       | owner         | lastUpdated                   | c8y_Availability.status |
|-------------|-------------------------|------------|---------------|-------------------------------|-------------------------|
| 497913      | mobile-device_0001      |            | ciuser01      | 2021-04-24T19:14:44.705Z      | AVAILABLE               |
| 497718      | mobile-device_0002      |            | ciuser01      | 2021-04-24T19:14:44.731Z      | AVAILABLE               |
| 497821      | mobile-device_0003      |            | ciuser01      | 2021-04-24T19:14:44.757Z      | AVAILABLE               |
| 497914      | mobile-device_0004      |            | ciuser01      | 2021-04-24T19:14:44.784Z      | AVAILABLE               |
| 497719      | mobile-device_0005      |            | ciuser01      | 2021-04-24T19:14:44.808Z      | AVAILABLE               |
2021-04-24T19:14:44.622Z        ERROR   commandError: max job limit exceeded. limit=5
```

---

### noAccept

Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect

**Example: Update a device without returning the object (more performance)**

Updating a large about of devices can put a load on the platform. In order to minimize the load, the `noAccept` parameter can be used as less processing is required by the platform as it does not need to send the update managed object back.

<CodeExample>

```bash
c8y devices list --name "mobile*" | c8y devices update --data "myOtherFragment={}" --noAccept
```

```powershell
Get-DeviceCollection -Name "mobile*" | batch | Update-Device -Data "myOtherFragment={}" -NoAccept
```

</CodeExample>

```bash title="output"
✓ Updated /inventory/managedObjects/497913 => 200 OK
✓ Updated /inventory/managedObjects/497718 => 200 OK
✓ Updated /inventory/managedObjects/497821 => 200 OK
✓ Updated /inventory/managedObjects/497914 => 200 OK
✓ Updated /inventory/managedObjects/497719 => 200 OK
```

---

### noColor

Don't use colors when displaying log entries on the console

Color is automatically disabled when the output is not written to the console output, i.e. when no pipeline, redirection or variable assignment is being used. But color can be enforced by using `--noColor=false`.

---

### noLog

Disables the activity log for the current command

A log in the activity log can be disabled for single command if you don't want it writing the command to file (maybe to increase performance by skipping the write-to-file call done by `go-c8y-cli`)

---

### noProxy

Ignore the proxy settings

`NoProxy` ignores all of proxy settings. Helpful when trying to diagnose proxy problems, or you need to ignore the `HTTPS_PROXY` environment variables.

**Example: Ignore any proxy environment variables**

The following command will not use the proxy settings when using the `noProxy` parameter.

<CodeExample>

```bash
export HTTPS_PROXY="http://10.0.0.1:8000"
c8y devices list --noProxy
```

```powershell
$env:HTTPS_PROXY = "http://10.0.0.1:8000"
Get-DeviceCollection -NoProxy
```

</CodeExample>


If you are in doubt about which proxy settings are being used when running a command then add `verbose` parameter to turn on the verbose logging, and the proxy/noproxy settings will be displayed on the console.

---

### nullInput

Disable reading from standard input (stdin). Useful if you are already redirecting to stdin inside a bash for/while loop.

---

### output

Output format i.e. table, json, csv, csvheader (default "table" (on terminal) otherwise "json")

:::info
The `table` format is the default when the output is being displayed in the console output, i.e. when no pipeline, redirection or variable assignment is being used. Otherwise `json` format is the default (when the server response is JSON of course, otherwise the raw text will be printed).
:::

**Example: Get fields from a devices and export as csv**

Print output as CSV (without headers) so that the data is easier to stream and parse to other 3rd party tools. `select` should also be used together in order to limit which fields are included in the output.

<CodeExample>

```bash
# csv without header
c8y devices list --name "mobile*" --select "id,name,type,creation*" --output csv

# csv with header
c8y devices list --name "mobile*" --select "id,name,type,creation*" --output csvheader
```

</CodeExample>

```csv title="output: csv without header"
497913,mobile-device_0001,,2021-04-24T19:14:01.202Z
497718,mobile-device_0002,,2021-04-24T19:14:01.254Z
497821,mobile-device_0003,,2021-04-24T19:14:01.301Z
497914,mobile-device_0004,,2021-04-24T19:14:01.334Z
497719,mobile-device_0005,,2021-04-24T19:14:01.364Z
```

```csv title="output: csv with header"
id,name,type,creationTime
497913,mobile-device_0001,,2021-04-24T19:14:01.202Z
497718,mobile-device_0002,,2021-04-24T19:14:01.254Z
497821,mobile-device_0003,,2021-04-24T19:14:01.301Z
497914,mobile-device_0004,,2021-04-24T19:14:01.334Z
497719,mobile-device_0005,,2021-04-24T19:14:01.364Z
```

If you want to include the csv headers in the output then 

---

### outputFile

Save JSON output to file (after select)

---

### outputFileRaw

Save raw response to file


The raw response (as returned by Cumulocity) can also be written to file in addition to displaying it on the console.

No view logic or select statements on the response will be applied. This can be useful if you want to tee the output

<CodeExample>

```bash
c8y alarms list -p 1--outputFileRaw test.json

# Or if you don't want any console output, just use redirection, but be sure to use raw!
c8y alarms list -p 1 --raw > test.json
```

</CodeExample>

```powershell title="output"
| id         | status            | type                         | severity      | count      | source.id  |
|------------|-------------------|------------------------------|---------------|------------|------------|
| 302        | ACKNOWLEDGED      | c8y_CepEsperDeprecation      | CRITICAL      | 133        | 115        |
```

```javascript title="test.json"
{"self":"https://{tenant}.example.com/alarm/alarms?pageSize=1&currentPage=1","statistics":{"totalPages":1,"currentPage":1,"pageSize":1},"alarms":[{"severity":"CRITICAL","creationTime":"2020-12-10T03:00:23.198Z","count":133,"history":{"auditRecords":[],"self":"https://{tenant}.example.com/audit/auditRecords"},"source":{"name":"CEP Engine {tenant}","self":"https://{tenant}.example.com/inventory/managedObjects/115","id":"115"},"type":"c8y_CepEsperDeprecation","firstOccurrenceTime":"2020-12-10T03:00:23.108Z","self":"https://{tenant}.example.com/alarm/alarms/302","time":"2021-04-24T03:00:10.405Z","id":"302","text":"Streaming analytics using CEL (Esper) is deprecated and no longer supported. Please refer to our documentation to migrate your real-time rules to Apama (https://cumulocity.com/guides/apama/overview-analytics/#migrate-from-esper)","status":"ACKNOWLEDGED"}]}
```

Or it can be used to download a binary from the platform

c8y_applications_storage_9397

<CodeExample>

```bash
c8y inventory find --query "has(c8y_IsBinary) and type eq 'c8y_applications_storage_*'" |
  c8y binaries get > my.zip
```

</CodeExample>

---

### pageSize

Maximum results per page (default 5)

Set the maximum number of results for commands which return a list of objects. There is usually an upper limit enforced by Cumulocity which is dependant on the Cumulocity object type (i.e. Events, Operations, Managed Objects etc.). By default it is set to `5`.


**Example: Get a collection of alarms and limit the results to 100 alarms**

<CodeExample>

```bash
c8y alarms list --pageSize 100
```

</CodeExample>

---

### progress

Show progress bar. This will also disable any other verbose output

---

### proxy

Proxy setting, i.e. http://10.0.0.1:8080

---

### queryParam

custom query parameters. i.e. --queryParam "withCustomOption=true,myOtherOption=myvalue"

---

### raw

Show raw response. This mode will force output=json and view=off


The `Raw` parameter tells the cli tool to return the Cumulocity response without an pre-processing, so that the response including the extra meta properties like `next`, `self` and `statistics`. This is useful if you want to process the original response from Cumulocity yourself, rather than only returning the array items.

This concept is applied to all entities (i.e. alarms, events, managed objects and operations etc.)

:::caution
`output` and `select` parameters will be ignored when using `raw`
:::

**Example: Get a list of alarms**

When getting an alarm list without the `raw` parameter, only the array containing the alarms is returned (i.e. `.alarms`). All of the other properties like `statistics` are removed. This is normally the preferred behavior because you are normally only interested in the actual alarms and not any additional meta properties. This reduces the code required when piping the alarm list to another function.

So getting a list of alarms without `raw` looks like this:

<CodeExample>

```bash
c8y alarms list -p 1 | jq

# or without jq
c8y alarms list -p 1 --output json --view off
```

```powershell
Get-AlarmCollection -PageSize 1 | tojson

# or without Powershell
Get-AlarmCollection -PageSize 1 -Output json -View off
```

</CodeExample>

```json
{
  "severity": "MAJOR",
  "creationTime": "2020-01-26T14:21:07.325Z",
  "count": 1,
  "history": {
    "auditRecords": [],
    "self": "https://{tenant}.eu-latest.cumulocity.com/audit/auditRecords"
  },
  "source": {
    "name": "TestDevicegIvTYAuwPz",
    "self": "https://{tenant}.eu-latest.cumulocity.com/inventory/managedObjects/131166",
    "id": "131166"
  },
  "type": "testType",
  "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms/130877",
  "time": "2020-01-26T14:21:06.921Z",
  "text": "Custom Event 1",
  "id": "130877",
  "status": "ACTIVE"
}
```

:::info
The `view` off is used to ensure all of the json properties are shown to us.
:::

Here is the same command but using the `raw` parameter.

<CodeExample>

```bash
c8y alarms list -p 1 --raw
```

</CodeExample>

```json title="output"
{
  "next": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms?pageSize=1&currentPage=2",
  "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms?pageSize=1&currentPage=1",
  "statistics": {
    "totalPages": 70,
    "currentPage": 1,
    "pageSize": 1
  },
  "alarms": [
    {
      "severity": "MAJOR",
      "creationTime": "2020-01-26T14:21:07.325Z",
      "count": 1,
      "history": "@{auditRecords=System.Object[]; self=https://{tenant}.eu-latest.cumulocity.com/audit/auditRecords}",
      "source": "@{name=TestDevicegIvTYAuwPz; self=https://{tenant}.eu-latest.cumulocity.com/inventory/managedObjects/131166; id=131166}",
      "type": "testType",
      "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms/130877",
      "time": "2020-01-26T14:21:06.921Z",
      "text": "Custom Event 1",
      "id": "130877",
      "status": "ACTIVE"
    }
  ]
}
```

---

### select

Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select "id,name,type,**.serialNumber"

---

### session

Session configuration

Session enables the use another c8y session for a single command without having to run `set-session`.

This is useful if you want to quickly see what is going on in another tenant without permanently changing your current session, or you want to transfer some information from a staging environment to a dev as a once-off action.

**Example: Get a device name from the staging tenant, and create a new device with the same name in a dev tenant**

<CodeExample>

```bash
devicename=$( c8y operations list --pageSize 1 --session staging-tenant -o csv --select name )

c8y devices create --name "$devicename" --session dev-tenant
```

```powershell
$Device = Get-OperationCollection -PageSize 1 -Session staging-tenant

New-Device -Name $Device.name -Session dev-tenant
```

</CodeExample>

---

### sessionPassword

Override session password

---

### sessionUsername

Override session username. i.e. peter or t1234/peter (with tenant)

---

### silentStatusCodes

Status codes which will not print out an error message

---

### timeout

Request timeout in seconds (default 600)

---

### totalPages

Total number of pages to get

---

### verbose

Display verbose information when sending a command to Cumulocity. The verbose messages help to understand what was sent to Cumulocity and what was received as well with other verbose messages related to the cli tool itself.

**Example: Display detailed logging about the command**

<CodeExample>

```bash
c8y devices list --verbose
```

</CodeExample>

```powershell title="output"
2021-04-24T08:44:34.559Z        INFO    Binding authorization environment variables
2021-04-24T08:44:34.564Z        INFO    activityLog path: /home/vscode/.go-c8y-cli/activityLog/c8y.activitylog.2021-04-24.json
2021-04-24T08:44:34.564Z        INFO    Loaded session: /workspaces/go-c8y-cli/.cumulocity/test.runner.ciuser.json
2021-04-24T08:44:34.564Z        INFO    command: c8y devices list --verbose
2021-04-24T08:44:34.592Z        INFO    Max jobs: 0
2021-04-24T08:44:34.595Z        INFO    worker 1: started job 1
2021-04-24T08:44:34.595Z        INFO    Current username: {tenant}/{username}
2021-04-24T08:44:34.595Z        INFO    Headers: map[Accept:[application/json] Authorization:[Basic  {base64 tenant/username:password}] User-Agent:[go-client] X-Application:[go-client]]
2021-04-24T08:44:34.595Z        INFO    Sending request: GET https://{host}/inventory/managedObjects?q=$filter=+$orderby=name
2021-04-24T08:44:34.698Z        INFO    Status code: 200
2021-04-24T08:44:34.698Z        INFO    Response time: 102ms
2021-04-24T08:44:34.698Z        INFO    Response Content-Type: application/vnd.com.nsn.cumulocity.managedobjectcollection+json;charset=UTF-8;ver=0.9
2021-04-24T08:44:34.698Z        INFO    Response Length: 5.6KB
2021-04-24T08:44:34.698Z        INFO    Unfiltered array size. len=5
2021-04-24T08:44:34.699Z        INFO    View mode: auto
2021-04-24T08:44:34.700Z        INFO    Detected view: id, name, type, owner, lastUpdated, c8y_Availability.status
| id          | name             | type         | owner         | lastUpdated                   | c8y_availability.status |
|-------------|------------------|--------------|---------------|-------------------------------|-------------------------|
| 494210      | device_0001      | ci_Test      | ciuser01      | 2021-04-24T08:44:28.080Z      |                         |
| 480957      | device_0002      | ci_Test      | ciuser01      | 2021-04-24T08:43:25.058Z      |                         |
| 481037      | device_0003      | ci_Test      | ciuser01      | 2021-04-24T08:43:25.095Z      |                         |
| 480861      | device_0004      | ci_Test      | ciuser01      | 2021-04-24T08:43:25.122Z      |                         |
| 480862      | device_0005      | ci_Test      | ciuser01      | 2021-04-24T08:43:25.148Z      |                         |
2021-04-24T08:44:34.702Z        INFO    worker 1: finished job 1 in 107ms
```

:::info
Verbose output is written to standard error, and the output response is written to standard output. The standard output can be redirected to null to only show the verbose messages using:

<CodeExample>

```bash
c8y devices list --verbose > /dev/null
```

```powershell
Get-DeviceCollection -Verbose | Out-Null
```

</CodeExample>

:::

---

### view

View option (default "auto")

Use views when displaying data on the terminal. Disable using `--view off`

---

### withError

Errors will be printed on stdout instead of stderr

---

### withTotalPages

`withTotalPages` requests that the response should contain the `.statistics.totalPages` property to see how many total pages exist.

In order to get an accurate total of entities, it is recommended to use `pageSize 1` along with `withTotalPages`. The `totalPages` property in the response will then display the total number of entities.

<CodeExample>

```bash
c8y alarms list --pageSize 1 --withTotalPages
```

</CodeExample>

```powershell title="output"
| totalPages | pageSize   | currentPage |
|------------|------------|-------------|
| 1          | 1          | 1           |
```

:::note
The above show the output when views have been applied, so it is only showing your the important information (i.e. only the `.statistics` fragment)
:::

---

### workers

Number of workers (default 1)

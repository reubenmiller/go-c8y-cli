---
layout: default
category: Concepts
title: Output Templates
---

import CodeExample from '@site/src/components/CodeExample';

:::caution
First introduced in version `2.32.0`.
:::

## Overview

Output templates provide a way to shape the output of a command and to combine with additional information that can be used in down stream commands. The output template is applied from a global flag `outputTemplate` and uses the same jsonnet template engine as the `template` flag.

For information about the template language (jsonnet) please refer to the [templates concept](./templates.md) page.

## Examples

The full power of output templates can be shown by a series of examples.

### Example: Getting number of alarms per device

A common scenario is to generate a report on a per-device level which checks how many alarms existing matching a specific criteria, for example how many alarms have been acknowledged in the last 7 days.

Before demonstrating output templates, it is useful to see why it is needed. If you are familiar with [chaining commands](./chaining-commands.md), then the following command shouldn't need much explaining, however for those who aren't, the following command simply gets a list of devices, and then sends one API call per device getting the count of alarms which are in the `ACKNOWLEDGED` state and have been created within the last 7 days.

<CodeExample>

```bash
c8y devices list \
| c8y alarms count \
    --status ACKNOWLEDGED \
    --dateFrom -7d
```

</CodeExample>

The output from the command above gives the following output.

```sh title="Output"
5
10
0
2
3
```

The output shown is the response from the Cumulocity API call being invoked by the `c8y alarms count` command. It does not include any information about which device the count is for making output less useful for reporting or for usage in downstream commands.

This is where an output template really shines. An output template allows you to combine the output of the `c8y alarms count` command with data that was piped into it (e.g. the device). The same command as above can be tweaked by adding the `outputTemplate` flag, and providing a template which combines the data that will make our report more useful.

<CodeExample>

```bash
c8y devices list \
| c8y alarms count \
    --status ACKNOWLEDGED \
    --dateFrom -7d \
    --outputTemplate "{deviceId: input.value.id, deviceName: input.value.name, totalAlarms: output}"
```

</CodeExample>

With the usage of the `outputTemplate` flag, the output is much more useful as it shows the device id and name and the total acknowledged alarms.

```sh title="Output"
| deviceId      | deviceName      | totalAlarms |
|---------------|-----------------|-------------|
| 211732        | linux001        | 5           |
| 337989        | linux002        | 10          |
| 213536        | linux003        | 0           |
| 219268        | linux004        | 2           |
| 219268        | linux005        | 3           |
```

### Example: Getting number of alarms per device

This is similar to the previous alarm count, however it differs slightly because it uses the paging statistics to get the total count instead of the count endpoint.

<CodeExample>

```bash
c8y devices list \
| c8y events list \
    --withTotalPages \
    --pageSize 1 \
    --outputTemplate "input.value + output.statistics" \
    --select id,name,totalPages
```

</CodeExample>


```sh title="Output"
| id          | name            | totalPages |
|-------------|-----------------|------------|
| 129203      | linux001        | 3          |
| 105203      | linux002        | 1          |
| 130313      | linux003        | 3          |
| 130323      | linux004        | 8          |
| 130311      | linux005        | 6          |
```

Now for an explanation about what is going on with the command as there are a few moving parts.

Firstly the total number of events can be retrieved by using a combination of the `--withTotalPages` and `--pageSize 1` which tells Cumulocity to include the `.statistics.totalPages` data will contains the total number of events (since we set the page size to 1).

Then, an output template is applied to the output which combines the piped device along with the statistics. We're doing a slightly lazy version which just merges the device object with the statistics object, as we'll be using a select flag to limit the amount of data.

Finally, the `select` flag is used to only show the `id`, `name` and `totalPages` properties.

Alternatively, an output template could have been used to reduce the data. The following command would also produce the same result.

<CodeExample>

```bash
c8y devices list \
| c8y events list \
    --withTotalPages \
    --pageSize 1 \
    --outputTemplate "{id: input.value.id, name: input.value.name, totalPages: output.statistics.totalPages}"
```

</CodeExample>


## Variables

The output template is very similar to the template engine, however it has a few extra variables which can be referenced when building the output. The following tables details which variables can be used.

|Name|Description|
|---|----|----|
|`output`|Output of the response (type depends on the response)|
|`request`|Object containing information about the request|
|`response`|Object with information about the response included header, url, response time etc.|
|`flags`|Object containing the flags which were set on the command|
|`input.index`|Pipeline iterator index (starting from 1)|
|`input.value`|Current pipeline item|


To show off the variables and example of each, the following command was used:

<CodeExample>

```bash
c8y devices list --name "test*" --outputTemplate "{flags: flags, response: response, request: request}"
```

</CodeExample>


### `flags`

Reference to the command's flags which are used. This can be useful either for documentation or to implement some conditional logic in the template based on the flag's values.

```json
{
  "name": "test*",
  "outputTemplate": "{flags: flags, response: response, request: request}"
}
```

:::note
Note for powershell users; The flags are the native `c8y` flags and not the PowerShell flags/parameters, though it should be fairly easy to see what the mapping is.
:::

### `request`

Request details about the request sent by the command.

```json
{
  "host": "test-ci-runner01.latest.stage.c8y.io",
  "method": "GET",
  "path": "/inventory/managedObjects",
  "pathEncoded": "/inventory/managedObjects?q=%24filter%3D+%24orderby%3Dname",
  "query": "q=$filter= $orderby=name",
  "queryParams": {
    "q": "$filter=(name eq 'test*') $orderby=name"
  },
  "url": "https://test-ci-runner01.latest.stage.c8y.io/inventory/managedObjects?q=%24filter%3D+%24orderby%3Dname"
}
```

### `response`

Response details included both the response (as a string), header values, and some additional meta information including statusCode, duration (in milliseconds).

```json
{
  "body": "<response as a string>",
  "contentLength": 7905,
  "contentType": "application/vnd.com.nsn.cumulocity.managedobjectcollection+json;charset=UTF-8;ver=0.9",
  "duration": 86,
  "header": {
    "Cache-Control": "no-cache, no-store, max-age=0, must-revalidate",
    "Connection": "keep-alive",
    "Content-Length": "7905",
    "Content-Type": "application/vnd.com.nsn.cumulocity.managedobject+json;charset=UTF-8;ver=0.9",
    "Date": "Sun, 04 Jun 2023 17:53:50 GMT",
    "Expires": "0",
    "Pragma": "no-cache",
    "Vary": "Accept-Encoding, User-Agent",
    "X-Content-Type-Options": "nosniff",
    "X-Frame-Options": "DENY"
  },
  "proto": "HTTP/1.1",
  "status": "200 OK",
  "statusCode": 200
}
```

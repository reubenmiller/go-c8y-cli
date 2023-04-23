---
title: Views
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Demo

<Video
  videoSrcURL="https://asciinema.org/a/416566/iframe?speed=1.0&autoplay=false&size=small&rows=30"
  videoTitle="Views example"
  width="90%"
  height="550px"
  ></Video>

## Overview

Views are used to help the user focus on the data that is important to them. They will only return properties of the json response which are defined in the activated view.

Views are not included in the go-c8y-cli binary itself, but rather provided by the [go-c8y-cli-addons](https://github.com/reubenmiller/go-c8y-cli-addons) repository which is installed when following the installation instructions. This allows the user to modify the views or add their own views without having to update go-c8y-cli.

The view logic behaves in the same way as the `select` parameter. Views are basically automatically applying the `select` parameter from a matching view.

## Views

### Activation

Views are activated by default when the output of the command is being written to the console. The view logic is also applied to any of following `output` options:

* csv
* csvheader
* json
* table

The view logic is **disabled** under any of the following conditions:

* Piping to a downstream command (chaining commands)
* Redirecting output (i.e. to file)
* Assigning to a variable
* The `--select` parameter is used
* The `--raw` parameter is used
* Server response is requested via `--output serverresponse`
* View is manually turned off using `--view off`

:::note
View logic is disabled when piping data because the downstream data should have access to the full object and not just the view data. This makes piping data more predictable as it does not depend on the user's local view definitions.
:::

### View definition

Each view has a criteria where it will be activated. go-c8y-cli will first scan for views in the configured directories, sort them by priority (from lowest number to highest), then pick the view which matches the response.

The following shows an example of a view definition.

```json title="file: devices.json"
{
    "version": "1",
    "definitions": [
        {
            "name": "device/agent",
            "priority": 400,
            "contentType": "",
            "self": "",
            "type": "",
            "fragments": ["c8y_IsDevice"],
            "columns": [
                "id",
                "name",
                "type",
                "owner",
                "lastUpdated",
                "c8y_Availability.status"
            ]
        }
    ]
}
```

:::note
View files are defined in a json file, and can contain multiple view definitions.
:::

### View matching

The first response item is used in the view detection. If the response contains an array (i.e. `.managedObjects[]`) then the first item from the array will be used in the detection.

A view definition supports matching by the following criteria:


* `contentType` - `Content-Type` value from the response header (supports regex)
* `self` - `.self` value of first entry (supports regex)
* `type` - `.type` value of first entry (supports regex)
* `fragments` - List of fragments (all fragments must exist)
* `requestPath` - Match the outgoing request PATH (supports regex)
* `requestMethod` - Match the outgoing request Method (e.g. GET, POST etc.) (supports regex)

:::note
The matching criteria can be combined to provide more precise matching.
::: 

:::note
If two or more definitions match on the same data, then the view with the lowest `priority` number will be chosen.
:::

## Tab completion

Views support full tab completion.

```
✓ ~ % c8y devices list --view device<TAB><TAB>
device/agent       -- [id name type owner lastUpdated c8y_Availability.status] | file: devices.json
deviceCredentials  -- [id tenantId username password self] | file: deviceCredentials.json
deviceGroup        -- [id lastUpdated name type owner] | file: deviceGroup.json
```

:::tip
If tab completion is not working, either you don't have any view paths configured, or your tab completion is generally not working (try re-installing following the install instructions)
:::

## Examples

### Manually choose a view

Views can also be manually selected by specifying the view name to the `view` parameter:

<CodeExample>

```bash
c8y devices list --view "device/agent"
```

</CodeExample>

### Forcing auto view detection when redirecting to file

Normally view detection is disabled when redirecting to file, however it can be force using `view auto`.

<CodeExample>

```bash
c8y devices list --view auto > myfile.json
```

</CodeExample>

### Turning off view logic

The view logic can be turned off using `view off`

<CodeExample>

```bash
c8y devices list --view off --output json
```

</CodeExample>

Alternatively, you can use the fact that piped data automatically turns off the view, and pipe the data directly to `jq` (if you have it installed).

This method is more convenient as piping the data will also force the output to json, thus it is less typing than then previous example.

<CodeExample transform="false">

```bash
c8y devices list | jq
```

</CodeExample>


## View definition examples

Check out the view definitions in [go-c8y-cli-addons](https://github.com/reubenmiller/go-c8y-cli-addons/tree/main/views/default) for more examples.

### Match on type and fragments

The following view is used to display smartgroup information. It combines both `type` and `fragments`.

```
{
    "version": "1",
    "definitions": [
        {
            "name": "smartgroup",
            "priority": 400,
            "type": "c8y_DynamicGroup",
            "fragments": ["c8y_IsDynamicGroup"],
            "columns": [
                "id",
                "name",
                "c8y_DeviceQueryString",
                "type",
                "owner",
                "lastUpdated",
                "c8y_IsDynamicGroup.invisible"
            ]
        }
    ]
}
```

### Match collection statistics

A special view used to display the summary information.

Since this view is using root fragments (fragments outside of the array data (i.e. `.managedObjects[]`)), the view can only be activated when using `withTotalPages` parameters are being used.

```
{
    "version": "1",
    "definitions": [
        {
            "name": "statistics",
            "priority": 100,
            "contentType": "collection",
            "fragments": ["statistics.pageSize", "self"],
            "columns": [
                "totalPages:statistics.totalPages",
                "pageSize:statistics.pageSize",
                "currentPage:statistics.currentPage"
            ]
        }
    ]
}
```

<CodeExample transform="false">

```bash
c8y devices list --view statistics --withTotalPages
```

</CodeExample>

```json title="Output"
| totalPages | pageSize   | currentPage |
|------------|------------|-------------|
| 19         | 20         | 1           |
```

### Match on Content-Type

The following contains multiple view definition files related to the user role object in Cumulocity. It uses the Content-Type value and

```json title="file: role.json"
{
    "version": "1",
    "definitions": [
        {
            "name": "role",
            "priority": 500,
            "contentType": "vnd.com.nsn.cumulocity.role\\+json",
            "columns": [
                "id",
                "name",
                "self"
            ]
        },
        {
            "name": "roleCollection",
            "priority": 500,
            "contentType": "vnd.com.nsn.cumulocity.roleCollection\\+json",
            "columns": [
                "id",
                "name",
                "self"
            ]
        },
        {
            "name": "roleReference",
            "priority": 500,
            "contentType": "vnd.com.nsn.cumulocity.roleReference\\+json",
            "columns": [
                "role.id",
                "role.name",
                "role.self"
            ]
        }
    ]
}
```

:::note
You will need to escape regex characters if you want the literal equivalent. i.e. `\\+json` if you want to match the `+json` string
:::

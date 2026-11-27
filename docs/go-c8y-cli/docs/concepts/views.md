---
title: Views
---

import CodeExample from '@site/src/components/CodeExample';
import AsciinemaPlayer from '@site/src/components/AsciinemaPlayer';

## Demo

<AsciinemaPlayer
  src="https://asciinema.org/a/416566.cast"
  rows={30}
  preload
  fit="width"
  theme="monokai"
  poster="npt:0:04"
/>

## Overview

Views are used to help the user focus on the data that is important to them. They will only return properties of the json response which are defined in the activated view.

Views are not included in the go-c8y-cli binary itself, but rather provided by the [go-c8y-cli-addons](https://github.com/reubenmiller/go-c8y-cli-addons) repository which is installed when following the installation instructions. This allows the user to modify the views or add their own views without having to update go-c8y-cli.

The view logic behaves in the same way as the `select` parameter. Views are basically automatically applying the `select` parameter from a matching view.

## Views

### Activation

Views are activated by default when the output of the command is being written to the console. The view logic is also applied to any of following `output` options:

* csv
* csvheader
* tsv
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

### Column widths

By default, the width of each column is calculated from the sampled rows (see `settings.views.sampleSize`), and limited by the `settings.views.columnMinWidth` and `settings.views.columnMaxWidth` settings.

A column can also be given as an object to control its width in the table output. String and object columns can be mixed in the same view.

```json title="file: devices.json"
{
    "version": "1",
    "definitions": [
        {
            "name": "device/agent",
            "priority": 400,
            "fragments": ["c8y_IsDevice"],
            "columns": [
                {"path": "id", "width": 12},
                "name",
                {"path": "description", "maxWidth": 40},
                {"name": "model", "path": "c8y_Hardware.model", "minWidth": 15},
                {"path": "c8y_Availability.status", "minWidth": 12}
            ]
        }
    ]
}
```

A column object supports the following properties:

* `path` - Property to display (required). Uses the same syntax as the `--select` flag, including wildcards such as `tags.**`
* `name` - Column name to display instead of the path. `{"name": "model", "path": "c8y_Hardware.model"}` is the same as `"model:c8y_Hardware.model"`
* `width` - Fixed column width in characters. Longer values are truncated or wrapped, depending on `settings.views.rowMode`. Takes precedence over `minWidth` and `maxWidth`
* `minWidth` - Minimum column width in characters. Overrides `settings.views.columnMinWidth` for this column
* `maxWidth` - Maximum column width in characters. Overrides `settings.views.columnMaxWidth` for this column

If the `path` contains a wildcard, then the width applies to all of the columns which it selects.

:::note
Column widths only affect the table output, so other output formats (e.g. json or csv) are not changed. A column is never narrower than its header, and columns are still removed if the table is wider than the terminal. The column widths are not used when the columns are selected with `--select`.
:::

### View matching

The view is detected once for each stream of results, from the first document. If the response contains an array (i.e. `.managedObjects[]`) then the first item from the array is used. The request path, request method and the `Content-Type` of the response are matched at the same time. If no view matches the first document, no view is applied to the rest of the stream.

When a command sends several requests (for example, when it follows multiple pages), the detection is repeated for the next response until a view matches. Once a view is chosen, it is used for the remaining documents.

A view definition supports matching by the following criteria:


* `contentType` - `Content-Type` value from the response header (supports regex)
* `self` - `.self` value of first entry (supports regex)
* `type` - `.type` value of first entry (supports regex)
* `fragments` - List of fragments (all fragments must exist)
* `requestPath` - Match the outgoing request PATH (supports regex)
* `requestMethod` - Match the outgoing request Method (e.g. GET, POST etc.) (supports regex)

Set `"autoMatch": false` to exclude a view from the automatic detection. The view can then only be used by name, either with `--view <name>` or as the [default view of an extension API command](/docs/concepts/extensions/api-commands/command-groups/#example-setting-the-default-view-of-a-command).

:::note
The matching criteria can be combined to provide more precise matching.
::: 

:::note
If two or more definitions match on the same data, then the view with the lowest `priority` number will be chosen.
:::

:::note
An extension API command can set its own default view. That view takes precedence over the automatic detection, but not over a view chosen by the user with `--view`.
:::

### Columns of a view

In the `table`, `csv`, `csvheader` and `tsv` output, every column of the view is always present, in the order of the view. If a document does not have the property, the cell is empty. This means the columns do not depend on which documents happen to be returned first.

```bash
c8y devices list --output csvheader --view auto
```

```csv title="Output"
id,name,c8y_Agent.version,c8y_Hardware.model
10030,vz-agent,1.2,M1
10029,vz-plain,,
```

The `json` output only contains the properties that a document has, so a document that lacks a column does not get an empty value.

The view is applied before `--flatten`, so the flattened output only contains the properties of the view.

```bash
c8y devices list --output json --view auto --flatten --compact
```

```json title="Output"
{"c8y_Agent.version":"1.2","c8y_Hardware.model":"M1","id":"10030","name":"vz-agent"}
{"id":"10029","name":"vz-plain"}
```

### Unknown view name

If the name given to `--view` does not match any view, the CLI prints a warning (`no matching view found. name=<name>`) and writes the output without a view.

```bash
c8y devices list --view nope
```

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

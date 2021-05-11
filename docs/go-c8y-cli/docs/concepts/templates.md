---
layout: default
category: Concepts
title: Templates
---

import CodeExample from '@site/src/components/CodeExample';

## Overview

Templates are a powerful way to create re-useable body templates with the ability to randomize your data. All commands which send a request body to Cumulocity support using a template.

The template combines with the existing command arguments to build the request's body which is sent to Cumulocity.

## Background

Creating data via the command line can be time consuming especially if there are are lot of required input parameters. To make things easier, the c8y cli tool supports data templates when creating or updating Cumulocity objects (i.e. measurements, managed objects, events, alarms etc.). The data templates allow the user to specify a file containing the template instead of having to provide all of the inputs manually on the command line.

The data templates are implemented using the data template language `jsonnet` (pronounced "jay sonnet"). This is an unofficial Google product which can be used to create json structures in a simple yet powerful way. `jsonnet` looks a lot like json however it is not so strict with the syntax, and you can also use variables, expressions, functions, inside a template. All of these features give the user a lot of flexibility to be able to create custom data structures quickly and efficiently.

Data templates are supported for all commands which create new data or edit existing data in Cumulocity.

Please read the examples to understand the usage of such templates, and checkout the [jsonnet documentation](https://jsonnet.org/) for more about the language and a live editor to experiment with its features.

## Example: Creating a measurement

Let's assume that you are building a weather monitoring application and you want to simulate some measurements in Cumulocity. You want to create a measurement which has two different series.

* `c8y_Weather.temperature`
* `c8y_Weather.barometricPressure`

You can create a jsonnet template file with the following contents.

```jsonnet title="file: example.jsonnet"
{
    c8y_Weather: {
        // simulate temperature in Brisbane, Australia: 15 - 35 °C
        temperature: {
            value: _.Int(35,15),
            unit: "°C",
        },

        // simulate barometric pressure: 980.0 - 1030.0 hPa with a precision of 1 decimal place
        barometricPressure: {
            value: _.Float(1030, 980, 1),
            unit: "hPa",
        },
    },
}
```

The template can then be used from the command line using the `template` parameter.

<CodeExample>

```bash
c8y measurements create --device 1234 --template ./example.jsonnet --dry
```

</CodeExample>

```json title="Request Body"
{
  "c8y_Weather": {
    "barometricPressure": {
      "unit": "hPa",
      "value": 1018.7
    },
    "temperature": {
      "unit": "°C",
      "value": 34
    }
  },
  "source": {
    "id": "1234"
  },
  "time": "2021-05-09T09:46:42.8280961Z",
  "type": "iot_weather_sensor1"
}
```

:::note
The template will be used as the base for constructing the body, then the additional parameters, like device will be added to it. This keeps your template free of hard coding any device ids.
:::

The template should mostly look familiar to you as jsonnet uses a json-like structure. However `jsonnet` extends json by adding some of the following features:

* Allowing comments (using `//`)
* Trailing commas
* Strings can be defined using single or double quotes
* Arithmetic, functions, local variables can be used

The example also uses some internally defined functions which are provided by go-c8y-ci, for example, `_.Int()` and `_.Float()`. These functions are not provided by default by jsonnet but are injected by go-c8y-cli into each template when it is being evaluated.

## Example: Template string to update a list of devices

Templates don't have to be stored in a file, you can also provide it as a string. This makes it possible to utilize some of the template functions quickly to update 

The following adds a fragment to a list of devices (via pipeline) and adds the timestamp when the fragment was applied using the `_.Now()` helper function.

<CodeExample>

```bash
c8y devices list |
    c8y devices update \
        --template "{c8y_ScriptResult: {status: 'OK', lastUpdated: _.Now()}}" \
        --select "id,name,c8y_ScriptResult" \
        --output json -c
```

</CodeExample>

```json title="Output"
{"c8y_ScriptResult":{"lastUpdated":"2021-05-09T11:04:05.650Z","status":"OK"},"id":"503627","name":"demo_001"}
{"c8y_ScriptResult":{"lastUpdated":"2021-05-09T11:04:05.652Z","status":"OK"},"id":"503533","name":"demo_002"}
{"c8y_ScriptResult":{"lastUpdated":"2021-05-09T11:04:05.750Z","status":"OK"},"id":"503624","name":"demo_003"}
{"c8y_ScriptResult":{"lastUpdated":"2021-05-09T11:04:05.785Z","status":"OK"},"id":"503625","name":"demo_004"}
{"c8y_ScriptResult":{"lastUpdated":"2021-05-09T11:04:05.818Z","status":"OK"},"id":"503626","name":"demo_005"}
```

## Template Variables

Variables can also be injected into your template at runtime. These variables can then be referenced from inside your jsonnet template.

For example if we want to generate a device managed object from a template, however you would like to change the type at runtime (via the command line).

Firstly when creating your jsonnet template, you can reference a template variable using the `var(name, [default])` function, where it accepts the name of the template variable and an optional default value (in case the variable is not provided by the user)

The following is an example of such a template:

```jsonnet
{
    name: "my device",
    type: var("type", "defaultType"),
    var("fragment", "c8y_Default"): {},
    c8y_IsDevice: {},
}
```

The template can then be used when creating a managed object, and the `type` and `fragment` variables can be injected by using the `templateVars` parameter.

<CodeExample>

```bash
c8y inventory create \
    --template "./examples/templates/device.jsonnet" \
    --templateVars "type=macOS,fragment=customer_Agent" \
    --dry
```

</CodeExample>

```json title="Output"
{
    "name": "my device",
    "type": "macOS",
    "customer_Agent": {},
    "c8y_IsDevice": {}
}
```

## Template Functions (added by go-c8y-cli)

Below lists the additional functions which are available in jsonnet template files. These functions are added to your template automatically by the c8y cli tool itself, and are not part of the standard jsonnet library. The built-in [jsonnet standard library](https://jsonnet.org/ref/stdlib.html) provides additional functions that can be used in combination with those injected by go-c8y-cli.


| Function | Description | Example |
|----------|-------------|---------|
|var(name, [defaultValue])|Reference a template variable from the `templateVars` cli parameter| **value of variable** |
|_.GetURLPath(url)| Get the URL path from a string | `/test?pageSize=1` |
|_.GetURLHost(url)| Get the hostname from a string | `https://example.com` |
|_.Int([max=100],[min=0])| Random integer (int64) between min and (max - 1) inclusive | `42` |
|_.Bool()| Generate a random boolean value | `true` |
|_.Float([max=100],[min=0])| Random float between min and max (inclusive)| `42.0` |
|_.Now([offset='0s'])| Generate ISO8601 date (with millisecond resolution) from a relative date or date string | `2021-05-09T07:48:43.949Z` |
|_.NowNano([offset='0s'])| Generate ISO8601 date (with nanosecond resolution) from a relative date or date string | `2021-05-09T07:48:33.0030745Z` |
|_.Name([prefix=''],[postfix=''])| Generate a random name with an optional prefix and postfix | `GfheJoa;Ktx,F56s` |
|_.Password([length=32])| Generate a randomized password of a specified length | `NlXYngK;bj!xeOvhpydJ4VIG6HuDcLBR`  |
|_.Hex([length=16])| Random Hexadecimal string of a given length. | `8b7e0736a5a6ed80` |
|_.Char([length=16])| Random string with only character a-zA-Z of a given length  | `exxUQqCDFwRHpUog` |
|_.Digit([length=16])| Random 0 padded string with only digits | `0261177197719716` |
|_.AlphaNumeric([length=16])| Random AlphaNumeric string of a given length | `f087oAAzjnvzkPdf` |
|_.StripKeys([object])| Strip protected Cumulocity properties from a object. The following properties are removed from the given object: additionParents, assetParents, childAdditions, childAssets, childDevices, deviceParents, creationTime, lastUpdated, self | `{}` |


### Example: Generating random data

Below is an example jsonnet template making use of the above functions provided by go-c8y-cli.

```jsonnet title="file: template.jsonnet"
{
    boolean: {
        Bool: _.Bool(),
    },
    numbers: {
        Float: _.Float(1, 0.5),
        Int: _.Int(50, -50),
    },
    dateString: {
        Now: _.Now('-5m'),
        NowNano: _.NowNano('-10d'),
    },
    url: {
        GetURLPath: _.GetURLPath('https://example.com/test?pageSize=1'),
        GetURLHost: _.GetURLHost('https://example.com/test?pageSize=1'),
    },
    strings: {
        Name: _.Name('device_', '_example'),
        Password: _.Password(32),
        Char: _.Char(16),
        Digit: _.Digit(10),
        Hex: _.Hex(24),
        AlphaNumeric: _.AlphaNumeric(20),
    }
}
```

<CodeExample>

```bash
c8y template execute --template "./template.jsonnet"
```

</CodeExample>


```json title="Output"
{
  "boolean": {
    "Bool": true
  },
  "dateString": {
    "Now": "2021-05-09T08:01:22.505Z",
    "NowNano": "2021-04-29T08:06:22.5057679Z"
  },
  "numbers": {
    "Float": 0.8675182149724747,
    "Int": 29
  },
  "strings": {
    "AlphaNumeric": "dm7J1CxhnH1KC8grDLnu",
    "Char": "ZsPIfwRTFtGafuMI",
    "Digit": "3496063026",
    "Hex": "ffc9ceb63ac73a463dd5d202",
    "Name": "device_q4b9L)p+WjMhJGsm_example",
    "Password": "u_ZsojR5ekPYh8WyvldDBNFincVOzUC["
  },
  "url": {
    "GetURLHost": "https://example.com",
    "GetURLPath": "/test?pageSize=1"
  }
}
```

:::tip
jsonnet has its own standard library which provides useful functions like uppercase, lowercase etc. For more information about jsonnet and its additional functions please checkout the following links:
* [jsonnet overview](https://jsonnet.org/)
* [jsonnet standard library](https://jsonnet.org/ref/stdlib.html)
:::

## Referencing piped input in a template

go-c8y-cli supports chaining multiple commands, the `template` function supports referencing the current pipeline item via a special variables injected into the template engine by go-c8y-cli before the template is evaluated.


| Variable | Description |
|----------|-------------|
| input.index | Pipeline iterator index (starting from 1) |
| input.value | Current pipeline item |


### Example: Referencing line delimited input in a template

If the piped input is not json, then the current line being processed will be available from the `input.value` variable in the template.

Each line is processed through the template separately. The following shows piping 3 lines being piped to the generic `c8y template execute` command and referencing the input in the template string to show how the values are mapped

<CodeExample>

```bash
echo -e "one\ntwo\nthree" |
    c8y template execute \
        --template "{ index: input.index, value: input.value, input:: '' }" \
        --output json -c
```

```powershell
"one", "two", "three" |
    c8y template execute `
        --template "{ index: input.index, value: input.value, input:: '' }" `
        --output json -c
```

```powershell
"one", "two", "three" |
    Invoke-Template `
        -Template "{ index: input.index, value: input.value, input:: '' }" `
        -Output json -Compact
```

</CodeExample>

```json title="Output"
{"index":1,"value":"one"}
{"index":2,"value":"two"}
{"index":3,"value":"three"}
```

:::note
`c8y template execute` automatically maps the piped input to the `input` property of the body. The above example is hiding this property by using the jsonnet syntax `prop::` which causes this property to be excluded in the output. 
:::

### Example: Referencing piped json in a template

When piping json lines to a command, the current json line row can be referenced from the `input.value` variable. Since json lines is being piped, this variable represents the complect json line object, and nested properties can be accessed directly from it, i.e. `input.value.type`, however an jsonnet will throw an exception if the property does not exist.

For example take the following json line file. Each line represents a full json object (it is not an array!)

```json title="file: input.json"
{"type": "linux_agent", "ramSize": "38MB"}
{"type": "macOS_agent", "ramSize": "40MB"}
{"type": "windows_agent", "ramSize": "50MB"}

```

The contents of the file can be piped to any command, in this case we are using command to create a new managed object from the piped input. The template reshapes the input value to custom fragments.

<CodeExample>

```bash
cat input.json |
    c8y inventory create \
        --template "{ type: input.value.type, agentProperties: { ram: input.value.ramSize } }" \
        --select id,type,agentProperties \
        --output json
```

</CodeExample>

```json title="Output"
{
  "agentProperties": {
    "ram": "38MB"
  },
  "id": "520078",
  "type": "linux_agent"
}
{
  "agentProperties": {
    "ram": "40MB"
  },
  "id": "519798",
  "type": "macOS_agent"
}
{
  "agentProperties": {
    "ram": "50MB"
  },
  "id": "520121",
  "type": "windows_agent"
}
```

### Example: Retry a failed operation

Retrying an operation just involved creating a new operation based on the contents of the failed operation.

The `select` parameter can be used to remove the readonly fragments from the failed operation before passing it to the create operation command.

The `template` variable just references passes the piped input untouched as data being piped is already in json format.

<CodeExample>

```bash
c8y operations list --status "FAILED" --select '**,!id,!deviceName!status,!creationTime' |
        c8y operations create --template "input.value"
```

</CodeExample>

### Example: Clone a device

Templates can be used to clone existing items by first fetching the reference item, piping it to the create command, and then using the `input.value` variable which references the current piped input line in the `template` parameter.

<CodeExample>

```bash
c8y devices list --pageSize 1 --select '**,copiedFrom:id,!id,!lastUpdated' |
    c8y devices create --name "myclone" --template "input.value" --output json
```

</CodeExample>

:::note
`select '**,copiedFrom:id,!id,!lastUpdated'` is used to select all of the input properties except for id and lastUpdated as these properties can not be part of the body for the create device command. It also maps the id field to `copiedFrom` so a reference to the source device will be present in the newly created device.
:::

```json title="Output"
{
  "copiedFrom": "507325",
  "creationTime": "2021-05-07T16:39:18.827Z",
  "id": "519792",
  "lastUpdated": "2021-05-09T21:06:40.620Z",
  "name": "myclone",
  "owner": "exampleuser",
  "self": "https://t12345.example.c8y.com/inventory/managedObjects/519792",
  "type": "altest"
}
```

## Template settings

Template (jsonnet) files can be stored in a folder and this folder can be added to the `settings.template.path` settings within the global or session configuration.

```json
{
    "settings": {
        "template.path": "/workspaces/go-c8y-cli/.cumulocity/templates"
    }
}
```

The template path will be used when looking up template filenames if the user does not specify a relative or full path to it.

i.e.

```bash
c8y template execute --template event.jsonnet
```

c8y will look for a file called "event.jsonnet" inside the `template.path` folder.

PSc8y also supports argument completion for template files if the path is set.

```powershell
PS /workspaces/go-c8y-cli> Invoke-Template -Template <tab>

alarm.jsonnet                 device.template.jsonnet       event.jsonnet                 measurement.advanced.jsonnet
```

## Complex Examples

### Create devices with variable software lists

Add a complex device managed object where the number of installed software applications in the `c8y_SoftwareList` fragment can be set via the command line.

```jsonnet title="file: device.template.jsonnet"
// Helper: Create software entry
local newSoftware(i) = {
    name: "app" + i,
    version: "1.0." + i,
    url: "",
};

// Output: Device Managed Object
{
    name: "name1",
    value: var("name", "defaultName"),
    type: var("type"),
    ["c8y_" + var("type")]: {},
    c8y_SoftwareList: [ newSoftware(i) for i in std.range(1, var("softwareCount", 1))],
}
```

<CodeExample>

```bash
c8y inventory create \
    --dry \
    --template ./examples/templates/device.template.jsonnet \
    --templateVars "softwareCount=2"
```

</CodeExample>

```json title="Output"
{
   "c8y_SoftwareList": [
     {
       "name": "app1",
       "url": "",
       "version": "1.0.1"
     },
     {
       "name": "app2",
       "url": "",
       "version": "1.0.2"
     }
   ],
   "c8y_defaultType": {},
   "name": "name1",
   "type": "defaultType",
   "value": "test"
}
```

## Template development

To help with the development of templates, there is a command which evaluates a template and prints the output to the console.

<CodeExample>

```bash
c8y template execute --template ./mytemplate.jsonnet
```

</CodeExample>

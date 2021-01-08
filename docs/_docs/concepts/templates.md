---
layout: default
category: Concepts
title: Templating
---

### Background

Creating data via the command line can be time consuming especially if there are are lot of required input parameters. To make things easier, the c8y cli tool supports data templates when creating or updating Cumulocity objects (i.e. measurements, managed objects, events, alarms etc.). The data templates allow the user to specify a file containing the template instead of having to provide all of the inputs manually on the command line.

The data templates are implemented using the data template lanaguage `jsonnet` (pronounced "jay sonnet"). This is an unofficial Google product which can be used to create json structures in a simple yet powerful way. `jsonnet` looks a lot like json however it is not so strict with the syntax, and you can also use variables, expresions, functions, inside a template. All of these features give the user a lot of flexibility to be able to create custom data structures quickly and efficiently.

Data templates are supported for all commands which create new data or edit existing data in Cumulocity.

Please read the examples to understand the usage of such templates, and checkout the [jsonnet documentation](https://jsonnet.org/) for more about the language and a live editor to experiment with its features.

### Example

Let's assume that you want to simulate create some mock measurements in Cumulocity to assist with prototyping. Say that you are building a weather application and you want to create a measurement which has two different series.

* `c8y_Weather.temperature`
* `c8y_Weather.barometricPressure`

You can create a jsonnet template file with the following contents:

```jsonnet
{
    c8y_Weather: {
        temperature: {
            value: rand.int,
            unit: "°C",
        },
        barometricPressure: {
            value: rand.float * 100 + 1000,
            unit: "Pa",
        },
    },
}
```

The template should mostly look familar to you as jsonnet uses a json-like structure. Some notable differences are that the properties do not required quotes and you are allowed to have trailing commas on fields. In addition the property values (and fields) can be defined as an expression rather than static values.

The example also uses some internally defined variables which are provided by the c8y cli tool: `rand.int` and `rand.float`.

`rand` is an object which is injected by the c8y cli tool, which has a few properties which contain randomized data. For example, `rand.int` returns a random integer between 0 and 100, and a `rand.float` returns a 32 bit float between 0 and 1.

The random values can be used to build up mock measurement, and you can control the values by using them in an expression. For example, if you wanted to use a value that ranges between -50 and 50, then you have subtract 50:

```jsonnet
{
    // Integer between -50 and 50
    value: rand.int - 50,
}
```

And if that is still not enough, you can use an if/else statement where it uses the value of `rand.bool` to switch between two different sets of expressions.

```jsonnet
{
    value: if rand.bool then -20 else rand.int * 10,
}
```

### Template Variables

Variables can also be injected into your template at runtime. These variables can then be referenced from inside your jsonnet template.

For example if we want to generate a device managed object from a template, however you would like to change the type at runtime (via the command line).

Firstly when creating your jsonnet template, you can reference any template variables using the `var()` function, where it accepts the name of the template variable and an optional default value (in case the variable is not provided by the user)

The following is an example of such a template:

```jsonnet
{
    name: "my device",
    type: var("type", "defaultType"),
    var("fragment"): {},
    c8y_IsDevice: {},
}
```

The template can then be used when creating a managed object, and the `type` and `fragment` variables can be injected by using the `templateVars` parameter.

Note: Multiple template variables can be provided, by using a comma separated list.

**Bash/zsh**

```sh
c8y inventory create \
    --template ./examples/templates/device.jsonnet \
    --templateVars "type=myCustomType1,fragment=myCustomObject" \
    --dry
```

**PowerShell**

```sh
New-ManagedObject `
    -Template ./examples/templates/measurement.jsonnet `
    -TemplateVars "type=myCustomType1,fragment=myCustomObject" `
    -WhatIf
```

**Output**

These command would produce the following body which would be sent to Cumulocity.

```json
{
    "name": "my device",
    "type": "myCustomType1",
    "myCustomObject": {},
    "c8y_IsDevice": {}
}
```


### Available (automatic) variables in templates

Below shows all of the variables which are available for use in jsonnet template files. There are multi randomized values which are injected into the template at runtime. These variables are added by the c8y cli tool iteself, and are not part of the standard jsonnet library.

| Variable | Description |
|-------|---------|
| rand.bool | Random boolean value, either true or false |
| rand.int | Random integer between 0 and 100 (inclusive) |
| rand.int2 | Random integer between 0 and 100 (inclusive) |
| rand.float | Random float32 between 0 and 1 |
| rand.float2 | Random float32 between 0 and 1 |
| rand.float3 | Random float32 between 0 and 1 |
| rand.float4 | Random float32 between 0 and 1 |

**Note**

The random values are assigned as variables and not functions. This means if you reference `rand.int` twice in the template, it will have the same value.

Additional information about jsonnet and it's synatax can be found [here](https://jsonnet.org/)


### Template settings

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

```sh
c8y template execute --template event.jsonnet
```

c8y will look for a file called "event.jsonnet" inside the `template.path` folder.

PSc8y also supports argument completion for template files if the path is set.

```powershell
PS /workspaces/go-c8y-cli> Invoke-Template -Template <tab>

alarm.jsonnet                 device.template.jsonnet       event.jsonnet                 measurement.advanced.jsonnet
```

**Note**

Argument completion for template filenames is not yet supporte for bash or zsh.

### Complex Example

Add a complex device managed object where the number of installed software applications in the `c8y_SoftwareList` fragment can be set via the command line.

*File: device.template.jsonnet*

```jsonnet
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

**Bash/zsh**

```sh
c8y inventory create \
    --dry \
    --template ./examples/templates/device.template.jsonnet \
    --templateVars "softwareCount=2"
```

**PowerShell**

```sh
New-ManagedObject `
    -Whatif `
    -Template ./examples/templates/device.template.jsonnet `
    -TemplateVars "softwareCount=2"
```

**Output**

The following output shows the body that will be uploaded to Cumulocity when creating the managed object.

```json
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

### Template development

To help with the development of templates, there is a command which evaluates a template and prints the output to the console.

**Bash/zsh**

```sh
c8y template execute --template ./mytemplate.jsonnet
```

**PowerShell**

```powershell
Invoke-Template -Template ./template.jsonnet
```

#### Input data (PowerShell)

`Invoke-Template` also supports piping of input data which can be referenced 

```powershell
$Template = @"
{
    type: base.name + '_' + rand.int,
    value: 1 + 2,
}
"@

$InputData = @(
    @{ name = "device1" },
    @{ name = "device2" }
)
$templateOutput = $InputData |
    Invoke-Template -Template $Template -Compress
```

**Output**

```json
{"name":"device1","type":"device1_70","value":3}
{"name":"device2","type":"device2_25","value":3}
```

Or the output can be piped to `ConvertFrom-Json`:

```powershell
$templateOutput = $InputData | Invoke-Template -Template $Template -Compress | ConvertFrom-Json
```

```sh
name  type     value
----  ----     -----
name  name_97      3
name2 name2_54     3
```

**Note**

When piping the output to `ConvertFrom-Json`, the `-Compress` option needs to be used otherwise it will return an error due to invalid JSON. This is because the output is a stream of json text and not a json array of results.

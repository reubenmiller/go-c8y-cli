---
layout: default
category: Concepts
title: Templating
---

### Data Templates


### Exercise

Assume that you want to simulate create some mock some measurements in Cumulocity to assist with prototyping. Let's say that you are building a weather application and you want to create a measurement which has two different series.

    * `c8y_Weather.temperature`
    * `c8y_Weather.barometricPressure`

You can create a simple jsonnet template file.

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

The template should mostly look familar to you as jsonnet uses a json-like structure. The interesting parts which might not be 100% clear are the two references to variables; `rand.int` and `rand.float`.

`rand` is an object which is injected by the c8y cli tool, which has a few properties which contain randomized data. `rand.int` returns a random integer between 0 and 10, and a `rand.float` returns a 32 bit float between 0 and 1.

The random values can be used to build up mock measurement, and you can control the range by using simple muliplication, divition and addition:

For example if you wanted to use a value that ranges between -50 and 50, then you have subtract 50:

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

#### Template Variables (TODO)

Variables can also be injected into your script, and the values can be passed at runtime, 

**input.vars.json**

{
    "type": "myCustom1",
    "softwareCount": 2
}



### Usage with cli

Once you have created a template you use it by passing the path to the template to the `template` parameter:

```sh
c8y measurements create \
    --device 12345 \
    --time "0s" \
    --type "c8y_Measurement" \
    --dry \
    --template ./examples/templates/measurement.jsonnet
```

```powershell
New-Measurement `
    -Device 12345 `
    -Time "0s" `
    -Type "c8y_Measurement" `
    -WhatIf `
    -Template ./examples/templates/measurement.jsonnet
```

### Available (automatic) variables

Below shows all of the variables which are available for use in the jsonnet template files. There are multi

| Variable | Description |
|-------|---------|
| rand.bool | Random boolean value, either true or false |
| rand.int | Random integer between 0 and 100 (inclusive) |
| rand.int2 | Random integer between 0 and 100 (inclusive) |
| rand.float | Random float32 between 0 and 1 |
| rand.float2 | Random float32 between 0 and 1 |
| rand.float3 | Random float32 between 0 and 1 |
| rand.float4 | Random float32 between 0 and 1 |


More information about jsonnet and it's synatax can be found [here](https://jsonnet.org/)



### Complex Example

Add a complex device managed object where the number of installed software applications in the c8y_SoftwareList fragment can be set via a variable

*device.template.jsonnet*

```jsonnet
// Helper: Create software entry
local newSoftware(i) = {
    name: "app" + i,
    version: "1.0." + i,
    url: "",
};

// Settings: Default values:
local defaults = {
    name: "test",
    type: "defaultType",
    softwareCount: 2,
} + vars;

// Output: Device Managed Object
{
    name: "name1",
    value: defaults.name,
    type: defaults.type,
    ["c8y_" + defaults.type]: {},
    c8y_SoftwareList: [ newSoftware(i) for i in std.range(1, defaults.softwareCount)],
}
```

**Bash/zsh**

```sh
c8y inventory create --dry --template ./examples/templates/device.template.jsonnet --templateVars "softwareCount=2"
```


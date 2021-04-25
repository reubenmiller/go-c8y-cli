---
layout: default
category: Concepts
title: Select Parameter
---

### Select parameter (new in v2.0.0)

The `select` parameter can be used to limit which fragments are returned by the cli tool and to provide a convenient way to modify the output response.

The properties to include in the output response can be given by using dot notation.

For example, given the following json:

```json
{
    "id": "1234",
    "c8y_Hardware": {
        "serialNumber": "abcdef"
    },
    "values": [1, 2, 3],
    "nested_values": [
        {
            "item": {
                "value": 10
            }
        }
    ]
}
```

The dot notations for the above json are:

```bash
id
c8y_Hardware.serialNumber
values.0
values.1
values.2
nested_values.0.item.value
```

Any combination of these can then be used in the `select` parameter: 

```bash
c8y devices list --select "id,values.0"
```

```json
{
    "id": "1234",
    "values": [1]
}
```

*Note*: Array indexes are mapped to `.<index>`, therefore there is no difference between between paths for an array index, and a json object using a number as a property. This may generate some unexpected results when using `--select` if you use numbers as properties in you Cumulocity data.


Entering explicit values is not very convenient, especially when some fragments can be very long, that's why the `select` parameter also supports usage of the following wildcard characters:

* `?` matches a single character not including the path delimiter `.`
* `*` matches zero or more characters not including the path delimiter `.`
* `**` (a.k.a. globstar) matches all characters including the path delimiter `.`

All dot notation paths are case-insensitive. If more than 1 property is matches the same property, then both will be returned.

### Select features

Below is a summary of actions which are supported by the `select` parameter:

|use case|usage|
|--|--|
|Get specific properties|`--select "id,name"`|
|Get root fragments which are not objects or arrays|`--select "*"`|
|Get all fragments (included nested objects and arrays)|`--select "**"`|
|Get all items matching a nested path pattern|`--select "**.self"`|
|Map property names (id->deviceId)|`--select "deviceId:id`|
|Get flattened json properties|`--select "**" --flatten`|
|Output results in CSV format (comma delimited) (with flattened json paths)|`--output csv` or `--output csvheader`|


## Examples

The following examples should show how the `select` parameter can be used. All of the examples reference the following json from a device (managed object).

**Reference device data used in each example**

```json
{
  "additionParents": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/additionParents"
  },
  "agent_Details": {
    "country": {
      "code": "61",
      "name": "AU"
    },
    "details": [1, 2, 3]
  },
  "assetParents": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/assetParents"
  },
  "c8y_Hardware": {
    "serialNumber": "XYDA010"
  },
  "c8y_IsDevice": {},
  "c8y_SoftwareList": [
    {
      "name": "app1",
      "url": "https://myexample.com/packages/app1/1.0.0",
      "version": "1.0.0"
    }, 
    {
      "name": "app2",
      "url": "https://myexample.com/packages/app2/2.0.0",
      "version": "2.0.0"
    }
  ],
  "childAdditions": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/childAdditions"
  },
  "childAssets": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/childAssets"
  },
  "childDevices": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/childDevices"
  },
  "company_Example": {},
  "creationTime": "2021-02-20T10:49:37.621Z",
  "deviceParents": {
    "references": [],
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735/deviceParents"
  },
  "id": "396735",
  "lastUpdated": "2021-02-20T10:49:37.621Z",
  "name": "device001",
  "owner": "ciuser01",
  "self": "https://t12345.example.c8y.com/inventory/managedObjects/396735",
  "type": "c8y_Linux"
}
```

##### Select non-nested fragments (i.e. literals like strings, numbers and booleans) using wildcards

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,nam*"
```

Note: In shells like zsh, bash, and fish, remember to include the values within quotes `"` to prevent `*` from being expanded by the shell interpreter.

**Output**

```json
{
  "id": "396806",
  "name": "1"
}
```

##### Select nested properties using dot notation

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,c8y_SoftwareList.**"

# or
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,c8y_SoftwareList.*.*"
```

**Output**

```json
{
  "c8y_SoftwareList": [
    {
      "name": "app1",
      "url": "https://myexample.com/packages/app1/1.0.0",
      "version": "1.0.0"
    }, 
    {
      "name": "app2",
      "url": "https://myexample.com/packages/app2/2.0.0",
      "version": "2.0.0"
    }
  ],
  "id": "396806"
}
```

##### Only select non-object/array properties

```bash
c8y devices list --select "*"
```

**Output**

```json
{
  "creationTime": "2021-02-20T10:49:28.308Z",
  "id": "396806",
  "lastUpdated": "2021-02-20T10:49:28.308Z",
  "name": "1",
  "owner": "ciuser01",
  "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806",
  "type": "c8y_MacOS"
}
```

##### Include all properties using a globstar (`**`)

```bash
c8y devices list --select "**"
```

**Output**

```json
{
  "additionParents": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/additionParents"
  },
  "agent_Details": {
    "country": {
      "code": "61",
      "name": "AU"
    },
    "details": [1, 2, 3]
  },
  "assetParents": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/assetParents"
  },
  "c8y_Hardware": {
    "serialNumber": "XYDA001"
  },
  "c8y_SoftwareList": [
    {
      "name": "app1",
      "url": "https://myexample.com/packages/app1/1.0.0",
      "version": "1.0.0"
    }, 
    {
      "name": "app2",
      "url": "https://myexample.com/packages/app2/2.0.0",
      "version": "2.0.0"
    }
  ],
  "childAdditions": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/childAdditions"
  },
  "childAssets": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/childAssets"
  },
  "childDevices": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/childDevices"
  },
  "creationTime": "2021-02-20T10:49:28.308Z",
  "deviceParents": {
    "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806/deviceParents"
  },
  "id": "396806",
  "lastUpdated": "2021-02-20T10:49:28.308Z",
  "name": "1",
  "owner": "ciuser01",
  "self": "https://t12345.example.c8y.com/inventory/managedObjects/396806",
  "type": "c8y_MacOS"
}
```

##### Flatten nested properties

```bash
c8y devices list --select "id,name,type,c8y_Software*.**" --flatten
```

**Output**

```json
{
  "c8y_SoftwareList.0.name": "app1",
  "c8y_SoftwareList.0.url": "https://myexample.com/packages/app1/1.0.0",
  "c8y_SoftwareList.0.version": "1.0.0",
  "c8y_SoftwareList.1.name": "app2",
  "c8y_SoftwareList.1.url": "https://myexample.com/packages/app2/2.0.0",
  "c8y_SoftwareList.1.version": "2.0.0",
  "id": "396806",
  "name": "1",
  "type": "c8y_MacOS"
}
```

The same data can also be returned as csv using the `--output csv` and `--output csvheader` options.


```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,name,type,c8y_Software*.**" --output csvheader
```

```csv
id,name,type,c8y_SoftwareList.0.name,c8y_SoftwareList.0.url,c8y_SoftwareList.0.version,c8y_SoftwareList.1.name,c8y_SoftwareList.1.url,c8y_SoftwareList.1.version
396806,1,c8y_MacOS,app1,https://myexample.com/packages/app1/1.0.0,1.0.0,app2,https://myexample.com/packages/app2/2.0.0,2.0.0
```

Or if you only want to return the `name` and `version` of each software item.

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,*software*.**.name,*software*.**.version" --output csvheader
```

```csv
id,c8y_SoftwareList.0.name,c8y_SoftwareList.1.name,c8y_SoftwareList.0.version,c8y_SoftwareList.1.version
396806,app1,app2,1.0.0,2.0.0
```

Or if you only want the first software package from each device

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "id,*software*.0.name,*software*.0.version" --output csvheader
```

```csv
id,c8y_SoftwareList.0.name,c8y_SoftwareList.0.version
396806,app1,1.0.0
```


##### Output as csv

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select "id,name" --output csv
```

**Output**

```csv
396806,1
396735,10
```

##### Output as csv with headers

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select "id,nam*" --output csvheader
```

**Output**

```csv
id,name
396806,1
396735,10
```

### Reshaping data using custom names

The output json can also be reshaped by adding a name before the dot notation path in the format of:
`<alias>:<path>`

For example, the following example shows how to map the following properties

* `id` to `deviceId`
* `name` to `deviceName`

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select "deviceId:id,deviceName:name"
```

**Output**

```json
{
  "deviceId": "396806",
  "deviceName": "1"
}
{
  "deviceId": "396735",
  "deviceName": "10"
}
```

Mapping objects works the same way, though you need to specify that you want the full objects by using a globstar, otherwise you will only return the last value (which is hard to predict)

#### Renaming a root fragment

Renaming the root fragment `agent_Details` to `agent` in the output:

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "agent:agent_Details.**"
```

**Output**

```json
{
  "agent": {
    "country": {
      "code": "61",
      "name": "AU"
    },
    "details": [1, 2, 3]
  }
}
```

#### Expanding nested properties using globstar

Mapping properties using globstar `**` is done differently as the globstar can return multiple values if present in the json response.

There are two special cases are listed below:

##### Case 1: Using globstar `**` at the beginning of the dot notation path.

If the dot notation path starts with `**.` it means that it will map every matching property (regardless where it is) to a nested property under the given alias.

So, for example, if you want to move the whole `agent_Details` fragment and all nested properties to a nested fragment, then you can use:

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select "info:**.agent_Details.**"
```

**Output**

```json
{
  "info": {
    "agent_Details": {
      "country": {
        "code": "61",
        "name": "AU"
      },
      "details": [1, 2, 3]
    }
  }
}
```


##### Case 2: Use globstar at the end of the path

Use a globstar at the end of the path renames the root fragment. The following example renames the `agent_Details` fragment to `info`.

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select "info:agent_Details.**"
```

**Output**

```json
{
  "info": {
    "country": {
      "code": "61",
      "name": "AU"
    },
    "details": [1, 2, 3]
  }
}
```

You can also use wildcards to move a nested property. The following maps all of the literal properties from `agent_Details.country` to `country`: 

```bash
c8y devices list --pageSize 1 --fragmentType company_Example --select "country:agent_Details.country.*"
```

**Output**

```json
{
  "country": {
    "code": "61",
    "name": "AU"
  }
}
```

#### Get a list of ids and save it to file

```bash
c8y devices list --pageSize 2 --fragmentType company_Example --select id --output csv > devices.csv
```

**File contents: devices.csv**

```bash
396806
396735
```

##### Get the application id by looking it up by its name

```bash
c8y applications get --id cockpit --select id --output csv
```

**Response**

```plaintext
7
```

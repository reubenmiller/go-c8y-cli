---
category: Concepts - Extensions - API based commands
title: Parameter types
---

Parameter (or flag) types are used to defined how a flag's value should be interpreted. The supported parameter types which can be referenced from the API specification are listed below.

## Introduction

Flags can be added to a command by placing a flag definition under any of the following sections:

* `pathParameters`
* `headerParameters`
* `queryParameters`
* `body`

A flag definition provides documentation about the flag, and how the flag should be interpreted. The definition not only contains how the value should be interpreted but also where the value should be written to when the API request is being generated.

For example, a minimal flag definition under the `queryParameters` section could look like this:

```yaml
queryParameters:
- name: name
  type: string
```

The following table shows how the parameter would be translated to a query parameter used in the outgoing HTTP request.

|Flag|Translated QueryParameter|
|----|----|
|`--name example`|`?name=example`|

Sometimes the name of the flag might be different to the corresponding query parameter name. This is an important usability aspect, as commands can provide a contextualized abstraction on top of the REST API, so a flag's meaning can be slighly different to the API's meaning. In this case a custom mapping can be provided by using the `property` field.

Extending the previous snippet, the `name` flag can be changed to write to write to the `fragment` query parameter instead of the `name`.

```yaml
queryParameters:
- name: name
  type: string
  property: fragment
```

Now using the value provided via the `--name` flag will be assigned to the `fragment` query parameter.

|Flag|QueryParameter|
|----|----|
|`--name example`|`?fragment=example`|

The same mapping principle can be applied to any of the other parameters (e.g. body, pathParameters, headerParameters etc.).


## Basic types

The following basic types are available for use.

### Boolean

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`boolean`|Boolean value| `--enable` |`true` or `false`|
|`booleanDefault`|Boolean value with default| `--enable` |`true` or `false` (depending on the `.default` value)|
|`optional_fragment`|Add optional fragment (empty json object)| `--enable` |`{}`|

### Date / Time

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`datetime`|Relative or fixed date/time string| `--dateFrom -10d` |`"2023-04-27T22:52:06.622+02:00"`|
|`date`|Relative or fixed time string| `--dateFrom -10d` |`"2023-04-27"`|

### Numbers (integer/float)

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`integer`|Integer value| `--value 42` |`42`|
|`float`|Float value| `--value 42.1` |`42.1`|

### String

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`string`|String value| `--item "text value"` |`"text value"`|
|`stringStatic`|Fixed string which is always added (set by `.value`)|N/A|`"foobar"`|
|`string[]`|List of strings|`--item one --item two`|`["one", "two"]`|
|`stringcsv[]`|List of strings as csv list|`--item one --item two`|`"one,two"`|


### File based

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`file`|File upload with optional meta data (Multipart FormData request)| `--file ./foobar.txt` |`<<raw file contents>>`|
|`fileContents`|File contents (for binary uploads)| `--file ./foobar.txt` |`<<raw file contents>>`|
|`fileContentsAsString`|File contents as a string (for usage in a json body)| `--file ./foobar.txt` |`"file":"<<json escaped file contents>>"`|
|`attachment`|File upload without optional meta data (Multipart FormData request)| `--file ./foobar.txt` |`<<raw file contents>>`|
|`binaryUploadURL`|Upload file as Inventory binary and return the URL| `--file ./foobar.txt` |`"https://{host}/inventory/binaries/12345"`|


### JSON

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`json_custom`|JSON shorthand (or json string)| `--mydata "foo.bar=true"` |`{"mydata":{"foo":{"bar":true}}}`|


## Cumulocity specific types

This section contains Cumulocity IoT specific types which most involve both tab completion and/or lookups depending on the exact type.

### Applications

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`application`|Application|`--application administration`|`"12"`|
|`applicationname`|Application name|`--application cockpit`|`"cockpit"`|
|`hostedapplication`|Hosted application (e.g. type=`HOSTED`)|`--application devicemanagement`|`"12"`|
|`microservice`|Microservice (type=`MICROSERVICE`)|`--microservice report-agent`|`"8"`|
|`microservicename`|Microservice name|`--microservice report-agent`|`"report-agent"`|
|`microserviceinstance`|Microservice instance (completion only) (use `.dependsOn` field to specify related microservice|`--id advanced-software-mgmt --instance <TAB><TAB>`|`"advanced-software-mgmt-scope-management-deployment-7597ddb65lj6"`|

### Devices / Agents / Sources

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`source`|Source (e.g. event, alarm, measurement)| `--source 12345` |`"12345"`|
|`id[]`|List of ids| `--id 1 --id 2` |`"12345"`|
|`agent[]`|List of agents| `--agent 1 --id 2` |`"1", "2"`|
|`device[]`|List of devices| `--device 1 --id 2` |`"1", "2"`|

### Device Groups

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`devicegroup[]`|Device group| `--group "germany"` |`"12345"`|
|`smartgroup[]`|Smart group| `--group "australia"` |`"12345"`|


### Device management

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`deviceservice[]`|Device service| `--device 12345 --service my` (use `.dependsOn` to set which flag provides the device) to be set |`"abcdefg"`|

### Device requests

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`devicerequest`|Device request|`--id abcdef`|`"abcdef"`|
|`devicerequest[]`|Device request array|`--id abcdef`|`["abcdef"]`|


### Misc.

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`certificate[]`|Trusted device certificate| `--cert 4fd8df0378f2cafd5e322c1aaa8b87300704e9a5` |`"4fd8df0378f2cafd5e322c1aaa8b87300704e9a5"`|
|`certificatefile`|Certificate file| `--certfile ./my.cert` |`<<contents of file>>`|


### Notifications2

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`subscriptionId`|Subscription id|`--id abcdef`|`"abcdef"`|
|`subscriptionName`|Subscription name|`--name abcdef`|`"abcdef"`|


### Repository

#### Configuration
|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`configuration[]`|Configuration repository item| - | `--config linux_conf` |`"12345"`|
|`configurationDetails`|Configuration repository item details| - | `--config linux_conf` |`{"configuration": {"name": "example-config","type": "agentConfig","url": "https://test.com/content/raw/app.json"}}`|

#### Device profile

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`deviceprofile[]`|Device profile| - | `--profile bundle-deployment` |`"12345"`|


#### Firmware

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`firmware[]`|Firmware repository item| - | `--firmware ubuntu-22_04` |`"12345"`|
|`firmwarename`|Firmware name| - | `--firmware ubuntu-22_04` |`"ubuntu-22_04"`|
|`firmwareversion[]`|Firmware version| `firmware` | `--version 1.0.0` |`"12345"`|
|`firmwareversionName`|Firmware version name| `firmware` | `--version 2.0.0` |`"2.0.0"`|
|`firmwarepatch[]`|Firmware patch| `firmware` | `--patch 1.0.1` |`"12345"`|
|`firmwarepatchName`|Firmware patch name| `firmware` | `--patch 1.0.1` |`"1.0.1"`|
|`firmwareDetails`|Firmware version details| `firmware` | `--firmware "ubuntu-22_04" --version "1.0.1"` |`{"version":{"name":"ubuntu-22_04","version":"1.0.1","url":"https://example.com"}}`|

#### Software

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`software[]`|Software| - | `--software vim` |`"12345"`|
|`softwareName`|Software name| - | `--software vim` |`"vim"`|
|`softwareDetails`|Software version details| `software` | `--software vim --version ""` |`"12345"`|
|`softwareversion[]`|Software version| `software` | `--software vim --version 1.0.0` |`"98765"`|
|`softwareversionName`|Software version name| `software` | `--software vim --version 1.0.0` |`"1.0.0"`|

### Tenant

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`tenant`|Tenant id| `--tenant t12345` |`"t12345"`|
|`tenantname`|Tenant name| `--tenant mytenant` |`"mytenant"`|


### Users / User groups / Roles

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`role[]`|User role| `--role ROLE_ALARM_*` |`"ROLE_ALARM_READ"`|
|`roleself[]`|User role url| `--role ROLE_ALARM_*` |`"https://{host}/user/roles/ROLE_ALARM_READ"`|
|`user[]`|User id| `--user john*` |`"john.smith@example.com"`|
|`userself[]`|User self url| `--user john*` |`"https://{host}/user/{tenant}/users/john.smith@example.com"`|
|`usergroup[]`|User group| `--usergroup admins` |`"2"`|

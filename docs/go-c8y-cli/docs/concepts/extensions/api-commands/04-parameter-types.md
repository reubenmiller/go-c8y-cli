---
category: Tutorials - Extensions
title: Parameter types
---

Parameter (or flag) types are used to defined how a flag's value should be interpreted. The supported parameter types which can be referenced from the API specification are listed below.

:::caution
This page is not finished yet, so don't bother reading it just yet.
:::

# Available types

## Basic types

### Boolean

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`boolean`|Boolean value| `--enable` |`true` or `false`|
|`booleanDefault`|Boolean value with default| `--enable` |`true` or `false` (depending on the `.value`)|
|`optional_fragment`|Add optional fragment| `--enable` |`"fragment":{}` (fragment is defined by the `.value` property|

### Dates

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`datetime`|Relative or fixed date/time string| `--dateFrom -10d` |`"2023-04-27T22:52:06.622+02:00"`|
|`date`|Relative or fixed time string| `--dateFrom -10d` |`"2023-04-27"`|

### Numbers (integer/float)

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`integer`|Integer value| `--item 42` |`42`|
|`float`|Float value| `--item 42.1` |`42.1`|

### String

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`string`|String value| `--item "text value"` |`"text value"`|
|`stringStatic`|Fixed string which is always added|N/A|`"foobar"`|
|`string[]`|List of strings|`--item one --item two`|`["one", "two"]`|
|`stringcsv[]`|List of strings as csv list|`--item one --item two`|`"one,two"`|


### File based

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`file`|File| `--file ./foobar.txt` |`<<raw text contents>>`|
|`fileContents`|File contents| `--file ./foobar.txt` |`<<raw text contents>>`|
|`fileContentsAsString`|File contents as a string (for usage in a json body)| `--file ./foobar.txt` |`"name":"{{json escaped file contents}}"`|
|`attachment`|File as an attachment| `--file ./foobar.txt` |`<<raw text contents>>`|
|`binaryUploadURL`|Upload file as Inventory binary and return the URL| `--file ./foobar.txt` |`https://{host}/inventory/binaries/12345`|


### JSON

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`json`|JSON or json short form| `--mydata "foo.bar=true"` |`{"foo":{"bar":true}}`|
|`json_custom`|Custom json body| `--mydata "foo=bar"` |`{"foo":{"bar":true}}`|


## Cumulocity specific types

### Application

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`application`|Application|`--application administration`|`"12"`|
|`applicationname`|Application name|`--application cockpit`|`"cockpit"`|
|`hostedapplication`|Hosted application (e.g. type=`HOSTED`)|`--application devicemanagement`|`"12"`|
|`microservice`|Microservice (type=`MICROSERVICE`)|`--microservice report-agent`|`"8"`|
|`microservicename`|Microservice name|`--microservice report-agent`|`"report-agent"`|
|`microserviceinstance`|Microservice instance|`--id advanced-software-mgmt --instance <TAB><TAB>`|`"advanced-software-mgmt-scope-management-deployment-7597ddb65lj6"`|


### Devices / Agents / Sources

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`source`|Source (e.g. event, alarm, measurement)| `--source 12345` |`"12345"`|
|`id[]`|List of ids| `--id 1 --id 2` |`"12345"`|
|`agent[]`|List of agents| `--agent 1 --id 2` |`"1", "2"`|
|`device[]`|List of devices| `--device 1 --id 2` |`"1", "2"`|


### Inventory Query API

queryExpression


### Tenant

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`tenant`|Tenant id| `--tenant t12345` |`"t12345"`|
|`tenantname`|Tenant name| `--tenant mytenant` |`"mytenant"`|


### Misc

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`certificate[]`|Trusted device certificate| `--cert abcdefg` |`"abcdefg"`|
|`certificatefile`|Certificate file| `--certfile ./my.cert` |`<<contents of file>>`|


inventoryChildType


### Device management

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`deviceservice[]`|Device service| `--service my` (Also requires `--device 12345`) to be set |`"abcdefg"`|

**TODO**: Add example which shows the dependsOn usage


### Repository

#### Configuration
|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`configuration[]`|Configuration repository item| - | `--config linux_conf` |`"12345"`|
|`configurationDetails`|Configuration repository item details| - | `--config linux_conf` |`TODO`|

#### Device profile

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`deviceprofile[]`|Device profile| - | `--profile bundle-deployment` |`TODO`|


#### Firmware

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`firmware[]`|Firmware repository item| - | `--firmware ubuntu-22_04` |`TODO`|
|`firmwarename`|Firmware name| - | `--firmware ubuntu-22_04` |`TODO`|
|`firmwareversion[]`|Firmware version| - | `--version 1.0.0` |`TODO`|
|`firmwarepatch[]`|Firmware patch| - | `--patch 1.0.1` |`TODO`|
|`firmwarepatchName`|Firmware patch name| - | `--patch 1.0.1` |`TODO`|
|`firmwareVersionName`|Firmware version name| - | `--version 1.0.1` |`TODO`|
|`firmwareDetails`|Firmware details| - | `--firmware 1.0.1 ` |`TODO`|

#### Software

|Type|Description|Depends On|Example usage|Example output|
|----|----|----|----|----|
|`software[]`|Software| - | `--software vim` |`"value"`|
|`softwareName`|Software name| - | `--software vim` |`"value"`|
|`softwareDetails`|Software details| - | `--software vim` |`"value"`|
|`softwareversion[]`|Software version| - | `--version 1.0.0` |`""`|
|`softwareversionName`|Software version name| - | `--version 1.0.0` |`""`|


### Device requests

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`devicerequest[]`|Device request|`--id abcdef`|`["abcdef"]`|

### Notifications 2

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`subscriptionId`|Subscription id|`--id abcdef`|`"abcdef"`|
|`subscriptionName`|Subscription name|`--name abcdef`|`"abcdef"`|

### Devices Groups

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`devicegroup[]`|Device group| `--group "germany"` |`"12345"`|
|`smartgroup[]`|Smart group| `--group "australia"` |`"12345"`|


### Users / User groups / Roles

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`role[]`|User role| `--role ROLE_ALARM_*` |`"ROLE_ALARM_READ"`|
|`roleself[]`|User role url| `--role ROLE_ALARM_*` |`"https://{host}/user/roles/ROLE_ALARM_READ"`|
|`user[]`|User id| `--user john*` |`"john.smith@example.com"`|
|`userself[]`|User self url| `--user john*` |`"https://{host}/user/{tenant}/users/john.smith@example.com"`|
|`usergroup[]`|User group| `--usergroup admins` |`"2"`|

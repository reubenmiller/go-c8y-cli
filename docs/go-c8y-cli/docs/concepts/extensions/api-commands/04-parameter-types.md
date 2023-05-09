---
category: Tutorials - Extensions
title: Parameter types
---

Parameter (or flag) types are used to defined how a flag's value should be interpreted. The supported parameter types which can be referenced from the API specification are listed below.

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
|`[]string`|List of strings|`--item one --item two`|`["one", "two"]`|
|`[]stringcsv`|List of strings as csv list|`--item one --item two`|`"one,two"`|


### File based

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`file`|File| `--file ./foobar.txt` |*raw text contents*|
|`fileContents`|File contents| `--file ./foobar.txt` |*raw text contents*|
|`attachment`|File as an attachment| `--file ./foobar.txt` |*raw text contents*|
|`binaryUploadURL`|Upload file as Inventory binary and return the URL| `--file ./foobar.txt` |`https://{host}/inventory/binaries/12345`|


### JSON

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`json`|JSON or json short form| `--mydata "foo.bar=true"` |`{"foo":{"bar":true}}`|
|`json_custom`|Custom json body| `--mydata "foo=bar"` |`{"foo":{"bar":true}}`|


## Cumulocity specific types

### Application

|Type|Description|Example|Completion|Lookup|
|----|----|----|----|
|`application`|Application|`12`|:white_check_mark:|:white_check_mark:|
|`applicationname`|Application name|`cockpit`|:white_check_mark:|:black_square_button:|
|`hostedapplication`|Hosted application (e.g. type=`HOSTED`)|`devicemanagement`|:white_check_mark:|:white_check_mark:|
|`microservice`|Microservice (type=`MICROSERVICE`)|`8`|:white_check_mark:|:white_check_mark:|
|`microservicename`|Microservice name|`report-agent`|:white_check_mark:|:black_square_button:|
|`microserviceinstance`|Microservice instance|``|:white_check_mark:|:black_square_button:|



### Devices / Agents / Sources

|Type|Description|Example usage|Example output|
|----|----|----|----|
|`source`|Source (e.g. event, alarm, measurement)| `--source 12345` |"12345"|
|`[]id`|List of ids| `--id 1 --id 2` |"12345"|
|`[]agent`|List of agents| `--agent 1 --id 2` |"1", "2"|
|`[]device`|List of devices| `--device 1 --id 2` |"1", "2"|


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
|`[]certificate`|Trusted device certificate| `--cert abcdefg` |"abcdefg"|
|`certificatefile`|Certificate file| `--certfile ./my.cert` |*Contents of certificate file*|


inventoryChildType


### Device management

[]deviceservice


### Repository

[]configuration
configurationDetails
[]deviceprofile
[]firmware
[]firmwareversion
[]firmwarepatch
firmwareName
firmwareVersionName
firmwareDetails
firmwarepatchName
[]software
softwareDetails
softwareName
[]softwareversion
softwareversionName



### Device requests

[]devicerequest

### Notifications 2

subscriptionId
subscriptionName

### Devices Groups

[]devicegroup
[]smartgroup


### Users / Roles

[]usergroup
[]userself
[]roleself
[]role
[]user

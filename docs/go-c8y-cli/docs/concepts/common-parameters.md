---
layout: default
category: Concepts
title: Common Parameters
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


### Verbose

Display verbose information when sending a command to Cumulocity. The verbose messages help to understand what was sent to Cumulocity and what was received as well with other verbose messages related to the cli tool itself.

**Example: Display detailed logging about the command**

<Tabs
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y devices list --verbose
```

</TabItem>
<TabItem value="powershell">

```powershell
$resp = Get-DeviceCollection -Verbose
```

</TabItem>
</Tabs>

*Output*

```powershell
VERBOSE: Performing the operation "Get device collection" on target "t12345678".
VERBOSE: C:\Users\xxxxx\Documents\WindowsPowerShell\Modules\PSc8y\1.1.0\Dependencies\c8y.windows.exe devices list --pretty=false --verbose --raw
2020/05/01 10:37:06 Using session environment variable: C:\Users\xxxxx\.cumulocity\session.json
2020/05/01 10:37:06 Using existing env variables. HTTP_PROXY [], http_proxy [], HTTPS_PROXY [], https_proxy [], NO_PROXY [], no_proxy []
2020/05/01 10:37:06 Using config file: C:\Users\xxxxx\.cumulocity\session.json
2020/05/01 10:37:06 Use tenant prefix: true
2020/05/01 10:37:06 Current username: t12345678/myusername
2020/05/01 10:37:06 Headers: map[Accept:[application/json] Authorization:[Basic dDEyMzQ1Njc4L215dXNlcm5hbWU6bmljZXRyeQo=] User-Agent:[go-client] X-Application:[go-client]]
2020/05/01 10:37:06 Sending request: GET https://goc8ycli-example.eu-latest.cumulocity.com/inventory/managedObjects?query=$filter=%28has%28c8y_IsDevice%29+or+has%28c8y_ModbusDevice%29%29+$orderby=name
2020/05/01 10:37:06 Body: null
2020/05/01 10:37:06 Status code: 200
VERBOSE: Statistics: currentPage=1, pageSize=5, totalPages=
```

### Dry

`--dry` displays the REST request that would have been sent if `-Dry` was not used. This option does NOT send a request to Cumulocity, it just displays the contents of the request on the console. This is useful if you want to see exact REST request and body being sent so you can copy it for use in other applications.

**Example: Check what request would be sent when using New-Device without creating the device**

<Tabs
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y devices create --name "my-test" --dry
```

</TabItem>
<TabItem value="powershell">

```powershell
New-Device -Name my-test -Dry
```

</TabItem>
</Tabs>

*Output*

```powershell
2020/05/01 10:54:16 Using session environment variable: C:\Users\******\.cumulocity\example-session1.json
2020/05/01 10:54:16 Using existing env variables. HTTP_PROXY [], http_proxy [], HTTPS_PROXY [], https_proxy [], NO_PROXY [], no_proxy []
2020/05/01 10:54:16 Using config file: C:\Users\******\.cumulocity\example-session1.json
2020/05/01 10:54:16 Use tenant prefix: true
2020/05/01 10:54:16 Input:
2020/05/01 10:54:16 Current username: {tenant}/{username}
2020/05/01 10:54:16 What If: Sending [POST] request to [https://{tenant}.eu-latest.cumulocity.com/inventory/managedObjects]

Headers:
User-Agent: go-client
X-Application: go-client
Content-Type: application/json
Authorization: Basic  {base64 tenant/username:password}
Accept: application/json

Body:
{
   "c8y_IsDevice": {},
   "name": "my-test"
 }
2020/05/01 10:54:16 Response time: 2ms
```


### PageSize

Set the maximum number of results for commands which return a list of objects. There is usually an upper limit enforced by Cumulocity which is dependant on the Cumulocity object type (i.e. Events, Operations, Managed Objects etc.). By default it is set to `5`.


**Example: Get a collection of alarms and limit the results to 100 alarms**

```powershell
Get-AlarmCollection -PageSize 100
```

### WithTotalPages

In the `PSc8y` module, the `-WithTotalPages` switch also changes the view to easily view the page statistics to see how many total pages exist.

In order to get an accurate total of entities, it is recommended to use `-PageSize 1` along with `-WithTotalPages`. The `totalPages` property will then display the total number of entities.

```powershell
Get-AlarmCollection -PageSize 1 -WithTotalPages
```

*Output*

```powershell

    self: https://mytenant.xxxxx.cumulocity.com/alarm/alarms?withTotalPages=true&pageSize=1&currentPage=1
    next: https://mytenant.xxxxx.cumulocity.com/alarm/alarms?withTotalPages=true&pageSize=1&currentPage=2
currentPage     pageSize        totalPages      alarms
-----------     --------        ----------      ------
1               1               44              {@{severity=MAJOR; creationTime=12/23/2019 18:58:46...
```

### Raw

The `Raw` option tells the cli tool to return the Cumulocity response without an pre-processing, so that the response including the extra meta properties like `next`, `self` and `statistics`. This is useful if you want to process the original response from Cumulocity yourself, rather than only returning the array items.

This concept is applied to all entities (i.e. alarms, events, managed objects and operations etc.)

**Example: Get a list of alarms**

When `Get-AlarmCollection` is used without the `-Raw` option, only the array containing the alarms is returned (i.e. `.alarms`). All of the other properties like `statistics` are removed. This is normally the preferred behavior because you are normally only interested in the actual alarms and not any additional meta properties. This reduces the code required when piping the alarm list to another function.

So getting a list of alarms without `-Raw` looks like this:

```powershell
Get-AlarmCollection -PageSize 1 | tojson
```

*Output*

```json
{
  "severity": "MAJOR",
  "creationTime": "2020-01-26T14:21:07.325Z",
  "count": 1,
  "history": {
    "auditRecords": [],
    "self": "https://{tenant}.eu-latest.cumulocity.com/audit/auditRecords"
  },
  "source": {
    "name": "TestDevicegIvTYAuwPz",
    "self": "https://{tenant}.eu-latest.cumulocity.com/inventory/managedObjects/131166",
    "id": "131166"
  },
  "type": "testType",
  "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms/130877",
  "time": "2020-01-26T14:21:06.921Z",
  "text": "Custom Event 1",
  "id": "130877",
  "status": "ACTIVE"
}
```

Here is the same command but using the `-Raw` option.

```bash
Get-AlarmCollection -PageSize 1 -Raw | tojson
```

*Output*

```json
{
  "next": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms?pageSize=1&currentPage=2",
  "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms?pageSize=1&currentPage=1",
  "statistics": {
    "totalPages": 70,
    "currentPage": 1,
    "pageSize": 1
  },
  "alarms": [
    {
      "severity": "MAJOR",
      "creationTime": "2020-01-26T14:21:07.325Z",
      "count": 1,
      "history": "@{auditRecords=System.Object[]; self=https://{tenant}.eu-latest.cumulocity.com/audit/auditRecords}",
      "source": "@{name=TestDevicegIvTYAuwPz; self=https://{tenant}.eu-latest.cumulocity.com/inventory/managedObjects/131166; id=131166}",
      "type": "testType",
      "self": "https://{tenant}.eu-latest.cumulocity.com/alarm/alarms/130877",
      "time": "2020-01-26T14:21:06.921Z",
      "text": "Custom Event 1",
      "id": "130877",
      "status": "ACTIVE"
    }
  ]
}
```

### OutputFile

The output file will be the raw response as returned from the Cumulocity platform.

```powershell
Get-AlarmCollection -PageSize 1 -OutputFile test.json
```

*Output*

```powershell
/Users/Shared/demo/test.json
```

Or it can be used to download a binary from the platform

c8y_applications_storage_9397

```powershell
Find-ManagedObjectCollection -Query "has(c8y_IsBinary) and type eq 'c8y_applications_storage_*'" -PageSize 1 |
    Get-Binary -OutputFile my.zip
```


### NoProxy

`NoProxy` ignores all of proxy settings. Helpful when trying to diagnose proxy problems, or you need to ignore the `HTTPS_PROXY` environment variables.

**Example: Ignore any proxy environment variables**

`Get-DeviceCollection` will not use the proxy settings when using `-NoProxy`.

```powershell
$env:HTTPS_PROXY = "http://10.0.0.1:8000"
Get-DeviceCollection -NoProxy
```

If you are in doubt about which proxy settings are being used when running a command then add `-Verbose` to turn on the verbose logging, and the proxy/noproxy settings will be displayed on the console.

### Session

Session enable to use another c8y session profile for a single command without having to use `Set-Session`.

This is useful if you want to quickly see what is going on in another tenant without permanently changing your current session, or you want to transfer some information from a staging environment to a dev as a once-off action.

**Example: Get a device name from the staging tenant, and create a new device with the same name in a dev tenant**

```powershell
$Device = Get-OperationCollection -PageSize 1 -Session staging-tenant

New-TestDevice -Name $Device.name -Session dev-tenant
```

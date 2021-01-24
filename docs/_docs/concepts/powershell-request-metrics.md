---
layout: default
category: Concepts
title: Powershell - Request Metrics
---

The PSc8y PowerShell modules provides some additional functionality that is not available when using Bash or zsh. PSc8y allows the user to access additional meta information about the command and the underlying request by providing it in different PowerShell streams.

The following sections details how to:
* Save the WhatIf (dry run) output to a variable
* Get meta information about the request such as response time, status code, request url, response headers etc.

These capabilities are possible via the use of PowerShell's common parameters, `InformationVariable` and `ErrorVariable`, which can be used to redirect information to variables or output streams.

The following links detail PowerShell's concepts regarding common parameters and stream redirection.
* [PowerShell Common Parameters](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_commonparameters)
* [PowerShell Redirection](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_redirection)


### Storing request information when using WhatIf

Information about requests including the request url, headers and body can be written to a variable for later use.

The usage of the `WhatIf` parameter means that the command will not actually send REST request to Cumulocity, but you can inspect it or use it as some form of API documentation.

##### Example: Redirect request information to a variable

```powershell
New-ManagedObject -Name "my custom data" -WhatIf -InformationVariable requestInfo

# Show the request data
$requestInfo
```

*Output*
```text
What If: Sending [POST] request to [https://example123.my-c8y.com/inventory/managedObjects]

Headers:
Accept: application/json
Authorization: Basic asdfasfd........
Content-Type: application/json
User-Agent: go-client
X-Application: go-client

Body:
{
  "name": "my custom data"
}
```

##### Example: Redirect request information to file

Alternatively you can store the request information to file by redirecting the Information stream (stream 6) to a file called `myrequest.txt`.

```powershell
New-ManagedObject -Name "my custom data" -WhatIf 6> myrequest.txt

# Read the myrequest.txt file
Get-Content myrequest.txt
```

##### Example: Store request from a custom api call

`Invoke-ClientRequest` also supports the `InformationVariable` parameter for use in custom scripts and modules.

```powershell
Invoke-ClientRequest `
  -Uri "/alarm/alarms" `
  -QueryParameters @{
    pageSize = "1";
  } `
  -Whatif `
  -InformationAction SilentlyContinue `
  -InformationVariable requestInfo

# Show request
$requestInfo
```

### Storing information about the response

When the `WhatIf` parameter is not being used, the InformationVariable contains information about the request and response.

The following information is available:

* Request url
* Request header
* Response status code (http status code)
* Response time (in milliseconds)
* Response header
* Response length (in Kilobyte)

##### Example: Save response time in a variable

The response time of the request can be accessed here

```powershell
$null = Get-ManagedObjectCollection -Verbose -InformationAction continue -InformationVariable responseInfo
$time = $responseInfo.MessageData.responseTime
```

Or you can display all of the available information about the request and response using:

```powershell
$responseInfo.MessageData | ConvertTo-Json
```

*Output*

```json
{
  "statusCode": "200",
  "responseHeader": "map[Cache-Control:[no-cache,no-store,must-revalidate] Connection:[keep-alive] Content-Type:[application/vnd.com.nsn.cumulocity.managedobjectcollection+json;charset=UTF-8;ver=0.9] Date:[Sat, 23 Jan 2021 19:29:44 GMT] Expires:[-1] Pragma:[no-cache]]",
  "responseTime": "95ms",
  "responseLength": "9.0KB"
}
```

##### Example: Get response times for custom API requests

`Invoke-ClientRequest` also supports accessing the request metrics via the InformationVariable.

```powershell
$Response = Invoke-ClientRequest `
  -Uri "/alarm/alarms" `
  -QueryParameters @{
    pageSize = "1";
  } `
  -InformationVariable responseInfo

$time = $responseInfo.MessageData.responseTime
```

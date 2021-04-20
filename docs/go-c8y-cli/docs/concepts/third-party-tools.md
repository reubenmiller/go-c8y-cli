---
layout: default
category: Concepts
title: Usage with 3rd party tools
---


### Re-using sessions information in 3rd party tools like curl

When a session is activated (via the `set-session` helper), environment variables are set containing information about the current session. This information can be re-used for other 3rd party tools like curl.

The `C8Y_HOST` environment variable contains the current Cumulocity base url related to the selected session.

The `C8Y_HEADER` environment variable contains the full authorization header (include header name), which makes it ideal for using with tools like `curl`.

The `C8Y_HEADER_AUTHORIZATION` environment variable contains the authorization header value (without the header name), which makes it more suitable for usage within PowerShell, but it can also be used just as easily in curl.

The examples below will show how these environment variables can be used in both shell (bash/zsh) and PowerShell.

#### bash-zsh: Sending c8y api requests using curl

**Get current user details**

```bash
curl --silent -H "$C8Y_HEADER" $C8Y_HOST/user/currentUser | jq

# or the same request but using the C8Y_HEADER_AUTHORIZATION environment variable
curl --silent -H "Authorization: $C8Y_HEADER_AUTHORIZATION" $C8Y_HOST/user/currentUser | jq
```

A more generic approach would be to create a small shell function which adds the authorization header and prefixes the Cumulocity host url automatically to requests.

The following shows a simple bash function that achieves this:

```bash
#
# c8ycurl: Helper which adds c8y authorization header and host url
#
c8ycurl ()
{ 
    curl --silent -H "$C8Y_HEADER" "${C8Y_HOST%%/}/${1#/}" "${@:2}"
}
```

Below shows examples how to then use the helper function.

```bash
c8ycurl /user/currentUser | jq

# Get a list of operation
c8ycurl "inventory/managedObjects?pageSize=1&q=\$filter=name+eq+'*'" -H "Accept: application/json" | jq


# Create a managed object (using POST)
c8ycurl "inventory/managedObjects" -XPOST -H "Accept: application/json" -H "Content-Type: application/json" -d "{\"name\":\"device_name\"}"
```

#### PowerShell: Sending c8y api requests using native PowerShell

To send native PowerShell requests, the session environment variables can be re-used. 

```powershell
# Get current user
Invoke-RestMethod -Headers @{ Authorization = $env:C8Y_HEADER_AUTHORIZATION } -Uri "$env:C8Y_HOST/user/currentUser"
```

PowerShell also supports setting default parameter values globally without having to create a helper function. The following set the default headers argument.

```powershell
$PSDefaultParameterValues["Invoke-RestMethod:Headers"] = @{ Authorization = $env:C8Y_HEADER_AUTHORIZATION }

Invoke-RestMethod -Uri "$env:C8Y_HOST/user/currentUser"

Invoke-RestMethod -Uri "$env:C8Y_HOST/inventory/managedObjects"
Invoke-RestMethod -Uri "$env:C8Y_HOST/user/currentUser"
```

Alternatively a helper function can be created which will add the required information. It might seem verbose, but most of the logic is just for passing on arguments to the native PowerShell `Invoke-RestMethod` cmdlet. You can also control it however you want by setting sensible defaults. You could also split the helper function into multiple, i.e. one for POSTs and PUTs and one for GETs etc.

```powershell
Function Invoke-MyRequest {
    [cmdletbinding()]
    Param(
        [string] $Path,

        [string] $Method = "GET",

        [object] $Body,

        [object] $Accept = "application/json",

        [object] $ContentType = "application/json",

        [hashtable] $Headers,

        # Additional options that will be passed to Invoke-RestMethod
        [hashtable] $AdditionalOptions
    )
    $options = @{
        Uri = "$env:C8Y_HOST".TrimEnd("/") + "/" + "$Path".TrimStart("/")
        Method = $Method
        ContentType = $ContentType
        Headers = @{ Authorization = $env:C8Y_HEADER_AUTHORIZATION }
    }
    if ($Accept) {
        $options.Headers.Accept = $Accept
    }
    if ($Headers) {
        $options.Headers += $Headers
    }
    if ($Body) {
        if ($Body -is [hashtable]) {
            $options.Body = ConvertTo-Json $Body -Compress
        } else {
            $options.Body = $Body
        }
    }
    if ($AdditionalOptions) {
        $options += $AdditionalOptions
    }

    Invoke-RestMethod @options
}
```

The `Invoke-MyRequest` helper function can then be called using:

```powershell
# Get current user
Invoke-MyRequest "/user/currentUser"

# Create managed object
Invoke-MyRequest "/inventory/managedObjects" -Method POST -Body @{ name = "test device name" }
```

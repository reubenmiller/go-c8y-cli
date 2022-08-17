Function Invoke-ClientRequest {
    <#
.SYNOPSIS
Send a rest request using the c8y

.DESCRIPTION
Send a custom rest request to Cumulocity using all of the options found on other command lets.
This is useful if you are extending PSc8y and want to send custom microservice requests, or
send requests which are not yet provided in the PSc8y module.

Example:

The following function sends a POST request to predefined microservice endpoint.
It accepts an input Body argument which will be used in the request.

The response is also converted from raw json (string) to Powershell objects so that advanced
filtering can be done on the response (i.e. using `Where-Object`)

```powershell
Function Invoke-MyMicroserviceEndpoint {
    [cmdletbinding()]
    Param(
        [hashtable] $Body
    )

    $options = @{
        Method = "POST"
        Uri = "/service/mymicroservice"
        Data = $Body
    }

    # Send request
    $response = Invoke-ClientRequest @options
    
    # Convert response from json to Powershell objects
    ConvertFrom-Json $response
}
```

.LINK
c8y rest

.EXAMPLE
Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test"

Create a new managed object with the name "test"

.EXAMPLE
Invoke-ClientRequest -Uri "/alarm/alarms" -QueryParameters @{ pageSize = "100" }

Get a list of alarms with page size of 100

.EXAMPLE
Invoke-ClientRequest -Uri "/alarm/alarms?pageSize=100"

Get a list of alarms with page size of 100

.EXAMPLE
Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test" -Headers @{ Custom-Value = "myValue"}

Create a new managed object but add a custom accept header value
#>
    [cmdletbinding()]
    Param(
        # Uri (or partial uri). i.e. /application/applications
        [Parameter(
            Mandatory = $true,
            Position = 0)]
        [string] $Uri,

        # Rest Method. Defaults to GET
        [ValidateSet("GET", "POST", "DELETE", "PUT", "HEAD")]
        [Microsoft.PowerShell.Commands.WebRequestMethod] $Method = 'GET',

        # Add custom headers to the rest request
        [hashtable] $Headers,

        # Input file to be uploaded as FormData
        [string] $InFile,

        # Uri query parameters
        [hashtable] $QueryParameters,

        # (Body) Content Type
        [string] $ContentType,

        # Accept header
        [string] $Accept,

        # HostName to use which overrides the given host
        [string] $HostName
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory create" -Exclude @(
            "Uri", "Method", "Headers", "QueryParameters", "InFile", "HostName"
        )
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = if ($Accept) { $Accept } else { "application/json" }
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($null -ne $QueryParameters) {
            $queryparams = New-Object System.Collections.ArrayList
            foreach ($key in $QueryParameters.Keys) {
                $value = $QueryParameters[$key]
                if ($value) {
                    $null = $c8yargs.AddRange(@("--customQueryParam", "${key}=${value}"))
                    # $null = $queryparams.Add("${key}=${value}")
                }
            }

            if ($queryparams.Count -gt 0) {
                $str = $queryparams -join "&"
                if ($Uri.Contains("?")) {
                    # uri already has some query parameters, so just append the new one to it
                    $Uri = $Uri + "&" + $str
                }
                else {
                    $Uri = $Uri + "?" + $str
                }
            }
        }

        if ($Method) {
            $null = $c8yargs.Add($Method)
        }

        $null = $c8yargs.Add($Uri)

        if ($null -ne $Headers) {
            foreach ($key in $Headers.Keys) {
                $null = $c8yargs.AddRange(@("-H=`"{0}: {1}`"" -f $key, $Headers[$key]))
            }
        }

        if ($HostName) {
            $null = $c8yargs.AddRange(@("--host", $HostName))
        }

        if ($InFile) {
            $null = $c8yargs.AddRange(@("--file", $InFile))
        }

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y rest $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y rest $c8yargs
        }
    }
}

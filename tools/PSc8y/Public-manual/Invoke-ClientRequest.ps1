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
    [cmdletbinding(
        SupportsShouldProcess = $true
    )]
    Param(
        [hashtable] $Body
    )

    $options = @{
        Method = "POST"
        Uri = "/service/mymicroservice"
        Data = $Body

        # Add these to support -WhatIf and -Verbose parameters
        WhatIfPreference = $WhatIfPreference `
        VerbosePreference = $VerbosePreference
    }

    # Send request
    $response = Invoke-ClientRequest @options
    
    # Convert response from json to Powershell objects
    ConvertFrom-Json $response
}
```

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
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "None")]
    Param(
        # Uri (or partial uri). i.e. /application/applications
        [Parameter(
            Mandatory = $true,
            Position = 0)]
        [string] $Uri,

        # Rest Method. Defaults to GET
        [ValidateSet("GET", "POST", "DELETE", "PUT", "HEAD")]
        [Microsoft.PowerShell.Commands.WebRequestMethod] $Method = 'GET',

        # Request body
        [object] $Data,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

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

        # Ignore the accept header
        [switch]
        $IgnoreAcceptHeader,

        # Timeout in seconds
        [Parameter()]
        [double]
        $TimeoutSec,

        # Pretty print json response
        [Parameter()]
        [switch]
        $Pretty,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session,

        # HostName to use which overrides the given host
        [string] $HostName,

        # Allow loading Cumulocity session setting from environment variables
        [Parameter()]
        [switch]
        $UseEnvironment
    )

    $c8y = Get-ClientBinary

    $c8yargs = New-Object System.Collections.ArrayList

    $null = $c8yargs.Add("rest")

    if ($Method) {
        $null = $c8yargs.Add($Method)
    }

    if ($null -ne $QueryParameters) {
        $queryparams = New-Object System.Collections.ArrayList
        foreach ($key in $QueryParameters.Keys) {
            $value = $QueryParameters[$key]
            if ($value) {
                $null = $queryparams.Add("${key}=${value}")
            }
        }

        if ($queryparams.Count -gt 0) {
            $str = $queryparams -join "&"
            if ($Uri.Contains("?")) {
                # uri already has some query parameters, so just append the new one to it
                $Uri = $Uri + "&" + $str
            } else {
                $Uri = $Uri + "?" + $str
            }
        }
    }

    $null = $c8yargs.Add($Uri)

    if ($null -ne $Data -and ![string]::IsNullOrEmpty($Data)) {
        if ($Data -is [string]) {
            if (Test-Json -InputObject $Data -WarningAction SilentlyContinue) {
                $null = $c8yargs.AddRange(@("--data", (ConvertTo-JsonArgument $Data)))
            } else {
                # allow shortform strings (intepreted by c8y cli tool)
                $null = $c8yargs.AddRange(@("--data", $Data))
            }
        } else {
            # Convert hashtables, psobject etc.
            $null = $c8yargs.AddRange(@("--data", (ConvertTo-JsonArgument $Data)))
        }

    }

    if (-not [string]::IsNullOrEmpty($Template)) {
        $null = $c8yargs.AddRange(@("--template", $Template))
    }

    if (-not [string]::IsNullOrEmpty($TemplateVars)) {
        $null = $c8yargs.AddRange(@("--templateVars", $TemplateVars))
    }

    if ($null -ne $Headers) {
        foreach ($key in $Headers.Keys) {
            $null = $c8yargs.AddRange(@("-H=`"{0}: {1}`"" -f $key, $Headers[$key]))
        }
    }

    if (-not [string]::IsNullOrEmpty($ContentType)) {
        $null = $c8yargs.AddRange(@("--contentType", $ContentType))
    }

    if (-not [string]::IsNullOrEmpty($Accept)) {
        $null = $c8yargs.AddRange(@("--accept", $Accept))
    }

    if ($IgnoreAcceptHeader) {
        $null = $c8yargs.Add("--ignoreAcceptHeader")
    }

    if ($HostName) {
        $null = $c8yargs.AddRange(@("--host", $HostName))
    }

    if ($TimeoutSec) {
        # Convert to milliseconds (cast to an integer)
        [int] $TimeoutInMS = $TimeoutSec * 1000
        $null = $c8yargs.AddRange(@("--timeout", $TimeoutInMS))
    }

    if ($InFile) {
        $null = $c8yargs.AddRange(@("--file", $InFile))
    }

    if ($OutputFile) {
        $null = $c8yargs.AddRange(@("--outputFile", $OutputFile))
    }

    if ($Raw) {
        $null = $c8yargs.Add("--raw")
    }

    if ($Session) {
        $null = $c8yargs.AddRange(@("--session", $Session))
    }

    if ($UseEnvironment) {
        $null = $c8yargs.Add("--useEnv")
    }

    if ($NoProxy) {
        $null = $c8yargs.Add("--noProxy")
    }

    $null = $c8yargs.Add("--pretty={0}" -f $Pretty.ToString().ToLower())

    if ($VerbosePreference) {
        $null = $c8yargs.Add("--verbose")
    }

    if ($WhatIfPreference) {
        $null = $c8yargs.Add("--dry")
    }

    Write-Verbose ("{0} {1}" -f $c8y, ($c8yargs -join " "))

    & $c8y $c8yargs
}

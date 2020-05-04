Function Invoke-ClientRequest {
<#
.SYNOPSIS
Send a rest request using the c8y

.EXAMPLE
Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test"

Create a new managed object with the name "test"

.EXAMPLE
Invoke-ClientRequest -Uri "/alarm/alarms" -QueryParameters @{ pageSize = "100" }

Get a list of alarms with page size of 100

.EXAMPLE
Invoke-ClientRequest -Uri "/alarm/alarms?pageSize=100"

Get a list of alarms with page size of 100
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

        # HostName to use which overrides the given host
        [string] $HostName,

        # Rest Method. Defaults to GET
        [Microsoft.PowerShell.Commands.WebRequestMethod] $Method,

        # Request body
        [object] $Data,

        # Input file to be uploaded as FormData
        [string] $InFile,

        # Uri query parameters
        [hashtable] $QueryParameters,

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
        $Session
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

    if ($null -ne $Data) {
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

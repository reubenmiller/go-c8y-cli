Function Invoke-CustomRestMethod {
    <#
.SYNOPSIS
Create a new C8Y Rest Request given a specific API string. It uses information from the active c8y session to build the full api request object

.EXAMPLE
Invoke-CustomRestMethod /alarm/alarms

Manually invoke the REST method GET on the alarms rest endpoint, to return all alarms

.EXAMPLE
rest /alarm/alarms

Invoke the rest method using the alias for the Invoke-CustomRestMethod (useful when testing out new c8y commands)

.EXAMPLE
rest myapp://health

Get the health status of the myapp microservice
#>
    [CmdletBinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "None")]
    Param(
        # Uri
        [Parameter(
            Mandatory = $true,
            Position = 0)]
        [string]
        $Uri,

        # REST API Method. i.e. GET, POST, PUT etc.
        [Parameter(
            Mandatory = $false,
            Position = 1)]
        [Microsoft.PowerShell.Commands.WebRequestMethod]
        $Method = "GET",

        # Content Type. Defaults to application/json (with utf8 encoding). If no encoding is set, then utf8 will be used
        [string]
        $ContentType,

        # Accept header type
        [string]
        $Accept,

        [Parameter(Mandatory = $false)]
        [System.Collections.IDictionary]
        $Headers,

        [Parameter(Mandatory = $false)]
        [System.Object]
        $Body,

        # Ignore the accept header
        [switch]
        $IgnoreAcceptHeader,

        # Return the result as text
        [switch]
        $AsText,

        # Don't send requests to the proxy
        [switch]
        $NoProxy,

        # Use an explicit proxy setting (instead of environmnet variables)
        [string]
        $Proxy,

        [int]
        $TimeoutSec,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Compress returns minimized json
        [Parameter()]
        [switch]
        $Compress,

        # Allow loading Cumulocity session setting from environment variables
        [Parameter()]
        [switch]
        $UseEnvironment,

        # Session path
        [Parameter()]
        [string]
        $Session
    )

    Begin {
        # Control default timeout setting by using a environment variable
        if (!$PSBoundParameters.ContainsKey("TimeoutSec")) {
            if ($Env:__C8Y_REST_TIMEOUT) {
                $TimeoutSec = $Env:__C8Y_REST_TIMEOUT
            }
            else {
                $TimeoutSec = 600
            }
        }

        $c8ycli = Get-ClientBinary
    }

    Process {
        $args = New-Object System.Collections.ArrayList

        $null = $args.Add("rest")

        if (![string]::IsNullOrEmpty($Method)) {
            $null = $args.Add($Method)
        }
        if (![string]::IsNullOrEmpty($Uri)) {
            $null = $args.Add($Uri)
        }
        if (![string]::IsNullOrEmpty($ContentType)) {
            $null = $args.AddRange(@("--contentType", $ContentType))
        }
        if (![string]::IsNullOrEmpty($Accept)) {
            $null = $args.AddRange(@("--accept", $Accept))
        }
        if ($IgnoreAcceptHeader) {
            $null = $args.Add("--ignoreAcceptHeader")
        }
        if ($Raw) {
            $null = $args.Add("--raw")
        }
        if (![string]::IsNullOrEmpty($Session)) {
            $null = $args.AddRange(@("--session", $Session))
        }
        if ($Compress) {
            $null = $args.Add("--pretty=false")
        }
        if (![string]::IsNullOrEmpty($Proxy)) {
            $null = $args.AddRange(@("--proxy", $Proxy))
        }
        if ($NoProxy) {
            $null = $args.Add("--noProxy")
        }
        if ($VerbosePreference) {
            $null = $args.Add("--verbose")
        }
        if ($WhatIfPreference) {
            $null = $args.Add("--dry")
        }
        if ($UseEnvironment) {
            $null = $args.Add("--useEnv")
        }

        if ($null -ne $Body) {
            $null = $args.AddRange(@(
                "--data",
                ("{0}" -f ((ConvertTo-Json $Body -Compress) -replace '"', '\"'))
            ))
        }

        $VerboseArgs = for ($i = 0; $i -lt $args.Count; $i++) {
            if ($args[$i] -eq "--data") {
                Write-Output $args[$i]
                $i += 1
                Write-Output ("'{0}'" -f $args[$i])
            } else {
                Write-Output $args[$i]
            }
        }

        Write-Verbose ("$c8ycli {0}" -f ($VerboseArgs -join " "))
        $RawResponse = & $c8ycli $args

        $ExitCode = $LASTEXITCODE
        if ($ExitCode -ne 0) {
            try {
                $errormessage = $RawResponse | Select-Object -First 1 | ConvertFrom-Json
                Write-Error ("{0}: {1}" -f @(
                        $errormessage.error,
                        $errormessage.message
                    ))
            }
            catch {
                Write-Error "c8y command failed. $RawResponse"
            }
            return
        }

        if ($AsText) {
            Write-Output $RawResponse
            return
        }

        $response = $RawResponse | ConvertFrom-Json
        $global:_rawdata = $response
        $global:_data = $response
        Write-Output $response
    }
}

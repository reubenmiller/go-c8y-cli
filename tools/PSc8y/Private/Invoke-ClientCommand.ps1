﻿Function Invoke-ClientCommand {
<# 
.SYNOPSIS
Run a Cumulocity client command using the c8y binary. Only intended for internal usage only

.DESCRIPTION
The command is a wrapper around the c8y binary which is used to send the rest request to Cumulocity.

The result will also be parsed, and Powershell type information will be added to the result set, so
only relevant information is shown.
#>
    [cmdletbinding()]
    Param(
        # Name of the command
        [Parameter(
            Mandatory = $true
        )]
        [string] $Noun,

        # Command verb, i.e. list, get, delete etc.
        [Parameter(
            Mandatory = $true
        )]
        [string] $Verb,

        # Parameters which should be passed to the c8y binary
        # The full parameter name should be used (i.e. --header, and not -H)
        [hashtable] $Parameters,

        [string] $Type = "c8y.item",

        # Type to be added to the result set. Used to control the view of the 
        # returned data in Powershell
        [string] $ItemType,

        # Name of the property to return a portion (fragment) of the data instead of the full
        # data set.
        [string] $ResultProperty,

        # Future Roadmap: Currently not used: Include all result sets
        [switch] $IncludeAll,

        # Future Roadmap: Current page to return
        [int] $CurrentPage,

        # Total number of pages to retrieve (only used with -IncludeAll)
        [int] $TotalPages,

        # Return the raw response rather than Powershell objects
        [switch] $Raw,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

    $c8yargs = New-Object System.Collections.ArrayList
    $null = $c8yargs.Add($Noun)
    $null = $c8yargs.Add($Verb)

    foreach ($iKey in $Parameters.Keys) {
        $Value = $Parameters[$iKey]

        foreach ($iValue in $Value) {
            if ("$Value" -notmatch "^$") {
                $key = $iKey[0].ToString().ToLowerInvariant() + $iKey.SubString(1)
                if ($Value -is [bool] -and $Value) {
                    $null = $c8yargs.AddRange(@("--${key}"))
                } else {
                    if ($key -eq "data") {
                        # due to cli parsing, data needs to be sent using "="
                        $null = $c8yargs.AddRange(@("--${key}", $Value))
                    } else {
                        if ($Value -match " ") {
                            # $null = $c8yargs.AddRange(@("--${key}", "$Value"))
                            $null = $c8yargs.Add("--${key}=`"$Value`"")
                        } else {
                            $null = $c8yargs.Add("--${key}=$Value")
                        }
                    }
                }
            }
        }
    }



    $null = $c8yargs.Add("--pretty=false")

    if ($WhatIfPreference) {
        $null = $c8yargs.Add("--dry")
    }

    if ($VerbosePreference) {
        $null = $c8yargs.Add("--verbose")
    }

    if ($TimeoutSec) {
        # Convert to milliseconds (cast to an integer)
        [int] $TimeoutInMS = $TimeoutSec * 1000
        $null = $c8yargs.AddRange(@("--timeout", $TimeoutInMS))
    }

    if ($CurrentPage) {
        $null = $c8yargs.AddRange(@("--currentPage", $CurrentPage))
    }

    if ($TotalPages) {
        $null = $c8yargs.AddRange(@("--totalPages", $TotalPages))
    }

    # Include all pagination results
    if ($IncludeAll) {
        # Write-Warning "IncludeAll operation is currently not implemented"
        $null = $c8yargs.Add("--includeAll")
    }

    # Only use streaming when results come in batches, and powershell needs the results before
    # the go binary is finished (i.e. realtime and paged results)
    $UsePipelineStreaming = ($IncludeAll -or $TotalPages -gt 0) -or $Verb -match "subscribe|subscribeAll"

    # Don't use the raw response, let go do everything
    if (-Not $UsePipelineStreaming) {
        $null = $c8yargs.Add("--raw")
    }

    $c8ycli = Get-ClientBinary
    Write-Verbose ("$c8ycli {0}" -f $c8yargs -join " ")

    $ExitCode = $null
    $ErrorOutput = @()
    try {
        if ($UsePipelineStreaming) {
            Write-Verbose "Streaming c8y binary output"
            $LastSaveWarning = "NOTE: This PSc8y automatic variable is not supported when using -IncludeAll or -TotalPages"
            $global:_rawdata = $LastSaveWarning
            $global:_data = $LastSaveWarning


            $JSONOptions = @{}
            if ($PSVersionTable.PSVersion.Major -gt 5) {
                $JSONOptions.Depth = 100
            }
            
            # Note: To enable the streaming of output result in the pipeline,
            # the value must be sent back as is
            if ($Raw) {
                $null = $c8yargs.Add("--raw")
            }
            & $c8ycli $c8yargs 2>&1 | Write-ClientMessage -PassThru | ForEach-Object {
                $line = "$_"
                
                # JSON should be returned on output stream
                if (-not $Raw -and ($line.StartsWith("[") -or $line.StartsWith("{"))) {
                    try {
                        ConvertFrom-Json -InputObject $line @JSONOptions |
                            Select-Object |
                            Add-PowershellType $ItemType
                    } catch {
                        Write-Warning "Could not decode json text. $_"
                        $line
                    }
                } else {
                    # Return data as is
                    $line
                }
            }
            $ExitCode = $LASTEXITCODE
            return
        } else {
            # Don't stream stdout/stderr, just call external binary and redirect messages for mapping
            # into powershell streams
            [array] $VerboseMessages = $( $RawResponse = & $c8ycli $c8yargs ) 2>&1
            $ErrorOutput = Write-ClientMessage $VerboseMessages -PassThru
            $ExitCode = $LASTEXITCODE
        }
    } catch {
        Write-Warning -Message $_.Exception.Message
        # do nothing, due to remote powershell session issue and $ErrorActionPreference being set to 'Stop'
        # https://github.com/PowerShell/PowerShell/issues/4002
    }

    if ($null -eq $ExitCode) {
        $ExitCode = $LASTEXITCODE
        $global:C8Y_EXITCODE = $ExitCode
    }
    if ($ExitCode -ne 0) {

        try {
            if ($RawResponse) {
                if ($PSVersionTable.PSVersion.Major -gt 5) {
                    $errormessage = $RawResponse | Out-String | ConvertFrom-Json -Depth 100 -ErrorAction SilentlyContinue
                } else {
                    $errormessage = $RawResponse | Out-String | ConvertFrom-Json -ErrorAction SilentlyContinue
                }
            } else {
                if ($ErrorOutput -is [string] -and $ErrorOutput.startsWith("{")) {
                    $errormessage = $ErrorOutput | ConvertFrom-Json
                } else {
                    $errormessage = $ErrorOutput
                }
            }
            
            if ($errormessage.error) {
                
                $errorText = switch ($errormessage.error) {
                    "commandError" {
                        @($errormessage.error, $errormessage.message).Where({$_ -ne ""}) -join ": "
                    }
                    default {
                        # server error
                        @("serverError", $errormessage.error, $errormessage.message).Where({$_ -ne ""}) -join ": "
                    }
                }
            } else {
                $errorText = $errormessage
            }
            
            Write-Error $errorText
            
        } catch {
            Write-Error "$RawResponse"
        }
        return
    }

    $isJSON = $false
    try {
        # Hide senstive data in the response
        if ($env:C8Y_LOGGER_HIDE_SENSITIVE -eq "true") {
            if ($env:C8Y_TENANT) {
                $RawResponse = $RawResponse -replace [regex]::Unescape($env:C8Y_TENANT), "{tenant}"
            }
            if ($env:C8Y_USERNAME) {
                # $RawResponse = $RawResponse -replace [regex]::Unescape($env:C8Y_USERNAME), "{username}"
            }
            if ($env:C8Y_PASSWORD) {
                # $RawResponse = $RawResponse -replace [regex]::Unescape($env:C8Y_PASSWORD), "{password}"
            }
        }

        $response = ""
        if ($null -ne $RawResponse) {
            $response = $RawResponse | ConvertFrom-Json
            $isJSON = $true
        }
    } catch {
        # ignore json errors, because sometimes the response is not json...so we want
        # to return it anyways
    }

    # Return quickly if a non-json response is detected
    if (!$isJSON) {
        Write-Verbose "non-json response detected"
        $global:_rawdata = $RawResponse
        $global:_data = $null
        $RawResponse
        return
    }

    $NestedData = Get-NestedProperty -InputObject $response -Name $ResultProperty

    if ($ResultProperty -and $ItemType) {
        $null = $NestedData `
            | Select-Object `
            | Add-PowershellType $ItemType
    }

    if ($response -and $Type) {
        $null = $response `
            | Select-Object `
            | Add-PowershellType $Type
    }

    $ReturnRawData = $Raw -or [string]::IsNullOrEmpty($ResultProperty) -or (
        $Parameters.ContainsKey("WithTotalPages") -and
        $Parameters["WithTotalPages"]
    )

    if ($response.statistics.pageSize) {
        Write-Verbose ("Statistics: currentPage={2}, pageSize={0}, totalPages={1}" -f @(
            $response.statistics.pageSize,
            $response.statistics.totalPages,
            $response.statistics.currentPage
        ))
    }

    
    if ($NestedData -and $response.statistics) {
        #$NestedData | Add-Member -MemberType NoteProperty -Name "PSc8yResult" -Value $_data
        # Add information to each element in the array

        $StatsAsJson = ConvertTo-Json @{
            next = $response.next
            pageSize = $response.statistics.pageSize
            totalPages = $response.statistics.totalPages
            currentPage = $response.statistics.currentPage
        } -Compress -Depth 100

        $NewScriptBlock = [scriptblock]::Create("ConvertFrom-Json '$StatsAsJson'")

        $null = $NestedData | Add-Member -Name "PSc8yRequestSource" -MemberType "ScriptMethod" -Value $NewScriptBlock
    }
   

    # Save last value for easier recall on command line
    $global:_rawdata = $response
    $global:_data = $null


    if ($null -ne $NestedData -or $NestedData.Count -ge 0) {
        $global:_data = $NestedData
    }

    if ($ReturnRawData -or
        ($null -eq $NestedData -and $null -eq $NestedData.Count)) {
        $response
    } else {
        Write-Output $NestedData
    }
}

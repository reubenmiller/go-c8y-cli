﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceRequest {
<#
.SYNOPSIS
Get device request

.DESCRIPTION
Get a device registration request

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceregistration_get

.EXAMPLE
PS> Get-DeviceRequest -Id "010af8dd0c102"

Get a new device request


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # New Device Request ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceregistration get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.newDeviceRequest+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration get $c8yargs
        }
        
    }

    End {}
}

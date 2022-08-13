﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceGroupChildAsset {
<#
.SYNOPSIS
Get child device reference

.DESCRIPTION
Get managed object child device reference

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_devices_get

.EXAMPLE
PS> Get-DeviceGroupChildAsset -Group $Agent.id -Child $Ref.id

Get an existing child asset reference


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Asset id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Group,

        # Asset reference id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Child
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups devices get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Group `
            | Group-ClientRequests `
            | c8y devicegroups devices get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Group `
            | Group-ClientRequests `
            | c8y devicegroups devices get $c8yargs
        }
        
    }

    End {}
}
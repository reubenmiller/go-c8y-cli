﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Firmware {
<#
.SYNOPSIS
Create firmware package

.DESCRIPTION
Create a new firmware package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_create

.EXAMPLE
PS> New-Firmware -Name "iot-linux" -Description "Linux image for IoT devices" -Data @{$type=@{}}

Create a firmware package


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # name
        [Parameter()]
        [string]
        $Name,

        # Description of the firmware package
        [Parameter()]
        [string]
        $Description,

        # Device type filter. Only allow firmware to be applied to devices of this type
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $DeviceType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $DeviceType `
            | Group-ClientRequests `
            | c8y firmware create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $DeviceType `
            | Group-ClientRequests `
            | c8y firmware create $c8yargs
        }
        
    }

    End {}
}

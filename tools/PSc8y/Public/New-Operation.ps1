﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Operation {
<#
.SYNOPSIS
Create operation

.DESCRIPTION
Create a new operation for an agent or device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/operations_create

.EXAMPLE
PS> New-Operation -Device $device.id -Description "Restart device" -Data @{ c8y_Restart = @{} }

Create operation for a device

.EXAMPLE
PS> Get-Device $device.id | New-Operation -Description "Restart device" -Data @{ c8y_Restart = @{} }

Create operation for a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the target device on which this operation should be performed.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Text description of the operation.
        [Parameter()]
        [string]
        $Description
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.operation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y operations create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y operations create $c8yargs
        }
        
    }

    End {}
}

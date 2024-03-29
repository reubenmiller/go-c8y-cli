﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Firmware {
<#
.SYNOPSIS
Delete firmware package

.DESCRIPTION
Delete an existing firmware package

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_delete

.EXAMPLE
PS> Remove-Firmware -Id $firmware.id

Delete a firmware package and all related versions

.EXAMPLE
PS> Get-ManagedObject -Id $firmware.id | Remove-Firmware

Delete a firmware package (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $firmware.id | Remove-Firmware -ForceCascade:$false

Delete a firmware package but keep the binaries


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware Package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Remove version and any related binaries
        [Parameter()]
        [switch]
        $ForceCascade
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y firmware delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y firmware delete $c8yargs
        }
        
    }

    End {}
}

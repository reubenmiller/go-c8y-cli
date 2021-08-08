﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ManagedObject {
<#
.SYNOPSIS
Delete managed object

.DESCRIPTION
Delete an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_delete

.EXAMPLE
PS> Remove-ManagedObject -Id $mo.id

Delete a managed object

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-ManagedObject

Delete a managed object (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Remove-ManagedObject -Cascade

Delete a managed object and all child devices


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # When set to true and the managed object is a device or group, all the hierarchy will be deleted.
        [Parameter()]
        [switch]
        $Cascade,

        # When set to true all the hierarchy will be deleted without checking the type of managed object. It takes precedence over the parameter cascade.
        [Parameter()]
        [switch]
        $ForceCascade,

        # When set to true and the managed object is a device, it deletes the associated device user (credentials).
        [Parameter()]
        [switch]
        $WithDeviceUser
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory delete"
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
            | c8y inventory delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventory delete $c8yargs
        }
        
    }

    End {}
}

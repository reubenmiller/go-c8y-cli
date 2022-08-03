﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildDeviceCollection {
<#
.SYNOPSIS
Get child device collection

.DESCRIPTION
Get a collection of managedObjects child references

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_children_list

.EXAMPLE
PS> Get-ChildDeviceCollection -Device $Device.id

Get a list of the child devices of an existing device

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Get-ChildDeviceCollection

Get a list of the child devices of an existing device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.
        [Parameter()]
        [switch]
        $WithChildren
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices children list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices children list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices children list $c8yargs
        }
        
    }

    End {}
}

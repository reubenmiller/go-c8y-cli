﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceChildCollection {
<#
.SYNOPSIS
Get child collection

.DESCRIPTION
Get a collection of child managedObjects

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_children_list

.EXAMPLE
PS> Get-DeviceChildCollection -Id 12345 -ChildType addition

Get a list of the child additions of an existing managed object

.EXAMPLE
PS> Get-ManagedObject -Id 12345 | Get-DeviceChildCollection -ChildType addition

Get a list of the child additions of an existing managed object (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Child relationship type (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('addition','asset','device')]
        [string]
        $ChildType,

        # Additional query filter
        [Parameter()]
        [string]
        $Query,

        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter()]
        [string]
        $QueryTemplate,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

        # Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.
        [Parameter()]
        [switch]
        $WithChildren,

        # When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)
        [Parameter()]
        [switch]
        $WithChildrenCount
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
            $Id `
            | Group-ClientRequests `
            | c8y devices children list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices children list $c8yargs
        }
        
    }

    End {}
}

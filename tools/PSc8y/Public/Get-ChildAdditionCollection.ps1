﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAdditionCollection {
<#
.SYNOPSIS
Get child addition collection

.DESCRIPTION
Get a collection of managedObjects child additions

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_additions_list

.EXAMPLE
PS> Get-ChildAdditionCollection -Id $software.id

Get a list of the child additions of an existing managed object

.EXAMPLE
PS> Get-ManagedObject -Id $software.id | Get-ChildAdditionCollection

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory additions list"
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
            | c8y inventory additions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventory additions list $c8yargs
        }
        
    }

    End {}
}

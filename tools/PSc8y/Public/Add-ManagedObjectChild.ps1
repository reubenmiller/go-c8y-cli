﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ManagedObjectChild {
<#
.SYNOPSIS
Assign child

.DESCRIPTION
Assign an existing managed object as a child to an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_children_assign

.EXAMPLE
PS> Add-ManagedObjectChild -Id $software.id -Child $version.id -ChildType addition

Add a related managed object as a child addition to an existing managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id where the child will be assigned to (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Managed object that will be assigned as a child (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Child,

        # Child relationship type (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('addition','asset','device')]
        [string]
        $ChildType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory children assign"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Child `
            | Group-ClientRequests `
            | c8y inventory children assign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y inventory children assign $c8yargs
        }
        
    }

    End {}
}

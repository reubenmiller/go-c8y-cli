# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-ManagedObjectCollection {
<#
.SYNOPSIS
Find managed object collection

.DESCRIPTION
Get a collection of managedObjects based on the Cumulocity query language

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_find

.EXAMPLE
PS> Find-ManagedObjectCollection -Query "name eq 'roomUpperFloor_*'"

Find all devices with their names starting with 'roomUpperFloor_'


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory find"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y inventory find $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y inventory find $c8yargs
        }
    }

    End {}
}

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

Find all managed objects with their names starting with 'roomUpperFloor_'


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject query
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Query,

        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter()]
        [string]
        $QueryTemplate,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by type
        [Parameter()]
        [string]
        $Type,

        # Only include agents
        [Parameter()]
        [switch]
        $Agents,

        # Filter by fragment type
        [Parameter()]
        [string]
        $FragmentType,

        # Filter by owner
        [Parameter()]
        [string]
        $Owner,

        # Filter by c8y_Availability.status
        [Parameter()]
        [ValidateSet('AVAILABLE','UNAVAILABLE','MAINTENANCE')]
        [string]
        $Availability,

        # Filter c8y_Availability.lastMessage to a specific date
        [Parameter()]
        [string]
        $LastMessageDateTo,

        # Filter c8y_Availability.lastMessage from a specific date
        [Parameter()]
        [string]
        $LastMessageDateFrom,

        # Filter creationTime.date to a specific date
        [Parameter()]
        [string]
        $CreationTimeDateTo,

        # Filter creationTime.date from a specific date
        [Parameter()]
        [string]
        $CreationTimeDateFrom,

        # Filter by group inclusion
        [Parameter()]
        [object[]]
        $Group,

        # Only include devices (deprecated)
        [Parameter()]
        [switch]
        $OnlyDevices,

        # Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved
        [Parameter()]
        [switch]
        $SkipChildrenNames,

        # Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.
        [Parameter()]
        [switch]
        $WithChildren,

        # When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)
        [Parameter()]
        [switch]
        $WithChildrenCount,

        # When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.
        [Parameter()]
        [switch]
        $WithGroups,

        # Include a flat list of all parents and grandparents of the given object
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
            $Query `
            | Group-ClientRequests `
            | c8y inventory find $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y inventory find $c8yargs
        }
        
    }

    End {}
}

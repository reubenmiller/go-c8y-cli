# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-DeviceServiceCollection {
<#
.SYNOPSIS
Find services

.DESCRIPTION
Find services of any device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_services_find

.EXAMPLE
PS> Find-DeviceServiceCollection

Find all services (from any device)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Additional query filter
        [Parameter()]
        [string]
        $Query,

        # Filter by service type e.g. systemd
        [Parameter()]
        [string]
        $ServiceType,

        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by service status (custom values allowed)
        [Parameter()]
        [ValidateSet('up','down','unknown')]
        [string]
        $Status,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices services find"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devices services find $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devices services find $c8yargs
        }
    }

    End {}
}

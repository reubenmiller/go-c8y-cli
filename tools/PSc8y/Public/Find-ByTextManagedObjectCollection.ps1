# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-ByTextManagedObjectCollection {
<#
.SYNOPSIS
Find managed object by text collection

.DESCRIPTION
Find a collection of managedObjects which match a given text value

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_findByText

.EXAMPLE
PS> Find-ByTextManagedObjectCollection -Text $Device1.name

Find a list of managed objects by text

.EXAMPLE
PS> Find-ByTextManagedObjectCollection -Text $Device1.name

Find managed objects which contain the text 'myText' (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Search for managed objects where a property value is equal to the given one. The following properties are examined: id, type, name, owner, externalIds. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Text,

        # ManagedObject type.
        [Parameter()]
        [string]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

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
        $WithParents,

        # (FEATURE_PREVIEW) Include c8y_LatestMeasurements fragment, which contains the latest measurement values reported by the device to the platform
        [Parameter()]
        [switch]
        $WithLatestValues
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory findByText"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Text `
            | Group-ClientRequests `
            | c8y inventory findByText $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Text `
            | Group-ClientRequests `
            | c8y inventory findByText $c8yargs
        }
        
    }

    End {}
}

# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SoftwareVersionCollection {
<#
.SYNOPSIS
Get software package version collection

.DESCRIPTION
Get a collection of software package versions (managedObjects) based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_list

.EXAMPLE
PS> Get-SoftwareVersionCollection -Software $software.id

Get a list of software package versions


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

        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter()]
        [string]
        $QueryTemplate,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

        # Software package id or name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Software,

        # Filter by version
        [Parameter()]
        [string]
        $Version,

        # Filter by url
        [Parameter()]
        [string]
        $Url,

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

        # Include parent references
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Software `
            | Group-ClientRequests `
            | c8y software versions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Software `
            | Group-ClientRequests `
            | c8y software versions list $c8yargs
        }
        
    }

    End {}
}

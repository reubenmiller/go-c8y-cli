# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObjectCollection {
<#
.SYNOPSIS
Get managed object collection

.DESCRIPTION
Get a collection of managedObjects based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_list

.EXAMPLE
PS> Get-ManagedObjectCollection

Get a list of managed objects

.EXAMPLE
PS> Get-ManagedObjectCollection -Ids $Device1.id, $Device2.id

Get a list of managed objects by id


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # List of ids.
        [Parameter()]
        [string[]]
        $Ids,

        # ManagedObject type.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

        # List of managed objects that are owned by the given username.
        [Parameter()]
        [string]
        $Owner,

        # managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z).
        [Parameter()]
        [string]
        $Text,

        # When set to `true` it returns managed objects which don't have any parent. If the current user doesn't have access to the parent, this is also root for the user
        [Parameter()]
        [switch]
        $OnlyRoots,

        # Search for a specific child addition and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAdditionId,

        # Search for a specific child asset and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAssetId,

        # Search for a specific child device and list all the groups to which it belongs.
        [Parameter()]
        [object[]]
        $ChildDeviceId,

        # Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved
        [Parameter()]
        [switch]
        $SkipChildrenNames,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents,

        # Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.
        [Parameter()]
        [switch]
        $WithChildren,

        # When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.
        [Parameter()]
        [switch]
        $WithGroups
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Type `
            | Group-ClientRequests `
            | c8y inventory list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Type `
            | Group-ClientRequests `
            | c8y inventory list $c8yargs
        }
        
    }

    End {}
}

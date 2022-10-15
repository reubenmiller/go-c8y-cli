# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceGroupCollection {
<#
.SYNOPSIS
Get device group collection

.DESCRIPTION
Get a collection of device groups based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_list

.EXAMPLE
PS> Get-DeviceGroupCollection -Name "parent*"

Get a collection of device groups with names that start with 'parent'


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter()]
        [string]
        $QueryTemplate,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

        # Additional query filter
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Query,

        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by type
        [Parameter()]
        [string]
        $Type,

        # Filter by fragment type
        [Parameter()]
        [string]
        $FragmentType,

        # Filter by owner
        [Parameter()]
        [string]
        $Owner,

        # Filter by group inclusion
        [Parameter()]
        [switch]
        $ExcludeRootGroup,

        # Filter by group inclusion
        [Parameter()]
        [object[]]
        $Group,

        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents,

        # Include names of child assets (only use where necessary as it is slow for large groups)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedobjectcollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Query `
            | Group-ClientRequests `
            | c8y devicegroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y devicegroups list $c8yargs
        }
        
    }

    End {}
}

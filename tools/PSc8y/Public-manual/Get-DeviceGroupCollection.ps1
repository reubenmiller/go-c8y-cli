Function Get-DeviceGroupCollection {
<#
.SYNOPSIS
Get a collection of device groups

.DESCRIPTION
Get a collection of device groups. Device groups are used to arrange devices together.

.LINK
c8y devicegroups list

.EXAMPLE
Get-DeviceGroupCollection -Name *Room*

Get all device groups with "Room" in their name

.EXAMPLE
Get-DeviceGroupCollection -Query "creationTime.date gt '2020-01-01T00:00:00Z'"

Get a list of devices groups which have been created more recently than 2020-01-01

#>
    [cmdletbinding(PositionalBinding=$true, HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device group name. Wildcards accepted
        [Parameter(Mandatory = $false)]
        [string]
        $Name,

        # Device group type.
        [Parameter(Mandatory = $false)]
        [string]
        $Type,

        # Device group fragment type.
        [Parameter(Mandatory = $false)]
        [string]
        $FragmentType,

        # Device group owner.
        [Parameter(Mandatory = $false)]
        [string]
        $Owner,

        # Query.
        [Parameter(Mandatory = $false)]
        [string]
        $Query,

        # Exclude root groups from the list
        [Parameter(Mandatory = $false)]
        [switch]
        $ExcludeRootGroup,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDeviceGroupCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicegroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicegroups list $c8yargs
        }
    }

    End {}
}

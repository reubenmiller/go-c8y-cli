Function Get-DeviceCollection {
<#
.SYNOPSIS
Get a collection of devices

.DESCRIPTION
Get a collection of devices in Cumulocity by using the inventory API.

.EXAMPLE
Get-DeviceCollection -Name *sensor*

Get all devices with "sensor" in their name

.EXAMPLE
Get-DeviceCollection -Name *sensor* -Type *c8y_* -PageSize 100

Get the first 100 devices with "sensor" in their name and has a type matching "c8y_"

.EXAMPLE
Get-DeviceCollection -Query "lastUpdated.date gt '2020-01-01T00:00:00Z'"

Get a list of devices which have been updated more recently than 2020-01-01

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device name. Wildcards accepted
        [Parameter(Mandatory = $false)]
        [string]
        $Name,

        # Device type.
        [Parameter(Mandatory = $false)]
        [string]
        $Type,

        # Device fragment type.
        [Parameter(Mandatory = $false)]
        [string]
        $FragmentType,

        # Device owner.
        [Parameter(Mandatory = $false)]
        [string]
        $Owner,

        # Query.
        [Parameter(Mandatory = $false)]
        [string]
        $Query,

        # Order results by a specific field. i.e. "name", "_id desc" or "creationTime.date asc".
        [Parameter(Mandatory = $false)]
        [string]
        $OrderBy,

        # Only include agents.
        [Parameter()]
        [switch]
        $Agents,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Name")) {
            $Parameters["name"] = $Name
        }
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("FragmentType")) {
            $Parameters["fragmentType"] = $FragmentType
        }
        if ($PSBoundParameters.ContainsKey("owner")) {
            $Parameters["owner"] = $Owner
        }
        if ($PSBoundParameters.ContainsKey("Query")) {
            $Parameters["query"] = $Query
        }
        if ($PSBoundParameters.ContainsKey("OrderBy")) {
            $Parameters["orderBy"] = $OrderBy
        }
        if ($PSBoundParameters.ContainsKey("Agents")) {
            $Parameters["agents"] = $Agents
        }
        if ($PSBoundParameters.ContainsKey("WithParents")) {
            $Parameters["withParents"] = $WithParents
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDeviceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customDevice+json"
            BoundParameters = $PSBoundParameters
        }
    }
    DynamicParam {
        Get-ClientCommonParameters -Type "Collection" -BoundParameters $PSBoundParameters
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devices list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devices list $c8yargs
        }
    }
}

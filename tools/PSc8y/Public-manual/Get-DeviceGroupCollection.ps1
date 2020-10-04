Function Get-DeviceGroupCollection {
<#
.SYNOPSIS
Get a collection of device groups

.DESCRIPTION
Get a collection of device groups. Device groups are used to arrange devices together.

.EXAMPLE
Get-DeviceGroupCollection -Name *Room*

Get all device groups with "Room" in their name

.EXAMPLE
Get-DeviceGroupCollection -Query "creationTime.date gt '2020-01-01T00:00:00Z'"

Get a list of devices groups which have been created more recently than 2020-01-01

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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
        $WithParents,

        # Maximum number of results
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateRange(1,2000)]
        [int]
        $PageSize,

        # Include total pages statistic
        [Parameter()]
        [switch]
        $WithTotalPages,

        # Get a specific page result
        [Parameter()]
        [int]
        $CurrentPage,

        # Maximum number of pages to retrieve when using -IncludeAll
        [Parameter()]
        [int]
        $TotalPages,

        # Include all results
        [Parameter()]
        [switch]
        $IncludeAll,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
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
        if ($PSBoundParameters.ContainsKey("ExcludeRootGroup")) {
            $Parameters["excludeRootGroup"] = $ExcludeRootGroup
        }
        if ($PSBoundParameters.ContainsKey("WithParents")) {
            $Parameters["withParents"] = $WithParents
        }
        if ($PSBoundParameters.ContainsKey("PageSize")) {
            $Parameters["pageSize"] = $PageSize
        }
        if ($PSBoundParameters.ContainsKey("WithTotalPages") -and $WithTotalPages) {
            $Parameters["withTotalPages"] = $WithTotalPages
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

    }

    Process {
        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject "")
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "devices" `
            -Verb "listDeviceGroups" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.customDeviceGroupCollection+json" `
            -ItemType "application/vnd.com.nsn.cumulocity.customDeviceGroup+json" `
            -ResultProperty "managedObjects" `
            -Raw:$Raw `
            -CurrentPage:$CurrentPage `
            -TotalPages:$TotalPages `
            -IncludeAll:$IncludeAll
    }

}

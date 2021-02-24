# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-ManagedObjectCollection {
<#
.SYNOPSIS
Get a collection of managedObjects based on Cumulocity query language

.DESCRIPTION
Get a collection of managedObjects based on Cumulocity query language

.EXAMPLE
PS> Find-ManagedObjectCollection -Query "name eq 'roomUpperFloor_*'"
Find all devices with their names starting with 'roomUpperFloor_'


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject query. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Query,

        # ManagedObject sort results by.
        [Parameter(Mandatory = $false)]
        [string]
        $OrderBy,

        # include a flat list of all parents and grandparents of the given object
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

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session,

        # Timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $Timeout
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Query")) {
            $Parameters["query"] = $Query
        }
        if ($PSBoundParameters.ContainsKey("OrderBy")) {
            $Parameters["orderBy"] = $OrderBy
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
        if ($PSBoundParameters.ContainsKey("Timeout")) {
            $Parameters["timeout"] = $Timeout
        }

    }

    Process {
        foreach ($item in @("")) {

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "inventory" `
                -Verb "find" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.managedObjectCollection+json" `
                -ItemType "application/vnd.com.nsn.cumulocity.managedObject+json" `
                -ResultProperty "managedObjects" `
                -Raw:$Raw `
                -CurrentPage:$CurrentPage `
                -TotalPages:$TotalPages `
                -IncludeAll:$IncludeAll
        }
    }

    End {}
}

# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AllTenantUsageSummaryStatistics {
<#
.SYNOPSIS
Get collection of tenant usage statistics summary

.DESCRIPTION
Get collection of tenant usage statistics summary

.EXAMPLE
PS> Get-AllTenantUsageSummaryStatistics

Get tenant summary statistics for all tenants

.EXAMPLE
PS> Get-AllTenantUsageSummaryStatistics -DateFrom "-30d"

Get tenant summary statistics collection for the last 30 days

.EXAMPLE
PS> Get-AllTenantUsageSummaryStatistics -DateFrom "-10d" -DateTo "-9d"

Get tenant summary statistics collection for the last 10 days, only return until the last 9 days


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Start date or date and time of the statistics.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of the statistics.
        [Parameter()]
        [string]
        $DateTo,

        # Show the full (raw) response from Cumulocity including pagination information
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
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
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
        foreach ($item in @("")) {


            Invoke-ClientCommand `
                -Noun "tenantStatistics" `
                -Verb "listSummaryAllTenants" `
                -Parameters $Parameters `
                -Type "application/json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}

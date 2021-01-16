# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantUsageSummaryStatistics {
<#
.SYNOPSIS
Get collection of tenant usage statistics summary

.DESCRIPTION
Get summary of requests and database usage from the start of this month until now

.EXAMPLE
PS> Get-TenantUsageSummaryStatistics

Get tenant summary statistics for the current tenant

.EXAMPLE
PS> Get-TenantUsageSummaryStatistics -DateFrom "-30d"

Get tenant summary statistics collection for the last 30 days

.EXAMPLE
PS> Get-TenantUsageSummaryStatistics -DateFrom "-10d" -DateTo "-9d"

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

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        foreach ($item in @("")) {


            Invoke-ClientCommand `
                -Noun "tenantStatistics" `
                -Verb "listSummaryForTenant" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.tenantUsageStatisticsSummary+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}

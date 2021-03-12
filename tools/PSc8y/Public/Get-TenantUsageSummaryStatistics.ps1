# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantUsageSummaryStatistics {
<#
.SYNOPSIS
Get tenant usage summary statistics

.DESCRIPTION
Get summary of requests and database usage from the start of this month until now

.LINK
c8y tenantStatistics listSummaryForTenant

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
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
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
        $DateTo
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantStatistics listSummaryForTenant"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenantUsageStatisticsSummary+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenantStatistics listSummaryForTenant $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenantStatistics listSummaryForTenant $c8yargs
        }
    }

    End {}
}

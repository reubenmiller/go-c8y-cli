# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantStatisticsCollection {
<#
.SYNOPSIS
Get tenant usage statistics

.DESCRIPTION
Get collection of tenant usage statistics

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenantstatistics_list

.EXAMPLE
PS> Get-TenantStatisticsCollection

Get tenant statistics collection

.EXAMPLE
PS> Get-TenantStatisticsCollection -DateFrom "-30d" -PageSize 30

Get tenant statistics collection for the last 30 days

.EXAMPLE
PS> Get-TenantStatisticsCollection -DateFrom "-3d" -DateTo "-2d"

Get tenant statistics collection for the day before yesterday


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
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantstatistics list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenantUsageStatisticsCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.tenantUsageStatisticsSummary+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenantstatistics list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenantstatistics list $c8yargs
        }
    }

    End {}
}

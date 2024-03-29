﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AllTenantUsageSummaryStatistics {
<#
.SYNOPSIS
Get all tenant usage summary statistics

.DESCRIPTION
Get collection of tenant usage statistics summary

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenantstatistics_listSummaryAllTenants

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantstatistics listSummaryAllTenants"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenantstatistics listSummaryAllTenants $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenantstatistics listSummaryAllTenants $c8yargs
        }
    }

    End {}
}

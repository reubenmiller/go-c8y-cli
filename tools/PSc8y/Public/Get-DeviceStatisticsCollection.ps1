# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceStatisticsCollection {
<#
.SYNOPSIS
Retrieve device statistics

.DESCRIPTION
Retrieve device statistics from a specific tenant (by a given ID). Either daily or monthly.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_statistics_list

.EXAMPLE
PS> Get-DeviceStatisticsCollection

Get device statistics


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Date of the queried day. When type is set to monthly then will be ignored.
        [Parameter()]
        [string]
        $Date,

        # Aggregation type. e.g. daily or monthly
        [Parameter()]
        [ValidateSet('daily','monthly')]
        [string]
        $Type,

        # The ID of the device to search for.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Tenant id. Defaults to current tenant (based on credentials)
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices statistics list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices statistics list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices statistics list $c8yargs
        }
        
    }

    End {}
}

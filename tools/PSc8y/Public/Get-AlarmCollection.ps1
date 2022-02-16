# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AlarmCollection {
<#
.SYNOPSIS
Get alarm collection

.DESCRIPTION
Get a collection of alarms based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_list

.EXAMPLE
PS> Get-AlarmCollection -Severity MAJOR -PageSize 100

Get alarms with the severity set to MAJOR

.EXAMPLE
PS> Get-AlarmCollection -DateFrom "-10m" -Status ACTIVE

Get active alarms which occurred in the last 10 minutes

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-AlarmCollection -Status ACTIVE

Get active alarms from a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Source device id.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Start date or date and time of alarm occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of alarm occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # Alarm type.
        [Parameter()]
        [string]
        $Type,

        # Comma separated alarm statuses, for example ACTIVE,CLEARED.
        [Parameter()]
        [ValidateSet('ACTIVE','ACKNOWLEDGED','CLEARED')]
        [string[]]
        $Status,

        # Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.
        [Parameter()]
        [ValidateSet('CRITICAL','MAJOR','MINOR','WARNING')]
        [string]
        $Severity,

        # When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.
        [Parameter()]
        [switch]
        $Resolved,

        # When set to true also alarms for related source devices will be included in the request. When this parameter is provided a source must be specified.
        [Parameter()]
        [switch]
        $WithSourceAssets,

        # When set to true also alarms for related source devices will be removed. When this parameter is provided also source must be defined.
        [Parameter()]
        [switch]
        $WithSourceDevices,

        # Start date or date and time of the alarm creation. Version >= 10.11
        [Parameter()]
        [string]
        $CreatedFrom,

        # End date or date and time of the alarm creation. Version >= 10.11
        [Parameter()]
        [string]
        $CreatedTo,

        # Start date or date and time of the last update made. Version >= 10.11
        [Parameter()]
        [string]
        $LastUpdatedFrom,

        # End date or date and time of the last update made. Version >= 10.11
        [Parameter()]
        [string]
        $LastUpdatedTo
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.alarmCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.alarm+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y alarms list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y alarms list $c8yargs
        }
        
    }

    End {}
}

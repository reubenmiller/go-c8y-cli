# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AlarmCount {
<#
.SYNOPSIS
Retrieve the total number of alarms

.DESCRIPTION
Count the total number of active alarms on your tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_count

.EXAMPLE
PS> Get-AlarmCount -Severity MAJOR

Get number of active alarms with the severity set to MAJOR

.EXAMPLE
PS> Get-AlarmCount -DateFrom "-10m" -Status ACTIVE

Get number of active alarms which occurred in the last 10 minutes

.EXAMPLE
PS> Get-AlarmCount -DateFrom "-10m" -Status ACTIVE -Device $Device.name

Get number of active alarms which occurred in the last 10 minutes on a device

.EXAMPLE
PS> Get-Device -Id $Device.id | Get-AlarmCount -DateFrom "-10m"

Get number of alarms from a list of devices using pipeline


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
        $WithSourceDevices
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms count"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "text/plain"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y alarms count $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y alarms count $c8yargs
        }
        
    }

    End {}
}

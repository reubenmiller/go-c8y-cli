﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-AlarmCollection {
<#
.SYNOPSIS
Delete alarm collection

.DESCRIPTION
Delete a collection of alarms by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_deleteCollection

.EXAMPLE
PS> Remove-AlarmCollection -Device "{{ randomdevice }}" -Severity MAJOR

Remove alarms on the device with the severity set to MAJOR

.EXAMPLE
PS> Remove-AlarmCollection -Device $device.id -DateFrom "-10m" -Status ACTIVE

Remove alarms on the device which are active and created in the last 10 minutes

.EXAMPLE
PS> Get-Device -Id $device.id | PSc8y\Remove-AlarmCollection -DateFrom "-10m" -Status ACTIVE

Remove alarms on the device which are active and created in the last 10 minutes (using pipeline)


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

        # When set to true also alarms for related source assets will be removed. When this parameter is provided also source must be defined.
        [Parameter()]
        [switch]
        $WithSourceAssets,

        # When set to true also alarms for related source devices will be removed. When this parameter is provided also source must be defined.
        [Parameter()]
        [switch]
        $WithSourceDevices
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms deleteCollection"
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
            | c8y alarms deleteCollection $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y alarms deleteCollection $c8yargs
        }
        
    }

    End {}
}

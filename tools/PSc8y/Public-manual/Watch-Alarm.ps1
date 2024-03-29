﻿Function Watch-Alarm {
<#
.SYNOPSIS
Watch realtime alarms

.DESCRIPTION
Watch realtime alarms

.LINK
c8y alarms subscribe

.EXAMPLE
PS> Watch-Alarm -Device 12345

Watch all alarms for a device

.EXAMPLE
Watch-Alarm -Device 12345 -Duration 600s | Foreach-object {
    $alarm = $_
    $daysOld = ($alarm.time - $alarm.creationTime).TotalDays
    if ($alarm.status -eq "ACTIVE" -and $daysOld -gt 1) {
        $alarm | Update-Alarm -Severity CRITICAL -Force
    }}

Subscribe to realtime alarm notifications for a device, and update the alarm severity to CRITICAL
if the alarm is active and was first created more than 1 day ago.

#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Device,

        # Duration to subscribe for. It accepts a duration, i.e. 1ms, 0.5s, 1m etc.
        [Parameter()]
        [string]
        $Duration,

        # End date or date and time of alarm occurrence.
        [Parameter()]
        [int]
        $Count,

        # Filter by realtime action types, i.e. CREATE,UPDATE,DELETE
        [Parameter()]
        [ValidateSet('CREATE','UPDATE','DELETE','')]
        [string[]]
        $ActionTypes
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y alarms subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y alarms subscribe $c8yargs
        }
    }

    End {}
}

Function Watch-Alarm {
<#
.SYNOPSIS
Watch realtime alarms

.DESCRIPTION
Watch realtime alarms

.EXAMPLE
PS> Watch-Alarm -Device 12345

Watch all alarms for a device

.EXAMPLE
Watch-Alarm -Device 12345 -DurationSec 600 | Foreach-object {
    $alarm = $_
    $daysOld = ($alarm.time - $alarm.creationTime).TotalDays
    if ($alarm.status -eq "ACTIVE" -and $daysOld -gt 1) {
        $alarm | Update-Alarm -Severity CRITICAL -Force
    }}

Subscribe to realtime alarm notifications for a device, and update the alarm severity to CRITICAL
if the alarm is active and was first created more than 1 day ago.

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Device,

        # Start date or date and time of alarm occurrence. (required)
        [Alias("DurationSec")]
        [Parameter()]
        [int]
        $Duration,

        # End date or date and time of alarm occurrence.
        [Parameter()]
        [int]
        $Count
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get" -BoundParameters $PSBoundParameters
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
        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $Device)
            )) {
            return
        }

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

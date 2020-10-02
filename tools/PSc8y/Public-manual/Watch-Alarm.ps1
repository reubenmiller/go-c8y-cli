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
        [object[]]
        $Device,

        # Start date or date and time of alarm occurrence. (required)
        [Parameter()]
        [int]
        $DurationSec,

        # End date or date and time of alarm occurrence.
        [Parameter()]
        [string]
        $Count,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("DurationSec")) {
            $Parameters["duration"] = $DurationSec
        }
        if ($PSBoundParameters.ContainsKey("Count")) {
            $Parameters["count"] = $Count
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
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["dryRun"] = $Session
        }

    }

    Process {
        $id = PSc8y\Expand-Id $Device
        if ($id) {
            $Parameters["device"] = PSc8y\Expand-Id $Device
        }

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }
        
        Invoke-ClientCommand `
            -Noun "alarms" `
            -Verb "subscribe" `
            -Parameters $Parameters `
            -Type "application/json"
    }

    End {}
}

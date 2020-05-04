Function Watch-Notification {
<#
.SYNOPSIS
Watch realtime notifications

.DESCRIPTION
Watch realtime notifications

.EXAMPLE
PS> Watch-Notification -Channel "/measurements/*" -DurationSec 90
Watch all measurements for 90 seconds

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(
            Mandatory = $true)]
        [string]
        $Channel,

        # Start date or date and time of notification occurrence. (required)
        [Parameter()]
        [int]
        $DurationSec,

        # End date or date and time of notification occurrence.
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
        if ($PSBoundParameters.ContainsKey("Channel")) {
            $Parameters["channel"] = $Channel
        }
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

    }

    Process {
        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "realtime" `
            -Verb "subscribe" `
            -Parameters $Parameters `
            -Type "application/json" `
            -ItemType "" `
            -ResultProperty "" `
            -Raw:$Raw `
            -IncludeAll:$IncludeAll
    }

    End {}
}

Function Watch-Operation {
<#
.SYNOPSIS
Watch realtime operations

.DESCRIPTION
Watch realtime operations

.EXAMPLE
PS> Watch-Operation -Device 12345
Watch all operations for a device

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

        # Start date or date and time of operation occurrence. (required)
        [Parameter()]
        [int]
        $DurationSec,

        # End date or date and time of operation occurrence.
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
            -Noun "operations" `
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

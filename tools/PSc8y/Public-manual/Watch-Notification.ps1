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
        # Channel
        [Parameter(
            Mandatory = $true)]
        [string]
        $Channel,

        # Start date or date and time of notification occurrence. (required)
        [Alias("DurationSec")]
        [Parameter()]
        [int]
        $Duration,

        # End date or date and time of notification occurrence.
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "realtime subscribe"
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
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        if ($ClientOptions.ConvertToPS) {
            c8y realtime subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y realtime subscribe $c8yargs
        }
    }

    End {}
}

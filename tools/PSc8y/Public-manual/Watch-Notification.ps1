Function Watch-Notification {
<#
.SYNOPSIS
Watch realtime notifications

.DESCRIPTION
Watch realtime notifications

.LINK
c8y realtime subscribe

.EXAMPLE
PS> Watch-Notification -Channel "/measurements/*" -DurationSec 90
Watch all measurements for 90 seconds

#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "realtime subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

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

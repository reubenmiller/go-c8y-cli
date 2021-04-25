Function Watch-NotificationChannel {
<#
.SYNOPSIS
Watch realtime device notifications

.DESCRIPTION
Watch realtime device notifications

.LINK
c8y realtime subscribeAll

.EXAMPLE
PS> Function Watch-NotificationChannel -Device 12345 -DurationSec 90
Watch all types of notifications for a device for 90 seconds

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "realtime subscribeAll"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y realtime subscribeAll $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y realtime subscribeAll $c8yargs
        }
    }

    End {}
}

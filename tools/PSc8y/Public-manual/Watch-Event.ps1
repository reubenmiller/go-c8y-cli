Function Watch-Event {
<#
.SYNOPSIS
Watch realtime events

.DESCRIPTION
Watch realtime events

.LINK
c8y events subscribe

.EXAMPLE
PS> Watch-Event -Device 12345
Watch all events for a device

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

        # Start date or date and time of event occurrence. (required)
        [Alias("DurationSec")]
        [Parameter()]
        [int]
        $Duration,

        # End date or date and time of event occurrence.
        [Parameter()]
        [int]
        $Count
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y events subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y events subscribe $c8yargs
        }
    }

    End {}
}

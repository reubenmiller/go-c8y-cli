Function Watch-Operation {
<#
.SYNOPSIS
Watch realtime operations

.DESCRIPTION
Watch realtime operations

.LINK
c8y operations subscribe

.EXAMPLE
PS> Watch-Operation -Device 12345
Watch all operations for a device

#>
    [cmdletbinding(PositionalBinding=$true, HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Device,

        # Start date or date and time of operation occurrence. (required)
        [Alias("DurationSec")]
        [Parameter()]
        [int]
        $Duration,

        # End date or date and time of operation occurrence.
        [Parameter()]
        [int]
        $Count
    )

    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y operations subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y operations subscribe $c8yargs
        }
    }

    End {}
}

﻿Function Watch-Measurement {
<#
.SYNOPSIS
Watch realtime measurements

.DESCRIPTION
Watch realtime measurements

.LINK
c8y measurements subscribe

.EXAMPLE
PS> Watch-Measurement -Device 12345
Watch all measurements for a device

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

        # End date or date and time of measurement occurrence.
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "measurements subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y measurements subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y measurements subscribe $c8yargs
        }
    }

    End {}
}

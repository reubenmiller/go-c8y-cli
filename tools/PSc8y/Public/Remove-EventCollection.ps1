# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-EventCollection {
<#
.SYNOPSIS
Delete event collection

.DESCRIPTION
Delete a collection of events by using a filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_deleteCollection

.EXAMPLE
PS> Remove-EventCollection -Type my_CustomType -DateFrom "-10d"

Remove events with type 'my_CustomType' that were created in the last 10 days

.EXAMPLE
PS> Remove-EventCollection -Device "{{ randomdevice }}"

Remove events from a device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Event type.
        [Parameter()]
        [string]
        $Type,

        # Fragment name from event.
        [Parameter()]
        [string]
        $FragmentType,

        # Start date or date and time of the event's creation (set by the platform during creation).
        [Parameter()]
        [string]
        $CreatedFrom,

        # End date or date and time of the event's creation (set by the platform during creation).
        [Parameter()]
        [string]
        $CreatedTo,

        # Start date or date and time of event occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of event occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # Return the newest instead of the oldest events. Must be used with dateFrom and dateTo parameters
        [Parameter()]
        [switch]
        $Revert
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events deleteCollection"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y events deleteCollection $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y events deleteCollection $c8yargs
        }
        
    }

    End {}
}

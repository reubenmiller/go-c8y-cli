# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-EventCollection {
<#
.SYNOPSIS
Get event collection

.DESCRIPTION
Get a collection of events based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_list

.EXAMPLE
PS> Get-EventCollection -Type "my_CustomType2" -DateFrom "-10d"

Get events with type 'my_CustomType' that were created in the last 10 days

.EXAMPLE
PS> Get-EventCollection -Device $Device.id

Get events from a device

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-EventCollection

Get events from a device (using pipeline)


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

        # Allows filtering events by the fragment's value, but only when provided together with fragmentType.
        [Parameter()]
        [string]
        $FragmentValue,

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

        # Start date or date and time of the last update made.
        [Parameter()]
        [string]
        $LastUpdatedFrom,

        # End date or date and time of the last update made.
        [Parameter()]
        [string]
        $LastUpdatedTo,

        # Return the newest instead of the oldest events. Must be used with dateFrom and dateTo parameters
        [Parameter()]
        [switch]
        $Revert,

        # When set to true also events for related source assets will be included in the request. When this parameter is provided a source must be specified.
        [Parameter()]
        [switch]
        $WithSourceAssets,

        # When set to true also events for related source devices will be included in the request. When this parameter is provided a source must be specified.
        [Parameter()]
        [switch]
        $WithSourceDevices
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.eventCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.event+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y events list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y events list $c8yargs
        }
        
    }

    End {}
}

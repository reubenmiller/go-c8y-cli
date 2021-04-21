# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Event {
<#
.SYNOPSIS
Create event

.DESCRIPTION
Create a new event for a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_create

.EXAMPLE
PS> New-Event -Device $device.id -Type c8y_TestAlarm -Text "Test event"

Create a new event for a device

.EXAMPLE
PS> Get-Device -Id $device.id | PSc8y\New-Event -Type c8y_TestAlarm -Time "-0s" -Text "Test event"

Create a new event for a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The ManagedObject which is the source of this event.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Time of the event. Defaults to current timestamp.
        [Parameter()]
        [string]
        $Time,

        # Identifies the type of this event.
        [Parameter()]
        [string]
        $Type,

        # Text description of the event.
        [Parameter()]
        [string]
        $Text
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.event+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y events create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y events create $c8yargs
        }
        
    }

    End {}
}

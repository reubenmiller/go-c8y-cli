﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Alarm {
<#
.SYNOPSIS
Create alarm

.DESCRIPTION
Create an alarm on a device or agent.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_create

.EXAMPLE
PS> New-Alarm -Device $device.id -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR

Create a new alarm for device

.EXAMPLE
PS> Get-Device -Id $device.id | PSc8y\New-Alarm -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR

Create a new alarm for device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The ManagedObject that the alarm originated from
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Identifies the type of this alarm, e.g. 'com_cumulocity_events_TamperEvent'.
        [Parameter()]
        [string]
        $Type,

        # Time of the alarm. Defaults to current timestamp.
        [Parameter()]
        [string]
        $Time,

        # Text description of the alarm.
        [Parameter()]
        [string]
        $Text,

        # The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.
        [Parameter()]
        [ValidateSet('CRITICAL','MAJOR','MINOR','WARNING')]
        [string]
        $Severity,

        # The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.
        [Parameter()]
        [ValidateSet('ACTIVE','ACKNOWLEDGED','CLEARED')]
        [string]
        $Status
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.alarm+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y alarms create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y alarms create $c8yargs
        }
        
    }

    End {}
}

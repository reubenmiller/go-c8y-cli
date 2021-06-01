# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Alarm {
<#
.SYNOPSIS
Update alarm

.DESCRIPTION
Update an alarm by its id

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_update

.EXAMPLE
PS> Update-Alarm -Id $Alarm.id -Status ACKNOWLEDGED

Acknowledge an existing alarm

.EXAMPLE
PS> Get-Alarm -Id $Alarm.id | PSc8y\Update-Alarm -Status ACKNOWLEDGED

Acknowledge an existing alarm (using pipeline)

.EXAMPLE
PS> Update-Alarm -Id $Alarm.id -Severity CRITICAL

Update severity of an existing alarm to CRITICAL


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Alarm id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Comma separated alarm statuses, for example ACTIVE,CLEARED.
        [Parameter()]
        [ValidateSet('ACTIVE','ACKNOWLEDGED','CLEARED')]
        [string]
        $Status,

        # Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.
        [Parameter()]
        [ValidateSet('CRITICAL','MAJOR','MINOR','WARNING')]
        [string]
        $Severity,

        # Text description of the alarm.
        [Parameter()]
        [string]
        $Text
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.alarm+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y alarms update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y alarms update $c8yargs
        }
        
    }

    End {}
}

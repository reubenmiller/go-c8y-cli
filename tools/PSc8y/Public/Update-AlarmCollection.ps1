# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-AlarmCollection {
<#
.SYNOPSIS
Update alarm collection

.DESCRIPTION
Update the status of a collection of alarms by using a filter. Currently only the status of alarms can be changed

.LINK
c8y alarms updateCollection

.EXAMPLE
PS> Update-AlarmCollection -Device $Device.id -Status ACTIVE -NewStatus ACKNOWLEDGED

Update the status of all active alarms on a device to ACKNOWLEDGED

.EXAMPLE
PS> Get-Device -Id $Device.id | PSc8y\Update-AlarmCollection -Status ACTIVE -NewStatus ACKNOWLEDGED

Update the status of all active alarms on a device to ACKNOWLEDGED (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The ManagedObject that the alarm originated from
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.
        [Parameter()]
        [ValidateSet('ACTIVE','ACKNOWLEDGED','CLEARED')]
        [string]
        $Status,

        # The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.
        [Parameter()]
        [ValidateSet('CRITICAL','MAJOR','MINOR','WARNING')]
        [string]
        $Severity,

        # When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.
        [Parameter()]
        [switch]
        $Resolved,

        # Start date or date and time of alarm occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of alarm occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # New status to be applied to all of the matching alarms (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('ACTIVE','ACKNOWLEDGED','CLEARED')]
        [string]
        $NewStatus
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms updateCollection"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Device)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y alarms updateCollection $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y alarms updateCollection $c8yargs
        }
        
    }

    End {}
}

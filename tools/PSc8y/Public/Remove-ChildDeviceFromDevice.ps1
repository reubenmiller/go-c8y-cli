# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ChildDeviceFromDevice {
<#
.SYNOPSIS
Delete child device reference

.DESCRIPTION
Delete child device reference

.EXAMPLE
PS> Remove-ChildDeviceFromDevice -Device $Device.id -ChildDevice $ChildDevice.id

Unassign a child device from its parent device


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Device,

        # Child device reference (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $ChildDevice
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences unassignChildDevice"
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
            $items = (PSc8y\Expand-Id $ChildDevice)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences unassignChildDevice $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences unassignChildDevice $c8yargs
        }
        
    }

    End {}
}

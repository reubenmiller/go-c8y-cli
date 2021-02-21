# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildDeviceToDevice {
<#
.SYNOPSIS
Create a child device reference

.DESCRIPTION
Create a child device reference

.EXAMPLE
PS> Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

Assign a device as a child device to an existing device

.EXAMPLE
PS> Get-ManagedObject -Id $ChildDevice.id | Add-ChildDeviceToDevice -Device $Device.id

Assign a device as a child device to an existing device (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device. (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Device,

        # New child device (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChild
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Device")) {
            $Parameters["device"] = PSc8y\Expand-Id $Device
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences assignChildDevice"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $NewChild)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $NewChild `
            | c8y inventoryReferences assignChildDevice $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $NewChild `
            | c8y inventoryReferences assignChildDevice $c8yargs
        }
        
    }

    End {}
}

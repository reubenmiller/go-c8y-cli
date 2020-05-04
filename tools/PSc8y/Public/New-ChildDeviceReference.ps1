# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-ChildDeviceReference {
<#
.SYNOPSIS
Create a child device reference

.DESCRIPTION
Create a child device reference

.EXAMPLE
PS> New-ChildDeviceReference -Device $Device.id -NewChild $ChildDevice.id
Assign a device as a child device to an existing device

.EXAMPLE
PS> Get-ManagedObject -Id $ChildDevice.id | New-ChildDeviceReference -Device $Device.id
Assign a device as a child device to an existing device (using pipeline)


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

        # new child device (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChild,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Device")) {
            $Parameters["device"] = $Device
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

    }

    Process {
        foreach ($item in (PSc8y\Expand-Device $NewChild)) {
            if ($item) {
                $Parameters["newChild"] = if ($item.id) { $item.id } else { $item }
            }

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "inventoryReferences" `
                -Verb "createChildDevice" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.managedObjectReference+json" `
                -ItemType "application/vnd.com.nsn.cumulocity.managedObject+json" `
                -ResultProperty "managedObject" `
                -Raw:$Raw
        }
    }

    End {}
}

Function Expand-Device {
<#
.SYNOPSIS
Expand a list of devices replacing any ids or names with the actual device object.

.NOTES
If the given object is already an device object, then it is added with no additional lookup

.PARAMETER InputObject
List of ids, names or device objects

.EXAMPLE
Expand-Device "mydevice"

Retrieve the device objects by name or id

.EXAMPLE
Get-DeviceCollection *test* | Expand-Device

Get all the device object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.


#>
    [cmdletbinding(
        # SupportsShouldProcess = $true,
        # ConfirmImpact = "None"
    )]
    Param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject
    )

    Process {
        [array] $AllDevices = foreach ($iDevice in $InputObject)
        {
            if ($iDevice.id) {
                $iDevice
            } else {
                if ($iDevice -match "^\d+$") {
                    Get-ManagedObject -Id $iDevice
                } else {
                    Get-DeviceCollection -Name $iDevice -WhatIf:$false
                }
            }
        }

        $AllDevices
    }
}

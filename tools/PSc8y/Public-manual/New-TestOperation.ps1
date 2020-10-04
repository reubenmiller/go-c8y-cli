Function New-TestOperation {
<#
.SYNOPSIS
Create a new test operation

.DESCRIPTION
Create a test operation for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestOperation

Create a new test device and then create an operation on it

.EXAMPLE
New-TestOperation -Device "myExistingDevice"

Create an operation on the existing device "myExistingDevice"
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [object] $Device,

        # Don't prompt for confirmation
        [switch] $Force
    )

    if ($null -ne $Device) {
        $iAgent = Expand-Device $Device
    }
    else {
        $iAgent = PSc8y\New-TestAgent -Force:$Force
    }

    # Fake device (if whatif prevented it from being created)
    if ($WhatIfPreference -and $null -eq $iAgent) {
        $iAgent = @{ id = "12345" }
    }

    PSc8y\New-Operation `
        -Device $iAgent.id `
        -Description "Test operation" `
        -Data @{
        c8y_Restart = @{
                parameters = @{ }
            }
        } `
        -Force:$Force
}

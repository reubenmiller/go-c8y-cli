Function New-TestEvent {
<#
.SYNOPSIS
Create a new test event

.DESCRIPTION
Create a test event for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestEvent

Create a new test device and then create an event on it

.EXAMPLE
New-TestEvent -Device "myExistingDevice"

Create an event on the existing device "myExistingDevice"
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

        # Add a dummy file to the event
        [switch] $WithBinary,

        # Don't prompt for confirmation
        [switch] $Force
    )

    if ($null -ne $Device) {
        $iDevice = Expand-Device $Device
    } else {
        $iDevice = PSc8y\New-TestDevice -Force:$Force
    }

    # Fake device (if whatif prevented it from being created)
    if ($WhatIfPreference -and $null -eq $iDevice) {
        $iDevice = @{ id = "12345" }
    }

    $c8yEvent = PSc8y\New-Event `
        -Device $iDevice.id `
        -Time "1970-01-01" `
        -Type "c8y_ci_TestEvent" `
        -Text "Test CI Event" `
        -Force:$Force

    if ($WithBinary) {
        if ($WhatIfPreference -and $null -eq $iDevice) {
            $c8yEvent = @{ id = "12345" }
        }

        $tempfile = New-TemporaryFile
        "Cumulocity test content" | Out-File -LiteralPath $tempfile
        $null = PSc8y\New-EventBinary `
            -Id $c8yEvent.id `
            -File $tempfile `
            -Force:$Force

        Remove-Item $tempfile
    }

    $c8yEvent
}

Function New-TestAlarm {
<#
.SYNOPSIS
Create a new test alarm

.DESCRIPTION
Create a test alarm for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestAlarm

Create a new test device and then create an alarm on it

.EXAMPLE
New-TestAlarm -Device "myExistingDevice"

Create an alarm on the existing device "myExistingDevice"
#>
    [cmdletbinding()]
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
        $iDevice = Expand-Device $Device
    } else {
        $iDevice = PSc8y\New-TestDevice -Force:$Force
    }

    PSc8y\New-Alarm `
        -Device $iDevice.id `
        -Time "1970-01-01" `
        -Type "c8y_ci_TestAlarm" `
        -Severity MAJOR `
        -Text "Test CI Alarm" `
        -Force:$Force
}

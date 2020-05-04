Function New-TestAlarm {
<#
.SYNOPSIS
Create a new test alarm
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [object] $Device,

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

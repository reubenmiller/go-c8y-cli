Function New-TestMeasurement {
<#
.SYNOPSIS
Create a new test measurement
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "None"
    )]
    Param(
        [object] $Device,

        # Value fragment type
        [string] $ValueFragmentType = "c8y_Temperature",

        # Value fragment series
        [string] $ValueFragmentSeries = "T",

        # Type
        [string] $Type = "C8yTemperatureReading",

        # Value
        [Double] $Value = 1.2345,

        # Unit. i.e. °C, m/s
        [string] $Unit = "°C",

        [switch] $Force
    )

    if ($null -eq $Device) {
        $iDevice = PSc8y\New-TestDevice -WhatIf:$false -Force:$Force
    } else {
        $iDevice = PSc8y\Expand-Device $Device
    }

    PSc8y\New-Measurement `
        -Device $iDevice.id `
        -Time "1970-01-01" `
        -Type $Type `
        -Data @{
            $ValueFragmentType = @{
                $ValueFragmentSeries = @{
                    value = $Value
                    unit = $Unit
                }
            }
        } `
        -Force:$Force
}

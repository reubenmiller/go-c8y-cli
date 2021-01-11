Function New-TestMeasurement {
<#
.SYNOPSIS
Create a new test measurement

.DESCRIPTION
Create a test measurement for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestMeasurement

Create a new test device and then create a measurement on it

.EXAMPLE
New-TestMeasurement -Device "myExistingDevice"

Create a measurement on the existing device "myExistingDevice"
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
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

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP")]
        [string]
        $ProcessingMode,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Don't prompt for confirmation
        [switch] $Force
    )

    Process {
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
            -ProcessingMode:$ProcessingMode `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force
    }
}

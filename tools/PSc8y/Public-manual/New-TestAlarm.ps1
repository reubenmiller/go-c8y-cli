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
        if ($null -ne $Device) {
            $iDevice = Expand-Device $Device
        } else {
            $iDevice = PSc8y\New-TestDevice -Force:$Force
        }

        # Fake device (if whatif prevented it from being created)
        if ($WhatIfPreference -and $null -eq $iDevice) {
            $iDevice = @{ id = "12345" }
        }

        if ($iDevice.id) {
            PSc8y\New-Alarm `
                -Device $iDevice.id `
                -Time "1970-01-01" `
                -Type "c8y_ci_TestAlarm" `
                -Severity MAJOR `
                -Text "Test CI Alarm" `
                -ProcessingMode:$ProcessingMode `
                -Template:$Template `
                -TemplateVars:$TemplateVars `
                -Force:$Force
        }
    }
}

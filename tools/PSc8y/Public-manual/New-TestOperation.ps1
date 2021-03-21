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
    [cmdletbinding()]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [object] $Device
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        $commonOptions = @{} + $PSBoundParameters
        $commonOptions.Remove("Device")

        if ($null -ne $Device) {
            $iAgent = Expand-Device $Device
        }
        else {
            $iAgent = PSc8y\New-TestAgent @commonOptions
        }

        # Fake device (if whatif prevented it from being created)
        if ($Dry -and $null -eq $iAgent) {
            $iAgent = @{ id = "12345" }
        }

        $options = @{} + $PSBoundParameters
        $options["Device"] = $iAgent.id
        $options["Description"] = "Test operation"

        if ($null -eq $options["Data"]) {
            $options["Data"] = @{}
        }
        $options["Data"].c8y_Restart = @{ parameters = @{ } }

        PSc8y\New-Operation @options
    }
}

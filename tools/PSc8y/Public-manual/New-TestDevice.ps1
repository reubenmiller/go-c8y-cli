Function New-TestDevice {
<# 
.SYNOPSIS
Create a new test device representation in Cumulocity

.DESCRIPTION
Create a new test device with a randomized name. Useful when performing mockups or prototyping.

The agent will have both the `c8y_IsDevice` fragments set.

.EXAMPLE
New-TestDevice

Create a test device

.EXAMPLE
1..10 | Foreach-Object { New-TestDevice -Force }

Create 10 test devices all with unique names

.EXAMPLE
1..10 | Foreach-Object { New-TestDevice -AsAgent -Force }

Create 10 test devices (with agent functionality) all with unique names

#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Device name prefix which is added before the randomized string
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testdevice",

        # Add agent fragment to the device
        [switch] $AsAgent,

        # Don't prompt for confirmation
        [switch] $Force
    )
    $Data = @{
        c8y_IsDevice = @{}
    }
    if ($AsAgent) {
        $Data.com_cumulocity_model_Agent = @{}
    }
    $DeviceName = New-RandomString -Prefix "${Name}_"
    $TestDevice = PSc8y\New-ManagedObject `
        -Name $DeviceName `
        -Data $Data `
        -Force

    $TestDevice
}

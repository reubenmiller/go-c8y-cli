Function New-TestDeviceGroup {
<# 
.SYNOPSIS
Create a new test device group

.DESCRIPTION
Create a new test device group with a randomized name. Useful when performing mockups or prototyping.

.EXAMPLE
New-TestDeviceGroup

Create a test device group

.EXAMPLE
1..10 | Foreach-Object { New-TestDeviceGroup -Force }

Create 10 test device groups all with unique names
#>
    [cmdletbinding()]
    Param(
        # Device group name prefix which is added before the randomized string
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testgroup",

        # Group type. Only device groups of type `Group` are visible as root folders in the UI
        [ValidateSet("Group", "SubGroup")]
        [string] $Type = "Group",

        # Don't prompt for confirmation
        [switch] $Force
    )
    $Data = @{
        c8y_IsDeviceGroup = @{ }
    }

    switch ($Type) {
        "SubGroup" {
            $Data.type = "c8y_DeviceSubGroup"
            break;
        }
        default {
            $Data.type = "c8y_DeviceGroup"
            break;
        }
    }

    $GroupName = New-RandomString -Prefix "${Name}_"
    PSc8y\New-ManagedObject `
        -Name $GroupName `
        -Data $Data `
        -Force:$Force
}

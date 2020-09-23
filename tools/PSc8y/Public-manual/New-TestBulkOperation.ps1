Function New-TestBulkOperation {
<#
.SYNOPSIS
Create a new test bulk operation

.DESCRIPTION
Create a test bulk operation for devices in a specific group.

If the group does not exist, then the group and devices will be created automatically

.EXAMPLE
New-TestBulkOperation

Create a new test bulk operation for a group of devices. The group and a list of devices
will be automatically created.

.EXAMPLE
New-TestBulkOperation -Group "myExistingDevice"

Create an operation on the existing device group "myExistingDevice"
#>
    [cmdletbinding(
        DefaultParameterSetName = "new-group-and-devices"
    )]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            ParameterSetName = "existing-device",
            Mandatory = $false,
            Position = 0
        )]
        [object] $Device,

        [Parameter(
            ParameterSetName = "existing-group",
            Mandatory = $true,
            Position = 0
        )]
        [object] $Group,

        # Don't prompt for confirmation
        [switch] $Force
    )

    $Group = PSc8y\New-TestDeviceGroup -TotalDevices 5

    PSc8y\New-BulkOperation `
        -StartDate "10s" `
        -CreationRampSec 10 `
        -Group $Group.id `
        -Operation @{
            c8y_Restart = @{
                parameters = @{ }
            }
        } `
        -Force:$Force
}

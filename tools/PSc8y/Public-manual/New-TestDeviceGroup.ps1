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

.EXAMPLE
New-TestDeviceGroup -TotalDevices 10

Create a test device group with 10 newly created devices
#>
    [cmdletbinding()]
    Param(
        # Device group name prefix which is added before the randomized string
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testgroup",

        # Group type. Only device groups of type `Group` are visible as root folders in the UI
        [ValidateSet("Group", "SubGroup")]
        [string] $Type = "Group",

        # Number of devices to create and assign to the group
        [int]
        $TotalDevices = 0
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        $options = @{} + $PSBoundParameters
        $options.Remove("Name")
        $options.Remove("Type")
        $options.Remove("TotalDevices")

        
        $TypeName = ""
        switch ($Type) {
            "SubGroup" {
                $TypeName = "c8y_DeviceSubGroup"
                break;
            }
            default {
                $TypeName = "c8y_DeviceGroup"
                break;
            }
        }

        $Data = @{
            c8y_IsDeviceGroup = @{ }
            type = $TypeName
        }

        $GroupName = New-RandomString -Prefix "${Name}_"
        $options["Name"] = $GroupName
        $options["Type"] = $TypeName
        $options["Data"] = $Data
        $Group = PSc8y\New-ManagedObject @options 
        
        if ($TotalDevices -gt 0) {
            for ($i = 0; $i -lt $TotalDevices; $i++) {
                $iDevice = PSc8y\New-TestAgent -Force -AsPSObject
                $null = PSc8y\Add-AssetToGroup -Group $Group.id -ChildDevice $iDevice.id -Force
            }
        }

        Write-Output $Group
    }
}

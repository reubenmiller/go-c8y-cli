Function New-TestDeviceGroup {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testgroup",

        [ValidateSet("Group", "SubGroup")]
        [string] $Type = "Group",

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

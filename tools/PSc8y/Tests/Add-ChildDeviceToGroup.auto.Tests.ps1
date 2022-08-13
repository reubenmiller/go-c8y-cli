. $PSScriptRoot/imports.ps1

Describe -Name "Add-ChildDeviceToGroup" {
    BeforeEach {

    }

    It -Skip "Add a device to a group" {
        $Response = PSc8y\Add-ChildDeviceToGroup -Group $Group.id -Child $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Add a device to a group by passing device and groups instead of an id or name" {
        $Response = PSc8y\Add-ChildDeviceToGroup -Group $Group -Child $Device
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Add multiple devices to a group. Alternatively `Get-DeviceCollection` can be used
to filter for a collection of devices and assign the results to a single group.
" {
        $Response = PSc8y\Get-Device $Device1.name, $Device2.name | Add-ChildDeviceToGroup -Group $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


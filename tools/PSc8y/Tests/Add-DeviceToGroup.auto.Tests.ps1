. $PSScriptRoot/imports.ps1

Describe -Name "Add-DeviceToGroup" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Group = PSc8y\New-TestDeviceGroup
        $Device1 = PSc8y\New-TestDevice
        $Device2 = PSc8y\New-TestDevice

    }

    It "Add a device to a group" {
        $Response = PSc8y\Add-DeviceToGroup -Group $Group.id -NewChildDevice $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add multiple devices to a group. Alternatively `Get-DeviceCollection` can be used
to filter for a collection of devices and assign the results to a single group.
" {
        $Response = PSc8y\Get-Device $Device1.name, $Device2.name | Add-DeviceToGroup -Group $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $Group.id
        PSc8y\Remove-ManagedObject -Id $Device1.id
        PSc8y\Remove-ManagedObject -Id $Device2.id

    }
}


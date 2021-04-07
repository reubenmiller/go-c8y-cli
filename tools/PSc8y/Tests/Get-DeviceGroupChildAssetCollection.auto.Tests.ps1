. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildAssetCollection" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-AssetToGroup -Group $Device.id -NewChildDevice $ChildDevice.id
        $Group = PSc8y\New-TestDeviceGroup
        $ChildGroup = PSc8y\New-TestDeviceGroup
        PSc8y\Add-AssetToGroup -Group $Group.id -NewChildGroup $ChildGroup.id

    }

    It "Get a list of the child assets of an existing device" {
        $Response = PSc8y\Get-DeviceGroupChildAssetCollection -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of the child assets of an existing group" {
        $Response = PSc8y\Get-DeviceGroupChildAssetCollection -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildDevice.id
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $ChildGroup.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}


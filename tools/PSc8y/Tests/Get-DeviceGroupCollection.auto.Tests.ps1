. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupCollection" {
    BeforeEach {
        $group = PSc8y\New-TestDeviceGroup -Name "parent"

    }

    It "Get a collection of device groups with names that start with 'parent'" {
        $Response = PSc8y\Get-DeviceGroupCollection -Name "parent*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $group.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceGroup" {
    BeforeEach {
        $group = PSc8y\New-TestDeviceGroup

    }

    It "Update device group by id" {
        $Response = PSc8y\Update-DeviceGroup -Id $group.id -Name "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update device group by name" {
        $Response = PSc8y\Update-DeviceGroup -Id $group.name -Name "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update device group custom properties" {
        $Response = PSc8y\Update-DeviceGroup -Id $group.name -Data @{ "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $group.id

    }
}


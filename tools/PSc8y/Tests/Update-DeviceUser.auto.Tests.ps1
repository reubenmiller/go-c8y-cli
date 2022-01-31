. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceUser" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Get device user by id" {
        $Response = PSc8y\Update-DeviceUser -Id $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get device user by name" {
        $Response = PSc8y\Update-DeviceUser -Id $device.name -Enabled
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceUser" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Get device user by id" {
        $Response = PSc8y\Get-DeviceUser -Id $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get device user by name" {
        $Response = PSc8y\Get-DeviceUser -Id $device.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


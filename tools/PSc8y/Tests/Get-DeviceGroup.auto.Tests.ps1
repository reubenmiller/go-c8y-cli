. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroup" {
    BeforeEach {
        $group = PSc8y\New-TestDeviceGroup

    }

    It "Get device group by id" {
        $Response = PSc8y\Get-DeviceGroup -Id $group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get device group by name" {
        $Response = PSc8y\Get-DeviceGroup -Id $group.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $group.id

    }
}


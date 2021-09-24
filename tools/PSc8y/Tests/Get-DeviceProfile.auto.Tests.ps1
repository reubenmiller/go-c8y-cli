. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceProfile" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Get a device profile" {
        $Response = PSc8y\Get-DeviceProfile -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a device profile (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-DeviceProfile
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}


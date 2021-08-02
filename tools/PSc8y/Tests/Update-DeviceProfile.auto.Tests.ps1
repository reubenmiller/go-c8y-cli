. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceProfile" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Update a device profile" {
        $Response = PSc8y\Update-DeviceProfile -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a device profile (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Update-DeviceProfile -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceProfile" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Delete a device profile" {
        $Response = PSc8y\Remove-DeviceProfile -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a device profile (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-DeviceProfile
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceAvailability" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice

    }

    It "Get a device's availability by id" {
        $Response = PSc8y\Get-DeviceAvailability -Id $Device.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Get a device's availability by name" {
        $Response = PSc8y\Get-DeviceAvailability -Id $Device.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Set-DeviceRequiredAvailability" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Set the required availability of a device by name to 10 minutes" {
        $Response = PSc8y\Set-DeviceRequiredAvailability -Device $device.id -Interval 10
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Set the required availability of a device (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $device.id | PSc8y\Set-DeviceRequiredAvailability -Interval 10
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-SupportedSeries" {
    BeforeEach {
        $device = PSc8y\New-TestDevice
        $Measurement = PSc8y\New-TestMeasurement -Device $device.id

    }

    It "Get the supported measurement series of a device by name" {
        $Response = PSc8y\Get-SupportedSeries -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get the supported measurement series of a device (using pipeline)" {
        $Response = PSc8y\Get-SupportedSeries -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


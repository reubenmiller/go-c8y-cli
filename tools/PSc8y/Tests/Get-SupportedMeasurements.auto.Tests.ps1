. $PSScriptRoot/imports.ps1

Describe -Name "Get-SupportedMeasurements" {
    BeforeEach {
        $device = PSc8y\New-TestDevice
        $Measurement = PSc8y\New-Measurement -Template "test.measurement.jsonnet" -Device $device.id

    }

    It "Get the supported measurements of a device by name" {
        $Response = PSc8y\Get-SupportedMeasurements -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get the supported measurements of a device (using pipeline)" {
        $Response = PSc8y\Get-SupportedMeasurements -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


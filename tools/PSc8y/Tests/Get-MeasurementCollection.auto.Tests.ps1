. $PSScriptRoot/imports.ps1

Describe -Name "Get-MeasurementCollection" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Measurement = New-Measurement -Template "test.measurement.jsonnet" -Device $Device.id -Type "TempReading"

    }

    It "Get a list of measurements" {
        $Response = PSc8y\Get-MeasurementCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of measurements for a particular device" {
        $Response = PSc8y\Get-MeasurementCollection -Device $Device.id -Type "TempReading"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get measurements from a device (using pipeline)" {
        $Response = PSc8y\Get-DeviceCollection -Name $Device.name | Get-MeasurementCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


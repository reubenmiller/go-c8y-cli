. $PSScriptRoot/imports.ps1

Describe -Name "Get-MeasurementSeries" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Measurement = New-Measurement -Template "{c8y_Temperature:{T:{value:_.Int(),unit:'°C'}}}" -Device $Device.id -Type "TempReading"
        $Measurement2 = New-Measurement -Template "{c8y_Temperature:{T:{value:_.Int(),unit:'°C'}}}" -Device $Device.id -Type "TempReading"

    }

    It "Get a list of measurements for a particular device" {
        $Response = PSc8y\Get-MeasurementSeries -Device $Device.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get measurement series c8y_Temperature.T on a device" {
        $Response = PSc8y\Get-MeasurementSeries -Device $Measurement2.source.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get measurement series from a device (using pipeline)" {
        $Response = PSc8y\Get-DeviceCollection -Name $Device.name | Get-MeasurementSeries -Series "c8y_Temperature.T"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $Measurement2.source.id

    }
}


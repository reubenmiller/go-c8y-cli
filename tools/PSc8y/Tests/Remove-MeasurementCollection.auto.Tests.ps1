. $PSScriptRoot/imports.ps1

Describe -Name "Remove-MeasurementCollection" {
    BeforeEach {
        $Measurement = New-TestDevice | New-Measurement -Template "test.measurement.jsonnet"

    }

    It -Skip "Delete measurements for a device" {
        $Response = PSc8y\Remove-MeasurementCollection -Device $Measurement.source.id
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete measurements older than 10 days for a device" {
        $Response = PSc8y\Remove-MeasurementCollection -Device $Measurement.source.id -DateTo "-10d"
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete measurements with a given fragment and older than 10 days for a device" {
        $Response = PSc8y\Remove-MeasurementCollection -Device $Measurement.source.id -DateTo "-10d" -FragmentType lmp
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Measurement.source.id

    }
}


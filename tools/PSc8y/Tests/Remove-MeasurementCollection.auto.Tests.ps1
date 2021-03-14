. $PSScriptRoot/imports.ps1

Describe -Name "Remove-MeasurementCollection" {
    BeforeEach {
        $Measurement = New-TestDevice | New-Measurement -Template "test.measurement.jsonnet"

    }

    It "Delete measurement collection for a device" {
        $Response = PSc8y\Remove-MeasurementCollection -Device $Measurement.source.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Measurement.source.id

    }
}


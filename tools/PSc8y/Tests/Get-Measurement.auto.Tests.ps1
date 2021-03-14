. $PSScriptRoot/imports.ps1

Describe -Name "Get-Measurement" {
    BeforeEach {
        $Measurement = New-TestDevice | New-Measurement -Template "test.measurement.jsonnet"

    }

    It "Get measurement" {
        $Response = PSc8y\Get-Measurement -Id $Measurement.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Measurement.source.id

    }
}


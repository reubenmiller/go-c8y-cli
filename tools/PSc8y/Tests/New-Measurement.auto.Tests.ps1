. $PSScriptRoot/imports.ps1

Describe -Name "New-Measurement" {
    BeforeEach {
        $TestDevice = PSc8y\New-TestDevice

    }

    It "Create measurement" {
        $Response = PSc8y\New-Measurement -Device $TestDevice.id -Time "0s" -Type "myType" -Data @{ c8y_Winding = @{ temperature = @{ value = 25.0; unit = "°C" } } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestDevice.id) {
            PSc8y\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue
        }

    }
}


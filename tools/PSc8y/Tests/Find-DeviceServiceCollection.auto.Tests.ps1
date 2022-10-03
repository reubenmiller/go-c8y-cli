. $PSScriptRoot/imports.ps1

Describe -Name "Find-DeviceServiceCollection" {
    BeforeEach {

    }

    It -Skip "Find all services (from any device)" {
        $Response = PSc8y\Find-DeviceServiceCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


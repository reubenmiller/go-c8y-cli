. $PSScriptRoot/imports.ps1

Describe -Name "Find-DeviceSoftwareCollection" {
    BeforeEach {

    }

    It -Skip "Find all software (from a device)" {
        $Response = PSc8y\Find-DeviceSoftwareCollection -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


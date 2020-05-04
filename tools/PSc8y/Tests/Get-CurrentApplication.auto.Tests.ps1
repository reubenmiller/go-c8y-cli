. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentApplication" {
    BeforeEach {

    }

    It -Skip "Get the current application (requires using application credentials)" {
        $Response = PSc8y\Get-CurrentApplication
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


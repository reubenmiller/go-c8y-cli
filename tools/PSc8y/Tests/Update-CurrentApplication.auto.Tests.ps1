. $PSScriptRoot/imports.ps1

Describe -Name "Update-CurrentApplication" {
    BeforeEach {

    }

    It -Skip "Update custom properties of the current application (requires using application credentials)" {
        $Response = PSc8y\Update-CurrentApplication -Data @{ myCustomProp = @{ value1 = 1}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


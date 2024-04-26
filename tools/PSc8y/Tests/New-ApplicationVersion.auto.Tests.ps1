. $PSScriptRoot/imports.ps1

Describe -Name "New-ApplicationVersion" {
    BeforeEach {

    }

    It -Skip "Create a new application version" {
        $Response = PSc8y\New-ApplicationVersion -Application 1234 -File ./myapp.zip -Version "2.0.0"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


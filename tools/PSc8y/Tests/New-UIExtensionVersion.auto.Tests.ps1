. $PSScriptRoot/imports.ps1

Describe -Name "New-UIExtensionVersion" {
    BeforeEach {

    }

    It -Skip "Create a new version for an extension" {
        $Response = PSc8y\New-UIExtensionVersion -Extension 1234 -File ./myapp.zip -Version "2.0.0"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


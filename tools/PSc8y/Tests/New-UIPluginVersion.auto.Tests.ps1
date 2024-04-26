. $PSScriptRoot/imports.ps1

Describe -Name "New-UIPluginVersion" {
    BeforeEach {

    }

    It -Skip "Create a new version for a plugin" {
        $Response = PSc8y\New-UIPluginVersion -Plugin 1234 -File ./myapp.zip -Version "2.0.0"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


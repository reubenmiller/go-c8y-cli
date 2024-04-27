. $PSScriptRoot/imports.ps1

Describe -Name "Update-UIPluginVersion" {
    BeforeEach {

    }

    It -Skip "Replace tags assigned to a version of a plugin" {
        $Response = PSc8y\Update-UIPluginVersion -Plugin 1234 -Version 1.0 -Tags tag1,latest
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


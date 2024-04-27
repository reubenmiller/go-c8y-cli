. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIPluginVersionCollection" {
    BeforeEach {

    }

    It -Skip "Get plugin versions" {
        $Response = PSc8y\Get-UIPluginVersionCollection -Plugin 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


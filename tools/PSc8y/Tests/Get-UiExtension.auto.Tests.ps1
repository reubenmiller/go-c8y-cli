. $PSScriptRoot/imports.ps1

Describe -Name "Get-UiExtension" {
    BeforeEach {

    }

    It "Get ui extension" {
        $Response = PSc8y\Get-UiExtension -Id 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


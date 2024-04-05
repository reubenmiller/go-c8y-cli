. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIExtensionVersionCollection" {
    BeforeEach {

    }

    It "Get extension versions" {
        $Response = PSc8y\Get-UIExtensionVersionCollection -Extension 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


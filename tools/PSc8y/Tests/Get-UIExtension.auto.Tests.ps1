. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIExtension" {
    BeforeEach {

    }

    It "Get ui extension" {
        $Response = PSc8y\Get-UIExtension -Id 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


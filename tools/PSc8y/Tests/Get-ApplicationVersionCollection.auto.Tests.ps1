. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationVersionCollection" {
    BeforeEach {

    }

    It "Get application versions" {
        $Response = PSc8y\Get-ApplicationVersionCollection -Application 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationCollection" {
    BeforeEach {

    }

    It "Get applications" {
        $Response = PSc8y\Get-ApplicationCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


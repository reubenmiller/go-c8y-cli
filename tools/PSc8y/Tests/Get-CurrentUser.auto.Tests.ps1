. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentUser" {
    BeforeEach {

    }

    It "Get the current user" {
        $Response = PSc8y\Get-CurrentUser
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


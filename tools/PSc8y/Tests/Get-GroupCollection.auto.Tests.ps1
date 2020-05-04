. $PSScriptRoot/imports.ps1

Describe -Name "Get-GroupCollection" {
    BeforeEach {

    }

    It "Get a list of user groups for the current tenant" {
        $Response = PSc8y\Get-GroupCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


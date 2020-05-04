. $PSScriptRoot/imports.ps1

Describe -Name "Get-RoleCollection" {
    BeforeEach {

    }

    It "Get a list of roles" {
        $Response = PSc8y\Get-RoleCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


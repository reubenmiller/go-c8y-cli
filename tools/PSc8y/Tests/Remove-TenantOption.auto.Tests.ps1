. $PSScriptRoot/imports.ps1

Describe -Name "Remove-TenantOption" {
    BeforeEach {
        New-TenantOption -Category "c8y_cli_tests" -Key "option3" -Value "3"

    }

    It "Delete a tenant option" {
        $Response = PSc8y\Remove-TenantOption -Category "c8y_cli_tests" -Key "option3"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


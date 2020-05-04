. $PSScriptRoot/imports.ps1

Describe -Name "Update-TenantOption" {
    BeforeEach {
        New-TenantOption -Category "c8y_cli_tests" -Key "option4" -Value "4"

    }

    It "Update a tenant option" {
        $Response = PSc8y\Update-TenantOption -Category "c8y_cli_tests" -Key "option4" -Value "0"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "option4"

    }
}


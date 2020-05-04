. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantOption" {
    BeforeEach {
        New-TenantOption -Category "c8y_cli_tests" -Key "option2" -Value "2"

    }

    It "Get a tenant option" {
        $Response = PSc8y\Get-TenantOption -Category "c8y_cli_tests" -Key "option2"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "option2"

    }
}


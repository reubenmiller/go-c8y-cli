. $PSScriptRoot/imports.ps1

Describe -Name "Remove-TenantOption" {
    BeforeEach {
        $option3 = New-RandomString -Prefix "option3"
        New-TenantOption -Category "c8y_cli_tests" -Key "$option3" -Value "3"

    }

    It "Delete a tenant option" {
        $Response = PSc8y\Remove-TenantOption -Category "c8y_cli_tests" -Key "$option3"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Update-TenantOptionBulk" {
    BeforeEach {
        $option5 = New-RandomString -Prefix "option5"
        $option6 = New-RandomString -Prefix "option6"
        New-TenantOption -Category "c8y_cli_tests" -Key "$option5" -Value "5"
        New-TenantOption -Category "c8y_cli_tests" -Key "$option6" -Value "6"

    }

    It "Update multiple tenant options" {
        $Response = PSc8y\Update-TenantOptionBulk -Category "c8y_cli_tests" -Data @{ $option5 = 0; $option6 = 1 }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "$option5"
        Remove-TenantOption -Category "c8y_cli_tests" -Key "$option6"

    }
}


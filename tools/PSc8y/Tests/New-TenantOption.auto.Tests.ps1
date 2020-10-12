. $PSScriptRoot/imports.ps1

Describe -Name "New-TenantOption" {
    BeforeEach {
        $option1 = New-RandomString -Prefix "option1"

    }

    It "Create a tenant option" {
        $Response = PSc8y\New-TenantOption -Category "c8y_cli_tests" -Key "$option1" -Value "1"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "$option1"

    }
}


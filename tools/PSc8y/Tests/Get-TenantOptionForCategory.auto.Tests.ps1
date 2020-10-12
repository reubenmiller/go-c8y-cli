. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantOptionForCategory" {
    BeforeEach {
        $option7 = New-RandomString -Prefix "option7"
        New-TenantOption -Category "c8y_cli_tests" -Key "$option7" -Value "7"

    }

    It "Get a list of options for a category" {
        $Response = PSc8y\Get-TenantOptionForCategory -Category "c8y_cli_tests"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "$option7"

    }
}


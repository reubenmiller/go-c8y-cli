. $PSScriptRoot/imports.ps1

Describe -Name "Update-TenantOptionEditable" {
    BeforeEach {
        $option8 = New-RandomString -Prefix "option8"
        New-TenantOption -Category "c8y_cli_tests" -Key "$option8" -Value "8"

    }

    It -Skip "Update editable property for an existing tenant option" {
        $Response = PSc8y\Update-TenantOptionEditable -Category "c8y_cli_tests" -Key "$option8" -Editable "true"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-TenantOption -Category "c8y_cli_tests" -Key "$option8"

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantTFASetting" {
    BeforeEach {

    }

    It -Skip "Get the Two-Factor-Authentication setting of a tenant" {
        $Response = PSc8y\Get-TenantTFASetting
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Update-TenantTFASetting" {
    BeforeEach {

    }

    It -Skip "Update the Tenant's TFA setting of the current tenant" {
        $Response = PSc8y\Update-TenantTFASetting -Strategy TOTP
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Update the Tenant's TFA setting to use time based one-time-password" {
        $Response = PSc8y\Update-TenantTFASetting -Tenant t12345 -Strategy TOTP
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


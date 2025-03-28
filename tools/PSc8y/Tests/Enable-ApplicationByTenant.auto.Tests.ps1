. $PSScriptRoot/imports.ps1

Describe -Name "Enable-ApplicationByTenant" {
    BeforeEach {

    }

    It -Skip "Enable an application of a tenant" {
        $Response = PSc8y\Enable-ApplicationByTenant -Tenant t12345 -Application myMicroservice
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Disable-ApplicationByTenant" {
    BeforeEach {

    }

    It -Skip "Disable an application of a tenant" {
        $Response = PSc8y\Disable-ApplicationByTenant -Tenant t12345 -Application myMicroservice
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


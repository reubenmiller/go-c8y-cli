. $PSScriptRoot/imports.ps1

Describe -Name "Disable-FeatureByTenant" {
    BeforeEach {

    }

    It -Skip "Disable a feature in the current tenant" {
        $Response = PSc8y\Disable-FeatureByTenant -Id example -Active -Tenant t12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


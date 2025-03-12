. $PSScriptRoot/imports.ps1

Describe -Name "Update-FeatureByTenant" {
    BeforeEach {

    }

    It -Skip "Enable a feature for a given specific tenant" {
        $Response = PSc8y\Update-FeatureByTenant -Id example -Active -Tenant t12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Enable-FeatureByTenant" {
    BeforeEach {

    }

    It -Skip "Enable a feature for a given tenant" {
        $Response = PSc8y\Enable-FeatureByTenant -Id example -Active -Tenant t12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


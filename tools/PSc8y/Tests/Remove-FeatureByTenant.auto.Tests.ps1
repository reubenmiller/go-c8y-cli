. $PSScriptRoot/imports.ps1

Describe -Name "Remove-FeatureByTenant" {
    BeforeEach {

    }

    It -Skip "Remove the feature override from a given tenant" {
        $Response = PSc8y\Remove-FeatureByTenant -Id example -Tenant t12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


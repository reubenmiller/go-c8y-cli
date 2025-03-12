. $PSScriptRoot/imports.ps1

Describe -Name "Get-FeatureByTenant" {
    BeforeEach {

    }

    It -Skip "Get the feature toggle override status for each child tenant" {
        $Response = PSc8y\Get-FeatureByTenant -Id example
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


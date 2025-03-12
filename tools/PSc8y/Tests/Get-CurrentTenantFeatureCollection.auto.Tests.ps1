. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentTenantFeatureCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of features for the current tenant" {
        $Response = PSc8y\Get-CurrentTenantFeatureCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


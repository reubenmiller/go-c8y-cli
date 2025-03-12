. $PSScriptRoot/imports.ps1

Describe -Name "Remove-CurrentTenantFeature" {
    BeforeEach {

    }

    It -Skip "Remove the feature override for the current tenant" {
        $Response = PSc8y\Remove-CurrentTenantFeature -Key example
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


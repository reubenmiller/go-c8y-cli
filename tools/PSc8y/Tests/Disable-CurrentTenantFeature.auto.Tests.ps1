. $PSScriptRoot/imports.ps1

Describe -Name "Disable-CurrentTenantFeature" {
    BeforeEach {

    }

    It -Skip "Disable a feature in the current tenant" {
        $Response = PSc8y\Disable-CurrentTenantFeature -Key example -Active
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


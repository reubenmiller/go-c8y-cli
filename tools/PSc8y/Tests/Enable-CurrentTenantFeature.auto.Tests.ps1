. $PSScriptRoot/imports.ps1

Describe -Name "Enable-CurrentTenantFeature" {
    BeforeEach {

    }

    It -Skip "Enable a feature in the current tenant" {
        $Response = PSc8y\Enable-CurrentTenantFeature -Key example -Active
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Update-CurrentTenantFeature" {
    BeforeEach {

    }

    It -Skip "Enable a feature in the current tenant" {
        $Response = PSc8y\Update-CurrentTenantFeature -Key example -Active
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


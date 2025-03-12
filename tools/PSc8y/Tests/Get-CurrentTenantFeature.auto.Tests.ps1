. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentTenantFeature" {
    BeforeEach {

    }

    It -Skip "Get a specific feature status in the current tenant" {
        $Response = PSc8y\Get-CurrentTenantFeature -Key example
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of tenants" {
        $Response = PSc8y\Get-TenantCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


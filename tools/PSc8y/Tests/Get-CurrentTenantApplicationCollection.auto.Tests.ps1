. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentTenantApplicationCollection" {
    BeforeEach {

    }

    It "Get a list of applications in the current tenant" {
        $Response = PSc8y\Get-CurrentTenantApplicationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


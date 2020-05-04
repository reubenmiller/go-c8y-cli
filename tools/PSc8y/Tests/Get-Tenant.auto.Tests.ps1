. $PSScriptRoot/imports.ps1

Describe -Name "Get-Tenant" {
    BeforeEach {

    }

    It -Skip "Get a tenant by name (from the management tenant)" {
        $Response = PSc8y\Get-Tenant -Id mycompany
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


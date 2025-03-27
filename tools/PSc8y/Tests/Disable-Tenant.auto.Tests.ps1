. $PSScriptRoot/imports.ps1

Describe -Name "Disable-Tenant" {
    BeforeEach {

    }

    It -Skip "Disable a tenant (from the management tenant)" {
        $Response = PSc8y\Disable-Tenant -Id mycompany
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


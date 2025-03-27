. $PSScriptRoot/imports.ps1

Describe -Name "Enable-Tenant" {
    BeforeEach {

    }

    It -Skip "Enable a tenant (from the management tenant)" {
        $Response = PSc8y\Enable-Tenant -Id mycompany
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Tenant" {
    BeforeEach {

    }

    It -Skip "Delete a tenant by name (from the management tenant)" {
        $Response = PSc8y\Remove-Tenant -Id mycompany
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


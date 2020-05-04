. $PSScriptRoot/imports.ps1

Describe -Name "Update-Tenant" {
    BeforeEach {

    }

    It -Skip "Update a tenant by name (from the mangement tenant)" {
        $Response = PSc8y\Update-Tenant -Id mycompany -ContactName "John Smith"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "New-Tenant" {
    BeforeEach {

    }

    It -Skip "Create a new tenant (from the management tenant)" {
        $Response = PSc8y\New-Tenant -Company "mycompany" -Domain "mycompany" -AdminName "admin" -Password "mys3curep9d8"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


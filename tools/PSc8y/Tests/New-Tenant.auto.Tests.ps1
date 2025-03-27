. $PSScriptRoot/imports.ps1

Describe -Name "New-Tenant" {
    BeforeEach {

    }

    It -Skip "Create a new tenant (from the management tenant)" {
        $Response = PSc8y\New-Tenant -Name "mycompany" -Domain "mycompany" -AdminEmail "admin@example.com" -AdminName "admin" -AdminPass "mys3curep9d8"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Create a new tenant and send a password reset email (from the management tenant)" {
        $Response = PSc8y\New-Tenant -Name "mycompany" -Domain "mycompany" -AdminEmail "admin@example.com" -AdminName "admin" -SendPasswordResetEmail
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


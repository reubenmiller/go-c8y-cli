. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantVersion" {
    BeforeEach {

    }

    It "Get the Cumulocity backend version" {
        $Response = PSc8y\Get-TenantVersion
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


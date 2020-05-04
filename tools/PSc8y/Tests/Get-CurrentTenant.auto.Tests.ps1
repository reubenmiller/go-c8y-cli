. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentTenant" {
    BeforeEach {

    }

    It "Get the current tenant (based on your current credentials)" {
        $Response = PSc8y\Get-CurrentTenant
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


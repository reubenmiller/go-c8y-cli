. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationCollectionByTenant" {
    BeforeEach {

    }

    It -Skip "Get a list of referenced applications on a given tenant (from management tenant)" {
        $Response = PSc8y\Get-ApplicationCollectionByTenant -Tenant mycompany
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


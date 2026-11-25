. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubTenant" {
    BeforeEach {

    }

    It -Skip "Get datahub tenant configuration" {
        $Response = PSc8y\Get-DataHubTenant
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubConfigurationCollection" {
    BeforeEach {

    }

    It -Skip "List the datahub offloading configurations" {
        $Response = PSc8y\Get-DataHubConfigurationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


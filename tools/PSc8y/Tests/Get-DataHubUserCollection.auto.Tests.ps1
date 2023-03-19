. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubUserCollection" {
    BeforeEach {

    }

    It -Skip "List the datahub users" {
        $Response = PSc8y\Get-DataHubUserCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


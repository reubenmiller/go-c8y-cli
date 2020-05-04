. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataBrokerConnectorCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of data broker connectors" {
        $Response = PSc8y\Get-DataBrokerConnectorCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


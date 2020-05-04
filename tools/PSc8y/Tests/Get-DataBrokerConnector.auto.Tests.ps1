. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataBrokerConnector" {
    BeforeEach {

    }

    It -Skip "Get a data broker connector" {
        $Response = PSc8y\Get-DataBrokerConnector -Id $DataBroker.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


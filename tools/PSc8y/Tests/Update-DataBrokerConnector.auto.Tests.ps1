. $PSScriptRoot/imports.ps1

Describe -Name "Update-DataBrokerConnector" {
    BeforeEach {

    }

    It -Skip "Change the status of a specific data broker connector by given connector id" {
        $Response = PSc8y\Update-DataBroker -Id 12345 -Status SUSPENDED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


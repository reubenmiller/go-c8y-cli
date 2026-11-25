. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubSchedulerCollection" {
    BeforeEach {

    }

    It -Skip "List the datahub scheduled items" {
        $Response = PSc8y\Get-DataHubSchedulerCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


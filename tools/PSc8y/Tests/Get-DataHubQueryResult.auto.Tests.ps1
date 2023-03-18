. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubQueryResult" {
    BeforeEach {

    }

    It -Skip "Get a list of alarms from datahub" {
        $Response = PSc8y\Get-DataHubQueryResult -Sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


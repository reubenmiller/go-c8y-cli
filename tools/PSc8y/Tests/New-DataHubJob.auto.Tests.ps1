. $PSScriptRoot/imports.ps1

Describe -Name "New-DataHubJob" {
    BeforeEach {

    }

    It -Skip "Create a new datahub job" {
        $Response = PSc8y\New-DataHubJob -Sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Create a new datahub job using context" {
        $Response = PSc8y\New-DataHubJob -Sql "SELECT * FROM alarms" -Context "myTenantIdDataLake", "Dremio", "myTenantId"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


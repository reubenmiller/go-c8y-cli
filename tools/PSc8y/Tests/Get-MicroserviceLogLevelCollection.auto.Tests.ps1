. $PSScriptRoot/imports.ps1

Describe -Name "Get-MicroserviceLogLevelCollection" {
    BeforeEach {

    }

    It -Skip "List log levels of microservice" {
        $Response = PSc8y\Get-MicroserviceLogLevelCollection -Name my-microservice
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Set-MicroserviceLogLevel" {
    BeforeEach {

    }

    It -Skip "Set log level of microservice" {
        $Response = PSc8y\Set-MicroserviceLogLevel -Name my-microservice -LoggerName org.example.microservice -LogLevel DEBUG
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


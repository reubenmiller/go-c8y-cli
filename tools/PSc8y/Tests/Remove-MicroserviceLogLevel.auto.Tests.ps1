. $PSScriptRoot/imports.ps1

Describe -Name "Remove-MicroserviceLogLevel" {
    BeforeEach {

    }

    It "Delete configured log level of microservice" {
        $Response = PSc8y\Remove-MicroserviceLogLevel -Name my-microservice -LoggerName org.example.microservice
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


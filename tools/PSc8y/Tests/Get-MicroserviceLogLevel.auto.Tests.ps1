. $PSScriptRoot/imports.ps1

Describe -Name "Get-MicroserviceLogLevel" {
    BeforeEach {

    }

    It "Get log level of microservice for a package" {
        $Response = PSc8y\Get-MicroserviceLogLevel -Name my-microservice -LoggerName org.example
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


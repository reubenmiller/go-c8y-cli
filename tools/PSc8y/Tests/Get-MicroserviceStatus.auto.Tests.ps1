. $PSScriptRoot/imports.ps1

Describe -Name "Get-MicroserviceStatus" {
    BeforeEach {

    }

    It "Get microservice status" {
        $Response = PSc8y\Get-MicroserviceStatus -Id 1234 -Dry
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get microservice status (using pipeline)" {
        $Response = PSc8y\Get-MicroserviceCollection | Get-MicroserviceStatus -Dry
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-MicroserviceCollection" {
    BeforeEach {

    }

    It "Get microservices" {
        $Response = PSc8y\Get-MicroserviceCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-MicroserviceBootstrapUser" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription

    }

    It -Skip "Get application bootstrap user" {
        $Response = PSc8y\Get-MicroserviceBootstrapUser -Id $App.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}


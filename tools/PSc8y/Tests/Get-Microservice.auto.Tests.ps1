. $PSScriptRoot/imports.ps1

Describe -Name "Get-Microservice" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription

    }

    It -Skip "Get an microservice by id" {
        $Response = PSc8y\Get-Microservice -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get an microservice by name" {
        $Response = PSc8y\Get-Microservice -Id $App.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}


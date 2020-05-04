. $PSScriptRoot/imports.ps1

Describe -Name "Enable-Microservice" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription

    }

    It -Skip "Enable (subscribe) to a microservice" {
        $Response = PSc8y\Enable-Microservice -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}


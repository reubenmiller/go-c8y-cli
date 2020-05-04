. $PSScriptRoot/imports.ps1

Describe -Name "Update-Microservice" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription

    }

    It -Skip "Update microservice availability to MARKET" {
        $Response = PSc8y\Update-Microservice -Id $App.id -Availability "MARKET"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}


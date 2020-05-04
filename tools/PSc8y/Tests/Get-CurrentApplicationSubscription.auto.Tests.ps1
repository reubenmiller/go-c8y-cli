. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentApplicationSubscription" {
    BeforeEach {

    }

    It -Skip "List the current application users/subscriptions" {
        $Response = PSc8y\Get-CurrentApplicationSubscription
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-Notification2SubscriptionCollection" {
    BeforeEach {

    }

    It -Skip "Get existing subscriptions" {
        $Response = PSc8y\Get-Notification2SubscriptionCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


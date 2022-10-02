. $PSScriptRoot/imports.ps1

Describe -Name "New-Notification2Token" {
    BeforeEach {

    }

    It -Skip "Create a new token which is valid for 30 minutes" {
        $Response = PSc8y\New-Notification2Token -Name testSubscription -Subscriber testSubscriber -ExpiresInMinutes 30
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


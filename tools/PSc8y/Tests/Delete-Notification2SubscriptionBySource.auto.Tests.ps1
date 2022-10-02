. $PSScriptRoot/imports.ps1

Describe -Name "Delete-Notification2SubscriptionBySource" {
    BeforeEach {

    }

    It "Delete a subscription associated with a device" {
        $Response = PSc8y\Delete-Notification2SubscriptionBySource -Device 12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


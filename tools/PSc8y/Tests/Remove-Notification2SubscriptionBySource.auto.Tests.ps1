. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Notification2SubscriptionBySource" {
    BeforeEach {

    }

    It -Skip "Delete a subscription associated with a device" {
        $Response = PSc8y\Remove-Notification2SubscriptionBySource -Device 12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


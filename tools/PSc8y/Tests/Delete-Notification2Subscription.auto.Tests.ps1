. $PSScriptRoot/imports.ps1

Describe -Name "Delete-Notification2Subscription" {
    BeforeEach {

    }

    It -Skip "Delete a subscription" {
        $Response = PSc8y\Delete-Notification2Subscription -Id 12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


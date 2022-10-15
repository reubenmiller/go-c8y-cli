. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Notification2Subscription" {
    BeforeEach {

    }

    It -Skip "Delete a subscription" {
        $Response = PSc8y\Remove-Notification2Subscription -Id 12345
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


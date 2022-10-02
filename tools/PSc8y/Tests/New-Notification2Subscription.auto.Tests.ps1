. $PSScriptRoot/imports.ps1

Describe -Name "New-Notification2Subscription" {
    BeforeEach {

    }

    It -Skip "Create a new subscription to operations for a specific device" {
        $Response = PSc8y\New-Notification2Subscription -Name deviceSub -Device 12345 -Context mo -ApiFilter operations
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


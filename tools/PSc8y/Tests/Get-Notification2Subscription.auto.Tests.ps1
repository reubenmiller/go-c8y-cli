. $PSScriptRoot/imports.ps1

Describe -Name "Get-Notification2Subscription" {
    BeforeEach {

    }

    It "Get an existing subscription" {
        $Response = PSc8y\Get-Notification2Subscription -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


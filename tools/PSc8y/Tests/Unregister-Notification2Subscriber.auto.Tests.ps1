. $PSScriptRoot/imports.ps1

Describe -Name "Unregister-Notification2Subscriber" {
    BeforeEach {

    }

    It "Unsubscribe a subscriber using its token" {
        $Response = PSc8y\Unregister-Notification2Subscriber -Token "eyJhbGciOiJSUzI1NiJ9"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


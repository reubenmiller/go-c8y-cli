. $PSScriptRoot/imports.ps1

Describe -Name "Get-Event" {
    BeforeEach {
        $TestEvent = PSc8y\New-TestEvent

    }

    It "Get event" {
        $Response = PSc8y\Get-Event -Id $TestEvent.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestEvent.source.id) {
            PSc8y\Remove-ManagedObject -Id $TestEvent.source.id -ErrorAction SilentlyContinue
        }

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Event" {
    BeforeEach {
        $TestEvent = PSc8y\New-TestEvent

    }

    It "Delete an event" {
        $Response = PSc8y\Remove-Event -Id $TestEvent.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        if ($TestEvent.source.id) {
            PSc8y\Remove-ManagedObject -Id $TestEvent.source.id -ErrorAction SilentlyContinue
        }

    }
}


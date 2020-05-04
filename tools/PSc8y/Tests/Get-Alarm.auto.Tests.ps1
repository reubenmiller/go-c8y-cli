. $PSScriptRoot/imports.ps1

Describe -Name "Get-Alarm" {
    BeforeEach {
        $TestAlarm = PSc8y\New-TestAlarm

    }

    It "Get alarm" {
        $Response = PSc8y\Get-Alarm -Id $TestAlarm.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestAlarm.source.id) {
            PSc8y\Remove-ManagedObject -Id $TestAlarm.source.id -ErrorAction SilentlyContinue
        }

    }
}


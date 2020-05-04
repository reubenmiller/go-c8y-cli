. $PSScriptRoot/imports.ps1

Describe -Name "Update-Alarm" {
    BeforeEach {
        $Device = New-TestDevice
        $Alarm = New-TestAlarm -Device $Device.id

    }

    It "Acknowledge an existing alarm" {
        $Response = PSc8y\Update-Alarm -Id $Alarm.id -Status ACKNOWLEDGED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Acknowledge an existing alarm (using pipeline)" {
        $Response = PSc8y\Get-Alarm -Id $Alarm.id | PSc8y\Update-Alarm -Status ACKNOWLEDGED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update severity of an existing alarm to CRITICAL" {
        $Response = PSc8y\Update-Alarm -Id $Alarm.id -Severity CRITICAL
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Get-AlarmCount" {
    BeforeEach {
        $Device = New-TestDevice
        $Alarm = PSc8y\New-Alarm -Device $Device.id -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR

    }

    It "Get number of active alarms with the severity set to MAJOR" {
        $Response = PSc8y\Get-AlarmCount -Severity MAJOR
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get number of active alarms which occurred in the last 10 minutes" {
        $Response = PSc8y\Get-AlarmCount -DateFrom "-10m" -Status ACTIVE
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get number of active alarms which occurred in the last 10 minutes on a device" {
        $Response = PSc8y\Get-AlarmCount -DateFrom "-10m" -Status ACTIVE -Device $Device.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get number of alarms from a list of devices using pipeline" {
        $Response = PSc8y\Get-Device -Id $Device.id | Get-AlarmCount -DateFrom "-10m"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


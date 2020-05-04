. $PSScriptRoot/imports.ps1

Describe -Name "Get-AlarmCollection" {
    BeforeEach {
        $Device = New-TestDevice
        $Alarm = PSc8y\New-Alarm -Device $Device.id -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR

    }

    It "Get alarms with the severity set to MAJOR" {
        $Response = PSc8y\Get-AlarmCollection -Severity MAJOR -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get active alarms which occurred in the last 10 minutes" {
        $Response = PSc8y\Get-AlarmCollection -DateFrom "-10m" -Status ACTIVE
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get active alarms from a device (using pipeline)" {
        $Response = PSc8y\Get-DeviceCollection -Name $Device.name | Get-AlarmCollection -Status ACTIVE
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


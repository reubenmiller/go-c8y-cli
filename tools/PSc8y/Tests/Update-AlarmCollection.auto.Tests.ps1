. $PSScriptRoot/imports.ps1

Describe -Name "Update-AlarmCollection" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Alarm = PSc8y\New-Alarm -Device $Device.id -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR

    }

    It "Update the status of all active alarms on a device to ACKNOWLEDGED" {
        $Response = PSc8y\Update-AlarmCollection -Device $Device.id -Status ACTIVE -NewStatus ACKNOWLEDGED
        $LASTEXITCODE | Should -Be 0
    }

    It "Update the status of all active alarms on a device to ACKNOWLEDGED (using pipeline)" {
        $Response = PSc8y\Get-Device -Id $Device.id | PSc8y\Update-AlarmCollection -Status ACTIVE -NewStatus ACKNOWLEDGED
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Remove-AlarmCollection" {
    BeforeEach {
        $device = New-TestDevice
        $TestDevice = PSc8y\New-TestDevice

    }

    It "Remove alarms on the device with the severity set to MAJOR" {
        $Response = PSc8y\Remove-AlarmCollection -Device $TestDevice.id -Severity MAJOR
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove alarms on the device which are active and created in the last 10 minutes" {
        $Response = PSc8y\Remove-AlarmCollection -Device $device.id -DateFrom "-10m" -Status ACTIVE
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove alarms on the device which are active and created in the last 10 minutes (using pipeline)" {
        $Response = PSc8y\Get-Device -Id $device.id | PSc8y\Remove-AlarmCollection -DateFrom "-10m" -Status ACTIVE
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id
        if ($TestDevice.id) {
            PSc8y\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue
        }

    }
}


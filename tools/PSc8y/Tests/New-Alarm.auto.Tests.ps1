. $PSScriptRoot/imports.ps1

Describe -Name "New-Alarm" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Create a new alarm for device" {
        $Response = PSc8y\New-Alarm -Device $device.id -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create a new alarm for device (using pipeline)" {
        $Response = PSc8y\Get-Device -Id $device.id | PSc8y\New-Alarm -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $device.id

    }
}


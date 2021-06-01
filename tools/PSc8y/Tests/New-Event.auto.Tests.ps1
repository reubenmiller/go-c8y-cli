. $PSScriptRoot/imports.ps1

Describe -Name "New-Event" {
    BeforeEach {
        $device = New-TestDevice

    }

    It "Create a new event for a device" {
        $Response = PSc8y\New-Event -Device $device.id -Type c8y_TestEvent -Text "Test event"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create a new event for a device (using pipeline)" {
        $Response = PSc8y\Get-Device -Id $device.id | PSc8y\New-Event -Type c8y_TestEvent -Time "-0s" -Text "Test event"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


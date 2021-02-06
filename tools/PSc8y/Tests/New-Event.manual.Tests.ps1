. $PSScriptRoot/imports.ps1

Describe -Name "New-Event" {
    BeforeEach {
        $device = New-TestDevice

    }

    It "Try to create a new event without the text field" {
        $Response = PSc8y\New-Event -Device $device.id -Type c8y_TestAlarm -ErrorVariable "ErrorMessages"
        $LASTEXITCODE | Should -Not -Be 0
        $Response | Should -BeNullOrEmpty
        $ErrorMessages[-1] -match "Body missing required properties: text" | Should -HaveCount 1
    }

    It "Create event where the template provides the required fields" {
        $options = @{
            Device = $device.id
            Type = "c8y_TestAlarm"
            Template = "{text: 'my custom text'}"
            ErrorVariable = "ErrorMessages"
        }
        $Response = PSc8y\New-Event @options
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $ErrorMessages -match "Body missing required properties: text" | Should -BeNullOrEmpty
    }

    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


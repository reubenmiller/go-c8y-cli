. $PSScriptRoot/imports.ps1

Describe -Name "New-Event" {
    BeforeEach {
        $device = New-TestDevice

    }

    It "Try to create a new event without the text field" {
        $ErrorMessages = $( $Response = PSc8y\New-Event -Device $device.id -Type c8y_TestAlarm -Force ) 2>&1
        $LASTEXITCODE | Should -Be 101
        $Response | Should -BeNullOrEmpty
        $ErrorMessages[-1] -match "Body is missing required properties: text" | Should -HaveCount 1
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
        $ErrorMessages -match "Body is missing required properties: text" | Should -BeNullOrEmpty
    }

    It "Create event where the template is missing required fields" {
        $options = @{
            Device = $device.id
            Type = "c8y_TestAlarm"
            Template = "{customText: 'my custom text'}"
        }
        $stderr = $( $Response = PSc8y\New-Event @options ) 2>&1
        $LASTEXITCODE | Should -Be 101
        $Response | Should -BeNullOrEmpty
        $stderr | Should -HaveCount 1
        $stderr | Should -Match "Body is missing required properties: text"
    }

    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessPassthroughConfiguration" {
    BeforeEach {

    }

    It "Create a SSH passthrough configuration to the localhost" {
        $Response = PSc8y\New-RemoteAccessPassthroughConfiguration -Device mydevice

        $LASTEXITCODE | Should -Be 0
    }

    It "Create a SSH passthrough configuration with custom details" {
        $Response = PSc8y\New-RemoteAccessPassthroughConfiguration -Device mydevice -Hostname customhost -Port 1234 -Name "My custom configuration"

        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


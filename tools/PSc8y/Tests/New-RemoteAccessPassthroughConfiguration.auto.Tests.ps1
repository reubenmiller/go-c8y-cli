. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessPassthroughConfiguration" {
    BeforeEach {

    }

    It -Skip "Create a SSH passthrough configuration to the localhost" {
        $Response = PSc8y\New-RemoteAccessPassthroughConfiguration -Device device01

        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Create a SSH passthrough configuration with custom details" {
        $Response = PSc8y\New-RemoteAccessPassthroughConfiguration -Device device01 -Hostname customhost -Port 1234 -Name "My custom configuration"

        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


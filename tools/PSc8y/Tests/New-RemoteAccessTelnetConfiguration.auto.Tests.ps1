. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessTelnetConfiguration" {
    BeforeEach {

    }

    It -Skip "Create a telnet configuration" {
        $Response = PSc8y\New-RemoteAccessTelnetConfiguration -Device device01

        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


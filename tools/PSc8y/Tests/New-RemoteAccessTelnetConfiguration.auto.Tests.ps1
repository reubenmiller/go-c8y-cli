. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessTelnetConfiguration" {
    BeforeEach {

    }

    It "Create a telnet configuration" {
        $Response = PSc8y\New-RemoteAccessTelnetConfiguration

        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


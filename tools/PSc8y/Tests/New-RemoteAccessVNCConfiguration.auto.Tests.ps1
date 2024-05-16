. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessVNCConfiguration" {
    BeforeEach {

    }

    It -Skip "Create a VNC configuration that does not require a password" {
        $Response = PSc8y\New-RemoteAccessVNCConfiguration -Device device01

        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Create a VNC configuration that requires a password" {
        $Response = PSc8y\New-RemoteAccessVNCConfiguration -Device device01 -Password 'asd08dcj23dsf{@#9}'
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


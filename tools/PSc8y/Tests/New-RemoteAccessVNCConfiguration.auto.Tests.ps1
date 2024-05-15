. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessVNCConfiguration" {
    BeforeEach {

    }

    It "Create a VNC configuration that does not require a password" {
        $Response = PSc8y\New-RemoteAccessVNCConfiguration

        $LASTEXITCODE | Should -Be 0
    }

    It "Create a VNC configuration that requires a password" {
        $Response = PSc8y\New-RemoteAccessVNCConfiguration -Password 'asd08dcj23dsf{@#9}'
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


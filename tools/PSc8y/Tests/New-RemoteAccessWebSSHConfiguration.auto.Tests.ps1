. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessWebSSHConfiguration" {
    BeforeEach {

    }

    It "Create a webssh configuration" {
        $Response = PSc8y\New-RemoteAccessWebSSHConfiguration

        $LASTEXITCODE | Should -Be 0
    }

    It "Create a webssh configuration with a custom hostname and port" {
        $Response = PSc8y\New-RemoteAccessWebSSHConfiguration -Hostname 127.0.0.1 -Port 2222
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


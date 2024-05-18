. $PSScriptRoot/imports.ps1

Describe -Name "New-RemoteAccessWebSSHConfiguration" {
    BeforeEach {

    }

    It -Skip "Create a webssh configuration (with username/password authentication)" {
        $Response = PSc8y\New-RemoteAccessWebSSHConfiguration -Device device01 -Username admin -Password "3Xz7cEj%oAmt#dnUMP*N"

        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Create a webssh configuration with a custom hostname and port (with ssh key authentication)" {
        $Response = PSc8y\New-RemoteAccessWebSSHConfiguration -Device device01 -Hostname 127.0.0.1 -Port 2222 -Username admin -PrivateKey "xxxx" -PublicKey "yyyyy"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


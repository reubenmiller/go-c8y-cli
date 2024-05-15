. $PSScriptRoot/imports.ps1

Describe -Name "Get-RemoteAccessConfiguration" {
    BeforeEach {

    }

    It "Get existing remote access configuration" {
        $Response = PSc8y\Get-RemoteAccessConfiguration -Device mydevice -Id 1
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


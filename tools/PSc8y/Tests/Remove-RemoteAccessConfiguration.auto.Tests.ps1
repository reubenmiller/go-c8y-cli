. $PSScriptRoot/imports.ps1

Describe -Name "Remove-RemoteAccessConfiguration" {
    BeforeEach {

    }

    It "Delete an existing remote access configuration" {
        $Response = PSc8y\Remove-RemoteAccessConfiguration -Device device01 -Id 1
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


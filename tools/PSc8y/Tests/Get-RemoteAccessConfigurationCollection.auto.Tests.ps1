. $PSScriptRoot/imports.ps1

Describe -Name "Get-RemoteAccessConfigurationCollection" {
    BeforeEach {

    }

    It "List remote access configurations for a given device" {
        $Response = PSc8y\Get-RemoteAccessConfigurationCollection -Device mydevice
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


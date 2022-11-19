. $PSScriptRoot/imports.ps1

Describe -Name "Send-Configuration" {
    BeforeEach {

    }

    It -Skip "Send a configuration file to a device" {
        $Response = PSc8y\Send-Configuration -Device mydevice -Configuration 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Send a configuration file to multiple devices" {
        $Response = PSc8y\Get-DeviceCollection | Send-Configuration -Configuration 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


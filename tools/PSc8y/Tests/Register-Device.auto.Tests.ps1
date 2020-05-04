. $PSScriptRoot/imports.ps1

Describe -Name "Register-Device" {
    BeforeEach {

    }

    It "Register a new device" {
        $Response = PSc8y\Register-Device -Id "ASDF098SD1J10912UD92JDLCNCU8"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceRequest -Id "ASDF098SD1J10912UD92JDLCNCU8"

    }
}


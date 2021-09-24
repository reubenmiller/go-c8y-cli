. $PSScriptRoot/imports.ps1

Describe -Name "Install-FirmwareVersion" {
    BeforeEach {

    }

    It -Skip "Get a firmware version" {
        $Response = PSc8y\Install-FirmwareVersion -Device $mo.id -Firmware linux-iot -Version 1.0.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


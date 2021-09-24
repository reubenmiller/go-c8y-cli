. $PSScriptRoot/imports.ps1

Describe -Name "Get-FirmwareVersionCollection" {
    BeforeEach {
        $name = New-RandomString -Prefix "firmware_"
        $firmware = New-Firmware -Name $name
        $firmwareVersion = PSc8y\New-FirmwareVersion -Firmware $firmware.id -Version "1.0.0" -Url "https://blob.azure.com/device-firmare/1.0.0/image.mender"

    }

    It "Get a list of firmware package versions" {
        $Response = PSc8y\Get-FirmwareVersionCollection -Firmware $firmware.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Firmware -Id $firmware.id

    }
}


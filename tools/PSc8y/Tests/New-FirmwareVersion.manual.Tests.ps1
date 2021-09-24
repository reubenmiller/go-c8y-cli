. $PSScriptRoot/imports.ps1

Describe -Name "New-FirmwareVersion" {
    BeforeEach {
        $name = New-RandomString -Prefix "firmware_"
        $firmware = New-Firmware -Name $name
    }

    It "Create a new version to an existing firmware package" {
        $Response = PSc8y\New-FirmwareVersion -Firmware $firmware.id -Version "1.0.0" -Url "https://blob.azure.com/device-firmare/1.0.0/image.mender"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Firmware -Id $firmware.id

    }
}


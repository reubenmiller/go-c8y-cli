. $PSScriptRoot/imports.ps1

Describe -Name "New-FirmwarePatch" {
    BeforeEach {
        $name = New-RandomString -Prefix "firmware_"
        $firmware = New-Firmware -Name $name
        $firmwareVersion = New-FirmwareVersion -Firmware $firmware.id -Version "1.9.0" -Url "https://test.com/blob/1.9.0/package.deb"

    }

    It "Create a new patch to an existing firmware package version" {
        $Response = PSc8y\New-FirmwarePatch -Firmware $firmware.id -Version "1.9.1" -DependencyVersion "1.9.0" -Url "https://test.com/blob/1.9.1/package.deb"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Firmware -Id $firmware.id

    }
}


. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceRequest" {
    BeforeEach {
        $serial_91019192078 = New-RandomString -Prefix "serial"
        $DeviceRequest = Register-Device -Id "$serial_91019192078"

    }

    It "Delete a new device request" {
        $Response = PSc8y\Remove-DeviceRequest -Id "$serial_91019192078"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


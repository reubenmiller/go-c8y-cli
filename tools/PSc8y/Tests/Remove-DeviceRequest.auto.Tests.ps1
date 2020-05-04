. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceRequest" {
    BeforeEach {
        $DeviceRequest = Register-Device -Id "91019192078"

    }

    It "Delete a new device request" {
        $Response = PSc8y\Remove-DeviceRequest -Id "91019192078"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


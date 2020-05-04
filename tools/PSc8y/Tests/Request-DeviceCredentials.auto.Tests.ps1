. $PSScriptRoot/imports.ps1

Describe -Name "Request-DeviceCredentials" {
    BeforeEach {

    }

    It -Skip "Request credentials for a new device" {
        $Response = PSc8y\Request-DeviceCredentials -Id "device-AD76-matrixer"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceRequest -Id "device-AD76-matrixer"

    }
}


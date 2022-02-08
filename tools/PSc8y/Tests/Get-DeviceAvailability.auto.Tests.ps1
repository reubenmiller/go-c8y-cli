. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceAvailability" {
    BeforeEach {
        $Device = PSc8y\Get-Device -Id "mobile-device01"

    }

    It "Get a device's availability by id" {
        $Response = PSc8y\Get-DeviceAvailability -Id $Device.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Get a device's availability by name" {
        $Response = PSc8y\Get-DeviceAvailability -Id $Device.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


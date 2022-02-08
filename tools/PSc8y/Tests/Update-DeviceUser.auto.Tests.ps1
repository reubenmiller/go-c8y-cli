. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceUser" {
    BeforeEach {
        $device = PSc8y\Get-Device -Id "mobile-device01"

    }

    It "Enable a device user" {
        $Response = PSc8y\Update-DeviceUser -Id $device.id -Enabled
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Disable a device user" {
        $Response = PSc8y\Update-DeviceUser -Id $device.name -Enabled:$false
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}


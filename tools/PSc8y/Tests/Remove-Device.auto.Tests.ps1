. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Device" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It -Skip "Remove device by id" {
        $Response = PSc8y\Remove-Device -Id $device.id
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Remove device by name" {
        $Response = PSc8y\Remove-Device -Id $device.name
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete device and related device user/credentials" {
        $Response = PSc8y\Remove-Device -Id "device01" -WithDeviceUser
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}


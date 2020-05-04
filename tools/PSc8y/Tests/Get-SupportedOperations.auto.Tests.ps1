. $PSScriptRoot/imports.ps1

Describe -Name "Get-SupportedOperations" {
    BeforeEach {
        $device = PSc8y\New-TestDevice
        $device = PSc8y\Update-ManagedObject -Id $device.id -Data @{ c8y_SupportedOperations = @( "c8y_Restart", "c8y_SoftwareList", "c8y_Firmware" ) }

    }

    It "Get the supported operations of a device by name" {
        $Response = PSc8y\Get-SupportedOperations -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get the supported operations of a device (using pipeline)" {
        $Response = PSc8y\Get-SupportedOperations -Device $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


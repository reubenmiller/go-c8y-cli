. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceCollection" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It -Skip "c8y devices list --name "sensor*" --type myType" {
        $Response = PSc8y\Get-DeviceCollection -Name "sensor*" -Type myType
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}


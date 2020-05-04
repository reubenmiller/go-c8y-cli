. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ChildDeviceReference" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\New-ChildDeviceReference -Device $Device.id -NewChild $ChildDevice.id

    }

    It "Unassign a child device from its parent device" {
        $Response = PSc8y\Remove-ChildDeviceReference -Device $Device.id -ChildDevice $ChildDevice.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildDevice.id
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


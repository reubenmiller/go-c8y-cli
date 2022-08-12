. $PSScriptRoot/imports.ps1

Describe -Name "Add-ChildAdditionToDevice" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice

    }

    It "Assign a managed object as a child addition to an existing device" {
        $Response = PSc8y\Add-ChildAdditionToDevice -Device $Device.id -Child $ChildDevice.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Assign a managed object as a child addition to an existing device (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $ChildDevice.id | Add-ChildDeviceToDevice -Device $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildDevice.id
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}


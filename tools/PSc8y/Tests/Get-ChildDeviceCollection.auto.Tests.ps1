. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildDeviceCollection" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

    }

    It "Get a list of the child devices of an existing device" {
        $Response = PSc8y\Get-ChildDeviceCollection -Device $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of the child devices of an existing device (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Get-ChildDeviceCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $ChildDevice.id

    }
}


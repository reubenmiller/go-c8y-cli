. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildDeviceReference" {
    BeforeEach {
        $Agent = New-TestAgent
        $Device = New-TestDevice
        $Ref = Add-ChildDeviceToDevice -Device $Agent.id -Child $Device.id

    }

    It "Get an existing child device reference" {
        $Response = PSc8y\Get-ChildDeviceReference -Device $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}


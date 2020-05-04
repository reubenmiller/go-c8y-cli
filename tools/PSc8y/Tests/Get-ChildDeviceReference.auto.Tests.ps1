. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildDeviceReference" {
    BeforeEach {
        $Agent = New-TestDevice -AsAgent
        $Device = New-TestDevice
        $Ref = New-ChildDeviceReference -Device $Agent.id -NewChild $Device.id

    }

    It "Get an existing child device reference" {
        $Response = PSc8y\Get-ChildDeviceReference -Device $Agent.id -Reference $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}


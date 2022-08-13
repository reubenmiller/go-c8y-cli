. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildAddition" {
    BeforeEach {
        $Agent = New-TestAgent
        $Device = New-TestDevice
        $Ref = Add-AdditionToDeviceGroup -Group $Agent.id -Child $Device.id

    }

    It "Get an existing child addition" {
        $Response = PSc8y\Get-ChildAddition -Id $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}


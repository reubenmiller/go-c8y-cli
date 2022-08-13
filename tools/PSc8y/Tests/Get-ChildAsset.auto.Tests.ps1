. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildAsset" {
    BeforeEach {
        $Agent = New-TestAgent
        $Device = New-TestDevice
        $Ref = Add-ChildAssetToDeviceGroup -Group $Agent.id -Child $Device.id

    }

    It "Get an existing child asset reference" {
        $Response = PSc8y\Get-ChildAsset -Id $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}


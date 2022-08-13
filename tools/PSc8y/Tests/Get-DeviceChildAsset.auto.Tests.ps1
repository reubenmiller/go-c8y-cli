. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceChildAsset" {
    BeforeEach {
        $Agent = New-TestAgent
        $Device = New-TestDevice
        $Ref = Add-ChildAssetToManagedObject -Id $agent.id -ChildDevice $Device.id

    }

    It "Get an existing child device reference" {
        $Response = PSc8y\Get-DeviceChildAsset -Device $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}


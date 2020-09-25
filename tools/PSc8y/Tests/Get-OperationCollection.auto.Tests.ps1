. $PSScriptRoot/imports.ps1

Describe -Name "Get-OperationCollection" {
    BeforeEach {
        $Agent = New-TestAgent
        $Operation1 = PSc8y\New-TestOperation -Device $Agent.id
        $Device = New-TestDevice
        Add-ChildDeviceToDevice -Device $Agent.id -NewChild $Device.id
        $Operation2 = PSc8y\New-TestOperation -Device $Device.id
        $Agent2 = New-TestAgent
        $Operation3 = PSc8y\New-TestOperation -Device $Agent2.id

    }

    It "Get a list of pending operations" {
        $Response = PSc8y\Get-OperationCollection -Status PENDING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of pending operations for a given agent and all of its child devices" {
        $Response = PSc8y\Get-OperationCollection -Agent $Agent.id -Status PENDING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of pending operations for a device" {
        $Response = PSc8y\Get-OperationCollection -Device $Device.id -Status PENDING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get operations from a device (using pipeline)" {
        $Response = PSc8y\Get-DeviceCollection -Name $Agent2.name | Get-OperationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -id $Agent.id
        Remove-ManagedObject -id $Device.id
        PSc8y\Remove-ManagedObject -Id $Agent2.id

    }
}

